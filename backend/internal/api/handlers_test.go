package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/echosh-labs/mercury-dasha/internal/api"
	"github.com/echosh-labs/mercury-dasha/internal/config"
	"github.com/echosh-labs/mercury-dasha/internal/db"
)

func setupTestRouter(t *testing.T) (*api.Router, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "mercury-test-*")
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
		Port:        "8080",
		BoltDBPath:  dbPath,
		ServiceName: "mercury-dasha-test",
		Environment: "test",
	}

	handler := api.NewHandler(cfg, store)
	router := api.NewRouter(handler)

	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}

	return router, cleanup
}

func TestHealthAndTelemetry(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// Test /healthz
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /healthz, got %d", rec.Code)
	}

	var healthResp api.HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &healthResp); err != nil {
		t.Fatalf("failed to parse /healthz response: %v", err)
	}
	if healthResp.Status != "UP" {
		t.Errorf("expected status UP, got %s", healthResp.Status)
	}

	// Test /api/telemetry
	reqTel := httptest.NewRequest(http.MethodGet, "/api/telemetry", nil)
	recTel := httptest.NewRecorder()
	router.ServeHTTP(recTel, reqTel)

	if recTel.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/telemetry, got %d", recTel.Code)
	}
}

func TestDashaDomainEndpoints(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	endpoints := []string{
		"/api/dasha/overview",
		"/api/dasha/nakshatras",
		"/api/dasha/alchemy",
	}

	for _, ep := range endpoints {
		req := httptest.NewRequest(http.MethodGet, ep, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("endpoint %s returned %d, expected 200", ep, rec.Code)
		}
	}
}

func TestMetaDocumentCRUD(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	docKey := "agent:dasha:test-config"
	payload := []byte(`{"agent_id":"dasha-01","active":true}`)

	// 1. Put document
	putReq := httptest.NewRequest(http.MethodPost, "/api/v1/meta/"+docKey, bytes.NewReader(payload))
	putReq.Header.Set("Content-Type", "application/json")
	putRec := httptest.NewRecorder()
	router.ServeHTTP(putRec, putReq)

	if putRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on PUT/POST meta, got %d: %s", putRec.Code, putRec.Body.String())
	}

	// 2. Get document
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/meta/"+docKey, nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on GET meta, got %d: %s", getRec.Code, getRec.Body.String())
	}

	// 3. List keys
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/meta", nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on LIST meta, got %d", listRec.Code)
	}

	// 4. Delete document
	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/meta/"+docKey, nil)
	delRec := httptest.NewRecorder()
	router.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on DELETE meta, got %d", delRec.Code)
	}
}
