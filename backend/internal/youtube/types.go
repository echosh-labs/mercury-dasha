package youtube

import (
	"context"
	"time"
)

// ChannelProfile contains verified YouTube channel info.
type ChannelProfile struct {
	ChannelID         string    `json:"channel_id"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	CustomURL         string    `json:"custom_url"`
	ThumbnailURL      string    `json:"thumbnail_url"`
	SubscriberCount   uint64    `json:"subscriber_count"`
	VideoCount        uint64    `json:"video_count"`
	UploadsPlaylistID string    `json:"uploads_playlist_id"`
	LastSyncedAt      time.Time `json:"last_synced_at"`
}

// UploadRequest defines input payload to launch a video upload.
type UploadRequest struct {
	// Source: either a local POSIX path (from Dropbox storehouse or local disk)
	FilePath            string   `json:"file_path"`
	
	// Video Metadata
	Title               string   `json:"title"`
	Description         string   `json:"description"`
	Tags                []string `json:"tags"`
	CategoryID          string   `json:"category_id"`       // Default: "22" (People & Blogs)
	PrivacyStatus       string   `json:"privacy_status"`    // "private", "unlisted", "public"
	MadeForKids         bool     `json:"made_for_kids"`
	Embeddable          bool     `json:"embeddable"`
	NotifySubscribers   bool     `json:"notify_subscribers"`

	// Astrological & Alchemical Enrichment
	AttachChronoContext bool     `json:"attach_chrono_context"`
}

// UploadJob tracks the progress of an asynchronous video upload.
type UploadJob struct {
	ID             string            `json:"id"`
	TriggerSource  string            `json:"trigger_source"` // "ui_manual", "pipeline_auto", "api"
	FilePath       string            `json:"file_path"`
	FileName       string            `json:"file_name"`
	Title          string            `json:"title,omitempty"`
	TotalBytes     int64             `json:"total_bytes"`
	BytesUploaded  int64             `json:"bytes_uploaded"`
	ProgressPct    float64           `json:"progress_pct"`
	Status         string            `json:"status"` // "queued", "uploading", "completed", "failed", "cancelled"
	VideoID        string            `json:"video_id,omitempty"`
	VideoURL       string            `json:"video_url,omitempty"`
	ErrorMessage   string            `json:"error_message,omitempty"`
	ChronoContext  map[string]string `json:"chrono_context,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	CompletedAt    *time.Time        `json:"completed_at,omitempty"`
}

// PipelineArtifact represents the input to the automated upload pipeline step.
type PipelineArtifact struct {
	JobID               string            `json:"job_id,omitempty"`
	SourcePath          string            `json:"source_path"`
	Title               string            `json:"title"`
	Description         string            `json:"description"`
	Tags                []string          `json:"tags"`
	CategoryID          string            `json:"category_id"`
	PrivacyStatus       string            `json:"privacy_status"`
	AttachChronoContext bool              `json:"attach_chrono_context"`
	Metadata            map[string]string `json:"metadata,omitempty"`
}

// PipelineResult represents the outcome of an automated upload pipeline step.
type PipelineResult struct {
	JobID        string  `json:"job_id"`
	VideoID      string  `json:"video_id"`
	VideoURL     string  `json:"video_url"`
	BytesSent    int64   `json:"bytes_sent"`
	DurationSecs float64 `json:"duration_secs"`
	Status       string  `json:"status"`
	ErrorMessage string  `json:"error_message,omitempty"`
}

// PipelineStep defines the contract for an autonomous pipeline stage.
type PipelineStep interface {
	Execute(ctx context.Context, artifact PipelineArtifact) (*PipelineResult, error)
}

// QuotaTracker monitors estimated daily YouTube Data API quota units (default 10,000 / day).
type QuotaTracker struct {
	DailyLimit   int       `json:"daily_limit"`
	UsedToday    int       `json:"used_today"`
	ResetTimeUTC time.Time `json:"reset_time_utc"`
}

// YouTubeStatusResponse returns full telemetry of the integration.
type YouTubeStatusResponse struct {
	Configured     bool            `json:"configured"`
	Authenticated  bool            `json:"authenticated"`
	TokenExpiry    *time.Time      `json:"token_expiry,omitempty"`
	Channel        *ChannelProfile `json:"channel,omitempty"`
	RecentJobs     []*UploadJob    `json:"recent_jobs"`
	EstimatedQuota QuotaTracker    `json:"quota"`
}

// VideoSummary is a lightweight item for recent uploads.
type VideoSummary struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	ThumbnailURL string    `json:"thumbnail_url"`
	PublishedAt  time.Time `json:"published_at"`
	Privacy      string    `json:"privacy"`
	WatchURL     string    `json:"watch_url"`
}
