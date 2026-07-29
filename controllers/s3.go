package controllers

import (
	"fmt"
	"time"

	"maincore_go/services"
	"maincore_go/utilities"

	"github.com/gin-gonic/gin"
)

type PresignedUploadURLRequest struct {
	Filename string `json:"filename" form:"filename" binding:"required"`
	Path     string `json:"path" form:"path"`
	Expiry   int    `json:"expiry" form:"expiry"` // Expiry in seconds, optional
}

// GetPresignedUploadURL handles generating presigned S3 upload URLs
func GetPresignedUploadURL(c *gin.Context) {
	var req PresignedUploadURLRequest
	if err := c.ShouldBind(&req); err != nil {
		utilities.BadRequest(c, "Filename is required", err.Error())
		return
	}

	key := req.Filename
	if req.Path != "" {
		key = fmt.Sprintf("%s/%s", req.Path, req.Filename)
	}

	var lifetime time.Duration
	if req.Expiry > 0 {
		lifetime = time.Duration(req.Expiry) * time.Second
	} else {
		lifetime = 15 * time.Minute
	}

	url, err := services.GeneratePresignedUploadURL(c.Request.Context(), key, lifetime)
	if err != nil {
		utilities.ServerError(c, err, "Failed to generate presigned upload URL")
		return
	}

	utilities.OK(c, gin.H{
		"upload_url": url,
		"key":        key,
		"expires_in": int(lifetime.Seconds()),
	}, "Presigned upload URL generated successfully")
}
