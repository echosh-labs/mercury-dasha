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
	"github.com/echosh-labs/mercury-dasha/internal/indexer"
)

// TextCatalogItem represents a written personal note or a spoken voice transcript.
type TextCatalogItem struct {
	ID                string   `json:"id"`
	Title             string   `json:"title"`
	Path              string   `json:"path"`
	SourceType        string   `json:"source_type"` // "written_note" | "spoken_transcript"
	Sanctuary         string   `json:"sanctuary"`
	SanctuaryLabel    string   `json:"sanctuary_label"`
	SubSanctuary      string   `json:"sub_sanctuary,omitempty"`
	NoteDate          string   `json:"note_date"`
	Year              int      `json:"year"`
	HasExplicitDate   bool     `json:"has_explicit_date"`
	SizeBytes         int64    `json:"size_bytes"`
	WordCount         int      `json:"word_count"`
	CharCount         int      `json:"char_count"`
	EstReadTimeMins   float64  `json:"est_read_time_mins"`
	DashaMahadasha    string   `json:"dasha_mahadasha,omitempty"`
	DashaAntardasha   string   `json:"dasha_antardasha,omitempty"`
	SacredMetal       string   `json:"sacred_metal,omitempty"`
	HermeticAxiom     string   `json:"hermetic_axiom,omitempty"`
	StoryArchetype    string   `json:"story_archetype,omitempty"`
	HasAudio          bool     `json:"has_audio"`
	AudioURL          string   `json:"audio_url,omitempty"`
	TranscriptURL     string   `json:"transcript_url,omitempty"`
	Snippet           string   `json:"snippet"`
	Tags              []string `json:"tags,omitempty"`
}

// TextCatalogResponse wraps the unified text library with rich classification metrics.
type TextCatalogResponse struct {
	Items  []TextCatalogItem `json:"items"`
	Total  int               `json:"total"`
	Counts struct {
		Total             int            `json:"total"`
		WrittenNotes      int            `json:"written_notes"`
		SpokenTranscripts int            `json:"spoken_transcripts"`
		WithAudio         int            `json:"with_audio"`
		Sanctuaries       map[string]int `json:"sanctuaries"`
		DashaPlanets      map[string]int `json:"dasha_planets"`
		Years             map[int]int    `json:"years"`
	} `json:"counts"`
}

// TextDocumentDetail returns full document content, formatting, and attached cosmic metadata.
type TextDocumentDetail struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Path            string   `json:"path"`
	Content         string   `json:"content"`
	Format          string   `json:"format"`
	SourceType      string   `json:"source_type"`
	Sanctuary       string   `json:"sanctuary"`
	SanctuaryLabel  string   `json:"sanctuary_label"`
	SubSanctuary    string   `json:"sub_sanctuary,omitempty"`
	NoteDate        string   `json:"note_date"`
	Year            int      `json:"year"`
	HasExplicitDate bool     `json:"has_explicit_date"`
	SizeBytes       int64    `json:"size_bytes"`
	WordCount       int      `json:"word_count"`
	CharCount       int      `json:"char_count"`
	EstReadTimeMins float64  `json:"est_read_time_mins"`
	DashaMahadasha  string   `json:"dasha_mahadasha,omitempty"`
	DashaAntardasha string   `json:"dasha_antardasha,omitempty"`
	SacredMetal     string   `json:"sacred_metal,omitempty"`
	HermeticAxiom   string   `json:"hermetic_axiom,omitempty"`
	StoryArchetype  string   `json:"story_archetype,omitempty"`
	HasAudio        bool     `json:"has_audio"`
	AudioURL        string   `json:"audio_url,omitempty"`
	TranscriptURL   string   `json:"transcript_url,omitempty"`
	Tags            []string `json:"tags,omitempty"`
}

