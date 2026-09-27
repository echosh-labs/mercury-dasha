package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/chrono"
	"github.com/echosh-labs/mercury-dasha/internal/dasha"
	"github.com/echosh-labs/mercury-dasha/internal/db"
	"github.com/echosh-labs/mercury-dasha/internal/timeline"
)

// ChronicleRequest defines the payload for compiling an audiovisual timeline for a character's Dasha epoch.
type ChronicleRequest struct {
	DashaPlanet      string  `json:"dasha_planet,omitempty"`       // Target planet (e.g. "jupiter", "mercury", "mars")
	Orientation      string  `json:"orientation,omitempty"`        // "16:9", "9:16", "1:1"
	DurationSec      float64 `json:"duration_sec,omitempty"`       // Total duration in seconds
	VoiceAudioPath   string  `json:"voice_audio_path,omitempty"`   // Voice narration file path
	AmbientMusicPath string  `json:"ambient_music_path,omitempty"` // Atmospheric background bed path
}

// ListCharactersHandler returns all stored character profiles with unified metadata.
func (h *Handler) ListCharactersHandler(w http.ResponseWriter, r *http.Request) {
	summaries, err := h.engine.ListSovereignProfileSummaries()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to list characters: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	keys := make([]string, 0, len(summaries))
	for _, s := range summaries {
		keys = append(keys, s.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"characters": summaries,
		"profiles":   keys,
		"summaries":  summaries,
		"count":      len(summaries),
	})
}

// GetCharacterHandler retrieves a character, their enriched timeline, or saved chronicles.
func (h *Handler) GetCharacterHandler(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"id parameter is required"}`, http.StatusBadRequest)
		return
	}

	// 1. Sub-resource: /timeline
	if strings.HasSuffix(id, "/timeline") {
		cleanID := strings.TrimSuffix(id, "/timeline")
		tl, err := h.engine.GetDetailedTimeline(cleanID)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				http.Error(w, `{"error":"character not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf(`{"error":"timeline lookup failed: %s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"character_id": cleanID,
			"timeline":     tl,
		})
		return
	}

	// 2. Sub-resource: /resonance
	if strings.HasSuffix(id, "/resonance") {
		cleanID := strings.TrimSuffix(id, "/resonance")
		coords := h.parseSolarCoords(r)
		hora := chrono.GetPlanetaryHoraWithCoords(time.Now().UTC(), coords)
		res, err := h.engine.GetProfileResonance(cleanID, hora.PlanetID)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				http.Error(w, `{"error":"character not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf(`{"error":"resonance calculation failed: %s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	// 3. Sub-resource: /chronicles (List or get specific chronicle)
	if strings.HasSuffix(id, "/chronicles") {
		cleanID := strings.TrimSuffix(id, "/chronicles")
		char, err := h.engine.GetSovereignProfile(cleanID)
		if err != nil {
			http.Error(w, `{"error":"character not found"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"character_id": cleanID,
			"chronicles":   char.Chronicles,
		})
		return
	}

	// Standard Character Dossier Lookup
	prof, err := h.engine.GetSovereignProfile(id)
	if err != nil && errors.Is(err, db.ErrNotFound) {
		if !strings.HasPrefix(id, "profile:") {
			prof, err = h.engine.GetSovereignProfile("profile:" + id)
		} else {
			prof, err = h.engine.GetSovereignProfile(strings.TrimPrefix(id, "profile:"))
		}
	}
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			http.Error(w, `{"error":"character not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf(`{"error":"storage error: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// Attach real-time topocentric symbiotic resonance
	coords := h.parseSolarCoords(r)
	hora := chrono.GetPlanetaryHoraWithCoords(time.Now().UTC(), coords)
	resonance := dasha.ResolveSymbioticResonance(prof, hora.PlanetID)
	prof.SymbioticResonance = &resonance

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(prof)
}

// SaveCharacterHandler handles character updates and chronicle compilation.
func (h *Handler) SaveCharacterHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"id parameter is required"}`, http.StatusBadRequest)
		return
	}

	// Action: POST /api/v1/characters/{id}/chronicle
	if strings.HasSuffix(id, "/chronicle") {
		cleanID := strings.TrimSuffix(id, "/chronicle")
		char, err := h.engine.GetSovereignProfile(cleanID)
		if err != nil && errors.Is(err, db.ErrNotFound) {
			if !strings.HasPrefix(cleanID, "profile:") {
				char, err = h.engine.GetSovereignProfile("profile:" + cleanID)
			}
		}
		if err != nil {
			http.Error(w, `{"error":"character not found"}`, http.StatusNotFound)
			return
		}

		var req ChronicleRequest
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&req)
		}

		// Resolve target period
		var targetPeriod *dasha.DashaPeriod
		if req.DashaPlanet != "" {
			targetPlanetID := dasha.PlanetID(strings.ToLower(req.DashaPlanet))
			// Look for matching period in character's timeline
			for _, m := range char.Timeline {
				if m.Planet == targetPlanetID {
					targetPeriod = &m
					break
				}
			}
		}
		if targetPeriod == nil {
			targetPeriod = &char.ActiveSnapshot.Mahadasha
		}

		opts := timeline.CharacterToTimelineOptions{
			Orientation:      timeline.Orientation(req.Orientation),
			DurationSec:      req.DurationSec,
			VoiceAudioPath:   req.VoiceAudioPath,
			AmbientMusicPath: req.AmbientMusicPath,
			IncludeOverlays:  true,
		}

		manifest, err := timeline.GenerateFromCharacterDasha(char, targetPeriod, opts)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"chronicle compilation failed: %s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		// Record reference in character profile
		chronicleRef := dasha.CharacterChronicleRef{
			ID:          manifest.ID,
			Title:       manifest.Title,
			DashaLevel:  string(targetPeriod.Level),
			Planet:      targetPeriod.Planet,
			DurationSec: manifest.DurationSec,
			Orientation: string(manifest.Canvas.Orientation),
			CreatedAt:   time.Now().UTC(),
		}
		char.Chronicles = append(char.Chronicles, chronicleRef)
		_ = h.engine.SaveSovereignProfile(char)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":    "created",
			"chronicle": chronicleRef,
			"manifest":  manifest,
		})
		return
	}

	// Standard Profile Save / Update
	var prof dasha.SovereignProfile
	if err := json.NewDecoder(r.Body).Decode(&prof); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid JSON body: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	prof.ID = id
	if err := h.engine.SaveSovereignProfile(&prof); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to persist character: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(prof)
}
