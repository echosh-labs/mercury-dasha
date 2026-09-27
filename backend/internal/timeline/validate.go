package timeline

import (
	"errors"
	"fmt"
	"strings"
)

// ValidationError aggregates individual validation failures.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a slice of ValidationError that implements error.
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var msgs []string
	for _, err := range ve {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// Validate inspects a TimelineManifest for structural, temporal, and referential integrity.
func (m *TimelineManifest) Validate() error {
	var errs ValidationErrors

	if strings.TrimSpace(m.ID) == "" {
		errs = append(errs, ValidationError{Field: "id", Message: "manifest ID cannot be empty"})
	}

	if strings.TrimSpace(m.Title) == "" {
		errs = append(errs, ValidationError{Field: "title", Message: "manifest title cannot be empty"})
	}

	// Canvas validation
	if m.Canvas.Width <= 0 || m.Canvas.Height <= 0 {
		errs = append(errs, ValidationError{Field: "canvas.resolution", Message: fmt.Sprintf("invalid canvas resolution: %dx%d", m.Canvas.Width, m.Canvas.Height)})
	}
	if m.Canvas.FPS <= 0 {
		errs = append(errs, ValidationError{Field: "canvas.fps", Message: fmt.Sprintf("invalid framerate: %d", m.Canvas.FPS)})
	}
	switch m.Canvas.Orientation {
	case OrientationLandscape16x9, OrientationPortrait9x16, OrientationSquare1x1:
		// valid
	case "":
		errs = append(errs, ValidationError{Field: "canvas.orientation", Message: "orientation cannot be empty"})
	default:
		errs = append(errs, ValidationError{Field: "canvas.orientation", Message: fmt.Sprintf("unsupported orientation: %s", m.Canvas.Orientation)})
	}

	// Audio Track validation & Sidechain reference index
	trackIDs := make(map[string]bool)
	for i, track := range m.AudioTracks {
		prefix := fmt.Sprintf("audio_tracks[%d]", i)
		if strings.TrimSpace(track.ID) == "" {
			errs = append(errs, ValidationError{Field: prefix + ".id", Message: "track ID cannot be empty"})
		} else {
			if trackIDs[track.ID] {
				errs = append(errs, ValidationError{Field: prefix + ".id", Message: fmt.Sprintf("duplicate track ID: %s", track.ID)})
			}
			trackIDs[track.ID] = true
		}

		if strings.TrimSpace(track.SourcePath) == "" {
			errs = append(errs, ValidationError{Field: prefix + ".source_path", Message: "audio source_path cannot be empty"})
		}

		if track.StartSec < 0 {
			errs = append(errs, ValidationError{Field: prefix + ".start_sec", Message: "start_sec cannot be negative"})
		}
		if track.EndSec > 0 && track.EndSec <= track.StartSec {
			errs = append(errs, ValidationError{Field: prefix + ".end_sec", Message: "end_sec must be strictly greater than start_sec"})
		}
		if track.Volume < 0 {
			errs = append(errs, ValidationError{Field: prefix + ".volume", Message: "volume cannot be negative"})
		}
		if track.FadeInSec < 0 {
			errs = append(errs, ValidationError{Field: prefix + ".fade_in_sec", Message: "fade_in_sec cannot be negative"})
		}
		if track.FadeOutSec < 0 {
			errs = append(errs, ValidationError{Field: prefix + ".fade_out_sec", Message: "fade_out_sec cannot be negative"})
		}
	}

	// Secondary ducking sidechain reference check
	for i, track := range m.AudioTracks {
		if track.Ducking != nil && track.Ducking.Enabled {
			prefix := fmt.Sprintf("audio_tracks[%d].ducking", i)
			if strings.TrimSpace(track.Ducking.SidechainFrom) == "" {
				errs = append(errs, ValidationError{Field: prefix + ".sidechain_from", Message: "sidechain_from track ID cannot be empty when ducking is enabled"})
			} else if !trackIDs[track.Ducking.SidechainFrom] {
				errs = append(errs, ValidationError{Field: prefix + ".sidechain_from", Message: fmt.Sprintf("referenced sidechain track '%s' does not exist in audio_tracks", track.Ducking.SidechainFrom)})
			} else if track.Ducking.SidechainFrom == track.ID {
				errs = append(errs, ValidationError{Field: prefix + ".sidechain_from", Message: "track cannot sidechain duck against itself"})
			}

			if track.Ducking.AttenuateTo <= 0 || track.Ducking.AttenuateTo > 1.0 {
				errs = append(errs, ValidationError{Field: prefix + ".attenuate_to", Message: "attenuate_to must be between 0.0 (exclusive) and 1.0 (inclusive)"})
			}
		}
	}

	// Visual Scene validation
	for i, scene := range m.Scenes {
		prefix := fmt.Sprintf("scenes[%d]", i)
		if strings.TrimSpace(scene.ID) == "" {
			errs = append(errs, ValidationError{Field: prefix + ".id", Message: "scene ID cannot be empty"})
		}
		if scene.StartSec < 0 {
			errs = append(errs, ValidationError{Field: prefix + ".start_sec", Message: "start_sec cannot be negative"})
		}
		if scene.DurationSec <= 0 {
			errs = append(errs, ValidationError{Field: prefix + ".duration_sec", Message: "duration_sec must be greater than 0"})
		}
		if strings.TrimSpace(scene.SourcePath) == "" {
			errs = append(errs, ValidationError{Field: prefix + ".source_path", Message: "scene source_path cannot be empty"})
		}
		if scene.TransitionSec > 0 && scene.TransitionSec > scene.DurationSec {
			errs = append(errs, ValidationError{Field: prefix + ".transition_sec", Message: "transition_sec cannot exceed scene duration"})
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// CheckEmptyTimeline confirms if the manifest contains either scenes or audio.
func (m *TimelineManifest) CheckEmptyTimeline() error {
	if len(m.Scenes) == 0 && len(m.AudioTracks) == 0 {
		return errors.New("timeline manifest must contain at least one visual scene or audio track")
	}
	return nil
}
