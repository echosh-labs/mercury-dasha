package timeline

import (
	"strings"
	"testing"
)

func TestNewTimelineManifest(t *testing.T) {
	m := NewTimelineManifest("manifest-001", "Alchemical Transmutation")
	if m.ID != "manifest-001" {
		t.Fatalf("expected ID manifest-001, got %s", m.ID)
	}
	if m.Title != "Alchemical Transmutation" {
		t.Fatalf("expected Title Alchemical Transmutation, got %s", m.Title)
	}
	if m.Canvas.Orientation != OrientationLandscape16x9 {
		t.Errorf("expected 16:9, got %s", m.Canvas.Orientation)
	}
	if m.Canvas.Width != 1920 || m.Canvas.Height != 1080 {
		t.Errorf("expected 1920x1080, got %dx%d", m.Canvas.Width, m.Canvas.Height)
	}
	if m.Canvas.FPS != 30 {
		t.Errorf("expected 30 fps, got %d", m.Canvas.FPS)
	}
}

func TestCanvasOrientations(t *testing.T) {
	tests := []struct {
		orientation Orientation
		wantW       int
		wantH       int
	}{
		{OrientationLandscape16x9, 1920, 1080},
		{OrientationPortrait9x16, 1080, 1920},
		{OrientationSquare1x1, 1080, 1080},
	}

	for _, tt := range tests {
		cfg := CanvasConfig{Orientation: tt.orientation}
		cfg.EnsureDefaults()
		if cfg.Width != tt.wantW || cfg.Height != tt.wantH {
			t.Errorf("orientation %s: want %dx%d, got %dx%d", tt.orientation, tt.wantW, tt.wantH, cfg.Width, cfg.Height)
		}
	}
}

func TestCalculateDuration(t *testing.T) {
	m := NewTimelineManifest("m1", "Duration Test")
	m.AddScene(VisualScene{
		ID:          "s1",
		StartSec:    0,
		DurationSec: 30,
		SourcePath:  "image1.png",
	})
	m.AddScene(VisualScene{
		ID:          "s2",
		StartSec:    30,
		DurationSec: 45,
		SourcePath:  "image2.png",
	})
	m.AddAudioTrack(AudioTrack{
		ID:         "a1",
		SourcePath: "ambient.mp3",
		StartSec:   0,
		EndSec:     85,
	})

	dur := m.CalculateDuration()
	if dur != 85.0 {
		t.Errorf("expected duration 85.0, got %f", dur)
	}
}

func TestValidateSuccess(t *testing.T) {
	m := NewTimelineManifest("m-valid", "Valid Manifest")
	m.AddAudioTrack(AudioTrack{
		ID:         "voice-1",
		Role:       AudioRoleVoice,
		SourcePath: "/path/to/voice.wav",
		StartSec:   0,
		EndSec:     60,
		Volume:     1.0,
	})
	m.AddAudioTrack(AudioTrack{
		ID:         "ambient-1",
		Role:       AudioRoleAmbientBed,
		SourcePath: "/path/to/ambient.mp3",
		StartSec:   0,
		EndSec:     60,
		Volume:     0.3,
		Ducking: &AudioDucking{
			Enabled:       true,
			AttenuateTo:   0.15,
			AttackMs:      150,
			ReleaseMs:     600,
			SidechainFrom: "voice-1",
		},
	})
	m.AddScene(VisualScene{
		ID:            "scene-1",
		StartSec:      0,
		DurationSec:   60,
		SourceType:    "image",
		SourcePath:    "/path/to/art.png",
		Motion:        MotionKenBurnsIn,
		MotionScale:   1.10,
		TransitionIn:  TransitionFadeToBlack,
		TransitionSec: 1.0,
	})
	m.EnsureDefaults()

	if err := m.Validate(); err != nil {
		t.Fatalf("expected valid manifest, got error: %v", err)
	}
}

