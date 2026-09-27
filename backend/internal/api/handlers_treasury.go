package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TreasurySuccessAlert represents an incoming or broadcasted treasury milestone event.
type TreasurySuccessAlert struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"` // "youtube_view", "youtube_subscriber", "youtube_milestone", "amra_subscriber", "treasury_tx", "revenue_yield"
	Title        string    `json:"title"`
	Message      string    `json:"message"`
	MetricLabel  string    `json:"metric_label,omitempty"`
	MetricValue  string    `json:"metric_value,omitempty"`
	Delta        any       `json:"delta,omitempty"`
	ChannelTitle string    `json:"channel_title,omitempty"`
	VideoTitle   string    `json:"video_title,omitempty"`
	VideoURL     string    `json:"video_url,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// TreasuryEventHandler receives pushed events from AMRA Treasury (port 8050) or external webhooks
// and broadcasts them via Chrono-Pulse SSE stream to all connected frontend clients.
func (h *Handler) TreasuryEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to read request body"})
		return
	}

	var alert TreasurySuccessAlert
	if err := json.Unmarshal(body, &alert); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid alert payload: " + err.Error()})
		return
	}

	if alert.ID == "" {
		alert.ID = fmt.Sprintf("treasury-%s-%d", alert.Type, time.Now().UnixNano())
	}
	if alert.Timestamp.IsZero() {
		alert.Timestamp = time.Now().UTC()
	}
	if alert.Title == "" {
		alert.Title = "🌟 AMRA Treasury Success Milestone"
	}

	// Broadcast to ChronoPulse SSE stream if active
	if h.metronome != nil {
		h.metronome.BroadcastEvent("treasury_success_alert", alert)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"received":  true,
		"alert":     alert,
		"broadcast": h.metronome != nil,
	})
}

// TreasuryStatusHandler returns the operational status of the sovereign AMRA Treasury link.
func (h *Handler) TreasuryStatusHandler(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{Timeout: 2 * time.Second}
	isLive := false

	resp, err := client.Get("http://127.0.0.1:8050/healthz")
	if err == nil {
		if resp.StatusCode == http.StatusOK {
			isLive = true
		}
		_ = resp.Body.Close()
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"service":     "amra-treasury",
		"service_url": "http://localhost:8050",
		"is_live":     isLive,
		"timestamp":   time.Now().UTC(),
	})
}
