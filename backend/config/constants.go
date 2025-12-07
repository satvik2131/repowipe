package config

import (
	"os"
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