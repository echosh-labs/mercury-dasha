package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/api"
	"github.com/echosh-labs/mercury-dasha/internal/config"
	"github.com/echosh-labs/mercury-dasha/internal/dasha"
	"github.com/echosh-labs/mercury-dasha/internal/db"
	"github.com/echosh-labs/mercury-dasha/internal/indexer"
)

// Handler processes MCP JSON-RPC 2.0 requests for Mercury Dasha.
type Handler struct {
	apiHandler *api.Handler
	store      db.StorageEngine
	cfg        *config.Config
}

// NewHandler creates a new Mercury Dasha MCP request handler.
func NewHandler(apiHandler *api.Handler, store db.StorageEngine, cfg *config.Config) *Handler {
	return &Handler{
		apiHandler: apiHandler,
		store:      store,
		cfg:        cfg,
	}
}

// HandleRequest processes an incoming JSON-RPC 2.0 request.
func (h *Handler) HandleRequest(ctx context.Context, raw []byte) *Response {
	var req Request
	if err := json.Unmarshal(raw, &req); err != nil {
		return &Response{
			JSONRPC: "2.0",
			Error:   &Error{Code: ErrCodeParse, Message: "Parse error: invalid JSON"},
		}
	}

	resp := &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		resp.Result = h.handleInitialize(req.Params)

	case "notifications/initialized":
		return nil

	case "ping":
		resp.Result = map[string]any{}

	case "tools/list":
		resp.Result = h.handleToolsList()

	case "tools/call":
		result, err := h.handleToolsCall(ctx, req.Params)
		if err != nil {
			resp.Result = &ToolResult{
				Content: []ToolContent{{Type: "text", Text: err.Error()}},
				IsError: true,
			}
		} else {
			resp.Result = result
		}

	case "resources/list":
		resList, err := h.handleResourcesList(ctx)
		if err != nil {
			resp.Error = &Error{Code: ErrCodeInternal, Message: err.Error()}
		} else {
			resp.Result = resList
		}

	case "resources/read":
		resContent, err := h.handleResourcesRead(ctx, req.Params)
		if err != nil {
			resp.Error = &Error{Code: ErrCodeInternal, Message: err.Error()}
		} else {
			resp.Result = resContent
		}

	default:
		resp.Error = &Error{
			Code:    ErrCodeNoMethod,
			Message: fmt.Sprintf("Method not found: %s", req.Method),
		}
	}

	return resp
}

func (h *Handler) handleInitialize(_ any) map[string]any {
	return map[string]any{
		"protocolVersion": ProtocolVersion,
		"serverInfo": ServerInfo{
			Name:    "mercury-dasha",
			Version: "1.0.0",
		},
		"capabilities": Capabilities{
			Tools:     &ToolCapability{ListChanged: false},
			Resources: &ResourceCapability{ListChanged: false},
		},
		"instructions": "Mercury Dasha is the Sovereign Scribe & Chrono-Matrix MCP Server for Justin Andrew Wood. It provides live astrological Dasha grounding, the 120-Year Vimshottari correspondence matrix (Text <-> Photo <-> Audio), Scribe's 10 sacred sanctuaries, and Google Recorder transcripts.",
	}
}

