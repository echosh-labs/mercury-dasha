package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/config"
	"github.com/echosh-labs/mercury-dasha/internal/db"
	"github.com/echosh-labs/mercury-dasha/internal/grecorder"
)

func setupTestRecorderHandler(t *testing.T) (*Handler, *http.ServeMux, string, func()) {
	tmpDir, err := os.MkdirTemp("", "mercury-grecorder-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := db.Open(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to open test db: %v", err)
	}

	recordingsDir := filepath.Join(tmpDir, "recordings")
	_ = os.MkdirAll(recordingsDir, 0755)

	// Create test manifest with a mock recording
	testManifest := grecorder.SyncManifest{
		Version:  "1.0",
		LastSync: time.Now(),
		Recordings: map[string]grecorder.ManifestEntry{
			"test-rec-1": {
				RecordingID:        "test-rec-1",
				Title:              "Hermetic Philosophy and Sacred Metals",
				RecordedAt:         time.Now().Add(-1 * time.Hour),
				Duration:           "04:20",
				DurationMs:         260000,
				AudioFile:          "test-rec-1.m4a",
				AudioSizeBytes:     1024,
				TranscriptTextFile: "test-rec-1.txt",
				TranscriptTextSize: 120,
				HasTranscript:      true,
			},
		},
	}
	manifestBytes, _ := json.Marshal(testManifest)
	_ = os.WriteFile(filepath.Join(recordingsDir, "manifest.json"), manifestBytes, 0644)

	// Create dummy audio and txt transcript
	_ = os.WriteFile(filepath.Join(recordingsDir, "test-rec-1.m4a"), []byte("dummy-m4a-data"), 0644)
	dummyTranscript := "[00:00] Speaker 1: As above, so below.\n[00:15] Speaker 2: Principles of Polarity and Vibration."
	_ = os.WriteFile(filepath.Join(recordingsDir, "test-rec-1.txt"), []byte(dummyTranscript), 0644)

	cfg := &config.Config{
		Port:                 "8080",
		BoltDBPath:           dbPath,
		DropboxLocalPath:     tmpDir,
		RecorderMCPURL:       "http://127.0.0.1:9999/mcp", // offline mock URL
		RecorderOutputDir:    recordingsDir,
		RecorderAutoSyncMins: 0, // poller disabled for test
	}

	h := NewHandler(cfg, store)
	router := NewRouter(h)

	cleanup := func() {
		h.Close()
		store.Close()
		os.RemoveAll(tmpDir)
	}

	return h, router.mux, recordingsDir, cleanup
}

func TestRecorderEndpoints_StatusAndList(t *testing.T) {
	_, mux, _, cleanup := setupTestRecorderHandler(t)
	defer cleanup()

	// 1. Test GET /api/v1/recorder/status
	req := httptest.NewRequest("GET", "/api/v1/recorder/status", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var status grecorder.StatusOverview
	if err := json.NewDecoder(rr.Body).Decode(&status); err != nil {
		t.Fatalf("failed to decode status: %v", err)
	}
	if status.TotalLocalRecordings != 1 {
		t.Errorf("expected 1 local recording, got %d", status.TotalLocalRecordings)
	}

	// 2. Test GET /api/v1/recorder/recordings
	reqList := httptest.NewRequest("GET", "/api/v1/recorder/recordings", nil)
	rrList := httptest.NewRecorder()
	mux.ServeHTTP(rrList, reqList)

	if rrList.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rrList.Code)
	}

	var listRes struct {
		Recordings []grecorder.Recording `json:"recordings"`
		Total      int                   `json:"total"`
	}
	if err := json.NewDecoder(rrList.Body).Decode(&listRes); err != nil {
		t.Fatalf("failed to decode list: %v", err)
	}
	if listRes.Total != 1 {
		t.Fatalf("expected 1 recording in list, got %d", listRes.Total)
	}
	if listRes.Recordings[0].Title != "Hermetic Philosophy and Sacred Metals" {
		t.Errorf("unexpected recording title: %s", listRes.Recordings[0].Title)
	}

	// 3. Test GET /api/v1/recorder/recordings/test-rec-1/transcript
	reqTrans := httptest.NewRequest("GET", "/api/v1/recorder/recordings/test-rec-1/transcript", nil)
	rrTrans := httptest.NewRecorder()
	mux.ServeHTTP(rrTrans, reqTrans)

	if rrTrans.Code != http.StatusOK {
		t.Fatalf("expected status 200 for transcript, got %d (body: %s)", rrTrans.Code, rrTrans.Body.String())
	}

	var payload grecorder.TranscriptPayload
	if err := json.NewDecoder(rrTrans.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode transcript: %v", err)
	}
	if len(payload.Paragraphs) != 2 {
		t.Fatalf("expected 2 paragraphs, got %d", len(payload.Paragraphs))
	}
	if payload.Paragraphs[0].Speaker != "Speaker 1" || payload.Paragraphs[0].Text != "As above, so below." {
		t.Errorf("unexpected paragraph 0: %+v", payload.Paragraphs[0])
	}

	// 4. Test GET /api/v1/recorder/recordings/test-rec-1/audio
	reqAudio := httptest.NewRequest("GET", "/api/v1/recorder/recordings/test-rec-1/audio", nil)
	rrAudio := httptest.NewRecorder()
	mux.ServeHTTP(rrAudio, reqAudio)

	if rrAudio.Code != http.StatusOK {
		t.Fatalf("expected status 200 for audio, got %d", rrAudio.Code)
	}
	if rrAudio.Header().Get("Content-Type") != "audio/mp4" {
		t.Errorf("expected Content-Type audio/mp4, got %s", rrAudio.Header().Get("Content-Type"))
	}
	if rrAudio.Body.String() != "dummy-m4a-data" {
		t.Errorf("unexpected audio data: %s", rrAudio.Body.String())
	}
}
