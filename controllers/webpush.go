package controllers

import (
	"maincore_go/config"
	"maincore_go/models"
	"maincore_go/utilities"

	"github.com/gin-gonic/gin"
)

type SubscribeInput struct {
	Endpoint string `json:"endpoint" binding:"required"`
	Keys     struct {
		P256dh string `json:"p256dh" binding:"required"`
		Auth   string `json:"auth" binding:"required"`
	} `json:"keys" binding:"required"`
}

type UnsubscribeInput struct {
	Endpoint string `json:"endpoint" binding:"required"`
}

func SubscribeToWebPush(c *gin.Context) {
	var input SubscribeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utilities.BadRequest(c, "Invalid input format", err.Error())
		return
	}

	userValue, exists := c.Get("user")
	if !exists {
		utilities.Unauthorized(c, "Unauthorized")
		return
	}
	userLogin := userValue.(*utilities.JwtPayload)

	sub := models.WebPushSubscription{
		UserID:   userLogin.ID,
		Endpoint: input.Endpoint,
		P256dh:   input.Keys.P256dh,
		Auth:     input.Keys.Auth,
	}

	// FirstOrCreate
	var existing models.WebPushSubscription
	if err := config.DB.Where("endpoint = ?", input.Endpoint).First(&existing).Error; err != nil {
		config.DB.Create(&sub)
	} else {
		existing.UserID = userLogin.ID
		existing.P256dh = input.Keys.P256dh
		existing.Auth = input.Keys.Auth
		config.DB.Save(&existing)
	}

	utilities.Created(c, nil, "Successfully subscribed to web push")
}

func UnsubscribeFromWebPush(c *gin.Context) {
	var input UnsubscribeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utilities.BadRequest(c, "Invalid input", err.Error())
		return
	}

	userValue, exists := c.Get("user")
	if !exists {
		utilities.Unauthorized(c, "Unauthorized")
		return
	}
	userLogin := userValue.(*utilities.JwtPayload)

	if err := config.DB.Where("user_id = ? AND endpoint = ?", userLogin.ID, input.Endpoint).Delete(&models.WebPushSubscription{}).Error; err != nil {
		utilities.ServerError(c, err, "Failed to unsubscribe")
		return
	}

	utilities.OK(c, nil, "Successfully unsubscribed")
}
