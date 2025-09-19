package main

import (
	"repowipe/config"
	"repowipe/routes"
	"time"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)


func main(){
	config.InitEnvVar()
	config.InitRedis()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // frontend origin
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	routes.Router(r)
	r.Run(":8080")
}