package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/echosh-labs/mercury-dasha/internal/config"
	"github.com/echosh-labs/mercury-dasha/internal/db"
)

func TestTreasuryHandlers_EventAndStatus(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mercury-treasury-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer store.Close()

	cfg := &config.Config{
		ServiceName: "mercury-dasha-test",
		Environment: "test",
		BoltDBPath:  dbPath,
	}

	h := NewHandler(cfg, store)
	defer h.Close()

	// 1. Post Treasury Success Event
	eventPayload := map[string]any{
		"type":          "youtube_view",
		"title":         "🎉 New YouTube View Detected",
		"message":       "A new organic view occurred on Toroidal Singularity.",
		"metric_label":  "+1 View",
		"metric_value":  "Total: 5",
		"delta":         1,
		"channel_title": "Justin Andrew Wood",
		"video_title":   "Toroidal Singularity (Akasha Spanda)",
		"video_url":     "https://youtu.be/hAyUSVMVW7k",
	}

	body, _ := json.Marshal(eventPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/treasury/events", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.TreasuryEventHandler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["received"] != true {
		t.Errorf("expected received=true, got %v", resp["received"])
	}

	// 2. Status Handler
	reqStatus := httptest.NewRequest(http.MethodGet, "/api/v1/treasury/status", nil)
	recStatus := httptest.NewRecorder()
	h.TreasuryStatusHandler(recStatus, reqStatus)

	if recStatus.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recStatus.Code, recStatus.Body.String())
	}

	var statusResp map[string]any
	if err := json.Unmarshal(recStatus.Body.Bytes(), &statusResp); err != nil {
		t.Fatalf("failed to decode status: %v", err)
	}

	if statusResp["service"] != "amra-treasury" {
		t.Errorf("expected service=amra-treasury, got %v", statusResp["service"])
	}
}
