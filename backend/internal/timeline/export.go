package timeline

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/youtube"
)

// ToJSON serializes the timeline manifest to formatted JSON bytes.
func (m *TimelineManifest) ToJSON() ([]byte, error) {
	return json.MarshalIndent(m, "", "  ")
}

// FromJSON parses a JSON payload into a validated TimelineManifest.
func FromJSON(data []byte) (*TimelineManifest, error) {
	var m TimelineManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to unmarshal timeline JSON: %w", err)
	}
	m.EnsureDefaults()
	if err := m.Validate(); err != nil {
		return nil, fmt.Errorf("timeline manifest validation failed: %w", err)
	}
	return &m, nil
}

// ToPipelineArtifact bridges the rendered video output to the YouTube Sovereign Studio upload pipeline.
func (m *TimelineManifest) ToPipelineArtifact(renderedVideoPath string, privacyStatus string, categoryID string) youtube.PipelineArtifact {
	if privacyStatus == "" {
		privacyStatus = "private" // Safe default for studio renders
	}
	if categoryID == "" {
		categoryID = "24" // Entertainment / Gaming / Audio-Visual standard
	}

	metadata := map[string]string{
		"manifest_id":      m.ID,
		"orientation":      string(m.Canvas.Orientation),
		"resolution":       fmt.Sprintf("%dx%d", m.Canvas.Width, m.Canvas.Height),
		"duration_sec":     fmt.Sprintf("%.2f", m.DurationSec),
		"audio_tracks_num": fmt.Sprintf("%d", len(m.AudioTracks)),
		"scenes_num":       fmt.Sprintf("%d", len(m.Scenes)),
	}

	if m.ChronoMeta != nil {
		metadata["vimshottari_lord"] = m.ChronoMeta.VimshottariLord
		metadata["planetary_hora"] = m.ChronoMeta.PlanetaryHora
		metadata["sacred_metal"] = m.ChronoMeta.SacredMetal
		metadata["hermetic_axiom"] = m.ChronoMeta.HermeticAxiom
	}

	return youtube.PipelineArtifact{
		JobID:               fmt.Sprintf("render-job-%s-%d", m.ID, time.Now().Unix()),
		SourcePath:          renderedVideoPath,
		Title:               m.Title,
		Description:         m.Description,
		Tags:                m.Tags,
		CategoryID:          categoryID,
		PrivacyStatus:       privacyStatus,
		AttachChronoContext: m.ChronoMeta != nil,
		Metadata:            metadata,
	}
}
