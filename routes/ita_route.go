package routes

import (
	controllers "hospitalbackend/controllers"
	"hospitalbackend/middleware"

	"github.com/gin-gonic/gin"
)

func ITARoutes(r *gin.Engine) {

	auth := middleware.AuthMiddleware()
	admin := middleware.EmployeeAndAdminOnly()

	// ==================== ITA ====================
	ita := r.Group("/ita")
	{
		// YEAR
		ita.GET("/years", controllers.GetYears)
		ita.POST("/years", auth, admin, controllers.CreateYear)

		// MOIT (ผูกกับปี)
		ita.GET("/years/:id/moit", controllers.GetMoitByYear)
		ita.POST("/years/:id/moit", auth, admin, controllers.CreateMoit)

		// FILE
		ita.GET("", controllers.GetAllITA)
		ita.POST("/upload", auth, admin, controllers.UploadITA)
	}
}
