package db

import (
	"io"
	"time"
)

type IndexEntry struct {
	ID        string         `json:"id"`        // Unique key, e.g. "books/Guild Archives/Rasa Shastra.pdf"
	Category  string         `json:"category"`  // "text", "audio", "video", "code", "books"
	Path      string         `json:"path"`      // Relative path to root
	FullPath  string         `json:"full_path"` // Absolute path on disk
	FileName  string         `json:"file_name"`
	Extension string         `json:"extension"` // e.g. ".pdf", ".txt", ".mp4"
	SizeBytes int64          `json:"size_bytes"`
	ModTime   time.Time      `json:"mod_time"`
	IndexedAt time.Time      `json:"indexed_at"`
	Tags      []string       `json:"tags,omitempty"`
	Snippet   string         `json:"snippet,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type IndexFilter struct {
	Query    string `json:"query"`
	Category string `json:"category"`
	Ext      string `json:"ext"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

// StorageEngine defines the standard repository interface for Mercury Dasha data persistence.
// Implementations include BoltDB (local/embedded) and future Cloud SQL / Firestore drivers.
type StorageEngine interface {
	PutJSON(bucket []byte, key string, rawJSON []byte) error
	GetJSON(bucket []byte, key string) ([]byte, error)
	Delete(bucket []byte, key string) error
	ListKeys(bucket []byte, prefix string) ([]string, error)
	Backup(w io.Writer) error
	GetStats() (*DBStats, error)
	BatchPutIndexEntries(entries []IndexEntry) error
	GetIndexEntry(id string) (*IndexEntry, error)
	SearchIndex(filter IndexFilter) ([]IndexEntry, int, error)
	GetIndexStats() (map[string]int64, error)
	SaveProfile(id string, data []byte) error
	GetProfile(id string) ([]byte, error)
	ListProfiles() ([]string, error)
	DeleteProfile(id string) error
	SaveSettings(data []byte) error
	GetSettings() ([]byte, error)
	SaveYouTubeToken(data []byte) error
	GetYouTubeToken() ([]byte, error)
	DeleteYouTubeToken() error
	SaveYouTubeJob(id string, data []byte) error
	GetYouTubeJob(id string) ([]byte, error)
	ListYouTubeJobs(limit int) ([][]byte, error)
	DeleteYouTubeJob(id string) error
	SaveEsotericContent(key string, data []byte) error
	GetEsotericContent(key string) ([]byte, error)
	ListEsotericContent(prefix string) ([]string, error)
	DeleteEsotericContent(key string) error
	SaveNakshatras(data []byte) error
	GetNakshatras() ([]byte, error)
	SaveAlchemy(data []byte) error
	GetAlchemy() ([]byte, error)
	SaveGCloudBilling(data []byte) error
	GetGCloudBilling() ([]byte, error)
	Close() error
}

// Ensure Store implements StorageEngine at compile-time.
var _ StorageEngine = (*Store)(nil)

