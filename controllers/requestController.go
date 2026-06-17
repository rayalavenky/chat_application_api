package controllers

import (
	"chat_application_api/models"
	"chat_application_api/services"
	"chat_application_api/utilis"
	"net/http"

	"github.com/gin-gonic/gin"
)

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

func GetSentRequests(c *gin.Context) {
	userID := c.Param("userId")

	requests, err := services.GetSentRequests(userID)

	if err != nil {
		utilis.Error(c, http.StatusInternalServerError, "failed to fetch sent requests")
		return
	}

	utilis.Success(c, http.StatusOK, "Sent requests fetched successfully", requests)
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

func RejectRequest(c *gin.Context) {
	requestID := c.Param("requestId")

	err := services.RejectRequest(requestID)
	if err != nil {
		utilis.Error(c, http.StatusInternalServerError, "failed to reject request")
		return
	}
	utilis.Success(c, http.StatusOK, "Request rejected successfully", nil)
}
