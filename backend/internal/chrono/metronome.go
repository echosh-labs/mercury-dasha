package chrono

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
	"github.com/echosh-labs/mercury-dasha/internal/db"
)

// ChaldeanOrder defines the descending planetary speed sequence for Hora calculations:
// Saturn -> Jupiter -> Mars -> Sun -> Venus -> Mercury -> Moon
var ChaldeanOrder = []dasha.PlanetID{
	dasha.PlanetSaturn,
	dasha.PlanetJupiter,
	dasha.PlanetMars,
	dasha.PlanetSun,
	dasha.PlanetVenus,
	dasha.PlanetMercury,
	dasha.PlanetMoon,
}

// DayRulers maps time.Weekday to the first Hora planetary ruler of that day at 06:00.
var DayRulers = map[time.Weekday]dasha.PlanetID{
	time.Sunday:    dasha.PlanetSun,
	time.Monday:    dasha.PlanetMoon,
	time.Tuesday:   dasha.PlanetMars,
	time.Wednesday: dasha.PlanetMercury,
	time.Thursday:  dasha.PlanetJupiter,
	time.Friday:    dasha.PlanetVenus,
	time.Saturday:  dasha.PlanetSaturn,
}

// HoraInfo captures the active planetary hour ruler, timing progress, and correspondences.
type HoraInfo struct {
	PlanetID               dasha.PlanetID `json:"planet_id"`
	Name                   string         `json:"name"`
	SanskritName           string         `json:"sanskrit_name"`
	Element                string         `json:"element"`
	ChakraCenter           string         `json:"chakra_center"`
	ColorHex               string         `json:"color_hex"`
	HermeticAxiom          string         `json:"hermetic_axiom"`
	SacredMetal            string         `json:"sacred_metal"`
	MetalSymbol            string         `json:"metal_symbol"`
	MagnumOpusStage        string         `json:"magnum_opus_stage"`
	StoryArchetype         string         `json:"story_archetype"`
	HourOfDay              int            `json:"hour_of_day"`
	DayLord                string         `json:"day_lord"`
	Diurnal                bool           `json:"diurnal"`
	Phase                  string         `json:"phase,omitempty"`
	StartTime              time.Time      `json:"start_time"`
	EndTime                time.Time      `json:"end_time"`
	ElapsedMinutes         float64        `json:"elapsed_minutes"`
	RemainingMinutes       float64        `json:"remaining_minutes"`
	ProgressPercent        float64        `json:"progress_percent"`
	BriefApplication       string         `json:"brief_application"`
	SuitableActivities     []string       `json:"suitable_activities,omitempty"`
	UnsuitableActivities   []string       `json:"unsuitable_activities,omitempty"`
	NextHourPlanet         string         `json:"next_hour_planet,omitempty"`
	ActiveSubHora          *SubHora       `json:"active_sub_hora,omitempty"`
	CurrentLagnaDegree     float64        `json:"current_lagna_degree"`
	LocalApparentSolarTime string         `json:"local_apparent_solar_time"`
	EquationOfTimeMinutes  float64        `json:"equation_of_time_minutes"`
}

// GetPlanetaryHora calculates the active Chaldean planetary hour at time t using standard celestial reference.
func GetPlanetaryHora(t time.Time) HoraInfo {
	return GetPlanetaryHoraWithCoords(t, SolarCoordinates{})
}

// GetPlanetaryHoraWithCoords calculates the active Chaldean planetary hour at time t with geographic coordinates.
func GetPlanetaryHoraWithCoords(t time.Time, coords SolarCoordinates) HoraInfo {
	sched := CalculateDaySchedule(t, coords, nil)
	act := sched.ActiveHour

	return HoraInfo{
		PlanetID:               act.PlanetID,
		Name:                   act.PlanetName,
		SanskritName:           act.SanskritName,
		Element:                act.Element,
		ChakraCenter:           act.ChakraCenter,
		ColorHex:               act.ColorHex,
		HermeticAxiom:          act.HermeticAxiom,
		SacredMetal:            act.SacredMetal,
		MetalSymbol:            act.MetalSymbol,
		MagnumOpusStage:        act.MagnumOpusStage,
		StoryArchetype:         dasha.PlanetsByID[act.PlanetID].StoryArchetype,
		HourOfDay:              act.Index,
		DayLord:                sched.DayLord,
		Diurnal:                act.Diurnal,
		Phase:                  act.Phase,
		StartTime:              act.StartTime,
		EndTime:                act.EndTime,
		ElapsedMinutes:         sched.ElapsedMinutes,
		RemainingMinutes:       sched.RemainingMinutes,
		ProgressPercent:        sched.ProgressPercent,
		BriefApplication:       act.BriefApplication,
		SuitableActivities:     act.SuitableActivities,
		UnsuitableActivities:   act.UnsuitableActivities,
		NextHourPlanet:         sched.NextHourPlanet,
		ActiveSubHora:          act.ActiveSubHora,
		CurrentLagnaDegree:     act.CurrentLagnaDegree,
		LocalApparentSolarTime: act.LocalApparentSolarTime,
		EquationOfTimeMinutes:  act.EquationOfTimeMinutes,
	}
}

