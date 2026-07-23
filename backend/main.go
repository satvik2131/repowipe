package main

import (
	"log"
	"os"
	"repowipe/config"
	"repowipe/routes"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.InitEnvVar()
	config.InitRedis()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	appEnv := os.Getenv("APP_ENV")
	log.Println("port=", port)
	log.Println("app_env=", appEnv)

	r := gin.Default()

	// Build allowed origins from env var (comma-separated)
	// e.g. ALLOWED_ORIGINS=http://localhost:3000,https://myapp.vercel.app
	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	allowedOrigins := []string{"http://localhost:3000"} // safe default
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

	// Frontend is deployed on Vercel — no static file serving needed here.
	// All unmatched routes return 404 for unknown API paths.
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "endpoint not found"})
	})

	routes.Router(r)

	r.Run(":" + port)
}