// TextCatalogHandler synthesizes /Dropbox/text documents and spoken transcripts into one catalog.
func (h *Handler) TextCatalogHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	filterSanctuary := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("sanctuary")))
	filterPlanet := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("dasha_planet")))
	filterYearStr := strings.TrimSpace(r.URL.Query().Get("year"))
	filterSource := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("source_type")))
	filterAudio := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("has_audio")))

	limit := 100
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}
	offset := 0
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	timeline := indexer.LoadDefaultTimeline(h.store)

	var allItems []TextCatalogItem
	knownPaths := make(map[string]bool)

	// 1. Ingest Indexed Written Notes from BoltDB
	if h.store != nil {
		docs, _, err := h.store.SearchIndex(db.IndexFilter{
			Category: "text",
			Limit:    2500,
		})
		if err == nil {
			for _, doc := range docs {
				knownPaths[doc.Path] = true

				sourceType := "written_note"
				if strings.Contains(doc.Path, "audio/recorder") || strings.Contains(doc.Path, "recorder/") {
					sourceType = "spoken_transcript"
				}

				sanctuaryID := "foundations_lore"
				sanctuaryLabel := "Foundations Lore & Musings"
				subSanctuary := ""
				if s, ok := doc.Metadata["sanctuary"].(string); ok && s != "" {
					sanctuaryID = s
				}
				if sl, ok := doc.Metadata["sanctuary_label"].(string); ok && sl != "" {
					sanctuaryLabel = sl
				}
				if sub, ok := doc.Metadata["sub_sanctuary"].(string); ok && sub != "" {
					subSanctuary = sub
				}

				noteDate := doc.ModTime.Format("2006-01-02")
				if nd, ok := doc.Metadata["note_date"].(string); ok && nd != "" {
					noteDate = nd
				}
				year := doc.ModTime.Year()
				if y, ok := doc.Metadata["year"].(float64); ok && y > 0 {
					year = int(y)
				} else if yInt, ok := doc.Metadata["year"].(int); ok && yInt > 0 {
					year = yInt
				}

				hasExplicitDate := false
				if hed, ok := doc.Metadata["has_explicit_date"].(bool); ok {
					hasExplicitDate = hed
				}

				wordCount := 0
				if wc, ok := doc.Metadata["sample_word_count"].(float64); ok {
					wordCount = int(wc)
				} else if wcInt, ok := doc.Metadata["sample_word_count"].(int); ok {
					wordCount = wcInt
				}

				charCount := 0
				if cc, ok := doc.Metadata["char_count"].(float64); ok {
					charCount = int(cc)
				} else if ccInt, ok := doc.Metadata["char_count"].(int); ok {
					charCount = ccInt
				}

				readTime := 0.5
				if rt, ok := doc.Metadata["est_read_time_mins"].(float64); ok && rt > 0 {
					readTime = rt
				}

				dashaMaha := ""
				if dm, ok := doc.Metadata["dasha_mahadasha"].(string); ok {
					dashaMaha = dm
				}
				dashaAntar := ""
				if da, ok := doc.Metadata["dasha_antardasha"].(string); ok {
					dashaAntar = da
				}
				metal := ""
				if m, ok := doc.Metadata["sacred_metal"].(string); ok {
					metal = m
				}
				axiom := ""
				if a, ok := doc.Metadata["hermetic_axiom"].(string); ok {
					axiom = a
				}
				archetype := ""
				if arc, ok := doc.Metadata["story_archetype"].(string); ok {
					archetype = arc
				}

				// If Dasha was not yet populated in metadata, calculate dynamically
				if dashaMaha == "" && len(timeline) > 0 {
					if parsedT, pErr := time.Parse("2006-01-02", noteDate); pErr == nil {
						snap := dasha.ResolveActiveSnapshot(timeline, parsedT)
						dashaMaha = snap.Mahadasha.PlanetName
						dashaAntar = snap.Antardasha.PlanetName
						mName, mLatin := indexer.PlanetSacredMetal(snap.Mahadasha.Planet)
						metal = mName + " (" + mLatin + ")"
						axiom = indexer.PlanetHermeticAxiom(snap.Mahadasha.Planet)
						archetype = indexer.PlanetStoryArchetype(snap.Mahadasha.Planet)
					}
				}

				// Title
				title := ""
				if t, ok := doc.Metadata["title"].(string); ok && t != "" {
					title = t
				} else {
					baseName := filepath.Base(doc.FileName)
					nameNoExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
					title = strings.ReplaceAll(nameNoExt, "_", " ")
					title = strings.ReplaceAll(title, "-", " ")
				}

				// Check for sidecar audio
				hasAudio := false
				audioURL := ""
				baseWithoutExt := strings.TrimSuffix(doc.FullPath, filepath.Ext(doc.FullPath))
				if _, statErr := os.Stat(baseWithoutExt + ".m4a"); statErr == nil {
					hasAudio = true
					audioURL = fmt.Sprintf("/api/v1/index/content?path=%s", url.QueryEscape(strings.TrimSuffix(doc.Path, filepath.Ext(doc.Path))+".m4a"))
				} else if _, statErr := os.Stat(baseWithoutExt + ".wav"); statErr == nil {
					hasAudio = true
					audioURL = fmt.Sprintf("/api/v1/index/content?path=%s", url.QueryEscape(strings.TrimSuffix(doc.Path, filepath.Ext(doc.Path))+".wav"))
				}

				item := TextCatalogItem{
					ID:              doc.ID,
					Title:           title,
					Path:            doc.Path,
					SourceType:      sourceType,
					Sanctuary:       sanctuaryID,
					SanctuaryLabel:  sanctuaryLabel,
					SubSanctuary:    subSanctuary,
					NoteDate:        noteDate,
					Year:            year,
					HasExplicitDate: hasExplicitDate,
					SizeBytes:       doc.SizeBytes,
					WordCount:       wordCount,
					CharCount:       charCount,
					EstReadTimeMins: readTime,
					DashaMahadasha:  dashaMaha,
					DashaAntardasha: dashaAntar,
					SacredMetal:     metal,
					HermeticAxiom:   axiom,
					StoryArchetype:  archetype,
					HasAudio:        hasAudio,
					AudioURL:        audioURL,
					Snippet:         doc.Snippet,
					Tags:            doc.Tags,
				}
				allItems = append(allItems, item)
			}
		}
	}

	// 2. Ingest Spoken Transcripts from Google Recorder Service
	if h.grecorder != nil {
		recList, err := h.grecorder.ListRecordings(r.Context(), "")
		if err == nil {
			for _, rec := range recList {
				if !rec.HasTranscript {
					continue
				}

				// Check if already in catalog from indexer walk
				if rec.TranscriptTxt != "" && knownPaths[rec.TranscriptTxt] {
					continue
				}

				noteDate := rec.RecordedAt.Format("2006-01-02")
				year := rec.RecordedAt.Year()

				// Resolve Dasha era
				dashaMaha := ""
				dashaAntar := ""
				metal := ""
				axiom := ""
				archetype := ""
				if len(timeline) > 0 {
					snap := dasha.ResolveActiveSnapshot(timeline, rec.RecordedAt)
					dashaMaha = snap.Mahadasha.PlanetName
					dashaAntar = snap.Antardasha.PlanetName
					mName, mLatin := indexer.PlanetSacredMetal(snap.Mahadasha.Planet)
					metal = mName + " (" + mLatin + ")"
					axiom = indexer.PlanetHermeticAxiom(snap.Mahadasha.Planet)
					archetype = indexer.PlanetStoryArchetype(snap.Mahadasha.Planet)
				}

				// Read transcript snippet & calculate words if file exists
				var snippet string
				var wordCount int
				var charCount int
				var sizeBytes int64
				readTime := 0.5

				if rec.TranscriptTxt != "" {
					if info, sErr := os.Stat(rec.TranscriptTxt); sErr == nil {
						sizeBytes = info.Size()
						if contentBytes, readErr := os.ReadFile(rec.TranscriptTxt); readErr == nil {
							fullText := string(contentBytes)
							wCnt, cCnt, rTime := indexer.CalculateTextMetrics(fullText)
							wordCount = wCnt
							charCount = cCnt
							readTime = rTime

							if len(fullText) > 400 {
								snippet = strings.TrimSpace(fullText[:400]) + "..."
							} else {
								snippet = strings.TrimSpace(fullText)
							}
						}
					}
				}

				if snippet == "" {
					snippet = fmt.Sprintf("Spoken transcript recording from %s. Duration: %s", rec.RecordedAt.Format("Jan 02, 2006"), rec.Duration)
				}

				item := TextCatalogItem{
					ID:              "recorder:" + rec.ID,
					Title:           rec.Title,
					Path:            rec.TranscriptTxt,
					SourceType:      "spoken_transcript",
					Sanctuary:       "spoken_transcripts",
					SanctuaryLabel:  "Spoken Transcripts",
					SubSanctuary:    "Voice Chronicle",
					NoteDate:        noteDate,
					Year:            year,
					HasExplicitDate: true,
					SizeBytes:       sizeBytes,
					WordCount:       wordCount,
					CharCount:       charCount,
					EstReadTimeMins: readTime,
					DashaMahadasha:  dashaMaha,
					DashaAntardasha: dashaAntar,
					SacredMetal:     metal,
					HermeticAxiom:   axiom,
					StoryArchetype:  archetype,
					HasAudio:        true,
					AudioURL:        fmt.Sprintf("/api/v1/recorder/recordings/%s/audio", rec.ID),
					TranscriptURL:   fmt.Sprintf("/api/v1/recorder/recordings/%s/transcript", rec.ID),
					Snippet:         snippet,
					Tags: []string{
						"text",
						"document",
						"transcript",
						"sanctuary:spoken_transcripts",
						"source:recorder",
						fmt.Sprintf("year:%d", year),
					},
				}
				allItems = append(allItems, item)
			}
		}
	}

	// 3. Compute Aggregated Counts
	var resp TextCatalogResponse
	resp.Counts.Sanctuaries = make(map[string]int)
	resp.Counts.DashaPlanets = make(map[string]int)
	resp.Counts.Years = make(map[int]int)

	for _, it := range allItems {
		resp.Counts.Total++
		if it.SourceType == "written_note" {
			resp.Counts.WrittenNotes++
		} else {
			resp.Counts.SpokenTranscripts++
		}
		if it.HasAudio {
			resp.Counts.WithAudio++
		}
		resp.Counts.Sanctuaries[it.Sanctuary]++
		if it.DashaMahadasha != "" {
			resp.Counts.DashaPlanets[strings.ToLower(it.DashaMahadasha)]++
		}
		if it.Year > 0 {
			resp.Counts.Years[it.Year]++
		}
	}

	// 4. Apply Filters
	var filtered []TextCatalogItem
	for _, it := range allItems {
		if filterSanctuary != "" && filterSanctuary != "all" && it.Sanctuary != filterSanctuary {
			continue
		}
		if filterSource != "" && filterSource != "all" && it.SourceType != filterSource {
			continue
		}
		if filterAudio == "true" && !it.HasAudio {
			continue
		}
		if filterAudio == "false" && it.HasAudio {
			continue
		}
		if filterPlanet != "" && filterPlanet != "all" {
			mahaMatch := strings.EqualFold(it.DashaMahadasha, filterPlanet)
			antarMatch := strings.EqualFold(it.DashaAntardasha, filterPlanet)
			if !mahaMatch && !antarMatch {
				continue
			}
		}
		if filterYearStr != "" && filterYearStr != "all" {
			if yInt, err := strconv.Atoi(filterYearStr); err == nil && it.Year != yInt {
				continue
			}
		}
		if q != "" {
			match := strings.Contains(strings.ToLower(it.Title), q) ||
				strings.Contains(strings.ToLower(it.Path), q) ||
				strings.Contains(strings.ToLower(it.Snippet), q) ||
				strings.Contains(strings.ToLower(it.SanctuaryLabel), q)
			if !match {
				continue
			}
		}
		filtered = append(filtered, it)
	}

	// 5. Sort Descending by NoteDate
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].NoteDate > filtered[j].NoteDate
	})

	resp.Total = len(filtered)
	if offset > len(filtered) {
		resp.Items = []TextCatalogItem{}
	} else {
		end := offset + limit
		if end > len(filtered) {
			end = len(filtered)
		}
		resp.Items = filtered[offset:end]
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// TextDocumentHandler serves the complete content and cosmic metadata for a note or transcript.
func (h *Handler) TextDocumentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rawPath := r.URL.Query().Get("path")
	id := r.URL.Query().Get("id")

	// 1. Check if requesting Google Recorder transcript by ID
	if strings.HasPrefix(id, "recorder:") || (id != "" && rawPath == "") {
		recID := strings.TrimPrefix(id, "recorder:")
		if h.grecorder != nil {
			transcript, err := h.grecorder.GetTranscript(r.Context(), recID)
			if err == nil {
				audioURL := fmt.Sprintf("/api/v1/recorder/recordings/%s/audio", recID)
				wCnt, cCnt, rTime := indexer.CalculateTextMetrics(transcript.RawText)

				doc := TextDocumentDetail{
					ID:              "recorder:" + recID,
					Title:           "Voice Chronicle Transcript",
					Path:            fmt.Sprintf("audio/recorder/%s.txt", recID),
					Content:         transcript.RawText,
					Format:          "txt",
					SourceType:      "spoken_transcript",
					Sanctuary:       "spoken_transcripts",
					SanctuaryLabel:  "Spoken Transcripts",
					SubSanctuary:    "Voice Chronicle",
					NoteDate:        time.Now().Format("2006-01-02"),
					Year:            time.Now().Year(),
					HasExplicitDate: true,
					SizeBytes:       int64(len(transcript.RawText)),
					WordCount:       wCnt,
					CharCount:       cCnt,
					EstReadTimeMins: rTime,
					HasAudio:        true,
					AudioURL:        audioURL,
					TranscriptURL:   fmt.Sprintf("/api/v1/recorder/recordings/%s/transcript", recID),
					Tags:            []string{"transcript", "voice", "recorder"},
				}
				_ = json.NewEncoder(w).Encode(doc)
				return
			}
		}
	}

	if rawPath == "" {
		http.Error(w, `{"error":"path or id parameter is required"}`, http.StatusBadRequest)
		return
	}

	// Clean path and ensure local traversal safety
	cleanRel := filepath.Clean(rawPath)
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

	contentBytes, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"could not read file: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	content := string(contentBytes)

	fileName := filepath.Base(fullPath)
	sanctuaryID, sanctuaryLabel, subSanctuary := indexer.ClassifySanctuary(cleanRel, fileName)
	noteDate, hasExplicitDate := indexer.ExtractDate(fileName, content, info.ModTime())
	wCnt, cCnt, rTime := indexer.CalculateTextMetrics(content)

	// Astrological Dasha Alignment
	timeline := indexer.LoadDefaultTimeline(h.store)
	dashaMaha := ""
	dashaAntar := ""
	metal := ""
	axiom := ""
	archetype := ""
	if len(timeline) > 0 {
		snap := dasha.ResolveActiveSnapshot(timeline, noteDate)
		dashaMaha = snap.Mahadasha.PlanetName
		dashaAntar = snap.Antardasha.PlanetName
		mName, mLatin := indexer.PlanetSacredMetal(snap.Mahadasha.Planet)
		metal = mName + " (" + mLatin + ")"
		axiom = indexer.PlanetHermeticAxiom(snap.Mahadasha.Planet)
		archetype = indexer.PlanetStoryArchetype(snap.Mahadasha.Planet)
	}

	// Formatted title
	ext := filepath.Ext(fileName)
	base := strings.TrimSuffix(fileName, ext)
	cleanTitle := strings.ReplaceAll(base, "_", " ")
	cleanTitle = strings.ReplaceAll(cleanTitle, "-", " ")
	if hasExplicitDate && sanctuaryID == "dated_journals" {
		cleanTitle = noteDate.Format("Jan 02, 2006") + " — " + cleanTitle
	}

	// Sidecar Audio check
	hasAudio := false
	audioURL := ""
	baseWithoutExt := strings.TrimSuffix(fullPath, filepath.Ext(fullPath))
	if _, statErr := os.Stat(baseWithoutExt + ".m4a"); statErr == nil {
		hasAudio = true
		audioURL = fmt.Sprintf("/api/v1/index/content?path=%s", url.QueryEscape(strings.TrimSuffix(cleanRel, filepath.Ext(cleanRel))+".m4a"))
	} else if _, statErr := os.Stat(baseWithoutExt + ".wav"); statErr == nil {
		hasAudio = true
		audioURL = fmt.Sprintf("/api/v1/index/content?path=%s", url.QueryEscape(strings.TrimSuffix(cleanRel, filepath.Ext(cleanRel))+".wav"))
	}

	sourceType := "written_note"
	if sanctuaryID == "spoken_transcripts" || strings.Contains(cleanRel, "audio/recorder") {
		sourceType = "spoken_transcript"
	}

	doc := TextDocumentDetail{
		ID:              cleanRel,
		Title:           cleanTitle,
		Path:            cleanRel,
		Content:         content,
		Format:          strings.TrimPrefix(ext, "."),
		SourceType:      sourceType,
		Sanctuary:       sanctuaryID,
		SanctuaryLabel:  sanctuaryLabel,
		SubSanctuary:    subSanctuary,
		NoteDate:        noteDate.Format("2006-01-02"),
		Year:            noteDate.Year(),
		HasExplicitDate: hasExplicitDate,
		SizeBytes:       info.Size(),
		WordCount:       wCnt,
		CharCount:       cCnt,
		EstReadTimeMins: rTime,
		DashaMahadasha:  dashaMaha,
		DashaAntardasha: dashaAntar,
		SacredMetal:     metal,
		HermeticAxiom:   axiom,
		StoryArchetype:  archetype,
		HasAudio:        hasAudio,
		AudioURL:        audioURL,
		Tags: []string{
			"text",
			"sanctuary:" + sanctuaryID,
			fmt.Sprintf("year:%d", noteDate.Year()),
		},
	}

	_ = json.NewEncoder(w).Encode(doc)
}

