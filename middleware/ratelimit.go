package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// CRIT-02: กัน brute force / credential stuffing / bcrypt CPU exhaustion บน POST /login
// ใช้ token bucket ต่อ IP (golang.org/x/time/rate) เก็บ state ใน memory ต่อ instance
// (พอสำหรับ deployment แบบ single instance บน Dokploy) — ถ้าขยายเป็นหลาย instance
// เมื่อไร ต้องย้าย state ไป Redis เพราะแต่ละ instance จะนับแยกกัน

// rateLimitConfig อ่านค่าจาก env ครั้งเดียวตอนสร้าง middleware
type rateLimitConfig struct {
	enabled bool          // LOGIN_RATE_ENABLED — kill switch (default true)
	max     int           // LOGIN_RATE_MAX — จำนวนครั้ง "ยั่งยืน" ต่อ window ต่อ IP
	window  time.Duration // LOGIN_RATE_WINDOW — หน้าต่างอ้างอิงที่ใช้คำนวณ refill
	burst   int           // LOGIN_RATE_BURST — ขนาด bucket (จำนวน request รวดเดียวที่ยอมก่อนโดนหน่วง)
	cleanup time.Duration // LOGIN_RATE_CLEANUP — รอบล้าง limiter ที่ไม่มี activity
}

// loadRateLimitConfig โหลดค่า config จาก env พร้อม default ที่ปลอดภัย
func loadRateLimitConfig() rateLimitConfig {
	cfg := rateLimitConfig{
		enabled: getenvBool("LOGIN_RATE_ENABLED", true),
		max:     getenvInt("LOGIN_RATE_MAX", 10),
		window:  getenvDuration("LOGIN_RATE_WINDOW", time.Minute),
		burst:   getenvInt("LOGIN_RATE_BURST", 10),
		cleanup: getenvDuration("LOGIN_RATE_CLEANUP", 5*time.Minute),
	}
	// guard ค่าที่เป็นไปไม่ได้ กัน divide-by-zero และ bucket ที่บล็อกทุกอย่าง
	if cfg.max < 1 {
		cfg.max = 1
	}
	if cfg.burst < 1 {
		cfg.burst = 1
	}
	if cfg.window <= 0 {
		cfg.window = time.Minute
	}
	if cfg.cleanup <= 0 {
		cfg.cleanup = 5 * time.Minute
	}
	return cfg
}

// ipLimiter ห่อ rate.Limiter พร้อม lastSeen ไว้ให้ cleanup รู้ว่าเมื่อไรลบทิ้งได้
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// loginRateLimiter จัดการ map ของ limiter ต่อ IP แบบ thread-safe
type loginRateLimiter struct {
	mu      sync.Mutex
	clients map[string]*ipLimiter
	cfg     rateLimitConfig
	limit   rate.Limit // อัตราเติม token = max ครั้ง ต่อ window
}

// newLoginRateLimiter สร้าง manager + เริ่ม goroutine cleanup
func newLoginRateLimiter(cfg rateLimitConfig) *loginRateLimiter {
	l := &loginRateLimiter{
		clients: make(map[string]*ipLimiter),
		cfg:     cfg,
		limit:   rate.Every(cfg.window / time.Duration(cfg.max)),
	}
	go l.cleanupLoop()
	return l
}

// getLimiter คืน *rate.Limiter ของ IP ที่ระบุ (สร้างใหม่ถ้ายังไม่มี)
// ถือ lock เฉพาะช่วง map lookup/insert แล้วปล่อยทันที — การนับ token (.Allow())
// ทำนอก lock เพราะ rate.Limiter thread-safe ในตัวอยู่แล้ว
func (l *loginRateLimiter) getLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	if c, ok := l.clients[ip]; ok {
		c.lastSeen = time.Now()
		return c.limiter
	}
	lim := rate.NewLimiter(l.limit, l.cfg.burst)
	l.clients[ip] = &ipLimiter{limiter: lim, lastSeen: time.Now()}
	return lim
}

// cleanupLoop ลบ limiter ที่ไม่มี activity เป็นระยะ กัน memory leak
// เงื่อนไขลบ: เงียบเกิน window **และ** bucket เติมกลับเต็มแล้ว (Tokens ครบ burst)
// การเช็ค Tokens ด้วยสำคัญ: ถ้าตั้ง burst > max การเติมเต็มใช้เวลานานกว่า window
// ถ้าลบตอนยังไม่เต็มแล้วสร้างใหม่ (เต็ม) = แจก token เกินที่ควร → reset โควตาได้
func (l *loginRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(l.cfg.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		for ip, c := range l.clients {
			if time.Since(c.lastSeen) > l.cfg.window && c.limiter.Tokens() >= float64(l.cfg.burst) {
				delete(l.clients, ip)
			}
		}
		l.mu.Unlock()
	}
}

// LoginRateLimit สร้าง gin middleware สำหรับ rate limit เฉพาะ POST /login
// เรียกครั้งเดียวตอน setup route (สร้าง manager + cleanup goroutine ตัวเดียวตลอดอายุโปรเซส)
// เกินโควตาตอบ 429 ตาม envelope เดิม { success:false, message } และ log เฉพาะตอนบล็อก
func LoginRateLimit() gin.HandlerFunc {
	cfg := loadRateLimitConfig()
	if !cfg.enabled {
		log.Println("[rate-limit] LOGIN_RATE_ENABLED=false — ปิด rate limit ของ /login")
		return func(c *gin.Context) { c.Next() }
	}

	rl := newLoginRateLimiter(cfg)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if rl.getLimiter(ip).Allow() {
			c.Next()
			return
		}

		// บล็อก — log เฉพาะตรงนี้ (ไม่ log ทุก request) พร้อมข้อมูลสืบสวน
		username := extractUsername(c)
		log.Printf("[rate-limit] blocked login ip=%s username=%q time=%s reason=rate_exceeded",
			ip, username, time.Now().Format(time.RFC3339))

		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"message": "พยายามเข้าสู่ระบบบ่อยเกินไป กรุณารอสักครู่แล้วลองใหม่",
		})
		c.Abort()
	}
}

// extractUsername ดึง username จาก body เพื่อ log ตอนบล็อกเท่านั้น (best-effort)
// เรียกเฉพาะบน block path ที่ c.Abort() ต่อทันที controller จึงไม่อ่าน body อีก
// จำกัดขนาดอ่าน 4KB กัน body ใหญ่ผิดปกติ; ล้มเหลว/พาร์สไม่ได้ = คืน "" ไม่โยน error
func extractUsername(c *gin.Context) string {
	if c.Request == nil || c.Request.Body == nil {
		return ""
	}
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 4096))
	if err != nil {
		return ""
	}
	// คืน body กลับเผื่อมี reader อื่นตามหลัง (ปกติ block path จะ abort อยู่แล้ว)
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	var payload struct {
		Username string `json:"username"`
	}
	if json.Unmarshal(body, &payload) != nil {
		return ""
	}
	return payload.Username
}

// --- env helpers (เฉพาะ package นี้) ---

func getenvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getenvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
