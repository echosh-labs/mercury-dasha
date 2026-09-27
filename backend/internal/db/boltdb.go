package db

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	ErrNotFound      = errors.New("record not found")
	ErrInvalidJSON   = errors.New("invalid JSON payload")
	ErrEmptyKey      = errors.New("key cannot be empty")
	ErrBucketNotFound = errors.New("bucket not found")

	BucketMeta              = []byte("dasha_meta")
	BucketEvents            = []byte("dasha_events")
	BucketSystem            = []byte("system")
	BucketAmraIdempotency   = []byte("amra_idempotency")
	BucketAmraSubscriptions = []byte("amra_subscriptions")
	BucketAmraLedger        = []byte("amra_ledger")
	BucketDropboxIndex      = []byte("dropbox_index")
	BucketDropboxStats      = []byte("dropbox_stats")
	BucketProfiles          = []byte("dasha_profiles")
	BucketSessions          = []byte("dnd_sessions")
	BucketAudioSlices       = []byte("audio_slices")
	BucketYouTube           = []byte("youtube_data")
	BucketEsoteric          = []byte("esoteric_content")
	BucketAmraBilling       = []byte("amra_billing")
	BucketNakshatras        = []byte("dasha_nakshatras")
	BucketAlchemy           = []byte("dasha_alchemy")
	BucketFoundationsSeeds  = []byte("foundations_seeds")
	BucketSelfHealing       = []byte("self_healing")
)

type Store struct {
	db     *bolt.DB
	dbPath string
}

type DBStats struct {
	Path           string    `json:"path"`
	SizeBytes      int64     `json:"size_bytes"`
	KeyCount       int       `json:"key_count"`
	TxN            int64     `json:"tx_total"`
	OpenTime       time.Time `json:"open_time"`
	AllocatedPages int64     `json:"allocated_pages"`
}

var openTime = time.Now()

func Open(dbPath string) (*Store, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory %s: %w", dir, err)
	}

	opts := &bolt.Options{
		Timeout: 2 * time.Second, // Prevent hung locks if another process exited abnormally
	}

	db, err := bolt.Open(dbPath, 0600, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open boltdb at %s: %w", dbPath, err)
	}

	// Initialize default schema buckets
	err = db.Update(func(tx *bolt.Tx) error {
		for _, bucketName := range [][]byte{
			BucketMeta,
			BucketEvents,
			BucketSystem,
			BucketAmraIdempotency,
			BucketAmraSubscriptions,
			BucketAmraLedger,
			BucketDropboxIndex,
			BucketDropboxStats,
			BucketProfiles,
			BucketSessions,
			BucketAudioSlices,
			BucketYouTube,
			BucketEsoteric,
			BucketAmraBilling,
			BucketNakshatras,
			BucketAlchemy,
			BucketFoundationsSeeds,
			BucketSelfHealing,
		} {
			if _, err := tx.CreateBucketIfNotExists(bucketName); err != nil {
				return fmt.Errorf("could not create bucket %s: %w", string(bucketName), err)
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	return &Store{
		db:     db,
		dbPath: dbPath,
	}, nil
}

func (s *Store) Close() error {
	if s.db != nil {
		err := s.db.Close()
		s.db = nil
		return err
	}
	return nil
}

// PutJSON validates and persists a raw JSON document by key.
func (s *Store) PutJSON(bucket []byte, key string, rawJSON []byte) error {
	if key == "" {
		return ErrEmptyKey
	}
	if !json.Valid(rawJSON) {
		return ErrInvalidJSON
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return fmt.Errorf("failed to get/create bucket %s: %w", string(bucket), err)
		}
		return b.Put([]byte(key), rawJSON)
	})
}

// GetJSON retrieves a raw JSON document by key without deserialization overhead.
func (s *Store) GetJSON(bucket []byte, key string) ([]byte, error) {
	if key == "" {
		return nil, ErrEmptyKey
	}

	var data []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return ErrBucketNotFound
		}
		v := b.Get([]byte(key))
		if v == nil {
			return ErrNotFound
		}
		// Data is only valid inside the transaction, make a safe copy
		data = make([]byte, len(v))
		copy(data, v)
		return nil
	})
	return data, err
}

