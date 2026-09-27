package youtube

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/db"
	"golang.org/x/oauth2"
)

func setupTestDB(t *testing.T) (db.StorageEngine, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "mercury-youtube-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := db.Open(dbPath)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("failed to open test db: %v", err)
	}

	cleanup := func() {
		_ = store.Close()
		_ = os.RemoveAll(tmpDir)
	}

	return store, cleanup
}

func TestEnvSecretProvider(t *testing.T) {
	ctx := context.Background()

	// 1. Missing credentials
	provider := NewEnvSecretProvider("", "", "")
	creds, err := provider.GetCredentials(ctx)
	if err == nil {
		t.Fatalf("expected error for missing credentials, got nil")
	}
	if creds.RedirectURL == "" {
		t.Errorf("expected default redirect URL to be populated, got empty")
	}

	// 2. Explicit credentials
	provider = NewEnvSecretProvider("test-client-id", "test-secret", "http://localhost:8080/callback")
	creds, err = provider.GetCredentials(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.ClientID != "test-client-id" || creds.ClientSecret != "test-secret" {
		t.Errorf("unexpected credentials: %+v", creds)
	}
}

func TestAuthManager_TokenLifecycle(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	secrets := NewEnvSecretProvider("test-id", "test-secret", "http://localhost:8080/cb")
	auth := NewAuthManager(secrets, store)

	if auth.IsAuthenticated() {
		t.Errorf("expected not authenticated initially")
	}

	// Save token
	tok := &oauth2.Token{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		Expiry:       time.Now().Add(1 * time.Hour),
	}
	if err := auth.SaveToken(tok); err != nil {
		t.Fatalf("failed to save token: %v", err)
	}

	if !auth.IsAuthenticated() {
		t.Errorf("expected authenticated after saving token")
	}

	loaded, err := auth.GetToken()
	if err != nil {
		t.Fatalf("failed to get token: %v", err)
	}
	if loaded.AccessToken != "mock-access-token" || loaded.RefreshToken != "mock-refresh-token" {
		t.Errorf("loaded token mismatch: %+v", loaded)
	}

	// Revoke token
	if err := auth.RevokeToken(context.Background()); err != nil {
		t.Fatalf("failed to revoke token: %v", err)
	}
	if auth.IsAuthenticated() {
		t.Errorf("expected not authenticated after revocation")
	}
}

func TestUploader_JobPersistenceAndCancellation(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	secrets := NewEnvSecretProvider("test-id", "test-secret", "http://localhost:8080/cb")
	auth := NewAuthManager(secrets, store)
	uploader := NewUploader(auth, store)

	job := &UploadJob{
		ID:            "job-test-1",
		TriggerSource: "test",
		FilePath:      "/tmp/sample.mp4",
		FileName:      "sample.mp4",
		TotalBytes:    1000,
		BytesUploaded: 500,
		ProgressPct:   50.0,
		Status:        "uploading",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := uploader.SaveJob(job); err != nil {
		t.Fatalf("failed to save job: %v", err)
	}

	loaded, err := uploader.GetJob("job-test-1")
	if err != nil {
		t.Fatalf("failed to get job: %v", err)
	}
	if loaded.ProgressPct != 50.0 || loaded.Status != "uploading" {
		t.Errorf("unexpected job: %+v", loaded)
	}

	// List jobs
	jobs, err := uploader.ListJobs(10)
	if err != nil {
		t.Fatalf("failed to list jobs: %v", err)
	}
	if len(jobs) != 1 || jobs[0].ID != "job-test-1" {
		t.Errorf("unexpected jobs list: %+v", jobs)
	}

	// Cancel job
	if err := uploader.CancelJob("job-test-1"); err != nil {
		t.Fatalf("failed to cancel job: %v", err)
	}
	cancelled, err := uploader.GetJob("job-test-1")
	if err != nil {
		t.Fatalf("failed to reload cancelled job: %v", err)
	}
	if cancelled.Status != "cancelled" {
		t.Errorf("expected status cancelled, got %s", cancelled.Status)
	}
}

func TestProgressReader(t *testing.T) {
	data := []byte("hello world video stream payload")
	total := int64(len(data))
	buf := bytes.NewReader(data)

	var lastUploaded int64
	var updatesCount int

	reader := NewProgressReader(buf, total, func(uploaded, total int64) {
		lastUploaded = uploaded
		updatesCount++
	})

	dst := &bytes.Buffer{}
	n, err := io.Copy(dst, reader)
	if err != nil {
		t.Fatalf("unexpected copy error: %v", err)
	}
	if n != total {
		t.Errorf("expected %d bytes copied, got %d", total, n)
	}
	if lastUploaded != total {
		t.Errorf("expected last uploaded %d, got %d", total, lastUploaded)
	}
	if updatesCount == 0 {
		t.Errorf("expected at least one update callback")
	}
}

func TestPipelineStep_MissingFile(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	secrets := NewEnvSecretProvider("test-id", "test-secret", "http://localhost:8080/cb")
	auth := NewAuthManager(secrets, store)
	uploader := NewUploader(auth, store)
	step := NewYouTubeUploadStep(uploader, func() string { return "ASTRO_SIG" })

	ctx := context.Background()
	res, err := step.Execute(ctx, PipelineArtifact{
		SourcePath: "/tmp/nonexistent-video-file-12345.mp4",
		Title:      "Test Video",
	})
	if err == nil {
		t.Fatalf("expected error for missing file, got nil")
	}
	if res.Status != "failed" {
		t.Errorf("expected result status failed, got %s", res.Status)
	}
}
