package middlewares

import (
	"strings"

	"maincore_go/config"
	"maincore_go/models"
	"maincore_go/utilities"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utilities.Unauthorized(c, "Unauthorized - No token provided")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utilities.Unauthorized(c, "Unauthorized - Invalid token format")
			c.Abort()
			return
		}
		token := parts[1]

		// Check session in DB
		var session models.Session
		if err := config.DB.Where("token = ?", token).First(&session).Error; err != nil {
			utilities.OtherResponse(c, 498, "Unauthorized - Invalid session", nil)
			c.Abort()
			return
		}

		decode, err := utilities.VerifyAccessToken(token)
		if err != nil {
			utilities.OtherResponse(c, 498, "Unauthorized - Invalid token", nil)
			c.Abort()
			return
		}

		if decode.Purpose != "ACCESS_TOKEN" {
			utilities.OtherResponse(c, 498, "Unauthorized - Invalid token purpose", nil)
			c.Abort()
			return
		}

		// Set the user in the context
		c.Set("user", decode)
		c.Next()
	}
}
