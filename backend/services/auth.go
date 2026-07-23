package services

import (
	"errors"
	"log"
	"net/http"
	"repowipe/config"
	"repowipe/types"
	"repowipe/utils"

	"github.com/gin-gonic/gin"
)

//fetches the access_token in exchange of temporary credentials
func FetchAccessToken(c *gin.Context, tempCred types.TempCode) (*types.AccessTokenResponse, error) {
	var tokenResp types.AccessTokenResponse

	query := map[string]string{
		"client_id":     config.ClientId,
		"client_secret": config.ClientSecret,
		"code":          tempCred.Code,
		"redirect_uri":  config.Redirect_Uri,
	}

	// Prefer form body — GitHub's documented exchange method.
	resp, err := utils.Client.R().
		SetHeader("Accept", "application/json").
		SetFormData(query).
		SetResult(&tokenResp).
		Post(config.AccessTokenUrl)

	if err != nil {
		log.Printf("Error making request: %v", err)
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("Error status: %d body=%s", resp.StatusCode(), resp.String())
		return nil, err
	}

	if tokenResp.AccessToken == "" {
		log.Printf("FetchAccessToken: empty access_token body=%s", resp.String())
		return nil, errors.New("empty access_token from github")
	}

	return &tokenResp, nil
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

// RevokeAccessToken invalidates a GitHub OAuth access token for this app.
// Best-effort: callers should still complete local session cleanup on failure.
func RevokeAccessToken(accessToken string) error {
	if accessToken == "" {
		return nil
	}

	url := config.RevokeTokenURL + config.ClientId + "/token"
	resp, err := utils.Client.R().
		SetBasicAuth(config.ClientId, config.ClientSecret).
		SetHeader("Accept", "application/vnd.github+json").
		SetHeader("X-GitHub-Api-Version", "2022-11-28").
		SetBody(map[string]string{"access_token": accessToken}).
		Delete(url)

	if err != nil {
		log.Printf("RevokeAccessToken: request error: %v", err)
		return err
	}

	// 204 No Content = success; 404 = already revoked / unknown — treat as ok.
	if resp.StatusCode() != http.StatusNoContent && resp.StatusCode() != http.StatusNotFound {
		log.Printf("RevokeAccessToken: unexpected status %d body=%s", resp.StatusCode(), resp.String())
		return errors.New("github token revoke failed")
	}

	return nil
}
