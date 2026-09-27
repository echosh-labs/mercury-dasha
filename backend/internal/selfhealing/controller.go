package selfhealing

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/config"
	"github.com/echosh-labs/mercury-dasha/internal/db"
	"github.com/echosh-labs/mercury-dasha/internal/portutil"
)

// Controller manages autonomous self-healing, periodic watchdog evaluation,
// memory reclamation, failure mitigation, and panic logging.
type Controller struct {
	mu                     sync.RWMutex
	store                  db.StorageEngine
	cfg                    *config.Config
	restartMgr             *RestartManager
	incidents              []Incident
	stats                  EngineStats
	watchdogTicker         *time.Ticker
	stopChan               chan struct{}
	running                atomic.Bool
	memoryWarningMB        float64
	memoryCriticalMB       float64
	goroutineWarning       int
	goroutineCritical      int
	consecutiveFailures    atomic.Uint64
	maxConsecutiveFailures uint64
	startTime              time.Time
	panicTimestamps        []time.Time
	panicMu                sync.Mutex
}

// NewController initializes the self-healing subsystem.
func NewController(store db.StorageEngine, cfg *config.Config, restartMgr *RestartManager) *Controller {
	memCritical := 1024.0
	if cfg.MemoryLimitMB > 0 {
		memCritical = cfg.MemoryLimitMB
	}
	memWarning := memCritical * 0.75

	c := &Controller{
		store:                  store,
		cfg:                    cfg,
		restartMgr:             restartMgr,
		incidents:              make([]Incident, 0),
		stopChan:               make(chan struct{}),
		memoryWarningMB:        memWarning,
		memoryCriticalMB:       memCritical,
		goroutineWarning:       500,
		goroutineCritical:      2000,
		maxConsecutiveFailures: 3,
		startTime:              time.Now(),
		panicTimestamps:        make([]time.Time, 0),
	}

	c.loadIncidents()
	return c
}

// Start initiates the background watchdog monitoring loop.
func (c *Controller) Start() {
	if c.running.Swap(true) {
		return
	}

	interval := 15 * time.Second
	if c.cfg.WatchdogIntervalSec > 0 {
		interval = time.Duration(c.cfg.WatchdogIntervalSec) * time.Second
	}

	c.watchdogTicker = time.NewTicker(interval)
	go c.watchdogLoop()
	log.Printf("🛡️ [Self-Healing Engine] Autonomous Watchdog active (interval: %v, memory limit: %.0f MB)", interval, c.memoryCriticalMB)
}

// Stop halts the watchdog monitoring routine.
func (c *Controller) Stop() {
	if !c.running.Swap(false) {
		return
	}
	if c.watchdogTicker != nil {
		c.watchdogTicker.Stop()
	}
	close(c.stopChan)
	log.Printf("[Self-Healing Engine] Watchdog stopped gracefully.")
}

func (c *Controller) watchdogLoop() {
	for {
		select {
		case <-c.stopChan:
			return
		case <-c.watchdogTicker.C:
			c.CheckAndHeal()
		}
	}
}

