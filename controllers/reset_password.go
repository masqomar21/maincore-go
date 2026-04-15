package controllers

import (
	"fmt"
	"math/rand"
	"time"

	"maincore_go/config"
	"maincore_go/models"
	"maincore_go/utilities"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type VerifyEmailInput struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyOtpInput struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

type ResetPasswordInput struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

func generateOTP() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%04d", r.Intn(10000))
}

func SearchEmail(c *gin.Context) {
	var input VerifyEmailInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utilities.BadRequest(c, "Invalid input", err.Error())
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		utilities.NotFound(c, "Email not found")
		return
	}

	// Delete old OTP if any
	config.DB.Where("user_id = ? AND purpose = ?", user.ID, models.OtpPurposeResetPassword).Delete(&models.Otp{})

	code := generateOTP()
	otp := models.Otp{
		UserID:    user.ID,
		Code:      code,
		Purpose:   models.OtpPurposeResetPassword,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	if err := config.DB.Create(&otp).Error; err != nil {
		utilities.ServerError(c, err, "Failed to create OTP")
		return
	}

	// In real application, we would send this via email/SMTP here.
	utilities.OK(c, nil, "OTP sent to email (mocked)")
}

func VerifyOtp(c *gin.Context) {
	var input VerifyOtpInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utilities.BadRequest(c, "Invalid input", err.Error())
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		utilities.NotFound(c, "User not found")
		return
	}

	var otp models.Otp
	if err := config.DB.Where("user_id = ? AND code = ? AND purpose = ?", user.ID, input.Code, models.OtpPurposeResetPassword).First(&otp).Error; err != nil {
		utilities.BadRequest(c, "Invalid OTP", nil)
		return
	}

	if time.Now().After(otp.ExpiresAt) {
		utilities.BadRequest(c, "OTP expired", nil)
		return
	}

	// If valid, we issue a special token for actually resetting the password
	payload := utilities.JwtPayload{
		ID:       user.ID,
		Name:     *user.Name,
		RoleType: "OTHER", // placeholder
		Purpose:  "RESET_PASSWORD",
	}

	token, _ := utilities.GenerateAccessToken(payload, 15*time.Minute)

	// Clean up OTP 
	config.DB.Delete(&otp)

	utilities.OK(c, gin.H{"token": token}, "OTP verified")
}

func ResetPassword(c *gin.Context) {
	var input ResetPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utilities.BadRequest(c, "Invalid input", err.Error())
		return
	}

	decode, err := utilities.VerifyAccessToken(input.Token)
	if err != nil || decode.Purpose != "RESET_PASSWORD" {
		utilities.Unauthorized(c, "Invalid or expired token")
		return
	}

	var user models.User
	if err := config.DB.Where("id = ?", decode.ID).First(&user).Error; err != nil {
		utilities.NotFound(c, "User not found")
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	hashedPassword := string(hash)

	config.DB.Model(&user).Update("password", &hashedPassword)

	config.DB.Create(&models.Logger{
		UserID:  user.ID,
		Process: models.ProcessUpdate,
		Detail:  "User reset password",
	})

	utilities.OK(c, nil, "Password successfully reset")
}
