package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/session"
)

// ListSessionsHandler returns all recorded D&D audio sessions.
// GET /api/v1/sessions?campaign=...
func (h *Handler) ListSessionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	campaign := r.URL.Query().Get("campaign")
	list, err := h.sessions.ListSessions(campaign)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to list sessions: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"sessions": list,
		"total":    len(list),
	})
}

// CreateSessionHandler initializes a new session record and sets up storage.
// POST /api/v1/sessions
func (h *Handler) CreateSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req session.AudioSession
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		http.Error(w, `{"error":"invalid JSON request body"}`, http.StatusBadRequest)
		return
	}

	req.EnsureDefaults()

	if err := h.sessions.CreateSession(&req); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to create session: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// If format is WAV, pre-initialize a standard 44-byte RIFF header on disk
	if strings.Contains(req.Format, "wav") {
		fullDiskPath := h.sessions.ResolveDiskPath(req.FilePath)
		header := session.GenerateWavHeader(req.Channels, req.SampleRate, req.BitDepth, 0)
		_, _ = session.AppendChunkToFile(fullDiskPath, header)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(req)
}

// GetSessionHandler returns a single session with markers and slices.
// GET /api/v1/sessions/{id...}
func (h *Handler) GetSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/sessions/")
	// Strip sub-routes like /markers or /chunk if called directly
	id = strings.Split(id, "/")[0]
	if id == "" {
		http.Error(w, `{"error":"session id is required"}`, http.StatusBadRequest)
		return
	}

	sess, err := h.sessions.GetSession(id)
	if err != nil {
		if err == session.ErrSessionNotFound {
			http.Error(w, `{"error":"session not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf(`{"error":"failed to get session: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sess)
}

// SessionChunkHandler appends raw audio bytes to the session file on disk.
// POST /api/v1/sessions/{id}/chunk
func (h *Handler) SessionChunkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Extract session ID from path /api/v1/sessions/{id}/chunk
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/sessions/"), "/")
	if len(parts) < 2 || parts[0] == "" {
		http.Error(w, `{"error":"invalid session chunk path"}`, http.StatusBadRequest)
		return
	}
	sessionID := parts[0]

	sess, err := h.sessions.GetSession(sessionID)
	if err != nil {
		http.Error(w, `{"error":"session not found"}`, http.StatusNotFound)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil || len(bodyBytes) == 0 {
		http.Error(w, `{"error":"empty chunk body"}`, http.StatusBadRequest)
		return
	}

	fullDiskPath := h.sessions.ResolveDiskPath(sess.FilePath)
	newSize, err := session.AppendChunkToFile(fullDiskPath, bodyBytes)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to append chunk: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"session_id": sessionID,
		"size_bytes": newSize,
		"status":     "appended",
	})
}

// CompleteSessionHandler finalizes a session, recalculating headers and duration.
// POST /api/v1/sessions/{id}/complete
func (h *Handler) CompleteSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/sessions/"), "/")
	if len(parts) < 2 || parts[0] == "" {
		http.Error(w, `{"error":"invalid complete session path"}`, http.StatusBadRequest)
		return
	}
	sessionID := parts[0]

	sess, err := h.sessions.CompleteSession(sessionID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to complete session: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sess)
}

// AddMarkerHandler adds a timestamped D&D event bookmark to the session.
// POST /api/v1/sessions/{id}/markers
func (h *Handler) AddMarkerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/sessions/"), "/")
	if len(parts) < 2 || parts[0] == "" {
		http.Error(w, `{"error":"invalid marker path"}`, http.StatusBadRequest)
		return
	}
	sessionID := parts[0]

	var marker session.SessionMarker
	if err := json.NewDecoder(r.Body).Decode(&marker); err != nil {
		http.Error(w, `{"error":"invalid marker payload"}`, http.StatusBadRequest)
		return
	}

	updated, err := h.sessions.AddMarker(sessionID, marker)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to add marker: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(updated)
}

// SessionStreamHandler streams the audio file with full HTTP 206 Range seeking support.
// GET /api/v1/sessions/{id}/stream
func (h *Handler) SessionStreamHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/sessions/"), "/")
	if len(parts) < 2 || parts[0] == "" {
		http.Error(w, `{"error":"invalid stream path"}`, http.StatusBadRequest)
		return
	}
	sessionID := parts[0]

	sess, err := h.sessions.GetSession(sessionID)
	if err != nil {
		http.Error(w, `{"error":"session not found"}`, http.StatusNotFound)
		return
	}

	fullDiskPath := h.sessions.ResolveDiskPath(sess.FilePath)
	fi, err := os.Stat(fullDiskPath)
	if err != nil || fi.IsDir() {
		http.Error(w, `{"error":"audio file not found on disk"}`, http.StatusNotFound)
		return
	}

	// Set content-disposition for inline playback
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", filepath.Base(sess.FilePath)))
	if strings.HasSuffix(strings.ToLower(sess.FilePath), ".wav") {
		w.Header().Set("Content-Type", "audio/wav")
	} else if strings.HasSuffix(strings.ToLower(sess.FilePath), ".webm") {
		w.Header().Set("Content-Type", "audio/webm")
	} else if strings.HasSuffix(strings.ToLower(sess.FilePath), ".mp3") {
		w.Header().Set("Content-Type", "audio/mpeg")
	}

	// http.ServeFile natively supports RFC 7233 range requests (HTTP 206 Partial Content)
	http.ServeFile(w, r, fullDiskPath)
}

