package services

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"repowipe/config"
	"repowipe/types"
	"repowipe/utils"

	"github.com/gin-gonic/gin"
)

// Repositories related services
func FetchRepos(c *gin.Context, accessToken string, page, visibility, sort, direction string) {
	var repos types.GitHubRepoList
	if page == "" {
		page = "1"
	}
	if visibility == "" {
		visibility = "all"
	}
	if sort == "" {
		sort = "updated"
	}
	if direction == "" {
		direction = "desc"
	}

	query := map[string]string{
		"page":        page,
		"per_page":    "10",
		"visibility":  visibility,
		"sort":        sort,
		"direction":   direction,
		"affiliation": "owner",
	}
	_, err := utils.Client.R().
		SetHeader("Authorization", "Bearer "+accessToken).
		SetQueryParams(query).
		SetResult(&repos).
		Get(config.GetRepoApi)

	if err != nil {
		log.Println("FetchRepos--", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, repos)
}

func DeleteRepos(c *gin.Context, accessToken, reponame, username string) error {
	resp, err := utils.Client.R().
		SetHeader("Authorization", "Bearer "+accessToken).
		Delete(config.DeleteApi + username + "/" + reponame)

	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return nil
	}

	status := resp.StatusCode()
	log.Println("status --", status)
	//Repos Deleted
	if status != 404 {
		c.JSON(http.StatusOK, "Repos Deleted!")
		return nil
	}

	return errors.New(reponame)
}

func SearchRepos(c *gin.Context, accessToken, username, reponame, language, visibility, kind, sort string) {
	var searchedRepos types.GitHubSearchResponse

	q := fmt.Sprintf("user:%s", username)
	if reponame != "" {
		q += " " + reponame + " in:name,description"
	}
	if language != "" {
		q += " language:" + language
	}
	switch visibility {
	case "public":
		q += " is:public"
	case "private":
		q += " is:private"
	}
	switch kind {
	case "forks":
		q += " fork:only archived:false"
	case "sources":
		q += " fork:false archived:false"
	case "archived":
		q += " archived:true"
	}

	if sort == "" || (sort != "stars" && sort != "updated" && sort != "forks") {
		sort = "updated"
	}

	resp, err := utils.Client.R().
		SetHeader("Authorization", "Bearer "+accessToken).
		SetQueryParams(map[string]string{
			"q":        q,
			"sort":     sort,
			"order":    "desc",
			"per_page": "30",
		}).
		SetResult(&searchedRepos).
		Get(config.SearchUri)

	if resp != nil {
		log.Println("Final_url---", resp.Request.URL)
	}
	if err != nil {
		log.Println("SearchRepos--", err)
		c.JSON(http.StatusConflict, nil)
		return
	}
	c.JSON(http.StatusOK, searchedRepos.Items)
}
