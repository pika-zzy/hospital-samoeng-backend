package routes

import (
	controllers "hospitalbackend/controllers"
	"hospitalbackend/middleware"

	"github.com/gin-gonic/gin"
)

func PermissionRoutes(r *gin.Engine) {
	roles := r.Group("/roles")
	roles.Use(middleware.AuthMiddleware())
	{
		// employee ต้องอ่านสิทธิ์ของตัวเองได้ (ใช้กรองเมนู sidebar)
		roles.GET("/:role/permissions", middleware.EmployeeAndAdminOnly(), controllers.GetRolePermissions)
		// แก้สิทธิ์ได้เฉพาะ admin
		roles.PUT("/:role/permissions", middleware.StaffOnly(), controllers.UpdateRolePermissions)
	}
}
