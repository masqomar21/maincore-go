package controllers

import (
	"strings"
	"time"

	"maincore_go/config"
	"maincore_go/models"
	"maincore_go/utilities"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utilities.BadRequest(c, "Invalid Input", err.Error())
		return
	}

	var role models.Role
	if err := config.DB.Where("role_type = ?", models.RoleTypeOther).First(&role).Error; err != nil {
		utilities.BadRequest(c, "Role not found", nil)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		utilities.ServerError(c, err, "Error hashing password")
		return
	}
	hashedPassword := string(hash)

	user := models.User{
		Name:     &input.Name,
		Email:    input.Email,
		Password: &hashedPassword,
		RoleID:   role.ID,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		utilities.ServerError(c, err, "Error creating user")
		return
	}

	utilities.Created(c, user, "Success")
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utilities.BadRequest(c, "Invalid Input", err.Error())
		return
	}

	var user models.User
	if err := config.DB.Preload("Role").Where("email = ?", input.Email).First(&user).Error; err != nil {
		utilities.NotFound(c, "User not found")
		return
	}

	if user.Password == nil {
		utilities.Unauthorized(c, "Password not match")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(input.Password)); err != nil {
		utilities.Unauthorized(c, "Password not match")
		return
	}

	payload := utilities.JwtPayload{
		ID:       user.ID,
		Name:     *user.Name,
		Role:     user.Role.Name,
		RoleType: string(user.Role.RoleType),
		Purpose:  "ACCESS_TOKEN",
	}

	token, err := utilities.GenerateAccessToken(payload, 24*time.Hour)
	if err != nil {
		utilities.ServerError(c, err, "Error generating token")
		return
	}

	session := models.Session{
		Token:  token,
		UserID: user.ID,
	}
	config.DB.Create(&session)

	// Log Activity
	config.DB.Create(&models.Logger{
		UserID:  user.ID,
		Process: models.ProcessLogin,
		Detail:  "User login",
	})

	// Response
	c.JSON(200, utilities.ResponsePayload{
		Status:  200,
		Message: "Success",
		Data: gin.H{
			"id":                  user.ID,
			"email":               user.Email,
			"name":                user.Name,
			"registeredViaGoogle": user.RegisteredViaGoogle,
			"token":               token,
		},
	})
}

func GetUserProfile(c *gin.Context) {
	userValue, exists := c.Get("user")
	if !exists {
		utilities.Unauthorized(c, "Unauthorized")
		return
	}

	userLogin := userValue.(*utilities.JwtPayload)

	var user models.User
	if err := config.DB.Preload("Role.RolePermissions.Permission").First(&user, userLogin.ID).Error; err != nil {
		utilities.NotFound(c, "User not found")
		return
	}

	mappedPermissions := []string{}
	for _, rp := range user.Role.RolePermissions {
		if rp.CanRead {
			mappedPermissions = append(mappedPermissions, "read:"+rp.Permission.Name)
		}
		if rp.CanWrite {
			mappedPermissions = append(mappedPermissions, "write:"+rp.Permission.Name)
		}
		if rp.CanUpdate {
			mappedPermissions = append(mappedPermissions, "update:"+rp.Permission.Name)
		}
		if rp.CanRestore {
			mappedPermissions = append(mappedPermissions, "restore:"+rp.Permission.Name)
		}
		if rp.CanDelete {
			mappedPermissions = append(mappedPermissions, "delete:"+rp.Permission.Name)
		}
	}

	// Prepare exact response shape matching original JS
	resData := gin.H{
		"id":                  user.ID,
		"name":                user.Name,
		"email":               user.Email,
		"registeredViaGoogle": user.RegisteredViaGoogle,
		"role": gin.H{
			"name":            user.Role.Name,
			"roleType":        user.Role.RoleType,
			"rolePermissions": mappedPermissions,
		},
	}

	utilities.OK(c, resData, "User profile retrieved successfully")
}

func Logout(c *gin.Context) {
	userValue, _ := c.Get("user")
	userLogin := userValue.(*utilities.JwtPayload)

	authHeader := c.GetHeader("Authorization")
	parts := strings.Split(authHeader, " ")
	token := parts[1]

	config.DB.Where("user_id = ? AND token = ?", userLogin.ID, token).Delete(&models.Session{})

	config.DB.Create(&models.Logger{
		UserID:  userLogin.ID,
		Process: models.ProcessLogout,
		Detail:  "User logout",
	})

	utilities.OK(c, nil, "Success")
}