// PulseMessage represents the periodic broadcast heartbeat sent via SSE.
type PulseMessage struct {
	Timestamp     time.Time                     `json:"timestamp"`
	Uptime        string                        `json:"uptime"`
	Hora          HoraInfo                      `json:"hora"`
	Alchemical    dasha.LiveAlchemicalAlignment `json:"alchemical"`
	NarrativeSeed string                        `json:"narrative_seed"`
	Telemetry     Telemetry                     `json:"telemetry"`
	DB            *db.DBStats                   `json:"db,omitempty"`
}

type Telemetry struct {
	AllocMB      float64 `json:"alloc_mb"`
	SysMB        float64 `json:"sys_mb"`
	Goroutines   int     `json:"goroutines"`
	NumGC        uint32  `json:"num_gc"`
}

// Metronome orchestrates background heartbeats and Server-Sent Events subscribers.
type Metronome struct {
	store           db.StorageEngine
	startTime       time.Time
	clientsMu       sync.RWMutex
	clients         map[chan []byte]SolarCoordinates
	stopChan        chan struct{}
	interval        time.Duration
	coordsResolver  func() SolarCoordinates
	lastDBStats     *db.DBStats
	lastDBStatsTime time.Time
}

// NewMetronome initializes a new Metronome heartbeat engine.
func NewMetronome(store db.StorageEngine, interval time.Duration) *Metronome {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	return &Metronome{
		store:     store,
		startTime: time.Now(),
		clients:   make(map[chan []byte]SolarCoordinates),
		stopChan:  make(chan struct{}),
		interval:  interval,
	}
}

// SetCoordsResolver registers a dynamic function to provide active geographic coordinates.
func (m *Metronome) SetCoordsResolver(fn func() SolarCoordinates) {
	m.coordsResolver = fn
}

func (m *Metronome) getCoords() SolarCoordinates {
	if m.coordsResolver != nil {
		coords := m.coordsResolver()
		if coords.Latitude != 0 || coords.Longitude != 0 {
			return coords
		}
	}
	return DefaultCoordinates
}

