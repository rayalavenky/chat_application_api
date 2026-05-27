package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"chat_application_api/config"
	"chat_application_api/middleware"
	"chat_application_api/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDB()

	config.ConnectRedis()

	router := gin.Default()

	router.Use(middleware.CORSMiddleware())

	routes.SetupRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	rand.Seed(time.Now().UnixNano())
	log.Printf("Server running on port %s", port)
	log.Fatal(router.Run(":" + port))
}
