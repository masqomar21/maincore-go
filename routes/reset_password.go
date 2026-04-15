package routes

import (
	"maincore_go/controllers"

	"github.com/gin-gonic/gin"
)

func ResetPasswordRoutes(router *gin.RouterGroup) {
	reset := router.Group("/reset-password")
	{
		reset.POST("/verify-email", controllers.SearchEmail)
		reset.POST("/verify-otp", controllers.VerifyOtp)
		reset.PUT("/change-password", controllers.ResetPassword)
	}
}
