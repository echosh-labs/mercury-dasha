package youtube

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/db"
	"golang.org/x/oauth2"
)

// Client is the primary sovereign facade for the YouTube subsystem in Mercury Dasha.
type Client struct {
	secrets      SecretProvider
	store        db.StorageEngine
	auth         *AuthManager
	uploader     *Uploader
	pipelineStep *YouTubeUploadStep
	enricher     func() string

	profileCache   *ChannelProfile
	profileCacheAt time.Time
	mu             sync.RWMutex
}

// NewClient constructs a YouTube Client engine.
func NewClient(secrets SecretProvider, store db.StorageEngine, enricher func() string) *Client {
	auth := NewAuthManager(secrets, store)
	uploader := NewUploader(auth, store)
	step := NewYouTubeUploadStep(uploader, enricher)

	return &Client{
		secrets:      secrets,
		store:        store,
		auth:         auth,
		uploader:     uploader,
		pipelineStep: step,
		enricher:     enricher,
	}
}

// SetAstrologicalEnricher configures the dynamic context signature callback.
func (c *Client) SetAstrologicalEnricher(fn func() string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enricher = fn
	c.pipelineStep = NewYouTubeUploadStep(c.uploader, fn)
}

// IsConfigured returns true if OAuth client credentials (client_id & client_secret) are present.
func (c *Client) IsConfigured(ctx context.Context) bool {
	creds, err := c.secrets.GetCredentials(ctx)
	return err == nil && creds != nil && strings.TrimSpace(creds.ClientID) != "" && strings.TrimSpace(creds.ClientSecret) != ""
}

// IsAuthenticated returns true if a valid OAuth2 token is stored in the database.
func (c *Client) IsAuthenticated() bool {
	return c.auth.IsAuthenticated()
}

// GetAuthURL generates the OAuth2 consent URL.
func (c *Client) GetAuthURL(ctx context.Context, state string) (string, error) {
	return c.auth.GetAuthURL(ctx, state)
}

// ExchangeCode processes the OAuth2 redirect callback and stores the token.
func (c *Client) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := c.auth.ExchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}
	// Invalidate profile cache upon new authentication
	c.mu.Lock()
	c.profileCache = nil
	c.mu.Unlock()
	return token, nil
}

// Disconnect revokes and clears stored OAuth tokens.
func (c *Client) Disconnect(ctx context.Context) error {
	c.mu.Lock()
	c.profileCache = nil
	c.mu.Unlock()
	return c.auth.RevokeToken(ctx)
}

// UploadVideo queues and executes a video upload.
func (c *Client) UploadVideo(ctx context.Context, req UploadRequest, triggerSource string) (*UploadJob, error) {
	c.mu.RLock()
	enricher := c.enricher
	c.mu.RUnlock()

	if triggerSource == "" {
		triggerSource = "ui_manual"
	}

	return c.uploader.ExecuteUpload(ctx, req, triggerSource, enricher)
}

// GetJob retrieves an upload job by ID.
func (c *Client) GetJob(id string) (*UploadJob, error) {
	return c.uploader.GetJob(id)
}

// ListJobs lists recent upload jobs.
func (c *Client) ListJobs(limit int) ([]*UploadJob, error) {
	return c.uploader.ListJobs(limit)
}

// CancelJob cancels an upload in progress.
func (c *Client) CancelJob(id string) error {
	return c.uploader.CancelJob(id)
}

// GetChannelProfile returns channel metadata, using a 10-minute in-memory cache.
func (c *Client) GetChannelProfile(ctx context.Context) (*ChannelProfile, error) {
	c.mu.RLock()
	if c.profileCache != nil && time.Since(c.profileCacheAt) < 10*time.Minute {
		cached := c.profileCache
		c.mu.RUnlock()
		return cached, nil
	}
	c.mu.RUnlock()

	profile, err := c.uploader.GetChannelProfile(ctx)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.profileCache = profile
	c.profileCacheAt = time.Now()
	c.mu.Unlock()

	return profile, nil
}

// ListRecentVideos fetches recent uploads for the authenticated channel.
func (c *Client) ListRecentVideos(ctx context.Context, maxResults int64) ([]VideoSummary, error) {
	return c.uploader.ListRecentVideos(ctx, maxResults)
}

// PipelineStep returns the automated pipeline stage instance.
func (c *Client) PipelineStep() PipelineStep {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.pipelineStep
}

// GetStatus returns the full status payload for telemetry and UI dashboards.
func (c *Client) GetStatus(ctx context.Context) YouTubeStatusResponse {
	configured := c.IsConfigured(ctx)
	authenticated := c.IsAuthenticated()

	var expiry *time.Time
	if authenticated {
		if token, err := c.auth.GetToken(); err == nil && token != nil {
			exp := token.Expiry
			if !exp.IsZero() {
				expiry = &exp
			}
		}
	}

	var channel *ChannelProfile
	if authenticated {
		channel, _ = c.GetChannelProfile(ctx)
	}

	recentJobs, _ := c.ListJobs(10)
	quota := c.uploader.GetQuota()

	return YouTubeStatusResponse{
		Configured:     configured,
		Authenticated:  authenticated,
		TokenExpiry:    expiry,
		Channel:        channel,
		RecentJobs:     recentJobs,
		EstimatedQuota: quota,
	}
}
