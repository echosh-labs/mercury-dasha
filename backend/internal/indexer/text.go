package indexer

import (
	"encoding/json"
	"math"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
	"github.com/echosh-labs/mercury-dasha/internal/db"
)

var (
	reDateDash    = regexp.MustCompile(`\b(19\d{2}|20\d{2})[-_](\d{1,2})[-_](\d{1,2})\b`)
	reDateCompact = regexp.MustCompile(`\b(19\d{2}|20\d{2})(\d{2})(\d{2})\b`)
	reDateSlash   = regexp.MustCompile(`\b(\d{1,2})/(\d{1,2})/(19\d{2}|20\d{2})\b`)
	reDateDot     = regexp.MustCompile(`\b(\d{1,2})\.(\d{1,2})\.(19\d{2}|20\d{2})\b`)
	reDateTextual = regexp.MustCompile(`(?i)\b(January|February|March|April|May|June|July|August|September|October|November|December|Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[a-z]*\s+(\d{1,2})(?:st|nd|rd|th)?[\s,]+(19\d{2}|20\d{2})\b`)
	reMonthYear   = regexp.MustCompile(`(?i)\b(January|February|March|April|May|June|July|August|September|October|November|December|Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[a-z]*\s+(19\d{2}|20\d{2})\b`)
	reYearOnly    = regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`)

	monthMap = map[string]time.Month{
		"jan": time.January, "feb": time.February, "mar": time.March, "apr": time.April,
		"may": time.May, "jun": time.June, "jul": time.July, "aug": time.August,
		"sep": time.September, "oct": time.October, "nov": time.November, "dec": time.December,
	}
)

// TextSanctuary represents one of the 14 thematic sanctuaries in the user's text vault.
type TextSanctuary struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
}

var AuthoritativeSanctuaries = []TextSanctuary{
	{ID: "blessings", Label: "Blessings & Invocations", Icon: "🕊️", Description: "Family and generational invocations for Sarah, Luke, Joshua, and Isaac"},
	{ID: "hermetic_kybalion", Label: "Hermeticism & Great Work", Icon: "⚡", Description: "The Kybalion, The Great Work, and The Healing Fire"},
	{ID: "martial_discipline", Label: "Martial Lore & Discipline", Icon: "🥋", Description: "Karate essays, discipline, and dojo philosophy"},
	{ID: "vision_dreams", Label: "Vision & Dream Logs", Icon: "🌙", Description: "Dream journals and subconscious exploration"},
	{ID: "devotion_prayer", Label: "Devotion & Healing", Icon: "🙏", Description: "Prayers, gratitude journals, and Chiron healing"},
	{ID: "vocation_manifestation", Label: "Vocation & Speeches", Icon: "💼", Description: "Work futures, public speeches, and mission statements"},
	{ID: "rosicrucian_study", Label: "Rosicrucian & Esoteric Study", Icon: "📖", Description: "American Rosae Crucis and esoteric papers"},
	{ID: "dated_journals", Label: "Chronological Journals", Icon: "📜", Description: "Date-anchored personal reflections"},
	{ID: "spoken_transcripts", Label: "Spoken Transcripts", Icon: "🎙️", Description: "Google Recorder & audio chronicle transcripts"},
	{ID: "foundations_lore", Label: "Foundations Lore & Musings", Icon: "🏛️", Description: "Philosophical musings, plays, and foundations lore"},
}

// ClassifySanctuary maps a file path and name to its thematic sanctuary domain.
func ClassifySanctuary(relPath, fileName string) (sanctuaryID, label, subSanctuary string) {
	norm := filepath.ToSlash(strings.ToLower(relPath))

	// 1. Spoken Transcripts
	if strings.Contains(norm, "audio/recorder") || (strings.HasSuffix(norm, ".txt") && strings.Contains(norm, "recorder")) {
		return "spoken_transcripts", "Spoken Transcripts", "Voice Chronicle"
	}

	// 2. Blessings & Generational Invocations
	if strings.Contains(norm, "/blessings") || strings.HasPrefix(norm, "blessings") || strings.Contains(norm, "text/blessings") {
		sub := ""
		for _, child := range []string{"sarah", "luke", "joshua", "isaac"} {
			if strings.Contains(norm, child) {
				sub = strings.Title(child)
				break
			}
		}
		return "blessings", "Blessings & Invocations", sub
	}

	// 3. Hermeticism & The Great Work
	if strings.Contains(norm, "kybalion") || strings.Contains(norm, "great-work") || strings.Contains(norm, "healing fire") {
		sub := ""
		if strings.Contains(norm, "kybalion") {
			sub = "The Kybalion"
		} else if strings.Contains(norm, "great-work") {
			sub = "The Great Work"
		} else if strings.Contains(norm, "healing fire") {
			sub = "The Healing Fire"
		}
		return "hermetic_kybalion", "Hermeticism & Great Work", sub
	}

	// 4. Martial Lore & Discipline
	if strings.Contains(norm, "karate") {
		sub := ""
		if strings.Contains(norm, "essays") {
			sub = "Karate Essays"
		}
		return "martial_discipline", "Martial Lore & Discipline", sub
	}

	// 5. Vision & Dream Logs
	if strings.Contains(norm, "dream") {
		return "vision_dreams", "Vision & Dream Logs", ""
	}

	// 6. Devotion & Healing
	if strings.Contains(norm, "prayer") || strings.Contains(norm, "gratitude") || strings.Contains(norm, "chiron") {
		sub := ""
		if strings.Contains(norm, "gratitude") {
			sub = "Gratitude"
		} else if strings.Contains(norm, "chiron") {
			sub = "Chiron"
		} else {
			sub = "Prayer"
		}
		return "devotion_prayer", "Devotion & Healing", sub
	}

	// 7. Rosicrucian & Esoteric Study
	if strings.Contains(norm, "american rosae crucis") || strings.Contains(norm, "study/additions") || strings.Contains(norm, "/paper") || strings.Contains(norm, "/study") {
		sub := ""
		if strings.Contains(norm, "american rosae crucis") {
			sub = "American Rosae Crucis"
		} else if strings.Contains(norm, "additions") {
			sub = "Study Additions"
		}
		return "rosicrucian_study", "Rosicrucian & Esoteric Study", sub
	}

	// 8. Vocation & Speeches
	if strings.Contains(norm, "work") || strings.Contains(norm, "speeches") || strings.Contains(norm, "vocabularypodcast") || strings.Contains(norm, "planning") {
		sub := ""
		if strings.Contains(norm, "speeches") {
			sub = "Speeches"
		} else if strings.Contains(norm, "vocabularypodcast") {
			sub = "Vocabulary Podcast"
		}
		return "vocation_manifestation", "Vocation & Speeches", sub
	}

	// 9. Chronological Journals (Dated in filename)
	if reDateDash.MatchString(fileName) || reDateCompact.MatchString(fileName) {
		return "dated_journals", "Chronological Journals", "Journal"
	}

	// 10. Default: Foundations Lore & Musings
	return "foundations_lore", "Foundations Lore & Musings", ""
}

// ExtractDate parses explicit dates from the filename or snippet, falling back to modTime.
func ExtractDate(fileName, snippet string, modTime time.Time) (time.Time, bool) {
	// 1. Check filename for YYYY-MM-DD or YYYY_MM_DD
	if m := reDateDash.FindStringSubmatch(fileName); len(m) == 4 {
		y, _ := strconv.Atoi(m[1])
		mo, _ := strconv.Atoi(m[2])
		d, _ := strconv.Atoi(m[3])
		if y >= 1900 && y <= 2100 && mo >= 1 && mo <= 12 && d >= 1 && d <= 31 {
			return time.Date(y, time.Month(mo), d, 12, 0, 0, 0, time.UTC), true
		}
	}

	// 2. Check filename for YYYYMMDD
	if m := reDateCompact.FindStringSubmatch(fileName); len(m) == 4 {
		y, _ := strconv.Atoi(m[1])
		mo, _ := strconv.Atoi(m[2])
		d, _ := strconv.Atoi(m[3])
		if y >= 1900 && y <= 2100 && mo >= 1 && mo <= 12 && d >= 1 && d <= 31 {
			return time.Date(y, time.Month(mo), d, 12, 0, 0, 0, time.UTC), true
		}
	}

	// 3. Check snippet header (up to first 35 lines or 3000 chars) for textual, slash, or ISO dates
	if snippet != "" {
		sample := snippet
		if len(sample) > 3000 {
			sample = sample[:3000]
		}
		lines := strings.Split(sample, "\n")
		checkLimit := 35
		if len(lines) < checkLimit {
			checkLimit = len(lines)
		}

		for i := 0; i < checkLimit; i++ {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				continue
			}

			// Textual: October 1, 1971 or Sept 11, 2009
			if m := reDateTextual.FindStringSubmatch(line); len(m) == 4 {
				mStr := strings.ToLower(m[1])
				if len(mStr) >= 3 {
					if mo, ok := monthMap[mStr[:3]]; ok {
						d, _ := strconv.Atoi(m[2])
						y, _ := strconv.Atoi(m[3])
						if y >= 1900 && y <= 2100 && d >= 1 && d <= 31 {
							return time.Date(y, mo, d, 12, 0, 0, 0, time.UTC), true
						}
					}
				}
			}

			// ISO: 2009-09-11
			if m := reDateDash.FindStringSubmatch(line); len(m) == 4 {
				y, _ := strconv.Atoi(m[1])
				mo, _ := strconv.Atoi(m[2])
				d, _ := strconv.Atoi(m[3])
				if y >= 1900 && y <= 2100 && mo >= 1 && mo <= 12 && d >= 1 && d <= 31 {
					return time.Date(y, time.Month(mo), d, 12, 0, 0, 0, time.UTC), true
				}
			}

			// Dot format: 01.10.1971 (Day.Month.Year)
			if m := reDateDot.FindStringSubmatch(line); len(m) == 4 {
				d, _ := strconv.Atoi(m[1])
				mo, _ := strconv.Atoi(m[2])
				y, _ := strconv.Atoi(m[3])
				if y >= 1900 && y <= 2100 && mo >= 1 && mo <= 12 && d >= 1 && d <= 31 {
					return time.Date(y, time.Month(mo), d, 12, 0, 0, 0, time.UTC), true
				}
			}

			// Slash format: 10/01/1971 (Month/Day/Year)
			if m := reDateSlash.FindStringSubmatch(line); len(m) == 4 {
				n1, _ := strconv.Atoi(m[1])
				n2, _ := strconv.Atoi(m[2])
				y, _ := strconv.Atoi(m[3])
				if y >= 1900 && y <= 2100 {
					if n1 >= 1 && n1 <= 12 && n2 >= 1 && n2 <= 31 {
						return time.Date(y, time.Month(n1), n2, 12, 0, 0, 0, time.UTC), true
					} else if n2 >= 1 && n2 <= 12 && n1 >= 1 && n1 <= 31 {
						return time.Date(y, time.Month(n2), n1, 12, 0, 0, 0, time.UTC), true
					}
				}
			}

			// Month Year: October 1971
			if m := reMonthYear.FindStringSubmatch(line); len(m) == 3 {
				mStr := strings.ToLower(m[1])
				if len(mStr) >= 3 {
					if mo, ok := monthMap[mStr[:3]]; ok {
						y, _ := strconv.Atoi(m[2])
						if y >= 1900 && y <= 2100 {
							return time.Date(y, mo, 1, 12, 0, 0, 0, time.UTC), true
						}
					}
				}
			}
		}
	}

	// 4. Check for YYYY in filename
	if m := reYearOnly.FindStringSubmatch(fileName); len(m) == 2 {
		y, _ := strconv.Atoi(m[1])
		if y >= 1900 && y <= 2088 {
			return time.Date(y, time.January, 1, 12, 0, 0, 0, time.UTC), true
		}
	}

	return modTime, false
}

// CalculateTextMetrics computes word count, character count, and estimated reading time.
func CalculateTextMetrics(text string) (wordCount, charCount int, readTimeMins float64) {
	charCount = len([]rune(text))
	words := strings.Fields(text)
	wordCount = len(words)
	readTimeMins = math.Max(0.5, math.Round((float64(wordCount)/200.0)*10)/10)
	return wordCount, charCount, readTimeMins
}

// PlanetSacredMetal maps a Vimshottari Graha to its sacred alchemical metal and Latin name.
func PlanetSacredMetal(p dasha.PlanetID) (name, latin string) {
	switch p {
	case dasha.PlanetSun:
		return "Gold", "Aurum"
	case dasha.PlanetMoon:
		return "Silver", "Argentum"
	case dasha.PlanetMars:
		return "Iron", "Ferrum"
	case dasha.PlanetMercury:
		return "Quicksilver", "Hydrargyrum"
	case dasha.PlanetJupiter:
		return "Tin", "Stannum"
	case dasha.PlanetVenus:
		return "Copper", "Cuprum"
	case dasha.PlanetSaturn:
		return "Lead", "Plumbum"
	case dasha.PlanetRahu:
		return "Electrum", "Gold-Silver"
	case dasha.PlanetKetu:
		return "Bronze", "Copper-Tin"
	default:
		return "Quicksilver", "Hydrargyrum"
	}
}

// PlanetHermeticAxiom maps a Vimshottari Graha to its corresponding Hermetic Axiom.
func PlanetHermeticAxiom(p dasha.PlanetID) string {
	switch p {
	case dasha.PlanetSun:
		return "The Principle of Mentalism (The All is Mind)"
	case dasha.PlanetMoon:
		return "The Principle of Correspondence (As Above, So Below)"
	case dasha.PlanetMercury:
		return "The Principle of Vibration (Nothing Rests; Everything Moves)"
	case dasha.PlanetVenus:
		return "The Principle of Polarity (Everything is Dual)"
	case dasha.PlanetMars:
		return "The Principle of Rhythm (The Pendulum-Swing)"
	case dasha.PlanetJupiter:
		return "The Principle of Cause and Effect (Every Cause has its Effect)"
	case dasha.PlanetSaturn:
		return "The Principle of Gender (Gender is in Everything)"
	case dasha.PlanetRahu:
		return "The Principle of Rhythm (The Cosmic Spiral)"
	case dasha.PlanetKetu:
		return "The Principle of Mentalism (The Divine Void)"
	default:
		return "The Principle of Mentalism"
	}
}

// PlanetStoryArchetype maps a Vimshottari Graha to its Foundations Story Archetype.
func PlanetStoryArchetype(p dasha.PlanetID) string {
	switch p {
	case dasha.PlanetSun:
		return "The Sovereign Architect & Radiant Solar Hero"
	case dasha.PlanetMoon:
		return "The Deep Intuitive & Lunar Weaver of Memory"
	case dasha.PlanetMars:
		return "The Fiery Champion & Indomitable Warrior"
	case dasha.PlanetMercury:
		return "The Divine Scribe & Winged Messenger of Wisdom"
	case dasha.PlanetJupiter:
		return "The Sacred Guru & Dispenser of Cosmic Grace"
	case dasha.PlanetVenus:
		return "The Supreme Artist & Alchemist of Harmony"
	case dasha.PlanetSaturn:
		return "The Stoic Hermit & Master of Chronos Time"
	case dasha.PlanetRahu:
		return "The Visionary Pioneer & Transgressor of Thresholds"
	case dasha.PlanetKetu:
		return "The Ascetic Mystic & Seeker of Liberation"
	default:
		return "The Sovereign Scribe"
	}
}

// EnhanceTextEntry performs deep semantic categorization, date extraction,
// word metric calculation, and astrological Dasha grounding on an IndexEntry.
func EnhanceTextEntry(entry *db.IndexEntry, timeline []dasha.DashaPeriod) {
	if entry.Metadata == nil {
		entry.Metadata = make(map[string]any)
	}

	sanctuaryID, sanctuaryLabel, subSanctuary := ClassifySanctuary(entry.Path, entry.FileName)
	entry.Metadata["sanctuary"] = sanctuaryID
	entry.Metadata["sanctuary_label"] = sanctuaryLabel
	if subSanctuary != "" {
		entry.Metadata["sub_sanctuary"] = subSanctuary
	}

	// Compute word counts and metrics from snippet or file
	textSample := entry.Snippet
	words, chars, readTime := CalculateTextMetrics(textSample)
	entry.Metadata["sample_word_count"] = words
	entry.Metadata["char_count"] = chars
	entry.Metadata["est_read_time_mins"] = readTime

	// Parse date
	noteDate, hasExplicitDate := ExtractDate(entry.FileName, entry.Snippet, entry.ModTime)
	entry.Metadata["note_date"] = noteDate.Format("2006-01-02")
	entry.Metadata["has_explicit_date"] = hasExplicitDate
	entry.Metadata["year"] = noteDate.Year()

	// Clean formatted display title
	ext := filepath.Ext(entry.FileName)
	base := strings.TrimSuffix(entry.FileName, ext)
	cleanTitle := strings.ReplaceAll(base, "_", " ")
	cleanTitle = strings.ReplaceAll(cleanTitle, "-", " ")
	if hasExplicitDate && (sanctuaryID == "dated_journals" || sanctuaryID == "spoken_transcripts") {
		cleanTitle = noteDate.Format("Jan 02, 2006") + " — " + cleanTitle
	}
	entry.Metadata["title"] = cleanTitle

	// Tags
	entry.Tags = append(entry.Tags,
		"text",
		"document",
		"sanctuary:"+sanctuaryID,
		"year:"+strconv.Itoa(noteDate.Year()),
	)
	if subSanctuary != "" {
		entry.Tags = append(entry.Tags, "sub:"+strings.ToLower(subSanctuary))
	}

	// Astrological & Alchemical Dasha Resolution
	if len(timeline) > 0 {
		snap := dasha.ResolveActiveSnapshot(timeline, noteDate)
		mahaGraha := snap.Mahadasha.Planet
		antarGraha := snap.Antardasha.Planet

		metalName, metalLatin := PlanetSacredMetal(mahaGraha)
		antarMetalName, _ := PlanetSacredMetal(antarGraha)
		hermeticAxiom := PlanetHermeticAxiom(mahaGraha)
		storyArchetype := PlanetStoryArchetype(mahaGraha)

		entry.Metadata["dasha_mahadasha"] = snap.Mahadasha.PlanetName
		entry.Metadata["dasha_mahadasha_sanskrit"] = snap.Mahadasha.SanskritName
		entry.Metadata["dasha_antardasha"] = snap.Antardasha.PlanetName
		entry.Metadata["dasha_antardasha_sanskrit"] = snap.Antardasha.SanskritName
		entry.Metadata["sacred_metal"] = metalName + " (" + metalLatin + ")"
		entry.Metadata["antardasha_metal"] = antarMetalName
		entry.Metadata["hermetic_axiom"] = hermeticAxiom
		entry.Metadata["story_archetype"] = storyArchetype

		entry.Tags = append(entry.Tags,
			"mahadasha:"+strings.ToLower(snap.Mahadasha.PlanetName),
			"antardasha:"+strings.ToLower(snap.Antardasha.PlanetName),
			"metal:"+strings.ToLower(metalName),
		)
	}

	// Entity and Archetype Recognition
	corpusLower := strings.ToLower(entry.FileName + " " + entry.Snippet)
	var entities []string
	keyEntities := []string{
		"vernon", "mark", "deborah", "janelle", "sarah", "luke", "joshua", "isaac", "norma",
		"bell canada", "karate", "chiron", "kybalion", "great work", "alchemy",
	}
	for _, ent := range keyEntities {
		if strings.Contains(corpusLower, ent) {
			entities = append(entities, ent)
			entry.Tags = append(entry.Tags, "entity:"+strings.ReplaceAll(ent, " ", "_"))
		}
	}
	if len(entities) > 0 {
		entry.Metadata["entities"] = entities
	}

	entry.Tags = dedupeStrings(entry.Tags)
}

// LoadDefaultTimeline retrieves the primary profile's timeline for Dasha grounding.
func LoadDefaultTimeline(store db.StorageEngine) []dasha.DashaPeriod {
	if store == nil {
		return nil
	}
	raw, err := store.GetProfile("profile:sovereign-genesis")
	if err != nil || len(raw) == 0 {
		keys, err := store.ListProfiles()
		if err == nil && len(keys) > 0 {
			raw, _ = store.GetProfile(keys[0])
		}
	}
	if len(raw) > 0 {
		var prof dasha.DashaProfile
		if err := json.Unmarshal(raw, &prof); err == nil && len(prof.Timeline) > 0 {
			return prof.Timeline
		}
	}
	return nil
}