// Start launches the background broadcast ticker.
func (m *Metronome) Start() {
	ticker := time.NewTicker(m.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				m.broadcastPulse()
			case <-m.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop shuts down the metronome and disconnects all clients.
func (m *Metronome) Stop() {
	select {
	case <-m.stopChan:
		return
	default:
		close(m.stopChan)
	}
	m.clientsMu.Lock()
	m.clients = make(map[chan []byte]SolarCoordinates)
	m.clientsMu.Unlock()
}

// ClientCount returns the number of currently connected SSE subscribers.
func (m *Metronome) ClientCount() int {
	if m == nil {
		return 0
	}
	m.clientsMu.RLock()
	defer m.clientsMu.RUnlock()
	return len(m.clients)
}

func (m *Metronome) broadcastPulse() {
	select {
	case <-m.stopChan:
		return
	default:
	}

	m.clientsMu.RLock()
	if len(m.clients) == 0 {
		m.clientsMu.RUnlock()
		return
	}

	// Group clients by geographic coordinates to maximize data efficiency and avoid duplicate calculations
	type targetGroup struct {
		coords   SolarCoordinates
		channels []chan []byte
	}
	groups := make(map[string]*targetGroup)

	for ch, coords := range m.clients {
		key := fmt.Sprintf("%.4f_%.4f", coords.Latitude, coords.Longitude)
		if g, exists := groups[key]; exists {
			g.channels = append(g.channels, ch)
		} else {
			groups[key] = &targetGroup{
				coords:   coords,
				channels: []chan []byte{ch},
			}
		}
	}
	m.clientsMu.RUnlock()

	for _, g := range groups {
		payload := m.generatePulse(g.coords)
		data, err := json.Marshal(payload)
		if err != nil {
			continue
		}

		sseFormatted := []byte(fmt.Sprintf("event: pulse\ndata: %s\n\n", string(data)))
		for _, ch := range g.channels {
			select {
			case <-m.stopChan:
				return
			case ch <- sseFormatted:
			default:
				// Non-blocking drop if client is lagging
			}
		}
	}
}

// BroadcastEvent broadcasts an arbitrary typed SSE event to all connected clients.
func (m *Metronome) BroadcastEvent(eventType string, data any) {
	bytesData, err := json.Marshal(data)
	if err != nil {
		return
	}
	msg := []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(bytesData)))

	m.clientsMu.RLock()
	defer m.clientsMu.RUnlock()

	for ch := range m.clients {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (m *Metronome) generatePulse(coords SolarCoordinates) PulseMessage {
	now := time.Now().UTC()
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	uptime := time.Since(m.startTime).Round(time.Second).String()

	// Data-efficiency standard: throttle BoltDB full B+Tree traversals
	// Cache DB stats for 30s rather than running read transactions on every 2-second pulse
	var dbStats *db.DBStats
	if m.store != nil {
		if time.Since(m.lastDBStatsTime) >= 30*time.Second || m.lastDBStats == nil {
			m.lastDBStats, _ = m.store.GetStats()
			m.lastDBStatsTime = time.Now()
		}
		dbStats = m.lastDBStats
	}

	hora := GetPlanetaryHoraWithCoords(now, coords)
	alchem := dasha.ResolveAlchemicalAlignment(hora.PlanetID)
	narrativeSeed := fmt.Sprintf(
		"Under the planetary hour of %s (%s, governing %s), the alchemical vessel undergoes %s. %s",
		hora.Name, hora.SacredMetal, hora.Element, alchem.MagnumOpusStage.Name, hora.HermeticAxiom,
	)

	return PulseMessage{
		Timestamp:     now,
		Uptime:        uptime,
		Hora:          hora,
		Alchemical:    alchem,
		NarrativeSeed: narrativeSeed,
		Telemetry: Telemetry{
			AllocMB:    float64(mem.Alloc) / (1024 * 1024),
			SysMB:      float64(mem.Sys) / (1024 * 1024),
			Goroutines: runtime.NumGoroutine(),
			NumGC:      mem.NumGC,
		},
		DB: dbStats,
	}
}

// SSEHandler serves the text/event-stream HTTP endpoint with location-aware real-time streaming.
func (m *Metronome) SSEHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Disable write deadline for persistent SSE streaming
	rc := http.NewResponseController(w)
	_ = rc.SetWriteDeadline(time.Time{})

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Determine subscriber coordinates: URL query param overrides > dynamic resolver > DefaultCoordinates
	clientCoords := m.getCoords()
	q := r.URL.Query()
	latStr := q.Get("lat")
	lonStr := q.Get("lon")
	if latStr != "" && lonStr != "" {
		var lat, lon float64
		if _, err := fmt.Sscanf(latStr, "%f", &lat); err == nil {
			if _, err := fmt.Sscanf(lonStr, "%f", &lon); err == nil {
				clientCoords = SolarCoordinates{Latitude: lat, Longitude: lon}
			}
		}
	}

	clientChan := make(chan []byte, 16)

	m.clientsMu.Lock()
	m.clients[clientChan] = clientCoords
	m.clientsMu.Unlock()

	defer func() {
		m.clientsMu.Lock()
		delete(m.clients, clientChan)
		m.clientsMu.Unlock()
	}()

	// Send initial pulse immediately on connect with authoritative location-accurate Hora
	initialPayload := m.generatePulse(clientCoords)
	initData, _ := json.Marshal(initialPayload)
	_, _ = fmt.Fprintf(w, "event: pulse\ndata: %s\n\n", string(initData))
	flusher.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case <-m.stopChan:
			return
		case <-heartbeat.C:
			if _, err := fmt.Fprintf(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case msg, ok := <-clientChan:
			if !ok {
				return
			}
			if _, err := w.Write(msg); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
