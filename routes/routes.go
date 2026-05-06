package routes

import (
	"chat_application_api/controllers"
	"chat_application_api/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	api := router.Group("/api")

	auth := api.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/refresh", controllers.RefreshToken)
		auth.POST("/login", controllers.Login)
		auth.POST("/logout", middleware.AuthMiddleware(), controllers.Logout)
		auth.POST("/forgot-password", controllers.ForgotPassword)
		auth.POST("/verify-otp", controllers.VerifyOTP)
		auth.POST("/reset-password", controllers.ResetPassword)
	}

	users := api.Group("/users", middleware.AuthMiddleware())
	{
		users.GET("", controllers.GetUsers)
		users.GET("/:id", controllers.GetUserByID)
		users.PUT("/update", controllers.UpdateProfile)
	}
}
