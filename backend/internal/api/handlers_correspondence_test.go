package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
	"github.com/echosh-labs/mercury-dasha/internal/db"
)

func TestTimelineCorrespondenceHandler(t *testing.T) {
	h, mux, _, cleanup := setupTestRecorderHandler(t)
	defer cleanup()

	// Seed profile with a two-era timeline: Jupiter (1993-2009) and Saturn (2009-2028)
	mockTimeline := []dasha.DashaPeriod{
		{
			Level:        dasha.LevelMahadasha,
			Planet:       dasha.PlanetJupiter,
			PlanetName:   "Jupiter",
			SanskritName: "Guru",
			StartDate:    time.Date(1993, 4, 13, 0, 0, 0, 0, time.UTC),
			EndDate:      time.Date(2009, 4, 13, 0, 0, 0, 0, time.UTC),
		},
		{
			Level:        dasha.LevelMahadasha,
			Planet:       dasha.PlanetSaturn,
			PlanetName:   "Saturn",
			SanskritName: "Shani",
			StartDate:    time.Date(2009, 4, 13, 0, 0, 0, 0, time.UTC),
			EndDate:      time.Date(2028, 4, 13, 0, 0, 0, 0, time.UTC),
			SubPeriods: []dasha.DashaPeriod{
				{
					Level:        dasha.LevelAntardasha,
					Planet:       dasha.PlanetJupiter,
					PlanetName:   "Jupiter",
					SanskritName: "Guru",
					StartDate:    time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC),
					EndDate:      time.Date(2028, 4, 13, 0, 0, 0, 0, time.UTC),
				},
			},
		},
	}
	mockProf := dasha.DashaProfile{
		ID:       "profile:sovereign-genesis",
		Timeline: mockTimeline,
	}
	profBytes, _ := json.Marshal(mockProf)
	_ = h.store.SaveProfile("profile:sovereign-genesis", profBytes)

	// Seed test text entry in Saturn era
	_ = h.store.BatchPutIndexEntries([]db.IndexEntry{
		{
			ID:        "text/2019-04-19.txt",
			Category:  "text",
			Path:      "text/2019-04-19.txt",
			FileName:  "2019-04-19.txt",
			Extension: ".txt",
			SizeBytes: 500,
			ModTime:   time.Date(2019, 4, 19, 12, 0, 0, 0, time.UTC),
			Snippet:   "Hermetic reflections on polarity.",
			Metadata: map[string]any{
				"dasha_mahadasha": "Saturn",
				"year":            2019,
				"note_date":       "2019-04-19",
				"sanctuary":       "hermetic_kybalion",
			},
		},
		{
			ID:        "Pictures/SarahBirth/img1.jpg",
			Category:  "photos",
			Path:      "Pictures/SarahBirth/img1.jpg",
			FileName:  "img1.jpg",
			Extension: ".jpg",
			SizeBytes: 2048,
			ModTime:   time.Date(2009, 5, 1, 12, 0, 0, 0, time.UTC),
			Metadata: map[string]any{
				"dasha_mahadasha": "Saturn",
				"year":            2009,
				"album":           "SarahBirth",
				"photo_date":      "2009-05-01",
				"sacred_metal":    "Lead (Plumbum)",
			},
		},
	})

	// Test GET /api/v1/timeline/correspondence
	req := httptest.NewRequest("GET", "/api/v1/timeline/correspondence", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp TimelineCorrespondenceResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Subject != "Justin Andrew Wood" {
		t.Errorf("Expected subject Justin Andrew Wood, got %s", resp.Subject)
	}
	if len(resp.Eras) != 2 {
		t.Fatalf("Expected 2 eras, got %d", len(resp.Eras))
	}

	// Verify Saturn era has text and photo
	var saturnEra *TimelineEraCorrespondence
	for i := range resp.Eras {
		if resp.Eras[i].Mahadasha == "Saturn" {
			saturnEra = &resp.Eras[i]
			break
		}
	}

	if saturnEra == nil {
		t.Fatalf("Saturn era not found in response")
	}
	if saturnEra.TotalTexts != 1 {
		t.Errorf("Expected 1 text in Saturn era, got %d", saturnEra.TotalTexts)
	}
	if saturnEra.TotalPhotos != 1 {
		t.Errorf("Expected 1 photo in Saturn era, got %d", saturnEra.TotalPhotos)
	}
	if len(saturnEra.PhotoAlbums) != 1 || saturnEra.PhotoAlbums[0].AlbumName != "SarahBirth" {
		t.Errorf("Expected SarahBirth album, got %+v", saturnEra.PhotoAlbums)
	}
}
