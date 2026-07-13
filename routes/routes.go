package routes

import (
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "hospitalbackend/docs" // เอกสาร Swagger ที่ generate จาก `swag init`
)

func SetupRouter() *gin.Engine {

	r := gin.Default()

	// Swagger UI ที่ /swagger/index.html — เปิดเฉพาะตอนตั้ง ENABLE_SWAGGER=true ใน .env
	// ปิดโดย default กันเผยโครงสร้าง API ทั้งหมดบน production
	if os.Getenv("ENABLE_SWAGGER") == "true" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
	r.Static("/uploads", "./uploads") // ให้ Gin เสิร์ฟไฟล์จากโฟลเดอร์ uploads

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
	return r
}
