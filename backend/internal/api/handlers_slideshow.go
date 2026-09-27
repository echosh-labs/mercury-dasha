package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/db"
	"github.com/echosh-labs/mercury-dasha/internal/timeline"
)

// ListSlideshowsHandler returns all hermetic slideshow chronicles stored in the sovereign repository.
func (h *Handler) ListSlideshowsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	slideshows, err := timeline.ListHermeticSlideshows(h.store)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to list hermetic slideshows: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// Always ensure canonical Temple of Illumination is present in the results
	hasCanon := false
	for _, s := range slideshows {
		if s.ID == "temple-of-illumination" {
			hasCanon = true
			break
		}
	}
	if !hasCanon {
		canon := timeline.NewTempleOfIlluminationSlideshow()
		slideshows = append([]*timeline.HermeticSlideshow{canon}, slideshows...)
	}

	summaries := make([]map[string]any, 0, len(slideshows))
	for _, s := range slideshows {
		summaries = append(summaries, map[string]any{
			"id":                 s.ID,
			"title":              s.Title,
			"subtitle":           s.Subtitle,
			"tradition":          s.Tradition,
			"author":             s.Author,
			"slide_count":        len(s.Slides),
			"total_duration_sec": s.TotalDurationSec,
			"tags":               s.Tags,
			"updated_at":         s.UpdatedAt,
		})
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"slideshows": summaries,
		"count":      len(summaries),
	})
}

// GetSlideshowHandler retrieves a complete hermetic slideshow document by ID.
func (h *Handler) GetSlideshowHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.PathValue("id")
	if id == "" {
		id = r.URL.Query().Get("id")
	}
	if id == "" {
		http.Error(w, `{"error":"id parameter is required"}`, http.StatusBadRequest)
		return
	}

	// Sub-resource check: /manifest
	if strings.HasSuffix(id, "/manifest") {
		cleanID := strings.TrimSuffix(id, "/manifest")
		h.generateSlideshowManifestByID(w, r, cleanID)
		return
	}

	slideshow, err := timeline.GetHermeticSlideshow(h.store, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			// If asking for canonical temple of illumination and not yet in store, return live default
			if id == "temple-of-illumination" {
				slideshow = timeline.NewTempleOfIlluminationSlideshow()
				_ = json.NewEncoder(w).Encode(slideshow)
				return
			}
			http.Error(w, `{"error":"slideshow not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf(`{"error":"failed to retrieve slideshow: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(slideshow)
}

// SaveSlideshowHandler stores or updates a hermetic slideshow chronicle in BoltDB and indexes its assets.
func (h *Handler) SaveSlideshowHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, `{"error":"method not allowed, must be POST or PUT"}`, http.StatusMethodNotAllowed)
		return
	}

	var s timeline.HermeticSlideshow
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid JSON body: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	if s.ID == "" {
		http.Error(w, `{"error":"slideshow id is required"}`, http.StatusBadRequest)
		return
	}

	s.EnsureDefaults()

	if err := timeline.SaveHermeticSlideshow(h.store, &s); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to save slideshow: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	baseStore := h.cfg.DropboxLocalPath
	if baseStore == "" {
		baseStore = "/home/justin/Dropbox"
	}
	_ = timeline.IndexSlideshowAssets(h.store, &s, baseStore)

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":    "saved",
		"slideshow": s,
	})
}

// SlideshowManifestHandler compiles a HermeticSlideshow into an audiovisual TimelineManifest.
func (h *Handler) SlideshowManifestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed, must be POST"}`, http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		id = r.URL.Query().Get("id")
	}
	if id == "" {
		http.Error(w, `{"error":"id parameter is required"}`, http.StatusBadRequest)
		return
	}

	if strings.HasSuffix(id, "/render") {
		h.RenderSlideshowVideoHandler(w, r)
		return
	}

	cleanID := strings.TrimSuffix(id, "/manifest")
	h.generateSlideshowManifestByID(w, r, cleanID)
}