func TestValidateErrors(t *testing.T) {
	// 1. Missing ID and Title
	m := &TimelineManifest{}
	err := m.Validate()
	if err == nil {
		t.Fatal("expected validation error for empty manifest")
	}

	// 2. Invalid Ducking Sidechain Reference
	m2 := NewTimelineManifest("m2", "Invalid Ducking")
	m2.AddAudioTrack(AudioTrack{
		ID:         "ambient-1",
		Role:       AudioRoleAmbientBed,
		SourcePath: "ambient.mp3",
		StartSec:   0,
		EndSec:     30,
		Volume:     0.5,
		Ducking: &AudioDucking{
			Enabled:       true,
			SidechainFrom: "non-existent-track",
			AttenuateTo:   0.2,
		},
	})
	m2.AddScene(VisualScene{
		ID:          "s1",
		StartSec:    0,
		DurationSec: 30,
		SourcePath:  "slide.png",
	})
	m2.EnsureDefaults()

	err2 := m2.Validate()
	if err2 == nil || !strings.Contains(err2.Error(), "does not exist in audio_tracks") {
		t.Fatalf("expected sidechain error, got: %v", err2)
	}

	// 3. Self-ducking error
	m3 := NewTimelineManifest("m3", "Self Ducking")
	m3.AddAudioTrack(AudioTrack{
		ID:         "track-1",
		Role:       AudioRoleVoice,
		SourcePath: "voice.wav",
		StartSec:   0,
		EndSec:     30,
		Volume:     1.0,
		Ducking: &AudioDucking{
			Enabled:       true,
			SidechainFrom: "track-1",
			AttenuateTo:   0.2,
		},
	})
	m3.AddScene(VisualScene{
		ID:          "s1",
		StartSec:    0,
		DurationSec: 30,
		SourcePath:  "slide.png",
	})
	m3.EnsureDefaults()

	err3 := m3.Validate()
	if err3 == nil || !strings.Contains(err3.Error(), "cannot sidechain duck against itself") {
		t.Fatalf("expected self-ducking error, got: %v", err3)
	}
}

func TestJSONRoundtrip(t *testing.T) {
	m := NewTimelineManifest("roundtrip-01", "Hermetic Symphony")
	m.AddAudioTrack(AudioTrack{
		ID:         "voice",
		Role:       AudioRoleVoice,
		SourcePath: "voice.wav",
		StartSec:   0,
		EndSec:     120,
		Volume:     1.0,
	})
	m.AddScene(VisualScene{
		ID:          "scene-01",
		StartSec:    0,
		DurationSec: 120,
		SourceType:  "image",
		SourcePath:  "hermetic.png",
		Motion:      MotionKenBurnsIn,
	})
	m.ChronoMeta = &ChronoBinding{
		VimshottariLord: "Mercury",
		PlanetaryHora:   "Mercury",
		SacredMetal:     "Quicksilver",
		HermeticAxiom:   "Principle of Polarity",
		MagnumOpusStage: "Albedo",
	}
	m.EnsureDefaults()

	data, err := m.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	parsed, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if parsed.ID != m.ID || parsed.Title != m.Title {
		t.Errorf("manifest mismatch after roundtrip: got ID %s, Title %s", parsed.ID, parsed.Title)
	}
	if parsed.ChronoMeta == nil || parsed.ChronoMeta.SacredMetal != "Quicksilver" {
		t.Errorf("chrono meta missing or altered after roundtrip: %+v", parsed.ChronoMeta)
	}
}

func TestToPipelineArtifact(t *testing.T) {
	m := NewTimelineManifest("pipe-01", "Sovereign Render")
	m.Description = "Broadcast ready output"
	m.Tags = []string{"alchemy", "audio-visual"}
	m.ChronoMeta = &ChronoBinding{
		PlanetaryHora: "Jupiter",
		SacredMetal:   "Tin",
	}
	m.EnsureDefaults()

	artifact := m.ToPipelineArtifact("/tmp/rendered.mp4", "unlisted", "28")

	if artifact.SourcePath != "/tmp/rendered.mp4" {
		t.Errorf("expected /tmp/rendered.mp4, got %s", artifact.SourcePath)
	}
	if artifact.PrivacyStatus != "unlisted" {
		t.Errorf("expected unlisted, got %s", artifact.PrivacyStatus)
	}
	if artifact.CategoryID != "28" {
		t.Errorf("expected 28, got %s", artifact.CategoryID)
	}
	if !artifact.AttachChronoContext {
		t.Error("expected AttachChronoContext to be true when ChronoMeta is present")
	}
	if artifact.Metadata["manifest_id"] != "pipe-01" {
		t.Errorf("expected pipe-01 in metadata, got %s", artifact.Metadata["manifest_id"])
	}
}
