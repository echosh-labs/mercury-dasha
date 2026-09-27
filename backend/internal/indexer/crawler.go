package indexer

import (
	"errors"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
	"github.com/echosh-labs/mercury-dasha/internal/db"
)

var defaultIgnoredDirs = map[string]bool{
	".git":           true,
	"node_modules":   true,
	"venv":           true,
	".venv":          true,
	"vendor":         true,
	"dist":           true,
	"build":          true,
	"target":         true,
	".next":          true,
	".cache":         true,
	"__pycache__":    true,
	".dropbox.cache": true,
	".pnpm-store":    true,
	".idea":          true,
	".vscode":        true,
}

type CrawlStatus struct {
	IsScanning      bool             `json:"is_scanning"`
	CurrentCategory string           `json:"current_category,omitempty"`
	StartedAt       *time.Time       `json:"started_at,omitempty"`
	FinishedAt      *time.Time       `json:"finished_at,omitempty"`
	TotalSeen       int64            `json:"total_seen"`
	TotalIndexed    int64            `json:"total_indexed"`
	ErrorsCount     int64            `json:"errors_count"`
	CategoryCounts  map[string]int64 `json:"category_counts"`
}

type Crawler struct {
	dropboxRoot string
	store       db.StorageEngine
	mu          sync.RWMutex
	isScanning  bool
	currentCat  string
	startedAt   time.Time
	finishedAt  time.Time
	totalSeen   int64
	totalIdx      int64
	errorsCount   int64
	dashaTimeline []dasha.DashaPeriod
}

func NewCrawler(dropboxRoot string, store db.StorageEngine) *Crawler {
	return &Crawler{
		dropboxRoot:   strings.TrimRight(dropboxRoot, "/"),
		store:         store,
		dashaTimeline: LoadDefaultTimeline(store),
	}
}

func (c *Crawler) GetStatus() CrawlStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var startedPtr, finishedPtr *time.Time
	if !c.startedAt.IsZero() {
		t := c.startedAt
		startedPtr = &t
	}
	if !c.finishedAt.IsZero() {
		t := c.finishedAt
		finishedPtr = &t
	}

	counts, _ := c.store.GetIndexStats()

	return CrawlStatus{
		IsScanning:      c.isScanning,
		CurrentCategory: c.currentCat,
		StartedAt:       startedPtr,
		FinishedAt:      finishedPtr,
		TotalSeen:       c.totalSeen,
		TotalIndexed:    c.totalIdx,
		ErrorsCount:     c.errorsCount,
		CategoryCounts:  counts,
	}
}

// StartScan launches an asynchronous indexing crawl across the given categories.
func (c *Crawler) StartScan(categories []string) (bool, error) {
	c.mu.Lock()
	if c.isScanning {
		c.mu.Unlock()
		return false, errors.New("scan already in progress")
	}

	if len(categories) == 0 {
		categories = []string{"text", "photos", "audio", "video", "code", "books", "media"}
	}

	c.isScanning = true
	c.startedAt = time.Now()
	c.finishedAt = time.Time{}
	c.totalSeen = 0
	c.totalIdx = 0
	c.errorsCount = 0
	c.mu.Unlock()

	go func() {
		defer func() {
			c.mu.Lock()
			c.isScanning = false
			c.finishedAt = time.Now()
			c.currentCat = ""
			c.mu.Unlock()
			log.Printf("[Indexer] Crawl completed in %v. Total indexed: %d", time.Since(c.startedAt), c.totalIdx)
		}()

		for _, cat := range categories {
			c.mu.Lock()
			c.currentCat = cat
			c.mu.Unlock()

			if err := c.scanCategory(cat); err != nil {
				log.Printf("[Indexer] Error scanning category %s: %v", cat, err)
				c.mu.Lock()
				c.errorsCount++
				c.mu.Unlock()
			}
		}
	}()

	return true, nil
}

