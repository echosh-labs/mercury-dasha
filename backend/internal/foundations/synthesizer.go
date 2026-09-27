package foundations

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
)

// SeedInput defines the input data to synthesize a StorySeed.
type SeedInput struct {
	Title               string
	SourcePath          string
	SourceType          string // "written_note" | "spoken_transcript"
	Content             string
	HasAudio            bool
	AudioURL            string
	TranscriptURL       string
	SanctuaryID         string
	SanctuaryLabel      string
	SubSanctuary        string
	NoteDate            string // "YYYY-MM-DD"
	Year                int
	ProtagonistOverride string
}

// ProtagonistArchetypes maps PlanetID to sovereign character personas.
var ProtagonistArchetypes = map[dasha.PlanetID]string{
	dasha.PlanetSaturn:  "The Alchemist Architect (Master of Time & Bone)",
	dasha.PlanetMercury: "The Quicksilver Scribe (Weaver of Volatile Words)",
	dasha.PlanetJupiter: "The Sovereign Preceptor (Keeper of the Sacred Tree)",
	dasha.PlanetSun:     "The Solar Sovereign (Illuminator of the Center)",
	dasha.PlanetMoon:    "The Tidal Mystic (Guardian of Subconscious Memory)",
	dasha.PlanetMars:    "The Forge Warrior (Igniter of Will & Iron)",
	dasha.PlanetVenus:   "The Harmony Weaver (Conjoiner of Divided Worlds)",
	dasha.PlanetRahu:    "The Eclipse Innovator (Disrupter of Stagnant Horizons)",
	dasha.PlanetKetu:    "The Silent Ascetic (Transcender of the Veil)",
}

// VisualPalettes maps PlanetID to aesthetic color sets.
var VisualPalettes = map[dasha.PlanetID][]string{
	dasha.PlanetSaturn:  {"#0f172a", "#1e293b", "#3b82f6", "#94a3b8"}, // Deep obsidian slate & lead blue
	dasha.PlanetMercury: {"#022c22", "#065f46", "#10b981", "#6ee7b7"}, // Quicksilver emerald & teal
	dasha.PlanetJupiter: {"#1c1917", "#78350f", "#d97706", "#fcd34d"}, // Royal tin & amber gold
	dasha.PlanetSun:     {"#451a03", "#b45309", "#f59e0b", "#fef08a"}, // Solar radiant aurum
	dasha.PlanetMoon:    {"#082f49", "#0e7490", "#06b6d4", "#e0f2fe"}, // Lunar tide & reflective argentum
	dasha.PlanetMars:    {"#450a0a", "#991b1b", "#ef4444", "#fca5a5"}, // Forge iron & martial crimson
	dasha.PlanetVenus:   {"#14532d", "#15803d", "#22c55e", "#86efac"}, // Copper green & harmonic jade
	dasha.PlanetRahu:    {"#2e1065", "#581c87", "#9333ea", "#d8b4fe"}, // Smoky ultraviolet & eclipse amethyst
	dasha.PlanetKetu:    {"#18181b", "#3f3f46", "#71717a", "#e4e4e7"}, // Ash gray & transcendental smoke
}

