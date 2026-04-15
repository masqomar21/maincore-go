package routes

import (
	"maincore_go/controllers"
	"maincore_go/middlewares"

	"github.com/gin-gonic/gin"
)

func MasterRoutes(router *gin.RouterGroup) {
	// Protected by Auth & Permissions
	master := router.Group("/")
	master.Use(middlewares.AuthMiddleware())
	master.Use(middlewares.GeneratePermissionList())

	// Users
	users := master.Group("/master/user")
	{
		users.GET("/", middlewares.RequirePermission("manage_users", "canRead"), controllers.ListUsers)
		users.POST("/", middlewares.RequirePermission("manage_users", "canWrite"), controllers.CreateUser)
		users.GET("/:id", middlewares.RequirePermission("manage_users", "canRead"), controllers.GetUser)
		users.DELETE("/:id", middlewares.RequirePermission("manage_users", "canDelete"), controllers.DeleteUser)
	}

	// Roles
	roles := master.Group("/master/role")
	{
		roles.GET("/", middlewares.RequirePermission("manage_roles", "canRead"), controllers.ListRoles)
	}

	// Notifications
	notifs := master.Group("/notification")
	{
		notifs.GET("/", controllers.ListNotifications)
		notifs.PUT("/read/:id", controllers.ReadNotification)
	}

	// Logs
	logs := master.Group("/log")
	{
		logs.GET("/", middlewares.RequirePermission("manage_logs", "canRead"), controllers.ListLogs)
	// Web Push
	webPush := master.Group("/web-push")
	{
		webPush.POST("/subscribe", controllers.SubscribeToWebPush)
		webPush.POST("/unsubscribe", controllers.UnsubscribeFromWebPush)
	}
}
}
