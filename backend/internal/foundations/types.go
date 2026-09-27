package foundations

import (
	"time"
)

// SceneBeat represents a single act or dramatic beat in a synthesized story arc.
type SceneBeat struct {
	Act             int      `json:"act"`                       // 1, 2, 3
	Title           string   `json:"title"`                     // e.g. "The Principle of Cause and Effect"
	Premise         string   `json:"premise"`                   // Dramatic situation or thesis
	EmotionalTone   string   `json:"emotional_tone"`            // e.g. "Contemplative Reflection", "Alchemical Transmutation"
	PromptDirective string   `json:"prompt_directive"`          // Visual prompt directive for video/image generation
	VisualKeywords  []string `json:"visual_keywords,omitempty"` // Mood, lighting, setting tags
}

// VisualDirectives specifies aesthetic and cinematic parameters derived from the cosmic epoch.
type VisualDirectives struct {
	ColorPalette    []string `json:"color_palette"`            // Hex colors aligned with Sacred Metal & Graha
	Atmosphere      string   `json:"atmosphere"`               // Mood description (e.g. "Nocturnal Alchemical Sanctuary")
	Lighting        string   `json:"lighting"`                 // Lighting directive (e.g. "Low-key candlelight with gold ember reflections")
	SuggestedMotion string   `json:"suggested_motion"`         // e.g. "ken_burns_zoom_in", "slow_pan"
}

// CosmicAlignment captures the Vimshottari Dasha era and alchemical stage when the note was written.
type CosmicAlignment struct {
	Mahadasha          string  `json:"mahadasha"`                   // e.g. "Saturn"
	MahadashaSanskrit  string  `json:"mahadasha_sanskrit"`          // e.g. "Shani"
	Antardasha         string  `json:"antardasha"`                  // e.g. "Mercury"
	AntardashaSanskrit string  `json:"antardasha_sanskrit"`         // e.g. "Budha"
	SacredMetal        string  `json:"sacred_metal"`                // e.g. "Lead & Quicksilver"
	HermeticAxiom      string  `json:"hermetic_axiom"`              // e.g. "The Principle of Cause and Effect"
	MagnumOpusStage    string  `json:"magnum_opus_stage"`           // e.g. "Calcination", "Sublimation"
	ActiveArchetype    string  `json:"active_archetype"`            // e.g. "The Alchemical Architect"
	ResonantFrequency  float64 `json:"resonant_frequency_hz"`       // e.g. 147.85 Hz
	ChakraFocus        string  `json:"chakra_focus,omitempty"`      // e.g. "Muladhara -> Vishuddha"
}

// StorySeed represents an archetypal narrative blueprint synthesized from personal writings or voice chronicles.
type StorySeed struct {
	ID               string           `json:"id"`
	Title            string           `json:"title"`
	SourcePath       string           `json:"source_path"`
	SourceType       string           `json:"source_type"` // "written_note" | "spoken_transcript"
	HasAudio         bool             `json:"has_audio"`
	AudioURL         string           `json:"audio_url,omitempty"`
	TranscriptURL    string           `json:"transcript_url,omitempty"`
	SanctuaryID      string           `json:"sanctuary_id"`
	SanctuaryLabel   string           `json:"sanctuary_label"`
	SubSanctuary     string           `json:"sub_sanctuary,omitempty"`
	NoteDate         string           `json:"note_date"`
	Year             int              `json:"year"`
	WordCount        int              `json:"word_count"`
	NarrativeHook    string           `json:"narrative_hook"`      // Opening thesis / premise
	DialogueAnchor   string           `json:"dialogue_anchor"`     // Quotable spoken or written passage for voiceover
	ProtagonistRole  string           `json:"protagonist_role"`    // Archetypal role aligned with Dasha ruler
	CosmicAlignment  CosmicAlignment  `json:"cosmic_alignment"`
	SceneBeats       []SceneBeat      `json:"scene_beats"`         // 3-Act structured narrative progression
	VisualDirectives VisualDirectives `json:"visual_directives"`
	Tags             []string         `json:"tags,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
}

// SeedFromTextRequest represents the input parameters for synthesizing a Story Seed.
type SeedFromTextRequest struct {
	Path        string `json:"path,omitempty"`        // Path to text note, e.g. "text/2013-03-15.txt"
	ID          string `json:"id,omitempty"`          // Note ID or recorder ID e.g. "recorder:12345"
	Content     string `json:"content,omitempty"`     // Direct text content (if not loading by path/id)
	Title       string `json:"title,omitempty"`       // Optional title override
	NoteDate    string `json:"note_date,omitempty"`   // Optional explicit date override (YYYY-MM-DD)
	Sanctuary   string `json:"sanctuary,omitempty"`   // Optional sanctuary override
	Protagonist string `json:"protagonist,omitempty"` // Optional custom protagonist role
	Save        bool   `json:"save,omitempty"`        // Whether to persist into BoltDB foundations_seeds
}

// CatalogItemRef abstracts catalog items for transit resonance calculations.
type CatalogItemRef struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Path            string   `json:"path"`
	SourceType      string   `json:"source_type"`
	Sanctuary       string   `json:"sanctuary"`
	SanctuaryLabel  string   `json:"sanctuary_label"`
	NoteDate        string   `json:"note_date"`
	Year            int      `json:"year"`
	DashaMahadasha  string   `json:"dasha_mahadasha,omitempty"`
	DashaAntardasha string   `json:"dasha_antardasha,omitempty"`
	SacredMetal     string   `json:"sacred_metal,omitempty"`
	HermeticAxiom   string   `json:"hermetic_axiom,omitempty"`
	HasAudio        bool     `json:"has_audio"`
	AudioURL        string   `json:"audio_url,omitempty"`
	Snippet         string   `json:"snippet,omitempty"`
	Tags            []string `json:"tags,omitempty"`
}

// ResonantExcerpt represents a life's work passage that resonates with the current celestial transit.
type ResonantExcerpt struct {
	CatalogID         string   `json:"catalog_id"`
	Title             string   `json:"title"`
	Path              string   `json:"path"`
	SourceType        string   `json:"source_type"`
	Sanctuary         string   `json:"sanctuary"`
	SanctuaryLabel    string   `json:"sanctuary_label"`
	NoteDate          string   `json:"note_date"`
	Year              int      `json:"year"`
	HasAudio          bool     `json:"has_audio"`
	AudioURL          string   `json:"audio_url,omitempty"`
	ResonanceReason   string   `json:"resonance_reason"`   // e.g. "Matches current Saturn Mahadasha", "Shares Hermetic Axiom"
	MatchingAttribute string   `json:"matching_attribute"` // "graha", "axiom", "metal"
	Snippet           string   `json:"snippet"`
	Tags              []string `json:"tags,omitempty"`
}

// ResonantTransitResponse wraps resonant life's work excerpts for the active transit.
type ResonantTransitResponse struct {
	CurrentTransit struct {
		Mahadasha     string `json:"mahadasha"`
		Antardasha    string `json:"antardasha"`
		SacredMetal   string `json:"sacred_metal"`
		HermeticAxiom string `json:"hermetic_axiom"`
		Archetype     string `json:"archetype"`
	} `json:"current_transit"`
	Count    int               `json:"count"`
	Excerpts []ResonantExcerpt `json:"excerpts"`
}
