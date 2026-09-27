package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/db"
)

func TestUnifiedAudioCatalogHandler(t *testing.T) {
	h, mux, recordingsDir, cleanup := setupTestRecorderHandler(t)
	defer cleanup()

	// Seed index with an audio vault entry
	dummyVaultDoc := db.IndexEntry{
		ID:        "audio/vault/guitar_riff.wav",
		Category:  "audio",
		Path:      "audio/vault/guitar_riff.wav",
		FullPath:  filepath.Join(recordingsDir, "guitar_riff.wav"),
		FileName:  "guitar_riff.wav",
		Extension: ".wav",
		SizeBytes: 1024,
		ModTime:   time.Now(),
		Metadata: map[string]any{
			"title":        "Studio Guitar Riff",
			"duration_sec": 42.5,
		},
	}
	if err := h.store.BatchPutIndexEntries([]db.IndexEntry{dummyVaultDoc}); err != nil {
		t.Fatalf("Failed to save vault doc: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/v1/audio/catalog", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp UnifiedAudioCatalogResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Total == 0 {
		t.Errorf("Expected at least 1 audio item, got %d", resp.Total)
	}

	foundRecorder := false
	foundVault := false
	for _, it := range resp.Items {
		if it.ID == "test-rec-1" {
			foundRecorder = true
			if !it.HasTranscript {
				t.Errorf("Expected recorder item to have transcript")
			}
			if it.Format != "m4a" {
				t.Errorf("Expected format m4a, got %s", it.Format)
			}
		}
		if it.Title == "Studio Guitar Riff" {
			foundVault = true
			if it.Format != "wav" {
				t.Errorf("Expected format wav, got %s", it.Format)
			}
			if it.Source != "vault" {
				t.Errorf("Expected source vault, got %s", it.Source)
			}
		}
	}

	if !foundRecorder {
		t.Errorf("Expected to find test-rec-1 in unified catalog items")
	}
	if !foundVault {
		t.Errorf("Expected to find Studio Guitar Riff in unified catalog items")
	}

	if resp.Counts.Total < 2 {
		t.Errorf("Expected Counts.Total >= 2, got %d", resp.Counts.Total)
	}
	if resp.Counts.WithTranscript < 1 {
		t.Errorf("Expected Counts.WithTranscript >= 1, got %d", resp.Counts.WithTranscript)
	}
}
