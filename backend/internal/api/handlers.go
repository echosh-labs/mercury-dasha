package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/amr"
	"github.com/echosh-labs/mercury-dasha/internal/amra"
	"github.com/echosh-labs/mercury-dasha/internal/chrono"
	"github.com/echosh-labs/mercury-dasha/internal/config"
	"github.com/echosh-labs/mercury-dasha/internal/dasha"
	"github.com/echosh-labs/mercury-dasha/internal/db"
	"github.com/echosh-labs/mercury-dasha/internal/dropbox"
	"github.com/echosh-labs/mercury-dasha/internal/grecorder"
	"github.com/echosh-labs/mercury-dasha/internal/indexer"
	"github.com/echosh-labs/mercury-dasha/internal/selfhealing"
	"github.com/echosh-labs/mercury-dasha/internal/session"
	"github.com/echosh-labs/mercury-dasha/internal/syncthing"
	"github.com/echosh-labs/mercury-dasha/internal/youtube"
)

var startTime = time.Now()

type visualCacheEntry struct {
	items    []UnifiedVisualItem
	cachedAt time.Time
}

type Handler struct {
	cfg         *config.Config
	store       db.StorageEngine
	engine      *dasha.Engine
	dropbox     *dropbox.Client
	indexer     *indexer.Crawler
	metronome   *chrono.Metronome
	sessions    *session.SessionStore
	youtube     *youtube.Client
	amraEngine  *amra.Engine
	amraHandler *amra.Handler
	syncthing   *syncthing.Subsystem
	syncthingH  *syncthing.HTTPHandler
	grecorder   *grecorder.Service
	selfHealing *selfhealing.Controller
	restartMgr  *selfhealing.RestartManager
	visualMu    sync.RWMutex
	visualCache *visualCacheEntry
}

func NewHandler(cfg *config.Config, store db.StorageEngine, dropClients ...*dropbox.Client) *Handler {
	var dropClient *dropbox.Client
	if len(dropClients) > 0 && dropClients[0] != nil {
		dropClient = dropClients[0]
	} else {
		dropClient = dropbox.NewClient(
			cfg.DropboxAccessToken,
			cfg.DropboxRefreshToken,
			cfg.DropboxAppKey,
			cfg.DropboxAppSecret,
			cfg.DropboxBackupPath,
		)
	}

	localPath := cfg.DropboxLocalPath
	if localPath == "" {
		localPath = "/home/justin/Dropbox"
	}
	crawl := indexer.NewCrawler(localPath, store)

	engine := dasha.NewEngine(store)

	metro := chrono.NewMetronome(store, 2*time.Second)

	sessionsPath := filepath.Join(localPath, "MercuryDasha", "sessions", "dnd")
	sessionStore := session.NewSessionStore(store, sessionsPath)

	ytSecrets := youtube.NewEnvSecretProvider(cfg.YouTubeClientID, cfg.YouTubeClientSecret, cfg.YouTubeRedirectURL)
	ytClient := youtube.NewClient(ytSecrets, store, func() string {
		coords := chrono.DefaultCoordinates
		if store != nil {
			if data, err := store.GetSettings(); err == nil && len(data) > 0 {
				var s config.SystemSettings
				if err := json.Unmarshal(data, &s); err == nil && (s.Latitude != 0 || s.Longitude != 0) {
					coords = chrono.SolarCoordinates{Latitude: s.Latitude, Longitude: s.Longitude}
				}
			}
		}
		hora := chrono.GetPlanetaryHoraWithCoords(time.Now().UTC(), coords)
		if hora.Name == "" {
			return ""
		}
		return fmt.Sprintf("══════════════════════════════════════════════════════════════════\n🌌 Mercury Dasha Sovereign Chronicle\n• Planetary Hora: %s (%s)\n• Sacred Metal: %s\n• Hermetic Axiom: %s\n• Story Archetype: %s\n• Host Substrate: echosh-labs single-binary engine (WSL2 Ubuntu)\n══════════════════════════════════════════════════════════════════",
			hora.Name, hora.SanskritName, hora.SacredMetal, hora.HermeticAxiom, hora.StoryArchetype)
	})

	amraEngine := amra.NewEngine(store)
	amraHandler := amra.NewHandler(amraEngine)

	dashaTimeline := indexer.LoadDefaultTimeline(store)
	syncthingSub := syncthing.NewSubsystem(cfg.SyncthingURL, cfg.SyncthingAPIKey, cfg.MediaLocalPath, store, dashaTimeline)
	syncthingH := syncthing.NewHTTPHandler(syncthingSub)

	h := &Handler{
		cfg:         cfg,
		store:       store,
		engine:      engine,
		dropbox:     dropClient,
		indexer:     crawl,
		metronome:   metro,
		sessions:    sessionStore,
		youtube:     ytClient,
		amraEngine:  amraEngine,
		amraHandler: amraHandler,
		syncthing:   syncthingSub,
		syncthingH:  syncthingH,
	}

	restartMgr := selfhealing.NewRestartManager(store, cfg.RestartToken)
	selfHealing := selfhealing.NewController(store, cfg, restartMgr)
	h.restartMgr = restartMgr
	h.selfHealing = selfHealing

	grecorderSvc := grecorder.NewService(cfg, store)
	grecorderSvc.StartBackgroundPoller()
	h.grecorder = grecorderSvc

	_ = dasha.SeedSovereignStorehouse(store)
	_ = dasha.MigrateProfiles(store)
	h.ensureSovereignProfileLocation()

	metro.SetCoordsResolver(func() chrono.SolarCoordinates {
		s := h.loadActiveSettings()
		return chrono.SolarCoordinates{
			Latitude:  s.Latitude,
			Longitude: s.Longitude,
		}
	})
	metro.Start()

	return h
}

// Syncthing returns the active Syncthing media synchronization subsystem.
func (h *Handler) Syncthing() *syncthing.Subsystem {
	return h.syncthing
}

// SelfHealing returns the controller managing health evaluation and auto-remediation.
func (h *Handler) SelfHealing() *selfhealing.Controller {
	return h.selfHealing
}

// RestartManager returns the coordinator managing graceful restarts and re-exec.
func (h *Handler) RestartManager() *selfhealing.RestartManager {
	return h.restartMgr
}

// Close gracefully terminates background metronome ticker, poller, watchdog, and syncthing event poller.
func (h *Handler) Close() {
	if h.selfHealing != nil {
		h.selfHealing.Stop()
	}
	if h.metronome != nil {
		h.metronome.Stop()
	}
	if h.grecorder != nil {
		h.grecorder.StopBackgroundPoller()
	}
	if h.syncthing != nil {
		h.syncthing.Stop()
	}
}

type HealthResponse struct {
	Status      string       `json:"status"`
	Service     string       `json:"service"`
	Version     string       `json:"version"`
	GoVersion   string       `json:"go_version"`
	Environment string       `json:"environment"`
	Timestamp   time.Time    `json:"timestamp"`
	DB          *db.DBStats  `json:"database,omitempty"`
}

// HealthzHandler responds with service liveness and BoltDB engine telemetry.
func (h *Handler) HealthzHandler(w http.ResponseWriter, r *http.Request) {
	stats, _ := h.store.GetStats()

	resp := HealthResponse{
		Status:      "UP",
		Service:     h.cfg.ServiceName,
		Version:     "1.0.0",
		GoVersion:   runtime.Version(),
		Environment: fmt.Sprintf("gcloud-run (%s)", h.cfg.ServiceName),
		Timestamp:   time.Now().UTC(),
		DB:          stats,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// TelemetryHandler returns live Go runtime stats and memory allocations.
func (h *Handler) TelemetryHandler(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(startTime).Round(time.Second).String()

	resp := map[string]any{
		"service":        h.cfg.ServiceName,
		"uptime":         uptime,
		"alloc_mb":       float64(m.Alloc) / (1024 * 1024),
		"total_alloc_mb": float64(m.TotalAlloc) / (1024 * 1024),
		"sys_mb":         float64(m.Sys) / (1024 * 1024),
		"num_gc":         m.NumGC,
		"goroutines":     runtime.NumGoroutine(),
		"timestamp":      time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// DashaOverviewHandler returns the planetary mahadasha cycles from the unified engine.
func (h *Handler) DashaOverviewHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h.engine.GetOverview())
}

// DashaNakshatrasHandler returns all 27 nakshatras (or filtered by ?planet=) from the unified engine.
func (h *Handler) DashaNakshatrasHandler(w http.ResponseWriter, r *http.Request) {
	planetFilter := r.URL.Query().Get("planet")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h.engine.GetNakshatras(planetFilter))
}

// DashaPlanetsHandler returns all 9 Vimshottari planetary lords from the unified engine.
func (h *Handler) DashaPlanetsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h.engine.GetPlanets())
}

