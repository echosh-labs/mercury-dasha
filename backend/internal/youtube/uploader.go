package youtube

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/db"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	ErrFileNotFound    = errors.New("source video file not found on disk")
	ErrJobNotFound     = errors.New("upload job not found")
	ErrAlreadyCanceled = errors.New("upload job was cancelled")
)

// ProgressReader wraps an io.Reader to report bytes read to a callback function.
type ProgressReader struct {
	reader   io.Reader
	total    int64
	uploaded int64
	onUpdate func(uploaded, total int64)
}

func NewProgressReader(r io.Reader, total int64, onUpdate func(uploaded, total int64)) *ProgressReader {
	return &ProgressReader{
		reader:   r,
		total:    total,
		onUpdate: onUpdate,
	}
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.uploaded += int64(n)
		if pr.onUpdate != nil {
			pr.onUpdate(pr.uploaded, pr.total)
		}
	}
	return n, err
}

// Uploader coordinates video file ingestion, resumable transfers, and job tracking.
type Uploader struct {
	auth          *AuthManager
	store         db.StorageEngine
	activeCancels map[string]context.CancelFunc
	mu            sync.RWMutex
	quotaMu       sync.Mutex
	quotaUsed     int
	quotaReset    time.Time
}

// NewUploader constructs an Uploader.
func NewUploader(auth *AuthManager, store db.StorageEngine) *Uploader {
	return &Uploader{
		auth:          auth,
		store:         store,
		activeCancels: make(map[string]context.CancelFunc),
		quotaReset:    nextMidnightUTC(),
	}
}

func nextMidnightUTC() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
}

func generateJobID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// SaveJob persists an upload job into BoltDB.
func (u *Uploader) SaveJob(job *UploadJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return u.store.SaveYouTubeJob(job.ID, data)
}

// GetJob retrieves an upload job by ID.
func (u *Uploader) GetJob(id string) (*UploadJob, error) {
	data, err := u.store.GetYouTubeJob(id)
	if err != nil {
		return nil, ErrJobNotFound
	}
	var job UploadJob
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, fmt.Errorf("failed to decode upload job: %w", err)
	}
	return &job, nil
}

// ListJobs returns the most recent upload jobs.
func (u *Uploader) ListJobs(limit int) ([]*UploadJob, error) {
	raws, err := u.store.ListYouTubeJobs(limit)
	if err != nil {
		return nil, err
	}
	var jobs []*UploadJob
	for _, raw := range raws {
		var j UploadJob
		if err := json.Unmarshal(raw, &j); err == nil {
			jobs = append(jobs, &j)
		}
	}
	return jobs, nil
}

// CancelJob cancels an in-progress upload job.
func (u *Uploader) CancelJob(id string) error {
	u.mu.Lock()
	cancel, exists := u.activeCancels[id]
	if exists {
		delete(u.activeCancels, id)
	}
	u.mu.Unlock()

	if exists && cancel != nil {
		cancel()
	}

	job, err := u.GetJob(id)
	if err == nil && job.Status == "uploading" {
		job.Status = "cancelled"
		job.ErrorMessage = "Upload cancelled by user"
		job.UpdatedAt = time.Now()
		_ = u.SaveJob(job)
	}

	return nil
}

func (u *Uploader) trackQuota(cost int) {
	u.quotaMu.Lock()
	defer u.quotaMu.Unlock()
	if time.Now().UTC().After(u.quotaReset) {
		u.quotaUsed = 0
		u.quotaReset = nextMidnightUTC()
	}
	u.quotaUsed += cost
}

// GetQuota returns current estimated quota usage.
func (u *Uploader) GetQuota() QuotaTracker {
	u.quotaMu.Lock()
	defer u.quotaMu.Unlock()
	if time.Now().UTC().After(u.quotaReset) {
		u.quotaUsed = 0
		u.quotaReset = nextMidnightUTC()
	}
	return QuotaTracker{
		DailyLimit:   10000,
		UsedToday:    u.quotaUsed,
		ResetTimeUTC: u.quotaReset,
	}
}

