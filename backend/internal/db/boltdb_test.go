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

func TestBoltDB_IndexSearch(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_index.db")

	store, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	entries := []db.IndexEntry{
		{
			ID:        "books/esoteric/hermetic.pdf",
			Category:  "books",
			Path:      "books/esoteric/hermetic.pdf",
			FullPath:  "/Dropbox/books/esoteric/hermetic.pdf",
			FileName:  "hermetic.pdf",
			Extension: ".pdf",
			SizeBytes: 1048576,
			Tags:      []string{"alchemy", "philosophy"},
		},
		{
			ID:        "audio/sessions/2026-voice.wav",
			Category:  "audio",
			Path:      "audio/sessions/2026-voice.wav",
			FullPath:  "/Dropbox/audio/sessions/2026-voice.wav",
			FileName:  "2026-voice.wav",
			Extension: ".wav",
			SizeBytes: 524288,
			Tags:      []string{"session"},
		},
		{
			ID:        "text/notes/mercury.txt",
			Category:  "text",
			Path:      "text/notes/mercury.txt",
			FullPath:  "/Dropbox/text/notes/mercury.txt",
			FileName:  "mercury.txt",
			Extension: ".txt",
			SizeBytes: 1024,
			Tags:      []string{"mercury", "notes"},
			Snippet:   "Mercury Dasha planetary cycle",
		},
	}

	if err := store.BatchPutIndexEntries(entries); err != nil {
		t.Fatalf("failed to batch put entries: %v", err)
	}

	// 1. Get by ID
	e, err := store.GetIndexEntry("books/esoteric/hermetic.pdf")
	if err != nil {
		t.Fatalf("failed to get entry by ID: %v", err)
	}
	if e.FileName != "hermetic.pdf" {
		t.Fatalf("expected hermetic.pdf, got %s", e.FileName)
	}

	// 2. Search by Category
	matches, count, err := store.SearchIndex(db.IndexFilter{Category: "books"})
	if err != nil {
		t.Fatalf("failed to search by category: %v", err)
	}
	if count != 1 || len(matches) != 1 {
		t.Fatalf("expected 1 match for books, got %d (len %d)", count, len(matches))
	}

	// 3. Search by Query
	matches, count, err = store.SearchIndex(db.IndexFilter{Query: "mercury"})
	if err != nil {
		t.Fatalf("failed to search by query: %v", err)
	}
	if count != 1 || len(matches) != 1 {
		t.Fatalf("expected 1 match for mercury, got %d", count)
	}

	// 4. Stats
	stats, err := store.GetIndexStats()
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	if stats["total"] != 3 || stats["books"] != 1 || stats["audio"] != 1 || stats["text"] != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

func TestBoltDB_EsotericContent(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_esoteric.db")

	store, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	// 1. Seed esoteric content
	if err := db.SeedEsotericContent(store); err != nil {
		t.Fatalf("failed to seed esoteric content: %v", err)
	}

	// 2. Idempotency check: seeding again should succeed without error
	if err := db.SeedEsotericContent(store); err != nil {
		t.Fatalf("seeding should be idempotent: %v", err)
	}

	// 3. Verify Arishadvarga retrieval
	rawDemons, err := store.GetEsotericContent("esoteric:arishadvarga")
	if err != nil {
		t.Fatalf("failed to get esoteric:arishadvarga: %v", err)
	}
	if len(rawDemons) == 0 {
		t.Fatal("esoteric:arishadvarga should not be empty")
	}

	// 4. Verify individual demon key
	kamaRaw, err := store.GetEsotericContent("esoteric:arishadvarga:kama")
	if err != nil {
		t.Fatalf("failed to get esoteric:arishadvarga:kama: %v", err)
	}
	if len(kamaRaw) == 0 {
		t.Fatal("kama document should not be empty")
	}

	// 5. Verify Āmra philosophy
	philRaw, err := store.GetEsotericContent("esoteric:amra_philosophy")
	if err != nil {
		t.Fatalf("failed to get esoteric:amra_philosophy: %v", err)
	}
	if len(philRaw) == 0 {
		t.Fatal("esoteric:amra_philosophy should not be empty")
	}

	// 6. List esoteric content
	keys, err := store.ListEsotericContent("esoteric:")
	if err != nil {
		t.Fatalf("failed to list esoteric keys: %v", err)
	}
	if len(keys) < 4 {
		t.Fatalf("expected at least 4 esoteric keys, got %d: %v", len(keys), keys)
	}

	// 7. Delete key
	if err := store.DeleteEsotericContent("esoteric:arishadvarga:kama"); err != nil {
		t.Fatalf("failed to delete esoteric key: %v", err)
	}
	if _, err := store.GetEsotericContent("esoteric:arishadvarga:kama"); err != db.ErrNotFound {
		t.Fatalf("expected ErrNotFound after deletion, got %v", err)
	}
}

func TestBoltDB_SovereignStorehouse(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_storehouse.db")

	store, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	// 1. Test Nakshatras storehouse methods
	nakJSON := []byte(`{"count":27,"nakshatras":[{"id":"ashwini","name":"Ashwini"}]}`)
	if err := store.SaveNakshatras(nakJSON); err != nil {
		t.Fatalf("SaveNakshatras failed: %v", err)
	}
	gotNak, err := store.GetNakshatras()
	if err != nil {
		t.Fatalf("GetNakshatras failed: %v", err)
	}
	if string(gotNak) != string(nakJSON) {
		t.Fatalf("expected %s, got %s", string(nakJSON), string(gotNak))
	}

	// 2. Test Alchemy storehouse methods
	alchemyJSON := []byte(`{"philosophy":"Spagyric Transmutation","sacred_metals":[{"id":"mercury","name":"Quicksilver"}]}`)
	if err := store.SaveAlchemy(alchemyJSON); err != nil {
		t.Fatalf("SaveAlchemy failed: %v", err)
	}
	gotAlchemy, err := store.GetAlchemy()
	if err != nil {
		t.Fatalf("GetAlchemy failed: %v", err)
	}
	if string(gotAlchemy) != string(alchemyJSON) {
		t.Fatalf("expected %s, got %s", string(alchemyJSON), string(gotAlchemy))
	}

	// 3. Test Cached Index Stats
	stats1, err := store.GetIndexStats()
	if err != nil {
		t.Fatalf("GetIndexStats failed: %v", err)
	}
	if stats1["total"] != 0 {
		t.Fatalf("expected total 0, got %d", stats1["total"])
	}

	// Insert entries and test cache invalidation
	entries := []db.IndexEntry{
		{ID: "text/note1.txt", Category: "text"},
		{ID: "audio/track1.mp3", Category: "audio"},
	}
	if err := store.BatchPutIndexEntries(entries); err != nil {
		t.Fatalf("BatchPutIndexEntries failed: %v", err)
	}

	stats2, err := store.GetIndexStats()
	if err != nil {
		t.Fatalf("GetIndexStats after insert failed: %v", err)
	}
	if stats2["total"] != 2 || stats2["text"] != 1 || stats2["audio"] != 1 {
		t.Fatalf("unexpected stats: %v", stats2)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

