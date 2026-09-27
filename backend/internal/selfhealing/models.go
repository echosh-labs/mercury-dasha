package selfhealing

import (
	"time"
)

// HealthStatus represents the high-level health state of a component or system.
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "HEALTHY"
	StatusDegraded  HealthStatus = "DEGRADED"
	StatusCritical  HealthStatus = "CRITICAL"
	StatusUnhealthy HealthStatus = "UNHEALTHY"
)

// ComponentStatus describes the operational state of a single subsystem.
type ComponentStatus struct {
	Name      string         `json:"name"`
	Status    HealthStatus   `json:"status"`
	Message   string         `json:"message"`
	LatencyMs float64        `json:"latency_ms,omitempty"`
	CheckedAt time.Time      `json:"checked_at"`
	Details   map[string]any `json:"details,omitempty"`
}

// SystemHealthReport contains the aggregated multi-factor health evaluation.
type SystemHealthReport struct {
	Service        string                     `json:"service"`
	Environment    string                     `json:"environment"`
	OverallStatus  HealthStatus               `json:"overall_status"`
	Timestamp      time.Time                  `json:"timestamp"`
	Uptime         string                     `json:"uptime"`
	UptimeSec      float64                    `json:"uptime_sec"`
	Components     map[string]ComponentStatus `json:"components"`
	Stats          EngineStats                `json:"stats"`
	SelfHealingActive bool                    `json:"self_healing_active"`
}

// EngineStats aggregates watchdog and self-healing telemetry.
type EngineStats struct {
	TotalChecks          uint64     `json:"total_checks"`
	AutoHealsTriggered   uint64     `json:"auto_heals_triggered"`
	PanicsRecovered      uint64     `json:"panics_recovered"`
	ConsecutiveFailures  uint64     `json:"consecutive_failures"`
	LastCheckTime        *time.Time `json:"last_check_time,omitempty"`
	LastHealedAt         *time.Time `json:"last_healed_at,omitempty"`
	LastHealAction       string     `json:"last_heal_action,omitempty"`
	RestartPending       bool       `json:"restart_pending"`
}

// Incident represents a diagnostic or self-healing event.
type Incident struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Severity    string    `json:"severity"` // "info", "warning", "critical"
	Component   string    `json:"component"` // "memory", "database", "goroutines", "http_panic", "port", "restart"
	Message     string    `json:"message"`
	Details     string    `json:"details,omitempty"`
	ActionTaken string    `json:"action_taken"` // "none", "forced_gc", "free_os_mem", "db_verify", "restart_scheduled", "restarted"
	Resolved    bool      `json:"resolved"`
}

// RestartRequest defines options sent when requesting a service restart.
type RestartRequest struct {
	Reason      string        `json:"reason"`
	Initiator   string        `json:"initiator"` // "remote_api", "watchdog", "operator", "cli"
	Delay       time.Duration `json:"delay_ms,omitempty"`
	Force       bool          `json:"force,omitempty"`
	DryRun      bool          `json:"dry_run,omitempty"`
}

// RestartRecord is an immutable audit log entry for a restart event.
type RestartRecord struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Reason    string    `json:"reason"`
	Initiator string    `json:"initiator"`
	Type      string    `json:"type"` // "remote_api", "watchdog_self_healing", "manual_signal"
	PID       int       `json:"pid"`
	Success   bool      `json:"success"`
	Details   string    `json:"details,omitempty"`
}

// RemediateRequest specifies manual or programmatic remediation requests.
type RemediateRequest struct {
	Action string `json:"action"` // "memory_gc", "free_os_mem", "db_verify", "port_check", "full_health_check"
}

// RemediateResponse provides the result of an attempted self-healing remediation.
type RemediateResponse struct {
	Action      string    `json:"action"`
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	Details     map[string]any `json:"details,omitempty"`
}
