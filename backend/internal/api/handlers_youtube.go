package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/echosh-labs/mercury-dasha/internal/amra"
	"github.com/echosh-labs/mercury-dasha/internal/youtube"
)

// YouTubeStatusHandler returns channel stats, auth state, quota usage, and recent jobs.
func (h *Handler) YouTubeStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Canonical-Service", "amra-treasury")
	w.Header().Set("X-Canonical-Port", "8050")
	w.Header().Set("X-Canonical-URL", "http://localhost:8050/api/v1/youtube/status")

	if h.youtube == nil {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"configured":        false,
			"authenticated":     false,
			"canonical_service": "amra-treasury",
			"canonical_port":    "8050",
			"canonical_url":     "http://localhost:8050/api/v1/youtube/status",
			"notice":            "YouTube Studio and publication pipeline has migrated to AMRA Treasury (Port 8050).",
		})
		return
	}

	status := h.youtube.GetStatus(r.Context())
	_ = json.NewEncoder(w).Encode(status)
}

// YouTubeAuthURLHandler generates the Google OAuth 2.0 consent URL.
func (h *Handler) YouTubeAuthURLHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.youtube == nil || !h.youtube.IsConfigured(r.Context()) {
		http.Error(w, `{"error":"youtube client_id or client_secret is not configured"}`, http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		b := make([]byte, 16)
		_, _ = rand.Read(b)
		state = hex.EncodeToString(b)
	}

	authURL, err := h.youtube.GetAuthURL(r.Context(), state)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{
		"auth_url": authURL,
		"state":    state,
	})
}

// YouTubeAuthCallbackHandler processes the Google OAuth2 redirect code.
func (h *Handler) YouTubeAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		http.Error(w, fmt.Sprintf("OAuth error: %s", errParam), http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	if h.youtube == nil {
		http.Error(w, "YouTube client engine not initialized", http.StatusInternalServerError)
		return
	}

	_, err := h.youtube.ExchangeCode(r.Context(), code)
	if err != nil {
		log.Printf("[YouTube OAuth] Token exchange failed: %v", err)
		http.Error(w, fmt.Sprintf("Token exchange failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect back to root SPA with pairing query param
	http.Redirect(w, r, "/?youtube_auth=success", http.StatusTemporaryRedirect)
}

// YouTubeDisconnectHandler revokes and wipes stored YouTube credentials.
func (h *Handler) YouTubeDisconnectHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.youtube == nil {
		http.Error(w, `{"error":"youtube engine not initialized"}`, http.StatusInternalServerError)
		return
	}

	if err := h.youtube.Disconnect(r.Context()); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "disconnected",
		"message": "YouTube credentials successfully revoked and cleared",
	})
}

// YouTubeUploadHandler handles either JSON requests with existing POSIX file paths or multipart video uploads.
func (h *Handler) YouTubeUploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.youtube == nil || !h.youtube.IsAuthenticated() {
		http.Error(w, `{"error":"youtube is not authenticated. Please pair your channel first."}`, http.StatusUnauthorized)
		return
	}

	contentType := r.Header.Get("Content-Type")
	var req youtube.UploadRequest

	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Parse multipart upload
		if err := r.ParseMultipartForm(500 << 20); err != nil { // 500 MB in-memory limit for headers/chunks
			http.Error(w, fmt.Sprintf(`{"error":"failed to parse multipart form: %v"}`, err), http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, `{"error":"form field 'file' is required"}`, http.StatusBadRequest)
			return
		}
		defer file.Close()

		uploadDir := filepath.Join(os.TempDir(), "mercury_youtube_uploads")
		_ = os.MkdirAll(uploadDir, 0755)
		tempPath := filepath.Join(uploadDir, fmt.Sprintf("%d_%s", r.Context().Value("ts"), header.Filename))

		dst, err := os.Create(tempPath)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"failed to save temporary file: %v"}`, err), http.StatusInternalServerError)
			return
		}
		if _, err := io.Copy(dst, file); err != nil {
			dst.Close()
			_ = os.Remove(tempPath)
			http.Error(w, fmt.Sprintf(`{"error":"failed to stream uploaded file: %v"}`, err), http.StatusInternalServerError)
			return
		}
		dst.Close()

		req.FilePath = tempPath
		req.Title = r.FormValue("title")
		req.Description = r.FormValue("description")
		req.CategoryID = r.FormValue("category_id")
		req.PrivacyStatus = r.FormValue("privacy_status")
		req.MadeForKids = r.FormValue("made_for_kids") == "true"
		req.Embeddable = r.FormValue("embeddable") != "false"
		req.AttachChronoContext = r.FormValue("attach_chrono_context") == "true"

		if tagsStr := r.FormValue("tags"); tagsStr != "" {
			parts := strings.Split(tagsStr, ",")
			for _, p := range parts {
				if t := strings.TrimSpace(p); t != "" {
					req.Tags = append(req.Tags, t)
				}
			}
		}
	} else {
		// JSON Body with local file path
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid json request body"}`, http.StatusBadRequest)
			return
		}
	}

	if strings.TrimSpace(req.FilePath) == "" {
		http.Error(w, `{"error":"file_path is required"}`, http.StatusBadRequest)
		return
	}

	// Verify file exists
	cleanPath := filepath.Clean(req.FilePath)
	if _, err := os.Stat(cleanPath); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"file not found at %q"}`, cleanPath), http.StatusNotFound)
		return
	}

	async := r.URL.Query().Get("async") != "false"

	if async {
		// Create initial queued job record
		job, err := h.youtube.UploadVideo(r.Context(), req, "ui_manual")
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(job)
		return
	}

	// Synchronous execution
	job, err := h.youtube.UploadVideo(r.Context(), req, "ui_manual")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(job)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(job)
}

// YouTubeJobsHandler lists recent upload jobs.
func (h *Handler) YouTubeJobsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.youtube == nil {
		http.Error(w, `{"error":"youtube engine not initialized"}`, http.StatusInternalServerError)
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	jobs, err := h.youtube.ListJobs(limit)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"total": len(jobs),
		"jobs":  jobs,
	})
}

// YouTubeJobDetailHandler handles GET job status and POST cancel.
func (h *Handler) YouTubeJobDetailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.youtube == nil {
		http.Error(w, `{"error":"youtube engine not initialized"}`, http.StatusInternalServerError)
		return
	}

	// Path: /api/v1/youtube/jobs/{id} or /api/v1/youtube/jobs/{id}/cancel
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/youtube/jobs/")
	parts := strings.Split(path, "/")
	id := parts[0]
	if id == "" {
		http.Error(w, `{"error":"job id required"}`, http.StatusBadRequest)
		return
	}

	isCancel := len(parts) > 1 && parts[1] == "cancel"

	if r.Method == http.MethodPost && isCancel {
		if err := h.youtube.CancelJob(id); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "cancelled",
			"job_id": id,
		})
		return
	}

	job, err := h.youtube.GetJob(id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"job %q not found"}`, id), http.StatusNotFound)
		return
	}

	_ = json.NewEncoder(w).Encode(job)
}

