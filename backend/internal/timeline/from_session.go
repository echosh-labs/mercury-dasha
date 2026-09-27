package timeline

import (
	"fmt"
	"sort"

	"github.com/echosh-labs/mercury-dasha/internal/session"
)

// SessionToTimelineOptions parameterizes the conversion of an AudioSession into a presentation.
type SessionToTimelineOptions struct {
	AmbientMusicPath string            // Optional path to atmospheric background audio
	DefaultImagePath string            // Default slide image for segments without specific art
	CategoryArtMap   map[string]string // Mapping of marker categories ("combat", "lore") to image paths
	Orientation      Orientation       // 16:9, 9:16, or 1:1
	EnableDucking    bool              // Auto-duck ambient music during session voice
}

// GenerateFromSession compiles an AudioSession and its timeline markers into a TimelineManifest.
func GenerateFromSession(sess *session.AudioSession, opts SessionToTimelineOptions) (*TimelineManifest, error) {
	if sess == nil {
		return nil, fmt.Errorf("session cannot be nil")
	}

	manifestID := fmt.Sprintf("timeline-%s", sess.ID)
	manifest := NewTimelineManifest(manifestID, sess.Title)
	manifest.Description = sess.Notes
	manifest.Tags = append(manifest.Tags, sess.Tags...)
	manifest.Tags = append(manifest.Tags, "session_render", sess.Campaign)

	if opts.Orientation != "" {
		manifest.Canvas.Orientation = opts.Orientation
		manifest.Canvas.EnsureDefaults()
	}

	// 1. Primary Voice Audio Track from Session
	voiceTrackID := fmt.Sprintf("voice-%s", sess.ID)
	voiceTrack := AudioTrack{
		ID:         voiceTrackID,
		Role:       AudioRoleVoice,
		SourcePath: sess.FilePath,
		StartSec:   0,
		EndSec:     sess.DurationSec,
		Volume:     1.0,
	}
	manifest.AddAudioTrack(voiceTrack)

	// 2. Ambient Background Music (Optional)
	if opts.AmbientMusicPath != "" {
		ambientTrack := AudioTrack{
			ID:         "ambient-bed",
			Role:       AudioRoleAmbientBed,
			SourcePath: opts.AmbientMusicPath,
			StartSec:   0,
			EndSec:     sess.DurationSec,
			Volume:     0.25,
			FadeInSec:  2.0,
			FadeOutSec: 3.0,
			Loop:       true,
		}

		if opts.EnableDucking {
			ambientTrack.Ducking = &AudioDucking{
				Enabled:       true,
				AttenuateTo:   0.15,
				AttackMs:      150,
				ReleaseMs     : 700,
				SidechainFrom: voiceTrackID,
			}
		}

		manifest.AddAudioTrack(ambientTrack)
	}

	// 3. Visual Scenes from Session Markers
	fallbackImage := opts.DefaultImagePath
	if fallbackImage == "" {
		fallbackImage = "default_mercury_card.png"
	}

	// Sort markers chronologically
	sortedMarkers := make([]session.SessionMarker, len(sess.Markers))
	copy(sortedMarkers, sess.Markers)
	sort.Slice(sortedMarkers, func(i, j int) bool {
		return sortedMarkers[i].TimestampMs < sortedMarkers[j].TimestampMs
	})

	motions := []MotionEffect{
		MotionKenBurnsIn,
		MotionPanLeft,
		MotionKenBurnsOut,
		MotionPanRight,
	}

	totalDuration := sess.DurationSec
	if totalDuration <= 0 && len(sortedMarkers) > 0 {
		totalDuration = float64(sortedMarkers[len(sortedMarkers)-1].TimestampMs)/1000.0 + 30.0
	}

	if len(sortedMarkers) == 0 {
		// Single continuous scene spanning entire audio
		sceneDur := totalDuration
		if sceneDur <= 0 {
			sceneDur = 60.0 // sensible fallback
		}
		manifest.AddScene(VisualScene{
			ID:           "scene-01-full",
			StartSec:     0,
			DurationSec:  sceneDur,
			SourceType:   "image",
			SourcePath:   fallbackImage,
			Motion:       MotionKenBurnsIn,
			MotionScale:  1.08,
			TransitionIn: TransitionFadeToBlack,
			Caption:      sess.Title,
			Subtitle:     sess.Campaign,
		})
	} else {
		for i, marker := range sortedMarkers {
			startSec := float64(marker.TimestampMs) / 1000.0
			var durationSec float64

			if i < len(sortedMarkers)-1 {
				nextStartSec := float64(sortedMarkers[i+1].TimestampMs) / 1000.0
				durationSec = nextStartSec - startSec
			} else {
				durationSec = totalDuration - startSec
			}

			if durationSec <= 0 {
				durationSec = 10.0 // minimum slide duration
			}

			// Resolve artwork by category
			imagePath := fallbackImage
			if opts.CategoryArtMap != nil {
				if mapped, ok := opts.CategoryArtMap[marker.Category]; ok && mapped != "" {
					imagePath = mapped
				}
			}

			motion := motions[i%len(motions)]
			trans := TransitionCrossfade
			if i == 0 {
				trans = TransitionFadeToBlack
			}

			scene := VisualScene{
				ID:            fmt.Sprintf("scene-%02d-%s", i+1, marker.Category),
				StartSec:      startSec,
				DurationSec:   durationSec,
				SourceType:    "image",
				SourcePath:    imagePath,
				Motion:        motion,
				MotionScale:   1.08,
				TransitionIn:  trans,
				TransitionSec: 1.0,
				Caption:       marker.Label,
				Subtitle:      marker.Notes,
			}
			manifest.AddScene(scene)
		}
	}

	// 4. Overlays: Reactive Audio Waveform
	manifest.AddOverlay(GlobalOverlay{
		ID:       "overlay-waveform",
		Type:     OverlayWaveform,
		Position: "bottom",
		StartSec: 0,
		EndSec:   totalDuration,
		Config: map[string]interface{}{
			"color":      "#06b6d4",
			"height_px":  64,
			"opacity":    0.85,
			"source_ref": voiceTrackID,
		},
	})

	manifest.EnsureDefaults()
	if err := manifest.Validate(); err != nil {
		return nil, fmt.Errorf("generated manifest validation failed: %w", err)
	}

	return manifest, nil
}
