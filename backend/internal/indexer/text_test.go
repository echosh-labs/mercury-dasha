package indexer

import (
	"testing"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
	"github.com/echosh-labs/mercury-dasha/internal/db"
)

func TestClassifySanctuary(t *testing.T) {
	tests := []struct {
		relPath      string
		fileName     string
		expectedID   string
		expectedSub  string
	}{
		{
			relPath:     "text/blessings/Sarah.txt",
			fileName:    "Sarah.txt",
			expectedID:  "blessings",
			expectedSub: "Sarah",
		},
		{
			relPath:     "text/blessings/Luke.txt",
			fileName:    "Luke.txt",
			expectedID:  "blessings",
			expectedSub: "Luke",
		},
		{
			relPath:     "text/kybalion/Axioms.txt",
			fileName:    "Axioms.txt",
			expectedID:  "hermetic_kybalion",
			expectedSub: "The Kybalion",
		},
		{
			relPath:     "text/great-work/Alchemy.txt",
			fileName:    "Alchemy.txt",
			expectedID:  "hermetic_kybalion",
			expectedSub: "The Great Work",
		},
		{
			relPath:     "text/karate/essays/Bushido.txt",
			fileName:    "Bushido.txt",
			expectedID:  "martial_discipline",
			expectedSub: "Karate Essays",
		},
		{
			relPath:     "text/dreams/flying.txt",
			fileName:    "flying.txt",
			expectedID:  "vision_dreams",
			expectedSub: "",
		},
		{
			relPath:     "text/prayer/gratitude.txt",
			fileName:    "gratitude.txt",
			expectedID:  "devotion_prayer",
			expectedSub: "Gratitude",
		},
		{
			relPath:     "text/study/American Rosae Crucis/paper.txt",
			fileName:    "paper.txt",
			expectedID:  "rosicrucian_study",
			expectedSub: "American Rosae Crucis",
		},
		{
			relPath:     "text/2013-03-15.txt",
			fileName:    "2013-03-15.txt",
			expectedID:  "dated_journals",
			expectedSub: "Journal",
		},
		{
			relPath:     "audio/recorder/2026-09-13_rec.txt",
			fileName:    "2026-09-13_rec.txt",
			expectedID:  "spoken_transcripts",
			expectedSub: "Voice Chronicle",
		},
	}

	for _, tc := range tests {
		id, _, sub := ClassifySanctuary(tc.relPath, tc.fileName)
		if id != tc.expectedID {
			t.Errorf("Path %s: expected sanctuary %s, got %s", tc.relPath, tc.expectedID, id)
		}
		if tc.expectedSub != "" && sub != tc.expectedSub {
			t.Errorf("Path %s: expected sub %s, got %s", tc.relPath, tc.expectedSub, sub)
		}
	}
}

func TestExtractDate(t *testing.T) {
	now := time.Now()

	// 1. Dash-separated date
	d1, explicit1 := ExtractDate("2013-03-15.txt", "", now)
	if !explicit1 {
		t.Errorf("Expected explicit date for 2013-03-15.txt")
	}
	if d1.Year() != 2013 || d1.Month() != 3 || d1.Day() != 15 {
		t.Errorf("Expected 2013-03-15, got %v", d1)
	}

	// 2. Compact date
	d2, explicit2 := ExtractDate("20190419.txt", "", now)
	if !explicit2 {
		t.Errorf("Expected explicit date for 20190419.txt")
	}
	if d2.Year() != 2019 || d2.Month() != 4 || d2.Day() != 19 {
		t.Errorf("Expected 2019-04-19, got %v", d2)
	}

	// 3. Date with trailing text
	d3, explicit3 := ExtractDate("2021-05-05 work futures.txt", "", now)
	if !explicit3 {
		t.Errorf("Expected explicit date for '2021-05-05 work futures.txt'")
	}
	if d3.Year() != 2021 || d3.Month() != 5 || d3.Day() != 5 {
		t.Errorf("Expected 2021-05-05, got %v", d3)
	}

	// 4. Year only
	d4, explicit4 := ExtractDate("MDP 2020.odt", "", now)
	if !explicit4 {
		t.Errorf("Expected explicit date for 'MDP 2020.odt'")
	}
	if d4.Year() != 2020 {
		t.Errorf("Expected 2020, got %v", d4)
	}

	// 5. Textual date in snippet header
	d5, explicit5 := ExtractDate("astrology.txt", "Birth Information\nOctober 1, 1971 04:40 AM\nSt. Catharines, Ontario", now)
	if !explicit5 {
		t.Errorf("Expected explicit date for October 1, 1971")
	}
	if d5.Year() != 1971 || d5.Month() != 10 || d5.Day() != 1 {
		t.Errorf("Expected 1971-10-01, got %v", d5)
	}

	// 6. Dot-separated date in snippet header
	d6, explicit6 := ExtractDate("vimshottari dasha.txt", "Vimshottari Dasa Chart\n01.10.1971", now)
	if !explicit6 {
		t.Errorf("Expected explicit date for 01.10.1971")
	}
	if d6.Year() != 1971 || d6.Month() != 10 || d6.Day() != 1 {
		t.Errorf("Expected 1971-10-01, got %v", d6)
	}

	// 7. Textual date for September 11, 2009
	d7, explicit7 := ExtractDate("longest_day.txt", "The Longest Day\nSeptember 11, 2009", now)
	if !explicit7 {
		t.Errorf("Expected explicit date for September 11, 2009")
	}
	if d7.Year() != 2009 || d7.Month() != 9 || d7.Day() != 11 {
		t.Errorf("Expected 2009-09-11, got %v", d7)
	}
}

