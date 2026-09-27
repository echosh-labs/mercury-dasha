package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
)

func TestProfileCRUD_Unified(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// 1. Create profile via POST /api/v1/profiles
	createPayload := []byte(`{
		"name": "Sovereign Genesis Test",
		"birth_date": "1992-06-15",
		"birth_time": "12:00",
		"timezone_offset": -4.0,
		"latitude": 43.1594,
		"longitude": -79.2469
	}`)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/profiles", bytes.NewReader(createPayload))
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on POST /api/v1/profiles, got %d: %s", createRec.Code, createRec.Body.String())
	}

	var createdProf dasha.SovereignProfile
	if err := json.Unmarshal(createRec.Body.Bytes(), &createdProf); err != nil {
		t.Fatalf("failed to decode created profile JSON: %v", err)
	}

	profileID := createdProf.ID
	if profileID == "" {
		t.Fatalf("expected non-empty profile ID")
	}

	// Verify unified astrological matrix
	if createdProf.Astrology.JanmaNakshatra.ID == "" {
		t.Errorf("expected JanmaNakshatra in Astrology matrix")
	}
	if createdProf.Astrology.StartingLord == "" {
		t.Errorf("expected StartingLord in Astrology matrix")
	}

	// Verify unified alchemical matrix
	if createdProf.Alchemy.SacredMetal.ID == "" {
		t.Errorf("expected SacredMetal in Alchemy matrix")
	}
	if createdProf.Alchemy.GoverningAxiom.Title == "" {
		t.Errorf("expected GoverningAxiom in Alchemy matrix")
	}

	// Verify active alchemical state
	if createdProf.ActiveAlchemy.MahadashaMetal.ID == "" {
		t.Errorf("expected MahadashaMetal in ActiveAlchemy")
	}

	// Verify timeline summary
	if len(createdProf.TimelineSummary) != 9 {
		t.Errorf("expected 9 Mahadasha entries in TimelineSummary, got %d", len(createdProf.TimelineSummary))
	}

	// 2. Fetch via GET /api/v1/profiles/{id}
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/profiles/"+profileID, nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on GET /api/v1/profiles/%s, got %d", profileID, getRec.Code)
	}

	var fetchedProf dasha.SovereignProfile
	if err := json.Unmarshal(getRec.Body.Bytes(), &fetchedProf); err != nil {
		t.Fatalf("failed to decode fetched profile JSON: %v", err)
	}
	if fetchedProf.SymbioticResonance == nil {
		t.Errorf("expected SymbioticResonance on GET profile")
	}

	// 3. Test GET /api/v1/profiles/{id}/resonance sub-resource
	resReq := httptest.NewRequest(http.MethodGet, "/api/v1/profiles/"+profileID+"/resonance", nil)
	resRec := httptest.NewRecorder()
	router.ServeHTTP(resRec, resReq)

	if resRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on GET /api/v1/profiles/%s/resonance, got %d", profileID, resRec.Code)
	}

	var res dasha.SymbioticResonance
	if err := json.Unmarshal(resRec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode resonance JSON: %v", err)
	}
	if res.TransitMetal.ID == "" {
		t.Errorf("expected TransitMetal in resonance response")
	}

	// 4. Test GET /api/v1/profiles/{id}/timeline sub-resource
	tlReq := httptest.NewRequest(http.MethodGet, "/api/v1/profiles/"+profileID+"/timeline", nil)
	tlRec := httptest.NewRecorder()
	router.ServeHTTP(tlRec, tlReq)

	if tlRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on GET /api/v1/profiles/%s/timeline, got %d", profileID, tlRec.Code)
	}

	// 5. Test GET /api/v1/profiles (listing with summaries)
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/profiles", nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on GET /api/v1/profiles, got %d", listRec.Code)
	}

	var listResp map[string]any
	if err := json.Unmarshal(listRec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to decode list response: %v", err)
	}
	if listResp["count"].(float64) < 1 {
		t.Errorf("expected at least 1 profile in list count")
	}

	// 6. Test GET /api/dasha/alchemy with profile_id
	alcReq := httptest.NewRequest(http.MethodGet, "/api/dasha/alchemy?profile_id="+profileID, nil)
	alcRec := httptest.NewRecorder()
	router.ServeHTTP(alcRec, alcReq)

	if alcRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on /api/dasha/alchemy with profile_id, got %d", alcRec.Code)
	}
	var alcResp map[string]any
	if err := json.Unmarshal(alcRec.Body.Bytes(), &alcResp); err != nil {
		t.Fatalf("failed to decode alchemy response: %v", err)
	}
	if alcResp["natal_alchemy"] == nil {
		t.Errorf("expected natal_alchemy in alchemy response when profile_id is provided")
	}

	// 7. Delete profile
	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/profiles/"+profileID, nil)
	delRec := httptest.NewRecorder()
	router.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on DELETE /api/v1/profiles/%s, got %d", profileID, delRec.Code)
	}
}
