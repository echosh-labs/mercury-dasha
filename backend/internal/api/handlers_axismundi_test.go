package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/echosh-labs/mercury-dasha/internal/axismundi"
	"github.com/echosh-labs/mercury-dasha/internal/config"
	"github.com/echosh-labs/mercury-dasha/internal/db"
)

func setupTestAxisMundiHandler(t *testing.T) (*Handler, *Router, func()) {
	tmpDir, err := os.MkdirTemp("", "mercury-axismundi-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := db.Open(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to open test db: %v", err)
	}

	cfg := &config.Config{
		ServiceName:  "mercury-dasha-test",
		Environment:  "test",
		BoltDBPath:   dbPath,
		AxisMundiURL: "http://127.0.0.1:58088", // Non-existent port for test isolation
	}

	h := NewHandler(cfg, store)
	router := NewRouter(h)

	cleanup := func() {
		h.Close()
		_ = store.Close()
		_ = os.RemoveAll(tmpDir)
	}

	return h, router, cleanup
}

func TestAxisMundiStatusHandler(t *testing.T) {
	_, router, cleanup := setupTestAxisMundiHandler(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/axis-mundi/status", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
	}

	var status axismundi.WorkspaceStatus
	if err := json.NewDecoder(rr.Body).Decode(&status); err != nil {
		t.Fatalf("Failed to decode status response: %v", err)
	}

	if status.Counts.KeepNotes != 0 {
		t.Errorf("Expected 0 initial keep notes, got %d", status.Counts.KeepNotes)
	}
}

func TestAxisMundiEventHandlerAndFeed(t *testing.T) {
	_, router, cleanup := setupTestAxisMundiHandler(t)
	defer cleanup()

	// 1. Ingest a Keep note via POST /api/v1/axis-mundi/events
	keepItem := axismundi.WorkspaceItem{
		ID:      "notes/test-keep-1",
		Type:    axismundi.TypeKeep,
		Title:   "Laboratory Opus Stage 1: Nigredo",
		Snippet: "Blackening and prima materia decomposition",
	}
	keepBody, _ := json.Marshal(keepItem)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/axis-mundi/events", bytes.NewReader(keepBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
	}

	var ingestResp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&ingestResp); err != nil {
		t.Fatalf("Failed to decode ingest response: %v", err)
	}
	if ingestResp["is_new"] != true {
		t.Errorf("Expected is_new: true, got %v", ingestResp["is_new"])
	}

	// 2. Ingest a Google Doc
	docItem := axismundi.WorkspaceItem{
		ID:      "docs/test-doc-1",
		Type:    axismundi.TypeDoc,
		Title:   "Principles of Hermetic Philosophy",
		Snippet: "Mental Transmutation treatise",
	}
	docBody, _ := json.Marshal(docItem)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/axis-mundi/events", bytes.NewReader(docBody))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on doc ingest, got %d", rr2.Code)
	}

	// 3. Query Feed via GET /api/v1/axis-mundi/feed
	feedReq := httptest.NewRequest(http.MethodGet, "/api/v1/axis-mundi/feed", nil)
	feedRR := httptest.NewRecorder()
	router.ServeHTTP(feedRR, feedReq)

	if feedRR.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for feed, got %d", feedRR.Code)
	}

	var feed axismundi.WorkspaceFeed
	if err := json.NewDecoder(feedRR.Body).Decode(&feed); err != nil {
		t.Fatalf("Failed to decode feed: %v", err)
	}

	if feed.Total != 2 {
		t.Errorf("Expected feed total 2, got %d", feed.Total)
	}
	if feed.Status.Counts.KeepNotes != 1 {
		t.Errorf("Expected 1 Keep note, got %d", feed.Status.Counts.KeepNotes)
	}
	if feed.Status.Counts.Docs != 1 {
		t.Errorf("Expected 1 Doc, got %d", feed.Status.Counts.Docs)
	}

	// 4. Test type filter: ?type=keep
	feedKeepReq := httptest.NewRequest(http.MethodGet, "/api/v1/axis-mundi/feed?type=keep", nil)
	feedKeepRR := httptest.NewRecorder()
	router.ServeHTTP(feedKeepRR, feedKeepReq)

	var keepFeed axismundi.WorkspaceFeed
	_ = json.NewDecoder(feedKeepRR.Body).Decode(&keepFeed)
	if keepFeed.Total != 1 {
		t.Errorf("Expected 1 filtered Keep note, got %d", keepFeed.Total)
	}
	if len(keepFeed.Items) > 0 && keepFeed.Items[0].Type != axismundi.TypeKeep {
		t.Errorf("Expected item type keep, got %s", keepFeed.Items[0].Type)
	}
}
