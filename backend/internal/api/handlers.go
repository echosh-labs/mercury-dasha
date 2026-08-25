package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/config"
	"github.com/echosh-labs/mercury-dasha/internal/db"
)

var startTime = time.Now()

type Handler struct {
	cfg   *config.Config
	store db.StorageEngine
}

func NewHandler(cfg *config.Config, store db.StorageEngine) *Handler {
	return &Handler{
		cfg:   cfg,
		store: store,
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

// DashaOverviewHandler returns the planetary mahadasha cycles.
func (h *Handler) DashaOverviewHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"mahadasha_lord":    "Mercury (17-Year Cycle)",
		"total_years":       120,
		"root_frequency_hz": 141.27,
		"planets": []map[string]any{
			{"name": "Mercury (Budha)", "sanskrit_name": "Budha", "duration_years": 17, "frequency_hz": 141.27, "element": "Earth/Air", "chakra": "Vishuddha (Throat)", "color_hex": "#10b981"},
			{"name": "Ketu", "sanskrit_name": "Ketu", "duration_years": 7, "frequency_hz": 211.44, "element": "Fire/Void", "chakra": "Sahasrara (Crown)", "color_hex": "#71717a"},
			{"name": "Venus (Shukra)", "sanskrit_name": "Shukra", "duration_years": 20, "frequency_hz": 221.23, "element": "Water", "chakra": "Swadhisthana (Sacral)", "color_hex": "#ec4899"},
			{"name": "Sun (Surya)", "sanskrit_name": "Surya", "duration_years": 6, "frequency_hz": 126.22, "element": "Fire", "chakra": "Manipura (Solar)", "color_hex": "#f59e0b"},
			{"name": "Moon (Chandra)", "sanskrit_name": "Chandra", "duration_years": 10, "frequency_hz": 210.42, "element": "Water", "chakra": "Ajna (Third Eye)", "color_hex": "#06b6d4"},
			{"name": "Mars (Mangala)", "sanskrit_name": "Mangala", "duration_years": 7, "frequency_hz": 144.72, "element": "Fire", "chakra": "Muladhara (Root)", "color_hex": "#ef4444"},
			{"name": "Rahu", "sanskrit_name": "Rahu", "duration_years": 18, "frequency_hz": 187.61, "element": "Smoke/Shadow", "chakra": "Astral Bridge", "color_hex": "#6366f1"},
			{"name": "Jupiter (Guru)", "sanskrit_name": "Guru", "duration_years": 16, "frequency_hz": 183.58, "element": "Ether", "chakra": "Anahata (Heart)", "color_hex": "#eab308"},
			{"name": "Saturn (Shani)", "sanskrit_name": "Shani", "duration_years": 19, "frequency_hz": 147.85, "element": "Air/Cold", "chakra": "Muladhara Base", "color_hex": "#3b82f6"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// DashaNakshatrasHandler returns Mercury-ruled nakshatras.
func (h *Handler) DashaNakshatrasHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"ruling_planet": "Mercury",
		"count":         3,
		"nakshatras": []map[string]any{
			{
				"id":        "ashlesha",
				"name":      "Ashlesha (The Clinging Star)",
				"range":     "16°40' - 30°00' Cancer",
				"frequency": 432,
				"deity":     "Nagas (Serpent Wisdom)",
				"symbol":    "Coiled Snake",
				"quality":   "Sharp, Penetrating Intuition",
			},
			{
				"id":        "jyeshtha",
				"name":      "Jyeshtha (The Eldest)",
				"range":     "16°40' - 30°00' Scorpio",
				"frequency": 528,
				"deity":     "Indra (King of the Gods)",
				"symbol":    "Circular Amulet / Talisman",
				"quality":   "Mastery, Directorial Will",
			},
			{
				"id":        "revati",
				"name":      "Revati (The Wealthy / Journey's End)",
				"range":     "16°40' - 30°00' Pisces",
				"frequency": 639,
				"deity":     "Pushan (Nourisher of Souls)",
				"symbol":    "Fish Swimming in Opposite Directions",
				"quality":   "Cosmic Transcendence & Safe Passage",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// DashaAlchemyHandler returns Hermetic axioms and quicksilver correspondences.
func (h *Handler) DashaAlchemyHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"philosophy": "Hermetic & Alchemical Principles",
		"metal":      "Quicksilver (Liquid Mercury)",
		"axioms": []map[string]string{
			{"title": "The Principle of Mentalism", "text": "THE ALL IS MIND; The Universe is Mental."},
			{"title": "The Principle of Correspondence", "text": "As above, so below; as below, so above."},
			{"title": "The Principle of Vibration", "text": "Nothing rests; everything moves; everything vibrates."},
			{"title": "The Principle of Polarity", "text": "Everything is Dual; everything has poles; opposites are identical in nature."},
			{"title": "The Principle of Rhythm", "text": "Everything flows, out and in; the pendulum-swing manifests in all things."},
			{"title": "The Principle of Cause and Effect", "text": "Every Cause has its Effect; every Effect has its Cause."},
			{"title": "The Principle of Gender", "text": "Gender is in everything; everything has its Masculine and Feminine Principles."},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
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
