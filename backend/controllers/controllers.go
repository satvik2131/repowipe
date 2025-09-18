package controllers

import (
	"log"
	"net/http"
	"repowipe/config"
	"repowipe/services"
	"repowipe/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

//verifies user
func VerifyUser(c *gin.Context){
	_,err  := c.Cookie("session_id")
	if err != nil {
		log.Println("err",err.Error())
		c.JSON(http.StatusOK,false)
		return;
	}

	//you get the cookie
	c.JSON(http.StatusOK,true)
}

//Receives temporary code from github and then pass it to get the 
//access token
func SetAccessToken( c *gin.Context){
	//github oauth temp code and status (to be exchanged for access_token)
	var tempCred types.TempCode;
	// Parse JSON request body into struct
	if err := c.ShouldBindJSON(&tempCred); err != nil {
		log.Println("error-=")
		c.JSON(http.StatusBadGateway,gin.H{"status":"invalid code credentials"})
		return
	}

	accessTokenResp := services.FetchAccessToken(c,tempCred)
	user := services.FetchUser(c,accessTokenResp.AccessToken)
	sessionID := saveToken(accessTokenResp.AccessToken)

	c.SetCookie(	
		"session_id",
		sessionID,
		0,
		"/",
		"localhost",
		true,
		true,
	)	
	c.JSON(200,gin.H{"user":user})
}

func FetchAllRepos (c *gin.Context){
	page := c.Query("page")
	sessionId,err := c.Cookie("session_id")
	if err != nil{
		c.JSON(http.StatusUnauthorized,nil)
		return
	}

	accessToken := getToken(sessionId)
	log.Println("access_token-=",accessToken)
	services.FetchRepos(c,accessToken,page)
}


func SearchRepos (c *gin.Context){
	sessionId,err := c.Cookie("session_id")
	username := c.Query("username")
	reponame := c.Query("reponame")

	if err != nil{
		c.JSON(http.StatusUnauthorized,nil)
		return
	}
	accessToken := getToken(sessionId)
	services.SearchRepos(c,accessToken,username,reponame)
}


func DeleteRepos(c *gin.Context){
		var deleteRepoData types.GithubRepoDelete
		if err := c.ShouldBindJSON(&deleteRepoData); err != nil{
			c.JSON(http.StatusBadRequest,nil)
		}
		services.DeleteRepos(c, deleteRepoData )
}



//Utility
//save access token to redis
func saveToken(access_token string) string {
	ctx := config.Ctx
	sessionID := uuid.New().String() // random unique ID
	config.RedisClient.Set(ctx,"session:" + sessionID,access_token,0)
	return sessionID
}


func getToken(tokenId string) string {
		ctx := config.Ctx
		accessToken,err := config.RedisClient.Get(ctx, "session:"+tokenId ).Result()
		if err != nil{
			panic(err)
		}
		return accessToken
	}



