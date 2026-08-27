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
		activityGroup.PUT("/:id",
			middleware.AuthMiddleware(),
			middleware.EmployeeAndAdminOnly(),
			activity_controller.UpdateActivity,
		)
		activityGroup.DELETE("/:id",
			middleware.AuthMiddleware(),
			middleware.EmployeeAndAdminOnly(),
			activity_controller.DeleteActivity,
		)

		// อัลบั้มรูปของกิจกรรม — รูปปกยังจัดการที่ POST/PUT /activities เหมือนเดิม
		// เพดานรวม (ปก + อัลบั้ม) = model.MaxActivityImages บังคับใน controller
		activityGroup.POST("/:id/images",
			middleware.AuthMiddleware(),
			middleware.EmployeeAndAdminOnly(),
			activity_controller.AddActivityImages,
		)
		activityGroup.DELETE("/:id/images/:imageId",
			middleware.AuthMiddleware(),
			middleware.EmployeeAndAdminOnly(),
			activity_controller.DeleteActivityImage,
		)
	}

}
