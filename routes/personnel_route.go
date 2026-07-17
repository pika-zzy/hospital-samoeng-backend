package routes

import (
	controllers_personnel "hospitalbackend/controllers"
	"hospitalbackend/middleware"

	"github.com/gin-gonic/gin"
)

func PersonnelRoutes(r *gin.Engine) {
	personnelGroup := r.Group("/personnel")
	{
		personnelGroup.GET("", controllers_personnel.GetAllPersonnel)
		personnelGroup.GET("/:id", controllers_personnel.GetPersonnelByID)
		personnelGroup.POST("",
			middleware.AuthMiddleware(),       //อันนี้เช็ค login
			middleware.EmployeeAndAdminOnly(), //อันนี้เช็ค rold adminonly
			controllers_personnel.AddNewPersonnal,
		)
		personnelGroup.PUT("/:id",
			middleware.AuthMiddleware(),
			middleware.EmployeeAndAdminOnly(),
			controllers_personnel.UpdatePersonnel,
		)
		personnelGroup.DELETE("/:id",
			middleware.AuthMiddleware(),
			middleware.EmployeeAndAdminOnly(),
			controllers_personnel.DeletePersonnel,
		)
	}
}
