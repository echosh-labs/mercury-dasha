package youtube

import (
	"context"
	"errors"
	"os"
	"strings"
)

var ErrMissingCredentials = errors.New("youtube client_id or client_secret is not configured")

// Credentials encapsulates the Google OAuth 2.0 application keys.
type Credentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
}

// SecretProvider abstracts credential retrieval to allow seamless evolution
// from simple environment variables to encrypted databases or cloud secret managers.
type SecretProvider interface {
	GetCredentials(ctx context.Context) (*Credentials, error)
}

// EnvSecretProvider resolves YouTube OAuth keys from system environment variables or .env.
type EnvSecretProvider struct {
	clientID     string
	clientSecret string
	redirectURL  string
}

// NewEnvSecretProvider creates a provider initialized with optional explicit credentials
// or falling back to environment variables.
func NewEnvSecretProvider(clientID, clientSecret, redirectURL string) *EnvSecretProvider {
	return &EnvSecretProvider{
		clientID:     strings.TrimSpace(clientID),
		clientSecret: strings.TrimSpace(clientSecret),
		redirectURL:  strings.TrimSpace(redirectURL),
	}
}

func (p *EnvSecretProvider) GetCredentials(_ context.Context) (*Credentials, error) {
	cid := p.clientID
	if cid == "" {
		cid = strings.TrimSpace(os.Getenv("YOUTUBE_CLIENT_ID"))
	}

	csec := p.clientSecret
	if csec == "" {
		csec = strings.TrimSpace(os.Getenv("YOUTUBE_CLIENT_SECRET"))
	}

	rurl := p.redirectURL
	if rurl == "" {
		rurl = strings.TrimSpace(os.Getenv("YOUTUBE_REDIRECT_URL"))
	}
	if rurl == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		rurl = "http://localhost:" + port + "/api/v1/youtube/auth/callback"
	}

	if cid == "" || csec == "" {
		return &Credentials{
			ClientID:     cid,
			ClientSecret: csec,
			RedirectURL:  rurl,
		}, ErrMissingCredentials
	}

	return &Credentials{
		ClientID:     cid,
		ClientSecret: csec,
		RedirectURL:  rurl,
	}, nil
}
