package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/echosh-labs/mercury-dasha/internal/api"
	"github.com/echosh-labs/mercury-dasha/internal/config"
	"github.com/echosh-labs/mercury-dasha/internal/db"
)

func TestIndexContentHandler_AMRTranscoding(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer store.Close()

	// Create a dummy Dropbox directory with a test AMR file
	mockDropbox := filepath.Join(tmpDir, "Dropbox")
	audioDir := filepath.Join(mockDropbox, "audio")
	if err := os.MkdirAll(audioDir, 0755); err != nil {
		t.Fatalf("failed to create mock audio dir: %v", err)
	}

	// Copy a real AMR sample if available, or write a minimal valid AMR-NB frame
	sampleSrc := "/home/justin/Dropbox/audio/20110603-1524.amr"
	testAmrPath := filepath.Join(audioDir, "test-recording.amr")
	if data, err := os.ReadFile(sampleSrc); err == nil {
		_ = os.WriteFile(testAmrPath, data, 0644)
	} else {
		// Minimal valid AMR frame (magic + 1 SID frame)
		minAMR := []byte("#!AMR\n\x44\x00\x00\x00\x00\x00")
		_ = os.WriteFile(testAmrPath, minAMR, 0644)
	}

	cfg := &config.Config{
		DropboxLocalPath: mockDropbox,
	}

	h := api.NewHandler(cfg, store)
	router := api.NewRouter(h)

	// 1. Request transcoded AMR -> should return audio/wav
	req := httptest.NewRequest(http.MethodGet, "/api/v1/index/content?path=audio/test-recording.amr", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "audio/wav" {
		t.Errorf("expected Content-Type audio/wav, got %s", contentType)
	}

	if rec.Header().Get("X-Mercury-Transcoded") != "amr-to-wav" {
		t.Errorf("expected X-Mercury-Transcoded header")
	}

	respBytes := rec.Body.Bytes()
	if !bytes.HasPrefix(respBytes, []byte("RIFF")) {
		t.Errorf("expected RIFF header in transcoded output")
	}

	// 2. Request raw AMR -> should return original AMR
	rawReq := httptest.NewRequest(http.MethodGet, "/api/v1/index/content?path=audio/test-recording.amr&raw=true", nil)
	rawRec := httptest.NewRecorder()
	router.ServeHTTP(rawRec, rawReq)

	if rawRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for raw request, got %d", rawRec.Code)
	}
	if !bytes.HasPrefix(rawRec.Body.Bytes(), []byte("#!AMR\n")) {
		t.Errorf("expected raw AMR magic header")
	}

	// 3. Request Range seek on transcoded WAV -> should return 206 Partial Content
	rangeReq := httptest.NewRequest(http.MethodGet, "/api/v1/index/content?path=audio/test-recording.amr", nil)
	rangeReq.Header.Set("Range", "bytes=0-1023")
	rangeRec := httptest.NewRecorder()
	router.ServeHTTP(rangeRec, rangeReq)

	if rangeRec.Code != http.StatusPartialContent {
		t.Fatalf("expected 206 Partial Content for Range seek, got %d", rangeRec.Code)
	}
	if rangeRec.Header().Get("Content-Range") == "" {
		t.Errorf("expected Content-Range header")
	}
	if len(rangeRec.Body.Bytes()) != 1024 {
		t.Errorf("expected 1024 bytes, got %d", len(rangeRec.Body.Bytes()))
	}
}
