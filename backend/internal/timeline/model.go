package timeline

import (
	"time"
)

// Orientation defines the aspect ratio standard for rendering.
type Orientation string

const (
	OrientationLandscape16x9 Orientation = "16:9" // Standard YouTube / Desktop (1920x1080)
	OrientationPortrait9x16  Orientation = "9:16" // YouTube Shorts / TikTok (1080x1920)
	OrientationSquare1x1     Orientation = "1:1"  // Social / Album art (1080x1080)
)

// CanvasConfig defines dimensions, frame rate, and background color.
type CanvasConfig struct {
	Orientation     Orientation `json:"orientation"`
	Width           int         `json:"width"`            // e.g. 1920
	Height          int         `json:"height"`           // e.g. 1080
	FPS             int         `json:"fps"`              // e.g. 30
	BackgroundColor string      `json:"background_color"` // e.g. "#050811" (Mercury Midnight)
}

// EnsureDefaults applies sensible defaults to canvas parameters.
func (c *CanvasConfig) EnsureDefaults() {
	if c.Orientation == "" {
		c.Orientation = OrientationLandscape16x9
	}
	if c.FPS <= 0 {
		c.FPS = 30
	}
	if c.BackgroundColor == "" {
		c.BackgroundColor = "#050811"
	}
	if c.Width <= 0 || c.Height <= 0 {
		switch c.Orientation {
		case OrientationPortrait9x16:
			c.Width = 1080
			c.Height = 1920
		case OrientationSquare1x1:
			c.Width = 1080
			c.Height = 1080
		case OrientationLandscape16x9:
			fallthrough
		default:
			c.Width = 1920
			c.Height = 1080
		}
	}
}

// AudioRole defines the functional role of an audio track within the mix.
type AudioRole string

const (
	AudioRoleVoice      AudioRole = "voice"   // Primary spoken dialogue or narration
	AudioRoleAmbientBed AudioRole = "ambient" // Background music or atmospheric drone
	AudioRoleSFX        AudioRole = "sfx"     // Point-in-time audio cue or sound effect
)

// AudioDucking configures automatic attenuation of background audio during speech.
type AudioDucking struct {
	Enabled       bool    `json:"enabled"`
	AttenuateTo   float64 `json:"attenuate_to"`   // Multiplier when ducked (e.g. 0.20 = -14dB)
	AttackMs      int     `json:"attack_ms"`      // Speed of volume reduction in ms (e.g. 150)
	ReleaseMs     int     `json:"release_ms"`     // Speed of volume recovery in ms (e.g. 600)
	SidechainFrom string  `json:"sidechain_from"` // ID of the voice track driving the ducking
}

// AudioTrack represents an audio stream to be composited into the mix.
type AudioTrack struct {
	ID         string        `json:"id"`
	Role       AudioRole     `json:"role"`
	SourcePath string        `json:"source_path"`          // POSIX path or storehouse relative path
	StartSec   float64       `json:"start_sec"`            // Timeline start offset
	EndSec     float64       `json:"end_sec,omitempty"`    // Timeline end offset (0 = entire track)
	OffsetSec  float64       `json:"offset_sec,omitempty"` // In-file audio seek start offset
	Volume     float64       `json:"volume"`               // 1.0 = unity gain (100%), 0.5 = 50%
	FadeInSec  float64       `json:"fade_in_sec,omitempty"`
	FadeOutSec float64       `json:"fade_out_sec,omitempty"`
	Loop       bool          `json:"loop,omitempty"`
	Ducking    *AudioDucking `json:"ducking,omitempty"`
}

// MotionEffect defines cinematic motion applied to a visual scene.
type MotionEffect string

const (
	MotionNone        MotionEffect = "none"
	MotionKenBurnsIn  MotionEffect = "ken_burns_zoom_in"  // Smooth zoom in (e.g. 1.0x -> 1.10x)
	MotionKenBurnsOut MotionEffect = "ken_burns_zoom_out" // Smooth zoom out (e.g. 1.10x -> 1.0x)
	MotionPanLeft     MotionEffect = "pan_left"           // Slow pan leftwards
	MotionPanRight    MotionEffect = "pan_right"          // Slow pan rightwards
)

