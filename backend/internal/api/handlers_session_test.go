package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/echosh-labs/mercury-dasha/internal/session"
)

func TestSessionLifecycleAndSlicing(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// 1. Create Session
	createPayload := []byte(`{
		"id": "session-test-dnd-1",
		"title": "Curse of Strahd - Castle Ravenloft",
		"campaign": "Curse of Strahd",
		"dm": "Justin",
		"format": "audio/wav",
		"sample_rate": 48000,
		"channels": 2,
		"bit_depth": 16,
		"input_device": "Steinberg UR-44 (Line 1/2)"
	}`)

	req := httptest.NewRequest("POST", "/api/v1/sessions", bytes.NewReader(createPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var createdSess session.AudioSession
	if err := json.NewDecoder(w.Body).Decode(&createdSess); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if createdSess.ID != "session-test-dnd-1" {
		t.Errorf("expected session ID session-test-dnd-1, got %s", createdSess.ID)
	}

	// 2. Append Audio Chunks (3 seconds of 48kHz 16-bit stereo PCM = 48000 * 4 * 3 = 576,000 bytes)
	audioChunk := make([]byte, 576000)
	chunkReq := httptest.NewRequest("POST", "/api/v1/sessions/session-test-dnd-1/chunk", bytes.NewReader(audioChunk))
	chunkReq.Header.Set("Content-Type", "application/octet-stream")
	chunkW := httptest.NewRecorder()
	router.ServeHTTP(chunkW, chunkReq)

	if chunkW.Code != http.StatusOK {
		t.Fatalf("expected chunk append status 200, got %d: %s", chunkW.Code, chunkW.Body.String())
	}

	// 3. Add Live Encounter Marker
	markerPayload := []byte(`{
		"timestamp_ms": 1500,
		"label": "Combat: Strahd Appears",
		"category": "combat",
		"notes": "Strahd stepped out of the shadows"
	}`)
	markerReq := httptest.NewRequest("POST", "/api/v1/sessions/session-test-dnd-1/markers", bytes.NewReader(markerPayload))
	markerReq.Header.Set("Content-Type", "application/json")
	markerW := httptest.NewRecorder()
	router.ServeHTTP(markerW, markerReq)

	if markerW.Code != http.StatusOK {
		t.Fatalf("expected marker status 200, got %d: %s", markerW.Code, markerW.Body.String())
	}

	// 4. Complete Session
	completeReq := httptest.NewRequest("POST", "/api/v1/sessions/session-test-dnd-1/complete", nil)
	completeW := httptest.NewRecorder()
	router.ServeHTTP(completeW, completeReq)

	if completeW.Code != http.StatusOK {
		t.Fatalf("expected complete status 200, got %d: %s", completeW.Code, completeW.Body.String())
	}

	var completedSess session.AudioSession
	if err := json.NewDecoder(completeW.Body).Decode(&completedSess); err != nil {
		t.Fatalf("failed to decode completed session: %v", err)
	}
	if completedSess.Status != "completed" {
		t.Errorf("expected status 'completed', got %s", completedSess.Status)
	}
	if completedSess.DurationSec < 2.99 || completedSess.DurationSec > 3.01 {
		t.Errorf("expected duration ~3.0s, got %f", completedSess.DurationSec)
	}

	// 5. List Sessions
	listReq := httptest.NewRequest("GET", "/api/v1/sessions?campaign=Curse%20of%20Strahd", nil)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("expected list status 200, got %d", listW.Code)
	}

	// 6. Test HTTP Range Streaming (RFC 7233 / HTTP 206)
	streamReq := httptest.NewRequest("GET", "/api/v1/sessions/session-test-dnd-1/stream", nil)
	streamReq.Header.Set("Range", "bytes=0-1023")
	streamW := httptest.NewRecorder()
	router.ServeHTTP(streamW, streamReq)

	if streamW.Code != http.StatusPartialContent {
		t.Errorf("expected status 206 Partial Content, got %d", streamW.Code)
	}
	if streamW.Body.Len() != 1024 {
		t.Errorf("expected 1024 bytes in partial stream, got %d", streamW.Body.Len())
	}

	// 7. Test Precision Audio Slicing (1000ms to 2000ms)
	sliceReq := httptest.NewRequest("GET", "/api/v1/sessions/session-test-dnd-1/slice?start_ms=1000&end_ms=2000&label=StrahdEntry", nil)
	sliceW := httptest.NewRecorder()
	router.ServeHTTP(sliceW, sliceReq)

	if sliceW.Code != http.StatusOK {
		t.Fatalf("expected slice status 200, got %d: %s", sliceW.Code, sliceW.Body.String())
	}
	if sliceW.Header().Get("Content-Type") != "audio/wav" {
		t.Errorf("expected audio/wav content type, got %s", sliceW.Header().Get("Content-Type"))
	}

	// Sliced WAV: 1.0s of 48kHz 16-bit stereo = 192,000 bytes + 44 bytes header = 192,044 bytes
	if sliceW.Body.Len() != 192044 {
		t.Errorf("expected 192044 bytes in slice WAV, got %d", sliceW.Body.Len())
	}

	// 8. Delete Session
	delReq := httptest.NewRequest("DELETE", "/api/v1/sessions/session-test-dnd-1", nil)
	delW := httptest.NewRecorder()
	router.ServeHTTP(delW, delReq)

	if delW.Code != http.StatusOK {
		t.Errorf("expected delete status 200, got %d", delW.Code)
	}
}
