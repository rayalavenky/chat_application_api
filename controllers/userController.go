package controllers

import (
	"chat_application_api/models"
	"chat_application_api/services"
	"chat_application_api/utilis"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	searchParam := c.Query("search")

	var search models.UserSearch
	if searchParam != "" {
		if err := json.Unmarshal([]byte(searchParam), &search); err != nil {
			utilis.Error(c, http.StatusBadRequest, "invalid search parameter")
			return
		}
	}

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

	pagination := utilis.GetPagination(c)

	users, totalRecords, err := services.GetUsers(search, userID, pagination)
	if err != nil {
		utilis.Error(c, http.StatusInternalServerError, "internal server error")
		return
	}
	utilis.Success(c, http.StatusOK, "Users fetched successfully", users, totalRecords)
}

func GetUserByID(c *gin.Context) {
	userID := c.Param("id")

	user, err := services.GetUserByID(userID)
	if err != nil {
		switch err.Error() {
		case "user not found":
			utilis.Error(c, http.StatusNotFound, "user not found")
		default:
			utilis.Error(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	utilis.Success(c, http.StatusOK, "User fetched successfully", user, 0)
}

func UpdateProfile(c *gin.Context) {
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

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utilis.Error(c, http.StatusBadRequest, utilis.FirstValidationMessage(err))
		return
	}

	user, err := services.UpdateProfile(userID, req)
	if err != nil {
		utilis.Error(c, http.StatusInternalServerError, "failed to update profile")
		return
	}

	utilis.Success(c, http.StatusOK, "Profile updated successfully", user, 0)
}
