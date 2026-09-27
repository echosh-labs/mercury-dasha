package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/chrono"
)

type SystemOverviewResponse struct {
	Service      string              `json:"service"`
	Version      string              `json:"version"`
	GoVersion    string              `json:"go_version"`
	Environment  string              `json:"environment"`
	Timestamp    string              `json:"timestamp"`
	Uptime       string              `json:"uptime"`
	UptimeSec    float64             `json:"uptime_sec"`
	RuntimeStats RuntimeOverview     `json:"runtime_stats"`
	Database     DatabaseOverview    `json:"database"`
	Dropbox      DropboxOverview     `json:"dropbox"`
	ChronoPulse  ChronoOverview      `json:"chrono_pulse"`
	YouTube      YouTubeOverview     `json:"youtube"`
	Architecture SubstrateOverview   `json:"architecture"`
	SelfHealing  SelfHealingOverview `json:"self_healing"`
}

type SelfHealingOverview struct {
	Active             bool   `json:"active"`
	Status             string `json:"status"`
	AutoHealsTriggered uint64 `json:"auto_heals_triggered"`
	PanicsRecovered    uint64 `json:"panics_recovered"`
	LastHealedAt       string `json:"last_healed_at,omitempty"`
	TotalChecks        uint64 `json:"total_checks"`
	RestartPending     bool   `json:"restart_pending"`
}

type YouTubeOverview struct {
	Configured    bool   `json:"configured"`
	Authenticated bool   `json:"authenticated"`
	ChannelName   string `json:"channel_name,omitempty"`
	VideoCount    uint64 `json:"video_count,omitempty"`
}

type RuntimeOverview struct {
	Goroutines   int     `json:"goroutines"`
	AllocMB      float64 `json:"alloc_mb"`
	TotalAllocMB float64 `json:"total_alloc_mb"`
	SysMB        float64 `json:"sys_mb"`
	NumGC        uint32  `json:"num_gc"`
}

type DatabaseOverview struct {
	Path           string         `json:"path"`
	SizeBytes      int64          `json:"size_bytes"`
	KeyCount       int            `json:"key_count"`
	OpenTime       string         `json:"open_time"`
	AllocatedPages int            `json:"allocated_pages"`
	CategoryCounts map[string]int64 `json:"category_counts"`
}

type DropboxOverview struct {
	LocalPath    string `json:"local_path"`
	LocalExists  bool   `json:"local_exists"`
	TotalIndexed int64  `json:"total_indexed"`
	Configured   bool   `json:"cloud_configured"`
	AccountEmail string `json:"account_email,omitempty"`
	AccountType  string `json:"account_type,omitempty"`
}

type ChronoOverview struct {
	ActiveHora   string `json:"active_hora"`
	SacredMetal  string `json:"sacred_metal"`
	Axiom        string `json:"axiom"`
	ClientCount  int    `json:"client_count"`
}

type SubstrateOverview struct {
	CanonicalRoot string   `json:"canonical_root"`
	HostSubstrate string   `json:"host_substrate"`
	SiblingRepos  []string `json:"sibling_repos"`
}

