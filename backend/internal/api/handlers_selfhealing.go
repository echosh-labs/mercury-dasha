package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/selfhealing"
)

func extractAdminToken(r *http.Request) string {
	if token := r.Header.Get("X-Admin-Token"); token != "" {
		return strings.TrimSpace(token)
	}
	if auth := r.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		}
		return strings.TrimSpace(auth)
	}
	return strings.TrimSpace(r.URL.Query().Get("token"))
}

// SystemHealthHandler returns a comprehensive multi-factor health report across database, memory, and goroutines.
func (h *Handler) SystemHealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.selfHealing == nil {
		http.Error(w, `{"error":"self-healing controller not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	report := h.selfHealing.EvaluateHealth()
	_ = json.NewEncoder(w).Encode(report)
}

// SelfHealingCheckHandler triggers an immediate diagnostic evaluation and executes any needed auto-remediation.
func (h *Handler) SelfHealingCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.selfHealing == nil {
		http.Error(w, `{"error":"self-healing controller not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	report := h.selfHealing.CheckAndHeal()
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message": "Health evaluation and auto-remediation check complete",
		"report":  report,
	})
}

// SelfHealingRemediateHandler executes targeted remediation (e.g. memory_gc, db_verify, port_check).
func (h *Handler) SelfHealingRemediateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.selfHealing == nil {
		http.Error(w, `{"error":"self-healing controller not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	var req selfhealing.RemediateRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	if req.Action == "" {
		req.Action = "memory_gc"
	}

	resp := h.selfHealing.Remediate(req.Action)
	if !resp.Success {
		w.WriteHeader(http.StatusBadRequest)
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// SelfHealingIncidentsHandler returns the journal of diagnostic incidents and auto-remediations.
func (h *Handler) SelfHealingIncidentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.selfHealing == nil {
		http.Error(w, `{"error":"self-healing controller not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if il, err := strconv.Atoi(l); err == nil && il > 0 {
			limit = il
		}
	}

	incidents := h.selfHealing.GetIncidents(limit)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"total":     len(incidents),
		"incidents": incidents,
	})
}

type RestartAPIRequest struct {
	Reason  string `json:"reason"`
	DelayMs int    `json:"delay_ms"`
	Force   bool   `json:"force"`
	DryRun  bool   `json:"dry_run"`
}

type RestartAPIResponse struct {
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Reason      string    `json:"reason"`
	Initiator   string    `json:"initiator"`
	PID         int       `json:"pid"`
	DelayMs     int       `json:"delay_ms"`
}

// RestartHandler receives remote administrative requests to gracefully restart the service.
func (h *Handler) RestartHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.restartMgr == nil {
		http.Error(w, `{"error":"restart manager not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	// Security verification
	token := extractAdminToken(r)
	if !h.restartMgr.ValidateToken(token) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "unauthorized: invalid or missing restart token",
		})
		return
	}

	var req RestartAPIRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	if req.Reason == "" {
		req.Reason = fmt.Sprintf("Remote restart requested by operator at %s", r.RemoteAddr)
	}
	if req.DelayMs <= 0 {
		req.DelayMs = 500
	}

	delayDur := time.Duration(req.DelayMs) * time.Millisecond

	restartReq := selfhealing.RestartRequest{
		Reason:    req.Reason,
		Initiator: "remote_api",
		Delay:     delayDur,
		Force:     req.Force,
		DryRun:    req.DryRun,
	}

	err := h.restartMgr.TriggerRestart(restartReq)
	if err != nil {
		if err == selfhealing.ErrRestartInProgress {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "service restart already in progress",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("failed to initiate restart: %v", err),
		})
		return
	}

	scheduled := time.Now().UTC().Add(delayDur)
	resp := RestartAPIResponse{
		Status:      "restarting",
		Message:     fmt.Sprintf("Mercury Dasha service is restarting in %dms", req.DelayMs),
		ScheduledAt: scheduled,
		Reason:      req.Reason,
		Initiator:   "remote_api",
		PID:         os.Getpid(),
		DelayMs:     req.DelayMs,
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// RestartHistoryHandler returns recent restart audit history and pending status.
func (h *Handler) RestartHistoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.restartMgr == nil {
		http.Error(w, `{"error":"restart manager not initialized"}`, http.StatusServiceUnavailable)
		return
	}

	history := h.restartMgr.GetHistory()
	_ = json.NewEncoder(w).Encode(map[string]any{
		"total":          len(history),
		"is_restarting":  h.restartMgr.IsRestarting(),
		"active_pid":     os.Getpid(),
		"restart_events": history,
	})
}
