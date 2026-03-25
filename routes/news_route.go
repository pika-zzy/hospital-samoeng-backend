package routes

import (
	controllers_news "hospitalbackend/controllers"
	"hospitalbackend/middleware"

	"github.com/gin-gonic/gin"
)

func NewsRoutes(r *gin.Engine) {

	newsGroup := r.Group("/news")
	{
		newsGroup.GET("", controllers_news.GetAllNews)
		newsGroup.GET("/:id", controllers_news.GetNewsByID)
		newsGroup.POST("",
			middleware.AuthMiddleware(), //อันนี้เช็ค login
			middleware.StaffOnly(),      //อันนี้เช็ค role adminonly
			controllers_news.CreateNews,
		)
	}

}
