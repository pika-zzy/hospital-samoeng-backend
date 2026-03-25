package middleware

import (
	"fmt"
	"hospitalbackend/utils"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(401, gin.H{"message": "missing token"})
			c.Abort()
			return
		}

		// ตัด Bearer ออก
		tokenString := authHeader[len("Bearer "):]

		claims := &utils.Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(401, gin.H{"message": "invalid token"})
			c.Abort()
			return
		}

		// เก็บข้อมูลไว้ใช้ต่อ
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		startTime := time.Now()

		fmt.Printf("URL = %s, Method = %s , Time = %s\n", c.Request.URL.Path, c.Request.Method, startTime)

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