// PrecisionSliceHandler extracts a sample-accurate slice and returns it as a standalone WAV.
// GET /api/v1/sessions/{id}/slice?start_ms=1000&end_ms=4500&label=Crit&save=true
func (h *Handler) PrecisionSliceHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/sessions/"), "/")
	if len(parts) < 2 || parts[0] == "" {
		http.Error(w, `{"error":"invalid slice path"}`, http.StatusBadRequest)
		return
	}
	sessionID := parts[0]

	sess, err := h.sessions.GetSession(sessionID)
	if err != nil {
		http.Error(w, `{"error":"session not found"}`, http.StatusNotFound)
		return
	}

	startMs, _ := strconv.ParseInt(r.URL.Query().Get("start_ms"), 10, 64)
	endMs, _ := strconv.ParseInt(r.URL.Query().Get("end_ms"), 10, 64)
	label := r.URL.Query().Get("label")
	category := r.URL.Query().Get("category")
	shouldSave := r.URL.Query().Get("save") == "true"

	if label == "" {
		label = fmt.Sprintf("Slice %d-%dms", startMs, endMs)
	}

	fullDiskPath := h.sessions.ResolveDiskPath(sess.FilePath)
	file, err := os.Open(fullDiskPath)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to open audio file: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		http.Error(w, `{"error":"failed to stat audio file"}`, http.StatusInternalServerError)
		return
	}

	// If save=true requested, persist slice to disk and register in DB
	if shouldSave {
		sliceFilename := fmt.Sprintf("%s_slice_%d_%d.wav", session.SanitizeFilename(sess.ID), startMs, endMs)
		sliceDestPath := filepath.Join(h.sessions.GetBasePath(), "slices", sliceFilename)
		savedSlice, saveErr := session.SavePrecisionSlice(fullDiskPath, sliceDestPath, startMs, endMs, label, category)
		if saveErr != nil {
			http.Error(w, fmt.Sprintf(`{"error":"failed to extract and save slice: %s"}`, saveErr.Error()), http.StatusBadRequest)
			return
		}
		_, _ = h.sessions.AddSlice(sessionID, *savedSlice)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(savedSlice)
		return
	}

	// Otherwise, stream slice on the fly with dynamic RIFF/WAVE header
	w.Header().Set("Content-Type", "audio/wav")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", fmt.Sprintf("slice_%s_%d_%d.wav", sess.ID, startMs, endMs)))

	slice, err := session.ExtractPrecisionSlice(file, fi.Size(), startMs, endMs, w)
	if err != nil {
		// Note: header may already have been sent if partial failure occurred
		return
	}
	_ = slice
}

// UploadSessionHandler accepts a full audio file from an external DAW / OBS Studio.
// POST /api/v1/sessions/upload (multipart/form-data)
func (h *Handler) UploadSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// 500MB max memory parsing, larger files stream to temp file
	if err := r.ParseMultipartForm(500 << 20); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to parse multipart form: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, `{"error":"'audio' file field is required"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	title := r.FormValue("title")
	campaign := r.FormValue("campaign")
	dm := r.FormValue("dm")
	notes := r.FormValue("notes")

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".wav"
	}

	sess := &session.AudioSession{
		Title:       title,
		Campaign:    campaign,
		DM:          dm,
		Notes:       notes,
		Format:      "audio/" + strings.TrimPrefix(ext, "."),
		InputDevice: "External DAW / OBS Studio Upload",
		Status:      "completed",
		StartTime:   time.Now(),
	}
	sess.EnsureDefaults()

	if err := h.sessions.CreateSession(sess); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to initialize session: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	fullDiskPath := h.sessions.ResolveDiskPath(sess.FilePath)
	destFile, err := os.Create(fullDiskPath)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to create session file: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	written, err := io.Copy(destFile, file)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to write audio data: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	sess.SizeBytes = written
	if ext == ".wav" {
		_ = session.UpdateWavHeaderSizes(fullDiskPath)
		if f, err := os.Open(fullDiskPath); err == nil {
			if info, err := session.ParseWavHeader(f, written); err == nil {
				sess.DurationSec = info.DurationSec
				sess.SampleRate = info.SampleRate
				sess.Channels = info.Channels
				sess.BitDepth = info.BitsPerSample
			}
			f.Close()
		}
	}

	finalSess, err := h.sessions.CompleteSession(sess.ID)
	if err != nil {
		finalSess = sess
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(finalSess)
}

// DeleteSessionHandler removes a session and its audio file.
// DELETE /api/v1/sessions/{id}
func (h *Handler) DeleteSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/sessions/")
	if id == "" {
		http.Error(w, `{"error":"session id is required"}`, http.StatusBadRequest)
		return
	}

	if err := h.sessions.DeleteSession(id); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to delete session: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"deleted": id,
		"status":  "success",
	})
}