func (h *Handler) handleToolsList() map[string]any {
	return map[string]any{
		"tools": []Tool{
			{
				Name:        "get_dasha_correspondence",
				Description: "Retrieve tri-partite narrative context (Text documents, photographs, audio recordings, and alchemical metadata) correlated across Justin's 120-Year Vimshottari timeline.",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"dasha_planet": map[string]any{
							"type":        "string",
							"description": "Filter by Vimshottari planet: Mars, Rahu, Jupiter, Saturn, Mercury, Ketu, Venus, Sun, Moon",
						},
						"year": map[string]any{
							"type":        "integer",
							"description": "Filter by calendar year (e.g. 1971, 1984, 2009, 2026)",
						},
						"entity": map[string]any{
							"type":        "string",
							"description": "Filter by character or entity mention (e.g. Vernon, Mark, Deborah, Luke, Sarah, Bell Canada, Karate, Chiron)",
						},
						"query": map[string]any{
							"type":        "string",
							"description": "Search keyword or text snippet match",
						},
					},
				},
			},
			{
				Name:        "get_active_cosmic_pulse",
				Description: "Get the real-time astrological Dasha and Hora status, active Mahadasha, Antardasha, sacred metal, Hermetic axiom, and story archetype for right now (or a specified timestamp).",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"timestamp": map[string]any{
							"type":        "string",
							"description": "Optional ISO-8601 timestamp (defaults to current time)",
						},
					},
				},
			},
			{
				Name:        "search_sanctuary_vault",
				Description: "Search Justin's written personal notes and sacred writings across the 10 thematic sanctuaries.",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"sanctuary": map[string]any{
							"type":        "string",
							"description": "Sanctuary ID: blessings, hermetic_kybalion, martial_discipline, vision_dreams, devotion_prayer, vocation_manifestation, rosicrucian_study, dated_journals, spoken_transcripts, foundations_lore, or 'all'",
						},
						"query": map[string]any{
							"type":        "string",
							"description": "Search keyword across titles and document content",
						},
						"limit": map[string]any{
							"type":        "integer",
							"description": "Maximum items to return (default 20)",
						},
					},
				},
			},
			{
				Name:        "read_text_document",
				Description: "Retrieve the full text content, word count, read time, and attached cosmic Dasha metadata for any note or transcript in the repository.",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]any{
							"type":        "string",
							"description": "Relative path within Dropbox (e.g. text/blessings/Sarah.txt or Migrated Paper Docs/the longest day in his eternal service.txt)",
						},
					},
					"required": []string{"path"},
				},
			},
			{
				Name:        "get_voice_chronicle",
				Description: "Search and retrieve Google Recorder voice recordings and transcripts synchronized through the local audio archive.",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"query": map[string]any{
							"type":        "string",
							"description": "Search query across voice recording transcripts and titles",
						},
						"recording_id": map[string]any{
							"type":        "string",
							"description": "Optional specific recording ID to retrieve",
						},
						"limit": map[string]any{
							"type":        "integer",
							"description": "Max recordings to return (default 15)",
						},
					},
				},
			},
			{
				Name:        "list_photo_albums",
				Description: "List chronological photo albums and event directories grounded in Justin's life eras with photo counts.",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"dasha_planet": map[string]any{
							"type":        "string",
							"description": "Filter albums by Dasha era: Saturn, Jupiter, etc.",
						},
						"year": map[string]any{
							"type":        "integer",
							"description": "Filter albums by year",
						},
					},
				},
			},
			{
				Name:        "search_sovereign_storehouse",
				Description: "Universal search across the 119,000+ indexed repository assets (text, photos, audio, video, code, books).",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"query": map[string]any{
							"type":        "string",
							"description": "Search term across paths, filenames, and tags",
						},
						"category": map[string]any{
							"type":        "string",
							"description": "Category: text, photos, audio, video, code, books",
						},
						"limit": map[string]any{
							"type":        "integer",
							"description": "Max results to return (default 25)",
						},
					},
				},
			},
			{
				Name:        "calculate_dasha_timeline",
				Description: "Calculate the exact 120-year Vimshottari Dasha timeline for any natal date/time and coordinates.",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"dob": map[string]any{
							"type":        "string",
							"description": "Date of birth in YYYY-MM-DD format",
						},
						"time": map[string]any{
							"type":        "string",
							"description": "Time of birth in HH:MM format (24-hour)",
						},
						"lat": map[string]any{
							"type":        "number",
							"description": "Latitude in decimal degrees",
						},
						"lon": map[string]any{
							"type":        "number",
							"description": "Longitude in decimal degrees",
						},
					},
					"required": []string{"dob", "time", "lat", "lon"},
				},
			},
		},
	}
}

