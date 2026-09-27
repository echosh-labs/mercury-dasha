package dasha

// SacredMetal represents one of the 7 classical alchemical metals and their esoteric correspondences.
type SacredMetal struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Symbol             string   `json:"symbol"`
	GoverningPlanet    PlanetID `json:"governing_planet"`
	GoverningGraha     string   `json:"governing_graha"`
	AlchemicalEssence  string   `json:"alchemical_essence"`
	TransmutationRole  string   `json:"transmutation_role"`
	ChakraCenter       string   `json:"chakra_center"`
	ColorHex           string   `json:"color_hex"`
	ConductivityRating string   `json:"conductivity_rating"`
	QuicksilverAffinity string  `json:"quicksilver_affinity"`
}

// HermeticAxiom represents one of the seven cosmic principles from the Kybalion / Hermetic corpus.
type HermeticAxiom struct {
	Number           int      `json:"number"`
	Title            string   `json:"title"`
	CanonicalText    string   `json:"canonical_text"`
	GoverningPlanet  PlanetID `json:"governing_planet"`
	AlchemicalStage  string   `json:"alchemical_stage"`
	EsotericMeaning  string   `json:"esoteric_meaning"`
	FoundationsTheme string   `json:"foundations_theme"`
}

// MagnumOpusStage defines a phase in the Great Alchemical Work.
type MagnumOpusStage struct {
	Index       int      `json:"index"`
	Name        string   `json:"name"`
	LatinName   string   `json:"latin_name"`
	Element     string   `json:"element"`
	Planet      PlanetID `json:"planet"`
	Description string   `json:"description"`
}

// SacredMetals defines the 7 classical metals, with Quicksilver (Mercury) as the catalytic sovereign.
var SacredMetals = []SacredMetal{
	{
		ID:                 "quicksilver",
		Name:               "Quicksilver (Hydrargyrum)",
		Symbol:             "☿",
		GoverningPlanet:    PlanetMercury,
		GoverningGraha:     "Budha",
		AlchemicalEssence:  "The Universal Solvent, Fluid Mind, High-Frequency Intellectual Conduit",
		TransmutationRole:  "The Living Bridge: Dissolves rigid structures to enable renewed synthesis.",
		ChakraCenter:       "Vishuddha (Throat / Voice of Logos)",
		ColorHex:           "#10b981",
		ConductivityRating: "Super-Fluid Mental Transmission",
		QuicksilverAffinity: "Self-Sovereign Core: The Quicksilver Engine of echosh-labs.",
	},
	{
		ID:                 "gold",
		Name:               "Gold (Aurum)",
		Symbol:             "☉",
		GoverningPlanet:    PlanetSun,
		GoverningGraha:     "Surya",
		AlchemicalEssence:  "Solar Consciousness, Incorruptibility, Divine Illumination",
		TransmutationRole:  "The Sovereign Summit: Perfected self-realization and royal light.",
		ChakraCenter:       "Manipura (Solar Plexus / Center of Will)",
		ColorHex:           "#f59e0b",
		ConductivityRating: "Pure Radiance",
		QuicksilverAffinity: "Quicksilver readily amalgams with Gold, binding fluid intellect to solar authority.",
	},
	{
		ID:                 "silver",
		Name:               "Silver (Argentum)",
		Symbol:             "☽",
		GoverningPlanet:    PlanetMoon,
		GoverningGraha:     "Chandra",
		AlchemicalEssence:  "Lunar Mirror, Psychic Receptivity, Tidal Memory & Subconscious",
		TransmutationRole:  "The Reflective Matrix: Purifying emotions and reflecting subtle impressions.",
		ChakraCenter:       "Ajna (Third Eye / Intuitive Sight)",
		ColorHex:           "#06b6d4",
		ConductivityRating: "Highest Electrical & Thermal Conductivity",
		QuicksilverAffinity: "Silver gives crystalline memory to Quicksilver's restless stream.",
	},
	{
		ID:                 "iron",
		Name:               "Iron (Ferrum)",
		Symbol:             "♂",
		GoverningPlanet:    PlanetMars,
		GoverningGraha:     "Mangala",
		AlchemicalEssence:  "Dynamic Will, Catalytic Friction, Martial Energy & Forge Fire",
		TransmutationRole:  "The Crucible Hammer: Tempered courage breaking inertia.",
		ChakraCenter:       "Muladhara (Root Center / Vital Drive)",
		ColorHex:           "#ef4444",
		ConductivityRating: "Magnetic Force",
		QuicksilverAffinity: "Quicksilver resists wetting Iron, providing the impervious flask in which Hermes brews.",
	},
	{
		ID:                 "tin",
		Name:               "Tin (Stannum)",
		Symbol:             "♃",
		GoverningPlanet:    PlanetJupiter,
		GoverningGraha:     "Guru",
		AlchemicalEssence:  "Philosophical Expansion, Benevolence, Structural Grace & Abundance",
		TransmutationRole:  "The Generous Architect: Expanding horizons and transmuting sorrow into wisdom.",
		ChakraCenter:       "Anahata (Heart / Higher Consciousness)",
		ColorHex:           "#eab308",
		ConductivityRating: "Expansive Resonance",
		QuicksilverAffinity: "Tin softens and expands Quicksilver into malleable philosophical alloys.",
	},
	{
		ID:                 "copper",
		Name:               "Copper (Cuprum)",
		Symbol:             "♀",
		GoverningPlanet:    PlanetVenus,
		GoverningGraha:     "Shukra",
		AlchemicalEssence:  "Aesthetic Harmony, Magnetism, Conjugal Affinity & Creative Love",
		TransmutationRole:  "The Harmonizer: Weaving disparate frequencies into coherent beauty.",
		ChakraCenter:       "Swadhisthana (Sacral Center / Creative Life)",
		ColorHex:           "#ec4899",
		ConductivityRating: "High Magnetic Affinity",
		QuicksilverAffinity: "Copper amalgamates instantly with Quicksilver, generating sacred talismanic luster.",
	},
	{
		ID:                 "lead",
		Name:               "Lead (Plumbum)",
		Symbol:             "♄",
		GoverningPlanet:    PlanetSaturn,
		GoverningGraha:     "Shani",
		AlchemicalEssence:  "Primal Density, Gravitas, Karmic Reckoning, The Ancient Chronos",
		TransmutationRole:  "The Base Material (Prima Materia): The heavy stone from which gold is drawn.",
		ChakraCenter:       "Muladhara Base (Earth Anchor / Bone Structure)",
		ColorHex:           "#3b82f6",
		ConductivityRating: "Dense Containment",
		QuicksilverAffinity: "Quicksilver extracts the philosophical spark hidden in Lead's somber weight.",
	},
}

