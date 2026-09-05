package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var ClientId, ClientSecret, Redirect_Uri string

// Legacy GitHub API constants (still used by older service helpers if referenced).
const (
	GithubAuthorizeURL = "https://github.com/login/oauth/authorize"
	GetUserApi         = "https://api.github.com/user"
	GetRepoApi         = "https://api.github.com/user/repos"
	AccessTokenUrl     = "https://github.com/login/oauth/access_token"
	SearchUri          = "https://api.github.com/search/repositories"
	DeleteApi          = "https://api.github.com/repos/"
	RevokeTokenURL     = "https://api.github.com/applications/"
)

func InitEnvVar() {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "production" {
		_ = godotenv.Load(".env.production")
	} else {
		_ = godotenv.Load(".env.development")
	}

	if os.Getenv("GIN_MODE") == "" {
		os.Setenv("GIN_MODE", "release")
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	Redirect_Uri = frontendURL + "/auth"

	ClientId = os.Getenv("GITHUB_CLIENT_ID")
	ClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")

	// At least one OAuth provider must be configured.
	hasGitHub := ClientId != "" && ClientSecret != ""
	hasGitLab := os.Getenv("GITLAB_CLIENT_ID") != "" && os.Getenv("GITLAB_CLIENT_SECRET") != ""
	hasBitbucket := os.Getenv("BITBUCKET_CLIENT_ID") != "" && os.Getenv("BITBUCKET_CLIENT_SECRET") != ""
	if !hasGitHub && !hasGitLab && !hasBitbucket {
		log.Fatal("configure at least one of GITHUB_*, GITLAB_*, or BITBUCKET_* client credentials")
	}
	if !hasGitHub {
		log.Println("warning: GITHUB_CLIENT_ID/SECRET not set")
	}
	if !hasGitLab {
		log.Println("warning: GITLAB_CLIENT_ID/SECRET not set")
	}
	if !hasBitbucket {
		log.Println("warning: BITBUCKET_CLIENT_ID/SECRET not set")
	}
}
