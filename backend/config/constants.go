package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)


var ClientId, ClientSecret,Redirect_Uri string
const (
	GetUserApi = "https://api.github.com/user"
 	GetRepoApi = "https://api.github.com/user/repos"
	AccessTokenUrl = "https://github.com/login/oauth/access_token"
	SearchUri = "https://api.github.com/search/repositories"
	DeleteApi = "https://api.github.com/repos/"
 )


func InitEnvVar(){
	   // Only try to load .env file in development
    // Production environments (like Railway) use environment variables directly
    if err := godotenv.Load(); err != nil {
        // Don't log error in production - .env file won't exist on Railway
        if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
            log.Println("No .env file found, using system environment variables")
        }
    }
    
    if os.Getenv("GIN_MODE") == "" {
        os.Setenv("GIN_MODE", "release")
    }

    env := os.Getenv("APP_ENV")
    if env == "development" {
        Redirect_Uri = "http://localhost:3000/auth"
    }else{
    	Redirect_Uri = "https://repowipe.site/auth"
    }
	
	ClientId = os.Getenv("GITHUB_CLIENT_ID")
	ClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
}