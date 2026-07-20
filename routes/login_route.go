package routes

import (
	controller_login "hospitalbackend/controllers"
	"hospitalbackend/middleware"

	"github.com/gin-gonic/gin"
)

func LoginRoute(r *gin.Engine) {

	r.POST("/login", controller_login.Login)

	// logout — ต้อง login อยู่ก่อน (auth) เพื่อรู้ว่าใครออก แล้ว bump token_version ของตัวเอง
	r.POST("/logout", middleware.AuthMiddleware(), controller_login.Logout)

	// จัดการผู้ใช้งาน (เฉพาะ admin)
	users := r.Group("/users")
	users.Use(middleware.AuthMiddleware(), middleware.StaffOnly())
	{
		users.GET("", controller_login.GetAllUsers)
		users.GET("/:id", controller_login.GetUserByID)
		users.POST("", controller_login.AddUser)
		users.DELETE("/:id", controller_login.DeleteUser)
		users.PATCH("/:id/status", controller_login.UpdateUserStatus)
	}
}