// DashaAlchemyHandler returns Hermetic axioms, sacred metals, and active alchemical state from the unified engine,
// enriched with the active user profile's natal alchemical constitution and symbiotic resonance when available.
func (h *Handler) DashaAlchemyHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	coords := h.parseSolarCoords(r)
	hora := chrono.GetPlanetaryHoraWithCoords(now, coords)
	data := h.engine.GetAlchemy(hora.PlanetID)

	prof := h.parseProfileFromQuery(r)
	if prof == nil {
		prof = h.loadActiveUserProfile()
	}
	if prof != nil {
		data["natal_alchemy"] = prof.Alchemy
		data["active_alchemy"] = prof.ActiveAlchemy
		resonance := dasha.ResolveSymbioticResonance(prof, hora.PlanetID)
		data["symbiotic_resonance"] = resonance
		data["active_profile_id"] = prof.ID
		data["active_profile_name"] = prof.Name
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}


// DashaCalculateHandler computes full Vimshottari Dasha profile and timeline.
func (h *Handler) DashaCalculateHandler(w http.ResponseWriter, r *http.Request) {
	var req dasha.CalculationRequest

	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"invalid JSON request: %s"}`, err.Error()), http.StatusBadRequest)
			return
		}
	} else {
		// Support GET with query params
		q := r.URL.Query()
		req.Name = q.Get("name")
		req.BirthDate = q.Get("birth_date")
		if req.BirthDate == "" {
			req.BirthDate = q.Get("dob")
		}
		req.BirthTime = q.Get("birth_time")
		if tzStr := q.Get("timezone_offset"); tzStr != "" {
			if tz, err := strconv.ParseFloat(tzStr, 64); err == nil {
				req.TimezoneOffset = tz
			}
		}
		if nakStr := q.Get("nakshatra_index"); nakStr != "" {
			if idx, err := strconv.Atoi(nakStr); err == nil {
				req.NakshatraIndex = idx
			}
		}
		if padaStr := q.Get("pada_number"); padaStr != "" {
			if p, err := strconv.Atoi(padaStr); err == nil {
				req.PadaNumber = p
			}
		}
		if degStr := q.Get("moon_degree"); degStr != "" {
			if deg, err := strconv.ParseFloat(degStr, 64); err == nil {
				req.MoonDegree = &deg
			}
		}
		if tTarget := q.Get("target_time"); tTarget != "" {
			req.TargetTime = &tTarget
		}
	}

	if req.BirthDate == "" {
		http.Error(w, `{"error":"birth_date parameter is required (format: YYYY-MM-DD)"}`, http.StatusBadRequest)
		return
	}

	profile, err := dasha.CalculateFromRequest(req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"calculation failed: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(profile)
}

// DashaHoraHandler returns the currently active Chaldean planetary hour (Hora).
func (h *Handler) DashaHoraHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	coords := h.parseSolarCoords(r)
	settings := h.loadActiveSettings()
	hora := chrono.GetPlanetaryHoraWithCoords(now, coords)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"timestamp": now,
		"settings":  settings,
		"hora":      hora,
	})
}

// loadActiveUserProfile loads the sovereign user profile (profile:sovereign-genesis) if available.
func (h *Handler) loadActiveUserProfile() *dasha.DashaProfile {
	if h.engine == nil {
		return nil
	}
	prof, err := h.engine.GetSovereignProfile("profile:sovereign-genesis")
	if err != nil || prof == nil {
		prof, err = h.engine.GetSovereignProfile("sovereign-genesis")
		if err != nil || prof == nil {
			return nil
		}
	}
	return prof
}

// ensureSovereignProfileLocation guarantees that the root user profile Sovereign Genesis has topocentric coordinates.
func (h *Handler) ensureSovereignProfileLocation() {
	if h.engine == nil {
		return
	}
	prof := h.loadActiveUserProfile()
	if prof != nil && prof.Latitude == 0 && prof.Longitude == 0 {
		prof.Latitude = 43.1594
		prof.Longitude = -79.2469
		prof.LocationName = "St. Catharines, ON"
		prof.Timezone = "America/Toronto"
		if prof.TimezoneOffset == 0 {
			prof.TimezoneOffset = -4.0
		}
		_ = h.engine.SaveSovereignProfile(prof)
	}
}

// loadActiveSettings retrieves active system settings, anchoring to the active user profile (Sovereign Genesis) first.
func (h *Handler) loadActiveSettings() config.SystemSettings {
	// 1. Check if user profile Sovereign Genesis defines location data
	if prof := h.loadActiveUserProfile(); prof != nil && (prof.Latitude != 0 || prof.Longitude != 0) {
		tz := prof.Timezone
		if tz == "" {
			tz = "America/Toronto"
		}
		locName := prof.LocationName
		if locName == "" {
			locName = "St. Catharines, ON"
		}
		return config.SystemSettings{
			Timezone:       tz,
			UTCOffsetHours: prof.TimezoneOffset,
			Latitude:       prof.Latitude,
			Longitude:      prof.Longitude,
			AutoDetect:     false,
			LocationName:   locName,
			UpdatedAt:      time.Now().UTC(),
		}
	}

	// 2. Check explicitly persisted system settings
	if h.store != nil {
		data, err := h.store.GetSettings()
		if err == nil && len(data) > 0 {
			var s config.SystemSettings
			if err := json.Unmarshal(data, &s); err == nil {
				return s
			}
		}
	}

	// 3. Fallback to sovereign defaults (St. Catharines / Niagara)
	return config.DefaultSystemSettings()
}

// parseSolarCoords extracts latitude and longitude from request query params, falling back to active user profile and system settings.
func (h *Handler) parseSolarCoords(r *http.Request) chrono.SolarCoordinates {
	q := r.URL.Query()
	latStr := q.Get("lat")
	lonStr := q.Get("lon")
	if latStr != "" && lonStr != "" {
		var lat, lon float64
		if _, err := fmt.Sscanf(latStr, "%f", &lat); err == nil {
			if _, err := fmt.Sscanf(lonStr, "%f", &lon); err == nil {
				return chrono.SolarCoordinates{Latitude: lat, Longitude: lon}
			}
		}
	}
	// Check profile in query if specified
	if prof := h.parseProfileFromQuery(r); prof != nil && (prof.Latitude != 0 || prof.Longitude != 0) {
		return chrono.SolarCoordinates{
			Latitude:  prof.Latitude,
			Longitude: prof.Longitude,
		}
	}
	// Fallback to active settings (which anchors to Sovereign Genesis user profile)
	settings := h.loadActiveSettings()
	return chrono.SolarCoordinates{
		Latitude:  settings.Latitude,
		Longitude: settings.Longitude,
	}
}

// parseProfileFromQuery fetches an optional DashaProfile from store for symbiotic resonance calculations.
func (h *Handler) parseProfileFromQuery(r *http.Request) *dasha.DashaProfile {
	profileID := r.URL.Query().Get("profile_id")
	if profileID == "" || h.engine == nil {
		return nil
	}
	data, err := h.engine.GetProfile(profileID)
	if err != nil || len(data) == 0 {
		if !strings.HasPrefix(profileID, "profile:") {
			data, err = h.engine.GetProfile("profile:" + profileID)
		} else {
			data, err = h.engine.GetProfile(strings.TrimPrefix(profileID, "profile:"))
		}
		if err != nil || len(data) == 0 {
			return nil
		}
	}
	var prof dasha.DashaProfile
	if err := json.Unmarshal(data, &prof); err == nil {
		return &prof
	}
	return nil
}

// HoraActiveHandler returns the currently active planetary hour, timing metrics, and symbiotic resonance.
func (h *Handler) HoraActiveHandler(w http.ResponseWriter, r *http.Request) {
	coords := h.parseSolarCoords(r)
	settings := h.loadActiveSettings()
	targetTime := time.Now().UTC()
	if tStr := r.URL.Query().Get("time"); tStr != "" {
		if t, err := time.Parse(time.RFC3339, tStr); err == nil {
			targetTime = t.UTC()
		}
	}

	profile := h.parseProfileFromQuery(r)
	sched := chrono.CalculateDaySchedule(targetTime, coords, profile)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"timestamp":            targetTime,
		"coordinates":          coords,
		"settings":             settings,
		"active_hour":          sched.ActiveHour,
		"elapsed_minutes":      sched.ElapsedMinutes,
		"remaining_minutes":    sched.RemainingMinutes,
		"progress_percent":     sched.ProgressPercent,
		"next_hour_planet":     sched.NextHourPlanet,
		"next_hour_start_time": sched.NextHourStartTime,
		"day_lord":             sched.DayLord,
		"dasha_resonance":      sched.DashaResonance,
	})
}

// HoraScheduleHandler returns the authoritative 24-hour planetary hour schedule for a given date/time.
func (h *Handler) HoraScheduleHandler(w http.ResponseWriter, r *http.Request) {
	coords := h.parseSolarCoords(r)
	targetTime := time.Now().UTC()
	q := r.URL.Query()
	if dateStr := q.Get("date"); dateStr != "" {
		if d, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetTime = time.Date(d.Year(), d.Month(), d.Day(), 12, 0, 0, 0, time.UTC)
		}
	}
	if tStr := q.Get("time"); tStr != "" {
		if t, err := time.Parse(time.RFC3339, tStr); err == nil {
			targetTime = t.UTC()
		}
	}

	profile := h.parseProfileFromQuery(r)
	sched := chrono.CalculateDaySchedule(targetTime, coords, profile)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sched)
}

// GetSettingsHandler returns the current system settings (timezone, coords, auto-detect).
func (h *Handler) GetSettingsHandler(w http.ResponseWriter, r *http.Request) {
	settings := h.loadActiveSettings()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(settings)
}

// SaveSettingsHandler updates and persists system settings to BoltDB.
func (h *Handler) SaveSettingsHandler(w http.ResponseWriter, r *http.Request) {
	var input config.SystemSettings
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error":"invalid json body"}`, http.StatusBadRequest)
		return
	}

	if input.Timezone == "" {
		input.Timezone = "America/Toronto"
	}
	// If coords are zero, resolve centroid from timezone
	if input.Latitude == 0 && input.Longitude == 0 {
		if centroid, ok := config.ResolveCentroid(input.Timezone); ok {
			input.Latitude = centroid.Latitude
			input.Longitude = centroid.Longitude
			if input.LocationName == "" {
				input.LocationName = centroid.Name
			}
		}
	}
	input.UpdatedAt = time.Now().UTC()

	data, err := json.Marshal(input)
	if err != nil {
		http.Error(w, `{"error":"failed to serialize settings"}`, http.StatusInternalServerError)
		return
	}

	if h.store != nil {
		if err := h.store.SaveSettings(data); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"failed to save settings: %s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
	}

	// Synchronize location to root user profile Sovereign Genesis
	if h.engine != nil {
		if profData, err := h.engine.GetProfile("profile:sovereign-genesis"); err == nil && len(profData) > 0 {
			var prof dasha.DashaProfile
			if err := json.Unmarshal(profData, &prof); err == nil {
				prof.Latitude = input.Latitude
				prof.Longitude = input.Longitude
				prof.LocationName = input.LocationName
				prof.Timezone = input.Timezone
				prof.TimezoneOffset = input.UTCOffsetHours
				if updatedProfData, err := json.Marshal(prof); err == nil {
					_ = h.engine.SaveProfile("profile:sovereign-genesis", updatedProfData)
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":  true,
		"settings": input,
	})
}

// ChronoPulseHandler streams real-time Server-Sent Events heartbeat.
func (h *Handler) ChronoPulseHandler(w http.ResponseWriter, r *http.Request) {
	if h.metronome == nil {
		http.Error(w, `{"error":"metronome stream unavailable"}`, http.StatusInternalServerError)
		return
	}
	h.metronome.SSEHandler(w, r)
}

// Profile Persistence Handlers
func (h *Handler) ListProfilesHandler(w http.ResponseWriter, r *http.Request) {
	summaries, err := h.engine.ListSovereignProfileSummaries()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to list profiles: %s"}`, err.Error()), http.StatusInternalServerError)
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


func (h *Handler) CreateProfileHandler(w http.ResponseWriter, r *http.Request) {
	var req dasha.CalculationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid request: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	if req.BirthDate == "" {
		http.Error(w, `{"error":"birth_date is required"}`, http.StatusBadRequest)
		return
	}

	prof, err := h.engine.Calculate(req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"calculation failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	if prof.ID == "" {
		prof.ID = "profile:" + strings.ToLower(strings.ReplaceAll(prof.Name, " ", "-"))
	}
	if !strings.HasPrefix(prof.ID, "profile:") {
		prof.ID = "profile:" + prof.ID
	}

	if err := h.engine.SaveSovereignProfile(prof); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to persist profile: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(prof)
}

func (h *Handler) GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"id parameter is required"}`, http.StatusBadRequest)
		return
	}

	// Sub-resource: /resonance
	if strings.HasSuffix(id, "/resonance") {
		cleanID := strings.TrimSuffix(id, "/resonance")
		coords := h.parseSolarCoords(r)
		hora := chrono.GetPlanetaryHoraWithCoords(time.Now().UTC(), coords)
		res, err := h.engine.GetProfileResonance(cleanID, hora.PlanetID)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				http.Error(w, `{"error":"profile not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf(`{"error":"resonance calculation failed: %s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
		return
	}

	// Sub-resource: /timeline
	if strings.HasSuffix(id, "/timeline") {
		cleanID := strings.TrimSuffix(id, "/timeline")
		tl, err := h.engine.GetDetailedTimeline(cleanID)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				http.Error(w, `{"error":"profile not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf(`{"error":"timeline lookup failed: %s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"profile_id": cleanID,
			"timeline":   tl,
		})
		return
	}

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
			http.Error(w, `{"error":"profile not found"}`, http.StatusNotFound)
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

func (h *Handler) SaveProfileHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"id parameter is required"}`, http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 5*1024*1024))
	if err != nil {
		http.Error(w, `{"error":"failed to read body"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var prof dasha.DashaProfile
	if err := json.Unmarshal(body, &prof); err != nil {
		http.Error(w, `{"error":"invalid JSON profile payload"}`, http.StatusBadRequest)
		return
	}

	if prof.ID == "" {
		prof.ID = id
	}
	if !strings.HasPrefix(prof.ID, "profile:") {
		prof.ID = "profile:" + prof.ID
	}

	dasha.PopulateUnifiedMatrices(&prof)

	if err := h.engine.SaveSovereignProfile(&prof); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to save profile: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// If this is the root sovereign profile, synchronize its location into system settings
	cleanID := strings.TrimPrefix(prof.ID, "profile:")
	if cleanID == "sovereign-genesis" || cleanID == "sovereign" {
		if prof.Latitude != 0 || prof.Longitude != 0 {
			settings := config.SystemSettings{
				Timezone:       prof.Timezone,
				UTCOffsetHours: prof.TimezoneOffset,
				Latitude:       prof.Latitude,
				Longitude:      prof.Longitude,
				LocationName:   prof.LocationName,
				AutoDetect:     false,
				UpdatedAt:      time.Now().UTC(),
			}
			if settings.Timezone == "" {
				settings.Timezone = "America/Toronto"
			}
			if settings.LocationName == "" {
				settings.LocationName = "St. Catharines, ON"
			}
			if sData, err := json.Marshal(settings); err == nil && h.store != nil {
				_ = h.store.SaveSettings(sData)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"id":      prof.ID,
		"profile": prof,
	})
}

func (h *Handler) DeleteProfileHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"id parameter is required"}`, http.StatusBadRequest)
		return
	}

	if err := h.engine.DeleteProfile(id); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to delete profile: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"success":true,"deleted":"` + id + `"}`))
}


