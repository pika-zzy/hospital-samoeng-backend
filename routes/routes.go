package routes

import (
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "hospitalbackend/docs" // เอกสาร Swagger ที่ generate จาก `swag init`
)

// trustedProxies คืนรายการ CIDR/IP ของ proxy ที่เชื่อถือ (Traefik/Dokploy)
// อ่านจาก env TRUSTED_PROXIES (คั่นด้วย comma) — ไม่ตั้งค่า = private ranges มาตรฐาน
// จำเป็นต่อความถูกต้องของ c.ClientIP(): ถ้าเชื่อทุก proxy (default ของ Gin) client
// ปลอม X-Forwarded-For แล้ว bypass rate limit ได้; ถ้าไม่เชื่อเลย ทุกคนจะเห็นเป็น IP
// ของ Traefik ตัวเดียวแล้วโดน 429 ร่วมกัน
func trustedProxies() []string {
	if v := os.Getenv("TRUSTED_PROXIES"); v != "" {
		parts := strings.Split(v, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if p = strings.TrimSpace(p); p != "" {
				out = append(out, p)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	// default: เชื่อเฉพาะช่วง private (Docker/Traefik อยู่ใน network ภายใน) + loopback
	// ⚠️ ควรตรวจ subnet จริงของ Dokploy หลัง deploy แล้วรัดให้แคบลงผ่าน TRUSTED_PROXIES
	return []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "127.0.0.1/32", "::1/128"}
}

func SetupRouter() *gin.Engine {

	r := gin.Default()

	// CRIT-03: ลด RAM buffer ของ multipart parsing จาก default 32MB → 8MB
	// (เพดานขนาด body จริงบังคับต่อ handler ด้วย MaxBytesReader ใน enforceUploadLimit)
	r.MaxMultipartMemory = 8 << 20

	// CRIT-02: ตั้ง trusted proxies ให้ c.ClientIP() อ่าน IP client จริงหลัง Traefik
	// ค่าผิด = security misconfig → fail fast
	if err := r.SetTrustedProxies(trustedProxies()); err != nil {
		log.Fatal("SetTrustedProxies ไม่สำเร็จ (ตรวจค่า TRUSTED_PROXIES): ", err)
	}

	// Swagger UI ที่ /swagger/index.html — เปิดเฉพาะตอนตั้ง ENABLE_SWAGGER=true ใน .env
	// ปิดโดย default กันเผยโครงสร้าง API ทั้งหมดบน production
	if os.Getenv("ENABLE_SWAGGER") == "true" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
	// ไฟล์อัปโหลด (public ไม่มี auth) — ย้ายจาก r.Static มาเป็น handler ของตัวเอง
	// เพื่อรองรับ ?download= ที่บังคับดาวน์โหลดข้าม origin ได้ ดู routes/uploads_route.go
	UploadsRoute(r)

	// Middleware สำหรับ CORS (อนุญาตให้ Frontend ที่รันบนพอร์ต 5173 เข้าถึง API ได้)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		// FIX: เพิ่ม PATCH — เดิมไม่มี ทำให้ PATCH /users/:id/status โดน browser block ตอน CORS preflight
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// เรียกใช้ Route ย่อยที่เราแยกไว้
	LoginRoute(r)
	NewsRoutes(r)
	ActivityRoutes(r)
	PersonnelRoutes(r)
	MenuRoutes(r)
	ITARoutes(r)
	MOITRoutes(r)
	PermissionRoutes(r)
	PopupRoutes(r)
	HeroRoutes(r)
	ContentRoutes(r)
	return r
}
