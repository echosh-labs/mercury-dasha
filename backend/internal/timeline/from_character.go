package timeline

import (
	"fmt"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
)

// CharacterToTimelineOptions parameterizes the generation of an audiovisual presentation from a character's Dasha epoch.
type CharacterToTimelineOptions struct {
	Orientation      Orientation // "16:9", "9:16", "1:1"
	DurationSec      float64     // Total presentation duration (default: 60.0s)
	VoiceAudioPath   string      // Optional spoken narrative or character dialogue audio
	AmbientMusicPath string      // Optional atmospheric background bed
	IncludeOverlays  bool        // Whether to attach ChronoBadge, Waveform, and LowerThird
}

// GenerateFromCharacterDasha compiles a Character profile and a specific Dasha period into a complete TimelineManifest.
func GenerateFromCharacterDasha(char *dasha.DashaProfile, targetPeriod *dasha.DashaPeriod, opts CharacterToTimelineOptions) (*TimelineManifest, error) {
	if char == nil {
		return nil, fmt.Errorf("character profile cannot be nil")
	}

	// If no specific period is provided, fall back to the currently active Mahadasha
	if targetPeriod == nil {
		targetPeriod = &char.ActiveSnapshot.Mahadasha
	}

	duration := opts.DurationSec
	if duration <= 0 {
		duration = 60.0 // Standard 1-minute chronicle preview
	}

	manifestID := fmt.Sprintf("chronicle-%s-%s-%d", char.ID, targetPeriod.Planet, time.Now().Unix())
	title := fmt.Sprintf("%s: The %s Epoch (%s)", char.Name, targetPeriod.PlanetName, targetPeriod.SacredMetal)
	if char.Title != "" {
		title = fmt.Sprintf("%s — %s: The %s Epoch", char.Name, char.Title, targetPeriod.PlanetName)
	}

	manifest := NewTimelineManifest(manifestID, title)
	manifest.Description = fmt.Sprintf("Audiovisual Sovereign Chronicle for %s during the %s %s (%s). Archetype: %s. Hermetic Axiom: %s.",
		char.Name, targetPeriod.PlanetName, string(targetPeriod.Level), targetPeriod.SacredMetal, targetPeriod.StoryArchetype, targetPeriod.HermeticAxiom)

	manifest.Tags = []string{
		"character_chronicle",
		"vimshottari_dasha",
		string(targetPeriod.Planet),
		targetPeriod.SacredMetal,
		char.Astrology.JanmaNakshatra.Name,
		char.Astrology.Lagna.Rashi,
	}

	if opts.Orientation != "" {
		manifest.Canvas.Orientation = opts.Orientation
		manifest.Canvas.EnsureDefaults()
	}

	// 1. ChronoBinding Metaphysics Stamping
	manifest.ChronoMeta = &ChronoBinding{
		VimshottariLord: fmt.Sprintf("%s (%s)", targetPeriod.PlanetName, targetPeriod.SanskritName),
		PlanetaryHora:   targetPeriod.PlanetName,
		SacredMetal:     targetPeriod.SacredMetal,
		HermeticAxiom:   targetPeriod.HermeticAxiom,
		MagnumOpusStage: char.Alchemy.MagnumOpusStage.Name,
	}

	// 2. Audio Composition
	var voiceTrackID string
	if opts.VoiceAudioPath != "" {
		voiceTrackID = fmt.Sprintf("voice-%s", manifestID)
		manifest.AddAudioTrack(AudioTrack{
			ID:         voiceTrackID,
			Role:       AudioRoleVoice,
			SourcePath: opts.VoiceAudioPath,
			StartSec:   0,
			EndSec:     duration,
			Volume:     1.0,
		})
	}

	// Resonant Frequency Bed
	dronePath := opts.AmbientMusicPath
	if dronePath == "" {
		dronePath = fmt.Sprintf("drone_%.0fhz_%s.wav", targetPeriod.FrequencyHz, targetPeriod.Planet)
	}

	droneTrack := AudioTrack{
		ID:         fmt.Sprintf("drone-%s", targetPeriod.Planet),
		Role:       AudioRoleAmbientBed,
		SourcePath: dronePath,
		StartSec:   0,
		EndSec:     duration,
		Volume:     0.28,
		FadeInSec:  2.0,
		FadeOutSec: 3.0,
		Loop:       true,
	}

	if voiceTrackID != "" {
		droneTrack.Ducking = &AudioDucking{
			Enabled:       true,
			AttenuateTo:   0.12,
			AttackMs:      150,
			ReleaseMs:     700,
			SidechainFrom: voiceTrackID,
		}
	}
	manifest.AddAudioTrack(droneTrack)

	// 3. 4-Scene Narrative Arc
	sceneDur := duration / 4.0
	if sceneDur < 10.0 {
		sceneDur = duration / 2.0
	}

	// Scene 1: Cosmic Genesis & Rising Lagna
	manifest.AddScene(VisualScene{
		ID:           "scene-01-genesis",
		StartSec:     0,
		DurationSec:  sceneDur,
		SourceType:   "svg_template",
		SourcePath:   "lagna_crucible_card.svg",
		Motion:       MotionKenBurnsIn,
		MotionScale:  1.08,
		TransitionIn: TransitionFadeToBlack,
		TransitionSec: 1.5,
		Caption:      char.Name,
		Subtitle:     fmt.Sprintf("Rising Lagna: %s • Janma Moon: %s (Pada %d)", char.Astrology.Lagna.Rashi, char.Astrology.JanmaNakshatra.Name, char.Astrology.JanmaPada),
	})

	// Scene 2: Embodiment Tripod & Bhavas
	manifest.AddScene(VisualScene{
		ID:           "scene-02-tripod",
		StartSec:     sceneDur,
		DurationSec:  sceneDur,
		SourceType:   "svg_template",
		SourcePath:   "tripod_embodiment_card.svg",
		Motion:       MotionPanLeft,
		MotionScale:  1.05,
		TransitionIn: TransitionCrossfade,
		TransitionSec: 1.2,
		Caption:      "Tripod of Embodiment",
		Subtitle:     fmt.Sprintf("Surya (Soul) in %s • Chandra (Mind) in %s • Dosha: %s", char.Astrology.Tripod.Surya.Rashi, char.Astrology.Tripod.Chandra.Rashi, char.Astrology.Ayurveda.Dosha),
	})

	// Scene 3: Alchemical Epoch Crucible
	manifest.AddScene(VisualScene{
		ID:           "scene-03-alchemy",
		StartSec:     sceneDur * 2.0,
		DurationSec:  sceneDur,
		SourceType:   "svg_template",
		SourcePath:   "alchemical_crucible_card.svg",
		Motion:       MotionKenBurnsOut,
		MotionScale:  1.10,
		TransitionIn: TransitionCrossfade,
		TransitionSec: 1.2,
		Caption:      fmt.Sprintf("%s Epoch — Sacred Metal %s (%s)", targetPeriod.PlanetName, targetPeriod.SacredMetal, targetPeriod.MetalSymbol),
		Subtitle:     fmt.Sprintf("Hermetic Principle: \"%s\"", targetPeriod.HermeticAxiom),
	})

	// Scene 4: Story Archetype & Living Horizon
	manifest.AddScene(VisualScene{
		ID:           "scene-04-horizon",
		StartSec:     sceneDur * 3.0,
		DurationSec:  duration - (sceneDur * 3.0),
		SourceType:   "svg_template",
		SourcePath:   "sovereign_horizon_card.svg",
		Motion:       MotionPanRight,
		MotionScale:  1.06,
		TransitionIn: TransitionCrossfade,
		TransitionSec: 1.2,
		Caption:      fmt.Sprintf("Archetype: %s", targetPeriod.StoryArchetype),
		Subtitle:     fmt.Sprintf("Vibrational Harmonic: %.1f Hz • %s", targetPeriod.FrequencyHz, dasha.FormatAgeRange(targetPeriod.AgeStart, targetPeriod.AgeEnd)),
	})

	// 4. Overlays
	if opts.IncludeOverlays {
		// Top-right Chrono Badge
		manifest.AddOverlay(GlobalOverlay{
			ID:       "overlay-chrono-badge",
			Type:     OverlayChronoBadge,
			Position: "top_right",
			StartSec: 0,
			EndSec:   duration,
			Config: map[string]interface{}{
				"planet_name":  targetPeriod.PlanetName,
				"sacred_metal": targetPeriod.SacredMetal,
				"metal_symbol": targetPeriod.MetalSymbol,
				"color_hex":    targetPeriod.ColorHex,
			},
		})

		// Lower third speaker banner
		epithet := char.Title
		if epithet == "" {
			epithet = "Sovereign Protagonist"
		}
		manifest.AddOverlay(GlobalOverlay{
			ID:       "overlay-lower-third",
			Type:     OverlayLowerThird,
			Position: "bottom",
			StartSec: 2.0,
			EndSec:   duration - 2.0,
			Config: map[string]interface{}{
				"name":         char.Name,
				"title":        epithet,
				"epoch":        fmt.Sprintf("%s Mahadasha", targetPeriod.PlanetName),
				"age_range":    dasha.FormatAgeRange(targetPeriod.AgeStart, targetPeriod.AgeEnd),
			},
		})

		// Waveform visualizer
		manifest.AddOverlay(GlobalOverlay{
			ID:       "overlay-waveform",
			Type:     OverlayWaveform,
			Position: "bottom_center",
			StartSec: 0,
			EndSec:   duration,
			Config: map[string]interface{}{
				"color_hex": targetPeriod.ColorHex,
				"opacity":   0.85,
			},
		})
	}

	manifest.DurationSec = duration
	manifest.UpdatedAt = time.Now()

	return manifest, nil
}