// ExecuteUpload launches an upload using the YouTube Data API v3.
func (u *Uploader) ExecuteUpload(ctx context.Context, req UploadRequest, triggerSource string, enricher func() string) (*UploadJob, error) {
	cleanPath := filepath.Clean(strings.TrimSpace(req.FilePath))
	info, err := os.Stat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrFileNotFound, cleanPath)
	}
	if info.IsDir() {
		return nil, errors.New("specified path is a directory, expected video file")
	}

	file, err := os.Open(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open video file: %w", err)
	}
	defer file.Close()

	tokenSource, err := u.auth.TokenSource(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication error: %w", err)
	}

	service, err := youtube.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize youtube service: %w", err)
	}

	// Prepare metadata
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(cleanPath), filepath.Ext(cleanPath))
	}

	description := req.Description
	if req.AttachChronoContext && enricher != nil {
		chronoSig := enricher()
		if chronoSig != "" {
			if description != "" {
				description += "\n\n"
			}
			description += chronoSig
		}
	}

	categoryID := req.CategoryID
	if categoryID == "" {
		categoryID = "22" // People & Blogs default
	}

	privacy := strings.ToLower(strings.TrimSpace(req.PrivacyStatus))
	if privacy != "public" && privacy != "unlisted" && privacy != "private" {
		privacy = "private" // Safe default
	}

	jobID := generateJobID()
	now := time.Now()
	job := &UploadJob{
		ID:            jobID,
		TriggerSource: triggerSource,
		FilePath:      cleanPath,
		FileName:      filepath.Base(cleanPath),
		Title:         title,
		TotalBytes:    info.Size(),
		BytesUploaded: 0,
		ProgressPct:   0.0,
		Status:        "uploading",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := u.SaveJob(job); err != nil {
		return nil, err
	}

	uploadCtx, cancel := context.WithCancel(ctx)
	u.mu.Lock()
	u.activeCancels[jobID] = cancel
	u.mu.Unlock()

	defer func() {
		u.mu.Lock()
		delete(u.activeCancels, jobID)
		u.mu.Unlock()
	}()

	videoObj := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       title,
			Description: description,
			Tags:        req.Tags,
			CategoryId:  categoryID,
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus:           privacy,
			SelfDeclaredMadeForKids: req.MadeForKids,
			Embeddable:              req.Embeddable,
		},
	}

	lastProgressSave := time.Now()
	progressReader := NewProgressReader(file, info.Size(), func(uploaded, total int64) {
		pct := 0.0
		if total > 0 {
			pct = float64(uploaded) / float64(total) * 100.0
		}
		job.BytesUploaded = uploaded
		job.ProgressPct = pct
		job.UpdatedAt = time.Now()

		// Throttle database writes to at most once per 2 seconds
		if time.Since(lastProgressSave) >= 2*time.Second || uploaded == total {
			lastProgressSave = time.Now()
			_ = u.SaveJob(job)
		}
	})

	call := service.Videos.Insert([]string{"snippet", "status"}, videoObj).Media(progressReader)
	call.Context(uploadCtx)

	uploadedVideo, err := call.Do()
	if err != nil {
		job.Status = "failed"
		job.ErrorMessage = err.Error()
		job.UpdatedAt = time.Now()
		_ = u.SaveJob(job)
		return job, err
	}

	completedAt := time.Now()
	job.Status = "completed"
	job.BytesUploaded = info.Size()
	job.ProgressPct = 100.0
	job.VideoID = uploadedVideo.Id
	job.VideoURL = fmt.Sprintf("https://youtu.be/%s", uploadedVideo.Id)
	job.CompletedAt = &completedAt
	job.UpdatedAt = completedAt
	_ = u.SaveJob(job)

	u.trackQuota(1600) // Videos.insert cost

	return job, nil
}

// GetChannelProfile fetches the verified channel profile for the connected account.
func (u *Uploader) GetChannelProfile(ctx context.Context) (*ChannelProfile, error) {
	tokenSource, err := u.auth.TokenSource(ctx)
	if err != nil {
		return nil, err
	}

	service, err := youtube.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to init youtube service: %w", err)
	}

	call := service.Channels.List([]string{"snippet", "statistics", "contentDetails"}).Mine(true)
	call.Context(ctx)
	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channel: %w", err)
	}

	if len(resp.Items) == 0 {
		return nil, errors.New("no youtube channel found for authenticated user")
	}

	ch := resp.Items[0]
	u.trackQuota(1)

	var thumb string
	if ch.Snippet.Thumbnails != nil {
		if ch.Snippet.Thumbnails.Default != nil {
			thumb = ch.Snippet.Thumbnails.Default.Url
		}
		if ch.Snippet.Thumbnails.Medium != nil {
			thumb = ch.Snippet.Thumbnails.Medium.Url
		}
	}

	var uploadsPlaylist string
	if ch.ContentDetails != nil && ch.ContentDetails.RelatedPlaylists != nil {
		uploadsPlaylist = ch.ContentDetails.RelatedPlaylists.Uploads
	}

	return &ChannelProfile{
		ChannelID:         ch.Id,
		Title:             ch.Snippet.Title,
		Description:       ch.Snippet.Description,
		CustomURL:         ch.Snippet.CustomUrl,
		ThumbnailURL:      thumb,
		SubscriberCount:   ch.Statistics.SubscriberCount,
		VideoCount:        ch.Statistics.VideoCount,
		UploadsPlaylistID: uploadsPlaylist,
		LastSyncedAt:      time.Now().UTC(),
	}, nil
}

// ListRecentVideos retrieves the latest uploaded videos on the channel.
func (u *Uploader) ListRecentVideos(ctx context.Context, maxResults int64) ([]VideoSummary, error) {
	tokenSource, err := u.auth.TokenSource(ctx)
	if err != nil {
		return nil, err
	}

	service, err := youtube.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, err
	}

	if maxResults <= 0 || maxResults > 50 {
		maxResults = 20
	}

	call := service.Search.List([]string{"snippet"}).ForMine(true).Type("video").Order("date").MaxResults(maxResults)
	call.Context(ctx)
	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to search uploaded videos: %w", err)
	}

	u.trackQuota(100) // Search.list cost

	var list []VideoSummary
	for _, item := range resp.Items {
		var thumb string
		if item.Snippet.Thumbnails != nil && item.Snippet.Thumbnails.Medium != nil {
			thumb = item.Snippet.Thumbnails.Medium.Url
		}
		pubTime, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		list = append(list, VideoSummary{
			ID:           item.Id.VideoId,
			Title:        item.Snippet.Title,
			Description:  item.Snippet.Description,
			ThumbnailURL: thumb,
			PublishedAt:  pubTime,
			WatchURL:     fmt.Sprintf("https://youtu.be/%s", item.Id.VideoId),
		})
	}

	return list, nil
}