// SynthesizeSeed synthesizes a complete Foundations Story Seed from document text, date, and Dasha timeline.
func SynthesizeSeed(input SeedInput, profile *dasha.DashaProfile, timeline []dasha.DashaPeriod) (*StorySeed, error) {
	if strings.TrimSpace(input.Content) == "" {
		return nil, fmt.Errorf("document content cannot be empty for story synthesis")
	}

	// 1. Resolve composition date & active Dasha era
	noteTime := time.Now()
	if input.NoteDate != "" {
		if parsed, err := time.Parse("2006-01-02", input.NoteDate); err == nil {
			noteTime = parsed
		}
	} else if input.Year > 0 {
		noteTime = time.Date(input.Year, time.June, 15, 12, 0, 0, 0, time.UTC)
	}

	snapshot := dasha.ResolveActiveSnapshot(timeline, noteTime)

	// 2. Derive Cosmic Alignment
	maha := snapshot.Mahadasha
	antar := snapshot.Antardasha
	mPlanet := dasha.PlanetsByID[maha.Planet]
	aPlanet := dasha.PlanetsByID[antar.Planet]

	sacredMetal := maha.SacredMetal
	if antar.SacredMetal != "" && antar.SacredMetal != maha.SacredMetal {
		sacredMetal = fmt.Sprintf("%s / %s", maha.SacredMetal, antar.SacredMetal)
	}

	hermeticAxiom := antar.HermeticAxiom
	if hermeticAxiom == "" {
		hermeticAxiom = maha.HermeticAxiom
	}

	// Determine Magnum Opus stage corresponding to planet
	opusStage := "Calcination"
	for _, stage := range dasha.MagnumOpusStages {
		if stage.Planet == maha.Planet || stage.Planet == antar.Planet {
			opusStage = stage.Name
			break
		}
	}

	cosmic := CosmicAlignment{
		Mahadasha:          mPlanet.Name,
		MahadashaSanskrit:  mPlanet.SanskritName,
		Antardasha:         aPlanet.Name,
		AntardashaSanskrit: aPlanet.SanskritName,
		SacredMetal:        sacredMetal,
		HermeticAxiom:      hermeticAxiom,
		MagnumOpusStage:    opusStage,
		ActiveArchetype:    snapshot.StoryContext.ActiveArchetype,
		ResonantFrequency:  mPlanet.RootFrequency,
		ChakraFocus:        fmt.Sprintf("%s -> %s", mPlanet.ChakraCenter, aPlanet.ChakraCenter),
	}

	// 3. Extract Narrative Hook & Dialogue Anchor
	hook := ExtractNarrativeHook(input.Content)
	anchor := ExtractDialogueAnchor(input.Content)
	words := countWords(input.Content)

	// 4. Protagonist Role
	protagonist := input.ProtagonistOverride
	if protagonist == "" {
		if role, ok := ProtagonistArchetypes[maha.Planet]; ok {
			protagonist = role
		} else {
			protagonist = fmt.Sprintf("The %s Seeker", mPlanet.Name)
		}
	}

	// 5. Generate 3-Act Alchemical Scene Beats
	sceneBeats := GenerateSceneBeats(input, cosmic, hook, anchor)

	// 6. Visual Directives
	palette := VisualPalettes[maha.Planet]
	if len(palette) == 0 {
		palette = []string{"#0f172a", "#3b82f6", "#10b981", "#f59e0b"}
	}

	visDirectives := VisualDirectives{
		ColorPalette:    palette,
		Atmosphere:      fmt.Sprintf("Sanctuary of %s: Introspective, atmospheric, grounded in the %s principle.", mPlanet.Name, hermeticAxiom),
		Lighting:        fmt.Sprintf("Chiaroscuro illumination accented with %s highlights and ember radiance.", sacredMetal),
		SuggestedMotion: "ken_burns_zoom_in",
	}

	// 7. Seed ID & Title
	datePrefix := noteTime.Format("20060102")
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(input.SourcePath+input.Content)))[:8]
	seedID := fmt.Sprintf("seed-%s-%s", datePrefix, hash)

	title := input.Title
	if title == "" {
		if input.SubSanctuary != "" {
			title = fmt.Sprintf("%s • %s Chronicle", input.SubSanctuary, mPlanet.Name)
		} else {
			title = fmt.Sprintf("%s: The %s Epoch", input.SanctuaryLabel, mPlanet.Name)
		}
	}

	// 8. Assemble Tags
	tagSet := make(map[string]bool)
	tagSet["foundations_seed"] = true
	tagSet["sanctuary:"+input.SanctuaryID] = true
	tagSet["graha:"+strings.ToLower(mPlanet.Name)] = true
	tagSet["metal:"+strings.ToLower(strings.ReplaceAll(sacredMetal, " ", "_"))] = true
	if input.SourceType != "" {
		tagSet["source:"+input.SourceType] = true
	}
	if input.HasAudio {
		tagSet["has_audio"] = true
	}
	if input.Year > 0 {
		tagSet[fmt.Sprintf("year:%d", input.Year)] = true
	}

	var tags []string
	for t := range tagSet {
		tags = append(tags, t)
	}

	seed := &StorySeed{
		ID:               seedID,
		Title:            title,
		SourcePath:       input.SourcePath,
		SourceType:       input.SourceType,
		HasAudio:         input.HasAudio,
		AudioURL:         input.AudioURL,
		TranscriptURL:    input.TranscriptURL,
		SanctuaryID:      input.SanctuaryID,
		SanctuaryLabel:   input.SanctuaryLabel,
		SubSanctuary:     input.SubSanctuary,
		NoteDate:         input.NoteDate,
		Year:             input.Year,
		WordCount:        words,
		NarrativeHook:    hook,
		DialogueAnchor:   anchor,
		ProtagonistRole:  protagonist,
		CosmicAlignment:  cosmic,
		SceneBeats:       sceneBeats,
		VisualDirectives: visDirectives,
		Tags:             tags,
		CreatedAt:        time.Now(),
	}

	return seed, nil
}

