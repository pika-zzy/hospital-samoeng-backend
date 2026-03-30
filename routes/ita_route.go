package routes

import (
	ita_controller "hospitalbackend/controllers"
	"hospitalbackend/middleware"

	"github.com/gin-gonic/gin"
)

func Itaroute(r *gin.Engine) {
	itaGroup := r.Group("/ita")
	{
		itaGroup.GET("", ita_controller.GetAllITA)
		itaGroup.GET("/:id", ita_controller.GetITAByID)
		itaGroup.POST("",
			middleware.AuthMiddleware(),       //อันนี้เช็ค login
			middleware.EmployeeAndAdminOnly(), //อันนี้เช็ค role adminonly
			ita_controller.CreateITA,
		)
	}
}
