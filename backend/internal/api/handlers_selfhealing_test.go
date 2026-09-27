package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/echosh-labs/mercury-dasha/internal/config"
	"github.com/echosh-labs/mercury-dasha/internal/db"
	"github.com/echosh-labs/mercury-dasha/internal/selfhealing"
)

func setupSelfHealingTestRouter(t *testing.T, restartToken string) (*Router, *Handler, *db.Store, func()) {
	tmpDir, err := os.MkdirTemp("", "api-selfhealing-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := db.Open(dbPath)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("failed to open boltdb: %v", err)
	}

	cfg := &config.Config{
		ServiceName:         "mercury-dasha-test",
		Environment:         "test",
		Port:                "18888",
		BoltDBPath:          dbPath,
		RestartToken:        restartToken,
		WatchdogIntervalSec: 60,
		MemoryLimitMB:       1024,
	}

	handler := NewHandler(cfg, store)
	router := NewRouter(handler)

	cleanup := func() {
		handler.Close()
		_ = store.Close()
		_ = os.RemoveAll(tmpDir)
	}

	return router, handler, store, cleanup
}

func TestSystemHealthEndpoint(t *testing.T) {
	router, _, _, cleanup := setupSelfHealingTestRouter(t, "")
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var report selfhealing.SystemHealthReport
	if err := json.Unmarshal(rec.Body.Bytes(), &report); err != nil {
		t.Fatalf("failed to unmarshal health report: %v", err)
	}

	if report.OverallStatus != selfhealing.StatusHealthy {
		t.Errorf("expected HEALTHY, got %s", report.OverallStatus)
	}

	if _, ok := report.Components["database"]; !ok {
		t.Errorf("expected database component in report")
	}
	if _, ok := report.Components["memory"]; !ok {
		t.Errorf("expected memory component in report")
	}
}

func TestSelfHealingCheckAndRemediate(t *testing.T) {
	router, _, _, cleanup := setupSelfHealingTestRouter(t, "")
	defer cleanup()

	// 1. On-demand check
	reqCheck := httptest.NewRequest(http.MethodPost, "/api/v1/system/self-healing/check", nil)
	recCheck := httptest.NewRecorder()
	router.ServeHTTP(recCheck, reqCheck)

	if recCheck.Code != http.StatusOK {
		t.Fatalf("check endpoint failed with code %d: %s", recCheck.Code, recCheck.Body.String())
	}

	// 2. Targeted remediation
	payload := `{"action":"memory_gc"}`
	reqRem := httptest.NewRequest(http.MethodPost, "/api/v1/system/self-healing/remediate", bytes.NewBufferString(payload))
	reqRem.Header.Set("Content-Type", "application/json")
	recRem := httptest.NewRecorder()
	router.ServeHTTP(recRem, reqRem)

	if recRem.Code != http.StatusOK {
		t.Fatalf("remediate endpoint failed with code %d: %s", recRem.Code, recRem.Body.String())
	}

	var remResp selfhealing.RemediateResponse
	if err := json.Unmarshal(recRem.Body.Bytes(), &remResp); err != nil {
		t.Fatalf("failed to unmarshal remResp: %v", err)
	}
	if !remResp.Success {
		t.Errorf("expected remediation success, got %v", remResp.Message)
	}

	// 3. Incidents inquiry
	reqInc := httptest.NewRequest(http.MethodGet, "/api/v1/system/self-healing/incidents?limit=10", nil)
	recInc := httptest.NewRecorder()
	router.ServeHTTP(recInc, reqInc)

	if recInc.Code != http.StatusOK {
		t.Fatalf("incidents query failed with code %d: %s", recInc.Code, recInc.Body.String())
	}

	var incData map[string]any
	if err := json.Unmarshal(recInc.Body.Bytes(), &incData); err != nil {
		t.Fatalf("failed to decode incidents: %v", err)
	}
	if total, ok := incData["total"].(float64); !ok || total == 0 {
		t.Errorf("expected at least 1 incident recorded, got %v", incData["total"])
	}
}

func TestRestartEndpointAuthAndExecution(t *testing.T) {
	secret := "secure-admin-key-xyz"
	router, handler, _, cleanup := setupSelfHealingTestRouter(t, secret)
	defer cleanup()

	// 1. Unauthorized request (missing token)
	reqNoAuth := httptest.NewRequest(http.MethodPost, "/api/v1/system/restart", bytes.NewBufferString(`{"reason":"test"}`))
	recNoAuth := httptest.NewRecorder()
	router.ServeHTTP(recNoAuth, reqNoAuth)

	if recNoAuth.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", recNoAuth.Code)
	}

	// 2. Unauthorized request (wrong token)
	reqWrongAuth := httptest.NewRequest(http.MethodPost, "/api/v1/system/restart", bytes.NewBufferString(`{"reason":"test"}`))
	reqWrongAuth.Header.Set("X-Admin-Token", "wrong-key")
	recWrongAuth := httptest.NewRecorder()
	router.ServeHTTP(recWrongAuth, reqWrongAuth)

	if recWrongAuth.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", recWrongAuth.Code)
	}

	// 3. Authorized request
	reqAuth := httptest.NewRequest(http.MethodPost, "/api/v1/system/restart", bytes.NewBufferString(`{"reason":"Manual operator restart","delay_ms":100,"dry_run":true}`))
	reqAuth.Header.Set("X-Admin-Token", secret)
	recAuth := httptest.NewRecorder()
	router.ServeHTTP(recAuth, reqAuth)

	if recAuth.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", recAuth.Code, recAuth.Body.String())
	}

	var resp RestartAPIResponse
	if err := json.Unmarshal(recAuth.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode restart response: %v", err)
	}
	if resp.Status != "restarting" {
		t.Errorf("expected status 'restarting', got %s", resp.Status)
	}

	// 4. Second concurrent restart should return 409 Conflict
	reqDup := httptest.NewRequest(http.MethodPost, "/api/v1/system/restart", bytes.NewBufferString(`{"reason":"Duplicate"}`))
	reqDup.Header.Set("X-Admin-Token", secret)
	recDup := httptest.NewRecorder()
	router.ServeHTTP(recDup, reqDup)

	if recDup.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict, got %d", recDup.Code)
	}

	// Read from restart manager channel to clear
	select {
	case req := <-handler.RestartManager().RestartChan():
		if req.Reason != "Manual operator restart" {
			t.Errorf("unexpected restart reason: %s", req.Reason)
		}
	default:
		t.Errorf("expected item on restart channel")
	}

	// 5. Restart history
	reqHist := httptest.NewRequest(http.MethodGet, "/api/v1/system/restart/history", nil)
	recHist := httptest.NewRecorder()
	router.ServeHTTP(recHist, reqHist)

	if recHist.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on history, got %d", recHist.Code)
	}
}

func TestPanicRecoveryMiddlewareIntegration(t *testing.T) {
	router, handler, _, cleanup := setupSelfHealingTestRouter(t, "")
	defer cleanup()

	// Register temporary panic route on router mux
	router.mux.HandleFunc("GET /api/v1/test/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("deliberate test crash")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test/panic", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 Internal Server Error, got %d", rec.Code)
	}

	// Verify panic was recorded in self-healing incident journal
	incidents := handler.SelfHealing().GetIncidents(5)
	if len(incidents) == 0 {
		t.Fatalf("expected incident logged for panic")
	}

	found := false
	for _, inc := range incidents {
		if inc.Component == "http_panic" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected http_panic incident to be recorded")
	}
}
