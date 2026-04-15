package utilities

import (
	"log"
	"net/http"

	"maincore_go/config"

	"github.com/gin-gonic/gin"
)

type ResponsePayload struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   interface{} `json:"error,omitempty"`
}

func OK(c *gin.Context, data interface{}, message string) {
	if message == "" {
		message = "Success"
	}
	c.JSON(http.StatusOK, ResponsePayload{
		Status:  http.StatusOK,
		Message: message,
		Data:    data,
	})
}

func Created(c *gin.Context, data interface{}, message string) {
	if message == "" {
		message = "Resource created"
	}
	c.JSON(http.StatusCreated, ResponsePayload{
		Status:  http.StatusCreated,
		Message: message,
		Data:    data,
	})
}

func BadRequest(c *gin.Context, message string, data interface{}) {
	if message == "" {
		message = "Bad request"
	}
	c.JSON(http.StatusBadRequest, ResponsePayload{
		Status:  http.StatusBadRequest,
		Message: message,
		Data:    data,
	})
}

func ValidateError(c *gin.Context, data interface{}) {
	c.JSON(http.StatusBadRequest, ResponsePayload{
		Status:  http.StatusBadRequest,
		Message: "Bad request",
		Data:    data,
	})
}

func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	c.JSON(http.StatusUnauthorized, ResponsePayload{
		Status:  http.StatusUnauthorized,
		Message: message,
		Data:    nil,
	})
}

func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "Forbidden"
	}
	c.JSON(http.StatusForbidden, ResponsePayload{
		Status:  http.StatusForbidden,
		Message: message,
		Data:    nil,
	})
}

func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "Data not found"
	}
	c.JSON(http.StatusNotFound, ResponsePayload{
		Status:  http.StatusNotFound,
		Message: message,
		Data:    nil,
	})
}

func ServerError(c *gin.Context, err error, message string) {
	if message == "" {
		message = "Internal server error"
	}
	log.Printf("Internal server error: %v", err)
	
	var errorDetail interface{} = nil
	if config.AppConfig.AppEnv == "development" && err != nil {
		errorDetail = err.Error()
	}

	c.JSON(http.StatusInternalServerError, ResponsePayload{
		Status:  http.StatusInternalServerError,
		Message: message,
		Error:   errorDetail,
		Data:    nil,
	})
}

func OtherResponse(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, ResponsePayload{
		Status:  status,
		Message: message,
		Data:    data,
	})
}
