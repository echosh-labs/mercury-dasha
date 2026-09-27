package dasha

// VimshottariPlanets defines the 9 classical planetary lords in canonical Vimshottari order,
// their durations (summing to 120 years), and their Hermetic and archetypal correspondences.
var VimshottariPlanets = []Planet{
	{
		ID:             PlanetKetu,
		Name:           "Ketu",
		SanskritName:   "Ketu (South Node)",
		DurationYears:  7,
		RootFrequency:  211.44,
		Element:        "Fire/Void",
		ChakraCenter:   "Sahasrara (Crown)",
		ColorHex:       "#71717a",
		HermeticAxiom:  "The Principle of Polarity: Dissolution of forms into origin.",
		StoryArchetype: "The Mystic / The Ascetic: Detachment, ego dissolution, spiritual liberation, and hidden insight.",
	},
	{
		ID:             PlanetVenus,
		Name:           "Venus",
		SanskritName:   "Shukra",
		DurationYears:  20,
		RootFrequency:  221.23,
		Element:        "Water",
		ChakraCenter:   "Swadhisthana (Sacral)",
		ColorHex:       "#ec4899",
		HermeticAxiom:  "The Principle of Gender: Creative synthesis and magnetism.",
		StoryArchetype: "The Creator / The Lover: Refinement, aesthetics, harmony, prosperity, and devotion to beauty.",
	},
	{
		ID:             PlanetSun,
		Name:           "Sun",
		SanskritName:   "Surya",
		DurationYears:  6,
		RootFrequency:  126.22,
		Element:        "Fire",
		ChakraCenter:   "Manipura (Solar Plexus)",
		ColorHex:       "#f59e0b",
		HermeticAxiom:  "The Principle of Mentalism: The central radiant consciousness.",
		StoryArchetype: "The Sovereign / The Hero: Self-realization, vitality, authority, inner radiance, and illumination.",
	},
	{
		ID:             PlanetMoon,
		Name:           "Moon",
		SanskritName:   "Chandra",
		DurationYears:  10,
		RootFrequency:  210.42,
		Element:        "Water",
		ChakraCenter:   "Ajna (Third Eye)",
		ColorHex:       "#06b6d4",
		HermeticAxiom:  "The Principle of Rhythm: The tidal ebb and flow of perception.",
		StoryArchetype: "The Weaver / The Dreamer: Memory, psychic receptivity, emotional nourishment, and subconscious currents.",
	},
	{
		ID:             PlanetMars,
		Name:           "Mars",
		SanskritName:   "Mangala",
		DurationYears:  7,
		RootFrequency:  144.72,
		Element:        "Fire",
		ChakraCenter:   "Muladhara (Root)",
		ColorHex:       "#ef4444",
		HermeticAxiom:  "The Principle of Cause and Effect: Will in catalytic motion.",
		StoryArchetype: "The Warrior / The Catalyst: Courage, initiative, conquest, dynamic drive, and transformative friction.",
	},
	{
		ID:             PlanetRahu,
		Name:           "Rahu",
		SanskritName:   "Rahu (North Node)",
		DurationYears:  18,
		RootFrequency:  187.61,
		Element:        "Smoke/Shadow",
		ChakraCenter:   "Astral Bridge",
		ColorHex:       "#6366f1",
		HermeticAxiom:  "The Principle of Vibration: Amplification and radical disruption.",
		StoryArchetype: "The Alchemist / The Trickster: Ambition, taboo exploration, obsessive breakthroughs, and boundary transgression.",
	},
	{
		ID:             PlanetJupiter,
		Name:           "Jupiter",
		SanskritName:   "Guru",
		DurationYears:  16,
		RootFrequency:  183.58,
		Element:        "Ether",
		ChakraCenter:   "Anahata (Heart)",
		ColorHex:       "#eab308",
		HermeticAxiom:  "The Principle of Correspondence: Expansion across micro and macrocosm.",
		StoryArchetype: "The Hierophant / The Sage: Wisdom, ethical architecture, grace, mentorship, and universal expansion.",
	},
	{
		ID:             PlanetSaturn,
		Name:           "Saturn",
		SanskritName:   "Shani",
		DurationYears:  19,
		RootFrequency:  147.85,
		Element:        "Air/Cold",
		ChakraCenter:   "Muladhara Base",
		ColorHex:       "#3b82f6",
		HermeticAxiom:  "The Principle of Cause and Effect: The strict reckoning of Chronos.",
		StoryArchetype: "The Architect / The Warden: Discipline, endurance, karmic consolidation, structural integrity, and mastery through time.",
	},
	{
		ID:             PlanetMercury,
		Name:           "Mercury",
		SanskritName:   "Budha",
		DurationYears:  17,
		RootFrequency:  141.27,
		Element:        "Earth/Air",
		ChakraCenter:   "Vishuddha (Throat)",
		ColorHex:       "#10b981",
		HermeticAxiom:  "The Principle of Mentalism: Quick transmigration of ideas and logos.",
		StoryArchetype: "The Messenger / The Polymath: Intellectual dexterity, discernment, commerce, code, and mercurial synthesis.",
	},
}

