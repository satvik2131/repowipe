package controllers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"repowipe/config"
	"repowipe/services"
	"repowipe/types"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// oauthStateTTL is how long the CSRF state token lives in Redis.
const oauthStateTTL = 5 * time.Minute

// GetGithubLoginURL generates the GitHub OAuth authorization URL entirely
// server-side and returns it to the frontend. The client_id, redirect_uri,
// scope, and a CSRF-proof state token are all set here — nothing sensitive
// is ever shipped to the browser.
func GetGithubLoginURL(c *gin.Context) {
	// Generate a cryptographically-random state token for CSRF protection.
	state := uuid.New().String()

	// Persist the state in Redis with a 5-minute TTL so we can validate it
	// when GitHub redirects back to /auth.
	ctx := config.Ctx
	if err := config.RedisClient.Set(ctx, "oauth:state:"+state, "1", oauthStateTTL).Err(); err != nil {
		log.Printf("GetGithubLoginURL: failed to persist state: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not initiate login"})
		return
	}

	scope := "repo,user,delete_repo"
	authURL := fmt.Sprintf(
		"%s?client_id=%s&redirect_uri=%s&scope=%s&state=%s",
		config.GithubAuthorizeURL,
		url.QueryEscape(config.ClientId),
		url.QueryEscape(config.Redirect_Uri),
		url.QueryEscape(scope),
		url.QueryEscape(state),
	)

	c.JSON(http.StatusOK, gin.H{"url": authURL})
}

// VerifyUser checks whether the request carries a valid session cookie.
func VerifyUser(c *gin.Context) {
	_, err := c.Cookie("session_id")
	if err != nil {
		log.Println("VerifyUser: no session cookie:", err)
		c.JSON(http.StatusOK, false)
		return
	}
	c.JSON(http.StatusOK, true)
}

// SetAccessToken exchanges the temporary GitHub OAuth code for an access token,
// fetches the authenticated user, stores the token in Redis, and sets a
// secure HttpOnly session cookie.
func SetAccessToken(c *gin.Context) {
	var tempCred types.TempCode
	if err := c.ShouldBindJSON(&tempCred); err != nil {
		log.Println("SetAccessToken: invalid payload:", err)
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid code credentials"})
		return
	}

	if tempCred.State == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "missing oauth state"})
		return
	}
	stateKey := "oauth:state:" + tempCred.State
	if err := config.RedisClient.Get(config.Ctx, stateKey).Err(); err != nil {
		log.Println("SetAccessToken: invalid oauth state:", err)
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid oauth state"})
		return
	}
	_ = config.RedisClient.Del(config.Ctx, stateKey).Err()

	accessTokenResp, err := services.FetchAccessToken(c, tempCred)
	if err != nil || accessTokenResp == nil || accessTokenResp.AccessToken == "" {
		log.Println("SetAccessToken: FetchAccessToken error:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "could not exchange code for token"})
		return
	}

	user := services.FetchUser(c, accessTokenResp.AccessToken)
	if user == nil {
		log.Println("SetAccessToken: FetchUser returned nil")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "could not fetch github user"})
		return
	}
	sessionID := saveToken(accessTokenResp.AccessToken)
	applySessionCookie(c, sessionID, 3600)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// Logout deletes the Redis session, revokes the GitHub OAuth token (best-effort),
// and clears the session cookie. Idempotent when no cookie is present.
func Logout(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err == nil && sessionID != "" {
		ctx := config.Ctx
		key := "session:" + sessionID

		accessToken, getErr := config.RedisClient.Get(ctx, key).Result()
		if getErr == nil && accessToken != "" {
			if revErr := services.RevokeAccessToken(accessToken); revErr != nil {
				log.Printf("Logout: GitHub token revoke failed (continuing): %v", revErr)
			}
		}

		if delErr := config.RedisClient.Del(ctx, key).Err(); delErr != nil {
			log.Printf("Logout: failed to delete Redis session: %v", delErr)
		}
	}

	applySessionCookie(c, "", -1)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ── helpers ──────────────────────────────────────────────────────────────────

// applySessionCookie sets or clears the session_id cookie using the same
// Domain / Secure / SameSite rules for login and logout so attributes never drift.
func applySessionCookie(c *gin.Context, value string, maxAge int) {
	// COOKIE_DOMAIN is the domain the session cookie is scoped to.
	// - Local dev: "localhost"
	// - Cross-site (frontend on Vercel, backend on Koyeb): leave empty so the
	//   cookie is host-only on the backend origin; browsers will still send it
	//   with credentialed requests when SameSite=None; Secure.
	cookieDomain := os.Getenv("COOKIE_DOMAIN")

	// For cross-site requests (frontend and backend on different registrable
	// domains) the cookie must be SameSite=None and Secure. In dev we keep
	// SameSite=Lax over http://localhost.
	if os.Getenv("APP_ENV") == "development" {
		c.SetSameSite(http.SameSiteLaxMode)
	} else {
		c.SetSameSite(http.SameSiteNoneMode)
	}

	secure := os.Getenv("APP_ENV") != "development"
	c.SetCookie("session_id", value, maxAge, "/", cookieDomain, secure, true)
}

// saveToken persists the GitHub access token in Redis under a random session
// ID and returns that session ID.
func saveToken(accessToken string) string {
	ctx := config.Ctx
	sessionID := uuid.New().String()
	config.RedisClient.Set(ctx, "session:"+sessionID, accessToken, 0)
	return sessionID
}

// getToken retrieves the GitHub access token from Redis for the given session
// ID, returning an empty string (and writing an Unauthorized response) if not
// found.
func getToken(c *gin.Context, sessionID string) string {
	ctx := config.Ctx
	accessToken, err := config.RedisClient.Get(ctx, "session:"+sessionID).Result()
	if err != nil {
		log.Println("getToken: session not found:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return ""
	}
	return accessToken
}
