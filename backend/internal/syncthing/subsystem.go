package syncthing

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
	"github.com/echosh-labs/mercury-dasha/internal/db"
	"github.com/echosh-labs/mercury-dasha/internal/indexer"
)

// Subsystem orchestrates the embedded Syncthing integration for Mercury Dasha.
type Subsystem struct {
	client       *Client
	processor    *MediaProcessor
	mediaRoot    string
	thumbnailDir string
	store        db.StorageEngine

	mu          sync.RWMutex
	lastEventID int64
	folderPaths map[string]string // folderID -> localPath
	isSyncing   bool
	devices     []DeviceConfig
	folders     []FolderConfig

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewSubsystem creates an embedded Syncthing subsystem.
func NewSubsystem(baseURL, apiKey, mediaRoot string, store db.StorageEngine, timeline []dasha.DashaPeriod) *Subsystem {
	if mediaRoot == "" {
		mediaRoot = "/home/justin/media"
	}
	thumbnailDir := filepath.Join(mediaRoot, ".thumbnails")
	_ = os.MkdirAll(thumbnailDir, 0755)

	client := NewClient(baseURL, apiKey)
	processor := NewMediaProcessor(store, timeline, mediaRoot, thumbnailDir)

	return &Subsystem{
		client:       client,
		processor:    processor,
		mediaRoot:    mediaRoot,
		thumbnailDir: thumbnailDir,
		store:        store,
		folderPaths:  make(map[string]string),
	}
}

// Start launches the background Syncthing Event Bus listener and vault backfill.
func (s *Subsystem) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	s.wg.Add(1)
	go s.runEventLoop(ctx)

	s.wg.Add(1)
	go s.initialBackfill(ctx)
}

// Stop cleanly terminates the Syncthing subsystem.
func (s *Subsystem) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
	log.Println("[Syncthing] Subsystem stopped.")
}

// Client returns the underlying Syncthing REST client.
func (s *Subsystem) Client() *Client {
	return s.client
}

// Processor returns the media pipeline processor.
func (s *Subsystem) Processor() *MediaProcessor {
	return s.processor
}

// UpdateTimeline updates the Dasha timeline reference.
func (s *Subsystem) UpdateTimeline(timeline []dasha.DashaPeriod) {
	s.processor.UpdateTimeline(timeline)
}

// runEventLoop maintains a persistent long-polling connection to Syncthing's /rest/events.
func (s *Subsystem) runEventLoop(ctx context.Context) {
	defer s.wg.Done()
	log.Println("[Syncthing] Reactive Event Bus listener started.")

	// Initial discovery of folder mappings
	s.refreshFolderMap(ctx)

	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		s.mu.RLock()
		since := s.lastEventID
		s.mu.RUnlock()

		events, err := s.client.GetEvents(ctx, since, 100, 30)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			// Connection refused or Syncthing down; retry with exponential backoff
			time.Sleep(backoff)
			if backoff < 15*time.Second {
				backoff *= 2
			}
			continue
		}
		backoff = time.Second // reset backoff on success

		for _, evt := range events {
			if evt.ID > since {
				since = evt.ID
			}
			s.handleEvent(ctx, evt)
		}

		s.mu.Lock()
		s.lastEventID = since
		s.mu.Unlock()
	}
}

func (s *Subsystem) handleEvent(ctx context.Context, evt Event) {
	switch evt.Type {
	case "ItemFinished":
		var data ItemFinishedData
		if err := json.Unmarshal(evt.Data, &data); err != nil {
			return
		}
		if data.Error != "" || data.Type != "file" || data.Action != "update" {
			return
		}

		// Resolve physical file path
		s.mu.RLock()
		basePath, ok := s.folderPaths[data.Folder]
		s.mu.RUnlock()

		if !ok || basePath == "" {
			s.refreshFolderMap(ctx)
			s.mu.RLock()
			basePath = s.folderPaths[data.Folder]
			s.mu.RUnlock()
		}

		if basePath == "" {
			basePath = s.mediaRoot
		}

		fullPath := filepath.Join(basePath, data.Item)
		log.Printf("[Syncthing] Event ItemFinished: %s (folder: %s)", data.Item, data.Folder)

		// Process and index immediately in background
		go func(fPath, fID string) {
			pCtx, pCancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer pCancel()
			if item, err := s.processor.ProcessMediaFile(pCtx, fPath, fID); err == nil && item != nil {
				log.Printf("[Syncthing] Autonomously ingested & indexed: %s [%s, Dasha: %s]", item.FileName, item.Category, item.DashaMahadasha)
			}
		}(fullPath, data.Folder)

	case "StateChanged":
		var data StateChangedData
		if err := json.Unmarshal(evt.Data, &data); err == nil {
			s.mu.Lock()
			s.isSyncing = (data.To == "syncing")
			s.mu.Unlock()
		}

	case "FolderCompletion", "DeviceConnected", "DeviceDisconnected":
		// Refresh telemetry cache
		go s.refreshStatus(context.Background())
	}
}

