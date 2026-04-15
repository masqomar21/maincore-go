package middlewares

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"maincore_go/config"
	"maincore_go/models"
	"maincore_go/utilities"

	"github.com/gin-gonic/gin"
)

type GeneratedPermissionList struct {
	Permission string `json:"permission"`
	CanRead    bool   `json:"canRead"`
	CanWrite   bool   `json:"canWrite"`
	CanUpdate  bool   `json:"canUpdate"`
	CanDelete  bool   `json:"canDelete"`
	CanRestore bool   `json:"canRestore"`
}

func GeneratePermissionList() gin.HandlerFunc {
	return func(c *gin.Context) {
		userValue, exists := c.Get("user")
		if !exists {
			utilities.Unauthorized(c, "Unauthorized - No user context found")
			c.Abort()
			return
		}

		userLogin := userValue.(*utilities.JwtPayload)

		key := fmt.Sprintf("user_permissions:%d", userLogin.ID)
		ctx := context.Background()

		// Try to GET from Redis
		cacheData, err := config.RedisClient.Get(ctx, key).Result()
		var permissionList []GeneratedPermissionList

		if err == nil && cacheData != "" {
			err = json.Unmarshal([]byte(cacheData), &permissionList)
			if err == nil {
				c.Set("permissionList", permissionList)
				c.Next()
				return
			}
		}

		// Pull from Database if Redis fails or empty
		var user models.User
		if err := config.DB.Preload("Role.RolePermissions.Permission").First(&user, userLogin.ID).Error; err != nil {
			utilities.Forbidden(c, "No permissions found")
			c.Abort()
			return
		}

		for _, rp := range user.Role.RolePermissions {
			permissionList = append(permissionList, GeneratedPermissionList{
				Permission: rp.Permission.Name,
				CanRead:    rp.CanRead,
				CanWrite:   rp.CanWrite,
				CanUpdate:  rp.CanUpdate,
				CanDelete:  rp.CanDelete,
				CanRestore: rp.CanRestore,
			})
		}

		// Serialize and store into Redis
		bytesData, err := json.Marshal(permissionList)
		if err == nil {
			config.RedisClient.Set(ctx, key, bytesData, time.Hour) // 1 Hour Cache
		} else {
			log.Printf("Failed to marshal permission list: %v", err)
		}

		c.Set("permissionList", permissionList)
		c.Next()
	}
}

func RequirePermission(permissionName string, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userValue, exists := c.Get("user")
		if !exists {
			utilities.Unauthorized(c, "Unauthorized")
			c.Abort()
			return
		}

		userLogin := userValue.(*utilities.JwtPayload)

		// Allow SUPER_ADMIN all privileges
		if userLogin.RoleType == "SUPER_ADMIN" {
			c.Next()
			return
		}

		permListValue, exists := c.Get("permissionList")
		if !exists {
			utilities.Forbidden(c, "No permissions loaded")
			c.Abort()
			return
		}

		permissionList := permListValue.([]GeneratedPermissionList)
		hasPermission := false

		for _, p := range permissionList {
			if p.Permission == permissionName {
				switch action {
				case "all":
					hasPermission = true
				case "canRead":
					hasPermission = p.CanRead
				case "canWrite":
					hasPermission = p.CanWrite
				case "canUpdate":
					hasPermission = p.CanUpdate
				case "canDelete":
					hasPermission = p.CanDelete
				case "canRestore":
					hasPermission = p.CanRestore
				}
			}
			if hasPermission {
				break
			}
		}

		if !hasPermission {
			msg := fmt.Sprintf("Forbidden - You do not have permission to %s %s", action, permissionName)
			utilities.Forbidden(c, msg)
			c.Abort()
			return
		}

		c.Next()
	}
}