// RenderSlideshowVideoHandler handles vocal audio upload and triggers offline FFmpeg compilation.
func (h *Handler) RenderSlideshowVideoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed, must be POST"}`, http.StatusMethodNotAllowed)
		return
	}

	rawID := r.PathValue("id")
	cleanID := strings.TrimSuffix(rawID, "/render")
	if cleanID == "" {
		cleanID = "temple-of-illumination"
	}

	_ = os.MkdirAll("renders", 0755)

	// Read audio file from multipart or body
	voicePath := ""
	if err := r.ParseMultipartForm(64 << 20); err == nil {
		file, header, fErr := r.FormFile("audio")
		if fErr == nil {
			defer file.Close()
			ext := filepath.Ext(header.Filename)
			if ext == "" {
				ext = ".wav"
			}
			voicePath = filepath.Join("renders", fmt.Sprintf("voice_%s_%d%s", cleanID, time.Now().Unix(), ext))
			out, cErr := os.Create(voicePath)
			if cErr == nil {
				_, _ = io.Copy(out, file)
				_ = out.Sync()
				_ = out.Close()
			}
		}
	}

	// Output video path
	outputVideo := filepath.Join("renders", fmt.Sprintf("%s_narration.mp4", cleanID))

	// Resolve slides directory across canonical paths and naming conventions
	altCleanID := strings.ReplaceAll(cleanID, "-", "_")
	dashCleanID := strings.ReplaceAll(cleanID, "_", "-")
	candidates := []string{
		filepath.Join("backend", "storage", "slideshows", altCleanID),
		filepath.Join("backend", "storage", "slideshows", dashCleanID),
		filepath.Join("storage", "slideshows", altCleanID),
		filepath.Join("storage", "slideshows", dashCleanID),
		filepath.Join("frontend", "public", "assets", "slideshows", altCleanID),
		filepath.Join("frontend", "public", "assets", "slideshows", dashCleanID),
		filepath.Join("/home/justin/code/echosh-labs/mercury-dasha/backend/storage/slideshows", altCleanID),
		filepath.Join("..", "backend", "storage", "slideshows", altCleanID),
	}

	slidesDir := ""
	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && fi.IsDir() {
			slidesDir = c
			break
		}
	}
	if slidesDir == "" {
		slidesDir = filepath.Join("backend", "storage", "slideshows", altCleanID)
	}

	// Resolve scripts/render_slideshow.py path
	candidateScripts := []string{
		"scripts/render_slideshow.py",
		filepath.Join("..", "scripts", "render_slideshow.py"),
		"/home/justin/code/echosh-labs/mercury-dasha/scripts/render_slideshow.py",
	}
	scriptPath := ""
	for _, cs := range candidateScripts {
		if fi, err := os.Stat(cs); err == nil && !fi.IsDir() {
			scriptPath = cs
			break
		}
	}
	if scriptPath == "" {
		scriptPath = "scripts/render_slideshow.py"
	}

	args := []string{
		scriptPath,
		"--slides-dir", slidesDir,
		"--output", outputVideo,
	}
	if voicePath != "" {
		args = append(args, "--voice-audio", voicePath)
	}

	// Resolve python executable: check python3 first, then python
	pyExec := "python3"
	if _, err := exec.LookPath(pyExec); err != nil {
		if _, err := exec.LookPath("python"); err == nil {
			pyExec = "python"
		}
	}

	cmd := exec.Command(pyExec, args...)
	cmd.Env = os.Environ()
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"render failed: %s", "details":%q}`, err.Error(), string(cmdOutput)), http.StatusInternalServerError)
		return
	}

	fi, _ := os.Stat(outputVideo)
	sizeMB := float64(0)
	if fi != nil {
		sizeMB = float64(fi.Size()) / (1024 * 1024)
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":       "rendered",
		"video_url":    fmt.Sprintf("/api/v1/index/content?path=%s", filepath.ToSlash(outputVideo)),
		"download_url": fmt.Sprintf("/api/v1/index/content?path=%s&download=1", filepath.ToSlash(outputVideo)),
		"file_size_mb": sizeMB,
		"slideshow_id": cleanID,
	})
}

func (h *Handler) generateSlideshowManifestByID(w http.ResponseWriter, r *http.Request, id string) {
	slideshow, err := timeline.GetHermeticSlideshow(h.store, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) && id == "temple-of-illumination" {
			slideshow = timeline.NewTempleOfIlluminationSlideshow()
		} else {
			http.Error(w, `{"error":"slideshow not found"}`, http.StatusNotFound)
			return
		}
	}

	var opts timeline.SlideshowOptions
	opts.Orientation = timeline.OrientationLandscape16x9
	if o := r.URL.Query().Get("orientation"); o != "" {
		opts.Orientation = timeline.Orientation(o)
	}
	opts.IncludeOverlays = true
	if r.URL.Query().Get("overlays") == "false" {
		opts.IncludeOverlays = false
	}
	if vPath := r.URL.Query().Get("voice_audio_path"); vPath != "" {
		opts.VoiceAudioPath = vPath
	}
	if mPath := r.URL.Query().Get("ambient_music_path"); mPath != "" {
		opts.AmbientMusicPath = mPath
	}

	manifest, err := slideshow.GenerateTimelineManifest(opts)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to generate manifest: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(manifest)
}
