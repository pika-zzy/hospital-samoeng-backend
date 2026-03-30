package routes

import (
	activity_controller "hospitalbackend/controllers"
	"hospitalbackend/middleware"

	"github.com/gin-gonic/gin"
)

func ActivityRoutes(r *gin.Engine) {

	activityGroup := r.Group("/activities")
	{
		activityGroup.GET("", activity_controller.GetAllActivities)
		activityGroup.GET("/:id", activity_controller.GetActivityByID)
		activityGroup.POST("",
			middleware.AuthMiddleware(),       //อันนี้เช็ค login
			middleware.EmployeeAndAdminOnly(), //อันนี้เช็ค role adminonly
			activity_controller.CreateActivity,
		)
	}

}
