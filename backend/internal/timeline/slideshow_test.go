package timeline_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/echosh-labs/mercury-dasha/internal/db"
	"github.com/echosh-labs/mercury-dasha/internal/timeline"
)

func TestNewTempleOfIlluminationSlideshow(t *testing.T) {
	s := timeline.NewTempleOfIlluminationSlideshow()
	if s == nil {
		t.Fatal("expected non-nil slideshow")
	}

	if s.ID != "temple-of-illumination" {
		t.Errorf("expected ID temple-of-illumination, got %s", s.ID)
	}

	if len(s.Slides) != 8 {
		t.Fatalf("expected 8 stanzas/slides, got %d", len(s.Slides))
	}

	if s.TotalDurationSec <= 0 {
		t.Errorf("expected positive total duration, got %.2f", s.TotalDurationSec)
	}

	// Verify Stanza 1
	slide1 := s.Slides[0]
	if slide1.Title != "The Ocean of Light" {
		t.Errorf("expected title 'The Ocean of Light', got '%s'", slide1.Title)
	}
	if slide1.SacredMetal != "Quicksilver" {
		t.Errorf("expected Quicksilver, got %s", slide1.SacredMetal)
	}
	if len(slide1.Verses) != 4 {
		t.Errorf("expected 4 verses in stanza 1, got %d", len(slide1.Verses))
	}

	// Verify Stanza 8
	slide8 := s.Slides[7]
	if slide8.Title != "The Innermost Circle" {
		t.Errorf("expected title 'The Innermost Circle', got '%s'", slide8.Title)
	}
	if slide8.Index != 8 {
		t.Errorf("expected index 8, got %d", slide8.Index)
	}
}

func TestGenerateTimelineManifest_FromSlideshow(t *testing.T) {
	s := timeline.NewTempleOfIlluminationSlideshow()

	opts := timeline.SlideshowOptions{
		Orientation:      timeline.OrientationLandscape16x9,
		IncludeOverlays:  true,
		EnableDucking:    true,
		VoiceAudioPath:   "audio/narration/temple_of_illumination.wav",
		AmbientMusicPath: "audio/ambient/drone_528hz.wav",
	}

	manifest, err := s.GenerateTimelineManifest(opts)
	if err != nil {
		t.Fatalf("failed to generate timeline manifest: %v", err)
	}

	if manifest == nil {
		t.Fatal("manifest should not be nil")
	}

	if len(manifest.Scenes) != 8 {
		t.Fatalf("expected 8 scenes in manifest, got %d", len(manifest.Scenes))
	}

	if len(manifest.AudioTracks) != 2 {
		t.Fatalf("expected 2 audio tracks (voice + ambient), got %d", len(manifest.AudioTracks))
	}

	if manifest.ChronoMeta == nil {
		t.Error("expected non-nil ChronoMeta")
	}

	if len(manifest.Overlays) != 3 {
		t.Errorf("expected 3 overlays (chrono badge, lower third, waveform), got %d", len(manifest.Overlays))
	}

	// Verify duration calculation
	if manifest.DurationSec < 90.0 {
		t.Errorf("expected duration >= 90s for 8 slides, got %.2f", manifest.DurationSec)
	}

	// Verify serialization
	jsonBytes, err := manifest.ToJSON()
	if err != nil {
		t.Fatalf("failed to marshal manifest to JSON: %v", err)
	}
	if len(jsonBytes) == 0 {
		t.Error("expected non-empty JSON output")
	}

	// Verify pipeline artifact conversion
	artifact := manifest.ToPipelineArtifact("rendered/temple_of_illumination.mp4", "unlisted", "24")
	if artifact.Title != manifest.Title {
		t.Errorf("expected artifact title '%s', got '%s'", manifest.Title, artifact.Title)
	}
}

func TestHermeticSlideshow_StoreAndIndex(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mercury-slideshow-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test bolt db: %v", err)
	}
	defer store.Close()

	s := timeline.NewTempleOfIlluminationSlideshow()

	// 1. Save to Hermetic Storage
	if err := timeline.SaveHermeticSlideshow(store, s); err != nil {
		t.Fatalf("failed to save slideshow: %v", err)
	}

	// 2. Retrieve from Hermetic Storage
	loaded, err := timeline.GetHermeticSlideshow(store, s.ID)
	if err != nil {
		t.Fatalf("failed to get slideshow: %v", err)
	}
	if loaded.Title != s.Title {
		t.Errorf("expected title '%s', got '%s'", s.Title, loaded.Title)
	}
	if len(loaded.Slides) != 8 {
		t.Errorf("expected 8 slides in loaded document, got %d", len(loaded.Slides))
	}

	// 3. List Hermetic Slideshows
	all, err := timeline.ListHermeticSlideshows(store)
	if err != nil {
		t.Fatalf("failed to list slideshows: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 slideshow, got %d", len(all))
	}

	// 4. Index Assets into Sovereign Storehouse
	if err := timeline.IndexSlideshowAssets(store, s, tmpDir); err != nil {
		t.Fatalf("failed to index slideshow assets: %v", err)
	}

	// Verify Search in Storehouse Index
	results, count, err := store.SearchIndex(db.IndexFilter{
		Query: "temple_of_illumination",
	})
	if err != nil {
		t.Fatalf("failed to search index: %v", err)
	}
	if count == 0 || len(results) == 0 {
		t.Errorf("expected indexed results for query 'temple_of_illumination', got %d", count)
	}
}
