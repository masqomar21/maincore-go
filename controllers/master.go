package controllers

import (
	"maincore_go/config"
	"maincore_go/models"
	"maincore_go/utilities"

	"github.com/gin-gonic/gin"
)

// List Users
func ListUsers(c *gin.Context) {
	var users []models.User
	if err := config.DB.Preload("Role").Find(&users).Error; err != nil {
		utilities.ServerError(c, err, "Error fetching users")
		return
	}
	utilities.OK(c, users, "Users retrieved")
}

func CreateUser(c *gin.Context) {
	// Re-uses Register type flow generally in admin dashboards
	Register(c)
}

func GetUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := config.DB.Preload("Role").First(&user, id).Error; err != nil {
		utilities.NotFound(c, "User not found")
		return
	}
	utilities.OK(c, user, "Success")
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := config.DB.Delete(&models.User{}, id).Error; err != nil {
		utilities.ServerError(c, err, "Error deleting user")
		return
	}
	utilities.OK(c, nil, "User deleted")
}

// List Roles
func ListRoles(c *gin.Context) {
	var roles []models.Role
	if err := config.DB.Preload("RolePermissions.Permission").Find(&roles).Error; err != nil {
		utilities.ServerError(c, err, "Error fetching roles")
		return
	}
	utilities.OK(c, roles, "Roles retrieved")
}

// Notifications
func ListNotifications(c *gin.Context) {
	userValue, _ := c.Get("user")
	userLogin := userValue.(*utilities.JwtPayload)

	var notifs []models.NotificationUser
	if err := config.DB.Preload("Notification").Where("user_id = ?", userLogin.ID).Find(&notifs).Error; err != nil {
		utilities.ServerError(c, err, "Error fetching notifications")
		return
	}
	utilities.OK(c, notifs, "Success")
}

func ReadNotification(c *gin.Context) {
	id := c.Param("id")
	userValue, _ := c.Get("user")
	userLogin := userValue.(*utilities.JwtPayload)

	if err := config.DB.Model(&models.NotificationUser{}).
		Where("id = ? AND user_id = ?", id, userLogin.ID).
		Update("read_status", true).Error; err != nil {
		utilities.ServerError(c, err, "Failed to update")
		return
	}
	utilities.OK(c, nil, "Notification read")
}

// Logs
func ListLogs(c *gin.Context) {
	var logs []models.Logger
	if err := config.DB.Preload("User").Order("created_at desc").Limit(100).Find(&logs).Error; err != nil {
		utilities.ServerError(c, err, "Error fetching logs")
		return
	}
	utilities.OK(c, logs, "Logs retrieved")
}
