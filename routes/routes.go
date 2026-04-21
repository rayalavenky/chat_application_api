package routes

import (
	"chat_application_api/controllers"
	"chat_application_api/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	api := router.Group("/api")

	publicAuth := api.Group("/auth")
	{
		publicAuth.POST("/register", controllers.Register)
		publicAuth.POST("/refresh", controllers.RefreshToken)
		publicAuth.POST("/login", controllers.Login)
	}

	// Protected routes
	protectedAuth := api.Group("/auth")
	protectedAuth.Use(middleware.AuthMiddleware())
	{
		protectedAuth.POST("/logout", controllers.Logout)
	}
}
