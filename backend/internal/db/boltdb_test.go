package db_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/echosh-labs/mercury-dasha/internal/db"
)

func TestBoltDB_CRUD(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	// 1. Put valid JSON
	key := "agent:dasha:001"
	validJSON := []byte(`{"id":"001","name":"Dasha Core","status":"active","tags":["mercury","engine"]}`)

	if err := store.PutJSON(db.BucketMeta, key, validJSON); err != nil {
		t.Fatalf("failed to put JSON: %v", err)
	}

	// 2. Reject invalid JSON
	invalidJSON := []byte(`{not valid json}`)
	if err := store.PutJSON(db.BucketMeta, "bad_key", invalidJSON); err != db.ErrInvalidJSON {
		t.Fatalf("expected ErrInvalidJSON, got %v", err)
	}

	// 3. Get JSON
	retrieved, err := store.GetJSON(db.BucketMeta, key)
	if err != nil {
		t.Fatalf("failed to get JSON: %v", err)
	}
	if !bytes.Equal(retrieved, validJSON) {
		t.Fatalf("data mismatch: expected %s, got %s", validJSON, retrieved)
	}

	// 4. List Keys with prefix
	keys, err := store.ListKeys(db.BucketMeta, "agent:dasha:")
	if err != nil {
		t.Fatalf("failed to list keys: %v", err)
	}
	if len(keys) != 1 || keys[0] != key {
		t.Fatalf("unexpected keys list: %v", keys)
	}

	// 5. Test Backup Snapshot
	var backupBuf bytes.Buffer
	if err := store.Backup(&backupBuf); err != nil {
		t.Fatalf("backup failed: %v", err)
	}
	if backupBuf.Len() == 0 {
		t.Fatal("backup snapshot is empty")
	}

	// 6. Delete
	if err := store.Delete(db.BucketMeta, key); err != nil {
		t.Fatalf("failed to delete: %v", err)
	}
	if _, err := store.GetJSON(db.BucketMeta, key); err != db.ErrNotFound {
		t.Fatalf("expected ErrNotFound after deletion, got %v", err)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