// CheckAndHeal evaluates system components and executes autonomous remedies if needed.
func (c *Controller) CheckAndHeal() SystemHealthReport {
	report := c.EvaluateHealth()

	c.mu.Lock()
	c.stats.TotalChecks++
	now := time.Now().UTC()
	c.stats.LastCheckTime = &now
	if c.restartMgr != nil {
		c.stats.RestartPending = c.restartMgr.IsRestarting()
	}
	c.mu.Unlock()

	// Assess need for automated healing
	if report.OverallStatus == StatusCritical || report.OverallStatus == StatusDegraded {
		fails := c.consecutiveFailures.Add(1)
		c.mu.Lock()
		c.stats.ConsecutiveFailures = fails
		c.mu.Unlock()

		// 1. Memory remediation
		if memComp, ok := report.Components["memory"]; ok && memComp.Status != StatusHealthy {
			freed, err := c.RemediateMemory()
			if err == nil {
				c.mu.Lock()
				c.stats.AutoHealsTriggered++
				c.stats.LastHealedAt = &now
				c.stats.LastHealAction = fmt.Sprintf("Forced GC freed %.2f MB", freed)
				c.mu.Unlock()
			}
		}

		// 2. Database verification / retry
		if dbComp, ok := report.Components["database"]; ok && dbComp.Status == StatusCritical {
			c.RecordIncident(Incident{
				ID:          fmt.Sprintf("inc-%d", time.Now().UnixNano()),
				Timestamp:   time.Now().UTC(),
				Severity:    "critical",
				Component:   "database",
				Message:     "BoltDB storehouse unresponsive during periodic watchdog check",
				ActionTaken: "db_verify",
				Resolved:    false,
			})
		}

		// 3. Escalation: If consecutive critical failures exceed threshold, trigger self-healing restart
		if fails >= c.maxConsecutiveFailures && c.restartMgr != nil && !c.restartMgr.IsRestarting() {
			log.Printf("⚠️ [Self-Healing Watchdog] Exceeded max consecutive critical failures (%d). Triggering autonomous service restart...", fails)
			c.RecordIncident(Incident{
				ID:          fmt.Sprintf("inc-%d", time.Now().UnixNano()),
				Timestamp:   time.Now().UTC(),
				Severity:    "critical",
				Component:   "system",
				Message:     fmt.Sprintf("Autonomous restart triggered after %d consecutive health check failures", fails),
				ActionTaken: "restart_scheduled",
				Resolved:    true,
			})
			_ = c.restartMgr.TriggerRestart(RestartRequest{
				Reason:    fmt.Sprintf("Self-healing watchdog: %d consecutive critical failures", fails),
				Initiator: "watchdog",
				Delay:     1 * time.Second,
				Force:     true,
			})
		}
	} else {
		// Reset failure counter on healthy evaluation
		c.consecutiveFailures.Store(0)
		c.mu.Lock()
		c.stats.ConsecutiveFailures = 0
		c.mu.Unlock()
	}

	return report
}

// EvaluateHealth performs non-destructive probes of all key subsystems.
func (c *Controller) EvaluateHealth() SystemHealthReport {
	components := make(map[string]ComponentStatus)
	now := time.Now().UTC()
	overall := StatusHealthy

	// 1. BoltDB Evaluation
	dbStatus := c.checkDatabase()
	components["database"] = dbStatus
	if dbStatus.Status == StatusCritical {
		overall = StatusCritical
	} else if dbStatus.Status == StatusDegraded && overall == StatusHealthy {
		overall = StatusDegraded
	}

	// 2. Memory Evaluation
	memStatus := c.checkMemory()
	components["memory"] = memStatus
	if memStatus.Status == StatusCritical {
		overall = StatusCritical
	} else if memStatus.Status == StatusDegraded && overall == StatusHealthy {
		overall = StatusDegraded
	}

	// 3. Goroutine Evaluation
	goroutineStatus := c.checkGoroutines()
	components["goroutines"] = goroutineStatus
	if goroutineStatus.Status == StatusCritical {
		overall = StatusCritical
	} else if goroutineStatus.Status == StatusDegraded && overall == StatusHealthy {
		overall = StatusDegraded
	}

	// 4. Contentious Port State Evaluation
	portStatus := c.checkPort()
	components["port"] = portStatus
	if portStatus.Status == StatusCritical && overall != StatusCritical {
		overall = StatusCritical
	}

	c.mu.RLock()
	engineStats := c.stats
	c.mu.RUnlock()

	up := time.Since(c.startTime)

	return SystemHealthReport{
		Service:           c.cfg.ServiceName,
		Environment:       c.cfg.Environment,
		OverallStatus:     overall,
		Timestamp:         now,
		Uptime:            up.Round(time.Second).String(),
		UptimeSec:         up.Seconds(),
		Components:        components,
		Stats:             engineStats,
		SelfHealingActive: c.running.Load(),
	}
}

