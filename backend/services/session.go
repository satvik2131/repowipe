package services

import (
	"encoding/json"
	"errors"
	"log"
	"repowipe/config"
	"repowipe/providers"
	"repowipe/types"
	"strings"
	"time"

	"github.com/google/uuid"
)

const sessionTTL = 24 * time.Hour

// oauthStatePayload is stored under oauth:state:{state}.
type oauthStatePayload struct {
	Provider types.Provider `json:"provider"`
	Mode     string         `json:"mode"` // login | link
}

// SaveOAuthState persists CSRF state with provider + mode.
// State format returned to the browser: {provider}:{mode}:{uuid}
func SaveOAuthState(state string, provider types.Provider, mode string) error {
	if mode != "link" {
		mode = "login"
	}
	b, err := json.Marshal(oauthStatePayload{Provider: provider, Mode: mode})
	if err != nil {
		return err
	}
	return config.RedisClient.Set(config.Ctx, "oauth:state:"+state, b, 5*time.Minute).Err()
}

// NewOAuthState creates a state token that embeds provider + mode for the frontend.
func NewOAuthState(provider types.Provider, mode string) string {
	if mode != "link" {
		mode = "login"
	}
	return string(provider) + ":" + mode + ":" + uuid.New().String()
}

// ConsumeOAuthState validates and deletes the state; returns provider + mode.
func ConsumeOAuthState(state string) (types.Provider, string, error) {
	key := "oauth:state:" + state
	raw, err := config.RedisClient.Get(config.Ctx, key).Result()
	if err != nil {
		return "", "", errors.New("invalid oauth state")
	}
	_ = config.RedisClient.Del(config.Ctx, key).Err()

	var payload oauthStatePayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		// Legacy: value was just "1"
		return types.ProviderGitHub, "login", nil
	}
	if !payload.Provider.Valid() {
		return "", "", errors.New("invalid oauth state provider")
	}
	if payload.Mode != "link" {
		payload.Mode = "login"
	}
	return payload.Provider, payload.Mode, nil
}

// GetSession loads a session document, migrating legacy plain-token sessions.
func GetSession(sessionID string) (*types.SessionDocument, error) {
	raw, err := config.RedisClient.Get(config.Ctx, "session:"+sessionID).Result()
	if err != nil {
		return nil, err
	}

	// Legacy: plain GitHub access token string
	if !strings.HasPrefix(strings.TrimSpace(raw), "{") {
		doc := &types.SessionDocument{
			Primary: types.ProviderGitHub,
			Providers: map[types.Provider]types.ProviderCredentials{
				types.ProviderGitHub: {AccessToken: raw},
			},
		}
		_ = SaveSession(sessionID, doc)
		return doc, nil
	}

	var doc types.SessionDocument
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, err
	}
	if doc.Providers == nil {
		doc.Providers = map[types.Provider]types.ProviderCredentials{}
	}
	return &doc, nil
}

// SaveSession persists the session document.
func SaveSession(sessionID string, doc *types.SessionDocument) error {
	b, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	return config.RedisClient.Set(config.Ctx, "session:"+sessionID, b, sessionTTL).Err()
}

// CreateSession creates a new session with a single primary provider.
func CreateSession(provider types.Provider, creds types.ProviderCredentials) (string, error) {
	sessionID := uuid.New().String()
	doc := &types.SessionDocument{
		Primary: provider,
		Providers: map[types.Provider]types.ProviderCredentials{
			provider: creds,
		},
	}
	if err := SaveSession(sessionID, doc); err != nil {
		return "", err
	}
	return sessionID, nil
}

// LinkProvider attaches another provider to an existing session.
func LinkProvider(sessionID string, provider types.Provider, creds types.ProviderCredentials) error {
	doc, err := GetSession(sessionID)
	if err != nil {
		return err
	}
	doc.Providers[provider] = creds
	return SaveSession(sessionID, doc)
}

// UnlinkProvider removes a non-primary provider.
func UnlinkProvider(sessionID string, provider types.Provider) error {
	doc, err := GetSession(sessionID)
	if err != nil {
		return err
	}
	if doc.Primary == provider {
		return errors.New("cannot unlink primary provider; logout or switch primary first")
	}
	creds, ok := doc.Providers[provider]
	if ok {
		if p, perr := providers.Get(provider); perr == nil {
			_ = p.Revoke(creds.AccessToken)
		}
		delete(doc.Providers, provider)
	}
	return SaveSession(sessionID, doc)
}

// DeleteSession revokes all tokens and removes the Redis key.
func DeleteSession(sessionID string) {
	doc, err := GetSession(sessionID)
	if err == nil && doc != nil {
		for name, creds := range doc.Providers {
			if p, perr := providers.Get(name); perr == nil {
				if revErr := p.Revoke(creds.AccessToken); revErr != nil {
					log.Printf("DeleteSession: revoke %s failed: %v", name, revErr)
				}
			}
		}
	}
	_ = config.RedisClient.Del(config.Ctx, "session:"+sessionID).Err()
}

// EnsureFreshToken refreshes Bitbucket tokens when near expiry.
func EnsureFreshToken(sessionID string, provider types.Provider) (string, error) {
	doc, err := GetSession(sessionID)
	if err != nil {
		return "", err
	}
	creds, ok := doc.Providers[provider]
	if !ok || creds.AccessToken == "" {
		return "", errors.New("provider not linked")
	}

	if (provider == types.ProviderBitbucket || provider == types.ProviderGitLab) {
		if creds.RefreshToken != "" && creds.ExpiresAt > 0 {
			if time.Now().Unix() >= creds.ExpiresAt-60 {
				p, perr := providers.Get(provider)
				if perr != nil {
					return "", perr
				}
				tok, rerr := p.RefreshToken(creds.RefreshToken)
				if rerr != nil {
					return "", rerr
				}
				creds.AccessToken = tok.AccessToken
				if tok.RefreshToken != "" {
					creds.RefreshToken = tok.RefreshToken
				}
				if tok.ExpiresIn > 0 {
					creds.ExpiresAt = time.Now().Unix() + int64(tok.ExpiresIn)
				}
				doc.Providers[provider] = creds
				_ = SaveSession(sessionID, doc)
			}
		}
	}
	return creds.AccessToken, nil
}

// ConnectionsFromSession builds the public connections response.
func ConnectionsFromSession(doc *types.SessionDocument) types.ConnectionsResponse {
	out := types.ConnectionsResponse{
		Primary:     doc.Primary,
		Connections: map[types.Provider]types.UserProfile{},
	}
	for name, creds := range doc.Providers {
		profile := creds.User
		if profile.Login == "" && creds.AccessToken != "" {
			if p, err := providers.Get(name); err == nil {
				if u, uerr := p.GetUser(creds.AccessToken); uerr == nil && u != nil {
					profile = *u
				}
			}
		}
		profile.Provider = name
		out.Connections[name] = profile
	}
	return out
}

// SessionExists reports whether a session key is present.
func SessionExists(sessionID string) bool {
	n, err := config.RedisClient.Exists(config.Ctx, "session:"+sessionID).Result()
	return err == nil && n > 0
}
