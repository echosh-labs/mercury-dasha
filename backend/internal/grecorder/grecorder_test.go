package grecorder

import (
	"testing"
	"time"
)

func TestParseTextTranscript(t *testing.T) {
	sample := `[00:00] Speaker 1: Welcome to the alchemical laboratory.
[00:05] Speaker 2: We are testing the quicksilver transmutation.
[01:15] Speaker 1: The Vimshottari Dasha planetary alignment is active.
Additional note on the second line.
[01:02:30] Speaker 3: Long session timestamp test.`

	paras := parseTextTranscript(sample)
	if len(paras) != 4 {
		t.Fatalf("expected 4 paragraphs, got %d", len(paras))
	}

	if paras[0].StartMs != 0 || paras[0].Speaker != "Speaker 1" {
		t.Errorf("unexpected para 0: %+v", paras[0])
	}

	if paras[1].StartMs != 5000 || paras[1].Speaker != "Speaker 2" {
		t.Errorf("unexpected para 1: %+v", paras[1])
	}

	if paras[2].StartMs != 75000 || paras[2].Speaker != "Speaker 1" {
		t.Errorf("unexpected para 2: %+v", paras[2])
	}
	if paras[2].Text != "The Vimshottari Dasha planetary alignment is active. Additional note on the second line." {
		t.Errorf("unexpected merged text: %s", paras[2].Text)
	}

	// 1 hr 2 min 30 sec = 3600 + 120 + 30 = 3750 sec = 3750000 ms
	if paras[3].StartMs != 3750000 || paras[3].Speaker != "Speaker 3" {
		t.Errorf("unexpected para 3: %+v", paras[3])
	}
}

func TestManifestSerialization(t *testing.T) {
	manifest := SyncManifest{
		Version:  "1.0",
		LastSync: time.Now(),
		Recordings: map[string]ManifestEntry{
			"rec-123": {
				RecordingID:    "rec-123",
				Title:          "Morning Dictation",
				AudioFile:      "2026-09-07_100123_Morning_rec-123.m4a",
				AudioSizeBytes: 1048576,
				HasTranscript:  true,
			},
		},
	}

	if len(manifest.Recordings) != 1 {
		t.Fatalf("expected 1 recording in manifest")
	}
}

func TestParseTextTranscript_GoogleRecorderFormat(t *testing.T) {
	sample := `[Speaker 1] (00:03)
Justin wood would.

[Speaker 2] (00:07)
Justin wood wood.`

	paras := parseTextTranscript(sample)
	if len(paras) != 2 {
		t.Fatalf("expected 2 paragraphs, got %d", len(paras))
	}
	if paras[0].Speaker != "Speaker 1" || paras[0].StartMs != 3000 || paras[0].Text != "Justin wood would." {
		t.Errorf("unexpected para 0: %+v", paras[0])
	}
	if paras[1].Speaker != "Speaker 2" || paras[1].StartMs != 7000 || paras[1].Text != "Justin wood wood." {
		t.Errorf("unexpected para 1: %+v", paras[1])
	}
}