// Delete removes a key from the bucket.
func (s *Store) Delete(bucket []byte, key string) error {
	if key == "" {
		return ErrEmptyKey
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return ErrBucketNotFound
		}
		return b.Delete([]byte(key))
	})
}

// ListKeys returns all keys in a bucket matching an optional prefix.
func (s *Store) ListKeys(bucket []byte, prefix string) ([]string, error) {
	var keys []string
	prefixBytes := []byte(prefix)

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return ErrBucketNotFound
		}
		c := b.Cursor()
		for k, _ := c.Seek(prefixBytes); k != nil && bytes.HasPrefix(k, prefixBytes); k, _ = c.Next() {
			keys = append(keys, string(k))
		}
		return nil
	})
	return keys, err
}

// Backup streams an atomic snapshot of the BoltDB file to the provided writer.
func (s *Store) Backup(w io.Writer) error {
	return s.db.View(func(tx *bolt.Tx) error {
		_, err := tx.WriteTo(w)
		return err
	})
}

// GetStats returns storage usage and transaction statistics.
func (s *Store) GetStats() (*DBStats, error) {
	var sizeBytes int64
	if fi, err := os.Stat(s.dbPath); err == nil {
		sizeBytes = fi.Size()
	}

	stats := s.db.Stats()
	var totalKeys int

	err := s.db.View(func(tx *bolt.Tx) error {
		for _, bName := range [][]byte{BucketMeta, BucketEvents, BucketSystem} {
			b := tx.Bucket(bName)
			if b != nil {
				totalKeys += b.Stats().KeyN
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &DBStats{
		Path:           s.dbPath,
		SizeBytes:      sizeBytes,
		KeyCount:       totalKeys,
		TxN:            int64(stats.TxN),
		OpenTime:       openTime,
		AllocatedPages: stats.TxStats.PageAlloc,
	}, nil
}

// BatchPutIndexEntries stores multiple IndexEntry documents in a single BoltDB transaction.
func (s *Store) BatchPutIndexEntries(entries []IndexEntry) error {
	if len(entries) == 0 {
		return nil
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketDropboxIndex)
		if b == nil {
			return ErrBucketNotFound
		}

		for _, entry := range entries {
			if entry.ID == "" {
				continue
			}
			raw, err := json.Marshal(entry)
			if err != nil {
				return fmt.Errorf("failed to marshal index entry %s: %w", entry.ID, err)
			}
			if err := b.Put([]byte(entry.ID), raw); err != nil {
				return err
			}
		}

		// Invalidate cached category stats so next query regenerates fresh counts
		if bStats := tx.Bucket(BucketDropboxStats); bStats != nil {
			_ = bStats.Delete([]byte("categories"))
		}

		return nil
	})
}

// GetIndexEntry retrieves an index entry by ID.
func (s *Store) GetIndexEntry(id string) (*IndexEntry, error) {
	if id == "" {
		return nil, ErrEmptyKey
	}

	var entry IndexEntry
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketDropboxIndex)
		if b == nil {
			return ErrBucketNotFound
		}
		v := b.Get([]byte(id))
		if v == nil {
			return ErrNotFound
		}
		return json.Unmarshal(v, &entry)
	})
	if err != nil {
		return nil, err
	}

	return &entry, nil
}

