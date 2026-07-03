package routes

import (
	controllers "hospitalbackend/controllers"
	"hospitalbackend/middleware"

	"github.com/gin-gonic/gin"
)

func PopupRoutes(r *gin.Engine) {

	auth := middleware.AuthMiddleware()
	admin := middleware.EmployeeAndAdminOnly()

	popup := r.Group("/popup")
	{
		// public — ฝั่งผู้ใช้เรียกทุกครั้งที่เข้าเว็บ
		popup.GET("", controllers.GetPopup)

		// admin — เปิด/ปิด + เปลี่ยนรูป
		popup.PATCH("/status", auth, admin, controllers.UpdatePopupStatus)
		popup.POST("/image", auth, admin, controllers.UploadPopupImage)
	}
}
