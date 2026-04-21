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
		utilis.Error(c, http.StatusBadRequest, utilis.FirstValidationMessage(err))
		return
	}

	user, err := services.Register(req)
	if err != nil {
		switch err.Error() {
		case "email already exists", "phone number already exists":
			utilis.Error(c, http.StatusConflict, err.Error())
		default:
			log.Printf("register error: %v", err)
			utilis.Error(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	utilis.Success(c, http.StatusCreated,
		"Registration successful. A temporary password has been sent to your email.",
		user,
	)
}

func RefreshToken(c *gin.Context) {
	var req models.RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utilis.Error(c, http.StatusBadRequest, utilis.FirstValidationMessage(err))
		return
	}

	tokens, err := services.RefreshToken(req)
	if err != nil {
		switch err.Error() {
		case "invalid or expired refresh token", "refresh token not found or already revoked":
			utilis.Error(c, http.StatusUnauthorized, err.Error())
		default:
			log.Printf("refresh error: %v", err)
			utilis.Error(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	utilis.Success(c, http.StatusOK, "Tokens refreshed", gin.H{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
	})
}

func Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utilis.Error(c, http.StatusBadRequest, utilis.FirstValidationMessage(err))
		return
	}

	loginResp, err := services.Login(req)
	if err != nil {
		switch err.Error() {
		case "invalid email or password":
			utilis.Error(c, http.StatusUnauthorized, err.Error())
		default:
			log.Printf("login error: %v", err)
			utilis.Error(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	utilis.Success(c, http.StatusOK, "Login Successful", loginResp)
}

func Logout(c *gin.Context) {
	userIDVal, exists := c.Get("userId")
	if !exists {
		utilis.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	userID, ok := userIDVal.(string)
	if !ok {
		utilis.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := services.Logout(userID); err != nil {
		log.Printf("logout error: %v", err)
		utilis.Error(c, http.StatusInternalServerError, "internal server error")
		return
	}

	utilis.Success(c, http.StatusOK, "Logout successful", nil)
}
