package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var ClientId, ClientSecret, Redirect_Uri string

// all github apis
const (
	GithubAuthorizeURL = "https://github.com/login/oauth/authorize"
	GetUserApi         = "https://api.github.com/user"
	GetRepoApi         = "https://api.github.com/user/repos"
	AccessTokenUrl     = "https://github.com/login/oauth/access_token"
	SearchUri          = "https://api.github.com/search/repositories"
	DeleteApi          = "https://api.github.com/repos/"
)

func InitEnvVar() {
	// Determine which .env file to load based on APP_ENV.
	// When deploying to Koyeb, env vars are injected directly — Load() will
	// silently fail and that is fine.
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "production" {
		_ = godotenv.Load(".env.production")
	} else {
		// Default to development
		_ = godotenv.Load(".env.development")
	}

	// Set GIN_MODE to release if not specified
	if os.Getenv("GIN_MODE") == "" {
		os.Setenv("GIN_MODE", "release")
	}

	// Redirect URI — must match the GitHub OAuth App callback setting.
	// Set FRONTEND_URL to e.g. https://myapp.vercel.app (no trailing slash).
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000" // safe local default
	}
	Redirect_Uri = frontendURL + "/auth"

	// Load required environment variables
	ClientId = os.Getenv("GITHUB_CLIENT_ID")
	ClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")

	// Validate required variables are set
	if ClientId == "" || ClientSecret == "" {
		log.Fatal("GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET must be set")
	}
}