// SearchIndex filters indexed entries by query, category, and extension with pagination.
func (s *Store) SearchIndex(filter IndexFilter) ([]IndexEntry, int, error) {
	q := strings.ToLower(strings.TrimSpace(filter.Query))
	cat := strings.ToLower(strings.TrimSpace(filter.Category))
	ext := strings.ToLower(strings.TrimSpace(filter.Ext))
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	} else if limit > 100000 {
		limit = 100000
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	var matches []IndexEntry
	totalCount := 0

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketDropboxIndex)
		if b == nil {
			return ErrBucketNotFound
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var item IndexEntry
			if err := json.Unmarshal(v, &item); err != nil {
				continue
			}

			// Filter by category
			if cat != "" && !strings.EqualFold(item.Category, cat) {
				continue
			}

			// Filter by extension
			if ext != "" && !strings.EqualFold(item.Extension, ext) {
				continue
			}

			// Filter by text query in name, path, or tags
			if q != "" {
				nameMatch := strings.Contains(strings.ToLower(item.FileName), q)
				pathMatch := strings.Contains(strings.ToLower(item.Path), q)
				tagMatch := false
				for _, t := range item.Tags {
					if strings.Contains(strings.ToLower(t), q) {
						tagMatch = true
						break
					}
				}
				if !nameMatch && !pathMatch && !tagMatch {
					continue
				}
			}

			if totalCount >= offset && len(matches) < limit {
				matches = append(matches, item)
			}
			totalCount++
		}
		return nil
	})

	if err != nil {
		return nil, 0, err
	}

	if matches == nil {
		matches = []IndexEntry{}
	}

	return matches, totalCount, nil
}

// GetIndexStats returns counts of indexed files per category with an O(1) cached fast path.
func (s *Store) GetIndexStats() (map[string]int64, error) {
	// 1. Fast path: check if cached stats exist in BucketDropboxStats
	var cachedData []byte
	_ = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketDropboxStats)
		if b != nil {
			v := b.Get([]byte("categories"))
			if v != nil {
				cachedData = make([]byte, len(v))
				copy(cachedData, v)
			}
		}
		return nil
	})

	if len(cachedData) > 0 {
		var stats map[string]int64
		if err := json.Unmarshal(cachedData, &stats); err == nil && stats["total"] > 0 {
			return stats, nil
		}
	}

	// 2. Slow path: cursor scan across index, compute counts, and cache
	stats := map[string]int64{
		"text":  0,
		"audio": 0,
		"video": 0,
		"code":  0,
		"books": 0,
		"total": 0,
	}

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketDropboxIndex)
		if b == nil {
			return ErrBucketNotFound
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var item IndexEntry
			if err := json.Unmarshal(v, &item); err != nil {
				continue
			}
			stats["total"]++
			cLower := strings.ToLower(item.Category)
			if _, exists := stats[cLower]; exists {
				stats[cLower]++
			} else {
				stats[cLower] = 1
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Persist computed stats in BucketDropboxStats for subsequent O(1) lookups
	if raw, err := json.Marshal(stats); err == nil {
		_ = s.PutJSON(BucketDropboxStats, "categories", raw)
	}

	return stats, nil
}

func (s *Store) SaveProfile(id string, data []byte) error {
	if strings.TrimSpace(id) == "" {
		return ErrEmptyKey
	}
	return s.PutJSON(BucketProfiles, id, data)
}

func (s *Store) GetProfile(id string) ([]byte, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrEmptyKey
	}
	return s.GetJSON(BucketProfiles, id)
}

func (s *Store) ListProfiles() ([]string, error) {
	return s.ListKeys(BucketProfiles, "")
}

func (s *Store) DeleteProfile(id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrEmptyKey
	}
	return s.Delete(BucketProfiles, id)
}

const KeySystemSettings = "settings:system"

// SaveSettings persists system settings into the system bucket.
func (s *Store) SaveSettings(data []byte) error {
	return s.PutJSON(BucketSystem, KeySystemSettings, data)
}

// GetSettings retrieves persisted system settings from the system bucket.
func (s *Store) GetSettings() ([]byte, error) {
	return s.GetJSON(BucketSystem, KeySystemSettings)
}

const KeyYouTubeToken = "auth:token"

// SaveYouTubeToken stores the serialized OAuth2 token.
func (s *Store) SaveYouTubeToken(data []byte) error {
	return s.PutJSON(BucketYouTube, KeyYouTubeToken, data)
}

