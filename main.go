package main

import (
	"log"
	"os"

	"chat_application_api/config"
	"chat_application_api/middleware"
	"chat_application_api/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDB()

	router := gin.Default()

	router.Use(middleware.CORSMiddleware())

	routes.SetupRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	log.Fatal(router.Run(":" + port))
}
