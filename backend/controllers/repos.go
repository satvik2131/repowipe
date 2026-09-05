package controllers

import (
	"log"
	"net/http"
	"repowipe/providers"
	"repowipe/types"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// FetchProviderRepos lists repos for :provider.
func FetchProviderRepos(c *gin.Context) {
	providerName := types.Provider(c.Param("provider"))
	if !providerName.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown provider"})
		return
	}
	sessionID, _, ok := requireSession(c)
	if !ok {
		return
	}
	token, ok := requireProviderToken(c, sessionID, providerName)
	if !ok {
		return
	}
	p, err := providers.Get(providerName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	visibility := c.DefaultQuery("visibility", "all")
	sort := c.DefaultQuery("sort", "updated")
	direction := c.DefaultQuery("direction", "desc")

	repos, err := p.ListRepos(token, page, visibility, sort, direction)
	if err != nil {
		log.Printf("FetchProviderRepos: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, repos)
}

// SearchProviderRepos searches repos for :provider.
func SearchProviderRepos(c *gin.Context) {
	providerName := types.Provider(c.Param("provider"))
	if !providerName.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown provider"})
		return
	}
	sessionID, doc, ok := requireSession(c)
	if !ok {
		return
	}
	token, ok := requireProviderToken(c, sessionID, providerName)
	if !ok {
		return
	}
	p, err := providers.Get(providerName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := c.Query("username")
	if username == "" {
		if creds, ok := doc.Providers[providerName]; ok && creds.User.Login != "" {
			username = creds.User.Login
		}
	}
	reponame := c.Query("q")
	if reponame == "" {
		reponame = c.Query("reponame")
	}
	language := c.Query("language")
	visibility := c.Query("visibility")
	kind := c.Query("kind")
	sort := c.DefaultQuery("sort", "updated")

	repos, err := p.SearchRepos(token, username, reponame, language, visibility, kind, sort)
	if err != nil {
		log.Printf("SearchProviderRepos: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, repos)
}

// DeleteProviderRepos deletes repos on :provider.
func DeleteProviderRepos(c *gin.Context) {
	providerName := types.Provider(c.Param("provider"))
	if !providerName.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown provider"})
		return
	}
	sessionID, doc, ok := requireSession(c)
	if !ok {
		return
	}
	token, ok := requireProviderToken(c, sessionID, providerName)
	if !ok {
		return
	}
	p, err := providers.Get(providerName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var body types.RepoDeleteRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	owner := body.Username
	if owner == "" {
		if creds, ok := doc.Providers[providerName]; ok {
			owner = creds.User.Login
		}
	}
	if owner == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username required"})
		return
	}

	var notFound []string
	for _, repo := range body.Repos {
		name := repo
		repoOwner := owner
		if strings.Contains(repo, "/") {
			parts := strings.SplitN(repo, "/", 2)
			repoOwner, name = parts[0], parts[1]
		}
		if err := p.DeleteRepo(token, repoOwner, name); err != nil {
			log.Printf("DeleteProviderRepos: %v", err)
			notFound = append(notFound, name)
		}
	}
	if len(notFound) > 0 {
		c.JSON(http.StatusNotFound, notFound)
		return
	}
	c.JSON(http.StatusOK, "Repos Deleted")
}

// Legacy GitHub routes — delegate to provider-scoped handlers.

func FetchAllRepos(c *gin.Context) {
	c.Params = append(c.Params, gin.Param{Key: "provider", Value: "github"})
	FetchProviderRepos(c)
}

func SearchRepos(c *gin.Context) {
	c.Params = append(c.Params, gin.Param{Key: "provider", Value: "github"})
	SearchProviderRepos(c)
}

func DeleteRepos(c *gin.Context) {
	c.Params = append(c.Params, gin.Param{Key: "provider", Value: "github"})
	DeleteProviderRepos(c)
}
