package main

import (
	"log"
	"os"
	"repowipe/config"
	"repowipe/routes"
	"repowipe/services"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	_ "repowipe/providers/bitbucket"
	_ "repowipe/providers/github"
	_ "repowipe/providers/gitlab"
)

func main() {
	config.InitEnvVar()
	config.InitRedis()
	services.StartTransferWorker()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	appEnv := os.Getenv("APP_ENV")
	log.Println("port=", port)
	log.Println("app_env=", appEnv)

	r := gin.Default()

	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	allowedOrigins := []string{"http://localhost:3000"}
	if allowedOriginsEnv != "" {
		allowedOrigins = strings.Split(allowedOriginsEnv, ",")
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "DELETE", "PUT", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "endpoint not found"})
	})

	routes.Router(r)

	r.Run(":" + port)
}
