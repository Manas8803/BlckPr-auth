package routes

import (
	controller "auth-service/main-app/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoute(router *gin.RouterGroup) {
	router.POST("/auth/login", controller.Login)
	router.POST("/auth/register", controller.Register)
	router.POST("/auth/otp", controller.ValidateOTP)
}
