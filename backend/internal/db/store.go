package db

import (
	"io"
)

// StorageEngine defines the standard repository interface for Mercury Dasha data persistence.
// Implementations include BoltDB (local/embedded) and future Cloud SQL / Firestore drivers.
type StorageEngine interface {
	PutJSON(bucket []byte, key string, rawJSON []byte) error
	GetJSON(bucket []byte, key string) ([]byte, error)
	Delete(bucket []byte, key string) error
	ListKeys(bucket []byte, prefix string) ([]string, error)
	Backup(w io.Writer) error
	GetStats() (*DBStats, error)
	Close() error
}

// Ensure Store implements StorageEngine at compile-time.
var _ StorageEngine = (*Store)(nil)
