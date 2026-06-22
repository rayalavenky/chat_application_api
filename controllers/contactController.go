package controllers

import (
	"chat_application_api/services"
	"chat_application_api/utilis"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetContacts(c *gin.Context) {
	userID := c.Param("userId")

	pagination := utilis.GetPagination(c)

	contacts, totalRecords, err := services.GetContacts(userID, pagination)

	if err != nil {
		utilis.Error(c, http.StatusInternalServerError, "failed to fetch contacts")
		return
	}

	utilis.Success(c, http.StatusOK, "Contacts fetched successfully", contacts, totalRecords)
}

func GetOnlineContacts(c *gin.Context) {
	userID := c.Param("userId")

	pagination := utilis.GetPagination(c)

	contacts, totalRecords, err := services.GetOnlineContacts(userID, pagination)

	if err != nil {
		utilis.Error(c, http.StatusInternalServerError, "failed to fetch online contacts")
		return
	}
	utilis.Success(c, http.StatusOK, "Online contacts fetched successfully", contacts, totalRecords)
}