// HermeticAxioms contains the 7 cosmic principles of Hermes Trismegistus.
var HermeticAxioms = []HermeticAxiom{
	{
		Number:           1,
		Title:            "The Principle of Mentalism",
		CanonicalText:    "THE ALL IS MIND; The Universe is Mental.",
		GoverningPlanet:  PlanetMercury,
		AlchemicalStage:  "Separation (Extraction of Pure Mind from Matter)",
		EsotericMeaning:  "Reality originates in consciousness. All phenomena are mental projections of the Supreme Mind.",
		FoundationsTheme: "Theme of Perception: Understanding the rules of reality before bending them.",
	},
	{
		Number:           2,
		Title:            "The Principle of Correspondence",
		CanonicalText:    "As above, so below; as below, so above.",
		GoverningPlanet:  PlanetJupiter,
		AlchemicalStage:  "Conjunction (Harmonizing Macrocosm and Microcosm)",
		EsotericMeaning:  "There is an invariant harmony between the planes of spirit, mind, and physical manifestation.",
		FoundationsTheme: "Theme of Architecture: The individual journey mirroring cosmic destiny.",
	},
	{
		Number:           3,
		Title:            "The Principle of Vibration",
		CanonicalText:    "Nothing rests; everything moves; everything vibrates.",
		GoverningPlanet:  PlanetRahu,
		AlchemicalStage:  "Fermentation (Spiritual Agitation & Awakening)",
		EsotericMeaning:  "Differences between matter, energy, and spirit result solely from varying frequencies of vibration.",
		FoundationsTheme: "Theme of Disruption: Breakthroughs caused by energetic shifts and harmonic resonance.",
	},
	{
		Number:           4,
		Title:            "The Principle of Polarity",
		CanonicalText:    "Everything is Dual; everything has poles; opposites are identical in nature, differing only in degree.",
		GoverningPlanet:  PlanetKetu,
		AlchemicalStage:  "Dissolution (Reconciliation of Opposing Forces)",
		EsotericMeaning:  "Love and hate, light and dark, hot and cold are degrees of the same spectrum, transmuted along the line.",
		FoundationsTheme: "Theme of Shadow: Integrating polar opposites into transcendental wholeness.",
	},
	{
		Number:           5,
		Title:            "The Principle of Rhythm",
		CanonicalText:    "Everything flows, out and in; the pendulum-swing manifests in all things; the rhythm compensates.",
		GoverningPlanet:  PlanetMoon,
		AlchemicalStage:  "Distillation (Repeated Heating & Cooling to Perfect Purity)",
		EsotericMeaning:  "In every cycle, the swing to the right is matched by the swing to the left. Mastery is polarizing above the ebb.",
		FoundationsTheme: "Theme of Tides: Weathering the low cycles with discipline until the rise returns.",
	},
	{
		Number:           6,
		Title:            "The Principle of Cause and Effect",
		CanonicalText:    "Every Cause has its Effect; every Effect has its Cause; Chance is but a name for Law not recognized.",
		GoverningPlanet:  PlanetSaturn,
		AlchemicalStage:  "Calcination (Burning away illusion under the flame of karmic truth)",
		EsotericMeaning:  "Nothing escapes the Universal Law. Masters become Causers rather than mere instruments of outside effects.",
		FoundationsTheme: "Theme of Consequence: Accepting sovereign responsibility for created outcomes.",
	},
	{
		Number:           7,
		Title:            "The Principle of Gender",
		CanonicalText:    "Gender is in everything; everything has its Masculine and Feminine Principles; Gender manifests on all planes.",
		GoverningPlanet:  PlanetVenus,
		AlchemicalStage:  "Coagulation (The Sacred Marriage of Active Will and Receptive Matrix)",
		EsotericMeaning:  "Creation requires the synergistic interplay of the assertive projection and the nurturing reception.",
		FoundationsTheme: "Theme of Creation: The synthesis of thought and feeling birthing living reality.",
	},
}

