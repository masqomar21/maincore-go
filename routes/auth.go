package routes

import (
	"maincore_go/controllers"
	"maincore_go/middlewares"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
		
		// Note: Google Auth integration will go here
	}

	profile := router.Group("/profile")
	profile.Use(middlewares.AuthMiddleware())
	profile.Use(middlewares.GeneratePermissionList())
	{
		profile.GET("/", controllers.GetUserProfile)
		profile.POST("/logout", controllers.Logout)
	}
}
