package amr_test

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"

	"github.com/echosh-labs/mercury-dasha/internal/amr"
)

func TestIsAMR(t *testing.T) {
	valid := []byte("#!AMR\n\x3c\x20")
	if !amr.IsAMR(valid) {
		t.Errorf("expected valid AMR header to be recognized")
	}

	invalid := []byte("RIFF1234WAVE")
	if amr.IsAMR(invalid) {
		t.Errorf("expected WAV header not to be recognized as AMR")
	}
}

func TestEncodeWAV(t *testing.T) {
	samples := []int16{0, 1000, -1000, 2000, -2000}
	sampleRate := 8000
	channels := 1

	wavBytes := amr.EncodeWAV(samples, sampleRate, channels)
	if len(wavBytes) != 44+len(samples)*2 {
		t.Fatalf("expected WAV size %d, got %d", 44+len(samples)*2, len(wavBytes))
	}

	if string(wavBytes[0:4]) != "RIFF" {
		t.Errorf("expected RIFF marker")
	}
	if string(wavBytes[8:12]) != "WAVE" {
		t.Errorf("expected WAVE marker")
	}
	if string(wavBytes[12:16]) != "fmt " {
		t.Errorf("expected 'fmt ' marker")
	}

	audioFormat := binary.LittleEndian.Uint16(wavBytes[20:22])
	if audioFormat != 1 {
		t.Errorf("expected PCM format 1, got %d", audioFormat)
	}

	numChannels := binary.LittleEndian.Uint16(wavBytes[22:24])
	if int(numChannels) != channels {
		t.Errorf("expected %d channels, got %d", channels, numChannels)
	}

	sr := binary.LittleEndian.Uint32(wavBytes[24:28])
	if int(sr) != sampleRate {
		t.Errorf("expected sample rate %d, got %d", sampleRate, sr)
	}

	bitsPerSample := binary.LittleEndian.Uint16(wavBytes[34:36])
	if bitsPerSample != 16 {
		t.Errorf("expected 16 bits per sample, got %d", bitsPerSample)
	}

	if string(wavBytes[36:40]) != "data" {
		t.Errorf("expected 'data' marker")
	}

	dataSize := binary.LittleEndian.Uint32(wavBytes[40:44])
	if int(dataSize) != len(samples)*2 {
		t.Errorf("expected data size %d, got %d", len(samples)*2, dataSize)
	}
}

func TestTranscodeRealAMRFile(t *testing.T) {
	// Look for one of the actual user AMR files in Dropbox
	samplePath := "/home/justin/Dropbox/audio/20110603-1524.amr"
	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		t.Skip("sample AMR file does not exist in test environment")
	}

	f, err := os.Open(samplePath)
	if err != nil {
		t.Fatalf("failed to open sample AMR file: %v", err)
	}
	defer f.Close()

	wavBytes, err := amr.TranscodeAMRToWAV(f)
	if err != nil {
		t.Fatalf("TranscodeAMRToWAV failed: %v", err)
	}

	if len(wavBytes) < 44 {
		t.Fatalf("WAV output too small: %d", len(wavBytes))
	}
	if !bytes.HasPrefix(wavBytes, []byte("RIFF")) {
		t.Errorf("missing RIFF header")
	}

	// Test caching
	tmpCache := t.TempDir()
	cachedPath, err := amr.GetOrCreateCachedWAV(tmpCache, "audio/20110603-1524.amr", samplePath)
	if err != nil {
		t.Fatalf("GetOrCreateCachedWAV failed: %v", err)
	}

	stat, err := os.Stat(cachedPath)
	if err != nil || stat.Size() == 0 {
		t.Fatalf("cached file invalid: %v", err)
	}

	// Call again to verify cache hit
	cachedPath2, err := amr.GetOrCreateCachedWAV(tmpCache, "audio/20110603-1524.amr", samplePath)
	if err != nil {
		t.Fatalf("GetOrCreateCachedWAV hit failed: %v", err)
	}
	if cachedPath != cachedPath2 {
		t.Errorf("expected same cached path, got %s and %s", cachedPath, cachedPath2)
	}
}