// TransitionType defines how a scene transitions onto the screen.
type TransitionType string

const (
	TransitionCut         TransitionType = "cut"
	TransitionCrossfade   TransitionType = "crossfade"
	TransitionFadeToBlack TransitionType = "fade_black"
	TransitionSlideLeft   TransitionType = "slide_left"
)

// VisualScene defines a discrete temporal slide or video segment.
type VisualScene struct {
	ID            string         `json:"id"`
	StartSec      float64        `json:"start_sec"`
	DurationSec   float64        `json:"duration_sec"`
	SourceType    string         `json:"source_type"` // "image", "svg_template", "color", "video"
	SourcePath    string         `json:"source_path"` // Dropbox path, local path, or template ID
	Motion        MotionEffect   `json:"motion"`
	MotionScale   float64        `json:"motion_scale,omitempty"` // e.g. 1.10
	TransitionIn  TransitionType `json:"transition_in"`
	TransitionSec float64        `json:"transition_sec,omitempty"` // Duration of transition in seconds
	Caption       string         `json:"caption,omitempty"`
	Subtitle      string         `json:"subtitle,omitempty"`
}

// OverlayType defines telemetry and visual visualizers stamped over the presentation.
type OverlayType string

const (
	OverlayWaveform    OverlayType = "waveform"     // Audio reactive waveform visualizer
	OverlayChronoBadge OverlayType = "chrono_badge" // Floating Hora / Alchemical Metal badge
	OverlayLowerThird  OverlayType = "lower_third"  // Speaker / topic banner
	OverlayWatermark   OverlayType = "watermark"    // Studio watermark / insignia
)

// GlobalOverlay defines persistent or timed UI elements stamped across scenes.
type GlobalOverlay struct {
	ID       string                 `json:"id"`
	Type     OverlayType            `json:"type"`
	Position string                 `json:"position"` // "bottom", "top_right", "bottom_left", "center"
	StartSec float64                `json:"start_sec"`
	EndSec   float64                `json:"end_sec,omitempty"` // 0 = entire presentation
	Config   map[string]interface{} `json:"config,omitempty"`
}

// ChronoBinding associates the timeline with live astronomical and alchemical state.
type ChronoBinding struct {
	VimshottariLord string `json:"vimshottari_lord"` // e.g. "Mercury (Budha)"
	PlanetaryHora   string `json:"planetary_hora"`   // e.g. "Mercury"
	SacredMetal     string `json:"sacred_metal"`     // e.g. "Quicksilver"
	HermeticAxiom   string `json:"hermetic_axiom"`   // e.g. "Principle of Polarity"
	MagnumOpusStage string `json:"magnum_opus_stage"` // e.g. "Albedo"
}

