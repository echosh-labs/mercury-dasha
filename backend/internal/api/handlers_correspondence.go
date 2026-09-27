package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
	"github.com/echosh-labs/mercury-dasha/internal/db"
	"github.com/echosh-labs/mercury-dasha/internal/indexer"
)

// PhotoCatalogItem represents a photograph linked to a Dasha era and album.
type PhotoCatalogItem struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Path            string `json:"path"`
	FileName        string `json:"file_name"`
	Album           string `json:"album"`
	PhotoDate       string `json:"photo_date"`
	Year            int    `json:"year"`
	SizeBytes       int64  `json:"size_bytes"`
	DashaMahadasha  string `json:"dasha_mahadasha,omitempty"`
	DashaAntardasha string `json:"dasha_antardasha,omitempty"`
	SacredMetal     string `json:"sacred_metal,omitempty"`
	URL             string `json:"url"`
}

// AudioItemSummary represents an audio recording or voice chronicle snippet.
type AudioItemSummary struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Path           string `json:"path"`
	AudioURL       string `json:"audio_url"`
	TranscriptURL  string `json:"transcript_url,omitempty"`
	Snippet        string `json:"snippet,omitempty"`
	RecordedAt     string `json:"recorded_at"`
	Year           int    `json:"year"`
	DashaMahadasha string `json:"dasha_mahadasha,omitempty"`
}

// AlbumSummary aggregates photo counts by distinct album directory.
type AlbumSummary struct {
	AlbumName string `json:"album_name"`
	Count     int    `json:"count"`
}

// TimelineEraCorrespondence holds the tri-partite assets (Text <-> Photo <-> Audio) for a Vimshottari era.
type TimelineEraCorrespondence struct {
	Mahadasha        string              `json:"mahadasha"`
	SanskritName     string              `json:"sanskrit_name"`
	SacredMetal      string              `json:"sacred_metal"`
	HermeticAxiom    string              `json:"hermetic_axiom"`
	StoryArchetype   string              `json:"story_archetype"`
	StartDate        string              `json:"start_date"`
	EndDate          string              `json:"end_date"`
	YearsLabel       string              `json:"years_label"`
	AgeRange         string              `json:"age_range"`
	IsActive         bool                `json:"is_active"`
	ActiveAntardasha string              `json:"active_antardasha,omitempty"`
	TotalTexts       int                 `json:"total_texts"`
	TotalPhotos      int                 `json:"total_photos"`
	TotalAudio       int                 `json:"total_audio"`
	TotalVideos      int                 `json:"total_videos"`
	PhotoAlbums      []AlbumSummary      `json:"photo_albums"`
	SampleTexts      []TextCatalogItem   `json:"sample_texts"`
	SamplePhotos     []PhotoCatalogItem  `json:"sample_photos"`
	SampleAudio      []AudioItemSummary  `json:"sample_audio"`
}

// TimelineCorrespondenceResponse presents the comprehensive 120-year correspondence matrix.
type TimelineCorrespondenceResponse struct {
	Subject       string                      `json:"subject"`
	DOB           string                      `json:"dob"`
	Birthplace    string                      `json:"birthplace"`
	Lagna         string                      `json:"lagna"`
	Nakshatra     string                      `json:"nakshatra"`
	ActiveEra     string                      `json:"active_era"`
	ActiveMetal   string                      `json:"active_metal"`
	Eras          []TimelineEraCorrespondence `json:"eras"`
	Ancestral     map[string]any              `json:"ancestral"`
	TotalIndexed  int                         `json:"total_indexed"`
	CountsSummary struct {
		TextDocs int `json:"text_documents"`
		Photos   int `json:"photos"`
		Audio    int `json:"audio_recordings"`
		Videos   int `json:"videos"`
	} `json:"counts_summary"`
}

