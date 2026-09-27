package boltyaml_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/echosh-labs/mercury-dasha/storage/boltyaml"
	bolt "go.etcd.io/bbolt"
)

type ConfigRecord struct {
	Name    string            `yaml:"name"`
	Version string            `yaml:"version"`
	Active  bool              `yaml:"active"`
	Tags    []string          `yaml:"tags"`
	Meta    map[string]string `yaml:"meta"`
}

func setupTestDB(t *testing.T) (*bolt.DB, *boltyaml.Store) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_boltyaml.db")

	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		t.Fatalf("failed to open bolt test db: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	store := boltyaml.NewStore(db)
	return db, store
}

func TestStore_PutAndGet(t *testing.T) {
	_, store := setupTestDB(t)

	bucket := []byte("configs")
	key := []byte("app-settings")

	expected := ConfigRecord{
		Name:    "Mercury Dasha",
		Version: "1.2.0",
		Active:  true,
		Tags:    []string{"astronomy", "alchemy", "ephemeris"},
		Meta: map[string]string{
			"env":    "production",
			"author": "Justin Andrew Wood",
		},
	}

	// 1. Put
	if err := store.Put(bucket, key, expected); err != nil {
		t.Fatalf("unexpected error on Put: %v", err)
	}

	// 2. Get
	var actual ConfigRecord
	if err := store.Get(bucket, key, &actual); err != nil {
		t.Fatalf("unexpected error on Get: %v", err)
	}

	if actual.Name != expected.Name {
		t.Errorf("expected name %q, got %q", expected.Name, actual.Name)
	}
	if actual.Version != expected.Version {
		t.Errorf("expected version %q, got %q", expected.Version, actual.Version)
	}
	if actual.Active != expected.Active {
		t.Errorf("expected active %v, got %v", expected.Active, actual.Active)
	}
	if len(actual.Tags) != len(expected.Tags) {
		t.Errorf("expected %d tags, got %d", len(expected.Tags), len(actual.Tags))
	}
	if actual.Meta["author"] != expected.Meta["author"] {
		t.Errorf("expected author %q, got %q", expected.Meta["author"], actual.Meta["author"])
	}
}

func TestStore_KeyNotFound(t *testing.T) {
	_, store := setupTestDB(t)

	bucket := []byte("configs")
	key := []byte("existing-key")

	// Create bucket with one key
	if err := store.Put(bucket, key, "some-value"); err != nil {
		t.Fatalf("unexpected error on Put: %v", err)
	}

	// Query non-existent key
	var val string
	err := store.Get(bucket, []byte("missing-key"), &val)
	if !errors.Is(err, boltyaml.ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestStore_BucketNotFound(t *testing.T) {
	_, store := setupTestDB(t)

	var val string
	err := store.Get([]byte("non-existent-bucket"), []byte("some-key"), &val)
	if !errors.Is(err, boltyaml.ErrBucketNotFound) {
		t.Fatalf("expected ErrBucketNotFound, got %v", err)
	}
}

func TestStore_Delete(t *testing.T) {
	_, store := setupTestDB(t)

	bucket := []byte("configs")
	key := []byte("temp-key")

	if err := store.Put(bucket, key, "temporary-data"); err != nil {
		t.Fatalf("unexpected error on Put: %v", err)
	}

	// Delete key
	if err := store.Delete(bucket, key); err != nil {
		t.Fatalf("unexpected error on Delete: %v", err)
	}

	// Confirm it's gone
	var val string
	err := store.Get(bucket, key, &val)
	if !errors.Is(err, boltyaml.ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound after deletion, got %v", err)
	}
}
