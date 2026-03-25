package routes

import (
	controller_login "hospitalbackend/controllers"

	"github.com/gin-gonic/gin"
)

func LoginRoute(r *gin.Engine) {

	r.POST("/login", controller_login.Login)
}
