package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/echosh-labs/mercury-dasha/internal/api"
	"github.com/echosh-labs/mercury-dasha/internal/dasha"
)

func TestCharacterChronicleAPI(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// 1. Create Character via POST /api/v1/characters
	createPayload := map[string]any{
		"name":            "Aurelius Drake",
		"birth_date":      "1990-08-20",
		"birth_time":      "14:30",
		"timezone_offset": -4.0,
		"latitude":        43.1594,
		"longitude":       -79.2469,
	}
	body, _ := json.Marshal(createPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/characters", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d: %s", rr.Code, rr.Body.String())
	}

	var createdChar dasha.DashaProfile
	if err := json.Unmarshal(rr.Body.Bytes(), &createdChar); err != nil {
		t.Fatalf("failed to decode created character: %v", err)
	}

	// 2. Compile Chronicle via POST /api/v1/characters/{id}/chronicle
	chronicleReq := api.ChronicleRequest{
		Orientation: "16:9",
		DurationSec: 60.0,
	}
	chronicleBody, _ := json.Marshal(chronicleReq)
	cReq := httptest.NewRequest(http.MethodPost, "/api/v1/characters/"+createdChar.ID+"/chronicle", bytes.NewReader(chronicleBody))
	cReq.Header.Set("Content-Type", "application/json")
	cRr := httptest.NewRecorder()
	router.ServeHTTP(cRr, cReq)

	if cRr.Code != http.StatusCreated {
		t.Fatalf("expected status 201 for chronicle, got %d: %s", cRr.Code, cRr.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(cRr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode chronicle response: %v", err)
	}

	if resp["status"] != "created" {
		t.Errorf("expected status 'created', got %v", resp["status"])
	}
	if resp["manifest"] == nil {
		t.Errorf("expected non-nil manifest in response")
	}

	// 3. Verify Chronicle appears in GET /api/v1/characters/{id}/chronicles
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/characters/"+createdChar.ID+"/chronicles", nil)
	listRr := httptest.NewRecorder()
	router.ServeHTTP(listRr, listReq)

	if listRr.Code != http.StatusOK {
		t.Fatalf("expected status 200 for chronicles list, got %d", listRr.Code)
	}

	var listResp map[string]any
	if err := json.Unmarshal(listRr.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to decode chronicles list: %v", err)
	}
	chronicles, ok := listResp["chronicles"].([]any)
	if !ok || len(chronicles) == 0 {
		t.Errorf("expected at least 1 chronicle in character chronicles list")
	}
}