// Meta JSON Handlers
func (h *Handler) GetMetaHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, `{"error":"key parameter is required"}`, http.StatusBadRequest)
		return
	}

	data, err := h.store.GetJSON(db.BucketMeta, key)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			http.Error(w, `{"error":"document not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf(`{"error":"internal storage error: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *Handler) PutMetaHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, `{"error":"key parameter is required"}`, http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 5*1024*1024))
	if err != nil {
		http.Error(w, `{"error":"failed to read request body"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := h.store.PutJSON(db.BucketMeta, key, body); err != nil {
		if errors.Is(err, db.ErrInvalidJSON) {
			http.Error(w, `{"error":"invalid JSON body payload"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf(`{"error":"failed to write document: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"success":true,"key":"` + key + `"}`))
}

func (h *Handler) DeleteMetaHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, `{"error":"key parameter is required"}`, http.StatusBadRequest)
		return
	}

	if err := h.store.Delete(db.BucketMeta, key); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to delete: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"success":true,"deleted":"` + key + `"}`))
}

func (h *Handler) ListMetaHandler(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	keys, err := h.store.ListKeys(db.BucketMeta, prefix)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to scan keys: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if keys == nil {
		keys = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"keys":   keys,
		"count":  len(keys),
		"prefix": prefix,
	})
}

func (h *Handler) BackupHandler(w http.ResponseWriter, r *http.Request) {
	filename := fmt.Sprintf("mercury-dasha-backup-%s.db", time.Now().UTC().Format("20060102-150405"))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))

	if err := h.store.Backup(w); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"snapshot failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
}

// Dropbox Handlers

// DropboxStatusHandler returns connection status, configured paths, and account info.
func (h *Handler) DropboxStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.dropbox == nil || !h.dropbox.IsConfigured() {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"configured": false,
			"base_path":  "/MercuryDasha/backups",
			"message":    "Dropbox integration is not configured. Set DROPBOX_ACCESS_TOKEN or DROPBOX_REFRESH_TOKEN.",
		})
		return
	}

	account, err := h.dropbox.GetAccountInfo(r.Context())
	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"configured": true,
			"base_path":  h.dropbox.GetBasePath(),
			"account":    nil,
			"error":      err.Error(),
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"configured": true,
		"base_path":  h.dropbox.GetBasePath(),
		"account":    account,
	})
}

// DropboxBackupHandler creates a BoltDB snapshot and uploads it directly to Dropbox.
func (h *Handler) DropboxBackupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.dropbox == nil || !h.dropbox.IsConfigured() {
		http.Error(w, `{"error":"dropbox integration is not configured"}`, http.StatusBadRequest)
		return
	}

	// 1. Capture snapshot in buffer
	var buf bytes.Buffer
	if err := h.store.Backup(&buf); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"snapshot generation failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// 2. Upload snapshot to Dropbox
	filename := fmt.Sprintf("mercury-dasha-backup-%s.db", time.Now().UTC().Format("20060102-150405"))
	targetPath := fmt.Sprintf("%s/%s", h.dropbox.GetBasePath(), filename)

	entry, err := h.dropbox.UploadFile(r.Context(), targetPath, &buf)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"dropbox upload failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":         true,
		"filename":        filename,
		"path":            entry.PathDisplay,
		"size_bytes":      entry.Size,
		"server_modified": entry.ServerModified,
	})
}

// DropboxFilesHandler lists backup snapshots in the Dropbox backup folder.
func (h *Handler) DropboxFilesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.dropbox == nil || !h.dropbox.IsConfigured() {
		http.Error(w, `{"error":"dropbox integration is not configured"}`, http.StatusBadRequest)
		return
	}

	folder := r.URL.Query().Get("path")
	entries, err := h.dropbox.ListFolder(r.Context(), folder)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"dropbox list_folder failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	if entries == nil {
		entries = []dropbox.FileEntry{}
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"base_path": h.dropbox.GetBasePath(),
		"entries":   entries,
		"count":     len(entries),
	})
}

// DropboxLinkHandler generates a temporary download link for a file.
func (h *Handler) DropboxLinkHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.dropbox == nil || !h.dropbox.IsConfigured() {
		http.Error(w, `{"error":"dropbox integration is not configured"}`, http.StatusBadRequest)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, `{"error":"path parameter is required"}`, http.StatusBadRequest)
		return
	}

	link, err := h.dropbox.GetTemporaryLink(r.Context(), path)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to generate download link: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"path": path,
		"link": link,
	})
}

// DropboxSearchHandler searches for files across Dropbox matching query ?q=.
func (h *Handler) DropboxSearchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.dropbox == nil || !h.dropbox.IsConfigured() {
		http.Error(w, `{"error":"dropbox integration is not configured"}`, http.StatusBadRequest)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, `{"error":"query parameter 'q' is required"}`, http.StatusBadRequest)
		return
	}

	folder := r.URL.Query().Get("path")
	matches, err := h.dropbox.SearchFiles(r.Context(), query, folder, 25)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"dropbox search failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	if matches == nil {
		matches = []dropbox.SearchMatch{}
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"query":   query,
		"path":    folder,
		"matches": matches,
		"count":   len(matches),
	})
}

// DropboxContentHandler retrieves or streams the raw content of a file from Dropbox.
func (h *Handler) DropboxContentHandler(w http.ResponseWriter, r *http.Request) {
	if h.dropbox == nil || !h.dropbox.IsConfigured() {
		http.Error(w, `{"error":"dropbox integration is not configured"}`, http.StatusBadRequest)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, `{"error":"path parameter is required"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", path))

	if _, err := h.dropbox.DownloadFile(r.Context(), path, w); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"dropbox download failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
}

