package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
	"github.com/echosh-labs/mercury-dasha/internal/foundations"
)

func TestFoundationsStorySeedEndpoints(t *testing.T) {
	h, mux, _, cleanup := setupTestRecorderHandler(t)
	defer cleanup()

	// Seed mock profile with timeline
	mockTimeline := []dasha.DashaPeriod{
		{
			Level:        dasha.LevelMahadasha,
			Planet:       dasha.PlanetSaturn,
			PlanetName:   "Saturn",
			SanskritName: "Shani",
			SacredMetal:  "Lead (Plumbum)",
			FrequencyHz:  147.85,
			StartDate:    time.Date(2007, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:      time.Date(2028, 1, 1, 0, 0, 0, 0, time.UTC),
			HermeticAxiom: "The Principle of Cause and Effect",
		},
	}
	mockProf := dasha.DashaProfile{
		ID:       "profile:sovereign-genesis",
		Timeline: mockTimeline,
	}
	profBytes, _ := json.Marshal(mockProf)
	_ = h.store.SaveProfile("profile:sovereign-genesis", profBytes)

	// Create test note on disk
	textDir := filepath.Join(h.cfg.DropboxLocalPath, "text", "blessings")
	_ = os.MkdirAll(textDir, 0755)

	content := `Blessing of Isaac:
May the Lord grant thee wisdom and steady discernment in all thy undertakings.
"Walk in the incorruptible light of integrity, and let no obstacle turn thee from the great work."
Honor the covenant of thy ancestors and maintain the sacred flame.`

	notePath := filepath.Join(textDir, "Blessing of Isaac.txt")
	_ = os.WriteFile(notePath, []byte(content), 0644)

	// 1. Test POST /api/v1/foundations/seeds/from-text
	reqBody, _ := json.Marshal(foundations.SeedFromTextRequest{
		Path:     "text/blessings/Blessing of Isaac.txt",
		NoteDate: "2017-06-15",
		Save:     true,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/foundations/seeds/from-text", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var seed foundations.StorySeed
	if err := json.Unmarshal(w.Body.Bytes(), &seed); err != nil {
		t.Fatalf("failed to decode response seed: %v", err)
	}

	if seed.ID == "" {
		t.Errorf("expected non-empty seed ID")
	}
	if seed.SanctuaryID != "blessings" {
		t.Errorf("expected sanctuary 'blessings', got %q", seed.SanctuaryID)
	}
	if len(seed.SceneBeats) != 3 {
		t.Errorf("expected 3 scene beats, got %d", len(seed.SceneBeats))
	}
	if seed.DialogueAnchor != "Walk in the incorruptible light of integrity, and let no obstacle turn thee from the great work." {
		t.Errorf("unexpected dialogue anchor: %q", seed.DialogueAnchor)
	}

	// 2. Test GET /api/v1/foundations/seeds (list)
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/foundations/seeds", nil)
	listW := httptest.NewRecorder()
	mux.ServeHTTP(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200 from list seeds, got %d", listW.Code)
	}

	var listResp struct {
		Total int                     `json:"total"`
		Seeds []foundations.StorySeed `json:"seeds"`
	}
	if err := json.Unmarshal(listW.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to decode list response: %v", err)
	}
	if listResp.Total != 1 {
		t.Errorf("expected 1 saved seed, got %d", listResp.Total)
	}

	// 3. Test GET /api/v1/foundations/seeds/{id}
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/foundations/seeds/"+seed.ID, nil)
	getW := httptest.NewRecorder()
	mux.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200 from get seed, got %d", getW.Code)
	}

	// 4. Test POST /api/v1/foundations/seeds/{id}/manifest
	manReq := httptest.NewRequest(http.MethodPost, "/api/v1/foundations/seeds/"+seed.ID+"/manifest?duration_sec=45", nil)
	manW := httptest.NewRecorder()
	mux.ServeHTTP(manW, manReq)

	if manW.Code != http.StatusOK {
		t.Fatalf("expected 200 from manifest endpoint, got %d: %s", manW.Code, manW.Body.String())
	}

	var manifestResp struct {
		ID          string  `json:"id"`
		DurationSec float64 `json:"duration_sec"`
		Scenes      []any   `json:"scenes"`
	}
	if err := json.Unmarshal(manW.Body.Bytes(), &manifestResp); err != nil {
		t.Fatalf("failed to decode manifest response: %v", err)
	}
	if manifestResp.DurationSec != 45 {
		t.Errorf("expected 45s duration, got %v", manifestResp.DurationSec)
	}

	// 5. Test GET /api/v1/foundations/seeds/resonant
	resReq := httptest.NewRequest(http.MethodGet, "/api/v1/foundations/seeds/resonant?limit=5", nil)
	resW := httptest.NewRecorder()
	mux.ServeHTTP(resW, resReq)

	if resW.Code != http.StatusOK {
		t.Fatalf("expected 200 from resonant endpoint, got %d", resW.Code)
	}

	// 6. Test DELETE /api/v1/foundations/seeds/{id}
	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/foundations/seeds/"+seed.ID, nil)
	delW := httptest.NewRecorder()
	mux.ServeHTTP(delW, delReq)

	if delW.Code != http.StatusOK {
		t.Fatalf("expected 200 from delete seed, got %d", delW.Code)
	}
}
