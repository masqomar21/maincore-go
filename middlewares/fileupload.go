package middlewares

import (
	"strings"

	"maincore_go/utilities"

	"github.com/gin-gonic/gin"
)

func FileUploadMiddleware(maxFileSize int64, allowedFileTypes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Just validate Content-Length early if possible
		if c.Request.ContentLength > maxFileSize {
			utilities.BadRequest(c, "File too large", nil)
			c.Abort()
			return
		}

		// When reading multiple parts, the user's controller will invoke c.FormFile...
		// But in this middleware, we can just attach the configuration into Context
		// or pre-parse. Gin handles file parsing very well inside the controller.
		// However, to mimic Express Multer `fileFilter`, we could parse multipart right here.

		err := c.Request.ParseMultipartForm(maxFileSize)
		if err != nil {
			if strings.Contains(err.Error(), "request body too large") {
				utilities.BadRequest(c, "File too large", nil)
			} else {
				// might not be a multipart request, ignoring or returning error?
			}
			// In Gin, you often parse the file directly in the controller (Bind).
			// We will just let the controller handle getting the file, but we'll set rules here.
		}

		c.Set("maxFileSize", maxFileSize)
		c.Set("allowedFileTypes", allowedFileTypes)
		c.Next()
	}
}

// CheckFileType is a helper for Controllers to use
func CheckFileType(mime string, allowedFileTypes []string) bool {
	for _, a := range allowedFileTypes {
		if strings.EqualFold(a, mime) {
			return true
		}
	}
	return false
}