func (c *Controller) checkDatabase() ComponentStatus {
	start := time.Now()
	cs := ComponentStatus{
		Name:      "database",
		Status:    StatusHealthy,
		CheckedAt: start.UTC(),
		Details:   make(map[string]any),
	}

	if c.store == nil {
		cs.Status = StatusDegraded
		cs.Message = "Storage engine uninitialized or nil"
		return cs
	}

	// Ping probe: test write and read in BucketSelfHealing
	pingKey := fmt.Sprintf("_ping_%d", os.Getpid())
	pingVal := []byte(fmt.Sprintf("%d", time.Now().UnixNano()))

	err := c.store.PutJSON(db.BucketSelfHealing, pingKey, pingVal)
	if err != nil {
		cs.Status = StatusCritical
		cs.Message = fmt.Sprintf("Database write check failed: %v", err)
		cs.LatencyMs = float64(time.Since(start).Microseconds()) / 1000.0
		return cs
	}

	readVal, err := c.store.GetJSON(db.BucketSelfHealing, pingKey)
	if err != nil || string(readVal) != string(pingVal) {
		cs.Status = StatusCritical
		cs.Message = "Database read validation check failed"
		cs.LatencyMs = float64(time.Since(start).Microseconds()) / 1000.0
		return cs
	}

	latency := float64(time.Since(start).Microseconds()) / 1000.0
	cs.LatencyMs = latency

	stats, _ := c.store.GetStats()
	if stats != nil {
		cs.Details["size_bytes"] = stats.SizeBytes
		cs.Details["allocated_pages"] = stats.AllocatedPages
		cs.Details["key_count"] = stats.KeyCount
	}

	if latency > 50.0 { // > 50ms read/write latency is degraded for embedded BoltDB
		cs.Status = StatusDegraded
		cs.Message = fmt.Sprintf("High storage latency: %.2f ms", latency)
	} else {
		cs.Status = StatusHealthy
		cs.Message = fmt.Sprintf("BoltDB responsive (latency: %.2f ms)", latency)
	}

	return cs
}

func (c *Controller) checkMemory() ComponentStatus {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	allocMB := float64(m.Alloc) / 1024.0 / 1024.0
	sysMB := float64(m.Sys) / 1024.0 / 1024.0

	cs := ComponentStatus{
		Name:      "memory",
		Status:    StatusHealthy,
		CheckedAt: time.Now().UTC(),
		Details: map[string]any{
			"alloc_mb":          allocMB,
			"sys_mb":            sysMB,
			"total_alloc_mb":    float64(m.TotalAlloc) / 1024.0 / 1024.0,
			"num_gc":            m.NumGC,
			"warning_limit_mb":  c.memoryWarningMB,
			"critical_limit_mb": c.memoryCriticalMB,
		},
	}

	if allocMB >= c.memoryCriticalMB {
		cs.Status = StatusCritical
		cs.Message = fmt.Sprintf("Critical memory usage: %.1f MB (limit: %.1f MB)", allocMB, c.memoryCriticalMB)
	} else if allocMB >= c.memoryWarningMB {
		cs.Status = StatusDegraded
		cs.Message = fmt.Sprintf("Elevated memory allocation: %.1f MB (warning threshold: %.1f MB)", allocMB, c.memoryWarningMB)
	} else {
		cs.Status = StatusHealthy
		cs.Message = fmt.Sprintf("Memory optimal (%.1f MB allocated, %.1f MB sys)", allocMB, sysMB)
	}

	return cs
}

