package dasha

import (
	"time"
)

// PlanetID represents the identifier for one of the 9 classical Vedic Grahas.
type PlanetID string

const (
	PlanetKetu    PlanetID = "ketu"
	PlanetVenus   PlanetID = "venus"
	PlanetSun     PlanetID = "sun"
	PlanetMoon    PlanetID = "moon"
	PlanetMars    PlanetID = "mars"
	PlanetRahu    PlanetID = "rahu"
	PlanetJupiter PlanetID = "jupiter"
	PlanetSaturn  PlanetID = "saturn"
	PlanetMercury PlanetID = "mercury"
)

// Planet holds astrological, alchemical, and vibrational correspondences for a Graha.
type Planet struct {
	ID             PlanetID `json:"id"`
	Name           string   `json:"name"`
	SanskritName   string   `json:"sanskrit_name"`
	DurationYears  float64  `json:"duration_years"` // Standard Vimshottari period (total = 120 yrs)
	RootFrequency  float64  `json:"root_frequency_hz"`
	Element        string   `json:"element"`
	ChakraCenter   string   `json:"chakra_center"`
	ColorHex       string   `json:"color_hex"`
	HermeticAxiom  string   `json:"hermetic_axiom"`
	StoryArchetype string   `json:"story_archetype"` // Foundational narrative archetype
}

// Pada represents one 3°20' quadrant of a Nakshatra (4 Padas per Nakshatra = 108 total in Zodiac).
type Pada struct {
	Number       int    `json:"number"`        // 1 to 4
	DegreesStart string `json:"degrees_start"` // e.g. "00°00'"
	DegreesEnd   string `json:"degrees_end"`   // e.g. "03°20'"
	NavamshaSign string `json:"navamsha_sign"` // Navamsha zodiac sign
}

// Nakshatra represents one of the 27 lunar mansions (13°20' / 800 arcminutes each).
type Nakshatra struct {
	Index        int      `json:"index"` // 1 to 27
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	SanskritName string   `json:"sanskrit_name"`
	DegreeStart  float64  `json:"degree_start"` // 0.0 to 360.0
	DegreeEnd    float64  `json:"degree_end"`
	ZodiacSpan   string   `json:"zodiac_span"` // e.g. "00°00' - 13°20' Aries"
	RulingPlanet PlanetID `json:"ruling_planet"`
	FrequencyHz  float64  `json:"frequency_hz"`
	Deity        string   `json:"deity"`
	Symbol       string   `json:"symbol"`
	Quality      string   `json:"quality"`
	Padas        []Pada   `json:"padas"`
}

// DashaPeriodLevel indicates the tier in the period hierarchy.
type DashaPeriodLevel string

const (
	LevelMahadasha       DashaPeriodLevel = "mahadasha"
	LevelAntardasha      DashaPeriodLevel = "antardasha"
	LevelPratyantardasha DashaPeriodLevel = "pratyantardasha"
)

// DashaPeriod represents a discrete segment of planetary rulership on the timeline.
type DashaPeriod struct {
	Level          DashaPeriodLevel `json:"level"`
	Planet         PlanetID         `json:"planet"`
	PlanetName     string           `json:"planet_name"`
	SanskritName   string           `json:"sanskrit_name"`
	ColorHex       string           `json:"color_hex"`
	DurationDays   float64          `json:"duration_days"`
	StartDate      time.Time        `json:"start_date"`
	EndDate        time.Time        `json:"end_date"`
	SacredMetal    string           `json:"sacred_metal,omitempty"`
	MetalSymbol    string           `json:"metal_symbol,omitempty"`
	HermeticAxiom  string           `json:"hermetic_axiom,omitempty"`
	StoryArchetype string           `json:"story_archetype,omitempty"`
	FrequencyHz    float64          `json:"frequency_hz,omitempty"`
	AgeStart       float64          `json:"age_start,omitempty"`
	AgeEnd         float64          `json:"age_end,omitempty"`
	SubPeriods     []DashaPeriod    `json:"sub_periods,omitempty"`
}

// ActivePeriodSnapshot captures the exact active cycle at a specified reference time.
type ActivePeriodSnapshot struct {
	ReferenceTime      time.Time    `json:"reference_time"`
	Mahadasha          DashaPeriod  `json:"mahadasha"`
	Antardasha         DashaPeriod  `json:"antardasha"`
	Pratyantardasha    DashaPeriod  `json:"pratyantardasha"`
	MahaElapsedDays    float64      `json:"maha_elapsed_days"`
	MahaRemainingDays  float64      `json:"maha_remaining_days"`
	MahaPercentDone    float64      `json:"maha_percent_done"`
	AntarElapsedDays   float64      `json:"antar_elapsed_days"`
	AntarRemainingDays float64      `json:"antar_remaining_days"`
	AntarPercentDone   float64      `json:"antar_percent_done"`
	StoryContext       StoryContext `json:"story_context"`
}

