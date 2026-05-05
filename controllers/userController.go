package controllers

import (
	"chat_application_api/services"
	"chat_application_api/utilis"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	users, err := services.GetUsers()
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
