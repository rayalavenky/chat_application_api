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

	users, err := services.GetUsers(search)
	if err != nil {
		utilis.Error(c, http.StatusInternalServerError, "internal server error")
		return
	}
	utilis.Success(c, http.StatusOK, "Users fetched successfully", users)
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

	utilis.Success(c, http.StatusOK, "User fetched successfully", user)
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

	utilis.Success(c, http.StatusOK, "Profile updated successfully", user)
}

func SendRequest(c *gin.Context) {
	var req models.UserRequest

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		utilis.Error(c, http.StatusBadRequest, utilis.FirstValidationMessage(err))
		return
	}
	err := services.SendRequest(req)

	if err != nil {
		utilis.Error(c, http.StatusInternalServerError, "failed to send request")
		return
	}

	utilis.Success(c, http.StatusOK, "Request sent successfully", nil)
}

func GetReceivedRequests(c *gin.Context) {
	userID := c.Param("userId")

	requests, err := services.GetReceivedRequests(userID)

	if err != nil {
		utilis.Error(c, http.StatusInternalServerError, "failed to fetch received requests")
		return
	}

	utilis.Success(c, http.StatusOK, "Received requests fetched successfully", requests)

}

func AcceptRequest(c *gin.Context) {
	requestID := c.Param("requestId")

	err := services.AcceptRequest(requestID)

	if err != nil {
		utilis.Error(c, http.StatusInternalServerError, "failed to accept request")
		return
	}

	utilis.Success(c, http.StatusOK, "Request accepted successfully", nil)
}

func GetContacts(c *gin.Context) {
	userID := c.Param("userId")

	contacts, err := services.GetContacts(userID)

	if err != nil {
		utilis.Error(c, http.StatusInternalServerError, "failed to fetch contacts")
		return
	}

	utilis.Success(c, http.StatusOK, "Contacts fetched successfully", contacts)
}
