package routes

import (
	"maincore_go/controllers"
	"maincore_go/middlewares"

	"github.com/gin-gonic/gin"
)

func S3Routes(router *gin.RouterGroup) {
	s3Group := router.Group("/s3")
	s3Group.Use(middlewares.AuthMiddleware())
	{
		s3Group.POST("/presigned-upload-url", controllers.GetPresignedUploadURL)
		s3Group.GET("/presigned-upload-url", controllers.GetPresignedUploadURL)
	}
}
