package session

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrSessionNotFound   = errors.New("audio session not found")
	ErrInvalidSessionID  = errors.New("invalid or empty session id")
	ErrInvalidSliceRange = errors.New("invalid slice range: start_ms must be less than end_ms and within session duration")
	ErrUnsupportedFormat = errors.New("unsupported audio format for precision slicing")
)

type SessionMarker struct {
	ID          string    `json:"id"`
	TimestampMs int64     `json:"timestamp_ms"` // Offset in milliseconds from session start
	WallTime    time.Time `json:"wall_time"`
	Label       string    `json:"label"`
	Category    string    `json:"category"` // "combat", "roleplay", "crit", "lore", "loot", "rest", "general"
	Notes       string    `json:"notes,omitempty"`
}

type AudioSlice struct {
	ID          string    `json:"id"`
	SessionID   string    `json:"session_id"`
	Label       string    `json:"label"`
	Category    string    `json:"category,omitempty"`
	StartMs     int64     `json:"start_ms"`
	EndMs       int64     `json:"end_ms"`
	DurationSec float64   `json:"duration_sec"`
	SampleRate  int       `json:"sample_rate"`
	Channels    int       `json:"channels"`
	BitDepth    int       `json:"bit_depth"`
	Format      string    `json:"format"` // "wav"
	SizeBytes   int64     `json:"size_bytes"`
	FilePath    string    `json:"file_path,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type AudioSession struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Campaign    string          `json:"campaign"`
	DM          string          `json:"dm,omitempty"`
	StartTime   time.Time       `json:"start_time"`
	EndTime     time.Time       `json:"end_time,omitempty"`
	DurationSec float64         `json:"duration_sec"`
	Status      string          `json:"status"` // "recording", "completed", "archived"
	Format      string          `json:"format"` // "audio/wav", "audio/webm", etc.
	SampleRate  int             `json:"sample_rate"` // e.g. 48000
	Channels    int             `json:"channels"`    // e.g. 2
	BitDepth    int             `json:"bit_depth"`   // e.g. 16, 24
	InputDevice string          `json:"input_device"`
	FilePath    string          `json:"file_path"` // relative path within storehouse
	SizeBytes   int64           `json:"size_bytes"`
	Markers     []SessionMarker `json:"markers"`
	Slices      []AudioSlice    `json:"slices,omitempty"`
	Notes       string          `json:"notes,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// EnsureDefaults fills missing required fields with sensible studio defaults.
func (s *AudioSession) EnsureDefaults() {
	if s.ID == "" {
		s.ID = "session-" + time.Now().Format("20060102-150405")
	}
	if s.Title == "" {
		s.Title = "D&D Session - " + time.Now().Format("Jan 02, 2006")
	}
	if s.Campaign == "" {
		s.Campaign = "Default Campaign"
	}
	if s.Status == "" {
		s.Status = "recording"
	}
	if s.Format == "" {
		s.Format = "audio/wav"
	}
	if s.SampleRate <= 0 {
		s.SampleRate = 48000 // Studio standard for video / OBS / Steinberg UR-44
	}
	if s.Channels <= 0 {
		s.Channels = 2 // Stereo
	}
	if s.BitDepth <= 0 {
		s.BitDepth = 16 // Standard PCM, 24 supported
	}
	if s.InputDevice == "" {
		s.InputDevice = "Steinberg UR-44 (Line In)"
	}
	if s.StartTime.IsZero() {
		s.StartTime = time.Now()
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}
	s.UpdatedAt = time.Now()
	if s.Markers == nil {
		s.Markers = make([]SessionMarker, 0)
	}
	if s.Slices == nil {
		s.Slices = make([]AudioSlice, 0)
	}
	if s.Tags == nil {
		s.Tags = make([]string, 0)
	}
}

// SanitizeFilename cleans the title or id for safe filesystem storage.
func SanitizeFilename(name string) string {
	clean := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		if r == ' ' {
			return '_'
		}
		return -1
	}, name)
	if clean == "" {
		return "session_audio"
	}
	return clean
}
