package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/axismundi"
)

// AxisMundiStatusHandler reports connectivity, operational mode, and observation metrics.
func (h *Handler) AxisMundiStatusHandler(w http.ResponseWriter, r *http.Request) {
	if h.axisMundi == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": "Axis Mundi listener subsystem is not initialized",
		})
		return
	}

	status := h.axisMundi.GetStatus()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}

// AxisMundiFeedHandler returns the aggregated feed of observed workspace items (Keep notes, Gmail, Docs, Sheets).
func (h *Handler) AxisMundiFeedHandler(w http.ResponseWriter, r *http.Request) {
	if h.axisMundi == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": "Axis Mundi listener subsystem is not initialized",
		})
		return
	}

	filterType := r.URL.Query().Get("type")
	newOnly := r.URL.Query().Get("new_only") == "true" || r.URL.Query().Get("new_only") == "1"

	feed := h.axisMundi.GetFeed(filterType, newOnly)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(feed)
}

// AxisMundiSyncHandler triggers an immediate on-demand poll against the Axis Mundi service.
func (h *Handler) AxisMundiSyncHandler(w http.ResponseWriter, r *http.Request) {
	if h.axisMundi == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": "Axis Mundi listener subsystem is not initialized",
		})
		return
	}

	ctx := r.Context()
	alerts, err := h.axisMundi.Poll(ctx)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":     err.Error(),
			"timestamp": time.Now(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":    true,
		"new_alerts": len(alerts),
		"alerts":     alerts,
		"status":     h.axisMundi.GetStatus(),
	})
}

// AxisMundiEventHandler receives pushed workspace events (e.g. from Axis Mundi webhooks or agents).
func (h *Handler) AxisMundiEventHandler(w http.ResponseWriter, r *http.Request) {
	if h.axisMundi == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": "Axis Mundi listener subsystem is not initialized",
		})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var item axismundi.WorkspaceItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid payload: " + err.Error()})
		return
	}

	if item.ID == "" || item.Title == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "item id and title are required"})
		return
	}

	alert, isNew := h.axisMundi.IngestEvent(item)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"received": true,
		"is_new":   isNew,
		"alert":    alert,
	})
}
