package routes

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {

	r := gin.Default()
	r.Static("/uploads", "./uploads") // ให้ Gin เสิร์ฟไฟล์จากโฟลเดอร์ uploads

	// Middleware สำหรับ CORS (อนุญาตให้ Frontend ที่รันบนพอร์ต 5173 เข้าถึง API ได้)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
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
	Itaroute(r)
	return r
}