// TextSanctuariesHandler returns metadata and descriptions for all 10 sanctuaries.
func (h *Handler) TextSanctuariesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"sanctuaries": indexer.AuthoritativeSanctuaries,
		"total":       len(indexer.AuthoritativeSanctuaries),
	})
}

// TextTimelineHandler returns notes and transcripts grouped chronologically by Dasha Era.
func (h *Handler) TextTimelineHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	timeline := indexer.LoadDefaultTimeline(h.store)
	if len(timeline) == 0 {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"eras":    []any{},
			"total":   0,
			"message": "no dasha timeline available",
		})
		return
	}

	type TimelineEraGroup struct {
		Mahadasha      string            `json:"mahadasha"`
		SanskritName   string            `json:"sanskrit_name"`
		SacredMetal    string            `json:"sacred_metal"`
		HermeticAxiom  string            `json:"hermetic_axiom"`
		StoryArchetype string            `json:"story_archetype"`
		StartDate      string            `json:"start_date"`
		EndDate        string            `json:"end_date"`
		TotalDocuments int               `json:"total_documents"`
		SampleItems    []TextCatalogItem `json:"sample_items"`
	}

	eraMap := make(map[string]*TimelineEraGroup)
	var orderedKeys []string

	for _, p := range timeline {
		key := p.PlanetName
		if _, exists := eraMap[key]; !exists {
			mName, mLatin := indexer.PlanetSacredMetal(p.Planet)
			eraMap[key] = &TimelineEraGroup{
				Mahadasha:      p.PlanetName,
				SanskritName:   p.SanskritName,
				SacredMetal:    mName + " (" + mLatin + ")",
				HermeticAxiom:  indexer.PlanetHermeticAxiom(p.Planet),
				StoryArchetype: indexer.PlanetStoryArchetype(p.Planet),
				StartDate:      p.StartDate.Format("2006-01-02"),
				EndDate:        p.EndDate.Format("2006-01-02"),
				SampleItems:    []TextCatalogItem{},
			}
			orderedKeys = append(orderedKeys, key)
		}
	}

	// Fetch indexed items and assign to eras
	if h.store != nil {
		docs, _, _ := h.store.SearchIndex(db.IndexFilter{Category: "text", Limit: 2500})
		for _, doc := range docs {
			if maha, ok := doc.Metadata["dasha_mahadasha"].(string); ok && maha != "" {
				if era, exists := eraMap[maha]; exists {
					era.TotalDocuments++
					if len(era.SampleItems) < 5 {
						title := ""
						if t, ok := doc.Metadata["title"].(string); ok {
							title = t
						} else {
							title = doc.FileName
						}
						nd := ""
						if d, ok := doc.Metadata["note_date"].(string); ok {
							nd = d
						}
						sanc := ""
						if s, ok := doc.Metadata["sanctuary"].(string); ok {
							sanc = s
						}
						era.SampleItems = append(era.SampleItems, TextCatalogItem{
							ID:        doc.ID,
							Title:     title,
							Path:      doc.Path,
							NoteDate:  nd,
							Sanctuary: sanc,
							Snippet:   doc.Snippet,
						})
					}
				}
			}
		}
	}

	var eras []*TimelineEraGroup
	for _, k := range orderedKeys {
		eras = append(eras, eraMap[k])
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"eras":  eras,
		"total": len(eras),
	})
}