// SystemOverviewHandler returns full-stack telemetry, DB stats, and architecture mapping.
func (h *Handler) SystemOverviewHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	up := time.Since(startTime)

	dbStats, _ := h.store.GetStats()
	idxCounts, _ := h.store.GetIndexStats()

	totalIdx := int64(0)
	if idxCounts != nil {
		totalIdx = idxCounts["total"]
	}

	hora := chrono.GetPlanetaryHora(time.Now().UTC())

	localPath := h.cfg.DropboxLocalPath
	if localPath == "" {
		localPath = "/home/justin/Dropbox"
	}
	_, localExists := os.Stat(localPath)

	clientCount := 0
	if h.metronome != nil {
		clientCount = h.metronome.ClientCount()
	}

	dbOverview := DatabaseOverview{
		Path:           h.cfg.BoltDBPath,
		CategoryCounts: idxCounts,
	}
	if dbStats != nil {
		dbOverview.SizeBytes = dbStats.SizeBytes
		dbOverview.KeyCount = dbStats.KeyCount
		dbOverview.OpenTime = dbStats.OpenTime.Format(time.RFC3339)
		dbOverview.AllocatedPages = int(dbStats.AllocatedPages)
	}

	ytOverview := YouTubeOverview{}
	if h.youtube != nil {
		ytOverview.Configured = h.youtube.IsConfigured(r.Context())
		ytOverview.Authenticated = h.youtube.IsAuthenticated()
	}

	resp := SystemOverviewResponse{
		Service:     h.cfg.ServiceName,
		Version:     "1.0.0",
		GoVersion:   runtime.Version(),
		Environment: h.cfg.Environment,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Uptime:      up.Round(time.Second).String(),
		UptimeSec:   up.Seconds(),
		RuntimeStats: RuntimeOverview{
			Goroutines:   runtime.NumGoroutine(),
			AllocMB:      float64(m.Alloc) / 1024 / 1024,
			TotalAllocMB: float64(m.TotalAlloc) / 1024 / 1024,
			SysMB:        float64(m.Sys) / 1024 / 1024,
			NumGC:        m.NumGC,
		},
		Database: dbOverview,
		Dropbox: DropboxOverview{
			LocalPath:    localPath,
			LocalExists:  localExists == nil,
			TotalIndexed: totalIdx,
			Configured:   h.dropbox != nil && h.dropbox.IsConfigured(),
		},
		ChronoPulse: ChronoOverview{
			ActiveHora:  hora.Name,
			SacredMetal: hora.SacredMetal,
			Axiom:       hora.HermeticAxiom,
			ClientCount: clientCount,
		},
		YouTube: ytOverview,
		Architecture: SubstrateOverview{
			CanonicalRoot: "/home/justin/code/echosh-labs/mercury-dasha",
			HostSubstrate: "WSL2 Ubuntu (Linux 6.6+)",
			SiblingRepos:  []string{"echosh", "foundations", "axis-mundi", "shaolin", "echosh-labs.com"},
		},
		SelfHealing: func() SelfHealingOverview {
			sh := SelfHealingOverview{
				Active: false,
				Status: "STANDBY",
			}
			if h.selfHealing != nil {
				report := h.selfHealing.EvaluateHealth()
				sh.Active = report.SelfHealingActive
				sh.Status = string(report.OverallStatus)
				sh.AutoHealsTriggered = report.Stats.AutoHealsTriggered
				sh.PanicsRecovered = report.Stats.PanicsRecovered
				sh.TotalChecks = report.Stats.TotalChecks
				sh.RestartPending = report.Stats.RestartPending
				if report.Stats.LastHealedAt != nil {
					sh.LastHealedAt = report.Stats.LastHealedAt.Format(time.RFC3339)
				}
			}
			return sh
		}(),
	}

	_ = json.NewEncoder(w).Encode(resp)
}

type ConversationEntry struct {
	ID         string    `json:"id"`
	Hemisphere string    `json:"hemisphere"` // "left" (Windows IDE) or "right" (WSL CLI)
	Path       string    `json:"path"`
	Date       time.Time `json:"date"`
	SizeBytes  int64     `json:"size_bytes"`
	Artifacts  []string  `json:"artifacts"`
}

type BrainDirectory struct {
	Path       string
	Hemisphere string // "left" or "right"
	Label      string
}

func getBicameralDirectories() []BrainDirectory {
	var list []BrainDirectory
	home, err := os.UserHomeDir()

	// 1. Right Hemisphere (WSL2 AGY CLI)
	if err == nil {
		cliBrain := filepath.Join(home, ".gemini", "antigravity-cli", "brain")
		list = append(list, BrainDirectory{
			Path:       cliBrain,
			Hemisphere: "right",
			Label:      "WSL2 AGY CLI",
		})

		wslBrain := filepath.Join(home, ".gemini", "antigravity", "brain")
		list = append(list, BrainDirectory{
			Path:       wslBrain,
			Hemisphere: "right",
			Label:      "WSL2 Antigravity",
		})
	}

	// 2. Left Hemisphere (Windows Antigravity IDE)
	list = append(list, BrainDirectory{
		Path:       "/mnt/c/Users/justi/.gemini/antigravity/brain",
		Hemisphere: "left",
		Label:      "Windows Antigravity IDE",
	})

	return list
}

func getBrainDirectories() []string {
	var candidates []string
	for _, b := range getBicameralDirectories() {
		candidates = append(candidates, b.Path)
	}
	return candidates
}

type HemisphereInfo struct {
	Name             string `json:"name"`
	Substrate        string `json:"substrate"`
	Path             string `json:"path"`
	Exists           bool   `json:"exists"`
	SessionCount     int    `json:"session_count"`
	InstructionChain string `json:"instruction_chain"`
}