// ExtractNarrativeHook pulls the opening thesis or leading philosophical sentence from content.
func ExtractNarrativeHook(content string) string {
	clean := strings.TrimSpace(content)
	lines := strings.Split(clean, "\n")

	for _, line := range lines {
		l := strings.TrimSpace(line)
		// Skip date headers, metadata lines, empty lines, and short labels
		if l == "" || strings.HasPrefix(l, "#") || strings.HasPrefix(l, "---") || strings.HasPrefix(l, "//") {
			continue
		}
		if utf8.RuneCountInString(l) < 20 {
			continue
		}
		// If line ends with colon or looks like a header, skip
		if strings.HasSuffix(l, ":") && utf8.RuneCountInString(l) < 40 {
			continue
		}

		// Find first or second sentence
		sentences := splitSentences(l)
		if len(sentences) > 0 && len(sentences[0]) > 25 {
			if len(sentences[0]) > 240 {
				return sentences[0][:237] + "..."
			}
			return sentences[0]
		}
		if len(l) > 240 {
			return l[:237] + "..."
		}
		return l
	}

	if len(clean) > 200 {
		return clean[:197] + "..."
	}
	return clean
}

// ExtractDialogueAnchor finds a potent, quotable passage suitable for spoken narration or character dialogue.
func ExtractDialogueAnchor(content string) string {
	clean := strings.TrimSpace(content)

	// 1. Look for explicit quoted text: "..."
	quoteRegex := regexp.MustCompile(`"([^"]{20,200})"`)
	matches := quoteRegex.FindAllStringSubmatch(clean, -1)
	if len(matches) > 0 {
		return matches[0][1]
	}

	// 2. Look for poignant sentences containing philosophical keywords
	keywords := []string{"remember", "must", "truth", "soul", "heart", "god", "father", "bless", "son", "daughter", "law", "mind", "light", "stone"}
	sentences := splitSentences(clean)

	for _, s := range sentences {
		sTrim := strings.TrimSpace(s)
		if len(sTrim) < 30 || len(sTrim) > 200 {
			continue
		}
		sLower := strings.ToLower(sTrim)
		for _, kw := range keywords {
			if strings.Contains(sLower, kw) {
				return sTrim
			}
		}
	}

	// 3. Fallback: Take a strong middle sentence
	if len(sentences) >= 3 {
		mid := sentences[len(sentences)/2]
		if len(mid) > 200 {
			return mid[:197] + "..."
		}
		return strings.TrimSpace(mid)
	}

	if len(clean) > 180 {
		return clean[:177] + "..."
	}
	return clean
}

// GenerateSceneBeats structures the note's theme into 3 distinct acts of personal transmutation.
func GenerateSceneBeats(input SeedInput, cosmic CosmicAlignment, hook, anchor string) []SceneBeat {
	topic := input.Title
	if topic == "" {
		topic = input.SanctuaryLabel
	}

	themeAxiom := cosmic.HermeticAxiom
	stage := cosmic.MagnumOpusStage

	// Act 1: The Principle / Genesis
	act1 := SceneBeat{
		Act:           1,
		Title:         fmt.Sprintf("Act I: Genesis under %s", cosmic.Mahadasha),
		Premise:       fmt.Sprintf("The seeker enters the crucible of experience. Core thesis: %s", hook),
		EmotionalTone: "Contemplative Reflection & Primal Stasis",
		PromptDirective: fmt.Sprintf(
			"Cinematic establishing shot of an ancient sanctuary bathed in %s tones, warm candle amber, solitary figure contemplating manuscripts.",
			cosmic.SacredMetal,
		),
		VisualKeywords: []string{"sanctuary", "scrolls", "candlelight", strings.ToLower(cosmic.Mahadasha)},
	}

	// Act 2: Alchemical Transmutation / The Friction
	dialogueExcerpt := anchor
	if len(dialogueExcerpt) > 80 {
		dialogueExcerpt = dialogueExcerpt[:77] + "..."
	}

	act2 := SceneBeat{
		Act:           2,
		Title:         fmt.Sprintf("Act II: The Crucible of %s", stage),
		Premise:       fmt.Sprintf("The Great Work confronts the soul. Hermetic alignment with '%s'. Voiceover: \"%s\"", themeAxiom, dialogueExcerpt),
		EmotionalTone: "Intense Transmutation & Dynamic Friction",
		PromptDirective: fmt.Sprintf(
			"Macro cinematic frame of dynamic alchemical transmutation, fluid metals merging with glowing embers, volatile quicksilver meeting dense lead.",
		),
		VisualKeywords: []string{"crucible", "transmutation", "friction", "hermetic_vessel"},
	}

	// Act 3: Sovereign Realization / Coagulation
	act3 := SceneBeat{
		Act:           3,
		Title:         fmt.Sprintf("Act III: The Sovereign Stone of %s", cosmic.Antardasha),
		Premise:       fmt.Sprintf("Wisdom is crystallized from experience. The protagonist embodies sovereign authority and returns to the world transformed."),
		EmotionalTone: "Serene Mastery & Sovereign Grounding",
		PromptDirective: fmt.Sprintf(
			"Expansive panoramic vista at dawn, golden hour atmospheric haze, the perfected alchemical stone resting upon a carved stone altar.",
		),
		VisualKeywords: []string{"dawn", "sovereign_stone", "golden_hour", "triumph"},
	}

	return []SceneBeat{act1, act2, act3}
}

