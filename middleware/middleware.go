package middleware

import (
	"errors"
	"fmt"
	"hospitalbackend/database"
	model "hospitalbackend/models"
	"hospitalbackend/utils"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(401, gin.H{"success": false, "message": "missing token"})
			c.Abort()
			return
		}

		// FIX: เช็ค prefix "Bearer " ก่อนตัด — ของเดิม slice ตรง ๆ ทำให้ panic ได้ถ้า header สั้นกว่า 7 ตัวอักษร
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(401, gin.H{"success": false, "message": "invalid token"})
			c.Abort()
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims := &utils.Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// HIGH-08: ยืนยัน signing method เป็น HMAC เท่านั้น — กัน algorithm-confusion attack
			// (เช่นปลอม token เป็น alg=none หรือสลับไปใช้ asymmetric)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(401, gin.H{"success": false, "message": "invalid token"})
			c.Abort()
			return
		}

		// CRIT-01: token stateless ตรวจ signature+exp ผ่านแล้วยังไม่พอ —
		// ต้องกลับไปถาม DB ว่า user ยังมีอยู่ / ยัง active / token ยังไม่ถูกเพิกถอน
		// อ่าน role สดจาก DB ด้วย (เผื่อ role เปลี่ยนหลังออก token)
		var user model.User
		err = database.DB.
			Select("id", "role", "is_active", "token_version").
			First(&user, claims.UserID).Error

		if err != nil {
			// แยก "ไม่พบ user" (ถูกลบ → เพิกถอน 401) ออกจาก DB error จริง (→ 500 ไม่ใช่ 401 หลอก)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(401, gin.H{"success": false, "message": "session หมดอายุ กรุณาเข้าสู่ระบบใหม่"})
			} else {
				c.JSON(500, gin.H{"success": false, "message": "เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง"})
			}
			c.Abort()
			return
		}

		// บัญชีถูกปิดใช้งาน หรือ token ถูกเพิกถอน (bump token_version) → เตะออกทันที
		if !user.IsActive || user.TokenVersion != claims.TokenVersion {
			c.JSON(401, gin.H{"success": false, "message": "session หมดอายุ กรุณาเข้าสู่ระบบใหม่"})
			c.Abort()
			return
		}

		// เก็บข้อมูลไว้ใช้ต่อ — ใช้ role สดจาก DB ไม่ใช่ค่าใน token
		c.Set("userID", user.ID)
		c.Set("role", user.Role)

		c.Next()
	}
}

func StaffOnly() gin.HandlerFunc {
	return func(c *gin.Context) {

		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(403, gin.H{"message": "คุณไม่มีสิทธิ์ในการเข้าถึง"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func EmployeeAndAdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {

		role, exists := c.Get("role")
		if !exists || role != "employee" && role != "admin" {
			c.JSON(403, gin.H{"message": "คุณไม่มีสิทธิ์ในการเข้าถึง"})
			c.Abort()
			return
		}

		c.Next()
	}
}
