package controllers

import (
	"chat_application_api/services"
	"chat_application_api/utilis"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetContacts(c *gin.Context) {
	userID := c.Param("userId")

	contacts, err := services.GetContacts(userID)

	if err != nil {
		utilis.Error(c, http.StatusInternalServerError, "failed to fetch contacts")
		return
	}

	utilis.Success(c, http.StatusOK, "Contacts fetched successfully", contacts)
}