// TimelineCorrespondenceHandler returns correlated Text, Images, and Audio grouped across the 120-Year Vimshottari timeline.
func (h *Handler) TimelineCorrespondenceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	filterPlanet := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("dasha_planet")))
	filterYearStr := strings.TrimSpace(r.URL.Query().Get("year"))
	filterEntity := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("entity")))
	searchQuery := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))

	timeline := indexer.LoadDefaultTimeline(h.store)
	now := time.Now()

	resp := TimelineCorrespondenceResponse{
		Subject:    "Justin Andrew Wood",
		DOB:        "1971-10-01 04:40 EDT",
		Birthplace: "St. Catharines, Ontario",
		Lagna:      "Kumbha (Aquarius) / Virgo Ascendant (Tropical)",
		Nakshatra:  "Dhanishta (Pada 2)",
		Eras:       []TimelineEraCorrespondence{},
		Ancestral: map[string]any{
			"title":        "Ancestral Foundation (Pre-1971)",
			"sacred_metal": "Adamantine",
			"era_span":     "1900–1971",
			"key_figures": []string{
				"Vernon Douglas Wood (1922–1984, Navy Petty Officer, Bell Canada Linesman)",
				"Dorothy Catherine Swepson (Grandmother)",
				"Mark Steven Wood (1950, Father)",
			},
			"description": "Ancestral origins in Welland & St. Catharines, telecommunications lineage with Bell Canada, and esoteric foundations.",
		},
	}

	if len(timeline) == 0 {
		http.Error(w, `{"error":"no dasha timeline profile found"}`, http.StatusNotFound)
		return
	}

	// Map of Mahadasha name -> Pointer to Era Correspondence
	eraMap := make(map[string]*TimelineEraCorrespondence)
	var orderedPlanets []string

	ageRanges := map[string]string{
		"Mars":    "Age 0.0 – 3.5",
		"Rahu":    "Age 3.5 – 21.5",
		"Jupiter": "Age 21.5 – 37.5",
		"Saturn":  "Age 37.5 – 56.5",
		"Mercury": "Age 56.5 – 73.5",
		"Ketu":    "Age 73.5 – 80.5",
		"Venus":   "Age 80.5 – 100.5",
		"Sun":     "Age 100.5 – 106.5",
		"Moon":    "Age 106.5 – 116.5",
	}

	for _, p := range timeline {
		key := p.PlanetName
		if _, exists := eraMap[key]; !exists {
			mName, mLatin := indexer.PlanetSacredMetal(p.Planet)
			isActive := now.After(p.StartDate) && now.Before(p.EndDate)
			activeAntar := ""
			if isActive {
				snap := dasha.ResolveActiveSnapshot(timeline, now)
				activeAntar = snap.Antardasha.PlanetName + " (" + snap.Antardasha.SanskritName + ")"
				resp.ActiveEra = key
				resp.ActiveMetal = mName + " (" + mLatin + ")"
			}

			startYear := p.StartDate.Year()
			endYear := p.EndDate.Year()
			yearsLbl := fmt.Sprintf("%d–%d", startYear, endYear)

			eraMap[key] = &TimelineEraCorrespondence{
				Mahadasha:        p.PlanetName,
				SanskritName:     p.SanskritName,
				SacredMetal:      mName + " (" + mLatin + ")",
				HermeticAxiom:    indexer.PlanetHermeticAxiom(p.Planet),
				StoryArchetype:   indexer.PlanetStoryArchetype(p.Planet),
				StartDate:        p.StartDate.Format("2006-01-02"),
				EndDate:          p.EndDate.Format("2006-01-02"),
				YearsLabel:       yearsLbl,
				AgeRange:         ageRanges[key],
				IsActive:         isActive,
				ActiveAntardasha: activeAntar,
				PhotoAlbums:      []AlbumSummary{},
				SampleTexts:      []TextCatalogItem{},
				SamplePhotos:     []PhotoCatalogItem{},
				SampleAudio:      []AudioItemSummary{},
			}
			orderedPlanets = append(orderedPlanets, key)
		}
	}

	// Album maps per era to aggregate album photo counts
	albumMaps := make(map[string]map[string]int)
	for _, p := range orderedPlanets {
		albumMaps[p] = make(map[string]int)
	}

	// 1. Fetch Text entries from Store
	if h.store != nil {
		textEntries, _, _ := h.store.SearchIndex(db.IndexFilter{Category: "text", Limit: 5000})
		for _, e := range textEntries {
			maha := ""
			if m, ok := e.Metadata["dasha_mahadasha"].(string); ok {
				maha = m
			}
			if maha == "" {
				continue
			}

			era, ok := eraMap[maha]
			if !ok {
				continue
			}

			// Apply filters
			if filterPlanet != "" && !strings.EqualFold(maha, filterPlanet) {
				continue
			}
			if filterYearStr != "" {
				if yInt, err := strconv.Atoi(filterYearStr); err == nil {
					if ey, ok := e.Metadata["year"].(int); ok && ey != yInt {
						continue
					}
				}
			}
			if filterEntity != "" {
				hasEnt := false
				if ents, ok := e.Metadata["entities"].([]any); ok {
					for _, item := range ents {
						if strings.Contains(strings.ToLower(fmt.Sprint(item)), filterEntity) {
							hasEnt = true
							break
						}
					}
				}
				if !hasEnt {
					continue
				}
			}
			if searchQuery != "" {
				match := strings.Contains(strings.ToLower(e.FileName), searchQuery) ||
					strings.Contains(strings.ToLower(e.Snippet), searchQuery) ||
					strings.Contains(strings.ToLower(e.Path), searchQuery)
				if !match {
					continue
				}
			}

			era.TotalTexts++
			resp.CountsSummary.TextDocs++

			if len(era.SampleTexts) < 12 {
				sanctuary, _ := e.Metadata["sanctuary"].(string)
				sanctLabel, _ := e.Metadata["sanctuary_label"].(string)
				subSanct, _ := e.Metadata["sub_sanctuary"].(string)
				title, _ := e.Metadata["title"].(string)
				if title == "" {
					title = e.FileName
				}
				nDate, _ := e.Metadata["note_date"].(string)
				yr, _ := e.Metadata["year"].(int)
				wCnt, _ := e.Metadata["sample_word_count"].(int)
				cCnt, _ := e.Metadata["char_count"].(int)
				rTime, _ := e.Metadata["est_read_time_mins"].(float64)

				item := TextCatalogItem{
					ID:              e.ID,
					Title:           title,
					Path:            e.Path,
					SourceType:      "written_note",
					Sanctuary:       sanctuary,
					SanctuaryLabel:  sanctLabel,
					SubSanctuary:    subSanct,
					NoteDate:        nDate,
					Year:            yr,
					SizeBytes:       e.SizeBytes,
					WordCount:       wCnt,
					CharCount:       cCnt,
					EstReadTimeMins: rTime,
					DashaMahadasha:  maha,
					Snippet:         e.Snippet,
					Tags:            e.Tags,
				}
				era.SampleTexts = append(era.SampleTexts, item)
			}
		}

		// 2. Fetch Photos entries from Store
		photoEntries, _, _ := h.store.SearchIndex(db.IndexFilter{Category: "photos", Limit: 10000})
		if len(photoEntries) == 0 {
			// fallback check "images"
			photoEntries, _, _ = h.store.SearchIndex(db.IndexFilter{Category: "images", Limit: 10000})
		}
		for _, e := range photoEntries {
			maha := ""
			if m, ok := e.Metadata["dasha_mahadasha"].(string); ok {
				maha = m
			}
			if maha == "" {
				continue
			}

			era, ok := eraMap[maha]
			if !ok {
				continue
			}

			if filterPlanet != "" && !strings.EqualFold(maha, filterPlanet) {
				continue
			}
			if filterYearStr != "" {
				if yInt, err := strconv.Atoi(filterYearStr); err == nil {
					if ey, ok := e.Metadata["year"].(int); ok && ey != yInt {
						continue
					}
				}
			}

			era.TotalPhotos++
			resp.CountsSummary.Photos++

			album, _ := e.Metadata["album"].(string)
			if album != "" {
				albumMaps[maha][album]++
			}

			if len(era.SamplePhotos) < 12 {
				pDate, _ := e.Metadata["photo_date"].(string)
				yr, _ := e.Metadata["year"].(int)
				metal, _ := e.Metadata["sacred_metal"].(string)
				photoItem := PhotoCatalogItem{
					ID:             e.ID,
					Title:          e.FileName,
					Path:           e.Path,
					FileName:       e.FileName,
					Album:          album,
					PhotoDate:      pDate,
					Year:           yr,
					SizeBytes:      e.SizeBytes,
					DashaMahadasha: maha,
					SacredMetal:    metal,
					URL:            fmt.Sprintf("/api/v1/dropbox/content?path=%s", e.Path),
				}
				era.SamplePhotos = append(era.SamplePhotos, photoItem)
			}
		}

		// 3. Fetch Audio recordings
		audioEntries, _, _ := h.store.SearchIndex(db.IndexFilter{Category: "audio", Limit: 5000})
		for _, e := range audioEntries {
			// Resolve dasha
			maha := ""
			if m, ok := e.Metadata["dasha_mahadasha"].(string); ok {
				maha = m
			} else {
				snap := dasha.ResolveActiveSnapshot(timeline, e.ModTime)
				maha = snap.Mahadasha.PlanetName
			}

			era, ok := eraMap[maha]
			if !ok {
				continue
			}

			era.TotalAudio++
			resp.CountsSummary.Audio++

			if len(era.SampleAudio) < 8 {
				audioItem := AudioItemSummary{
					ID:             e.ID,
					Title:          e.FileName,
					Path:           e.Path,
					AudioURL:       fmt.Sprintf("/api/v1/dropbox/content?path=%s", e.Path),
					RecordedAt:     e.ModTime.Format("2006-01-02"),
					Year:           e.ModTime.Year(),
					DashaMahadasha: maha,
				}
				era.SampleAudio = append(era.SampleAudio, audioItem)
			}
		}

		// 4. Incorporate Google Recorder synchronized transcripts if available
		if h.grecorder != nil {
			recs, err := h.grecorder.ListRecordings(r.Context(), "")
			if err == nil {
				for _, rec := range recs {
					rDate := rec.RecordedAt
					if rDate.IsZero() {
						rDate = time.Now()
					}
					snap := dasha.ResolveActiveSnapshot(timeline, rDate)
					maha := snap.Mahadasha.PlanetName
					if era, ok := eraMap[maha]; ok {
						era.TotalAudio++
						resp.CountsSummary.Audio++

						if len(era.SampleAudio) < 12 {
							snippet := fmt.Sprintf("Voice recording from %s. Duration: %s", rec.RecordedAt.Format("Jan 02, 2006"), rec.Duration)
							era.SampleAudio = append(era.SampleAudio, AudioItemSummary{
								ID:             "recorder:" + rec.ID,
								Title:          rec.Title,
								Path:           rec.TranscriptTxt,
								AudioURL:       fmt.Sprintf("/api/v1/recorder/recordings/%s/audio", rec.ID),
								TranscriptURL:  fmt.Sprintf("/api/v1/recorder/recordings/%s/transcript", rec.ID),
								Snippet:        snippet,
								RecordedAt:     rDate.Format("2006-01-02"),
								Year:           rDate.Year(),
								DashaMahadasha: maha,
							})
						}
					}
				}
			}
		}
	}

	// Finalize albums summary per era
	for _, p := range orderedPlanets {
		era := eraMap[p]
		aMap := albumMaps[p]
		for aName, count := range aMap {
			era.PhotoAlbums = append(era.PhotoAlbums, AlbumSummary{
				AlbumName: aName,
				Count:     count,
			})
		}
		sort.Slice(era.PhotoAlbums, func(i, j int) bool {
			return era.PhotoAlbums[i].Count > era.PhotoAlbums[j].Count
		})

		// Append to response
		if filterPlanet == "" || strings.EqualFold(p, filterPlanet) {
			resp.Eras = append(resp.Eras, *era)
		}
	}

	resp.TotalIndexed = resp.CountsSummary.TextDocs + resp.CountsSummary.Photos + resp.CountsSummary.Audio

	_ = json.NewEncoder(w).Encode(resp)
}
