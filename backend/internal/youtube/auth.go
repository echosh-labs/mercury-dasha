package youtube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/db"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

var (
	ErrNotAuthenticated = errors.New("youtube is not authenticated; connect your channel via OAuth")
	ErrInvalidState      = errors.New("invalid or expired oauth state parameter")
)

// AuthManager coordinates OAuth 2.0 flows, persistent token lifecycle, and auto-refresh persistence.
type AuthManager struct {
	secrets SecretProvider
	store   db.StorageEngine
	mu      sync.RWMutex
}

// NewAuthManager initializes the authentication manager.
func NewAuthManager(secrets SecretProvider, store db.StorageEngine) *AuthManager {
	return &AuthManager{
		secrets: secrets,
		store:   store,
	}
}

// GetOAuthConfig retrieves application credentials and builds the oauth2.Config.
func (a *AuthManager) GetOAuthConfig(ctx context.Context) (*oauth2.Config, error) {
	creds, err := a.secrets.GetCredentials(ctx)
	if err != nil {
		return nil, err
	}

	return &oauth2.Config{
		ClientID:     creds.ClientID,
		ClientSecret: creds.ClientSecret,
		RedirectURL:  creds.RedirectURL,
		Scopes: []string{
			youtube.YoutubeUploadScope,
			youtube.YoutubeReadonlyScope,
			"https://www.googleapis.com/auth/yt-analytics.readonly",
			"https://www.googleapis.com/auth/yt-analytics-monetary.readonly",
		},
		Endpoint: google.Endpoint,
	}, nil
}

// GetAuthURL generates a Google OAuth 2.0 authorization URL requesting offline access and refresh token.
func (a *AuthManager) GetAuthURL(ctx context.Context, state string) (string, error) {
	cfg, err := a.GetOAuthConfig(ctx)
	if err != nil {
		return "", err
	}

	url := cfg.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	)
	return url, nil
}

// ExchangeCode exchanges an authorization code for an OAuth2 token and stores it in BoltDB.
func (a *AuthManager) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	cfg, err := a.GetOAuthConfig(ctx)
	if err != nil {
		return nil, err
	}

	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	if err := a.SaveToken(token); err != nil {
		return nil, fmt.Errorf("failed to save oauth token: %w", err)
	}

	return token, nil
}

// SaveToken serializes and persists an OAuth2 token to BoltDB.
func (a *AuthManager) SaveToken(token *oauth2.Token) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to serialize token: %w", err)
	}

	return a.store.SaveYouTubeToken(data)
}

// GetToken loads and deserializes the stored OAuth2 token from BoltDB.
func (a *AuthManager) GetToken() (*oauth2.Token, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	data, err := a.store.GetYouTubeToken()
	if err != nil {
		return nil, ErrNotAuthenticated
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to decode stored token: %w", err)
	}

	return &token, nil
}

// IsAuthenticated reports whether a valid token exists in the database.
func (a *AuthManager) IsAuthenticated() bool {
	token, err := a.GetToken()
	return err == nil && token != nil && token.RefreshToken != ""
}

// RevokeToken attempts to revoke the current token with Google and removes it from BoltDB.
func (a *AuthManager) RevokeToken(ctx context.Context) error {
	token, err := a.GetToken()
	if err == nil && token != nil {
		tokenToRevoke := token.RefreshToken
		if tokenToRevoke == "" {
			tokenToRevoke = token.AccessToken
		}
		if tokenToRevoke != "" {
			reqURL := fmt.Sprintf("https://oauth2.googleapis.com/revoke?token=%s", url.QueryEscape(tokenToRevoke))
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
			if err == nil {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				client := &http.Client{Timeout: 10 * time.Second}
				_, _ = client.Do(req)
			}
		}
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	return a.store.DeleteYouTubeToken()
}

// TokenSource returns an auto-refreshing oauth2.TokenSource that automatically
// writes updated tokens back into BoltDB when refreshed.
func (a *AuthManager) TokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	cfg, err := a.GetOAuthConfig(ctx)
	if err != nil {
		return nil, err
	}

	token, err := a.GetToken()
	if err != nil {
		return nil, err
	}

	rawSource := cfg.TokenSource(ctx, token)
	return &persistentTokenSource{
		source: rawSource,
		store:  a.store,
		last:   token,
	}, nil
}

type persistentTokenSource struct {
	source oauth2.TokenSource
	store  db.StorageEngine
	last   *oauth2.Token
	mu     sync.Mutex
}

func (pts *persistentTokenSource) Token() (*oauth2.Token, error) {
	pts.mu.Lock()
	defer pts.mu.Unlock()

	newToken, err := pts.source.Token()
	if err != nil {
		return nil, err
	}

	if pts.last == nil || newToken.AccessToken != pts.last.AccessToken {
		pts.last = newToken
		if data, err := json.Marshal(newToken); err == nil {
			_ = pts.store.SaveYouTubeToken(data)
		}
	}

	return newToken, nil
}
