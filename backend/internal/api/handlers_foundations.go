package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
	"github.com/echosh-labs/mercury-dasha/internal/db"
	"github.com/echosh-labs/mercury-dasha/internal/foundations"
	"github.com/echosh-labs/mercury-dasha/internal/indexer"
	"github.com/echosh-labs/mercury-dasha/internal/timeline"
)

// FoundationsSeedFromTextHandler synthesizes a structured Foundations Story Seed from text or a transcript.
func (h *Handler) FoundationsSeedFromTextHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed, must be POST"}`, http.StatusMethodNotAllowed)
		return
	}

	var req foundations.SeedFromTextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid JSON request: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	var content string
	var title string
	var sourcePath string
	var sourceType string
	var hasAudio bool
	var audioURL string
	var transcriptURL string
	var sanctuaryID string
	var sanctuaryLabel string
	var subSanctuary string
	var noteDate string
	var year int

	timelineList := indexer.LoadDefaultTimeline(h.store)

	// Case 1: Google Recorder Transcript ID
	if strings.HasPrefix(req.ID, "recorder:") || (req.ID != "" && req.Path == "") {
		recID := strings.TrimPrefix(req.ID, "recorder:")
		if h.grecorder != nil {
			transcript, err := h.grecorder.GetTranscript(r.Context(), recID)
			if err == nil && transcript != nil {
				content = transcript.RawText
				sourcePath = fmt.Sprintf("audio/recorder/%s.txt", recID)
				sourceType = "spoken_transcript"
				sanctuaryID = "spoken_transcripts"
				sanctuaryLabel = "Spoken Transcripts"
				subSanctuary = "Voice Chronicle"
				hasAudio = true
				audioURL = fmt.Sprintf("/api/v1/recorder/recordings/%s/audio", recID)
				transcriptURL = fmt.Sprintf("/api/v1/recorder/recordings/%s/transcript", recID)

				// Find recording metadata for date/title
				if recs, lErr := h.grecorder.ListRecordings(r.Context(), ""); lErr == nil {
					for _, rItem := range recs {
						if rItem.ID == recID {
							title = rItem.Title
							if !rItem.RecordedAt.IsZero() {
								noteDate = rItem.RecordedAt.Format("2006-01-02")
								year = rItem.RecordedAt.Year()
							}
							break
						}
					}
				}
			}
		}
	}

	// Case 2: Document Path in Dropbox
	if content == "" && req.Path != "" {
		cleanRel := filepath.Clean(req.Path)
		cleanRel = strings.TrimPrefix(cleanRel, "/")
		cleanRel = strings.TrimPrefix(cleanRel, "\\")

		fullPath := filepath.Join(h.cfg.DropboxLocalPath, cleanRel)
		if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(h.cfg.DropboxLocalPath)) {
			http.Error(w, `{"error":"access forbidden: path traversal outside root"}`, http.StatusForbidden)
			return
		}

		info, err := os.Stat(fullPath)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"file not found: %s"}`, err.Error()), http.StatusNotFound)
			return
		}

		b, err := os.ReadFile(fullPath)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"failed to read document: %s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		content = string(b)
		sourcePath = cleanRel
		sourceType = "written_note"

		fileName := filepath.Base(fullPath)
		sanctuaryID, sanctuaryLabel, subSanctuary = indexer.ClassifySanctuary(cleanRel, fileName)
		nDate, _ := indexer.ExtractDate(fileName, content, info.ModTime())
		noteDate = nDate.Format("2006-01-02")
		year = nDate.Year()

		// Derive clean title
		nameNoExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		title = strings.ReplaceAll(nameNoExt, "_", " ")

		// Check for sidecar audio
		basePathNoExt := strings.TrimSuffix(cleanRel, filepath.Ext(cleanRel))
		for _, audioExt := range []string{".m4a", ".mp3", ".wav", ".aac"} {
			audioRel := basePathNoExt + audioExt
			audioFull := filepath.Join(h.cfg.DropboxLocalPath, audioRel)
			if _, aErr := os.Stat(audioFull); aErr == nil {
				hasAudio = true
				audioURL = fmt.Sprintf("/api/v1/index/content?path=%s", url.QueryEscape(audioRel))
				break
			}
		}
	}

	// Case 3: Raw Content provided directly in request
	if content == "" && req.Content != "" {
		content = req.Content
		sourceType = "written_note"
		sourcePath = "custom/user_input.txt"
		sanctuaryID = "foundations_lore"
		sanctuaryLabel = "Foundations Lore & Musings"
		noteDate = time.Now().Format("2006-01-02")
		year = time.Now().Year()
	}

	if content == "" {
		http.Error(w, `{"error":"no valid content could be resolved from path, id, or content parameters"}`, http.StatusBadRequest)
		return
	}

	// Apply explicit overrides
	if req.Title != "" {
		title = req.Title
	}
	if req.NoteDate != "" {
		noteDate = req.NoteDate
		if parsed, err := time.Parse("2006-01-02", req.NoteDate); err == nil {
			year = parsed.Year()
		}
	}
	if req.Sanctuary != "" {
		sanctuaryID = req.Sanctuary
	}

	input := foundations.SeedInput{
		Title:               title,
		SourcePath:          sourcePath,
		SourceType:          sourceType,
		Content:             content,
		HasAudio:            hasAudio,
		AudioURL:            audioURL,
		TranscriptURL:       transcriptURL,
		SanctuaryID:         sanctuaryID,
		SanctuaryLabel:      sanctuaryLabel,
		SubSanctuary:        subSanctuary,
		NoteDate:            noteDate,
		Year:                year,
		ProtagonistOverride: req.Protagonist,
	}

	seed, err := foundations.SynthesizeSeed(input, nil, timelineList)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to synthesize story seed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// Persist seed if requested
	if req.Save {
		seedJSON, err := json.Marshal(seed)
		if err == nil {
			_ = h.store.PutJSON(db.BucketFoundationsSeeds, seed.ID, seedJSON)
		}
	}

	_ = json.NewEncoder(w).Encode(seed)
}

// FoundationsListSeedsHandler lists saved story seeds from BoltDB.
func (h *Handler) FoundationsListSeedsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	keys, err := h.store.ListKeys(db.BucketFoundationsSeeds, "")
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to list story seeds: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	var seeds []foundations.StorySeed
	for _, k := range keys {
		raw, err := h.store.GetJSON(db.BucketFoundationsSeeds, k)
		if err == nil && len(raw) > 0 {
			var seed foundations.StorySeed
			if err := json.Unmarshal(raw, &seed); err == nil {
				seeds = append(seeds, seed)
			}
		}
	}

	sort.Slice(seeds, func(i, j int) bool {
		return seeds[i].CreatedAt.After(seeds[j].CreatedAt)
	})

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"total": len(seeds),
		"seeds": seeds,
	})
}

// FoundationsGetSeedHandler retrieves a saved story seed by ID.
func (h *Handler) FoundationsGetSeedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract ID from path e.g. /api/v1/foundations/seeds/{id}
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var id string
	if len(pathParts) >= 5 {
		id = pathParts[4]
	}
	if id == "" {
		id = r.URL.Query().Get("id")
	}

	if id == "" {
		http.Error(w, `{"error":"seed id is required"}`, http.StatusBadRequest)
		return
	}

	raw, err := h.store.GetJSON(db.BucketFoundationsSeeds, id)
	if err != nil || len(raw) == 0 {
		http.Error(w, `{"error":"story seed not found"}`, http.StatusNotFound)
		return
	}

	w.Write(raw)
}

// FoundationsDeleteSeedHandler deletes a saved story seed by ID.
func (h *Handler) FoundationsDeleteSeedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var id string
	if len(pathParts) >= 5 {
		id = pathParts[4]
	}
	if id == "" {
		id = r.URL.Query().Get("id")
	}

	if id == "" {
		http.Error(w, `{"error":"seed id is required"}`, http.StatusBadRequest)
		return
	}

	if err := h.store.Delete(db.BucketFoundationsSeeds, id); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to delete story seed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"deleted": true,
		"id":      id,
	})
}

// FoundationsResonantSeedsHandler queries life's work for items matching the active celestial transit.
func (h *Handler) FoundationsResonantSeedsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	timelineList := indexer.LoadDefaultTimeline(h.store)
	now := time.Now()
	if tParam := r.URL.Query().Get("time"); tParam != "" {
		if parsed, err := time.Parse(time.RFC3339, tParam); err == nil {
			now = parsed
		}
	}

	limit := 10
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	snapshot := dasha.ResolveActiveSnapshot(timelineList, now)

	// Gather CatalogItemRef items from BoltDB index and Google Recorder
	var catalogRefs []foundations.CatalogItemRef

	// 1. BoltDB text entries
	filter := db.IndexFilter{
		Category: "text",
		Limit:    1000,
	}
	entries, _, err := h.store.SearchIndex(filter)
	if err == nil {
		for _, e := range entries {
			sanctuaryID, sanctuaryLabel, _ := indexer.ClassifySanctuary(e.Path, e.FileName)
			nDate, _ := indexer.ExtractDate(e.FileName, e.Snippet, e.ModTime)

			snap := dasha.ResolveActiveSnapshot(timelineList, nDate)
			mName, mLatin := indexer.PlanetSacredMetal(snap.Mahadasha.Planet)

			catalogRefs = append(catalogRefs, foundations.CatalogItemRef{
				ID:              e.ID,
				Title:           e.FileName,
				Path:            e.Path,
				SourceType:      "written_note",
				Sanctuary:       sanctuaryID,
				SanctuaryLabel:  sanctuaryLabel,
				NoteDate:        nDate.Format("2006-01-02"),
				Year:            nDate.Year(),
				DashaMahadasha:  snap.Mahadasha.PlanetName,
				DashaAntardasha: snap.Antardasha.PlanetName,
				SacredMetal:     mName + " (" + mLatin + ")",
				HermeticAxiom:   indexer.PlanetHermeticAxiom(snap.Mahadasha.Planet),
				Snippet:         e.Snippet,
				Tags:            e.Tags,
			})
		}
	}

	// 2. Google Recorder entries
	if h.grecorder != nil {
		recordings, rErr := h.grecorder.ListRecordings(r.Context(), "")
		if rErr == nil {
			for _, rec := range recordings {
				if !rec.HasTranscript {
					continue
				}
				recDate := rec.RecordedAt
				if recDate.IsZero() {
					recDate = time.Now()
				}
				snap := dasha.ResolveActiveSnapshot(timelineList, recDate)
				mName, mLatin := indexer.PlanetSacredMetal(snap.Mahadasha.Planet)

				catalogRefs = append(catalogRefs, foundations.CatalogItemRef{
					ID:              "recorder:" + rec.ID,
					Title:           rec.Title,
					Path:            fmt.Sprintf("audio/recorder/%s.txt", rec.ID),
					SourceType:      "spoken_transcript",
					Sanctuary:       "spoken_transcripts",
					SanctuaryLabel:  "Spoken Transcripts",
					NoteDate:        recDate.Format("2006-01-02"),
					Year:            recDate.Year(),
					DashaMahadasha:  snap.Mahadasha.PlanetName,
					DashaAntardasha: snap.Antardasha.PlanetName,
					SacredMetal:     mName + " (" + mLatin + ")",
					HermeticAxiom:   indexer.PlanetHermeticAxiom(snap.Mahadasha.Planet),
					HasAudio:        true,
					AudioURL:        fmt.Sprintf("/api/v1/recorder/recordings/%s/audio", rec.ID),
					Snippet:         rec.Title,
				})
			}
		}
	}

	excerpts := foundations.FindResonantExcerpts(catalogRefs, &snapshot, limit)

	resp := foundations.ResonantTransitResponse{
		Count:    len(excerpts),
		Excerpts: excerpts,
	}
	resp.CurrentTransit.Mahadasha = snapshot.Mahadasha.PlanetName
	resp.CurrentTransit.Antardasha = snapshot.Antardasha.PlanetName
	resp.CurrentTransit.SacredMetal = snapshot.Mahadasha.SacredMetal
	resp.CurrentTransit.HermeticAxiom = snapshot.Antardasha.HermeticAxiom
	resp.CurrentTransit.Archetype = snapshot.StoryContext.ActiveArchetype

	_ = json.NewEncoder(w).Encode(resp)
}

// FoundationsSeedManifestHandler bridges a StorySeed into a TimelineManifest for video compilation.
func (h *Handler) FoundationsSeedManifestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed, must be POST"}`, http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var id string
	if len(pathParts) >= 5 {
		id = pathParts[4]
	}
	if id == "" {
		id = r.URL.Query().Get("id")
	}

	var seed *foundations.StorySeed

	// If ID provided, load from store
	if id != "" {
		raw, err := h.store.GetJSON(db.BucketFoundationsSeeds, id)
		if err == nil && len(raw) > 0 {
			var s foundations.StorySeed
			if uErr := json.Unmarshal(raw, &s); uErr == nil {
				seed = &s
			}
		}
	}

	// If body contains direct seed JSON
	if seed == nil {
		var s foundations.StorySeed
		if err := json.NewDecoder(r.Body).Decode(&s); err == nil && s.ID != "" {
			seed = &s
		}
	}

	if seed == nil {
		http.Error(w, `{"error":"valid story seed or seed ID required to generate manifest"}`, http.StatusBadRequest)
		return
	}

	orientation := timeline.OrientationLandscape16x9
	if o := r.URL.Query().Get("orientation"); o != "" {
		orientation = timeline.Orientation(o)
	}

	duration := 60.0
	if dStr := r.URL.Query().Get("duration_sec"); dStr != "" {
		if d, err := strconv.ParseFloat(dStr, 64); err == nil && d > 0 {
			duration = d
		}
	}

	manifest, err := foundations.SeedToTimelineManifest(seed, orientation, duration)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to generate timeline manifest: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(manifest)
}
