package timeline

import (
	"testing"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/session"
)

func TestGenerateFromSessionWithMarkers(t *testing.T) {
	now := time.Now()
	sess := &session.AudioSession{
		ID:          "session-dnd-042",
		Title:       "The Tomb of Horrors: Chamber of Quicksilver",
		Campaign:    "Ecosystem Chronicles",
		DurationSec: 180.0,
		FilePath:    "/home/justin/Dropbox/sessions/session-dnd-042.wav",
		Notes:       "An epic confrontation with the Alchemical Golem.",
		Tags:        []string{"dnd", "boss_fight"},
		Markers: []session.SessionMarker{
			{
				ID:          "marker-1",
				TimestampMs: 0,
				WallTime:    now,
				Label:       "Entering the Sanctum",
				Category:    "lore",
				Notes:       "Ancient runes glow along the obsidian walls.",
			},
			{
				ID:          "marker-2",
				TimestampMs: 45000, // 45 sec
				WallTime:    now.Add(45 * time.Second),
				Label:       "The Golem Awakens!",
				Category:    "combat",
				Notes:       "Initiative is rolled.",
			},
			{
				ID:          "marker-3",
				TimestampMs: 120000, // 120 sec
				WallTime:    now.Add(120 * time.Second),
				Label:       "Shattering the Core & Looting",
				Category:    "loot",
				Notes:       "The Philosopher's Stone is retrieved.",
			},
		},
	}

	opts := SessionToTimelineOptions{
		AmbientMusicPath: "/home/justin/Dropbox/audio/dungeon_ambience.mp3",
		DefaultImagePath: "/home/justin/Dropbox/art/generic_dungeon.png",
		CategoryArtMap: map[string]string{
			"lore":   "/home/justin/Dropbox/art/ancient_scroll.png",
			"combat": "/home/justin/Dropbox/art/golem_battle.png",
			"loot":   "/home/justin/Dropbox/art/philosophers_stone.png",
		},
		Orientation:   OrientationLandscape16x9,
		EnableDucking: true,
	}

	manifest, err := GenerateFromSession(sess, opts)
	if err != nil {
		t.Fatalf("GenerateFromSession failed: %v", err)
	}

	// 1. Check Audio Tracks
	if len(manifest.AudioTracks) != 2 {
		t.Fatalf("expected 2 audio tracks (voice + ambient), got %d", len(manifest.AudioTracks))
	}
	voiceTrack := manifest.AudioTracks[0]
	if voiceTrack.Role != AudioRoleVoice || voiceTrack.SourcePath != sess.FilePath {
		t.Errorf("voice track misconfigured: %+v", voiceTrack)
	}
	ambientTrack := manifest.AudioTracks[1]
	if ambientTrack.Role != AudioRoleAmbientBed || ambientTrack.Ducking == nil {
		t.Errorf("ambient track misconfigured: %+v", ambientTrack)
	}
	if !ambientTrack.Ducking.Enabled || ambientTrack.Ducking.SidechainFrom != voiceTrack.ID {
		t.Errorf("ducking sidechain link broken: %+v", ambientTrack.Ducking)
	}

	// 2. Check Scenes
	if len(manifest.Scenes) != 3 {
		t.Fatalf("expected 3 scenes from markers, got %d", len(manifest.Scenes))
	}

	// Scene 1: Lore (0s - 45s)
	s1 := manifest.Scenes[0]
	if s1.StartSec != 0 || s1.DurationSec != 45.0 {
		t.Errorf("scene 1 timing mismatch: start=%.1f, dur=%.1f", s1.StartSec, s1.DurationSec)
	}
	if s1.SourcePath != "/home/justin/Dropbox/art/ancient_scroll.png" {
		t.Errorf("scene 1 art mismatch: %s", s1.SourcePath)
	}
	if s1.Caption != "Entering the Sanctum" {
		t.Errorf("scene 1 caption mismatch: %s", s1.Caption)
	}

	// Scene 2: Combat (45s - 120s => 75s duration)
	s2 := manifest.Scenes[1]
	if s2.StartSec != 45.0 || s2.DurationSec != 75.0 {
		t.Errorf("scene 2 timing mismatch: start=%.1f, dur=%.1f", s2.StartSec, s2.DurationSec)
	}
	if s2.SourcePath != "/home/justin/Dropbox/art/golem_battle.png" {
		t.Errorf("scene 2 art mismatch: %s", s2.SourcePath)
	}

	// Scene 3: Loot (120s - 180s => 60s duration)
	s3 := manifest.Scenes[2]
	if s3.StartSec != 120.0 || s3.DurationSec != 60.0 {
		t.Errorf("scene 3 timing mismatch: start=%.1f, dur=%.1f", s3.StartSec, s3.DurationSec)
	}
	if s3.SourcePath != "/home/justin/Dropbox/art/philosophers_stone.png" {
		t.Errorf("scene 3 art mismatch: %s", s3.SourcePath)
	}

	// 3. Check Overlays
	if len(manifest.Overlays) != 1 || manifest.Overlays[0].Type != OverlayWaveform {
		t.Errorf("expected waveform overlay, got %+v", manifest.Overlays)
	}

	// 4. Manifest Integrity
	if err := manifest.Validate(); err != nil {
		t.Fatalf("generated manifest failed validation: %v", err)
	}
}

func TestGenerateFromSessionNoMarkers(t *testing.T) {
	sess := &session.AudioSession{
		ID:          "session-short-01",
		Title:       "Quick Monologue",
		DurationSec: 50.0,
		FilePath:    "/path/to/mono.wav",
	}

	opts := SessionToTimelineOptions{
		DefaultImagePath: "/art/default.png",
	}

	manifest, err := GenerateFromSession(sess, opts)
	if err != nil {
		t.Fatalf("failed without markers: %v", err)
	}

	if len(manifest.Scenes) != 1 {
		t.Fatalf("expected 1 scene, got %d", len(manifest.Scenes))
	}
	if manifest.Scenes[0].DurationSec != 50.0 {
		t.Errorf("expected scene duration 50.0, got %f", manifest.Scenes[0].DurationSec)
	}
	if manifest.Scenes[0].SourcePath != "/art/default.png" {
		t.Errorf("expected default art, got %s", manifest.Scenes[0].SourcePath)
	}
}

func TestGenerateFromSessionNil(t *testing.T) {
	_, err := GenerateFromSession(nil, SessionToTimelineOptions{})
	if err == nil {
		t.Fatal("expected error for nil session")
	}
}