func (c *Controller) checkGoroutines() ComponentStatus {
	count := runtime.NumGoroutine()
	cs := ComponentStatus{
		Name:      "goroutines",
		Status:    StatusHealthy,
		CheckedAt: time.Now().UTC(),
		Details: map[string]any{
			"count":     count,
			"warning":   c.goroutineWarning,
			"critical":  c.goroutineCritical,
		},
	}

	if count >= c.goroutineCritical {
		cs.Status = StatusCritical
		cs.Message = fmt.Sprintf("Goroutine leak suspected: %d active routines", count)
	} else if count >= c.goroutineWarning {
		cs.Status = StatusDegraded
		cs.Message = fmt.Sprintf("High goroutine count: %d active routines", count)
	} else {
		cs.Status = StatusHealthy
		cs.Message = fmt.Sprintf("Goroutine count nominal (%d active routines)", count)
	}

	return cs
}

func (c *Controller) checkPort() ComponentStatus {
	cleanPort := c.cfg.Port
	cs := ComponentStatus{
		Name:      "port",
		Status:    StatusHealthy,
		CheckedAt: time.Now().UTC(),
		Details: map[string]any{
			"configured_port": cleanPort,
		},
	}

	// Find any other rogue processes on our port
	procs, err := portutil.FindProcessesOnPort(cleanPort)
	if err == nil && len(procs) > 0 {
		cs.Status = StatusDegraded
		cs.Message = fmt.Sprintf("Port %s has %d extraneous processes detected", cleanPort, len(procs))
		cs.Details["contentious_processes"] = procs
	} else {
		cs.Message = fmt.Sprintf("Port :%s sovereign and uncontended", cleanPort)
	}

	return cs
}

// RemediateMemory executes garbage collection and releases OS memory pages.
func (c *Controller) RemediateMemory() (float64, error) {
	var mBefore, mAfter runtime.MemStats
	runtime.ReadMemStats(&mBefore)

	runtime.GC()
	debug.FreeOSMemory()

	runtime.ReadMemStats(&mAfter)

	beforeMB := float64(mBefore.Alloc) / 1024.0 / 1024.0
	afterMB := float64(mAfter.Alloc) / 1024.0 / 1024.0
	freedMB := beforeMB - afterMB

	msg := fmt.Sprintf("Forced GC and OS memory purge completed: %.2f MB -> %.2f MB (reclaimed: %.2f MB)", beforeMB, afterMB, freedMB)
	log.Printf("🧹 [Self-Healing Engine] %s", msg)

	c.RecordIncident(Incident{
		ID:          fmt.Sprintf("inc-%d", time.Now().UnixNano()),
		Timestamp:   time.Now().UTC(),
		Severity:    "info",
		Component:   "memory",
		Message:     msg,
		ActionTaken: "forced_gc",
		Resolved:    true,
	})

	return freedMB, nil
}

// Remediate performs a requested remediation action on demand.
func (c *Controller) Remediate(action string) RemediateResponse {
	resp := RemediateResponse{
		Action:    action,
		Timestamp: time.Now().UTC(),
		Details:   make(map[string]any),
	}

	switch action {
	case "memory_gc", "free_os_mem":
		freed, err := c.RemediateMemory()
		if err != nil {
			resp.Success = false
			resp.Message = fmt.Sprintf("Memory remediation failed: %v", err)
		} else {
			resp.Success = true
			resp.Message = fmt.Sprintf("Successfully reclaimed %.2f MB via garbage collection and FreeOSMemory", freed)
			resp.Details["freed_mb"] = freed
		}

	case "db_verify":
		dbCheck := c.checkDatabase()
		resp.Success = dbCheck.Status == StatusHealthy
		resp.Message = dbCheck.Message
		resp.Details["db_status"] = dbCheck

	case "port_check":
		portCheck := c.checkPort()
		resp.Success = portCheck.Status == StatusHealthy
		resp.Message = portCheck.Message
		resp.Details["port_status"] = portCheck

	case "full_health_check":
		report := c.CheckAndHeal()
		resp.Success = report.OverallStatus == StatusHealthy
		resp.Message = fmt.Sprintf("Full health evaluation: %s", report.OverallStatus)
		resp.Details["report"] = report

	default:
		resp.Success = false
		resp.Message = fmt.Sprintf("Unknown remediation action %q", action)
	}

	return resp
}

