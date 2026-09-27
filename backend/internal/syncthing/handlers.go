package syncthing

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// HTTPHandler exposes REST endpoints for the Syncthing subsystem.
type HTTPHandler struct {
	subsystem *Subsystem
}

// NewHTTPHandler creates HTTP handler wrapping the Syncthing subsystem.
func NewHTTPHandler(subsystem *Subsystem) *HTTPHandler {
	return &HTTPHandler{
		subsystem: subsystem,
	}
}

// StatusHandler reports aggregated Syncthing subsystem health and telemetry.
func (h *HTTPHandler) StatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status, err := h.subsystem.GetStatus(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": err.Error(),
		})
		return
	}
	_ = json.NewEncoder(w).Encode(status)
}

// FeedHandler returns recent synchronized media items.
func (h *HTTPHandler) FeedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	limitStr := r.URL.Query().Get("limit")
	limit := 30
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	items := h.subsystem.GetFeed(limit)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"total": len(items),
		"items": items,
	})
}

// RescanHandler commands Syncthing to trigger an immediate rescan of media repositories.
func (h *HTTPHandler) RescanHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	folderID := r.URL.Query().Get("folder")

	if err := h.subsystem.TriggerRescan(r.Context(), folderID); err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Syncthing rescan initiated",
		"folder":  folderID,
	})
}

// StreamHandler streams high-resolution media files with byte-range HTTP seeking.
func (h *HTTPHandler) StreamHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "missing path parameter", http.StatusBadRequest)
		return
	}
	h.subsystem.StreamMedia(w, r, path)
}

// ThumbnailHandler serves generated poster thumbnails or image previews.
func (h *HTTPHandler) ThumbnailHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = r.URL.Query().Get("path")
	}
	if name == "" {
		http.Error(w, "missing name or path parameter", http.StatusBadRequest)
		return
	}
	h.subsystem.ServeThumbnail(w, r, name)
}
