package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/echosh-labs/mercury-dasha/internal/api"
	"github.com/echosh-labs/mercury-dasha/internal/config"
	"github.com/echosh-labs/mercury-dasha/internal/db"
	"github.com/echosh-labs/mercury-dasha/internal/timeline"
)

func setupTestSlideshowRouter(t *testing.T) (http.Handler, db.StorageEngine, string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "mercury-slideshow-api-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	cfg := &config.Config{
		ServiceName:      "mercury-dasha-test",
		Environment:      "test",
		DropboxLocalPath: tmpDir,
	}

	// Seed canonical slideshow
	if err := timeline.SeedTempleOfIlluminationSlideshow(store, tmpDir); err != nil {
		t.Fatalf("failed to seed slideshow: %v", err)
	}

	handler := api.NewHandler(cfg, store)
	router := api.NewRouter(handler)
	return router, store, tmpDir
}

func TestSlideshowAPI_ListSlideshows(t *testing.T) {
	router, store, tmpDir := setupTestSlideshowRouter(t)
	defer store.Close()
	defer os.RemoveAll(tmpDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/slideshows", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Count      int              `json:"count"`
		Slideshows []map[string]any `json:"slideshows"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if resp.Count == 0 || len(resp.Slideshows) == 0 {
		t.Fatalf("expected at least 1 slideshow, got %d", resp.Count)
	}

	found := false
	for _, s := range resp.Slideshows {
		if s["id"] == "temple-of-illumination" {
			found = true
			if s["slide_count"] != float64(8) {
				t.Errorf("expected 8 slides, got %v", s["slide_count"])
			}
			break
		}
	}
	if !found {
		t.Error("temple-of-illumination not found in slideshow list")
	}
}

func TestSlideshowAPI_GetSlideshow(t *testing.T) {
	router, store, tmpDir := setupTestSlideshowRouter(t)
	defer store.Close()
	defer os.RemoveAll(tmpDir)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/slideshows/temple-of-illumination", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var s timeline.HermeticSlideshow
	if err := json.Unmarshal(rec.Body.Bytes(), &s); err != nil {
		t.Fatalf("failed to unmarshal slideshow: %v", err)
	}

	if s.ID != "temple-of-illumination" {
		t.Errorf("expected ID temple-of-illumination, got %s", s.ID)
	}
	if len(s.Slides) != 8 {
		t.Fatalf("expected 8 slides, got %d", len(s.Slides))
	}
	if s.Slides[0].Title != "The Ocean of Light" {
		t.Errorf("expected slide 1 title 'The Ocean of Light', got '%s'", s.Slides[0].Title)
	}
}

func TestSlideshowAPI_GenerateManifest(t *testing.T) {
	router, store, tmpDir := setupTestSlideshowRouter(t)
	defer store.Close()
	defer os.RemoveAll(tmpDir)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/slideshows/temple-of-illumination/manifest?orientation=16:9", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var manifest timeline.TimelineManifest
	if err := json.Unmarshal(rec.Body.Bytes(), &manifest); err != nil {
		t.Fatalf("failed to unmarshal timeline manifest: %v", err)
	}

	if len(manifest.Scenes) != 8 {
		t.Fatalf("expected 8 scenes in manifest, got %d", len(manifest.Scenes))
	}
	if manifest.Canvas.Orientation != timeline.OrientationLandscape16x9 {
		t.Errorf("expected 16:9 orientation, got %s", manifest.Canvas.Orientation)
	}
	if len(manifest.AudioTracks) == 0 {
		t.Error("expected audio tracks in generated manifest")
	}
}
