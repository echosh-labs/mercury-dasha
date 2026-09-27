// DESCRIPTION: Key-value persistence engine wrapping bbolt with automatic YAML encoding and decoding.

package boltyaml

import (
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
	"gopkg.in/yaml.v3"
)

var (
	ErrBucketNotFound = errors.New("bucket not found")
	ErrKeyNotFound    = errors.New("key not found")
)

// Store provides YAML persistence operations over BoltDB.
type Store struct {
	db *bolt.DB
}

// NewStore initializes a new Store instance wrapping an open BoltDB handle.
func NewStore(db *bolt.DB) *Store {
	return &Store{db: db}
}

// Put serializes a value into YAML and writes it to the specified bucket and key.
func (s *Store) Put(bucket []byte, key []byte, value any) error {
	payload, err := yaml.Marshal(value)
	if err != nil {
		return fmt.Errorf("yaml marshal failed: %w", err)
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return fmt.Errorf("create bucket failed: %w", err)
		}
		return b.Put(key, payload)
	})
}

// Get fetches a YAML payload by bucket and key, unmarshaling it into the target destination.
func (s *Store) Get(bucket []byte, key []byte, dest any) error {
	return s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return ErrBucketNotFound
		}

		raw := b.Get(key)
		if raw == nil {
			return ErrKeyNotFound
		}

		if err := yaml.Unmarshal(raw, dest); err != nil {
			return fmt.Errorf("yaml unmarshal failed: %w", err)
		}
		return nil
	})
}

// Delete removes a key-value entry from the specified bucket.
func (s *Store) Delete(bucket []byte, key []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return ErrBucketNotFound
		}
		return b.Delete(key)
	})
}
