package routes

import (
	controllers "hospitalbackend/controllers"
	"hospitalbackend/middleware"

	"github.com/gin-gonic/gin"
)

func HeroRoutes(r *gin.Engine) {

	auth := middleware.AuthMiddleware()
	admin := middleware.EmployeeAndAdminOnly()

	hero := r.Group("/hero")
	{
		// public — หน้าแรกเรียกมาแสดง carousel
		hero.GET("", controllers.GetHeroSlides)

		// admin — เพิ่มรูป / แก้ลำดับ / ลบ (ลบรูปสุดท้ายไม่ได้)
		hero.POST("", auth, admin, controllers.CreateHeroSlide)
		hero.PUT("/:id", auth, admin, controllers.UpdateHeroSlide)
		hero.DELETE("/:id", auth, admin, controllers.DeleteHeroSlide)
	}
}