// StoryContext provides rich archetypal narrative hooks for the Foundations storytelling engine.
type StoryContext struct {
	ActiveArchetype    string   `json:"active_archetype"`
	ThematicPhase      string   `json:"thematic_phase"`
	AlchemicalElement  string   `json:"alchemical_element"`
	ChakraFocus        string   `json:"chakra_focus"`
	ResonantFrequency  float64  `json:"resonant_frequency_hz"`
	HermeticPrinciples []string `json:"hermetic_principles"`
	NarrativePrompt    string   `json:"narrative_prompt"`
}

// CalculationRequest represents input parameters for Dasha calculation.
type CalculationRequest struct {
	Name           string   `json:"name,omitempty"`
	BirthDate      string   `json:"birth_date"`                // "YYYY-MM-DD"
	BirthTime      string   `json:"birth_time"`                // "HH:MM" (optional, default "12:00")
	TimezoneOffset float64  `json:"timezone_offset,omitempty"` // Hours offset from UTC (e.g. -4 for EDT)
	Latitude       float64  `json:"latitude,omitempty"`
	Longitude      float64  `json:"longitude,omitempty"`
	NakshatraIndex int      `json:"nakshatra_index,omitempty"` // Optional manual override: 1-27
	PadaNumber     int      `json:"pada_number,omitempty"`     // 1-4
	MoonDegree     *float64 `json:"moon_degree,omitempty"`     // Optional manual override: 0.0 - 360.0
	TargetTime     *string  `json:"target_time,omitempty"`     // Reference time for active snapshot (default: now)
}

// DashaPeriodSummary represents a compact summary of a major period on the timeline.
type DashaPeriodSummary struct {
	Level        DashaPeriodLevel `json:"level"`
	Planet       PlanetID         `json:"planet"`
	PlanetName   string           `json:"planet_name"`
	SanskritName string           `json:"sanskrit_name"`
	ColorHex     string           `json:"color_hex"`
	DurationDays float64          `json:"duration_days"`
	StartDate    time.Time        `json:"start_date"`
	EndDate      time.Time        `json:"end_date"`
	SacredMetal  string           `json:"sacred_metal,omitempty"`
	MetalSymbol  string           `json:"metal_symbol,omitempty"`
	AgeRange     string           `json:"age_range,omitempty"`
}

// VedicPanchanga captures the 5 classical limbs of time at birth (Pancha-anga).
type VedicPanchanga struct {
	Vara              string  `json:"vara"`                // e.g. "Mangalavara (Tuesday)"
	VaraLord          string  `json:"vara_lord"`           // e.g. "Mars (Mangala)"
	TithiNumber       int     `json:"tithi_number"`        // 1 to 30
	TithiName         string  `json:"tithi_name"`          // e.g. "Shukla Dashami", "Purnima", "Amavasya"
	Paksha            string  `json:"paksha"`              // "Shukla" (Waxing) or "Krishna" (Waning)
	TithiPercent      float64 `json:"tithi_percent"`       // Completion percentage [0, 100)
	YogaNumber        int     `json:"yoga_number"`         // 1 to 27
	YogaName          string  `json:"yoga_name"`           // e.g. "Saubhagya", "Siddhi"
	YogaMeaning       string  `json:"yoga_meaning"`        // Benefic / Malefic quality
	KaranaNumber      int     `json:"karana_number"`       // 1 to 60
	KaranaName        string  `json:"karana_name"`         // e.g. "Vishti (Bhadra)", "Bava"
	KaranaType        string  `json:"karana_type"`         // "Movable" or "Fixed"
	SiderealSunDegree float64 `json:"sidereal_sun_degree"` // Lahiri Sidereal Sun degree [0, 360)
}

// AyurvedicConstitution defines the biological triad and character totems anchored to Janma Nakshatra.
type AyurvedicConstitution struct {
	Dosha          string `json:"dosha"`           // "Vata", "Pitta", "Kapha"
	DoshaQualities string `json:"dosha_qualities"` // Energetic traits
	Gana           string `json:"gana"`            // "Deva", "Manushya", "Rakshasa"
	YoniTotem      string `json:"yoni_totem"`      // e.g. "Lioness (Simha)"
	YoniAnimal     string `json:"yoni_animal"`     // e.g. "Lion"
	Nadi           string `json:"nadi"`            // "Adi" (Initial), "Madhya" (Middle), "Antya" (Final)
}