// GetYouTubeToken retrieves the serialized OAuth2 token.
func (s *Store) GetYouTubeToken() ([]byte, error) {
	return s.GetJSON(BucketYouTube, KeyYouTubeToken)
}

// DeleteYouTubeToken removes the stored OAuth2 token.
func (s *Store) DeleteYouTubeToken() error {
	return s.Delete(BucketYouTube, KeyYouTubeToken)
}

// SaveYouTubeJob saves an upload job record.
func (s *Store) SaveYouTubeJob(id string, data []byte) error {
	if strings.TrimSpace(id) == "" {
		return ErrEmptyKey
	}
	return s.PutJSON(BucketYouTube, "job:"+id, data)
}

// GetYouTubeJob retrieves an upload job record by ID.
func (s *Store) GetYouTubeJob(id string) ([]byte, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrEmptyKey
	}
	return s.GetJSON(BucketYouTube, "job:"+id)
}

// ListYouTubeJobs lists all upload job records up to limit.
func (s *Store) ListYouTubeJobs(limit int) ([][]byte, error) {
	if limit <= 0 {
		limit = 50
	}
	var jobs [][]byte
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketYouTube)
		if b == nil {
			return ErrBucketNotFound
		}
		c := b.Cursor()
		prefix := []byte("job:")
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			item := make([]byte, len(v))
			copy(item, v)
			jobs = append(jobs, item)
			if len(jobs) >= limit {
				break
			}
		}
		return nil
	})
	return jobs, err
}

// DeleteYouTubeJob removes a job record by ID.
func (s *Store) DeleteYouTubeJob(id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrEmptyKey
	}
	return s.Delete(BucketYouTube, "job:"+id)
}

// SaveEsotericContent stores an esoteric text or philosophical document.
func (s *Store) SaveEsotericContent(key string, data []byte) error {
	if strings.TrimSpace(key) == "" {
		return ErrEmptyKey
	}
	return s.PutJSON(BucketEsoteric, key, data)
}

// GetEsotericContent retrieves an esoteric document by key.
func (s *Store) GetEsotericContent(key string) ([]byte, error) {
	if strings.TrimSpace(key) == "" {
		return nil, ErrEmptyKey
	}
	return s.GetJSON(BucketEsoteric, key)
}

// ListEsotericContent lists all keys matching prefix in esoteric_content bucket.
func (s *Store) ListEsotericContent(prefix string) ([]string, error) {
	return s.ListKeys(BucketEsoteric, prefix)
}

// DeleteEsotericContent deletes an esoteric document by key.
func (s *Store) DeleteEsotericContent(key string) error {
	if strings.TrimSpace(key) == "" {
		return ErrEmptyKey
	}
	return s.Delete(BucketEsoteric, key)
}

// SaveGCloudBilling persists a GCloud billing snapshot or configuration.
func (s *Store) SaveGCloudBilling(data []byte) error {
	return s.PutJSON(BucketAmraBilling, "latest", data)
}

// GetGCloudBilling retrieves the latest GCloud billing snapshot.
func (s *Store) GetGCloudBilling() ([]byte, error) {
	return s.GetJSON(BucketAmraBilling, "latest")
}

// SaveNakshatras persists the authoritative 27 Nakshatras catalog to BoltDB.
func (s *Store) SaveNakshatras(data []byte) error {
	return s.PutJSON(BucketNakshatras, "catalog", data)
}

// GetNakshatras retrieves the authoritative 27 Nakshatras catalog from BoltDB.
func (s *Store) GetNakshatras() ([]byte, error) {
	return s.GetJSON(BucketNakshatras, "catalog")
}

// SaveAlchemy persists the alchemical knowledge base (metals, axioms, stages) to BoltDB.
func (s *Store) SaveAlchemy(data []byte) error {
	return s.PutJSON(BucketAlchemy, "catalog", data)
}

// GetAlchemy retrieves the alchemical knowledge base from BoltDB.
func (s *Store) GetAlchemy() ([]byte, error) {
	return s.GetJSON(BucketAlchemy, "catalog")
}
