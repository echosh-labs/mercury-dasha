package session

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/echosh-labs/mercury-dasha/internal/db"
)

func TestPrecisionWavSlicing(t *testing.T) {
	channels := 2
	sampleRate := 48000
	bitsPerSample := 16
	durationSec := 5.0 // 5 seconds audio

	wavData := SynthesizeSilentWav(channels, sampleRate, bitsPerSample, durationSec)
	r := bytes.NewReader(wavData)

	// 1. Verify Header Parsing
	info, err := ParseWavHeader(r, int64(len(wavData)))
	if err != nil {
		t.Fatalf("ParseWavHeader failed: %v", err)
	}

	if info.Channels != 2 {
		t.Errorf("expected 2 channels, got %d", info.Channels)
	}
	if info.SampleRate != 48000 {
		t.Errorf("expected 48000 sample rate, got %d", info.SampleRate)
	}
	if info.BitsPerSample != 16 {
		t.Errorf("expected 16 bits per sample, got %d", info.BitsPerSample)
	}
	if info.DurationSec < 4.99 || info.DurationSec > 5.01 {
		t.Errorf("expected ~5.0s duration, got %f", info.DurationSec)
	}

	// 2. Extract Precision Slice: 1000ms to 2500ms (1.5 seconds)
	var sliceBuf bytes.Buffer
	startMs := int64(1000)
	endMs := int64(2500)

	slice, err := ExtractPrecisionSlice(r, int64(len(wavData)), startMs, endMs, &sliceBuf)
	if err != nil {
		t.Fatalf("ExtractPrecisionSlice failed: %v", err)
	}

	if slice.StartMs != startMs || slice.EndMs != endMs {
		t.Errorf("slice boundaries mismatch: got %d-%d, want %d-%d", slice.StartMs, slice.EndMs, startMs, endMs)
	}
	if slice.DurationSec < 1.49 || slice.DurationSec > 1.51 {
		t.Errorf("expected ~1.5s slice duration, got %f", slice.DurationSec)
	}

	// 3. Verify Extracted Slice Header
	sliceReader := bytes.NewReader(sliceBuf.Bytes())
	sliceInfo, err := ParseWavHeader(sliceReader, int64(sliceBuf.Len()))
	if err != nil {
		t.Fatalf("failed to parse extracted slice header: %v", err)
	}

	expectedFrames := int64(float64(sampleRate) * 1.5)
	expectedBytes := 44 + (expectedFrames * int64(channels*(bitsPerSample/8)))
	if int64(sliceBuf.Len()) != expectedBytes {
		t.Errorf("expected slice size %d bytes, got %d", expectedBytes, sliceBuf.Len())
	}

	if sliceInfo.DurationSec < 1.49 || sliceInfo.DurationSec > 1.51 {
		t.Errorf("slice parsed duration expected ~1.5s, got %f", sliceInfo.DurationSec)
	}
}

func TestAppendChunkAndHeaderUpdate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mercury-test-audio-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "test_recording.wav")

	// Write 44-byte initial header
	header := GenerateWavHeader(2, 48000, 16, 0)
	size, err := AppendChunkToFile(filePath, header)
	if err != nil {
		t.Fatalf("failed to write header: %v", err)
	}
	if size != 44 {
		t.Errorf("expected size 44, got %d", size)
	}

	// Append 1 second of stereo 16-bit 48kHz audio (48000 * 4 = 192,000 bytes)
	oneSecBytes := make([]byte, 192000)
	size, err = AppendChunkToFile(filePath, oneSecBytes)
	if err != nil {
		t.Fatalf("failed to append chunk: %v", err)
	}
	if size != 192044 {
		t.Errorf("expected total size 192044, got %d", size)
	}

	// Update header sizes
	if err := UpdateWavHeaderSizes(filePath); err != nil {
		t.Fatalf("failed to update header sizes: %v", err)
	}

	// Read and verify
	f, err := os.Open(filePath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	info, err := ParseWavHeader(f, 192044)
	if err != nil {
		t.Fatalf("failed to parse updated WAV: %v", err)
	}

	if info.DataSize != 192000 {
		t.Errorf("expected data size 192000, got %d", info.DataSize)
	}
	if info.DurationSec < 0.99 || info.DurationSec > 1.01 {
		t.Errorf("expected duration ~1.0s, got %f", info.DurationSec)
	}
}

func TestSessionStoreBoltDB(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mercury-test-session-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	store, err := db.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	audioPath := filepath.Join(tempDir, "audio")
	sessionStore := NewSessionStore(store, audioPath)

	sess := &AudioSession{
		ID:          "test-session-101",
		Title:       "Descent into Avernus - Session 1",
		Campaign:    "Avernus",
		DM:          "Justin",
		Format:      "audio/wav",
		SampleRate:  48000,
		Channels:    2,
		BitDepth:    16,
		InputDevice: "Steinberg UR-44 (Line In)",
	}

	if err := sessionStore.CreateSession(sess); err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Retrieve
	retrieved, err := sessionStore.GetSession("test-session-101")
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if retrieved.Title != sess.Title {
		t.Errorf("expected title %s, got %s", sess.Title, retrieved.Title)
	}

	// Add Marker
	marker := SessionMarker{
		TimestampMs: 5000,
		Label:       "Combat: Hellhound Ambush",
		Category:    "combat",
		Notes:       "Party rolled initiative",
	}
	updated, err := sessionStore.AddMarker("test-session-101", marker)
	if err != nil {
		t.Fatalf("AddMarker failed: %v", err)
	}
	if len(updated.Markers) != 1 {
		t.Fatalf("expected 1 marker, got %d", len(updated.Markers))
	}
	if updated.Markers[0].Label != "Combat: Hellhound Ambush" {
		t.Errorf("marker label mismatch: %s", updated.Markers[0].Label)
	}

	// List
	list, err := sessionStore.ListSessions("")
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 session in list, got %d", len(list))
	}

	// Complete Session
	completed, err := sessionStore.CompleteSession("test-session-101")
	if err != nil {
		t.Fatalf("CompleteSession failed: %v", err)
	}
	if completed.Status != "completed" {
		t.Errorf("expected status 'completed', got %s", completed.Status)
	}

	// Delete Session
	if err := sessionStore.DeleteSession("test-session-101"); err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	_, err = sessionStore.GetSession("test-session-101")
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}
