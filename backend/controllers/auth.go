package controllers

import (
	"log"
	"net/http"
	"os"
	"repowipe/config"
	"repowipe/providers"
	"repowipe/services"
	"repowipe/types"
	"time"

	"github.com/gin-gonic/gin"
)

// GetProviderLoginURL returns the OAuth authorize URL for :provider.
// Query mode=login|link (default login). Link requires an existing session.
func GetProviderLoginURL(c *gin.Context) {
	providerName := types.Provider(c.Param("provider"))
	if !providerName.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown provider"})
		return
	}
	p, err := providers.Get(providerName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mode := c.DefaultQuery("mode", "login")
	if mode == "link" {
		sessionID, err := c.Cookie("session_id")
		if err != nil || sessionID == "" || !services.SessionExists(sessionID) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "must be logged in to link a provider"})
			return
		}
	} else {
		mode = "login"
	}

	state := services.NewOAuthState(providerName, mode)
	if err := services.SaveOAuthState(state, providerName, mode); err != nil {
		log.Printf("GetProviderLoginURL: persist state: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not initiate login"})
		return
	}

	redirectURI := oauthRedirectURI(providerName, mode)
	authURL, err := p.AuthorizeURL(redirectURI, state, mode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not build authorize url"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": authURL})
}

// ProviderCallback exchanges the OAuth code and creates or links a session.
func ProviderCallback(c *gin.Context) {
	providerName := types.Provider(c.Param("provider"))
	if !providerName.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown provider"})
		return
	}

	var req types.AuthCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if req.Code == "" || req.State == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "missing code or state"})
		return
	}

	stateProvider, mode, err := services.ConsumeOAuthState(req.State)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid oauth state"})
		return
	}
	if stateProvider != providerName {
		c.JSON(http.StatusForbidden, gin.H{"error": "provider mismatch"})
		return
	}
	if req.Mode == "link" || req.Mode == "login" {
		mode = req.Mode
	}

	p, err := providers.Get(providerName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	redirectURI := oauthRedirectURI(providerName, mode)
	tok, err := p.ExchangeCode(req.Code, redirectURI)
	if err != nil || tok == nil || tok.AccessToken == "" {
		log.Printf("ProviderCallback: exchange failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "could not exchange code for token"})
		return
	}

	user, err := p.GetUser(tok.AccessToken)
	if err != nil || user == nil {
		log.Printf("ProviderCallback: GetUser failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "could not fetch user"})
		return
	}

	creds := types.ProviderCredentials{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		User:         *user,
	}
	if tok.ExpiresIn > 0 {
		creds.ExpiresAt = time.Now().Unix() + int64(tok.ExpiresIn)
	}

	if mode == "link" {
		sessionID, cerr := c.Cookie("session_id")
		if cerr != nil || sessionID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "must be logged in to link"})
			return
		}
		if err := services.LinkProvider(sessionID, providerName, creds); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "session not found"})
			return
		}
		doc, _ := services.GetSession(sessionID)
		c.JSON(http.StatusOK, gin.H{
			"user":        primaryUser(doc),
			"connections": services.ConnectionsFromSession(doc),
			"mode":        "link",
		})
		return
	}

	sessionID, err := services.CreateSession(providerName, creds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create session"})
		return
	}
	applySessionCookie(c, sessionID, int(24*time.Hour/time.Second))

	doc, _ := services.GetSession(sessionID)
	c.JSON(http.StatusOK, gin.H{
		"user":        user,
		"connections": services.ConnectionsFromSession(doc),
		"mode":        "login",
	})
}

// GetConnections returns linked providers for the current session.
func GetConnections(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	doc, err := services.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.JSON(http.StatusOK, services.ConnectionsFromSession(doc))
}

// UnlinkProvider removes a non-primary linked provider.
func UnlinkProvider(c *gin.Context) {
	providerName := types.Provider(c.Param("provider"))
	if !providerName.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown provider"})
		return
	}
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if err := services.UnlinkProvider(sessionID, providerName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	doc, _ := services.GetSession(sessionID)
	c.JSON(http.StatusOK, services.ConnectionsFromSession(doc))
}

// VerifyUser checks whether the request carries a valid session cookie.
func VerifyUser(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" || !services.SessionExists(sessionID) {
		c.JSON(http.StatusOK, false)
		return
	}
	c.JSON(http.StatusOK, true)
}

// Logout deletes the Redis session, revokes tokens, and clears the cookie.
func Logout(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err == nil && sessionID != "" {
		services.DeleteSession(sessionID)
	}
	applySessionCookie(c, "", -1)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// SetAccessToken is the legacy GitHub callback path.
func SetAccessToken(c *gin.Context) {
	c.Params = append(c.Params, gin.Param{Key: "provider", Value: "github"})
	ProviderCallback(c)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func oauthRedirectURI(provider types.Provider, mode string) string {
	// Exact FRONTEND_URL/auth — provider + mode live in Redis/state, not the URI,
	// so OAuth app callback registration stays a single URL per environment.
	_ = provider
	_ = mode
	return config.Redirect_Uri
}

func applySessionCookie(c *gin.Context, value string, maxAge int) {
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	if os.Getenv("APP_ENV") == "development" {
		c.SetSameSite(http.SameSiteLaxMode)
	} else {
		c.SetSameSite(http.SameSiteNoneMode)
	}
	secure := os.Getenv("APP_ENV") != "development"
	c.SetCookie("session_id", value, maxAge, "/", cookieDomain, secure, true)
}

func primaryUser(doc *types.SessionDocument) *types.UserProfile {
	if doc == nil {
		return nil
	}
	if creds, ok := doc.Providers[doc.Primary]; ok {
		u := creds.User
		u.Provider = doc.Primary
		if u.Login != "" {
			return &u
		}
	}
	for name, creds := range doc.Providers {
		u := creds.User
		u.Provider = name
		return &u
	}
	return nil
}

func requireSession(c *gin.Context) (string, *types.SessionDocument, bool) {
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return "", nil, false
	}
	doc, err := services.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return "", nil, false
	}
	return sessionID, doc, true
}

func requireProviderToken(c *gin.Context, sessionID string, provider types.Provider) (string, bool) {
	token, err := services.EnsureFreshToken(sessionID, provider)
	if err != nil || token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "provider not linked: " + string(provider)})
		return "", false
	}
	return token, true
}
