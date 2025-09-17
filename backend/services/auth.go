package services

import (
	"log"
	"net/http"
	"repowipe/config"
	"repowipe/types"
	"repowipe/utils"

	"github.com/gin-gonic/gin"
)

//fetches the access_token in exchange of temporary credentials
func FetchAccessToken(c *gin.Context, tempCred types.TempCode) types.AccessTokenResponse {
	var tokenResp types.AccessTokenResponse

	resp, err := utils.Client.R().
		SetQueryParams(map[string]string{
			"client_id":     config.ClientId,
			"client_secret": config.ClientSecret,
			"code":         tempCred.Code,
			"redirect_uri": config.Redirect_Uri,
		}).
		SetResult(&tokenResp).
		Post(config.AccessTokenUrl)

	if err != nil {
		log.Printf("Error making request: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return types.AccessTokenResponse{}
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("Error status: %d", resp.StatusCode())
		c.JSON(resp.StatusCode(), gin.H{"error": "Failed to get access token"})
		return types.AccessTokenResponse{}
	}

	return tokenResp
}


func FetchUser(c *gin.Context, accessToken string) types.User {
	var user types.User

	resp, err := utils.Client.R().
		SetHeader("Authorization", "Bearer "+accessToken).
		SetResult(&user).
		Get(config.GetUserApi)

	if err != nil {
		log.Printf("Error fetching user: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return types.User{}
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("Error status: %d", resp.StatusCode())
		c.JSON(resp.StatusCode(), gin.H{"error": "Failed to fetch user"})
		return types.User{}
	}

	log.Println("user---",user)

	return user
}