type BicameralOverviewResponse struct {
	LeftHemisphere  HemisphereInfo `json:"left_hemisphere"`
	RightHemisphere HemisphereInfo `json:"right_hemisphere"`
	CorpusCallosum  string         `json:"corpus_callosum"`
	TotalSessions   int            `json:"total_sessions"`
	ActiveSession   string         `json:"active_session,omitempty"`
	Timestamp       string         `json:"timestamp"`
}

// AgentBicameralHandler returns bicameral dual-hemisphere telemetry and status.
func (h *Handler) AgentBicameralHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dirs := getBicameralDirectories()
	leftCount := 0
	rightCount := 0
	leftExists := false
	rightExists := false
	leftPath := ""
	rightPath := ""

	for _, d := range dirs {
		des, err := os.ReadDir(d.Path)
		if err != nil {
			continue
		}
		count := 0
		for _, de := range des {
			if de.IsDir() {
				count++
			}
		}
		if d.Hemisphere == "left" {
			leftExists = true
			leftPath = d.Path
			leftCount += count
		} else {
			rightExists = true
			if rightPath == "" {
				rightPath = d.Path
			}
			rightCount += count
		}
	}

	resp := BicameralOverviewResponse{
		LeftHemisphere: HemisphereInfo{
			Name:             "Left Hemisphere (Windows Antigravity IDE)",
			Substrate:        "Windows 11 Desktop (Host Substrate)",
			Path:             leftPath,
			Exists:           leftExists,
			SessionCount:     leftCount,
			InstructionChain: "User Global Rules (<RULE[user_global]>) + wise-bardeen/GEMINI.md",
		},
		RightHemisphere: HemisphereInfo{
			Name:             "Right Hemisphere (WSL2 Ubuntu AGY CLI)",
			Substrate:        "WSL2 Ubuntu 24.04 (POSIX Ext4 Substrate)",
			Path:             rightPath,
			Exists:           rightExists,
			SessionCount:     rightCount,
			InstructionChain: "Monorepo Root AGENTS.md + mercury-dasha/GEMINI.md",
		},
		CorpusCallosum: "/home/justin/.gemini/bicameral (manifest.json)",
		TotalSessions:  leftCount + rightCount,
		ActiveSession:  "65df33eb-aefb-4206-b758-6072bfa69b3a",
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// AgentConversationsHandler lists recent agent sessions and available markdown artifacts across both hemispheres.
func (h *Handler) AgentConversationsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var entries []ConversationEntry
	seen := make(map[string]bool)

	for _, bdir := range getBicameralDirectories() {
		des, err := os.ReadDir(bdir.Path)
		if err != nil {
			continue
		}

		for _, de := range des {
			if !de.IsDir() || seen[de.Name()] {
				continue
			}

			fullDir := filepath.Join(bdir.Path, de.Name())
			info, err := de.Info()
			if err != nil {
				continue
			}

			// Scan for immediate artifacts in conversation folder
			artifacts := []string{}
			var totalSize int64
			files, err := os.ReadDir(fullDir)
			if err == nil {
				for _, f := range files {
					if f.IsDir() {
						continue
					}
					base := f.Name()
					if finfo, err := f.Info(); err == nil {
						totalSize += finfo.Size()
					}
					if strings.HasSuffix(base, ".md") && !strings.Contains(base, ".metadata.") {
						artifacts = append(artifacts, base)
					}
				}
			}

			seen[de.Name()] = true
			entries = append(entries, ConversationEntry{
				ID:         de.Name(),
				Hemisphere: bdir.Hemisphere,
				Path:       fullDir,
				Date:       info.ModTime(),
				SizeBytes:  totalSize,
				Artifacts:  artifacts,
			})
		}
	}

	// Sort descending by date
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date.After(entries[j].Date)
	})

	_ = json.NewEncoder(w).Encode(map[string]any{
		"total":         len(entries),
		"conversations": entries,
	})
}

// AgentArtifactHandler returns the raw markdown text for a specified conversation artifact.
func (h *Handler) AgentArtifactHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Query().Get("id")
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "walkthrough.md"
	}

	// Prevent directory traversal
	cleanName := filepath.Base(name)
	cleanID := filepath.Base(id)
	if !strings.HasSuffix(cleanName, ".md") {
		http.Error(w, `{"error":"only markdown (.md) artifacts may be viewed"}`, http.StatusBadRequest)
		return
	}

	var content []byte
	var foundPath string
	for _, brainDir := range getBrainDirectories() {
		candidate := filepath.Join(brainDir, cleanID, cleanName)
		data, err := os.ReadFile(candidate)
		if err == nil {
			content = data
			foundPath = candidate
			break
		}
	}

	if content == nil {
		http.Error(w, fmt.Sprintf(`{"error":"artifact %q not found in session %q"}`, cleanName, cleanID), http.StatusNotFound)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":      cleanID,
		"name":    cleanName,
		"path":    foundPath,
		"content": string(content),
	})
}

