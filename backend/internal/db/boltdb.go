package db

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	ErrNotFound      = errors.New("record not found")
	ErrInvalidJSON   = errors.New("invalid JSON payload")
	ErrEmptyKey      = errors.New("key cannot be empty")
	ErrBucketNotFound = errors.New("bucket not found")

	BucketMeta   = []byte("dasha_meta")
	BucketEvents = []byte("dasha_events")
	BucketSystem = []byte("system")
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
		for _, bucketName := range [][]byte{BucketMeta, BucketEvents, BucketSystem} {
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
		return s.db.Close()
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
		b := tx.Bucket(bucket)
		if b == nil {
			return ErrBucketNotFound
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