// TimelineManifest is the master declarative document describing a multi-track audiovisual render.
type TimelineManifest struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
	Canvas      CanvasConfig    `json:"canvas"`
	DurationSec float64         `json:"duration_sec"` // Total presentation length in seconds
	AudioTracks []AudioTrack    `json:"audio_tracks"`
	Scenes      []VisualScene   `json:"scenes"`
	Overlays    []GlobalOverlay `json:"overlays,omitempty"`
	ChronoMeta  *ChronoBinding  `json:"chrono_meta,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// NewTimelineManifest creates a blank timeline manifest with standard defaults.
func NewTimelineManifest(id, title string) *TimelineManifest {
	now := time.Now()
	m := &TimelineManifest{
		ID:          id,
		Title:       title,
		Tags:        make([]string, 0),
		AudioTracks: make([]AudioTrack, 0),
		Scenes:      make([]VisualScene, 0),
		Overlays:    make([]GlobalOverlay, 0),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	m.Canvas.EnsureDefaults()
	return m
}

// AddAudioTrack appends an audio track to the manifest.
func (m *TimelineManifest) AddAudioTrack(t AudioTrack) {
	if m.AudioTracks == nil {
		m.AudioTracks = make([]AudioTrack, 0)
	}
	m.AudioTracks = append(m.AudioTracks, t)
	m.UpdatedAt = time.Now()
}

// AddScene appends a visual scene to the manifest.
func (m *TimelineManifest) AddScene(s VisualScene) {
	if m.Scenes == nil {
		m.Scenes = make([]VisualScene, 0)
	}
	m.Scenes = append(m.Scenes, s)
	m.UpdatedAt = time.Now()
}

// AddOverlay appends a global overlay to the manifest.
func (m *TimelineManifest) AddOverlay(o GlobalOverlay) {
	if m.Overlays == nil {
		m.Overlays = make([]GlobalOverlay, 0)
	}
	m.Overlays = append(m.Overlays, o)
	m.UpdatedAt = time.Now()
}

// CalculateDuration determines the duration based on the furthest reaching scene or audio track.
func (m *TimelineManifest) CalculateDuration() float64 {
	var maxSec float64

	for _, s := range m.Scenes {
		end := s.StartSec + s.DurationSec
		if end > maxSec {
			maxSec = end
		}
	}

	for _, a := range m.AudioTracks {
		end := a.EndSec
		if end <= 0 && a.StartSec > 0 {
			end = a.StartSec
		}
		if end > maxSec {
			maxSec = end
		}
	}

	return maxSec
}

// EnsureDefaults applies sensible fallbacks and fills missing metadata.
func (m *TimelineManifest) EnsureDefaults() {
	if m.ID == "" {
		m.ID = "timeline-" + time.Now().Format("20060102-150405")
	}
	if m.Title == "" {
		m.Title = "Untitled Mercury Presentation"
	}
	m.Canvas.EnsureDefaults()

	if m.DurationSec <= 0 {
		m.DurationSec = m.CalculateDuration()
	}

	if m.Tags == nil {
		m.Tags = make([]string, 0)
	}
	if m.AudioTracks == nil {
		m.AudioTracks = make([]AudioTrack, 0)
	}
	if m.Scenes == nil {
		m.Scenes = make([]VisualScene, 0)
	}
	if m.Overlays == nil {
		m.Overlays = make([]GlobalOverlay, 0)
	}

	// Apply track defaults
	for i := range m.AudioTracks {
		if m.AudioTracks[i].Volume <= 0 {
			m.AudioTracks[i].Volume = 1.0
		}
		if m.AudioTracks[i].Role == "" {
			m.AudioTracks[i].Role = AudioRoleAmbientBed
		}
		if m.AudioTracks[i].Ducking != nil && m.AudioTracks[i].Ducking.Enabled {
			if m.AudioTracks[i].Ducking.AttenuateTo <= 0 {
				m.AudioTracks[i].Ducking.AttenuateTo = 0.20
			}
			if m.AudioTracks[i].Ducking.AttackMs <= 0 {
				m.AudioTracks[i].Ducking.AttackMs = 150
			}
			if m.AudioTracks[i].Ducking.ReleaseMs <= 0 {
				m.AudioTracks[i].Ducking.ReleaseMs = 600
			}
		}
	}

	// Apply scene defaults
	for i := range m.Scenes {
		if m.Scenes[i].Motion == "" {
			m.Scenes[i].Motion = MotionNone
		}
		if m.Scenes[i].MotionScale <= 0 && m.Scenes[i].Motion != MotionNone {
			m.Scenes[i].MotionScale = 1.08
		}
		if m.Scenes[i].TransitionIn == "" {
			m.Scenes[i].TransitionIn = TransitionCut
		}
		if m.Scenes[i].SourceType == "" {
			m.Scenes[i].SourceType = "image"
		}
	}

	now := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
}