func (c *Crawler) scanCategory(category string) error {
	var subDirs []string
	switch category {
	case "text":
		subDirs = []string{"text", "Migrated Paper Docs", "Great Work"}
	case "photos", "images":
		subDirs = []string{"Pictures", "Family Room/images", "MercuryDasha/uploads", "images", "media", "media/phone"}
	case "video":
		subDirs = []string{"video", "media", "media/phone", "Pictures"}
	case "media":
		subDirs = []string{"media", "media/phone", "Pictures", "video"}
	default:
		subDirs = []string{category}
	}

	for _, sub := range subDirs {
		catDir := filepath.Join(c.dropboxRoot, sub)
		fi, err := os.Stat(catDir)
		if err != nil || !fi.IsDir() {
			continue // directory does not exist or isn't a directory
		}

		log.Printf("[Indexer] Scanning directory: %s (%s)", catDir, category)

		var batch []db.IndexEntry
		batchSize := 200

		_ = filepath.WalkDir(catDir, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				c.mu.Lock()
				c.errorsCount++
				c.mu.Unlock()
				return nil // skip on error
			}

			name := d.Name()

			// Filter out ignored directories
			if d.IsDir() {
				if defaultIgnoredDirs[name] || strings.HasPrefix(name, ".") && name != "." {
					return filepath.SkipDir
				}
				return nil
			}

			// Skip hidden files
			if strings.HasPrefix(name, ".") {
				return nil
			}

			c.mu.Lock()
			c.totalSeen++
			c.mu.Unlock()

			info, err := d.Info()
			if err != nil {
				return nil
			}

			relPath, err := filepath.Rel(c.dropboxRoot, path)
			if err != nil {
				relPath = path
			}
			relPath = filepath.ToSlash(relPath)

			ext := strings.ToLower(filepath.Ext(name))

			// If scanning photos, only index image files
			if (category == "photos" || category == "images") && !IsPhotoExtension(ext) {
				return nil
			}
			// If scanning video, only index video files
			if category == "video" && !IsVideoExtension(ext) {
				return nil
			}

			actualCat := category
			if category == "media" {
				if IsPhotoExtension(ext) {
					actualCat = "photos"
				} else if IsVideoExtension(ext) {
					actualCat = "video"
				} else {
					return nil
				}
			}

			entry := db.IndexEntry{
				ID:        relPath,
				Category:  actualCat,
				Path:      relPath,
				FullPath:  path,
				FileName:  name,
				Extension: ext,
				SizeBytes: info.Size(),
				ModTime:   info.ModTime(),
				IndexedAt: time.Now().UTC(),
			}

			// Extract text snippet for lightweight documents
			if isTextDocument(ext) && info.Size() > 0 && info.Size() < 10*1024*1024 {
				entry.Snippet = extractSnippet(path, 1024)
			}

			// Enhance with category tags and domain metadata
			if entry.Category == "text" {
				EnhanceTextEntry(&entry, c.dashaTimeline)
			} else if entry.Category == "photos" || entry.Category == "images" {
				EnhancePhotoEntry(&entry, c.dashaTimeline)
			} else {
				EnhanceEntry(&entry)
			}

			batch = append(batch, entry)
			if len(batch) >= batchSize {
				if err := c.store.BatchPutIndexEntries(batch); err != nil {
					log.Printf("[Indexer] Failed batch write: %v", err)
				} else {
					c.mu.Lock()
					c.totalIdx += int64(len(batch))
					c.mu.Unlock()
				}
				batch = batch[:0]
			}

			return nil
		})

		// Flush remaining batch
		if len(batch) > 0 {
			if err := c.store.BatchPutIndexEntries(batch); err != nil {
				log.Printf("[Indexer] Failed final batch write: %v", err)
			} else {
				c.mu.Lock()
				c.totalIdx += int64(len(batch))
				c.mu.Unlock()
			}
		}
	}

	return nil
}

func isTextDocument(ext string) bool {
	switch ext {
	case ".txt", ".md", ".json", ".yaml", ".yml", ".toml", ".sh", ".py", ".go", ".ts", ".js", ".html", ".css", ".sql":
		return true
	default:
		return false
	}
}

func extractSnippet(path string, maxBytes int64) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	buf := make([]byte, maxBytes)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return ""
	}

	// Clean snippet string
	s := string(buf[:n])
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.TrimSpace(s)
}

// IsVideoExtension returns true if the extension matches a common video format.
func IsVideoExtension(ext string) bool {
	switch strings.ToLower(ext) {
	case ".mp4", ".mov", ".m4v", ".mkv", ".avi", ".webm", ".3gp", ".wmv":
		return true
	default:
		return false
	}
}
