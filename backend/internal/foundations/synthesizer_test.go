package foundations

import (
	"testing"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
	"github.com/echosh-labs/mercury-dasha/internal/timeline"
)

func TestSynthesizeSeedFromText(t *testing.T) {
	// Create mock profile and timeline
	prof, err := dasha.CalculateFromRequest(dasha.CalculationRequest{
		BirthDate:      "1982-11-23",
		BirthTime:      "16:45",
		TimezoneOffset: -5,
		NakshatraIndex: 24,
		PadaNumber:     4,
	})
	if err != nil {
		t.Fatalf("failed to calculate dasha profile: %v", err)
	}
	tl := prof.Timeline

	content := `The Principle of Cause and Effect governs all manifestation in the mortal plane.
Every cause has its effect; every effect has its cause. Chance is simply a law not recognized.
We must remember that the true master polarizes at the mental plane of causation, refusing to be mere pawns moved across the board by environmental tides.
In this crucible of discipline, we forge the sovereign will.`

	input := SeedInput{
		Title:          "The Law of Causation",
		SourcePath:     "text/hermetic/cause_effect.txt",
		SourceType:     "written_note",
		Content:        content,
		HasAudio:       true,
		AudioURL:       "/api/v1/index/content?path=audio/cause_effect.m4a",
		SanctuaryID:    "hermetic_kybalion",
		SanctuaryLabel: "Hermetic & Kybalion Studies",
		NoteDate:       "2013-03-15",
		Year:           2013,
	}

	seed, err := SynthesizeSeed(input, prof, tl)
	if err != nil {
		t.Fatalf("SynthesizeSeed returned unexpected error: %v", err)
	}

	if seed.ID == "" {
		t.Errorf("expected non-empty seed ID")
	}
	if seed.NarrativeHook == "" {
		t.Errorf("expected extracted narrative hook, got empty string")
	}
	if seed.DialogueAnchor == "" {
		t.Errorf("expected dialogue anchor, got empty string")
	}
	if len(seed.SceneBeats) != 3 {
		t.Errorf("expected 3 scene beats, got %d", len(seed.SceneBeats))
	}
	if seed.CosmicAlignment.Mahadasha == "" {
		t.Errorf("expected cosmic Mahadasha to be populated")
	}
	if seed.CosmicAlignment.SacredMetal == "" {
		t.Errorf("expected cosmic SacredMetal to be populated")
	}

	t.Logf("Synthesized Seed: ID=%s, Hook=%s, Graha=%s, Metal=%s, Axiom=%s",
		seed.ID, seed.NarrativeHook, seed.CosmicAlignment.Mahadasha, seed.CosmicAlignment.SacredMetal, seed.CosmicAlignment.HermeticAxiom)

	// Test Manifest generation
	manifest, err := SeedToTimelineManifest(seed, timeline.OrientationLandscape16x9, 60.0)
	if err != nil {
		t.Fatalf("SeedToTimelineManifest returned error: %v", err)
	}

	if len(manifest.Scenes) != 3 {
		t.Errorf("expected 3 scenes in manifest, got %d", len(manifest.Scenes))
	}
	if len(manifest.AudioTracks) != 2 {
		t.Errorf("expected 2 audio tracks (voice + ambient drone), got %d", len(manifest.AudioTracks))
	}
	if manifest.ChronoMeta == nil {
		t.Errorf("expected ChronoMeta on manifest")
	}
}

func TestExtractNarrativeHookAndAnchor(t *testing.T) {
	text := `# Daily Journal
Date: 2019-07-22

The soul expands only when it embraces voluntary struggle and refuses the sedative of complacency.
"To know the divine order is to align our speech with silence."
We continue the great work under the midnight sun.`

	hook := ExtractNarrativeHook(text)
	if hook != "The soul expands only when it embraces voluntary struggle and refuses the sedative of complacency" {
		t.Errorf("unexpected hook: %q", hook)
	}

	anchor := ExtractDialogueAnchor(text)
	if anchor != "To know the divine order is to align our speech with silence." {
		t.Errorf("unexpected anchor: %q", anchor)
	}
}
