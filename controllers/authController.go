package controllers

import (
	"log"
	"net/http"

	"chat_application_api/models"
	"chat_application_api/services"
	"chat_application_api/utilis"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": utilis.FormatValidationError(err),
		})
		return
	}

	user, err := services.Register(req)
	if err != nil {
		switch err.Error() {
		case "email already exists", "phone number already exists":
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			log.Printf("register error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful. A temporary password has been sent to your email.",
		"user":    user,
	})
}

func RefreshToken(c *gin.Context) {
	var req models.RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": utilis.FormatValidationError(err),
		})
		return
	}

	tokens, err := services.RefreshToken(req)
	if err != nil {
		switch err.Error() {
		case "invalid or expired refresh token", "refresh token not found or already revoked":
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
	})
}