type RouteDoc struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// APICatalogHandler returns all registered endpoints with metadata for live testing.
func (h *Handler) APICatalogHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	routes := []RouteDoc{
		// Dasha Observatory
		{"POST", "/api/dasha/calculate", "Dasha Observatory", "Calculate 120-year Vimshottari hierarchy from birth ephemeris or nakshatra/pada."},
		{"GET", "/api/dasha/overview", "Dasha Observatory", "Get Vimshottari system overview and 9 planetary resonance frequencies."},
		{"GET", "/api/dasha/nakshatras", "Dasha Observatory", "List all 27 Vedic Nakshatras with deities, symbols, and qualities."},
		{"GET", "/api/dasha/planets", "Dasha Observatory", "List the 9 Vimshottari Grahas with cycle years and chakra centers."},
		{"GET", "/api/dasha/hora", "Dasha Observatory", "Real-time planetary hora ruler, sacred metal, axiom, and story archetype."},
		{"GET", "/api/dasha/alchemy", "Alchemical Laboratory", "The 7 Sacred Metals, 7 Hermetic Axioms, and Magnum Opus stages."},
		{"POST", "/api/v1/profiles/{id}", "Dasha Profiles", "Save calculated natal chart and dasha profile to BoltDB."},
		{"GET", "/api/v1/profiles", "Dasha Profiles", "List all stored dasha profiles in BoltDB."},
		{"GET", "/api/v1/profiles/{id}", "Dasha Profiles", "Retrieve stored dasha profile by ID."},
		{"DELETE", "/api/v1/profiles/{id}", "Dasha Profiles", "Delete stored dasha profile by ID."},

		// Chrono-Pulse Metronome
		{"GET", "/api/v1/stream/pulse", "Chrono-Pulse", "Server-Sent Events (SSE) 2-second metronome broadcasting live hora and telemetry."},

		// BoltDB Storehouse
		{"GET", "/api/v1/meta", "BoltDB Storehouse", "List all document keys in BoltDB meta bucket."},
		{"GET", "/api/v1/meta/{key}", "BoltDB Storehouse", "Get document JSON by key."},
		{"POST", "/api/v1/meta/{key}", "BoltDB Storehouse", "Put or update document JSON by key."},
		{"DELETE", "/api/v1/meta/{key}", "BoltDB Storehouse", "Delete document by key."},
		{"GET", "/api/v1/backup", "BoltDB Storehouse", "Stream raw BoltDB snapshot as downloadable binary."},

		// Dropbox Sovereign Storehouse
		{"GET", "/api/v1/index/status", "Dropbox Storehouse", "Crawl status and file counts across books, code, audio, video, text."},
		{"GET", "/api/v1/index/search", "Dropbox Storehouse", "Indexed search across 85,119 files with category and extension filters."},
		{"GET", "/api/v1/index/content", "Dropbox Storehouse", "HTTP 206 Partial Content / Range media streaming and document retrieval."},
		{"POST", "/api/v1/index/scan", "Dropbox Storehouse", "Trigger asynchronous crawl across local Dropbox folders."},
		{"GET", "/api/v1/dropbox/status", "Dropbox Cloud", "OAuth2 account status, quota, and non-expiring token state."},
		{"POST", "/api/v1/dropbox/backup", "Dropbox Cloud", "Upload BoltDB snapshot directly to /MercuryDasha/backups via OAuth."},
		{"GET", "/api/v1/dropbox/files", "Dropbox Cloud", "List backup archives stored in Dropbox cloud backup path."},
		{"GET", "/api/v1/dropbox/link", "Dropbox Cloud", "Generate temporary 4-hour direct streaming share link."},

		// Syncthing Media Ingestion Subsystem & Vault
		{"GET", "/api/v1/syncthing/status", "Syncthing Vault", "Real-time daemon health, paired device connections, and media folder sync metrics."},
		{"GET", "/api/v1/syncthing/feed", "Syncthing Vault", "Chronological feed of recently synced photos and videos with astrological Dasha grounding."},
		{"POST", "/api/v1/syncthing/rescan", "Syncthing Vault", "Command Syncthing daemon to execute immediate folder re-scan."},
		{"GET", "/api/v1/media/stream", "Syncthing Vault", "Stream high-resolution media with byte-range HTTP 206 Partial Content seeking."},
		{"GET", "/api/v1/media/thumbnail", "Syncthing Vault", "Serve generated video/photo poster thumbnail images."},

		// Axis Mundi Workspace Ingestion Hub
		{"GET", "/api/v1/axis-mundi/status", "Axis Mundi Hub", "Health status, connection mode, and statistics for Keep, Gmail, Docs, and Sheets monitoring."},
		{"GET", "/api/v1/axis-mundi/feed", "Axis Mundi Hub", "Aggregated feed of observed Google Workspace items (Keep notes, Gmail threads, Docs, Sheets)."},
		{"POST", "/api/v1/axis-mundi/sync", "Axis Mundi Hub", "Trigger immediate on-demand poll against Axis Mundi service."},
		{"POST", "/api/v1/axis-mundi/events", "Axis Mundi Hub", "Ingest pushed workspace event and broadcast pulse alert."},

		// YouTube Sovereign Video Uploader & Studio
		{"GET", "/api/v1/youtube/status", "YouTube Studio", "OAuth2 account status, verified channel metadata, and estimated quota units."},
		{"GET", "/api/v1/youtube/auth/url", "YouTube Studio", "Generate Google OAuth 2.0 authorization URL with offline consent."},
		{"GET", "/api/v1/youtube/auth/callback", "YouTube Studio", "OAuth2 redirect receiver and token persistence handler."},
		{"POST", "/api/v1/youtube/auth/disconnect", "YouTube Studio", "Revoke OAuth tokens and purge YouTube credentials from BoltDB."},
		{"POST", "/api/v1/youtube/upload", "YouTube Studio", "Enqueue video upload from local POSIX path or multipart form."},
		{"GET", "/api/v1/youtube/jobs", "YouTube Studio", "List active and historical video upload jobs."},
		{"GET", "/api/v1/youtube/jobs/{id}", "YouTube Studio", "Get real-time upload progress and video URL."},
		{"POST", "/api/v1/youtube/jobs/{id}/cancel", "YouTube Studio", "Cancel an in-progress video upload."},
		{"GET", "/api/v1/youtube/videos", "YouTube Studio", "List authenticated channel's uploaded videos."},

		// System & Agent Mission Control
		{"GET", "/healthz", "Mission Control", "Service health check, Go version, and database open stats."},
		{"GET", "/api/telemetry", "Mission Control", "Go runtime memory, goroutines, and GC statistics."},
		{"GET", "/api/v1/system/overview", "Mission Control", "Full-stack system diagnostics, DB telemetry, and substrate mapping."},
		{"GET", "/api/v1/system/routes", "Mission Control", "Interactive API catalog of all registered endpoints."},
		{"GET", "/api/v1/agent/conversations", "Mission Control", "Enumerate Antigravity agent brain sessions and report artifacts."},
		{"GET", "/api/v1/agent/artifact", "Mission Control", "Read and stream markdown artifacts (walkthrough.md, implementation_plan.md)."},
		{"GET", "/api/v1/agent/bicameral", "Mission Control", "Bicameral dual-hemisphere telemetry (Windows IDE Left & WSL CLI Right)."},

		// Autonomous Self-Healing & Remote Service Restart
		{"GET", "/api/v1/system/health", "Self-Healing", "Multi-factor health evaluation across BoltDB, memory, and goroutines."},
		{"POST", "/api/v1/system/self-healing/check", "Self-Healing", "Run on-demand diagnostic probe and trigger automated remediation."},
		{"POST", "/api/v1/system/self-healing/remediate", "Self-Healing", "Execute targeted self-healing action (memory_gc, free_os_mem, db_verify)."},
		{"GET", "/api/v1/system/self-healing/incidents", "Self-Healing", "Retrieve real-time and persistent journal of self-healing incidents."},
		{"POST", "/api/v1/system/restart", "Self-Healing", "Issue authorized remote restart of the Mercury Dasha service."},
		{"GET", "/api/v1/system/restart/history", "Self-Healing", "Audit trail of remote and automated service restarts."},
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"total":  len(routes),
		"routes": routes,
	})
}
