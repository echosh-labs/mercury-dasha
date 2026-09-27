package syncthing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
)

func TestProcessorIngestion(t *testing.T) {
	tmpDir := t.TempDir()
	mediaDir := filepath.Join(tmpDir, "media")
	thumbDir := filepath.Join(tmpDir, ".thumbnails")
	_ = os.MkdirAll(mediaDir, 0755)

	testImg := filepath.Join(mediaDir, "PXL_20260924_120000123.jpg")
	if err := os.WriteFile(testImg, []byte("fake-jpeg-data"), 0644); err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}

	timeline := []dasha.DashaPeriod{
		{
			Level:      dasha.LevelMahadasha,
			Planet:     dasha.PlanetSaturn,
			PlanetName: "Saturn",
			StartDate:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:    time.Date(2039, 1, 1, 0, 0, 0, 0, time.UTC),
			SubPeriods: []dasha.DashaPeriod{
				{
					Level:      dasha.LevelAntardasha,
					Planet:     dasha.PlanetJupiter,
					PlanetName: "Jupiter",
					StartDate:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					EndDate:    time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
		},
	}

	processor := NewMediaProcessor(nil, timeline, mediaDir, thumbDir)

	ctx := context.Background()
	item, err := processor.ProcessMediaFile(ctx, testImg, "test-folder")
	if err != nil {
		t.Fatalf("ProcessMediaFile failed: %v", err)
	}
	if item == nil {
		t.Fatal("expected item, got nil")
	}

	if item.Category != "photos" {
		t.Errorf("expected category 'photos', got '%s'", item.Category)
	}
	if item.DashaMahadasha != "Saturn" {
		t.Errorf("expected Mahadasha 'Saturn', got '%s'", item.DashaMahadasha)
	}
	if item.DashaAntardasha != "Jupiter" {
		t.Errorf("expected Antardasha 'Jupiter', got '%s'", item.DashaAntardasha)
	}

	feed := processor.GetRecentFeed(10)
	if len(feed) != 1 {
		t.Errorf("expected feed length 1, got %d", len(feed))
	}
}

func TestHTTPHandlers(t *testing.T) {
	tmpDir := t.TempDir()
	mediaDir := filepath.Join(tmpDir, "media")
	_ = os.MkdirAll(mediaDir, 0755)

	testImg := filepath.Join(mediaDir, "test.jpg")
	_ = os.WriteFile(testImg, []byte("fake-image"), 0644)

	sub := NewSubsystem("http://127.0.0.1:8384", "test-key", mediaDir, nil, nil)
	httpHandler := NewHTTPHandler(sub)

	// Test Feed
	req := httptest.NewRequest(http.MethodGet, "/api/v1/syncthing/feed", nil)
	w := httptest.NewRecorder()
	httpHandler.FeedHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("FeedHandler returned %d", w.Code)
	}

	// Test Stream
	sReq := httptest.NewRequest(http.MethodGet, "/api/v1/media/stream?path=test.jpg", nil)
	sW := httptest.NewRecorder()
	httpHandler.StreamHandler(sW, sReq)
	if sW.Code != http.StatusOK {
		t.Errorf("StreamHandler returned %d", sW.Code)
	}
	if sW.Body.String() != "fake-image" {
		t.Errorf("StreamHandler body mismatch: %s", sW.Body.String())
	}
}
