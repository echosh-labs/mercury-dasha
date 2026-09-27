package api

import (
	"encoding/json"
	"io"
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

// DiscoveryRequest represents a batch of discoveries submitted by Sovereign Observer.
type DiscoveryRequest struct {
	Source      string                    `json:"source"`
	Discoveries []axismundi.WorkspaceItem `json:"discoveries"`
	Items       []axismundi.WorkspaceItem `json:"items"`
}

// AxisMundiDiscoveriesHandler receives newly registered files, emails, or keep notes discovered by Sovereign Observer.
func (h *Handler) AxisMundiDiscoveriesHandler(w http.ResponseWriter, r *http.Request) {
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

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "failed to read body: " + err.Error()})
		return
	}

	var items []axismundi.WorkspaceItem
	source := "sovereign_observer"

	// 1. Try decoding as DiscoveryRequest object
	var discReq DiscoveryRequest
	if err := json.Unmarshal(bodyBytes, &discReq); err == nil && (len(discReq.Discoveries) > 0 || len(discReq.Items) > 0) {
		if discReq.Source != "" {
			source = discReq.Source
		}
		items = append(items, discReq.Discoveries...)
		items = append(items, discReq.Items...)
	} else {
		// 2. Try decoding as array of WorkspaceItem
		var rawList []axismundi.WorkspaceItem
		if err := json.Unmarshal(bodyBytes, &rawList); err == nil && len(rawList) > 0 {
			items = rawList
		} else {
			// 3. Try decoding as single WorkspaceItem
			var singleItem axismundi.WorkspaceItem
			if err := json.Unmarshal(bodyBytes, &singleItem); err == nil && singleItem.ID != "" {
				items = []axismundi.WorkspaceItem{singleItem}
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid discovery payload: expected object or array with item id and title"})
				return
			}
		}
	}

	// Filter valid items
	var validItems []axismundi.WorkspaceItem
	for _, it := range items {
		if it.ID != "" && it.Title != "" {
			if it.Source == "" {
				it.Source = source
			}
			validItems = append(validItems, it)
		}
	}

	if len(validItems) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "no valid items in discovery payload"})
		return
	}

	newAlerts := h.axisMundi.ProcessItems(validItems)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":    true,
		"source":     source,
		"received":   len(validItems),
		"new_alerts": len(newAlerts),
		"alerts":     newAlerts,
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

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "failed to read body: " + err.Error()})
		return
	}

	// Try single item first for backward compatibility
	var item axismundi.WorkspaceItem
	if err := json.Unmarshal(bodyBytes, &item); err == nil && item.ID != "" && item.Title != "" {
		alert, isNew := h.axisMundi.IngestEvent(item)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"received": true,
			"is_new":   isNew,
			"alert":    alert,
		})
		return
	}

	// Fallback to discovery request
	var discReq DiscoveryRequest
	if err := json.Unmarshal(bodyBytes, &discReq); err == nil && (len(discReq.Discoveries) > 0 || len(discReq.Items) > 0) {
		items := append(discReq.Discoveries, discReq.Items...)
		newAlerts := h.axisMundi.ProcessItems(items)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"received":   true,
			"new_alerts": len(newAlerts),
			"alerts":     newAlerts,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": "item id and title are required"})
}
