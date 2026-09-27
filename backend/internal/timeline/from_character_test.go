package timeline

import (
	"testing"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
)

func TestGenerateFromCharacterDasha(t *testing.T) {
	req := dasha.CalculationRequest{
		Name:           "Cadence Vael",
		BirthDate:      "1992-06-15",
		BirthTime:      "12:00",
		TimezoneOffset: -4,
		Latitude:       43.1594,
		Longitude:      -79.2469,
	}

	prof, err := dasha.CalculateFromRequest(req)
	if err != nil {
		t.Fatalf("failed to calculate profile: %v", err)
	}
	prof.Title = "The Alchemical Architect"

	opts := CharacterToTimelineOptions{
		Orientation:     OrientationLandscape16x9,
		DurationSec:     60.0,
		IncludeOverlays: true,
	}

	manifest, err := GenerateFromCharacterDasha(prof, nil, opts)
	if err != nil {
		t.Fatalf("failed to generate character timeline: %v", err)
	}

	if manifest.ID == "" {
		t.Errorf("expected non-empty manifest ID")
	}
	if manifest.DurationSec != 60.0 {
		t.Errorf("expected 60.0s duration, got %f", manifest.DurationSec)
	}
	if manifest.Canvas.Orientation != OrientationLandscape16x9 {
		t.Errorf("expected 16:9 orientation, got %s", manifest.Canvas.Orientation)
	}
	if len(manifest.Scenes) != 4 {
		t.Errorf("expected 4 scenes, got %d", len(manifest.Scenes))
	}
	if len(manifest.AudioTracks) < 1 {
		t.Errorf("expected at least 1 audio track, got %d", len(manifest.AudioTracks))
	}
	if len(manifest.Overlays) != 3 {
		t.Errorf("expected 3 overlays, got %d", len(manifest.Overlays))
	}
	if manifest.ChronoMeta == nil {
		t.Fatalf("expected non-nil ChronoMeta")
	}
	if manifest.ChronoMeta.SacredMetal == "" {
		t.Errorf("expected non-empty SacredMetal in ChronoMeta")
	}

	// Validate JSON Serialization
	data, err := manifest.ToJSON()
	if err != nil {
		t.Fatalf("failed to serialize manifest to JSON: %v", err)
	}
	if len(data) == 0 {
		t.Errorf("expected non-empty JSON data")
	}

	// Validate Pipeline Artifact conversion
	artifact := manifest.ToPipelineArtifact("/tmp/output.mp4", "private", "24")
	if artifact.Title == "" {
		t.Errorf("expected non-empty artifact title")
	}
	if artifact.Metadata["vimshottari_lord"] == "" {
		t.Errorf("expected vimshottari_lord in metadata")
	}
}