// FindResonantExcerpts matches catalog documents against the active celestial transit.
func FindResonantExcerpts(items []CatalogItemRef, snapshot *dasha.ActivePeriodSnapshot, limit int) []ResonantExcerpt {
	if limit <= 0 {
		limit = 10
	}

	currentMaha := strings.ToLower(snapshot.Mahadasha.PlanetName)
	currentAntar := strings.ToLower(snapshot.Antardasha.PlanetName)
	currentAxiom := strings.ToLower(snapshot.Antardasha.HermeticAxiom)
	currentMetal := strings.ToLower(snapshot.Mahadasha.SacredMetal)

	var results []ResonantExcerpt

	for _, it := range items {
		matchReason := ""
		matchAttr := ""

		itMaha := strings.ToLower(it.DashaMahadasha)
		itAntar := strings.ToLower(it.DashaAntardasha)
		itAxiom := strings.ToLower(it.HermeticAxiom)
		itMetal := strings.ToLower(it.SacredMetal)

		// Priority 1: Exact Mahadasha and Antardasha harmony
		if itMaha == currentMaha && itAntar == currentAntar {
			matchReason = fmt.Sprintf("Written during the identical %s / %s cosmic epoch", snapshot.Mahadasha.PlanetName, snapshot.Antardasha.PlanetName)
			matchAttr = "graha"
		} else if itMaha == currentMaha {
			matchReason = fmt.Sprintf("Shares current %s Mahadasha period", snapshot.Mahadasha.PlanetName)
			matchAttr = "graha"
		} else if itAntar == currentAntar {
			matchReason = fmt.Sprintf("Governed by current %s Antardasha sub-cycle", snapshot.Antardasha.PlanetName)
			matchAttr = "graha"
		} else if itAxiom != "" && strings.Contains(itAxiom, currentAxiom) {
			matchReason = fmt.Sprintf("Expresses the resonant Hermetic Axiom: %s", snapshot.Antardasha.HermeticAxiom)
			matchAttr = "axiom"
		} else if itMetal != "" && strings.Contains(itMetal, currentMetal) {
			matchReason = fmt.Sprintf("Resonates with current Sacred Metal: %s", snapshot.Mahadasha.SacredMetal)
			matchAttr = "metal"
		}

		if matchReason != "" {
			results = append(results, ResonantExcerpt{
				CatalogID:         it.ID,
				Title:             it.Title,
				Path:              it.Path,
				SourceType:        it.SourceType,
				Sanctuary:         it.Sanctuary,
				SanctuaryLabel:    it.SanctuaryLabel,
				NoteDate:          it.NoteDate,
				Year:              it.Year,
				HasAudio:          it.HasAudio,
				AudioURL:          it.AudioURL,
				ResonanceReason:   matchReason,
				MatchingAttribute: matchAttr,
				Snippet:           it.Snippet,
				Tags:              it.Tags,
			})
			if len(results) >= limit {
				break
			}
		}
	}

	return results
}

// splitSentences divides text into sentences by punctuation marks.
func splitSentences(text string) []string {
	re := regexp.MustCompile(`[.!?]+(?:\s+|$)`)
	splits := re.Split(text, -1)
	var out []string
	for _, s := range splits {
		t := strings.TrimSpace(s)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

// countWords computes word count.
func countWords(s string) int {
	return len(strings.Fields(s))
}
