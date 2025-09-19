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

//Repositories related services
func FetchRepos (c *gin.Context,accessToken string,page string)  {
	var repos types.GitHubRepoList
	query := map[string]string{
		"page":page,
		"per_page":"10",
	}
	_,err := utils.Client.R().
	SetHeader("Authorization", "Bearer " + accessToken ).
	SetQueryParams(query).
	SetResult(&repos).
	Get(config.GetRepoApi)

	if err != nil{
		log.Println("FetchRepos--",err)
		c.JSON(http.StatusBadRequest,err)
		return
	}

	
	c.JSON(http.StatusOK,repos)
}


func DeleteRepos(c *gin.Context, accessToken ,reponame, username string ) (error) {
	resp,err := utils.Client.R().
		SetHeader("Authorization", "Bearer " + accessToken ).
		Delete(config.DeleteApi  + username +"/"+reponame )
	
	if err != nil {
		c.JSON(http.StatusBadRequest,err)
		return nil
	}

	//Repos Deleted
	if resp.Status() != "404 Not Found"{
		c.JSON(http.StatusOK,"Repos Deleted!")
		return nil;
	}

	
	return errors.New(reponame);
}


func SearchRepos(c *gin.Context, accessToken, username, reponame string ){
	var searchedRepos types.GitHubSearchResponse
	queryParams := fmt.Sprintf("?q=user:%s+%s",username,reponame)
	uri := config.SearchUri + queryParams
	resp,err := utils.Client.R().
				SetHeader("Authorization","Bearer "+ accessToken).
				SetResult(&searchedRepos).
				Get(uri)

	log.Println("Final_url---",resp.Request.URL)
	if err != nil {
		log.Println("SearchRepos--",err)
		c.JSON(http.StatusConflict,nil)
		return 
	}
	c.JSON(http.StatusOK,searchedRepos.Items)
}