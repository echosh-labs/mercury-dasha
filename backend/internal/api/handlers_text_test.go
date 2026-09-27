package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
	"github.com/echosh-labs/mercury-dasha/internal/db"
)

func TestTextCatalogEndpoints(t *testing.T) {
	h, mux, recordingsDir, cleanup := setupTestRecorderHandler(t)
	defer cleanup()

	// Seed a mock profile with timeline so Dasha timeline endpoints have data
	mockTimeline := []dasha.DashaPeriod{
		{
			Level:        dasha.LevelMahadasha,
			Planet:       dasha.PlanetSaturn,
			PlanetName:   "Saturn",
			SanskritName: "Shani",
			StartDate:    time.Date(2007, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	mockProf := dasha.DashaProfile{
		ID:       "profile:sovereign-genesis",
		Timeline: mockTimeline,
	}
	profBytes, _ := json.Marshal(mockProf)
	_ = h.store.SaveProfile("profile:sovereign-genesis", profBytes)

	// Seed a written personal note into the test store
	textDir := filepath.Join(h.cfg.DropboxLocalPath, "text")
	_ = os.MkdirAll(textDir, 0755)

	noteContent := "The Kybalion teaches that everything is dual; everything has poles; everything has its pair of opposites."
	noteFile := filepath.Join(textDir, "2019-04-19.txt")
	_ = os.WriteFile(noteFile, []byte(noteContent), 0644)

	dummyTextEntry := db.IndexEntry{
		ID:        "text/2019-04-19.txt",
		Category:  "text",
		Path:      "text/2019-04-19.txt",
		FullPath:  noteFile,
		FileName:  "2019-04-19.txt",
		Extension: ".txt",
		SizeBytes: int64(len(noteContent)),
		ModTime:   time.Date(2019, 4, 19, 12, 0, 0, 0, time.UTC),
		Snippet:   noteContent,
		Metadata: map[string]any{
			"sanctuary":       "dated_journals",
			"sanctuary_label": "Chronological Journals",
			"note_date":       "2019-04-19",
			"year":            2019,
			"dasha_mahadasha": "Saturn",
			"dasha_antardasha": "Sun",
			"sacred_metal":    "Lead (Plumbum)",
		},
	}
	if err := h.store.BatchPutIndexEntries([]db.IndexEntry{dummyTextEntry}); err != nil {
		t.Fatalf("Failed to save dummy text entry: %v", err)
	}

	// 1. Test GET /api/v1/text/catalog
	{
		req := httptest.NewRequest("GET", "/api/v1/text/catalog", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK for text catalog, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp TextCatalogResponse
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to decode text catalog response: %v", err)
		}

		if resp.Total == 0 {
			t.Errorf("Expected at least 1 text item, got %d", resp.Total)
		}

		// Verify both written note and spoken transcript exist
		foundWritten := false
		foundTranscript := false
		for _, item := range resp.Items {
			if item.SourceType == "written_note" {
				foundWritten = true
				if item.Sanctuary != "dated_journals" {
					t.Errorf("Expected sanctuary dated_journals, got %s", item.Sanctuary)
				}
			}
			if item.SourceType == "spoken_transcript" {
				foundTranscript = true
				if !item.HasAudio {
					t.Errorf("Expected transcript item to have HasAudio=true")
				}
			}
		}

		if !foundWritten {
			t.Errorf("Expected to find written_note in text catalog items")
		}
		if !foundTranscript {
			t.Errorf("Expected to find spoken_transcript in text catalog items")
		}
	}

	// 2. Test GET /api/v1/text/document
	{
		req := httptest.NewRequest("GET", "/api/v1/text/document?path=text/2019-04-19.txt", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK for text document, got %d: %s", rr.Code, rr.Body.String())
		}

		var doc TextDocumentDetail
		if err := json.NewDecoder(rr.Body).Decode(&doc); err != nil {
			t.Fatalf("Failed to decode document detail: %v", err)
		}

		if doc.Content != noteContent {
			t.Errorf("Expected content '%s', got '%s'", noteContent, doc.Content)
		}
		if doc.Sanctuary != "dated_journals" {
			t.Errorf("Expected sanctuary dated_journals, got %s", doc.Sanctuary)
		}
		if doc.WordCount <= 0 {
			t.Errorf("Expected positive word count, got %d", doc.WordCount)
		}
	}

	// 3. Test GET /api/v1/text/sanctuaries
	{
		req := httptest.NewRequest("GET", "/api/v1/text/sanctuaries", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK for text sanctuaries, got %d", rr.Code)
		}

		var sResp map[string]any
		_ = json.NewDecoder(rr.Body).Decode(&sResp)
		if total, ok := sResp["total"].(float64); !ok || total < 5 {
			t.Errorf("Expected at least 5 sanctuaries, got %v", sResp["total"])
		}
	}

	// 4. Test GET /api/v1/text/timeline
	{
		req := httptest.NewRequest("GET", "/api/v1/text/timeline", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK for text timeline, got %d: %s", rr.Code, rr.Body.String())
		}

		var tResp map[string]any
		_ = json.NewDecoder(rr.Body).Decode(&tResp)
		if eras, ok := tResp["eras"].([]any); !ok || len(eras) == 0 {
			t.Errorf("Expected timeline eras list, got %v", tResp["eras"])
		}
	}
	_ = recordingsDir
}