func (s *Subsystem) refreshFolderMap(ctx context.Context) {
	folders, err := s.client.GetFolders(ctx)
	if err != nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.folders = folders
	for _, f := range folders {
		path := f.Path
		if strings.HasPrefix(path, "~/") {
			home, _ := os.UserHomeDir()
			path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
		s.folderPaths[f.ID] = path
	}
}

func (s *Subsystem) refreshStatus(ctx context.Context) {
	s.refreshFolderMap(ctx)
	devices, err := s.client.GetDevices(ctx)
	if err == nil {
		s.mu.Lock()
		s.devices = devices
		s.mu.Unlock()
	}
}

// walkFollowingSymlinks traverses root and follows symlinked directories up to maxDepth without looping.
func walkFollowingSymlinks(root string, fn func(path string, info os.FileInfo) error) error {
	visited := make(map[string]bool)

	var walk func(dir string, depth int) error
	walk = func(dir string, depth int) error {
		if depth > 5 {
			return nil
		}
		realDir, err := filepath.EvalSymlinks(dir)
		if err != nil {
			return nil
		}
		if visited[realDir] {
			return nil
		}
		visited[realDir] = true

		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil
		}

		for _, entry := range entries {
			name := entry.Name()
			if strings.HasPrefix(name, ".") {
				continue
			}
			entryPath := filepath.Join(dir, name)
			fi, err := os.Stat(entryPath)
			if err != nil {
				continue
			}
			if fi.IsDir() {
				if err := walk(entryPath, depth+1); err != nil {
					return err
				}
			} else {
				if err := fn(entryPath, fi); err != nil {
					return err
				}
			}
		}
		return nil
	}

	return walk(root, 0)
}

// initialBackfill crawls /home/justin/media following symlinks to backfill existing items into BoltDB.
func (s *Subsystem) initialBackfill(ctx context.Context) {
	defer s.wg.Done()

	// Brief pause to let server startup settle
	select {
	case <-ctx.Done():
		return
	case <-time.After(2 * time.Second):
	}

	log.Printf("[Syncthing] Starting asynchronous vault scan in %s...", s.mediaRoot)

	type fileItem struct {
		path    string
		modTime time.Time
	}
	var mediaFiles []fileItem

	_ = walkFollowingSymlinks(s.mediaRoot, func(path string, info os.FileInfo) error {
		ext := strings.ToLower(filepath.Ext(info.Name()))
		if indexer.IsPhotoExtension(ext) || indexer.IsVideoExtension(ext) {
			mediaFiles = append(mediaFiles, fileItem{path: path, modTime: info.ModTime()})
		}
		return nil
	})

	// Sort newest first
	sort.Slice(mediaFiles, func(i, j int) bool {
		return mediaFiles[i].modTime.After(mediaFiles[j].modTime)
	})

	log.Printf("[Syncthing] Discovered %d media files in vault to index.", len(mediaFiles))

	count := 0
	for _, f := range mediaFiles {
		if ctx.Err() != nil {
			break
		}
		if _, err := s.processor.ProcessMediaFile(ctx, f.path, "vault"); err == nil {
			count++
		}
	}

	log.Printf("[Syncthing] Initial vault scan completed. %d media items active in catalog.", count)
}

// GetStatus returns the unified health and telemetry payload.
func (s *Subsystem) GetStatus(ctx context.Context) (*StatusResponse, error) {
	s.refreshStatus(ctx)

	live, version, err := s.client.Ping(ctx)
	if err != nil {
		live = false
	}

	s.mu.RLock()
	devices := s.devices
	folders := s.folders
	isSyncing := s.isSyncing
	lastID := s.lastEventID
	s.mu.RUnlock()

	// Calculate total vault files and size following symlinks
	var totalFiles int
	var totalBytes int64
	_ = walkFollowingSymlinks(s.mediaRoot, func(path string, info os.FileInfo) error {
		totalFiles++
		totalBytes += info.Size()
		return nil
	})

	recent := s.processor.GetRecentFeed(20)

	return &StatusResponse{
		SyncthingLive:   live,
		SyncthingURL:    s.client.baseURL,
		Version:         version,
		MediaRoot:       s.mediaRoot,
		TotalVaultFiles: totalFiles,
		TotalVaultBytes: totalBytes,
		Devices:         devices,
		Folders:         folders,
		LastEventID:     lastID,
		IsSyncing:       isSyncing,
		RecentItems:     recent,
	}, nil
}

// GetFeed returns up to limit recent items from memory feed.
func (s *Subsystem) GetFeed(limit int) []RecentMediaItem {
	return s.processor.GetRecentFeed(limit)
}