// MagnumOpusStages defines the 7 stages of the Great Alchemical Work.
var MagnumOpusStages = []MagnumOpusStage{
	{Index: 1, Name: "Calcination", LatinName: "Calcinatio", Element: "Fire", Planet: PlanetMars, Description: "Reduction of ego and inertia to ash under transformative flame."},
	{Index: 2, Name: "Dissolution", LatinName: "Solutio", Element: "Water", Planet: PlanetMoon, Description: "Dissolving rigid thoughts in subconscious waters of renewal."},
	{Index: 3, Name: "Separation", LatinName: "Separatio", Element: "Air", Planet: PlanetMercury, Description: "Precise discernment: separating the subtle from the gross."},
	{Index: 4, Name: "Conjunction", LatinName: "Conjunctio", Element: "Earth/Water", Planet: PlanetVenus, Description: "The sacred union of previously divided inner elements."},
	{Index: 5, Name: "Fermentation", LatinName: "Fermentatio", Element: "Ether", Planet: PlanetJupiter, Description: "Inoculation of new spiritual inspiration into the vessel."},
	{Index: 6, Name: "Distillation", LatinName: "Distillatio", Element: "Air/Fire", Planet: PlanetMercury, Description: "Repeated vaporization and condensation into quintessential clarity."},
	{Index: 7, Name: "Coagulation", LatinName: "Coagulatio", Element: "Earth", Planet: PlanetSaturn, Description: "Solidification of the Philosopher's Stone: sovereign embodiment in physical reality."},
}

// MetalsByPlanet maps PlanetID to SacredMetal for O(1) lookup.
var MetalsByPlanet = func() map[PlanetID]SacredMetal {
	m := make(map[PlanetID]SacredMetal, len(SacredMetals))
	for _, metal := range SacredMetals {
		m[metal.GoverningPlanet] = metal
	}
	return m
}()

// AxiomsByPlanet maps PlanetID to HermeticAxiom for O(1) lookup.
var AxiomsByPlanet = func() map[PlanetID]HermeticAxiom {
	m := make(map[PlanetID]HermeticAxiom, len(HermeticAxioms))
	for _, axiom := range HermeticAxioms {
		m[axiom.GoverningPlanet] = axiom
	}
	return m
}()

// LiveAlchemicalAlignment describes the active alchemical frequency of the engine right now.
type LiveAlchemicalAlignment struct {
	ActivePlanet    PlanetID        `json:"active_planet"`
	ActiveMetal     SacredMetal     `json:"active_metal"`
	GoverningAxiom  HermeticAxiom   `json:"governing_axiom"`
	MagnumOpusStage MagnumOpusStage `json:"magnum_opus_stage"`
	AlchemicalMotto string          `json:"alchemical_motto"`
}

// ResolveAlchemicalAlignment resolves the alchemical alignment for any planet.
func ResolveAlchemicalAlignment(planet PlanetID) LiveAlchemicalAlignment {
	metal, ok := MetalsByPlanet[planet]
	if !ok {
		metal = MetalsByPlanet[PlanetMercury] // Default to Quicksilver
	}

	axiom, ok := AxiomsByPlanet[planet]
	if !ok {
		axiom = AxiomsByPlanet[PlanetMercury]
	}

	var stage MagnumOpusStage
	for _, s := range MagnumOpusStages {
		if s.Planet == planet {
			stage = s
			break
		}
	}
	if stage.Name == "" {
		stage = MagnumOpusStages[2] // Separation (Mercury)
	}

	return LiveAlchemicalAlignment{
		ActivePlanet:    planet,
		ActiveMetal:     metal,
		GoverningAxiom:  axiom,
		MagnumOpusStage: stage,
		AlchemicalMotto: "Solve et Coagula — Dissolve and Coagulate through the Quicksilver Metronome.",
	}
}
