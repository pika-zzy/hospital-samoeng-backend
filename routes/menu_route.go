package routes

import (
	controllers_menu "hospitalbackend/controllers"
	"hospitalbackend/middleware"

	"github.com/gin-gonic/gin"
)

func MenuRoutes(r *gin.Engine) {
	menuGroup := r.Group("/menu")
	{
		menuGroup.GET("", controllers_menu.GetAllMenu)
		menuGroup.GET("/:id", controllers_menu.GetMenuByID)
		menuGroup.POST("",
			middleware.AuthMiddleware(),       //อันนี้เช็ค login
			middleware.EmployeeAndAdminOnly(), //อันนี้เช็ค role adminonly
			controllers_menu.AddNewMenu,
		)
		menuGroup.PUT("/:id",
			middleware.AuthMiddleware(),
			middleware.EmployeeAndAdminOnly(),
			controllers_menu.UpdateMenu,
		)
		menuGroup.DELETE("/:id",
			middleware.AuthMiddleware(),
			middleware.EmployeeAndAdminOnly(),
			controllers_menu.DeleteMenu,
		)
	}
}
