package routes

import (
	"hospitalbackend/controllers"
	"hospitalbackend/middleware"

	"github.com/gin-gonic/gin"
)

// ContentRoutes — หน้าเนื้อหาที่แก้ได้จากหลังบ้าน (ความปลอดภัยด้านยา / ชมรมจริยธรรม / PDPA)
//
// อ่านได้ทุกคน (หน้าเว็บสาธารณะ) — เพิ่ม/แก้/ลบ ต้อง login (employee หรือ admin)
// เหมือน personnel/news/activity
//
// section อ้างด้วย slug ส่วน group/file อ้างด้วย id — แยก path prefix กัน
// (/content/sections, /content/groups, /content/files) จึงไม่ชน wildcard ของ Gin
func ContentRoutes(r *gin.Engine) {
	content := r.Group("/content")
	{
		content.GET("/sections", controllers.GetContentSections)
		content.GET("/sections/:slug", controllers.GetContentSection)

		auth := content.Group("")
		auth.Use(middleware.AuthMiddleware(), middleware.EmployeeAndAdminOnly())
		{
			auth.POST("/sections", controllers.CreateContentSection)
			auth.PUT("/sections/:slug", controllers.UpdateContentSection)
			auth.DELETE("/sections/:slug", controllers.DeleteContentSection)

			auth.POST("/sections/:slug/groups", controllers.CreateContentGroup)
			auth.PUT("/groups/:id", controllers.UpdateContentGroup)
			auth.DELETE("/groups/:id", controllers.DeleteContentGroup)

			auth.POST("/groups/:id/files", controllers.AddContentFiles)
			auth.PUT("/files/:id", controllers.UpdateContentFile)
			auth.DELETE("/files/:id", controllers.DeleteContentFile)
		}
	}
}
