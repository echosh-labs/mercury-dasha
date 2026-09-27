package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/echosh-labs/mercury-dasha/internal/config"
	"github.com/echosh-labs/mercury-dasha/internal/db"
	"github.com/echosh-labs/mercury-dasha/internal/youtube"
)

func setupTestRouter(t *testing.T, ytClientID, ytSecret string) (*Router, db.StorageEngine, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "mercury-api-yt-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := db.Open(dbPath)
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		t.Fatalf("failed to open test db: %v", err)
	}

	cfg := &config.Config{
		Port:                "8080",
		BoltDBPath:          dbPath,
		ServiceName:         "mercury-dasha-test",
		Environment:         "test",
		YouTubeClientID:     ytClientID,
		YouTubeClientSecret: ytSecret,
		YouTubeRedirectURL:  "http://localhost:8080/api/v1/youtube/auth/callback",
	}

	handler := NewHandler(cfg, store)
	router := NewRouter(handler)

	cleanup := func() {
		handler.Close()
		_ = store.Close()
		_ = os.RemoveAll(tmpDir)
	}

	return router, store, cleanup
}

func TestYouTubeEndpoints_StatusAndAuth(t *testing.T) {
	// 1. Unconfigured credentials
	router, _, cleanup := setupTestRouter(t, "", "")
	defer cleanup()

	// GET /api/v1/youtube/status
	req := httptest.NewRequest("GET", "/api/v1/youtube/status", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var status youtube.YouTubeStatusResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &status); err != nil {
		t.Fatalf("failed to unmarshal status: %v", err)
	}
	if status.Configured || status.Authenticated {
		t.Errorf("expected configured=false, authenticated=false; got %+v", status)
	}

	// GET /api/v1/youtube/auth/url -> should fail when unconfigured
	req = httptest.NewRequest("GET", "/api/v1/youtube/auth/url", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for missing credentials, got %d", rr.Code)
	}

	// 2. Configured credentials
	routerConfigured, _, cleanupConfigured := setupTestRouter(t, "mock-client-id.apps.googleusercontent.com", "mock-client-secret")
	defer cleanupConfigured()

	req = httptest.NewRequest("GET", "/api/v1/youtube/auth/url", nil)
	rr = httptest.NewRecorder()
	routerConfigured.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (body: %s)", rr.Code, rr.Body.String())
	}

	var authResp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &authResp); err != nil {
		t.Fatalf("failed to unmarshal auth url resp: %v", err)
	}
	if !strings.Contains(authResp["auth_url"], "accounts.google.com") {
		t.Errorf("expected google oauth url, got %s", authResp["auth_url"])
	}
	if authResp["state"] == "" {
		t.Errorf("expected non-empty state")
	}
}

func TestYouTubeEndpoints_UploadUnauthenticated(t *testing.T) {
	router, _, cleanup := setupTestRouter(t, "mock-id", "mock-secret")
	defer cleanup()

	reqBody := `{"file_path":"/tmp/fake.mp4","title":"My Title"}`
	req := httptest.NewRequest("POST", "/api/v1/youtube/upload", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got %d", rr.Code)
	}
}

func TestYouTubeEndpoints_JobsListAndDisconnect(t *testing.T) {
	router, _, cleanup := setupTestRouter(t, "mock-id", "mock-secret")
	defer cleanup()

	// GET /api/v1/youtube/jobs
	req := httptest.NewRequest("GET", "/api/v1/youtube/jobs", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var jobsResp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &jobsResp); err != nil {
		t.Fatalf("failed to unmarshal jobs resp: %v", err)
	}
	if jobsResp["total"].(float64) != 0 {
		t.Errorf("expected total 0 jobs, got %v", jobsResp["total"])
	}

	// POST /api/v1/youtube/auth/disconnect
	req = httptest.NewRequest("POST", "/api/v1/youtube/auth/disconnect", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}

func TestYouTubeEndpoints_AnalyticsAndFinanceUnauth(t *testing.T) {
	router, _, cleanup := setupTestRouter(t, "mock-id", "mock-secret")
	defer cleanup()

	// GET /api/v1/youtube/analytics -> 401 when not authenticated
	req := httptest.NewRequest("GET", "/api/v1/youtube/analytics", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for unauth analytics, got %d", rr.Code)
	}

	// GET /api/v1/youtube/finance -> 401 when not authenticated
	req = httptest.NewRequest("GET", "/api/v1/youtube/finance", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for unauth finance, got %d", rr.Code)
	}

	// POST /api/v1/youtube/finance/sync-amra -> 401 when not authenticated
	req = httptest.NewRequest("POST", "/api/v1/youtube/finance/sync-amra", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for unauth sync-amra, got %d", rr.Code)
	}

	// GET /api/v1/amra/metrics -> 200 OK
	req = httptest.NewRequest("GET", "/api/v1/amra/metrics", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for amra metrics, got %d", rr.Code)
	}

	// GET /api/v1/amra/plans -> 200 OK
	req = httptest.NewRequest("GET", "/api/v1/amra/plans", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for amra plans, got %d", rr.Code)
	}
}

