package routes

import (
	controllers "hospitalbackend/controllers"
	"hospitalbackend/middleware"

	"github.com/gin-gonic/gin"
)

func MOITRoutes(r *gin.Engine) {

	auth := middleware.AuthMiddleware()
	admin := middleware.EmployeeAndAdminOnly()

	// ==================== MOIT ====================
	moit := r.Group("/moit")
	{
		moit.GET("", controllers.GetAllMoit)
		// ดู topic ของ moit
		moit.GET("/:id/topics", controllers.GetMoitTopics)

		// แก้ไข / ลบ moit
		moit.PUT("/:id", auth, admin, controllers.UpdateMoit)
		moit.DELETE("/:id", auth, admin, controllers.DeleteMoit)

		// สร้าง topic
		moit.POST("/:id/topics", auth, admin, controllers.CreateTopic)
	}

	// ==================== TOPIC ====================
	topic := r.Group("/topics")
	{
		topic.PUT("/:id", auth, admin, controllers.UpdateTopic)
		topic.DELETE("/:id", auth, admin, controllers.DeleteTopic)

		// สร้าง item
		topic.POST("/:id/items", auth, admin, controllers.CreateItem)
	}

	// ==================== ITEM ====================
	item := r.Group("/items")
	{
		item.PUT("/:id", auth, admin, controllers.UpdateItem)
		item.DELETE("/:id", auth, admin, controllers.DeleteItem)
	}
}
