package controllers

import (
	"log"
	"net/http"
	"repowipe/services"
	"repowipe/types"

	"github.com/gin-gonic/gin"
)

// FetchAllRepos returns a paginated list of the authenticated user's repos.
func FetchAllRepos(c *gin.Context) {
	page := c.Query("page")
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	accessToken := getToken(c, sessionID)
	if accessToken == "" {
		return // getToken already wrote the 401 response
	}
	services.FetchRepos(c, accessToken, page)
}

// SearchRepos proxies a search query to the GitHub search API.
func SearchRepos(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	username := c.Query("username")
	reponame := c.Query("reponame")

	accessToken := getToken(c, sessionID)
	if accessToken == "" {
		return
	}
	services.SearchRepos(c, accessToken, username, reponame)
}

// DeleteRepos deletes each repo in the request body for the authenticated user.
func DeleteRepos(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	accessToken := getToken(c, sessionID)
	if accessToken == "" {
		return
	}

	var deleteRepoData types.GithubRepoDelete
	if err := c.ShouldBindJSON(&deleteRepoData); err != nil {
		log.Println("DeleteRepos: invalid payload:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	var notFoundRepos []string
	for _, repo := range deleteRepoData.Repos {
		if err := services.DeleteRepos(c, accessToken, repo, deleteRepoData.Username); err != nil {
			log.Println("DeleteRepos: repo not found:", err)
			notFoundRepos = append(notFoundRepos, err.Error())
		}
	}

	if len(notFoundRepos) > 0 {
		c.JSON(http.StatusNotFound, notFoundRepos)
		return
	}
	c.JSON(http.StatusOK, "Repos Deleted")
}