// NatalLagna represents the topocentric Sidereal Ascendant at birth.
type NatalLagna struct {
	Degree        float64   `json:"degree"`         // [0, 360)
	Rashi         string    `json:"rashi"`          // e.g. "Mesha (Aries)"
	RashiSanskrit string    `json:"rashi_sanskrit"` // e.g. "Mesha"
	RashiIndex    int       `json:"rashi_index"`    // 1 to 12
	DegreeInSign  float64   `json:"degree_in_sign"` // [0, 30)
	Lord          PlanetID  `json:"lord"`           // e.g. "mars"
	LordName      string    `json:"lord_name"`      // e.g. "Mars (Mangala)"
	Nakshatra     Nakshatra `json:"nakshatra"`      // Rising lunar mansion
	Pada          int       `json:"pada"`           // 1 to 4
	Symbol        string    `json:"symbol"`         // Sign glyph e.g. "♈"
	Element       string    `json:"element"`        // "Fire", "Earth", "Air", "Water"
}

// LuminaryPlacement describes the sign, degree, and house placement of a celestial body from Lagna.
type LuminaryPlacement struct {
	Degree       float64   `json:"degree"`         // [0, 360)
	Rashi        string    `json:"rashi"`          // e.g. "Tula (Libra)"
	RashiIndex   int       `json:"rashi_index"`    // 1 to 12
	DegreeInSign float64   `json:"degree_in_sign"` // [0, 30)
	RashiLord    PlanetID  `json:"rashi_lord"`     // Ruling planet of sign
	HouseNumber  int       `json:"house_number"`   // 1 to 12 from Lagna
	HouseName    string    `json:"house_name"`     // e.g. "Tanu", "Dhana", "Bhagya"
	Nakshatra    Nakshatra `json:"nakshatra"`      // Lunar mansion
	Pada         int       `json:"pada"`           // 1 to 4
}

// BhavaDetail describes one of the 12 classical houses (Bhavas) relative to the Sidereal Lagna.
type BhavaDetail struct {
	HouseNumber  int      `json:"house_number"`  // 1 to 12
	Name         string   `json:"name"`          // e.g. "Tanu (1st)", "Dhana (2nd)"
	SanskritName string   `json:"sanskrit_name"` // e.g. "Tanu", "Dhana", "Sahaja"
	Domain       string   `json:"domain"`        // Life sphere
	Rashi        string   `json:"rashi"`         // Sign occupying house
	RashiLord    string   `json:"rashi_lord"`    // Planet governing house
	Luminaries   []string `json:"luminaries"`    // Bodies placed in this house e.g. ["Lagna"], ["Surya"], ["Chandra"]
}

// TripodOfEmbodiment synthesizes the core cosmic trinity: Lagna (Body), Surya (Soul), Chandra (Mind).
type TripodOfEmbodiment struct {
	Lagna   NatalLagna         `json:"lagna"`
	Surya   LuminaryPlacement  `json:"surya"`
	Chandra LuminaryPlacement  `json:"chandra"`
	Bhavas  []BhavaDetail      `json:"bhavas"`
}

// GrahaDignity represents the evaluated classical Vedic dignity (Avastha / Sambandha) of a planet.
type GrahaDignity struct {
	Level       string `json:"level"`       // "EXALTED", "MOOLATRIKONA", "OWN_SIGN", "GREAT_FRIEND", "FRIEND", "NEUTRAL", "ENEMY", "GREAT_ENEMY", "DEBILITATED"
	Name        string `json:"name"`        // Display name e.g. "Exalted (Param Ucha)"
	Description string `json:"description"` // Energetic interpretation
	ColorHex    string `json:"color_hex"`   // UI badge hex
}