// DropboxMetaHandler returns detailed metadata for a given Dropbox path.
func (h *Handler) DropboxMetaHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.dropbox == nil || !h.dropbox.IsConfigured() {
		http.Error(w, `{"error":"dropbox integration is not configured"}`, http.StatusBadRequest)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, `{"error":"path parameter is required"}`, http.StatusBadRequest)
		return
	}

	meta, err := h.dropbox.GetFileMetadata(r.Context(), path)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to get metadata: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(meta)
}

// IndexScanHandler triggers an asynchronous background scan of Dropbox folders.
func (h *Handler) IndexScanHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"POST method required"}`, http.StatusMethodNotAllowed)
		return
	}

	catParam := r.URL.Query().Get("category")
	var categories []string
	if catParam != "" {
		for _, c := range strings.Split(catParam, ",") {
			cTrim := strings.TrimSpace(c)
			if cTrim != "" {
				categories = append(categories, cTrim)
			}
		}
	}

	started, err := h.indexer.StartScan(categories)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"started": false,
			"error":   err.Error(),
			"status":  h.indexer.GetStatus(),
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"started":    started,
		"categories": categories,
		"status":     h.indexer.GetStatus(),
	})
}

// IndexStatusHandler reports current indexing progress and file counts.
func (h *Handler) IndexStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h.indexer.GetStatus())
}

// IndexSearchHandler performs indexed queries across text, audio, video, code, and books.
func (h *Handler) IndexSearchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query().Get("q")
	cat := r.URL.Query().Get("category")
	ext := r.URL.Query().Get("ext")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}
	offset := 0
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	filter := db.IndexFilter{
		Query:    query,
		Category: cat,
		Ext:      ext,
		Limit:    limit,
		Offset:   offset,
	}

	results, total, err := h.store.SearchIndex(filter)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"search failed: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"query":    query,
		"category": cat,
		"ext":      ext,
		"limit":    limit,
		"offset":   offset,
		"total":    total,
		"results":  results,
	})
}

// IndexEntryHandler retrieves detailed metadata for a single indexed file.
func (h *Handler) IndexEntryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error":"id parameter is required"}`, http.StatusBadRequest)
		return
	}

	entry, err := h.store.GetIndexEntry(id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			http.Error(w, `{"error":"entry not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf(`{"error":"failed to retrieve entry: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(entry)
}

// IndexContentHandler is a hybrid streamer: serves from local filesystem if present (with HTTP Range seeking),
// or falls back to streaming from Dropbox Cloud API if remote.
// Supports transparent on-the-fly transcoding of legacy AMR audio files into standard PCM WAV.
func (h *Handler) IndexContentHandler(w http.ResponseWriter, r *http.Request) {
	relPath := r.URL.Query().Get("path")
	if relPath == "" {
		relPath = r.URL.Query().Get("id")
	}
	if relPath == "" {
		http.Error(w, `{"error":"path or id parameter is required"}`, http.StatusBadRequest)
		return
	}

	// Sanitize relPath
	relPath = strings.TrimPrefix(relPath, "/")
	cleanRel := filepath.Clean(relPath)

	isAMR := strings.EqualFold(filepath.Ext(cleanRel), ".amr")
	rawRequested := r.URL.Query().Get("raw") == "true" || r.URL.Query().Get("download") == "1"

	// 1. Try local filesystem if configured
	localRoot := h.cfg.DropboxLocalPath
	if localRoot == "" {
		localRoot = "/home/justin/Dropbox"
	}

	fullLocalPath := filepath.Join(localRoot, cleanRel)
	if fi, err := os.Stat(fullLocalPath); err == nil && !fi.IsDir() {
		// If AMR audio and raw not explicitly requested, transcode/serve cached WAV
		if isAMR && !rawRequested {
			cacheDir := filepath.Join(".data", "cache", "audio")
			if cachedWav, err := amr.GetOrCreateCachedWAV(cacheDir, cleanRel, fullLocalPath); err == nil {
				w.Header().Set("Content-Type", "audio/wav")
				w.Header().Set("X-Mercury-Transcoded", "amr-to-wav")
				http.ServeFile(w, r, cachedWav)
				return
			} else {
				log.Printf("[IndexContentHandler] AMR transcode error for %s: %v", cleanRel, err)
			}
		}

		// Native file server handles Range headers, Content-Type, Last-Modified, etc.
		http.ServeFile(w, r, fullLocalPath)
		return
	}

	// 1b. Check local working directory / project root (e.g. renders/ or storage/)
	if fi, err := os.Stat(cleanRel); err == nil && !fi.IsDir() {
		http.ServeFile(w, r, cleanRel)
		return
	}

	// 2. Cloud Fallback if local file not found
	if h.dropbox != nil && h.dropbox.IsConfigured() {
		remotePath := "/" + cleanRel

		// If remote AMR file and raw not requested, download to cache and transcode to WAV
		if isAMR && !rawRequested {
			cacheDir := filepath.Join(".data", "cache", "audio")
			_ = os.MkdirAll(cacheDir, 0755)
			hash := sha256.Sum256([]byte(cleanRel))
			tmpAmr := filepath.Join(cacheDir, fmt.Sprintf("cloud_%x.amr", hash[:6]))
			if f, err := os.Create(tmpAmr); err == nil {
				if _, err := h.dropbox.DownloadFile(r.Context(), remotePath, f); err == nil {
					_ = f.Close()
					if cachedWav, err := amr.GetOrCreateCachedWAV(cacheDir, cleanRel, tmpAmr); err == nil {
						w.Header().Set("Content-Type", "audio/wav")
						w.Header().Set("X-Mercury-Transcoded", "amr-to-wav")
						http.ServeFile(w, r, cachedWav)
						return
					}
				} else {
					_ = f.Close()
				}
			}
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", filepath.Base(cleanRel)))
		if _, err := h.dropbox.DownloadFile(r.Context(), remotePath, w); err == nil {
			return
		}
	}

	http.Error(w, fmt.Sprintf(`{"error":"file %q not found locally or in cloud"}`, cleanRel), http.StatusNotFound)
}

// GetEsotericCatalogHandler lists available esoteric document keys in BoltDB.
func (h *Handler) GetEsotericCatalogHandler(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	if prefix == "" {
		prefix = "esoteric:"
	}
	keys, err := h.store.ListEsotericContent(prefix)
	if err != nil {
		http.Error(w, `{"error":"failed to list esoteric documents"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"keys":   keys,
		"count":  len(keys),
		"prefix": prefix,
		"source": "boltdb:esoteric_content",
	})
}

// GetArishadvargaHandler returns the 6 classical inner adversaries and their transmutations.
func (h *Handler) GetArishadvargaHandler(w http.ResponseWriter, r *http.Request) {
	raw, err := h.store.GetEsotericContent("esoteric:arishadvarga")
	if err != nil {
		// Fallback to default demons if not yet seeded
		demons := db.DefaultArishadvargaDemons()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"demons": demons,
			"count":  len(demons),
			"source": "default_seed",
		})
		return
	}

	var demons []db.ArishadvargaDemon
	if err := json.Unmarshal(raw, &demons); err != nil {
		http.Error(w, `{"error":"failed to parse arishadvarga catalog"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"demons": demons,
		"count":  len(demons),
		"source": "boltdb:esoteric_content",
	})
}

// GetEsotericDocHandler retrieves a specific esoteric document by key.
func (h *Handler) GetEsotericDocHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, `{"error":"document key is required"}`, http.StatusBadRequest)
		return
	}

	fullKey := key
	if !strings.HasPrefix(fullKey, "esoteric:") {
		fullKey = "esoteric:" + fullKey
	}

	data, err := h.store.GetEsotericContent(fullKey)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			http.Error(w, `{"error":"esoteric document not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// RecorderStatusHandler returns Google Recorder MCP server status and local sync stats.
func (h *Handler) RecorderStatusHandler(w http.ResponseWriter, r *http.Request) {
	if h.grecorder == nil {
		http.Error(w, `{"error":"recorder service not initialized"}`, http.StatusServiceUnavailable)
		return
	}
	status := h.grecorder.GetStatus(r.Context())
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}

// RecorderSyncHandler triggers on-demand synchronization of recordings via MCP into Dropbox.
func (h *Handler) RecorderSyncHandler(w http.ResponseWriter, r *http.Request) {
	if h.grecorder == nil {
		http.Error(w, `{"error":"recorder service not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	opts := grecorder.SyncOptions{
		All:                true,
		DownloadAudio:      true,
		DownloadTranscript: true,
		TranscriptFormat:   "both",
		DownloadMetadata:   true,
	}

	// Parse JSON body if present
	if r.Header.Get("Content-Type") == "application/json" && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&opts)
	}

	// Override with query params if supplied
	if l := r.URL.Query().Get("limit"); l != "" {
		if iv, err := strconv.Atoi(l); err == nil {
			opts.Limit = iv
			opts.All = false
		}
	}
	if q := r.URL.Query().Get("query"); q != "" {
		opts.Query = q
	}
	if r.URL.Query().Get("force") == "true" {
		opts.Force = true
	}

	syncStatus, err := h.grecorder.Sync(r.Context(), opts)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":  err.Error(),
			"status": syncStatus,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(syncStatus)
}