// PlanetsByID maps PlanetID to Planet for fast O(1) lookup.
var PlanetsByID = func() map[PlanetID]Planet {
	m := make(map[PlanetID]Planet, len(VimshottariPlanets))
	for _, p := range VimshottariPlanets {
		m[p.ID] = p
	}
	return m
}()

// NakshatraNames contains the canonical 27 lunar mansions in zodiacal order.
var Nakshatras = []Nakshatra{
	// 1. Ashwini (00°00' - 13°20' Aries)
	{
		Index: 1, ID: "ashwini", Name: "Ashwini", SanskritName: "Aśvinī",
		DegreeStart: 0.0, DegreeEnd: 13.333333333333334, ZodiacSpan: "00°00' - 13°20' Aries",
		RulingPlanet: PlanetKetu, FrequencyHz: 288.0,
		Deity: "Ashvins (Twin Celestial Physicians)", Symbol: "Horse's Head",
		Quality: "Swift, Healing, Pioneering Initiative",
		Padas: []Pada{
			{Number: 1, DegreesStart: "00°00' Aries", DegreesEnd: "03°20' Aries", NavamshaSign: "Aries"},
			{Number: 2, DegreesStart: "03°20' Aries", DegreesEnd: "06°40' Aries", NavamshaSign: "Taurus"},
			{Number: 3, DegreesStart: "06°40' Aries", DegreesEnd: "10°00' Aries", NavamshaSign: "Gemini"},
			{Number: 4, DegreesStart: "10°00' Aries", DegreesEnd: "13°20' Aries", NavamshaSign: "Cancer"},
		},
	},
	// 2. Bharani (13°20' - 26°40' Aries)
	{
		Index: 2, ID: "bharani", Name: "Bharani", SanskritName: "Bharaṇī",
		DegreeStart: 13.333333333333334, DegreeEnd: 26.666666666666668, ZodiacSpan: "13°20' - 26°40' Aries",
		RulingPlanet: PlanetVenus, FrequencyHz: 320.0,
		Deity: "Yama (Lord of Death & Dharma)", Symbol: "Yoni / Vulva",
		Quality: "Restraint, Transformation, Bearer of Fire",
		Padas: []Pada{
			{Number: 1, DegreesStart: "13°20' Aries", DegreesEnd: "16°40' Aries", NavamshaSign: "Leo"},
			{Number: 2, DegreesStart: "16°40' Aries", DegreesEnd: "20°00' Aries", NavamshaSign: "Virgo"},
			{Number: 3, DegreesStart: "20°00' Aries", DegreesEnd: "23°20' Aries", NavamshaSign: "Libra"},
			{Number: 4, DegreesStart: "23°20' Aries", DegreesEnd: "26°40' Aries", NavamshaSign: "Scorpio"},
		},
	},
	// 3. Krittika (26°40' Aries - 10°00' Taurus)
	{
		Index: 3, ID: "krittika", Name: "Krittika", SanskritName: "Kṛttikā",
		DegreeStart: 26.666666666666668, DegreeEnd: 40.0, ZodiacSpan: "26°40' Aries - 10°00' Taurus",
		RulingPlanet: PlanetSun, FrequencyHz: 341.3,
		Deity: "Agni (God of Sacred Fire)", Symbol: "Razor / Knife / Flame",
		Quality: "Penetrating, Purifying, Razor-Sharp Discernment",
		Padas: []Pada{
			{Number: 1, DegreesStart: "26°40' Aries", DegreesEnd: "30°00' Aries", NavamshaSign: "Sagittarius"},
			{Number: 2, DegreesStart: "00°00' Taurus", DegreesEnd: "03°20' Taurus", NavamshaSign: "Capricorn"},
			{Number: 3, DegreesStart: "03°20' Taurus", DegreesEnd: "06°40' Taurus", NavamshaSign: "Aquarius"},
			{Number: 4, DegreesStart: "06°40' Taurus", DegreesEnd: "10°00' Taurus", NavamshaSign: "Pisces"},
		},
	},
	// 4. Rohini (10°00' - 23°20' Taurus)
	{
		Index: 4, ID: "rohini", Name: "Rohini", SanskritName: "Rohiṇī",
		DegreeStart: 40.0, DegreeEnd: 53.333333333333336, ZodiacSpan: "10°00' - 23°20' Taurus",
		RulingPlanet: PlanetMoon, FrequencyHz: 360.0,
		Deity: "Brahma / Prajapati (The Creator)", Symbol: "Chariot / Temple / Cart",
		Quality: "Fertility, Sensory Beauty, Seductive Growth",
		Padas: []Pada{
			{Number: 1, DegreesStart: "10°00' Taurus", DegreesEnd: "13°20' Taurus", NavamshaSign: "Aries"},
			{Number: 2, DegreesStart: "13°20' Taurus", DegreesEnd: "16°40' Taurus", NavamshaSign: "Taurus"},
			{Number: 3, DegreesStart: "16°40' Taurus", DegreesEnd: "20°00' Taurus", NavamshaSign: "Gemini"},
			{Number: 4, DegreesStart: "20°00' Taurus", DegreesEnd: "23°20' Taurus", NavamshaSign: "Cancer"},
		},
	},
	// 5. Mrigashira (23°20' Taurus - 06°40' Gemini)
	{
		Index: 5, ID: "mrigashira", Name: "Mrigashira", SanskritName: "Mṛgaśīrṣā",
		DegreeStart: 53.333333333333336, DegreeEnd: 66.66666666666667, ZodiacSpan: "23°20' Taurus - 06°40' Gemini",
		RulingPlanet: PlanetMars, FrequencyHz: 384.0,
		Deity: "Soma (Moon God of Nectar)", Symbol: "Deer's Head",
		Quality: "Seeking, Inquisitive, Roaming Explorer",
		Padas: []Pada{
			{Number: 1, DegreesStart: "23°20' Taurus", DegreesEnd: "26°40' Taurus", NavamshaSign: "Leo"},
			{Number: 2, DegreesStart: "26°40' Taurus", DegreesEnd: "30°00' Taurus", NavamshaSign: "Virgo"},
			{Number: 3, DegreesStart: "00°00' Gemini", DegreesEnd: "03°20' Gemini", NavamshaSign: "Libra"},
			{Number: 4, DegreesStart: "03°20' Gemini", DegreesEnd: "06°40' Gemini", NavamshaSign: "Scorpio"},
		},
	},
	// 6. Ardra (06°40' - 20°00' Gemini)
	{
		Index: 6, ID: "ardra", Name: "Ardra", SanskritName: "Ārdrā",
		DegreeStart: 66.66666666666667, DegreeEnd: 80.0, ZodiacSpan: "06°40' - 20°00' Gemini",
		RulingPlanet: PlanetRahu, FrequencyHz: 405.0,
		Deity: "Rudra (God of the Storm)", Symbol: "Teardrop / Diamond",
		Quality: "Stormy, Cathartic, Destruction of Illusion",
		Padas: []Pada{
			{Number: 1, DegreesStart: "06°40' Gemini", DegreesEnd: "10°00' Gemini", NavamshaSign: "Sagittarius"},
			{Number: 2, DegreesStart: "10°00' Gemini", DegreesEnd: "13°20' Gemini", NavamshaSign: "Capricorn"},
			{Number: 3, DegreesStart: "13°20' Gemini", DegreesEnd: "16°40' Gemini", NavamshaSign: "Aquarius"},
			{Number: 4, DegreesStart: "16°40' Gemini", DegreesEnd: "20°00' Gemini", NavamshaSign: "Pisces"},
		},
	},
	// 7. Punarvasu (20°00' Gemini - 03°20' Cancer)
	{
		Index: 7, ID: "punarvasu", Name: "Punarvasu", SanskritName: "Punarvasu",
		DegreeStart: 80.0, DegreeEnd: 93.33333333333333, ZodiacSpan: "20°00' Gemini - 03°20' Cancer",
		RulingPlanet: PlanetJupiter, FrequencyHz: 432.0,
		Deity: "Aditi (Cosmic Mother of Gods)", Symbol: "Bow & Quiver of Arrows",
		Quality: "Renewal, Return of Light, Restoration",
		Padas: []Pada{
			{Number: 1, DegreesStart: "20°00' Gemini", DegreesEnd: "23°20' Gemini", NavamshaSign: "Aries"},
			{Number: 2, DegreesStart: "23°20' Gemini", DegreesEnd: "26°40' Gemini", NavamshaSign: "Taurus"},
			{Number: 3, DegreesStart: "26°40' Gemini", DegreesEnd: "30°00' Gemini", NavamshaSign: "Gemini"},
			{Number: 4, DegreesStart: "00°00' Cancer", DegreesEnd: "03°20' Cancer", NavamshaSign: "Cancer"},
		},
	},
	// 8. Pushya (03°20' - 16°40' Cancer)
	{
		Index: 8, ID: "pushya", Name: "Pushya", SanskritName: "Puṣya",
		DegreeStart: 93.33333333333333, DegreeEnd: 106.66666666666667, ZodiacSpan: "03°20' - 16°40' Cancer",
		RulingPlanet: PlanetSaturn, FrequencyHz: 480.0,
		Deity: "Brihaspati (Guru of the Celestials)", Symbol: "Cow's Udder / Lotus Flower",
		Quality: "Supreme Nourishment, Spiritual Auspiciousness",
		Padas: []Pada{
			{Number: 1, DegreesStart: "03°20' Cancer", DegreesEnd: "06°40' Cancer", NavamshaSign: "Leo"},
			{Number: 2, DegreesStart: "06°40' Cancer", DegreesEnd: "10°00' Cancer", NavamshaSign: "Virgo"},
			{Number: 3, DegreesStart: "10°00' Cancer", DegreesEnd: "13°20' Cancer", NavamshaSign: "Libra"},
			{Number: 4, DegreesStart: "13°20' Cancer", DegreesEnd: "16°40' Cancer", NavamshaSign: "Scorpio"},
		},
	},
	// 9. Ashlesha (16°40' - 30°00' Cancer) - Mercury Ruled
	{
		Index: 9, ID: "ashlesha", Name: "Ashlesha", SanskritName: "Āśleṣā",
		DegreeStart: 106.66666666666667, DegreeEnd: 120.0, ZodiacSpan: "16°40' - 30°00' Cancer",
		RulingPlanet: PlanetMercury, FrequencyHz: 432.0,
		Deity: "Nagas (Serpent Wisdom & Kundalini)", Symbol: "Coiled Serpent",
		Quality: "Sharp, Penetrating Intuition, Mystical Binding",
		Padas: []Pada{
			{Number: 1, DegreesStart: "16°40' Cancer", DegreesEnd: "20°00' Cancer", NavamshaSign: "Sagittarius"},
			{Number: 2, DegreesStart: "20°00' Cancer", DegreesEnd: "23°20' Cancer", NavamshaSign: "Capricorn"},
			{Number: 3, DegreesStart: "23°20' Cancer", DegreesEnd: "26°40' Cancer", NavamshaSign: "Aquarius"},
			{Number: 4, DegreesStart: "26°40' Cancer", DegreesEnd: "30°00' Cancer", NavamshaSign: "Pisces"},
		},
	},
	// 10. Magha (00°00' - 13°20' Leo)
	{
		Index: 10, ID: "magha", Name: "Magha", SanskritName: "Maghā",
		DegreeStart: 120.0, DegreeEnd: 133.33333333333334, ZodiacSpan: "00°00' - 13°20' Leo",
		RulingPlanet: PlanetKetu, FrequencyHz: 512.0,
		Deity: "Pitris (Ancestral Spirits / Forefathers)", Symbol: "Royal Throne Chamber",
		Quality: "Lineage, Nobility, Ancestral Legacy",
		Padas: []Pada{
			{Number: 1, DegreesStart: "00°00' Leo", DegreesEnd: "03°20' Leo", NavamshaSign: "Aries"},
			{Number: 2, DegreesStart: "03°20' Leo", DegreesEnd: "06°40' Leo", NavamshaSign: "Taurus"},
			{Number: 3, DegreesStart: "06°40' Leo", DegreesEnd: "10°00' Leo", NavamshaSign: "Gemini"},
			{Number: 4, DegreesStart: "10°00' Leo", DegreesEnd: "13°20' Leo", NavamshaSign: "Cancer"},
		},
	},
	// 11. Purva Phalguni (13°20' - 26°40' Leo)
	{
		Index: 11, ID: "purva_phalguni", Name: "Purva Phalguni", SanskritName: "Pūrva Phālgunī",
		DegreeStart: 133.33333333333334, DegreeEnd: 146.66666666666666, ZodiacSpan: "13°20' - 26°40' Leo",
		RulingPlanet: PlanetVenus, FrequencyHz: 528.0,
		Deity: "Bhaga (God of Delight & Wealth)", Symbol: "Hammock / Front Legs of Bed",
		Quality: "Creative Passion, Conjugal Union, Leisure",
		Padas: []Pada{
			{Number: 1, DegreesStart: "13°20' Leo", DegreesEnd: "16°40' Leo", NavamshaSign: "Leo"},
			{Number: 2, DegreesStart: "16°40' Leo", DegreesEnd: "20°00' Leo", NavamshaSign: "Virgo"},
			{Number: 3, DegreesStart: "20°00' Leo", DegreesEnd: "23°20' Leo", NavamshaSign: "Libra"},
			{Number: 4, DegreesStart: "23°20' Leo", DegreesEnd: "26°40' Leo", NavamshaSign: "Scorpio"},
		},
	},
	// 12. Uttara Phalguni (26°40' Leo - 10°00' Virgo)
	{
		Index: 12, ID: "uttara_phalguni", Name: "Uttara Phalguni", SanskritName: "Uttara Phālgunī",
		DegreeStart: 146.66666666666666, DegreeEnd: 160.0, ZodiacSpan: "26°40' Leo - 10°00' Virgo",
		RulingPlanet: PlanetSun, FrequencyHz: 576.0,
		Deity: "Aryaman (God of Patronage & Alliances)", Symbol: "Four Legs of Bed",
		Quality: "Honor, Lasting Contracts, Philanthropy",
		Padas: []Pada{
			{Number: 1, DegreesStart: "26°40' Leo", DegreesEnd: "30°00' Leo", NavamshaSign: "Sagittarius"},
			{Number: 2, DegreesStart: "00°00' Virgo", DegreesEnd: "03°20' Virgo", NavamshaSign: "Capricorn"},
			{Number: 3, DegreesStart: "03°20' Virgo", DegreesEnd: "06°40' Virgo", NavamshaSign: "Aquarius"},
			{Number: 4, DegreesStart: "06°40' Virgo", DegreesEnd: "10°00' Virgo", NavamshaSign: "Pisces"},
		},
	},
	// 13. Hasta (10°00' - 23°20' Virgo)
	{
		Index: 13, ID: "hasta", Name: "Hasta", SanskritName: "Hasta",
		DegreeStart: 160.0, DegreeEnd: 173.33333333333334, ZodiacSpan: "10°00' - 23°20' Virgo",
		RulingPlanet: PlanetMoon, FrequencyHz: 600.0,
		Deity: "Savitr (Solar Artisan of Dawn)", Symbol: "Open Hand / Fist",
		Quality: "Dexterity, Craftsmanship, Manifesting Skill",
		Padas: []Pada{
			{Number: 1, DegreesStart: "10°00' Virgo", DegreesEnd: "13°20' Virgo", NavamshaSign: "Aries"},
			{Number: 2, DegreesStart: "13°20' Virgo", DegreesEnd: "16°40' Virgo", NavamshaSign: "Taurus"},
			{Number: 3, DegreesStart: "16°40' Virgo", DegreesEnd: "20°00' Virgo", NavamshaSign: "Gemini"},
			{Number: 4, DegreesStart: "20°00' Virgo", DegreesEnd: "23°20' Virgo", NavamshaSign: "Cancer"},
		},
	},
	// 14. Chitra (23°20' Virgo - 06°40' Libra)
	{
		Index: 14, ID: "chitra", Name: "Chitra", SanskritName: "Citrā",
		DegreeStart: 173.33333333333334, DegreeEnd: 186.66666666666666, ZodiacSpan: "23°20' Virgo - 06°40' Libra",
		RulingPlanet: PlanetMars, FrequencyHz: 640.0,
		Deity: "Vishvakarma (Divine Celestial Architect)", Symbol: "Gleaming Jewel / Pearl",
		Quality: "Brilliance, Visual Form, Illusion Mastery",
		Padas: []Pada{
			{Number: 1, DegreesStart: "23°20' Virgo", DegreesEnd: "26°40' Virgo", NavamshaSign: "Leo"},
			{Number: 2, DegreesStart: "26°40' Virgo", DegreesEnd: "30°00' Virgo", NavamshaSign: "Virgo"},
			{Number: 3, DegreesStart: "00°00' Libra", DegreesEnd: "03°20' Libra", NavamshaSign: "Libra"},
			{Number: 4, DegreesStart: "03°20' Libra", DegreesEnd: "06°40' Libra", NavamshaSign: "Scorpio"},
		},
	},
	// 15. Swati (06°40' - 20°00' Libra)
	{
		Index: 15, ID: "swati", Name: "Swati", SanskritName: "Svātī",
		DegreeStart: 186.66666666666666, DegreeEnd: 200.0, ZodiacSpan: "06°40' - 20°00' Libra",
		RulingPlanet: PlanetRahu, FrequencyHz: 672.0,
		Deity: "Vayu (Wind God of Prana)", Symbol: "Shoot of Plant Bending in Wind",
		Quality: "Independence, Flexibility, Movement, Breath",
		Padas: []Pada{
			{Number: 1, DegreesStart: "06°40' Libra", DegreesEnd: "10°00' Libra", NavamshaSign: "Sagittarius"},
			{Number: 2, DegreesStart: "10°00' Libra", DegreesEnd: "13°20' Libra", NavamshaSign: "Capricorn"},
			{Number: 3, DegreesStart: "13°20' Libra", DegreesEnd: "16°40' Libra", NavamshaSign: "Aquarius"},
			{Number: 4, DegreesStart: "16°40' Libra", DegreesEnd: "20°00' Libra", NavamshaSign: "Pisces"},
		},
	},
	// 16. Vishakha (20°00' Libra - 03°20' Scorpio)
	{
		Index: 16, ID: "vishakha", Name: "Vishakha", SanskritName: "Viśākhā",
		DegreeStart: 200.0, DegreeEnd: 213.33333333333334, ZodiacSpan: "20°00' Libra - 03°20' Scorpio",
		RulingPlanet: PlanetJupiter, FrequencyHz: 720.0,
		Deity: "Indra & Agni (Chieftain & Sacred Fire)", Symbol: "Triumphal Arch / Potter's Wheel",
		Quality: "Single-Pointed Focus, Goal Triumph",
		Padas: []Pada{
			{Number: 1, DegreesStart: "20°00' Libra", DegreesEnd: "23°20' Libra", NavamshaSign: "Aries"},
			{Number: 2, DegreesStart: "23°20' Libra", DegreesEnd: "26°40' Libra", NavamshaSign: "Taurus"},
			{Number: 3, DegreesStart: "26°40' Libra", DegreesEnd: "30°00' Libra", NavamshaSign: "Gemini"},
			{Number: 4, DegreesStart: "00°00' Scorpio", DegreesEnd: "03°20' Scorpio", NavamshaSign: "Cancer"},
		},
	},
	// 17. Anuradha (03°20' - 16°40' Scorpio)
	{
		Index: 17, ID: "anuradha", Name: "Anuradha", SanskritName: "Anurādhā",
		DegreeStart: 213.33333333333334, DegreeEnd: 226.66666666666666, ZodiacSpan: "03°20' - 16°40' Scorpio",
		RulingPlanet: PlanetSaturn, FrequencyHz: 768.0,
		Deity: "Mitra (God of Divine Friendship & Light)", Symbol: "Row of Furrows / Staff",
		Quality: "Devotion, Organizational Fellowship, Esoteric Travel",
		Padas: []Pada{
			{Number: 1, DegreesStart: "03°20' Scorpio", DegreesEnd: "06°40' Scorpio", NavamshaSign: "Leo"},
			{Number: 2, DegreesStart: "06°40' Scorpio", DegreesEnd: "10°00' Scorpio", NavamshaSign: "Virgo"},
			{Number: 3, DegreesStart: "10°00' Scorpio", DegreesEnd: "13°20' Scorpio", NavamshaSign: "Libra"},
			{Number: 4, DegreesStart: "13°20' Scorpio", DegreesEnd: "16°40' Scorpio", NavamshaSign: "Scorpio"},
		},
	},
	// 18. Jyeshtha (16°40' - 30°00' Scorpio) - Mercury Ruled
	{
		Index: 18, ID: "jyeshtha", Name: "Jyeshtha", SanskritName: "Jyeṣṭhā",
		DegreeStart: 226.66666666666666, DegreeEnd: 240.0, ZodiacSpan: "16°40' - 30°00' Scorpio",
		RulingPlanet: PlanetMercury, FrequencyHz: 528.0,
		Deity: "Indra (King of Gods & Dragon Slayer)", Symbol: "Circular Talisman / Umbrella",
		Quality: "Mastery, Seniority, Directorial Sovereignty",
		Padas: []Pada{
			{Number: 1, DegreesStart: "16°40' Scorpio", DegreesEnd: "20°00' Scorpio", NavamshaSign: "Sagittarius"},
			{Number: 2, DegreesStart: "20°00' Scorpio", DegreesEnd: "23°20' Scorpio", NavamshaSign: "Capricorn"},
			{Number: 3, DegreesStart: "23°20' Scorpio", DegreesEnd: "26°40' Scorpio", NavamshaSign: "Aquarius"},
			{Number: 4, DegreesStart: "26°40' Scorpio", DegreesEnd: "30°00' Scorpio", NavamshaSign: "Pisces"},
		},
	},
	// 19. Mula (00°00' - 13°20' Sagittarius)
	{
		Index: 19, ID: "mula", Name: "Mula", SanskritName: "Mūla",
		DegreeStart: 240.0, DegreeEnd: 253.33333333333334, ZodiacSpan: "00°00' - 13°20' Sagittarius",
		RulingPlanet: PlanetKetu, FrequencyHz: 810.0,
		Deity: "Nirriti (Goddess of Dissolution & Calamity)", Symbol: "Tied Bunch of Roots",
		Quality: "Root Extraction, Uprooting, Galactic Center",
		Padas: []Pada{
			{Number: 1, DegreesStart: "00°00' Sagittarius", DegreesEnd: "03°20' Sagittarius", NavamshaSign: "Aries"},
			{Number: 2, DegreesStart: "03°20' Sagittarius", DegreesEnd: "06°40' Sagittarius", NavamshaSign: "Taurus"},
			{Number: 3, DegreesStart: "06°40' Sagittarius", DegreesEnd: "10°00' Sagittarius", NavamshaSign: "Gemini"},
			{Number: 4, DegreesStart: "10°00' Sagittarius", DegreesEnd: "13°20' Sagittarius", NavamshaSign: "Cancer"},
		},
	},
	// 20. Purva Ashadha (13°20' - 26°40' Sagittarius)
	{
		Index: 20, ID: "purva_ashadha", Name: "Purva Ashadha", SanskritName: "Pūrva Aṣāḍhā",
		DegreeStart: 253.33333333333334, DegreeEnd: 266.6666666666667, ZodiacSpan: "13°20' - 26°40' Sagittarius",
		RulingPlanet: PlanetVenus, FrequencyHz: 864.0,
		Deity: "Apas (Cosmic Waters of Life)", Symbol: "Winnowing Basket / Elephant Tusk",
		Quality: "Invincible Declaration, Purification, Cleansing",
		Padas: []Pada{
			{Number: 1, DegreesStart: "13°20' Sagittarius", DegreesEnd: "16°40' Sagittarius", NavamshaSign: "Leo"},
			{Number: 2, DegreesStart: "16°40' Sagittarius", DegreesEnd: "20°00' Sagittarius", NavamshaSign: "Virgo"},
			{Number: 3, DegreesStart: "20°00' Sagittarius", DegreesEnd: "23°20' Sagittarius", NavamshaSign: "Libra"},
			{Number: 4, DegreesStart: "23°20' Sagittarius", DegreesEnd: "26°40' Sagittarius", NavamshaSign: "Scorpio"},
		},
	},
	// 21. Uttara Ashadha (26°40' Sagittarius - 10°00' Capricorn)
	{
		Index: 21, ID: "uttara_ashadha", Name: "Uttara Ashadha", SanskritName: "Uttara Aṣāḍhā",
		DegreeStart: 266.6666666666667, DegreeEnd: 280.0, ZodiacSpan: "26°40' Sagittarius - 10°00' Capricorn",
		RulingPlanet: PlanetSun, FrequencyHz: 900.0,
		Deity: "Vishvadevas (Universal Cosmic Gods)", Symbol: "Small Planks / Elephant Tusk",
		Quality: "Enduring Victory, Righteous Conquest, Stability",
		Padas: []Pada{
			{Number: 1, DegreesStart: "26°40' Sagittarius", DegreesEnd: "30°00' Sagittarius", NavamshaSign: "Sagittarius"},
			{Number: 2, DegreesStart: "00°00' Capricorn", DegreesEnd: "03°20' Capricorn", NavamshaSign: "Capricorn"},
			{Number: 3, DegreesStart: "03°20' Capricorn", DegreesEnd: "06°40' Capricorn", NavamshaSign: "Aquarius"},
			{Number: 4, DegreesStart: "06°40' Capricorn", DegreesEnd: "10°00' Capricorn", NavamshaSign: "Pisces"},
		},
	},
	// 22. Shravana (10°00' - 23°20' Capricorn)
	{
		Index: 22, ID: "shravana", Name: "Shravana", SanskritName: "Śravaṇa",
		DegreeStart: 280.0, DegreeEnd: 293.3333333333333, ZodiacSpan: "10°00' - 23°20' Capricorn",
		RulingPlanet: PlanetMoon, FrequencyHz: 960.0,
		Deity: "Vishnu (Preserver of the Universe)", Symbol: "Ear / Three Footprints",
		Quality: "Deep Listening, Oral Wisdom Tradition, Preservation",
		Padas: []Pada{
			{Number: 1, DegreesStart: "10°00' Capricorn", DegreesEnd: "13°20' Capricorn", NavamshaSign: "Aries"},
			{Number: 2, DegreesStart: "13°20' Capricorn", DegreesEnd: "16°40' Capricorn", NavamshaSign: "Taurus"},
			{Number: 3, DegreesStart: "16°40' Capricorn", DegreesEnd: "20°00' Capricorn", NavamshaSign: "Gemini"},
			{Number: 4, DegreesStart: "20°00' Capricorn", DegreesEnd: "23°20' Capricorn", NavamshaSign: "Cancer"},
		},
	},
	// 23. Dhanishta (23°20' Capricorn - 06°40' Aquarius)
	{
		Index: 23, ID: "dhanishta", Name: "Dhanishta", SanskritName: "Dhaniṣṭhā",
		DegreeStart: 293.3333333333333, DegreeEnd: 306.6666666666667, ZodiacSpan: "23°20' Capricorn - 06°40' Aquarius",
		RulingPlanet: PlanetMars, FrequencyHz: 1024.0,
		Deity: "Eight Vasus (Elemental Gods of Abundance)", Symbol: "Musical Drum (Damaru) / Flute",
		Quality: "Rhythm, Symphonic Wealth, Martial Syncopation",
		Padas: []Pada{
			{Number: 1, DegreesStart: "23°20' Capricorn", DegreesEnd: "26°40' Capricorn", NavamshaSign: "Leo"},
			{Number: 2, DegreesStart: "26°40' Capricorn", DegreesEnd: "30°00' Capricorn", NavamshaSign: "Virgo"},
			{Number: 3, DegreesStart: "00°00' Aquarius", DegreesEnd: "03°20' Aquarius", NavamshaSign: "Libra"},
			{Number: 4, DegreesStart: "03°20' Aquarius", DegreesEnd: "06°40' Aquarius", NavamshaSign: "Scorpio"},
		},
	},
	// 24. Shatabhisha (06°40' - 20°00' Aquarius)
	{
		Index: 24, ID: "shatabhisha", Name: "Shatabhisha", SanskritName: "Śatabhiṣaj",
		DegreeStart: 306.6666666666667, DegreeEnd: 320.0, ZodiacSpan: "06°40' - 20°00' Aquarius",
		RulingPlanet: PlanetRahu, FrequencyHz: 1080.0,
		Deity: "Varuna (God of Waters & Celestial Oceans)", Symbol: "Empty Circle / Hundred Physicians",
		Quality: "Enigmatic, Cosmic Veil, Esoteric Medicine",
		Padas: []Pada{
			{Number: 1, DegreesStart: "06°40' Aquarius", DegreesEnd: "10°00' Aquarius", NavamshaSign: "Sagittarius"},
			{Number: 2, DegreesStart: "10°00' Aquarius", DegreesEnd: "13°20' Aquarius", NavamshaSign: "Capricorn"},
			{Number: 3, DegreesStart: "13°20' Aquarius", DegreesEnd: "16°40' Aquarius", NavamshaSign: "Aquarius"},
			{Number: 4, DegreesStart: "16°40' Aquarius", DegreesEnd: "20°00' Aquarius", NavamshaSign: "Pisces"},
		},
	},
	// 25. Purva Bhadrapada (20°00' Aquarius - 03°20' Pisces)
	{
		Index: 25, ID: "purva_bhadrapada", Name: "Purva Bhadrapada", SanskritName: "Pūrva Bhādrapadā",
		DegreeStart: 320.0, DegreeEnd: 333.3333333333333, ZodiacSpan: "20°00' Aquarius - 03°20' Pisces",
		RulingPlanet: PlanetJupiter, FrequencyHz: 1152.0,
		Deity: "Aja Ekapada (One-Footed Cosmic Serpent)", Symbol: "Swords / Two-Faced Man / Front of Funeral Cot",
		Quality: "Fierce Spiritual Penance, Kundalini Fire",
		Padas: []Pada{
			{Number: 1, DegreesStart: "20°00' Aquarius", DegreesEnd: "23°20' Aquarius", NavamshaSign: "Aries"},
			{Number: 2, DegreesStart: "23°20' Aquarius", DegreesEnd: "26°40' Aquarius", NavamshaSign: "Taurus"},
			{Number: 3, DegreesStart: "26°40' Aquarius", DegreesEnd: "30°00' Aquarius", NavamshaSign: "Gemini"},
			{Number: 4, DegreesStart: "00°00' Pisces", DegreesEnd: "03°20' Pisces", NavamshaSign: "Cancer"},
		},
	},
	// 26. Uttara Bhadrapada (03°20' - 16°40' Pisces)
	{
		Index: 26, ID: "uttara_bhadrapada", Name: "Uttara Bhadrapada", SanskritName: "Uttara Bhādrapadā",
		DegreeStart: 333.3333333333333, DegreeEnd: 346.6666666666667, ZodiacSpan: "03°20' - 16°40' Pisces",
		RulingPlanet: PlanetSaturn, FrequencyHz: 1200.0,
		Deity: "Ahirbudhnya (Serpent of the Oceanic Depths)", Symbol: "Back of Funeral Cot / Twin in Depths",
		Quality: "Wisdom of the Depths, Solitary Contemplation",
		Padas: []Pada{
			{Number: 1, DegreesStart: "03°20' Pisces", DegreesEnd: "06°40' Pisces", NavamshaSign: "Leo"},
			{Number: 2, DegreesStart: "06°40' Pisces", DegreesEnd: "10°00' Pisces", NavamshaSign: "Virgo"},
			{Number: 3, DegreesStart: "10°00' Pisces", DegreesEnd: "13°20' Pisces", NavamshaSign: "Libra"},
			{Number: 4, DegreesStart: "13°20' Pisces", DegreesEnd: "16°40' Pisces", NavamshaSign: "Scorpio"},
		},
	},
	// 27. Revati (16°40' - 30°00' Pisces) - Mercury Ruled
	{
		Index: 27, ID: "revati", Name: "Revati", SanskritName: "Revatī",
		DegreeStart: 346.6666666666667, DegreeEnd: 360.0, ZodiacSpan: "16°40' - 30°00' Pisces",
		RulingPlanet: PlanetMercury, FrequencyHz: 639.0,
		Deity: "Pushan (Nourisher of Flocks & Safe Conductor)", Symbol: "Fish Swimming in Opposite Directions",
		Quality: "Journey's End, Transcendence, Safe Passage",
		Padas: []Pada{
			{Number: 1, DegreesStart: "16°40' Pisces", DegreesEnd: "20°00' Pisces", NavamshaSign: "Sagittarius"},
			{Number: 2, DegreesStart: "20°00' Pisces", DegreesEnd: "23°20' Pisces", NavamshaSign: "Capricorn"},
			{Number: 3, DegreesStart: "23°20' Pisces", DegreesEnd: "26°40' Pisces", NavamshaSign: "Aquarius"},
			{Number: 4, DegreesStart: "26°40' Pisces", DegreesEnd: "30°00' Pisces", NavamshaSign: "Pisces"},
		},
	},
}

// GetNakshatraByIndex returns Nakshatra by 1-based index (1 to 27).
func GetNakshatraByIndex(index int) (Nakshatra, bool) {
	if index < 1 || index > 27 {
		return Nakshatra{}, false
	}
	return Nakshatras[index-1], true
}

// GetNakshatraByDegree finds which Nakshatra encompasses the given degree (0.0 to 360.0).
func GetNakshatraByDegree(deg float64) (Nakshatra, int, float64) {
	// Normalize degree to [0, 360)
	for deg < 0 {
		deg += 360
	}
	for deg >= 360 {
		deg -= 360
	}

	nakSpan := 360.0 / 27.0 // 13.333333333333334
	idx := int(deg / nakSpan)
	if idx >= 27 {
		idx = 26
	}

	nak := Nakshatras[idx]
	traversed := deg - nak.DegreeStart

	padaSpan := nakSpan / 4.0 // 3.3333333333333335
	padaNum := int(traversed/padaSpan) + 1
	if padaNum > 4 {
		padaNum = 4
	}

	return nak, padaNum, traversed
}