// GrahaPlacement details the precise sidereal position, sign, house, nakshatra, and dignity of a Vedic Graha.
type GrahaPlacement struct {
	ID           PlanetID     `json:"id"`             // "sun", "moon", "mars", "mercury", "jupiter", "venus", "saturn", "rahu", "ketu"
	Name         string       `json:"name"`           // e.g. "Mars"
	SanskritName string       `json:"sanskrit_name"`  // e.g. "Mangala"
	Degree       float64      `json:"degree"`         // [0, 360)
	Rashi        string       `json:"rashi"`          // e.g. "Mesha (Aries)"
	RashiIndex   int          `json:"rashi_index"`    // 1 to 12
	DegreeInSign float64      `json:"degree_in_sign"` // [0, 30)
	HouseNumber  int          `json:"house_number"`   // 1 to 12 from Lagna
	HouseName    string       `json:"house_name"`     // e.g. "Tanu", "Karma"
	Nakshatra    Nakshatra    `json:"nakshatra"`      // Lunar mansion
	Pada         int          `json:"pada"`           // 1 to 4
	Dignity      GrahaDignity `json:"dignity"`        // Evaluated essential dignity
	IsRetrograde bool         `json:"is_retrograde"`  // True if retrograde motion
}

// NatalAstrology captures the core natal astrological matrix derived from the birth chart.
type NatalAstrology struct {
	SiderealMoonDegree float64               `json:"sidereal_moon_degree"`
	SiderealSunDegree  float64               `json:"sidereal_sun_degree"`
	JanmaNakshatra     Nakshatra             `json:"janma_nakshatra"`
	JanmaPada          int                   `json:"janma_pada"`
	PadaDetail         Pada                  `json:"pada_detail"`
	StartingLord       PlanetID              `json:"starting_lord"`
	StartingLordName   string                `json:"starting_lord_name"`
	BalanceYears       float64               `json:"balance_years"`
	ElementalTattva    string                `json:"elemental_tattva"`
	Panchanga          VedicPanchanga        `json:"panchanga"`
	Ayurveda           AyurvedicConstitution `json:"ayurveda"`
	Lagna              NatalLagna            `json:"lagna"`
	Tripod             TripodOfEmbodiment    `json:"tripod"`
	Grahas             []GrahaPlacement      `json:"grahas"`
}

// NatalAlchemy encapsulates the inherent alchemical constitution anchored to the birth graha and nakshatra.
type NatalAlchemy struct {
	SacredMetal         SacredMetal     `json:"sacred_metal"`
	GoverningAxiom      HermeticAxiom   `json:"governing_axiom"`
	MagnumOpusStage     MagnumOpusStage `json:"magnum_opus_stage"`
	ChakraAnchor        string          `json:"chakra_anchor"`
	ResonantFrequencyHz float64         `json:"resonant_frequency_hz"`
	QuicksilverAffinity string          `json:"quicksilver_affinity"`
	AlchemicalMotto     string          `json:"alchemical_motto"`
}

// ActiveAlchemicalState describes the active transmutation vessel of currently running Dasha periods.
type ActiveAlchemicalState struct {
	MahadashaMetal      SacredMetal   `json:"mahadasha_metal"`
	MahadashaAxiom      HermeticAxiom `json:"mahadasha_axiom"`
	AntardashaMetal     SacredMetal   `json:"antardasha_metal"`
	AntardashaAxiom     HermeticAxiom `json:"antardasha_axiom"`
	TransmutationVessel string        `json:"transmutation_vessel"`
}

// SymbioticResonance evaluates how the current transit hora resonates with the user's natal and dasha lords.
type SymbioticResonance struct {
	TransitHoraPlanet     PlanetID      `json:"transit_hora_planet"`
	TransitHoraPlanetName string        `json:"transit_hora_planet_name"`
	TransitMetal          SacredMetal   `json:"transit_metal"`
	TransitAxiom          HermeticAxiom `json:"transit_axiom"`
	IsJanmaResonant       bool          `json:"is_janma_resonant"`
	IsMahadashaResonant   bool          `json:"is_mahadasha_resonant"`
	IsAntardashaResonant  bool          `json:"is_antardasha_resonant"`
	ResonanceTier         string        `json:"resonance_tier"` // "SOVEREIGN_JANMA_CONJUNCTION" | "MAHADASHA_HARMONIC" | "ANTARDASHA_HARMONIC" | "NEUTRAL_TRANSIT"
	TransmutationGuidance string        `json:"transmutation_guidance"`
}