func (h *Handler) handleToolsCall(ctx context.Context, params any) (*ToolResult, error) {
	pMap, ok := params.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid params: expected object")
	}

	name, _ := pMap["name"].(string)
	args, _ := pMap["arguments"].(map[string]any)
	if args == nil {
		args = make(map[string]any)
	}

	timeline := indexer.LoadDefaultTimeline(h.store)

	switch name {
	case "get_dasha_correspondence":
		filterPlanet, _ := args["dasha_planet"].(string)
		filterEntity, _ := args["entity"].(string)
		query, _ := args["query"].(string)
		var year int
		if yVal, ok := args["year"]; ok {
			switch v := yVal.(type) {
			case float64:
				year = int(v)
			case int:
				year = v
			case string:
				year, _ = strconv.Atoi(v)
			}
		}

		// Query items from store
		textItems, _, _ := h.store.SearchIndex(db.IndexFilter{Category: "text", Limit: 5000})
		photoItems, _, _ := h.store.SearchIndex(db.IndexFilter{Category: "photos", Limit: 10000})

		type EraSummary struct {
			Planet         string   `json:"planet"`
			Sanskrit       string   `json:"sanskrit"`
			Metal          string   `json:"metal"`
			Axiom          string   `json:"axiom"`
			Archetype      string   `json:"archetype"`
			Span           string   `json:"span"`
			TotalTexts     int      `json:"total_texts"`
			TotalPhotos    int      `json:"total_photos"`
			SampleTexts    []string `json:"sample_texts"`
			SampleAlbums   []string `json:"sample_albums"`
		}

		eraSummaries := make(map[string]*EraSummary)
		for _, p := range timeline {
			mName, mLatin := indexer.PlanetSacredMetal(p.Planet)
			eraSummaries[p.PlanetName] = &EraSummary{
				Planet:       p.PlanetName,
				Sanskrit:     p.SanskritName,
				Metal:        mName + " (" + mLatin + ")",
				Axiom:        indexer.PlanetHermeticAxiom(p.Planet),
				Archetype:    indexer.PlanetStoryArchetype(p.Planet),
				Span:         fmt.Sprintf("%d–%d", p.StartDate.Year(), p.EndDate.Year()),
				SampleTexts:  []string{},
				SampleAlbums: []string{},
			}
		}

		// Filter texts
		for _, t := range textItems {
			maha, _ := t.Metadata["dasha_mahadasha"].(string)
			if maha == "" {
				continue
			}
			if filterPlanet != "" && !strings.EqualFold(maha, filterPlanet) {
				continue
			}
			if year > 0 {
				if y, ok := t.Metadata["year"].(int); ok && y != year {
					continue
				}
			}
			if filterEntity != "" {
				has := false
				if ents, ok := t.Metadata["entities"].([]any); ok {
					for _, e := range ents {
						if strings.Contains(strings.ToLower(fmt.Sprint(e)), strings.ToLower(filterEntity)) {
							has = true
							break
						}
					}
				}
				if !has {
					continue
				}
			}
			if query != "" {
				m := strings.Contains(strings.ToLower(t.FileName), strings.ToLower(query)) ||
					strings.Contains(strings.ToLower(t.Snippet), strings.ToLower(query))
				if !m {
					continue
				}
			}

			if era, ok := eraSummaries[maha]; ok {
				era.TotalTexts++
				if len(era.SampleTexts) < 8 {
					nDate, _ := t.Metadata["note_date"].(string)
					era.SampleTexts = append(era.SampleTexts, fmt.Sprintf("[%s] %s (%s)", nDate, t.FileName, t.Path))
				}
			}
		}

		// Filter photos
		albumSets := make(map[string]map[string]bool)
		for _, p := range photoItems {
			maha, _ := p.Metadata["dasha_mahadasha"].(string)
			if maha == "" {
				continue
			}
			if filterPlanet != "" && !strings.EqualFold(maha, filterPlanet) {
				continue
			}
			if year > 0 {
				if y, ok := p.Metadata["year"].(int); ok && y != year {
					continue
				}
			}

			if era, ok := eraSummaries[maha]; ok {
				era.TotalPhotos++
				album, _ := p.Metadata["album"].(string)
				if album != "" {
					if albumSets[maha] == nil {
						albumSets[maha] = make(map[string]bool)
					}
					albumSets[maha][album] = true
				}
			}
		}

		for maha, aSet := range albumSets {
			if era, ok := eraSummaries[maha]; ok {
				for alb := range aSet {
					if len(era.SampleAlbums) < 8 {
						era.SampleAlbums = append(era.SampleAlbums, alb)
					}
				}
			}
		}

		outJSON, _ := json.MarshalIndent(eraSummaries, "", "  ")
		return &ToolResult{
			Content: []ToolContent{{Type: "text", Text: string(outJSON)}},
		}, nil

	case "get_active_cosmic_pulse":
		t := time.Now()
		if tsStr, ok := args["timestamp"].(string); ok && tsStr != "" {
			if parsed, err := time.Parse(time.RFC3339, tsStr); err == nil {
				t = parsed
			}
		}

		snap := dasha.ResolveActiveSnapshot(timeline, t)
		metalName, metalLatin := indexer.PlanetSacredMetal(snap.Mahadasha.Planet)
		antarMetal, _ := indexer.PlanetSacredMetal(snap.Antardasha.Planet)

		res := map[string]any{
			"timestamp":        t.Format(time.RFC3339),
			"active_mahadasha": snap.Mahadasha.PlanetName,
			"mahadasha_sanskrit": snap.Mahadasha.SanskritName,
			"active_antardasha": snap.Antardasha.PlanetName,
			"antardasha_sanskrit": snap.Antardasha.SanskritName,
			"sacred_metal":     metalName + " (" + metalLatin + ")",
			"antardasha_metal": antarMetal,
			"hermetic_axiom":   indexer.PlanetHermeticAxiom(snap.Mahadasha.Planet),
			"story_archetype":  indexer.PlanetStoryArchetype(snap.Mahadasha.Planet),
			"mahadasha_start":  snap.Mahadasha.StartDate.Format("2006-01-02"),
			"mahadasha_end":    snap.Mahadasha.EndDate.Format("2006-01-02"),
			"current_age":      fmt.Sprintf("%.1f years", t.Sub(time.Date(1971, 10, 1, 4, 40, 0, 0, time.UTC)).Hours()/(24*365.25)),
		}
		outJSON, _ := json.MarshalIndent(res, "", "  ")
		return &ToolResult{Content: []ToolContent{{Type: "text", Text: string(outJSON)}}}, nil

	case "search_sanctuary_vault":
		sanct, _ := args["sanctuary"].(string)
		q, _ := args["query"].(string)
		limit := 20
		if lVal, ok := args["limit"].(float64); ok && lVal > 0 {
			limit = int(lVal)
		}

		entries, _, err := h.store.SearchIndex(db.IndexFilter{
			Category: "text",
			Query:    q,
			Limit:    limit * 5,
		})
		if err != nil {
			return nil, err
		}

		var matches []map[string]any
		for _, e := range entries {
			sID, _ := e.Metadata["sanctuary"].(string)
			if sanct != "" && sanct != "all" && !strings.EqualFold(sID, sanct) {
				continue
			}
			matches = append(matches, map[string]any{
				"path":            e.Path,
				"file_name":       e.FileName,
				"sanctuary":       sID,
				"sanctuary_label": e.Metadata["sanctuary_label"],
				"note_date":       e.Metadata["note_date"],
				"year":            e.Metadata["year"],
				"dasha":           e.Metadata["dasha_mahadasha"],
				"sacred_metal":    e.Metadata["sacred_metal"],
				"snippet":         e.Snippet,
			})
			if len(matches) >= limit {
				break
			}
		}

		outJSON, _ := json.MarshalIndent(matches, "", "  ")
		return &ToolResult{Content: []ToolContent{{Type: "text", Text: string(outJSON)}}}, nil

	case "read_text_document":
		path, _ := args["path"].(string)
		if path == "" {
			return nil, fmt.Errorf("path parameter is required")
		}

		cleanRel := filepath.Clean(strings.TrimPrefix(strings.TrimPrefix(path, "/"), "\\"))
		fullPath := filepath.Join(h.cfg.DropboxLocalPath, cleanRel)
		contentBytes, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("could not read file %s: %v", path, err)
		}

		content := string(contentBytes)
		fileName := filepath.Base(fullPath)
		info, _ := os.Stat(fullPath)
		modTime := time.Now()
		if info != nil {
			modTime = info.ModTime()
		}

		sanctID, sanctLabel, subSanct := indexer.ClassifySanctuary(cleanRel, fileName)
		noteDate, hasExplicit := indexer.ExtractDate(fileName, content, modTime)
		wCnt, cCnt, rTime := indexer.CalculateTextMetrics(content)

		snap := dasha.ResolveActiveSnapshot(timeline, noteDate)
		metalName, metalLatin := indexer.PlanetSacredMetal(snap.Mahadasha.Planet)

		res := map[string]any{
			"path":              cleanRel,
			"file_name":         fileName,
			"sanctuary":         sanctID,
			"sanctuary_label":   sanctLabel,
			"sub_sanctuary":     subSanct,
			"note_date":         noteDate.Format("2006-01-02"),
			"has_explicit_date": hasExplicit,
			"word_count":        wCnt,
			"char_count":        cCnt,
			"est_read_time_min": rTime,
			"dasha_mahadasha":   snap.Mahadasha.PlanetName,
			"dasha_antardasha":  snap.Antardasha.PlanetName,
			"sacred_metal":      metalName + " (" + metalLatin + ")",
			"hermetic_axiom":    indexer.PlanetHermeticAxiom(snap.Mahadasha.Planet),
			"content":           content,
		}
		outJSON, _ := json.MarshalIndent(res, "", "  ")
		return &ToolResult{Content: []ToolContent{{Type: "text", Text: string(outJSON)}}}, nil

	case "get_voice_chronicle":
		q, _ := args["query"].(string)
		recID, _ := args["recording_id"].(string)
		limit := 15
		if lVal, ok := args["limit"].(float64); ok && lVal > 0 {
			limit = int(lVal)
		}

		// Search recorder entries from store
		entries, _, err := h.store.SearchIndex(db.IndexFilter{
			Category: "audio",
			Query:    q,
			Limit:    limit * 3,
		})
		if err != nil {
			return nil, err
		}

		var results []map[string]any
		for _, e := range entries {
			if !strings.Contains(e.Path, "recorder") {
				continue
			}
			if recID != "" && !strings.Contains(e.FileName, recID) {
				continue
			}
			results = append(results, map[string]any{
				"path":        e.Path,
				"file_name":   e.FileName,
				"recorded_at": e.ModTime.Format("2006-01-02"),
				"year":        e.ModTime.Year(),
				"size_bytes":  e.SizeBytes,
			})
			if len(results) >= limit {
				break
			}
		}

		outJSON, _ := json.MarshalIndent(results, "", "  ")
		return &ToolResult{Content: []ToolContent{{Type: "text", Text: string(outJSON)}}}, nil

	case "list_photo_albums":
		planetFilter, _ := args["dasha_planet"].(string)
		var yearFilter int
		if yVal, ok := args["year"]; ok {
			switch v := yVal.(type) {
			case float64:
				yearFilter = int(v)
			case int:
				yearFilter = v
			case string:
				yearFilter, _ = strconv.Atoi(v)
			}
		}

		entries, _, _ := h.store.SearchIndex(db.IndexFilter{Category: "photos", Limit: 15000})
		albumCounts := make(map[string]map[string]int)

		for _, e := range entries {
			maha, _ := e.Metadata["dasha_mahadasha"].(string)
			if maha == "" {
				continue
			}
			if planetFilter != "" && !strings.EqualFold(maha, planetFilter) {
				continue
			}
			if yearFilter > 0 {
				if y, ok := e.Metadata["year"].(int); ok && y != yearFilter {
					continue
				}
			}

			album, _ := e.Metadata["album"].(string)
			if album == "" {
				album = "General Archive"
			}
			if albumCounts[maha] == nil {
				albumCounts[maha] = make(map[string]int)
			}
			albumCounts[maha][album]++
		}

		outJSON, _ := json.MarshalIndent(albumCounts, "", "  ")
		return &ToolResult{Content: []ToolContent{{Type: "text", Text: string(outJSON)}}}, nil

	case "search_sovereign_storehouse":
		q, _ := args["query"].(string)
		cat, _ := args["category"].(string)
		limit := 25
		if lVal, ok := args["limit"].(float64); ok && lVal > 0 {
			limit = int(lVal)
		}

		entries, total, err := h.store.SearchIndex(db.IndexFilter{
			Category: cat,
			Query:    q,
			Limit:    limit,
		})
		if err != nil {
			return nil, err
		}

		res := map[string]any{
			"query":    q,
			"category": cat,
			"total":    total,
			"results":  entries,
		}
		outJSON, _ := json.MarshalIndent(res, "", "  ")
		return &ToolResult{Content: []ToolContent{{Type: "text", Text: string(outJSON)}}}, nil

	case "calculate_dasha_timeline":
		dob, _ := args["dob"].(string)
		tStr, _ := args["time"].(string)
		lat, _ := args["lat"].(float64)
		lon, _ := args["lon"].(float64)

		calcProfile, err := dasha.CalculateFromRequest(dasha.CalculationRequest{
			Name:           "Calculated Chart",
			BirthDate:      dob,
			BirthTime:      tStr,
			Latitude:       lat,
			Longitude:      lon,
			TimezoneOffset: -4.0,
		})
		if err != nil {
			return nil, fmt.Errorf("calculation error: %v", err)
		}

		outJSON, _ := json.MarshalIndent(calcProfile, "", "  ")
		return &ToolResult{Content: []ToolContent{{Type: "text", Text: string(outJSON)}}}, nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (h *Handler) handleResourcesList(_ context.Context) (map[string]any, error) {
	return map[string]any{
		"resources": []Resource{
			{
				URI:         "mercury://timeline/active",
				Name:        "Active Cosmic Pulse",
				Description: "Real-time active Mahadasha, Antardasha, Hora, sacred metal, and Hermetic axiom.",
				MimeType:    "application/json",
			},
			{
				URI:         "mercury://timeline/matrix",
				Name:        "120-Year Vimshottari Timeline Matrix",
				Description: "Master 120-year correspondence matrix linking life eras with texts, photos, and voice chronicles.",
				MimeType:    "application/json",
			},
			{
				URI:         "mercury://sanctuaries",
				Name:        "Authoritative Text Sanctuaries",
				Description: "Catalog of the 10 thematic sanctuaries for personal writings, prayers, blessings, and lore.",
				MimeType:    "application/json",
			},
			{
				URI:         "mercury://profile/sovereign-genesis",
				Name:        "Sovereign Genesis Natal Profile",
				Description: "Canonical birth chart, Kumbha Lagna, Dhanishta Pada 2, and natal grounding.",
				MimeType:    "application/json",
			},
			{
				URI:         "mercury://ancestral",
				Name:        "Ancestral Foundation",
				Description: "Origins and genealogy: Vernon Douglas Wood (1922-1984), Dorothy Catherine Swepson, Bell Canada telecommunications.",
				MimeType:    "application/json",
			},
		},
	}, nil
}