// TriggerRescan calls Syncthing to rescan a specific folder or all media folders.
func (s *Subsystem) TriggerRescan(ctx context.Context, folderID string) error {
	return s.client.RescanFolder(ctx, folderID)
}

// StreamMedia serves high-speed video/image files with HTTP 206 byte-range seeking across mediaRoot and Dropbox vaults.
func (s *Subsystem) StreamMedia(w http.ResponseWriter, r *http.Request, relPath string) {
	relPath = strings.TrimPrefix(filepath.Clean(relPath), string(filepath.Separator))
	if strings.Contains(relPath, "..") {
		http.Error(w, "invalid path traversal", http.StatusBadRequest)
		return
	}

	// 1. Check in mediaRoot (e.g. /home/justin/media/pixel7/...)
	fullPath := filepath.Join(s.mediaRoot, relPath)
	fi, err := os.Stat(fullPath)
	if err != nil || fi.IsDir() {
		// 2. Check in Dropbox root (e.g. /home/justin/Dropbox/Pictures/... or /home/justin/Dropbox/video/...)
		dropboxRoot := "/home/justin/Dropbox"
		dbPath := filepath.Join(dropboxRoot, relPath)
		if fi2, err2 := os.Stat(dbPath); err2 == nil && !fi2.IsDir() {
			fullPath = dbPath
			fi = fi2
		} else {
			http.Error(w, "media not found", http.StatusNotFound)
			return
		}
	}

	f, err := os.Open(fullPath)
	if err != nil {
		http.Error(w, "could not open media", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(fullPath))
	switch ext {
	case ".mp4":
		w.Header().Set("Content-Type", "video/mp4")
	case ".mov":
		w.Header().Set("Content-Type", "video/quicktime")
	case ".webm":
		w.Header().Set("Content-Type", "video/webm")
	case ".avi":
		w.Header().Set("Content-Type", "video/vnd.avi")
	case ".mkv":
		w.Header().Set("Content-Type", "video/x-matroska")
	case ".m4v":
		w.Header().Set("Content-Type", "video/x-m4v")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".webp":
		w.Header().Set("Content-Type", "image/webp")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	w.Header().Set("Accept-Ranges", "bytes")
	http.ServeContent(w, r, fi.Name(), fi.ModTime(), f)
}

// ServeThumbnail serves cached poster frame images, photo previews, or generates on-demand video thumbnails.
func (s *Subsystem) ServeThumbnail(w http.ResponseWriter, r *http.Request, thumbNameOrPath string) {
	if strings.Contains(thumbNameOrPath, "..") {
		http.Error(w, "invalid path traversal", http.StatusBadRequest)
		return
	}
	thumbNameOrPath = strings.TrimPrefix(filepath.Clean(thumbNameOrPath), string(filepath.Separator))

	// 1. Check if it exists directly in thumbnailDir (e.g. thumb_015e1d7f2932e647.jpg)
	baseName := filepath.Base(thumbNameOrPath)
	cachedPath := filepath.Join(s.thumbnailDir, baseName)
	if fi, err := os.Stat(cachedPath); err == nil && !fi.IsDir() {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		http.ServeFile(w, r, cachedPath)
		return
	}

	// 2. Check if it's a photo in mediaRoot or Dropbox
	ext := strings.ToLower(filepath.Ext(thumbNameOrPath))
	if indexer.IsPhotoExtension(ext) {
		fullPath := filepath.Join(s.mediaRoot, thumbNameOrPath)
		if fi, err := os.Stat(fullPath); err == nil && !fi.IsDir() {
			w.Header().Set("Cache-Control", "public, max-age=86400")
			http.ServeFile(w, r, fullPath)
			return
		}
		dropboxPath := filepath.Join("/home/justin/Dropbox", thumbNameOrPath)
		if fi, err := os.Stat(dropboxPath); err == nil && !fi.IsDir() {
			w.Header().Set("Cache-Control", "public, max-age=86400")
			http.ServeFile(w, r, dropboxPath)
			return
		}
	}

	// 3. If it's a video, generate thumbnail on demand if possible
	if indexer.IsVideoExtension(ext) {
		fullPath := filepath.Join(s.mediaRoot, thumbNameOrPath)
		if _, err := os.Stat(fullPath); err != nil {
			fullPath = filepath.Join("/home/justin/Dropbox", thumbNameOrPath)
		}
		if _, err := os.Stat(fullPath); err == nil {
			thumbFile := s.processor.generateVideoThumbnail(r.Context(), fullPath, filepath.Base(fullPath))
			if thumbFile != "" {
				genPath := filepath.Join(s.thumbnailDir, thumbFile)
				if _, err := os.Stat(genPath); err == nil {
					w.Header().Set("Content-Type", "image/jpeg")
					w.Header().Set("Cache-Control", "public, max-age=86400")
					http.ServeFile(w, r, genPath)
					return
				}
			}
		}
	}

	http.Error(w, "thumbnail not found", http.StatusNotFound)
}