// ProfileSummary is a lightweight projection for dashboards, dropdowns, and profile selection.
type ProfileSummary struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	BirthDate      string    `json:"birth_date"`
	LocationName   string    `json:"location_name"`
	NakshatraName  string    `json:"nakshatra_name"`
	NakshatraIndex int       `json:"nakshatra_index"`
	PadaNumber     int       `json:"pada_number"`
	StartingLord   string    `json:"starting_lord"`
	NatalMetal     string    `json:"natal_metal"`
	ActiveMaha     string    `json:"active_mahadasha"`
	Dosha          string    `json:"dosha,omitempty"`
	Gana           string    `json:"gana,omitempty"`
	Tithi               string    `json:"tithi,omitempty"`
	LagnaRashi          string    `json:"lagna_rashi,omitempty"`
	SunRashi            string    `json:"sun_rashi,omitempty"`
	MoonRashi           string    `json:"moon_rashi,omitempty"`
	YoniTotem           string    `json:"yoni_totem,omitempty"`
	ElementalTattva     string    `json:"elemental_tattva,omitempty"`
	AlchemicalMotto     string    `json:"alchemical_motto,omitempty"`
	MagnumOpusStage     string    `json:"magnum_opus_stage,omitempty"`
	ChakraAnchor        string    `json:"chakra_anchor,omitempty"`
	ResonantFrequencyHz float64   `json:"resonant_frequency_hz,omitempty"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// DashaProfile represents the unified sovereign profile output for a birth chart,
// fusing the astronomical ephemeris, natal astrology, natal alchemy, and 120-year timeline.
type DashaProfile struct {
	ID             string                  `json:"id"`
	Name           string                  `json:"name"`
	Title          string                  `json:"title,omitempty"`
	Backstory      string                  `json:"backstory,omitempty"`
	AvatarIcon     string                  `json:"avatar_icon,omitempty"`
	Chronicles     []CharacterChronicleRef `json:"chronicles,omitempty"`
	BirthDate      string               `json:"birth_date,omitempty"`
	BirthTime      string               `json:"birth_time,omitempty"`
	Timezone       string               `json:"timezone,omitempty"`      // e.g. "America/Toronto"
	TimezoneOffset float64              `json:"timezone_offset"`
	Latitude       float64              `json:"latitude,omitempty"`       // Decimal degrees, e.g. 43.1594
	Longitude      float64              `json:"longitude,omitempty"`      // Decimal degrees, e.g. -79.2469
	LocationName   string               `json:"location_name,omitempty"`  // e.g. "St. Catharines, ON"
	Mode           string               `json:"mode,omitempty"`
	NakshatraIndex int                  `json:"nakshatra_index,omitempty"`
	PadaNumber     int                  `json:"pada_number,omitempty"`
	BirthTimeUTC   time.Time            `json:"birth_time_utc"`
	MoonDegree     float64              `json:"moon_degree"`
	JanmaNakshatra Nakshatra            `json:"janma_nakshatra"`
	JanmaPada      int                  `json:"janma_pada"`
	BalanceYears   float64              `json:"balance_years"` // Remaining years of 1st Mahadasha at birth
	StartingLord   PlanetID             `json:"starting_lord"`
	ActiveSnapshot ActivePeriodSnapshot `json:"active_snapshot"`
	Timeline       []DashaPeriod        `json:"timeline,omitempty"` // Full 120-year sequence (omitted in compact summary mode)

	// First-Class Top-Level Astronomical and Alchemical Projections
	NakshatraName       string    `json:"nakshatra_name,omitempty"`
	LagnaRashi          string    `json:"lagna_rashi,omitempty"`
	SunRashi            string    `json:"sun_rashi,omitempty"`
	MoonRashi           string    `json:"moon_rashi,omitempty"`
	NatalMetal          string    `json:"natal_metal,omitempty"`
	AlchemicalMotto     string    `json:"alchemical_motto,omitempty"`
	Dosha               string    `json:"dosha,omitempty"`
	Gana                string    `json:"gana,omitempty"`
	Tithi               string    `json:"tithi,omitempty"`
	YoniTotem           string    `json:"yoni_totem,omitempty"`
	ElementalTattva     string    `json:"elemental_tattva,omitempty"`
	MagnumOpusStage     string    `json:"magnum_opus_stage,omitempty"`
	ChakraAnchor        string    `json:"chakra_anchor,omitempty"`
	ResonantFrequencyHz float64   `json:"resonant_frequency_hz,omitempty"`

	// Sovereign Unified Domain Extensions
	Astrology          NatalAstrology        `json:"astrology"`
	Alchemy            NatalAlchemy          `json:"alchemy"`
	ActiveAlchemy      ActiveAlchemicalState `json:"active_alchemy"`
	TimelineSummary    []DashaPeriodSummary  `json:"timeline_summary"`
	SymbioticResonance *SymbioticResonance   `json:"symbiotic_resonance,omitempty"`
	CreatedAt          time.Time             `json:"created_at,omitempty"`
	UpdatedAt          time.Time             `json:"updated_at,omitempty"`
}

// SovereignProfile is an alias for the unified DashaProfile.
type SovereignProfile = DashaProfile