// RecordIncident adds a diagnostic or self-healing event to the history log and BoltDB.
func (c *Controller) RecordIncident(inc Incident) {
	c.mu.Lock()
	c.incidents = append(c.incidents, inc)
	if len(c.incidents) > 100 {
		c.incidents = c.incidents[len(c.incidents)-100:]
	}
	c.mu.Unlock()

	if c.store != nil {
		data, err := json.Marshal(inc)
		if err == nil {
			key := fmt.Sprintf("incident:%d:%s", inc.Timestamp.UnixNano(), inc.ID)
			_ = c.store.PutJSON(db.BucketSelfHealing, key, data)
		}
	}
}

// RecordPanic intercepts an unhandled panic in HTTP handlers or workers,
// records it into the incident journal, and schedules restart if a panic storm occurs.
func (c *Controller) RecordPanic(rec any, stack []byte, path string) {
	now := time.Now().UTC()
	c.mu.Lock()
	c.stats.PanicsRecovered++
	c.mu.Unlock()

	c.panicMu.Lock()
	c.panicTimestamps = append(c.panicTimestamps, now)
	// Prune panic timestamps older than 60 seconds
	cutoff := now.Add(-60 * time.Second)
	recentPanics := 0
	filtered := make([]time.Time, 0, len(c.panicTimestamps))
	for _, t := range c.panicTimestamps {
		if t.After(cutoff) {
			filtered = append(filtered, t)
			recentPanics++
		}
	}
	c.panicTimestamps = filtered
	c.panicMu.Unlock()

	stackSnippet := string(stack)
	if len(stackSnippet) > 500 {
		stackSnippet = stackSnippet[:500] + "... (truncated)"
	}

	inc := Incident{
		ID:          fmt.Sprintf("panic-%d", time.Now().UnixNano()),
		Timestamp:   now,
		Severity:    "critical",
		Component:   "http_panic",
		Message:     fmt.Sprintf("Panic caught on %s: %v", path, rec),
		Details:     stackSnippet,
		ActionTaken: "none",
		Resolved:    true,
	}

	// Check if panic threshold reached for automated restart
	if recentPanics >= 5 && c.restartMgr != nil && !c.restartMgr.IsRestarting() {
		inc.ActionTaken = "restart_scheduled"
		log.Printf("🚨 [Self-Healing Engine] Panic storm detected (%d panics in 60s). Scheduling emergency restart...", recentPanics)
		_ = c.restartMgr.TriggerRestart(RestartRequest{
			Reason:    fmt.Sprintf("Emergency restart: %d panics in 60 seconds", recentPanics),
			Initiator: "panic_watchdog",
			Delay:     500 * time.Millisecond,
			Force:     true,
		})
	}

	c.RecordIncident(inc)
}

// GetIncidents returns recent incidents in reverse chronological order.
func (c *Controller) GetIncidents(limit int) []Incident {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if limit <= 0 || limit > len(c.incidents) {
		limit = len(c.incidents)
	}

	out := make([]Incident, limit)
	for i := 0; i < limit; i++ {
		out[i] = c.incidents[len(c.incidents)-1-i]
	}
	return out
}

func (c *Controller) loadIncidents() {
	if c.store == nil {
		return
	}
	keys, err := c.store.ListKeys(db.BucketSelfHealing, "incident:")
	if err != nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, k := range keys {
		data, err := c.store.GetJSON(db.BucketSelfHealing, k)
		if err != nil || len(data) == 0 {
			continue
		}
		var inc Incident
		if err := json.Unmarshal(data, &inc); err == nil {
			c.incidents = append(c.incidents, inc)
		}
	}
}
