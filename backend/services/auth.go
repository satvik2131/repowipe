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
func FetchAccessToken(c *gin.Context, tempCred types.TempCode)(*types.AccessTokenResponse,error) {
	var tokenResp types.AccessTokenResponse

	query := map[string]string{
		"client_id":     config.ClientId,
		"client_secret": config.ClientSecret,
		"code":         tempCred.Code,
		"redirect_uri": config.Redirect_Uri,
	}

	resp, err := utils.Client.R().
		SetQueryParams(query).
		SetResult(&tokenResp).
		Post(config.AccessTokenUrl)


	if err != nil {
		log.Printf("Error making request: %v", err)
		return nil,err
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("Error status: %d", resp.StatusCode())
		return nil,err
	}

	return &tokenResp,nil
}


func FetchUser(c *gin.Context, accessToken string) any  {
	var user types.User

	resp, err := utils.Client.R().
		SetHeader("Authorization", "Bearer "+accessToken).
		SetResult(&user).
		Get(config.GetUserApi)
		

	if err != nil {
		log.Printf("Error fetching user: %v", err)
		return nil
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("Error status: %d", resp.StatusCode())
		return nil
	}

	return user
}
