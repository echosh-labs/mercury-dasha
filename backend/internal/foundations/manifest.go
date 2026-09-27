package foundations

import (
	"fmt"

	"github.com/echosh-labs/mercury-dasha/internal/timeline"
)

// SeedToTimelineManifest transforms a StorySeed into a fully configured TimelineManifest ready for rendering.
func SeedToTimelineManifest(seed *StorySeed, orientation timeline.Orientation, durationSec float64) (*timeline.TimelineManifest, error) {
	if seed == nil {
		return nil, fmt.Errorf("story seed cannot be nil")
	}

	if durationSec <= 0 {
		durationSec = 60.0
	}

	manifestID := fmt.Sprintf("manifest-%s", seed.ID)
	title := seed.Title
	if title == "" {
		title = fmt.Sprintf("Foundations Chronicle: %s", seed.SanctuaryLabel)
	}

	manifest := timeline.NewTimelineManifest(manifestID, title)
	manifest.Description = fmt.Sprintf(
		"Audiovisual Chronicle synthesized from %s (%s). Hook: %s. Hermetic Axiom: %s. Sacred Metal: %s.",
		seed.SourcePath, seed.SanctuaryLabel, seed.NarrativeHook, seed.CosmicAlignment.HermeticAxiom, seed.CosmicAlignment.SacredMetal,
	)

	manifest.Tags = append(seed.Tags, "foundations_story_seed", "audiovisual_chronicle")

	if orientation != "" {
		manifest.Canvas.Orientation = orientation
	} else {
		manifest.Canvas.Orientation = timeline.OrientationLandscape16x9
	}
	manifest.Canvas.EnsureDefaults()

	// 1. ChronoBinding Metaphysics Stamping
	manifest.ChronoMeta = &timeline.ChronoBinding{
		VimshottariLord: fmt.Sprintf("%s (%s)", seed.CosmicAlignment.Mahadasha, seed.CosmicAlignment.MahadashaSanskrit),
		PlanetaryHora:   seed.CosmicAlignment.Antardasha,
		SacredMetal:     seed.CosmicAlignment.SacredMetal,
		HermeticAxiom:   seed.CosmicAlignment.HermeticAxiom,
		MagnumOpusStage: seed.CosmicAlignment.MagnumOpusStage,
	}

	// 2. Audio Composition
	var voiceTrackID string
	if seed.HasAudio && (seed.AudioURL != "" || seed.SourcePath != "") {
		voiceTrackID = fmt.Sprintf("voice-%s", seed.ID)
		audioPath := seed.AudioURL
		if audioPath == "" {
			audioPath = seed.SourcePath
		}

		manifest.AddAudioTrack(timeline.AudioTrack{
			ID:         voiceTrackID,
			Role:       timeline.AudioRoleVoice,
			SourcePath: audioPath,
			StartSec:   0,
			EndSec:     durationSec,
			Volume:     1.0,
		})
	}

	// Ambient Harmonic Drone Bed
	dronePath := fmt.Sprintf("drone_%.0fhz_%s.wav", seed.CosmicAlignment.ResonantFrequency, seed.CosmicAlignment.Mahadasha)
	droneTrack := timeline.AudioTrack{
		ID:         fmt.Sprintf("drone-%s", seed.CosmicAlignment.Mahadasha),
		Role:       timeline.AudioRoleAmbientBed,
		SourcePath: dronePath,
		StartSec:   0,
		EndSec:     durationSec,
		Volume:     0.26,
		FadeInSec:  2.0,
		FadeOutSec: 3.0,
		Loop:       true,
	}

	if voiceTrackID != "" {
		droneTrack.Ducking = &timeline.AudioDucking{
			Enabled:       true,
			AttenuateTo:   0.12, // -15 dB ducking
			AttackMs:      150,
			ReleaseMs:     700,
			SidechainFrom: voiceTrackID,
		}
	}
	manifest.AddAudioTrack(droneTrack)

	// 3. Visual Scenes generated from SceneBeats
	numBeats := len(seed.SceneBeats)
	if numBeats == 0 {
		numBeats = 3
	}
	sceneDur := durationSec / float64(numBeats)

	for i, beat := range seed.SceneBeats {
		start := float64(i) * sceneDur
		sceneID := fmt.Sprintf("scene-%02d-act%d", i+1, beat.Act)

		motion := timeline.MotionKenBurnsIn
		if i == 1 {
			motion = timeline.MotionPanLeft
		} else if i == 2 {
			motion = timeline.MotionKenBurnsOut
		}

		trans := timeline.TransitionCrossfade
		if i == 0 {
			trans = timeline.TransitionFadeToBlack
		}

		subtitle := beat.Premise
		if i == 0 && seed.NarrativeHook != "" {
			subtitle = seed.NarrativeHook
		} else if i == 1 && seed.DialogueAnchor != "" {
			subtitle = fmt.Sprintf("\"%s\"", seed.DialogueAnchor)
		} else if i == 2 && seed.CosmicAlignment.HermeticAxiom != "" {
			subtitle = seed.CosmicAlignment.HermeticAxiom
		}

		manifest.AddScene(timeline.VisualScene{
			ID:            sceneID,
			StartSec:      start,
			DurationSec:   sceneDur,
			SourceType:    "color",
			SourcePath:    seed.VisualDirectives.ColorPalette[0],
			Motion:        motion,
			MotionScale:   1.08,
			TransitionIn:  trans,
			TransitionSec: 1.5,
			Caption:       beat.Title,
			Subtitle:      subtitle,
		})
	}

	// 4. Overlays: ChronoBadge & Waveform
	manifest.AddOverlay(timeline.GlobalOverlay{
		ID:       "chrono-badge-top",
		Type:     timeline.OverlayChronoBadge,
		Position: "top_right",
		StartSec: 0,
		EndSec:   durationSec,
		Config: map[string]interface{}{
			"lord":  seed.CosmicAlignment.Mahadasha,
			"metal": seed.CosmicAlignment.SacredMetal,
			"axiom": seed.CosmicAlignment.HermeticAxiom,
		},
	})

	if seed.HasAudio {
		manifest.AddOverlay(timeline.GlobalOverlay{
			ID:       "waveform-bottom",
			Type:     timeline.OverlayWaveform,
			Position: "bottom",
			StartSec: 0,
			EndSec:   durationSec,
			Config: map[string]interface{}{
				"color":      seed.VisualDirectives.ColorPalette[1],
				"opacity":    0.85,
				"style":      "mirror_bars",
				"track_side": voiceTrackID,
			},
		})
	}

	manifest.DurationSec = manifest.CalculateDuration()
	return manifest, nil
}
