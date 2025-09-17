package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)


var ClientId, ClientSecret string
const (
	GetUserApi = "https://api.github.com/user"
 	GetRepoApi = "https://api.github.com/user/repos"
	AccessTokenUrl = "https://github.com/login/oauth/access_token"
	Redirect_Uri = "http://localhost:3000/auth/"
	SearchUri = "https://api.github.com/search/repositories"
 )


func InitEnvVar(){
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}	

	ClientId = os.Getenv("GITHUB_CLIENT_ID")
	ClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
}