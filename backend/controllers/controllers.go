package controllers

import (
	"log"
	"net/http"
	"os"
	"repowipe/config"
	"repowipe/services"
	"repowipe/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// verifies user
func VerifyUser(c *gin.Context) {
	_, err := c.Cookie("session_id")
	if err != nil {
		log.Println("err", err.Error())
		c.JSON(http.StatusOK, false)
		return
	}

	//you get the cookie
	c.JSON(http.StatusOK, true)
}

// Receives temporary code from github and then pass it to get the
// access token
func SetAccessToken(c *gin.Context) {
	//github oauth temp code and status (to be exchanged for access_token)
	var tempCred types.TempCode

	// Parse JSON request body into struct
	if err := c.ShouldBindJSON(&tempCred); err != nil {
		log.Println("error-=")
		c.JSON(http.StatusForbidden, gin.H{"status": "invalid code credentials"})
		return
	}

	accessTokenResp, err := services.FetchAccessToken(c, tempCred)
	if err != nil {
		c.JSON(http.StatusUnauthorized, err)
	}

	user := services.FetchUser(c, accessTokenResp.AccessToken)
	sessionID := saveToken(accessTokenResp.AccessToken)
	var allowedCookiesDomains string
	if env := os.Getenv("APP_ENV"); env == "development" {
		allowedCookiesDomains = "localhost"
	} else {
		allowedCookiesDomains = "repowipe.site"
	}

	c.SetCookie(
		"session_id",
		sessionID,
		3600,
		"/",
		allowedCookiesDomains,
		true,
		true,
	)
	c.JSON(200, gin.H{"user": user})
}

func FetchAllRepos(c *gin.Context) {
	page := c.Query("page")
	sessionId, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusUnauthorized, nil)
		return
	}

	accessToken := getToken(c, sessionId)
	services.FetchRepos(c, accessToken, page)
}

func SearchRepos(c *gin.Context) {
	sessionId, err := c.Cookie("session_id")
	username := c.Query("username")
	reponame := c.Query("reponame")

	if err != nil {
		c.JSON(http.StatusUnauthorized, nil)
		return
	}
	accessToken := getToken(c, sessionId)
	services.SearchRepos(c, accessToken, username, reponame)
}

func DeleteRepos(c *gin.Context) {
	sessionId, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusUnauthorized, nil)
	}

	accessToken := getToken(c, sessionId)
	var notFoundRepos []string

	var deleteRepoData types.GithubRepoDelete
	if err := c.ShouldBindJSON(&deleteRepoData); err != nil {
		log.Println("controller-del-repo: ", err.Error())
		c.JSON(http.StatusBadRequest, nil)
	}
	username := deleteRepoData.Username
	for _, repo := range deleteRepoData.Repos {
		err := services.DeleteRepos(c, accessToken, repo, username)
		if err != nil {
			log.Println("repo not found", err.Error())
			notFoundRepos = append(notFoundRepos, err.Error())
		}
	}

	if len(notFoundRepos) > 0 {
		c.JSON(http.StatusNotFound, notFoundRepos)
	} else {
		c.JSON(http.StatusOK, "Repos Deleted")
	}
}

// Utility
// save access token to redis
func saveToken(access_token string) string {
	ctx := config.Ctx
	sessionID := uuid.New().String() // random unique ID
	config.RedisClient.Set(ctx, "session:"+sessionID, access_token, 0)
	return sessionID
}

func getToken(c *gin.Context, tokenId string) string {
	ctx := config.Ctx
	accessToken, err := config.RedisClient.Get(ctx, "session:"+tokenId).Result()

	if err != nil {
		c.JSON(http.StatusUnauthorized, "Unauthorized")
	}

	return accessToken
}