// RecorderSyncStatusHandler reports current or last sync progress.
func (h *Handler) RecorderSyncStatusHandler(w http.ResponseWriter, r *http.Request) {
	if h.grecorder == nil {
		http.Error(w, `{"error":"recorder service not initialized"}`, http.StatusServiceUnavailable)
		return
	}
	status := h.grecorder.GetSyncStatus()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}

// RecorderRecordingsHandler lists recordings from local manifest merged with remote MCP catalog.
func (h *Handler) RecorderRecordingsHandler(w http.ResponseWriter, r *http.Request) {
	if h.grecorder == nil {
		http.Error(w, `{"error":"recorder service not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		q = r.URL.Query().Get("query")
	}

	recordings, err := h.grecorder.ListRecordings(r.Context(), q)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"recordings": recordings,
		"total":      len(recordings),
	})
}

// RecorderTranscriptHandler returns formatted diarized paragraphs and raw text for a recording.
func (h *Handler) RecorderTranscriptHandler(w http.ResponseWriter, r *http.Request) {
	if h.grecorder == nil {
		http.Error(w, `{"error":"recorder service not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	id := r.PathValue("id")
	id = strings.TrimPrefix(id, "/")
	id = strings.TrimSuffix(id, "/transcript")
	id = strings.TrimSuffix(id, "/audio")
	if id == "" {
		id = r.URL.Query().Get("id")
	}
	if id == "" {
		http.Error(w, `{"error":"recording id is required"}`, http.StatusBadRequest)
		return
	}

	transcript, err := h.grecorder.GetTranscript(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(transcript)
}

// RecorderAudioHandler streams the local .m4a audio file with HTTP 206 Range seeking.
func (h *Handler) RecorderAudioHandler(w http.ResponseWriter, r *http.Request) {
	if h.grecorder == nil {
		http.Error(w, `{"error":"recorder service not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	id := r.PathValue("id")
	id = strings.TrimPrefix(id, "/")
	id = strings.TrimSuffix(id, "/audio")
	id = strings.TrimSuffix(id, "/transcript")
	if id == "" {
		id = r.URL.Query().Get("id")
	}
	if id == "" {
		http.Error(w, `{"error":"recording id is required"}`, http.StatusBadRequest)
		return
	}

	filePath, err := h.grecorder.GetAudioFilePath(id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "audio/mp4")
	w.Header().Set("Accept-Ranges", "bytes")
	http.ServeFile(w, r, filePath)
}

// UnifiedAudioItem represents an audio asset from any source in Mercury Dasha.
type UnifiedAudioItem struct {
	ID                string         `json:"id"`
	Title             string         `json:"title"`
	Source            string         `json:"source"` // "recorder" | "vault" | "session"
	Category          string         `json:"category"`
	Format            string         `json:"format"` // "m4a", "wav", "mp3", "amr", "flac", "ogg"
	Duration          string         `json:"duration,omitempty"`
	DurationMs        int64          `json:"duration_ms,omitempty"`
	SizeBytes         int64          `json:"size_bytes"`
	RecordedAt        string         `json:"recorded_at,omitempty"`
	ModTime           string         `json:"mod_time,omitempty"`
	HasTranscript     bool           `json:"has_transcript"`
	IsSynced          bool           `json:"is_synced"`
	AudioURL          string         `json:"audio_url"`
	TranscriptURL     string         `json:"transcript_url,omitempty"`
	TranscriptSnippet string         `json:"transcript_snippet,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
	Tags              []string       `json:"tags,omitempty"`
	Path              string         `json:"path"`
}

// UnifiedAudioCatalogResponse wraps the unified library response with statistics.
type UnifiedAudioCatalogResponse struct {
	Items  []UnifiedAudioItem `json:"items"`
	Total  int                `json:"total"`
	Counts struct {
		Total          int            `json:"total"`
		WithTranscript int            `json:"with_transcript"`
		Recorder       int            `json:"recorder"`
		Vault          int            `json:"vault"`
		Sessions       int            `json:"sessions"`
		Formats        map[string]int `json:"formats"`
	} `json:"counts"`
}

// UnifiedAudioCatalogHandler synthesizes Google Recorder items and Dropbox vault audio into one catalog.
func (h *Handler) UnifiedAudioCatalogHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	filterSource := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("source")))
	filterFormat := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	filterTranscript := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("has_transcript")))

	limit := 200
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

	var allItems []UnifiedAudioItem
	knownPaths := make(map[string]bool)
	knownBaseNames := make(map[string]bool)

	// 1. Ingest Google Recorder items
	if h.grecorder != nil {
		recList, err := h.grecorder.ListRecordings(r.Context(), "")
		if err == nil {
			for _, rec := range recList {
				var sizeBytes int64
				if rec.AudioPath != "" {
					knownPaths[rec.AudioPath] = true
					knownBaseNames[filepath.Base(rec.AudioPath)] = true
					if info, statErr := os.Stat(rec.AudioPath); statErr == nil {
						sizeBytes = info.Size()
					}
				}

				item := UnifiedAudioItem{
					ID:            rec.ID,
					Title:         rec.Title,
					Source:        "recorder",
					Category:      "voice_chronicle",
					Format:        "m4a",
					Duration:      rec.Duration,
					DurationMs:    rec.DurationMs,
					SizeBytes:     sizeBytes,
					RecordedAt:    rec.RecordedAt.Format(time.RFC3339),
					ModTime:       rec.RecordedAt.Format(time.RFC3339),
					HasTranscript: rec.HasTranscript,
					IsSynced:      rec.IsSynced,
					AudioURL:      fmt.Sprintf("/api/v1/recorder/recordings/%s/audio", rec.ID),
					TranscriptURL: fmt.Sprintf("/api/v1/recorder/recordings/%s/transcript", rec.ID),
					Path:          rec.AudioPath,
				}
				allItems = append(allItems, item)
			}
		}
	}

	// 2. Ingest BoltDB Indexed Audio Vault Files
	if h.store != nil {
		vaultDocs, _, err := h.store.SearchIndex(db.IndexFilter{
			Category: "audio",
			Limit:    2000,
		})
		if err == nil {
			for _, doc := range vaultDocs {
				if knownPaths[doc.FullPath] || knownBaseNames[doc.FileName] {
					continue
				}
				if strings.Contains(doc.Path, "audio/recorder") {
					continue
				}
				// CRITICAL FIX: Skip any studio / D&D session audio files from general index.
				// Step 3 (h.sessions.ListSessions) is the authoritative single source of truth for all sessions.
				if strings.HasPrefix(doc.Path, "MercuryDasha/sessions/") || strings.Contains(doc.Path, "sessions/dnd") {
					continue
				}

				knownPaths[doc.FullPath] = true
				knownPaths[doc.Path] = true
				knownBaseNames[doc.FileName] = true

				src := "vault"
				ext := strings.TrimPrefix(strings.ToLower(doc.Extension), ".")
				if ext == "" {
					ext = "wav"
				}

				// Check for sidecar transcript (.txt or .json with same base name)
				hasTranscript := false
				baseWithoutExt := strings.TrimSuffix(doc.FullPath, filepath.Ext(doc.FullPath))
				if _, statErr := os.Stat(baseWithoutExt + ".txt"); statErr == nil {
					hasTranscript = true
				} else if _, statErr := os.Stat(baseWithoutExt + ".json"); statErr == nil {
					hasTranscript = true
				}

				title := ""
				if t, ok := doc.Metadata["title"].(string); ok && t != "" {
					title = t
				} else {
					baseName := filepath.Base(doc.FileName)
					nameNoExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
					title = strings.ReplaceAll(nameNoExt, "_", " ")
					title = strings.ReplaceAll(title, "-", " ")
				}

				var duration string
				var durationMs int64
				if dur, ok := doc.Metadata["duration_sec"].(float64); ok && dur > 0 {
					durationMs = int64(dur * 1000)
					mins := int(dur) / 60
					secs := int(dur) % 60
					duration = fmt.Sprintf("%02d:%02d", mins, secs)
				}

				item := UnifiedAudioItem{
					ID:            doc.ID,
					Title:         title,
					Source:        src,
					Category:      "audio_vault",
					Format:        ext,
					Duration:      duration,
					DurationMs:    durationMs,
					SizeBytes:     doc.SizeBytes,
					ModTime:       doc.ModTime.Format(time.RFC3339),
					RecordedAt:    doc.ModTime.Format(time.RFC3339),
					HasTranscript: hasTranscript,
					IsSynced:      true,
					AudioURL:      fmt.Sprintf("/api/v1/index/content?path=%s", url.QueryEscape(doc.Path)),
					Tags:          doc.Tags,
					Metadata:      doc.Metadata,
					Path:          doc.Path,
				}
				allItems = append(allItems, item)
			}
		}
	}

	// 3. Ingest D&D / Studio Sessions from session store
	if h.sessions != nil {
		sessList, err := h.sessions.ListSessions("")
		if err == nil {
			for _, sess := range sessList {
				// Filter out stale / abandoned 0-byte recording sessions
				if sess.Status == "recording" && sess.SizeBytes == 0 {
					continue
				}
				if knownPaths[sess.FilePath] || (sess.ID != "" && knownPaths[sess.ID]) {
					continue
				}
				knownPaths[sess.FilePath] = true
				if sess.ID != "" {
					knownPaths[sess.ID] = true
				}

				ext := "wav"
				if strings.Contains(sess.Format, "webm") {
					ext = "webm"
				} else if strings.Contains(sess.Format, "flac") {
					ext = "flac"
				} else if strings.Contains(sess.Format, "mp3") {
					ext = "mp3"
				}

				var duration string
				var durationMs int64
				if sess.DurationSec > 0 {
					durationMs = int64(sess.DurationSec * 1000)
					mins := int(sess.DurationSec) / 60
					secs := int(sess.DurationSec) % 60
					duration = fmt.Sprintf("%02d:%02d", mins, secs)
				}

				recTime := sess.StartTime.Format(time.RFC3339)
				if sess.StartTime.IsZero() {
					recTime = sess.CreatedAt.Format(time.RFC3339)
				}

				item := UnifiedAudioItem{
					ID:            sess.ID,
					Title:         sess.Title,
					Source:        "session",
					Category:      "studio_session",
					Format:        ext,
					Duration:      duration,
					DurationMs:    durationMs,
					SizeBytes:     sess.SizeBytes,
					ModTime:       sess.UpdatedAt.Format(time.RFC3339),
					RecordedAt:    recTime,
					HasTranscript: false,
					IsSynced:      true,
					AudioURL:      fmt.Sprintf("/api/v1/sessions/%s/stream", sess.ID),
					Tags:          sess.Tags,
					Path:          sess.FilePath,
				}
				allItems = append(allItems, item)
			}
		}
	}

	// 4. Compute Counts across entire collection
	var resp UnifiedAudioCatalogResponse
	resp.Counts.Formats = make(map[string]int)
	for _, it := range allItems {
		resp.Counts.Total++
		if it.HasTranscript {
			resp.Counts.WithTranscript++
		}
		switch it.Source {
		case "recorder":
			resp.Counts.Recorder++
		case "vault":
			resp.Counts.Vault++
		case "session":
			resp.Counts.Sessions++
		}
		resp.Counts.Formats[it.Format]++
	}

	// 4. Apply Filters
	var filtered []UnifiedAudioItem
	for _, it := range allItems {
		if filterSource != "" && filterSource != "all" && it.Source != filterSource {
			continue
		}
		if filterFormat != "" && filterFormat != "all" && it.Format != filterFormat {
			continue
		}
		if filterTranscript == "true" && !it.HasTranscript {
			continue
		}
		if filterTranscript == "false" && it.HasTranscript {
			continue
		}
		if q != "" {
			match := strings.Contains(strings.ToLower(it.Title), q) ||
				strings.Contains(strings.ToLower(it.Path), q) ||
				strings.Contains(strings.ToLower(it.Format), q)
			if !match {
				continue
			}
		}
		filtered = append(filtered, it)
	}

	// 5. Sort Descending by ModTime / RecordedAt
	sort.Slice(filtered, func(i, j int) bool {
		ti := filtered[i].RecordedAt
		if ti == "" {
			ti = filtered[i].ModTime
		}
		tj := filtered[j].RecordedAt
		if tj == "" {
			tj = filtered[j].ModTime
		}
		return ti > tj
	})

	resp.Total = len(filtered)
	if offset > len(filtered) {
		resp.Items = []UnifiedAudioItem{}
	} else {
		end := offset + limit
		if end > len(filtered) {
			end = len(filtered)
		}
		resp.Items = filtered[offset:end]
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// AudioTranscribePlaceholderHandler handles future on-demand transcription requests.
func (h *Handler) AudioTranscribePlaceholderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		ID   string `json:"id"`
		Path string `json:"path"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":      "queued_placeholder",
		"message":     "On-demand transcription scheduled for future Voice AI pipeline. Asset is registered in audio catalog.",
		"id":          req.ID,
		"path":        req.Path,
		"enqueued_at": time.Now().UTC().Format(time.RFC3339),
	})
}

// SyncthingStatusHandler reports aggregated Syncthing subsystem health and telemetry.
func (h *Handler) SyncthingStatusHandler(w http.ResponseWriter, r *http.Request) {
	if h.syncthingH == nil {
		http.Error(w, "syncthing subsystem not initialized", http.StatusServiceUnavailable)
		return
	}
	h.syncthingH.StatusHandler(w, r)
}

// SyncthingFeedHandler returns recent synchronized media items.
func (h *Handler) SyncthingFeedHandler(w http.ResponseWriter, r *http.Request) {
	if h.syncthingH == nil {
		http.Error(w, "syncthing subsystem not initialized", http.StatusServiceUnavailable)
		return
	}
	h.syncthingH.FeedHandler(w, r)
}

// SyncthingRescanHandler triggers an immediate rescan of media repositories.
func (h *Handler) SyncthingRescanHandler(w http.ResponseWriter, r *http.Request) {
	if h.syncthingH == nil {
		http.Error(w, "syncthing subsystem not initialized", http.StatusServiceUnavailable)
		return
	}
	h.InvalidateVisualCache()
	h.syncthingH.RescanHandler(w, r)
}

// MediaStreamHandler streams high-resolution media files with byte-range HTTP seeking.
func (h *Handler) MediaStreamHandler(w http.ResponseWriter, r *http.Request) {
	if h.syncthingH == nil {
		http.Error(w, "syncthing subsystem not initialized", http.StatusServiceUnavailable)
		return
	}
	h.syncthingH.StreamHandler(w, r)
}

// MediaThumbnailHandler serves generated poster thumbnails.
func (h *Handler) MediaThumbnailHandler(w http.ResponseWriter, r *http.Request) {
	if h.syncthingH == nil {
		http.Error(w, "syncthing subsystem not initialized", http.StatusServiceUnavailable)
		return
	}
	h.syncthingH.ThumbnailHandler(w, r)
}

// InvalidateVisualCache clears the in-memory visual catalog cache.
func (h *Handler) InvalidateVisualCache() {
	h.visualMu.Lock()
	defer h.visualMu.Unlock()
	h.visualCache = nil
}

// UnifiedVisualItem represents a photo or video asset across Pixel 7 mobile sync and Dropbox archives.
type UnifiedVisualItem struct {
	ID              string         `json:"id"`
	Title           string         `json:"title"`
	FileName        string         `json:"file_name"`
	Source          string         `json:"source"` // "pixel7" | "dropbox_photos" | "dropbox_videos" | "vault"
	Category        string         `json:"category"` // "photos" | "video"
	Format          string         `json:"format"` // "jpg", "png", "mp4", "mov", "webm", etc.
	Path            string         `json:"path"`
	FullPath        string         `json:"full_path"`
	SizeBytes       int64          `json:"size_bytes"`
	ModTime         string         `json:"mod_time"`
	MediaDate       string         `json:"media_date,omitempty"`
	Year            int            `json:"year,omitempty"`
	DurationSec     float64        `json:"duration_sec,omitempty"`
	DurationStr     string         `json:"duration_str,omitempty"`
	Resolution      string         `json:"resolution,omitempty"`
	ThumbnailURL    string         `json:"thumbnail_url"`
	StreamURL       string         `json:"stream_url"`
	Tags            []string       `json:"tags,omitempty"`
	DashaMahadasha  string         `json:"dasha_mahadasha,omitempty"`
	DashaAntardasha string         `json:"dasha_antardasha,omitempty"`
	SacredMetal     string         `json:"sacred_metal,omitempty"`
	Hora            string         `json:"hora,omitempty"`
	HermeticAxiom   string         `json:"hermetic_axiom,omitempty"`
	DeviceModel     string         `json:"device_model,omitempty"`
	Location        string         `json:"location,omitempty"`
	Album           string         `json:"album,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

// UnifiedVisualCatalogResponse wraps the federated visual catalog with aggregated facets.
type UnifiedVisualCatalogResponse struct {
	Items  []UnifiedVisualItem `json:"items"`
	Total  int                 `json:"total"`
	Offset int                 `json:"offset"`
	Limit  int                 `json:"limit"`
	Counts struct {
		Total         int            `json:"total"`
		Photos        int            `json:"photos"`
		Videos        int            `json:"videos"`
		Pixel7        int            `json:"pixel7"`
		DropboxPhotos int            `json:"dropbox_photos"`
		DropboxVideos int            `json:"dropbox_videos"`
		Years         []int          `json:"years"`
		Dashas        map[string]int `json:"dashas"`
	} `json:"counts"`
}

func (h *Handler) convertIndexEntryToVisual(entry db.IndexEntry) UnifiedVisualItem {
	ext := strings.TrimPrefix(strings.ToLower(entry.Extension), ".")
	isPixel7 := strings.HasPrefix(entry.Path, "pixel7/") || entry.Metadata["album"] == "pixel7"
	if !isPixel7 {
		for _, t := range entry.Tags {
			if strings.EqualFold(t, "phone") {
				isPixel7 = true
				break
			}
		}
	}

	src := "vault"
	if isPixel7 {
		src = "pixel7"
	} else if entry.Category == "photos" {
		src = "dropbox_photos"
	} else if entry.Category == "video" {
		src = "dropbox_videos"
	}

	title := entry.FileName
	if t, ok := entry.Metadata["title"].(string); ok && t != "" {
		title = t
	} else {
		base := strings.TrimSuffix(entry.FileName, filepath.Ext(entry.FileName))
		title = strings.ReplaceAll(base, "_", " ")
		title = strings.ReplaceAll(title, "-", " ")
	}

	var durationSec float64
	var durationStr string
	if dur, ok := entry.Metadata["duration_sec"].(float64); ok && dur > 0 {
		durationSec = dur
		m := int(dur) / 60
		s := int(dur) % 60
		durationStr = fmt.Sprintf("%02d:%02d", m, s)
	}

	var res string
	if r, ok := entry.Metadata["resolution"].(string); ok {
		res = r
	}

	var model string
	if m, ok := entry.Metadata["device_model"].(string); ok {
		model = m
	} else if isPixel7 {
		model = "Pixel 7"
	}

	var loc string
	if l, ok := entry.Metadata["location"].(string); ok {
		loc = l
	}

	var album string
	if a, ok := entry.Metadata["album"].(string); ok {
		album = a
	} else {
		album = filepath.Base(filepath.Dir(entry.Path))
	}

	var photoDate string
	var year int
	if pd, ok := entry.Metadata["photo_date"].(string); ok && pd != "" {
		photoDate = pd
		if t, err := time.Parse("2006-01-02", pd); err == nil {
			year = t.Year()
		}
	}
	if year == 0 {
		if y, ok := entry.Metadata["year"].(float64); ok && y > 0 {
			year = int(y)
		} else if y, ok := entry.Metadata["year"].(int); ok && y > 0 {
			year = y
		} else {
			year = entry.ModTime.Year()
		}
	}
	if photoDate == "" {
		photoDate = entry.ModTime.Format("2006-01-02")
	}

	var maha, antar, metal, hora, hermetic string
	if m, ok := entry.Metadata["dasha_mahadasha"].(string); ok {
		maha = m
	}
	if a, ok := entry.Metadata["dasha_antardasha"].(string); ok {
		antar = a
	}
	if sm, ok := entry.Metadata["sacred_metal"].(string); ok {
		metal = sm
	}
	if h, ok := entry.Metadata["hora"].(string); ok {
		hora = h
	}
	if ax, ok := entry.Metadata["hermetic_axiom"].(string); ok {
		hermetic = ax
	}

	streamURL := fmt.Sprintf("/api/v1/media/stream?path=%s", url.QueryEscape(entry.Path))
	thumbURL := streamURL
	if entry.Category == "video" {
		thumbURL = fmt.Sprintf("/api/v1/media/thumbnail?path=%s", url.QueryEscape(entry.Path))
	}

	return UnifiedVisualItem{
		ID:              entry.ID,
		Title:           title,
		FileName:        entry.FileName,
		Source:          src,
		Category:        entry.Category,
		Format:          ext,
		Path:            entry.Path,
		FullPath:        entry.FullPath,
		SizeBytes:       entry.SizeBytes,
		ModTime:         entry.ModTime.Format(time.RFC3339),
		MediaDate:       photoDate,
		Year:            year,
		DurationSec:     durationSec,
		DurationStr:     durationStr,
		Resolution:      res,
		ThumbnailURL:    thumbURL,
		StreamURL:       streamURL,
		Tags:            entry.Tags,
		DashaMahadasha:  maha,
		DashaAntardasha: antar,
		SacredMetal:     metal,
		Hora:            hora,
		HermeticAxiom:   hermetic,
		DeviceModel:     model,
		Location:        loc,
		Album:           album,
		Metadata:        entry.Metadata,
	}
}

func (h *Handler) getVisualCatalogItems() []UnifiedVisualItem {
	h.visualMu.RLock()
	if h.visualCache != nil && time.Since(h.visualCache.cachedAt) < 5*time.Minute {
		cached := h.visualCache.items
		h.visualMu.RUnlock()
		return cached
	}
	h.visualMu.RUnlock()

	h.visualMu.Lock()
	defer h.visualMu.Unlock()

	// Double check cache
	if h.visualCache != nil && time.Since(h.visualCache.cachedAt) < 5*time.Minute {
		return h.visualCache.items
	}

	var allEntries []db.IndexEntry
	if h.store != nil {
		photos, _, err := h.store.SearchIndex(db.IndexFilter{
			Category: "photos",
			Limit:    100000,
		})
		if err == nil {
			allEntries = append(allEntries, photos...)
		}

		videos, _, err := h.store.SearchIndex(db.IndexFilter{
			Category: "video",
			Limit:    100000,
		})
		if err == nil {
			allEntries = append(allEntries, videos...)
		}
	}

	items := make([]UnifiedVisualItem, 0, len(allEntries))
	for _, entry := range allEntries {
		items = append(items, h.convertIndexEntryToVisual(entry))
	}

	// Sort newest first by ModTime / MediaDate
	sort.Slice(items, func(i, j int) bool {
		dateI := items[i].MediaDate
		if dateI == "" {
			dateI = items[i].ModTime
		}
		dateJ := items[j].MediaDate
		if dateJ == "" {
			dateJ = items[j].ModTime
		}
		return dateI > dateJ
	})

	h.visualCache = &visualCacheEntry{
		items:    items,
		cachedAt: time.Now(),
	}
	return items
}

// UnifiedVisualCatalogHandler serves federated visual media from Pixel 7 and Dropbox archives.
func (h *Handler) UnifiedVisualCatalogHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	filterSource := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("source")))
	filterCategory := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("category")))
	filterDasha := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("dasha")))
	sortBy := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("sort")))

	var filterYear int
	if yStr := r.URL.Query().Get("year"); yStr != "" {
		if y, err := strconv.Atoi(yStr); err == nil && y > 0 {
			filterYear = y
		}
	}

	limit := 60
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
			if limit > 500 {
				limit = 500
			}
		}
	}
	offset := 0
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	allItems := h.getVisualCatalogItems()

	var resp UnifiedVisualCatalogResponse
	resp.Counts.Dashas = make(map[string]int)
	yearSet := make(map[int]bool)

	// Collect global counts and facets
	for _, item := range allItems {
		resp.Counts.Total++
		if item.Category == "photos" {
			resp.Counts.Photos++
		} else if item.Category == "video" {
			resp.Counts.Videos++
		}

		if item.Source == "pixel7" {
			resp.Counts.Pixel7++
		} else if item.Source == "dropbox_photos" {
			resp.Counts.DropboxPhotos++
		} else if item.Source == "dropbox_videos" {
			resp.Counts.DropboxVideos++
		}

		if item.Year > 0 {
			yearSet[item.Year] = true
		}
		if item.DashaMahadasha != "" {
			resp.Counts.Dashas[item.DashaMahadasha]++
		}
	}

	for y := range yearSet {
		resp.Counts.Years = append(resp.Counts.Years, y)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(resp.Counts.Years)))

	// Filter
	var filtered []UnifiedVisualItem
	for _, item := range allItems {
		// Category filter
		if filterCategory != "" && filterCategory != "all" {
			if !strings.EqualFold(item.Category, filterCategory) {
				continue
			}
		}

		// Source filter
		if filterSource != "" && filterSource != "all" {
			if filterSource == "pixel7" && item.Source != "pixel7" {
				continue
			}
			if (filterSource == "dropbox_photos" || filterSource == "photos") && item.Source != "dropbox_photos" {
				continue
			}
			if (filterSource == "dropbox_videos" || filterSource == "video" || filterSource == "videos") && item.Source != "dropbox_videos" {
				continue
			}
			if filterSource == "vault" && item.Source == "pixel7" {
				continue
			}
		}

		// Year filter
		if filterYear > 0 && item.Year != filterYear {
			continue
		}

		// Dasha filter
		if filterDasha != "" {
			if !strings.EqualFold(item.DashaMahadasha, filterDasha) &&
				!strings.EqualFold(item.DashaAntardasha, filterDasha) {
				continue
			}
		}

		// Text Query filter
		if q != "" {
			match := strings.Contains(strings.ToLower(item.FileName), q) ||
				strings.Contains(strings.ToLower(item.Title), q) ||
				strings.Contains(strings.ToLower(item.Path), q) ||
				strings.Contains(strings.ToLower(item.Album), q) ||
				strings.Contains(strings.ToLower(item.DashaMahadasha), q) ||
				strings.Contains(strings.ToLower(item.SacredMetal), q) ||
				strings.Contains(strings.ToLower(item.DeviceModel), q)

			if !match {
				for _, tag := range item.Tags {
					if strings.Contains(strings.ToLower(tag), q) {
						match = true
						break
					}
				}
			}
			if !match {
				continue
			}
		}

		filtered = append(filtered, item)
	}

	// Apply Sorting if non-default
	switch sortBy {
	case "date_asc":
		sort.Slice(filtered, func(i, j int) bool {
			dI := filtered[i].MediaDate
			if dI == "" {
				dI = filtered[i].ModTime
			}
			dJ := filtered[j].MediaDate
			if dJ == "" {
				dJ = filtered[j].ModTime
			}
			return dI < dJ
		})
	case "size_desc":
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].SizeBytes > filtered[j].SizeBytes
		})
	case "name_asc":
		sort.Slice(filtered, func(i, j int) bool {
			return strings.ToLower(filtered[i].FileName) < strings.ToLower(filtered[j].FileName)
		})
	}

	resp.Total = len(filtered)
	resp.Offset = offset
	resp.Limit = limit

	if offset < len(filtered) {
		end := offset + limit
		if end > len(filtered) {
			end = len(filtered)
		}
		resp.Items = filtered[offset:end]
	} else {
		resp.Items = []UnifiedVisualItem{}
	}

	_ = json.NewEncoder(w).Encode(resp)
}
