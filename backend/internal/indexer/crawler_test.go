package indexer_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/db"
	"github.com/echosh-labs/mercury-dasha/internal/indexer"
)

func TestIndexer_BooksMetadata(t *testing.T) {
	tags, meta := indexer.ParseBookMetadata("books/Guild Archives/Rasa Shastra - The Hidden Art.pdf", "Rasa Shastra - The Hidden Art.pdf", 5000)
	if meta["author"] != "Rasa Shastra" {
		t.Errorf("expected author 'Rasa Shastra', got %v", meta["author"])
	}
	if meta["title"] != "The Hidden Art" {
		t.Errorf("expected title 'The Hidden Art', got %v", meta["title"])
	}
	if meta["collection"] != "Guild Archives" {
		t.Errorf("expected collection 'Guild Archives', got %v", meta["collection"])
	}

	foundPDF := false
	for _, tag := range tags {
		if tag == "pdf" {
			foundPDF = true
		}
	}
	if !foundPDF {
		t.Errorf("expected 'pdf' tag in tags: %v", tags)
	}
}

func TestIndexer_CrawlDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	dropboxRoot := filepath.Join(tmpDir, "Dropbox")

	// Create directories
	textDir := filepath.Join(dropboxRoot, "text")
	codeDir := filepath.Join(dropboxRoot, "code", "myproject", "node_modules", "dep")
	_ = os.MkdirAll(textDir, 0755)
	_ = os.MkdirAll(codeDir, 0755)

	// Create test files
	_ = os.WriteFile(filepath.Join(textDir, "note.txt"), []byte("Mercury Dasha notes"), 0644)
	_ = os.WriteFile(filepath.Join(dropboxRoot, "code", "myproject", "main.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(codeDir, "dep.js"), []byte("module.exports = {}"), 0644) // should be skipped!

	store, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer store.Close()

	crawler := indexer.NewCrawler(dropboxRoot, store)
	started, err := crawler.StartScan([]string{"text", "code"})
	if err != nil || !started {
		t.Fatalf("failed to start scan: %v", err)
	}

	// Wait for background scan to finish
	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		status := crawler.GetStatus()
		if !status.IsScanning && status.FinishedAt != nil {
			break
		}
	}

	// Verify text entry was indexed
	matches, count, err := store.SearchIndex(db.IndexFilter{Query: "note"})
	if err != nil {
		t.Fatalf("failed to search store: %v", err)
	}
	if count != 1 || len(matches) != 1 {
		t.Errorf("expected 1 match for note, got %d", count)
	} else if matches[0].Snippet != "Mercury Dasha notes" {
		t.Errorf("expected snippet 'Mercury Dasha notes', got %q", matches[0].Snippet)
	}

	// Verify node_modules was skipped
	matches, count, err = store.SearchIndex(db.IndexFilter{Query: "dep.js"})
	if err != nil {
		t.Fatalf("failed to search store: %v", err)
	}
	if count != 0 {
		t.Errorf("expected node_modules file dep.js to be skipped, but found %d matches", count)
	}
}
