package main

import (
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

	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "DELETE", "PUT", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))


	// Serve static assets (JS, CSS, images, etc.)
	r.Static("/assets", "./static/assets")
	
	// Serve favicon
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	// Catch-all handler for React Router
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		
		// If it's an API request, return 404
		if strings.HasPrefix(path, "/api") {
			c.JSON(404, gin.H{"error": "API endpoint not found"})
			return
		}
		
		// If it's requesting a static file that doesn't exist, return 404
		if strings.Contains(path, ".") && !strings.HasSuffix(path, "/") {
			c.String(404, "File not found")
			return
		}
		
		// Otherwise, serve the React app
		c.File("./static/index.html")
	})

	routes.Router(r)


	r.Run(":8080")
}