package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var ClientId, ClientSecret, Redirect_Uri string

//all github apis
const (
	GetUserApi     = "https://api.github.com/user"
	GetRepoApi     = "https://api.github.com/user/repos"
	AccessTokenUrl = "https://github.com/login/oauth/access_token"
	SearchUri      = "https://api.github.com/search/repositories"
	DeleteApi      = "https://api.github.com/repos/"
)

func InitEnvVar() {
	// Try to load .env file - only works in local development
	// Silently fails in Docker/Koyeb/Railway where env vars are injected
	_ = godotenv.Load()

	// Set GIN_MODE to release if not specified
	if os.Getenv("GIN_MODE") == "" {
		os.Setenv("GIN_MODE", "release")
	}

	// Determine redirect URI based on environment
	env := os.Getenv("APP_ENV")
	if env == "development" {
		Redirect_Uri = "http://localhost:3000/auth"
	} else {
		Redirect_Uri = "https://repowipe.site/auth"
	}

	// Load required environment variables
	ClientId = os.Getenv("GITHUB_CLIENT_ID")
	ClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")

	// Validate required variables are set
	if ClientId == "" || ClientSecret == "" {
		log.Fatal("GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET must be set")
	}
}