// YouTubeVideosHandler lists recent uploaded videos on the channel.
func (h *Handler) YouTubeVideosHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.youtube == nil || !h.youtube.IsAuthenticated() {
		http.Error(w, `{"error":"youtube is not authenticated"}`, http.StatusUnauthorized)
		return
	}

	limit := int64(20)
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.ParseInt(l, 10, 64); err == nil && n > 0 {
			limit = n
		}
	}

	videos, err := h.youtube.ListRecentVideos(r.Context(), limit)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"total":  len(videos),
		"videos": videos,
	})
}

// YouTubeAnalyticsHandler returns viewer retention and performance metrics.
func (h *Handler) YouTubeAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.youtube == nil || !h.youtube.IsAuthenticated() {
		http.Error(w, `{"error":"youtube is not authenticated"}`, http.StatusUnauthorized)
		return
	}

	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	report, err := h.youtube.GetAnalyticsReport(r.Context(), startDate, endDate)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(report)
}

// YouTubeFinanceHandler returns monetization, ad revenue, and CPM metrics.
func (h *Handler) YouTubeFinanceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.youtube == nil || !h.youtube.IsAuthenticated() {
		http.Error(w, `{"error":"youtube is not authenticated"}`, http.StatusUnauthorized)
		return
	}

	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	report, err := h.youtube.GetFinancialReport(r.Context(), startDate, endDate)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(report)
}

// YouTubeSyncAmraHandler ingests daily YouTube revenue records into the AMRA audit ledger.
func (h *Handler) YouTubeSyncAmraHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.youtube == nil || !h.youtube.IsAuthenticated() {
		http.Error(w, `{"error":"youtube is not authenticated"}`, http.StatusUnauthorized)
		return
	}
	if h.amraEngine == nil {
		http.Error(w, `{"error":"amra financial core not initialized"}`, http.StatusInternalServerError)
		return
	}

	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	report, err := h.youtube.GetFinancialReport(r.Context(), startDate, endDate)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	ch, _ := h.youtube.GetChannelProfile(r.Context())
	channelID := "unknown"
	if ch != nil {
		channelID = ch.ChannelID
	}

	syncedCount := 0
	skippedCount := 0
	var totalLedgedDollars float64

	for _, row := range report.DailyRows {
		if row.EstimatedRevenue <= 0 {
			continue
		}

		_, err := h.amraEngine.IngestYouTubeRevenue(r.Context(), amra.YouTubeRevenueRecord{
			Day:              row.Day,
			EstimatedRevenue: row.EstimatedRevenue,
			MonetizedPlays:   row.MonetizedPlaybacks,
			CPM:              row.CPM,
			ChannelID:        channelID,
		})
		if err != nil {
			if errors.Is(err, amra.ErrDuplicateEvent) {
				skippedCount++
				continue
			}
			log.Printf("[AMRA Bridge] Failed to ingest revenue for day %s: %v", row.Day, err)
			continue
		}
		syncedCount++
		totalLedgedDollars += row.EstimatedRevenue
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":        "synced",
		"synced_count":  syncedCount,
		"skipped_count": skippedCount,
		"total_dollars": totalLedgedDollars,
		"channel_id":    channelID,
	})
}