func (h *Handler) handleResourcesRead(_ context.Context, params any) (map[string]any, error) {
	pMap, ok := params.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid params: expected object")
	}
	uri, _ := pMap["uri"].(string)
	if uri == "" {
		return nil, fmt.Errorf("uri is required")
	}

	timeline := indexer.LoadDefaultTimeline(h.store)
	now := time.Now()

	switch uri {
	case "mercury://timeline/active":
		snap := dasha.ResolveActiveSnapshot(timeline, now)
		metalName, metalLatin := indexer.PlanetSacredMetal(snap.Mahadasha.Planet)
		data := map[string]any{
			"timestamp":        now.Format(time.RFC3339),
			"active_mahadasha": snap.Mahadasha.PlanetName,
			"active_antardasha": snap.Antardasha.PlanetName,
			"sacred_metal":     metalName + " (" + metalLatin + ")",
			"hermetic_axiom":   indexer.PlanetHermeticAxiom(snap.Mahadasha.Planet),
			"story_archetype":  indexer.PlanetStoryArchetype(snap.Mahadasha.Planet),
		}
		raw, _ := json.MarshalIndent(data, "", "  ")
		return map[string]any{
			"contents": []ResourceContent{{URI: uri, MimeType: "application/json", Text: string(raw)}},
		}, nil

	case "mercury://timeline/matrix":
		type EraBlock struct {
			Planet   string `json:"planet"`
			Sanskrit string `json:"sanskrit"`
			Years    string `json:"years"`
			Metal    string `json:"metal"`
		}
		var eras []EraBlock
		for _, p := range timeline {
			mName, _ := indexer.PlanetSacredMetal(p.Planet)
			eras = append(eras, EraBlock{
				Planet:   p.PlanetName,
				Sanskrit: p.SanskritName,
				Years:    fmt.Sprintf("%d–%d", p.StartDate.Year(), p.EndDate.Year()),
				Metal:    mName,
			})
		}
		raw, _ := json.MarshalIndent(eras, "", "  ")
		return map[string]any{
			"contents": []ResourceContent{{URI: uri, MimeType: "application/json", Text: string(raw)}},
		}, nil

	case "mercury://sanctuaries":
		raw, _ := json.MarshalIndent(indexer.AuthoritativeSanctuaries, "", "  ")
		return map[string]any{
			"contents": []ResourceContent{{URI: uri, MimeType: "application/json", Text: string(raw)}},
		}, nil

	case "mercury://profile/sovereign-genesis":
		profRaw, err := h.store.GetProfile("profile:sovereign-genesis")
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"contents": []ResourceContent{{URI: uri, MimeType: "application/json", Text: string(profRaw)}},
		}, nil

	case "mercury://ancestral":
		data := map[string]any{
			"title":        "Ancestral Foundation (Pre-1971)",
			"sacred_metal": "Adamantine",
			"span":         "1900–1971",
			"key_figures": []map[string]string{
				{
					"name":        "Vernon Douglas Wood",
					"dates":       "1922 – March 24, 1984",
					"service":     "Royal Canadian Navy Petty Officer; Bell Canada Linesman (30+ years, retired 1980)",
					"resting":     "Welland, Ontario",
					"key_archive": "vernon-douglas-bell-canada.jpg, bell-canada.jpg, the longest day in his eternal service.txt",
				},
				{
					"name":    "Dorothy Catherine Swepson",
					"dates":   "Grandmother",
					"resting": "Welland, Ontario",
				},
				{
					"name":  "Mark Steven Wood",
					"dates": "Born 1950",
					"role":  "Father",
				},
			},
		}
		raw, _ := json.MarshalIndent(data, "", "  ")
		return map[string]any{
			"contents": []ResourceContent{{URI: uri, MimeType: "application/json", Text: string(raw)}},
		}, nil

	default:
		return nil, fmt.Errorf("resource not found: %s", uri)
	}
}
