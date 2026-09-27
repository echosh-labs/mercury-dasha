package youtube

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
)

// YouTubeUploadStep implements the PipelineStep interface for automated video publishing pipelines.
type YouTubeUploadStep struct {
	uploader *Uploader
	enricher func() string
}

// NewYouTubeUploadStep creates an upload step ready to be linked into automated build chains.
func NewYouTubeUploadStep(uploader *Uploader, enricher func() string) *YouTubeUploadStep {
	return &YouTubeUploadStep{
		uploader: uploader,
		enricher: enricher,
	}
}

// Execute performs the automated publishing stage for a finished video asset.
func (s *YouTubeUploadStep) Execute(ctx context.Context, artifact PipelineArtifact) (*PipelineResult, error) {
	start := time.Now()

	if artifact.SourcePath == "" {
		return nil, errors.New("pipeline artifact missing source_path")
	}

	info, err := os.Stat(artifact.SourcePath)
	if err != nil {
		return &PipelineResult{
			JobID:        artifact.JobID,
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("artifact file not accessible: %v", err),
			DurationSecs: time.Since(start).Seconds(),
		}, err
	}

	uploadReq := UploadRequest{
		FilePath:            artifact.SourcePath,
		Title:               artifact.Title,
		Description:         artifact.Description,
		Tags:                artifact.Tags,
		CategoryID:          artifact.CategoryID,
		PrivacyStatus:       artifact.PrivacyStatus,
		AttachChronoContext: artifact.AttachChronoContext,
	}

	job, err := s.uploader.ExecuteUpload(ctx, uploadReq, "pipeline_auto", s.enricher)
	duration := time.Since(start).Seconds()

	if err != nil {
		jobID := ""
		if job != nil {
			jobID = job.ID
		}
		return &PipelineResult{
			JobID:        jobID,
			Status:       "failed",
			ErrorMessage: err.Error(),
			DurationSecs: duration,
		}, err
	}

	return &PipelineResult{
		JobID:        job.ID,
		VideoID:      job.VideoID,
		VideoURL:     job.VideoURL,
		BytesSent:    info.Size(),
		DurationSecs: duration,
		Status:       "completed",
	}, nil
}