func TestCalculateTextMetrics(t *testing.T) {
	sample := "The Lips of Wisdom are closed, except to the ears of Understanding. As above, so below."
	words, chars, readTime := CalculateTextMetrics(sample)

	if words != 16 {
		t.Errorf("Expected 16 words, got %d", words)
	}
	if chars <= 0 {
		t.Errorf("Expected positive char count, got %d", chars)
	}
	if readTime < 0.5 {
		t.Errorf("Expected read time >= 0.5 mins, got %f", readTime)
	}
}

func TestEnhanceTextEntryWithDasha(t *testing.T) {
	// Create mock timeline with a Jupiter Mahadasha and Saturn Antardasha in 2013
	start := time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	timeline := []dasha.DashaPeriod{
		{
			Level:        dasha.LevelMahadasha,
			Planet:       dasha.PlanetJupiter,
			PlanetName:   "Jupiter",
			SanskritName: "Guru",
			StartDate:    start,
			EndDate:      end,
			SubPeriods: []dasha.DashaPeriod{
				{
					Level:        dasha.LevelAntardasha,
					Planet:       dasha.PlanetSaturn,
					PlanetName:   "Saturn",
					SanskritName: "Shani",
					StartDate:    time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
					EndDate:      time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
		},
	}

	entry := db.IndexEntry{
		ID:        "text/2013-03-15.txt",
		Category:  "text",
		Path:      "text/2013-03-15.txt",
		FileName:  "2013-03-15.txt",
		Extension: ".txt",
		SizeBytes: 3500,
		ModTime:   time.Date(2013, 3, 15, 12, 0, 0, 0, time.UTC),
		Snippet:   "Reflections on discipline and the great work.",
	}

	EnhanceTextEntry(&entry, timeline)

	if entry.Metadata["sanctuary"] != "dated_journals" {
		t.Errorf("Expected sanctuary dated_journals, got %v", entry.Metadata["sanctuary"])
	}
	if entry.Metadata["dasha_mahadasha"] != "Jupiter" {
		t.Errorf("Expected Mahadasha Jupiter, got %v", entry.Metadata["dasha_mahadasha"])
	}
	if entry.Metadata["dasha_antardasha"] != "Saturn" {
		t.Errorf("Expected Antardasha Saturn, got %v", entry.Metadata["dasha_antardasha"])
	}
	if entry.Metadata["sacred_metal"] != "Tin (Stannum)" {
		t.Errorf("Expected Tin (Stannum), got %v", entry.Metadata["sacred_metal"])
	}
	if entry.Metadata["hermetic_axiom"] == "" {
		t.Errorf("Expected non-empty hermetic_axiom")
	}
	if entry.Metadata["story_archetype"] == "" {
		t.Errorf("Expected non-empty story_archetype")
	}
}
