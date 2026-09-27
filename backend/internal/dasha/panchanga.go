package dasha

import (
	"fmt"
	"math"
	"time"
)

// CalculateSiderealSun computes the Sun's sidereal longitude [0, 360) using Meeus solar elements and Lahiri Ayanamsha.
func CalculateSiderealSun(t time.Time) float64 {
	year := t.Year()
	month := int(t.Month())
	day := float64(t.Day()) + (float64(t.Hour()) / 24.0) + (float64(t.Minute()) / 1440.0) + (float64(t.Second()) / 86400.0)

	if month <= 2 {
		year -= 1
		month += 12
	}
	A := math.Floor(float64(year) / 100.0)
	B := 2.0 - A + math.Floor(A/4.0)
	jd := math.Floor(365.25*float64(year+4716)) + math.Floor(30.6001*float64(month+1)) + day + B - 1524.5

	// Julian centuries from J2000.0
	T := (jd - 2451545.0) / 36525.0

	// Mean Solar Longitude L0 (degrees)
	L0 := math.Mod(280.46646+T*(36000.76983+T*0.0003032), 360.0)
	if L0 < 0 {
		L0 += 360.0
	}

	// Mean Solar Anomaly M (degrees)
	M := 357.52911 + T*(35999.05029-0.0001537*T)
	Mrad := M * (math.Pi / 180.0)

	// Equation of Center C (degrees)
	C := math.Sin(Mrad)*(1.914602-T*(0.004817+0.000014*T)) +
		math.Sin(2.0*Mrad)*(0.019993-0.000101*T) +
		math.Sin(3.0*Mrad)*0.000289

	sunTrueLong := L0 + C

	// Apparent aberration
	omega := 125.04 - 1934.136*T
	lambda := sunTrueLong - 0.00569 - 0.00478*math.Sin(omega*(math.Pi/180.0))

	// Lahiri Ayanamsha (matching Moon ephemeris)
	ayanamsha := 23.85709 + (50.29/3600.0)*(T*100.0)

	siderealSun := lambda - ayanamsha
	for siderealSun < 0 {
		siderealSun += 360.0
	}
	for siderealSun >= 360.0 {
		siderealSun -= 360.0
	}
	return siderealSun
}

var varaNames = []string{
	"Ravivara (Sunday)", "Somavara (Monday)", "Mangalavara (Tuesday)",
	"Budhavara (Wednesday)", "Guruvara (Thursday)", "Shukravara (Friday)", "Shanivara (Saturday)",
}

var varaLords = []string{
	"Sun (Surya)", "Moon (Chandra)", "Mars (Mangala)",
	"Mercury (Budha)", "Jupiter (Guru)", "Venus (Shukra)", "Saturn (Shani)",
}

var tithiNames = []string{
	"Pratipada", "Dvitiya", "Tritiya", "Chaturthi", "Panchami",
	"Shashti", "Saptami", "Ashtami", "Navami", "Dashami",
	"Ekadashi", "Dvadashi", "Trayodashi", "Chaturdashi", "Purnima",
	"Pratipada", "Dvitiya", "Tritiya", "Chaturthi", "Panchami",
	"Shashti", "Saptami", "Ashtami", "Navami", "Dashami",
	"Ekadashi", "Dvadashi", "Trayodashi", "Chaturdashi", "Amavasya",
}

type yogaEntry struct {
	Name    string
	Meaning string
}

var solilunarYogas = []yogaEntry{
	{"Vishkambha", "Malefic - Obstacle & Supported Pillar"},
	{"Priti", "Benefic - Love, Mutual Affection & Joy"},
	{"Ayushman", "Benefic - Longevity, Vitality & Health"},
	{"Saubhagya", "Benefic - Sovereign Prosperity & Fortune"},
	{"Shobhana", "Benefic - Radiance, Elegance & Virtue"},
	{"Atiganda", "Malefic - Severe Obstacles & Trial"},
	{"Sukarma", "Benefic - Auspicious Works & Righteous Deeds"},
	{"Dhriti", "Benefic - Fortitude, Courage & Constancy"},
	{"Shula", "Malefic - Sharp Weapon & Piercing Strife"},
	{"Ganda", "Malefic - Perilous Knot & Challenge"},
	{"Vriddhi", "Benefic - Growth, Abundance & Expansion"},
	{"Dhruva", "Benefic - Firmness, Permanence & Reliability"},
	{"Vyaghata", "Malefic - Fierce Impact & Striking Blows"},
	{"Harshana", "Benefic - Delight, Merriment & Rejoicing"},
	{"Vajra", "Malefic - Thunderbolt & Indomitable Force"},
	{"Siddhi", "Benefic - Accomplishment, Attainment & Mastery"},
	{"Vyatipata", "Malefic - Calamity & Intense Cataclysm"},
	{"Variyan", "Benefic - Supremacy, Eminence & Comfort"},
	{"Parigha", "Malefic - Iron Barricade & Obstruction"},
	{"Shiva", "Benefic - Auspiciousness, Transcendence & Peace"},
	{"Siddha", "Benefic - Perfection, Enlightenment & Skill"},
	{"Sadhya", "Benefic - Attainable Fulfillment & Accomplishment"},
	{"Shubha", "Benefic - Purity, Grace & Auspicious Radiance"},
	{"Shukla", "Benefic - Brightness, Luminescence & Virtue"},
	{"Brahma", "Benefic - Creative Essence & Universal Truth"},
	{"Indra", "Benefic - Sovereign Leadership & Honor"},
	{"Vaidhriti", "Malefic - Critical Divergence & Subversion"},
}

// ResolvePanchanga derives the 5 cosmic limbs of time for a given birth moment and planetary longitudes.
func ResolvePanchanga(birthUTC time.Time, moonDeg, sunDeg float64) VedicPanchanga {
	// 1. Vara (Day Lord)
	weekday := int(birthUTC.Weekday())
	vara := varaNames[weekday]
	varaLord := varaLords[weekday]

	// 2. Tithi (Lunar Phase)
	diff := moonDeg - sunDeg
	for diff < 0 {
		diff += 360.0
	}
	for diff >= 360.0 {
		diff -= 360.0
	}

	tithiNumber := int(diff/12.0) + 1
	if tithiNumber > 30 {
		tithiNumber = 30
	}
	tithiPercent := math.Mod(diff, 12.0) / 12.0 * 100.0

	paksha := "Shukla"
	if tithiNumber > 15 {
		paksha = "Krishna"
	}

	rawTithiName := tithiNames[tithiNumber-1]
	tithiName := fmt.Sprintf("%s %s", paksha, rawTithiName)
	if tithiNumber == 15 {
		tithiName = "Purnima (Full Moon)"
	} else if tithiNumber == 30 {
		tithiName = "Amavasya (New Moon)"
	}

	// 3. Yoga (Solilunar angular sum, 27 intervals of 13°20')
	sumDeg := math.Mod(sunDeg+moonDeg, 360.0)
	if sumDeg < 0 {
		sumDeg += 360.0
	}
	yogaSpan := 360.0 / 27.0 // 13.333333333333334 deg
	yogaNumber := int(sumDeg/yogaSpan) + 1
	if yogaNumber > 27 {
		yogaNumber = 27
	}
	yoga := solilunarYogas[yogaNumber-1]

	// 4. Karana (Half-Tithi, 60 half-tithis per lunar cycle)
	karanaNumber := int(diff/6.0) + 1
	if karanaNumber > 60 {
		karanaNumber = 60
	}

	var karanaName, karanaType string
	switch {
	case karanaNumber == 1:
		karanaName = "Kimstughna"
		karanaType = "Fixed"
	case karanaNumber >= 2 && karanaNumber <= 57:
		movableKaranas := []string{
			"Bava", "Balava", "Kaulava", "Taitila", "Gara", "Vanija", "Vishti (Bhadra)",
		}
		idx := (karanaNumber - 2) % 7
		karanaName = movableKaranas[idx]
		karanaType = "Movable"
	case karanaNumber == 58:
		karanaName = "Shakuni"
		karanaType = "Fixed"
	case karanaNumber == 59:
		karanaName = "Chatushpada"
		karanaType = "Fixed"
	case karanaNumber == 60:
		karanaName = "Naga"
		karanaType = "Fixed"
	}

	return VedicPanchanga{
		Vara:              vara,
		VaraLord:          varaLord,
		TithiNumber:       tithiNumber,
		TithiName:         tithiName,
		Paksha:            paksha,
		TithiPercent:      math.Round(tithiPercent*100) / 100,
		YogaNumber:        yogaNumber,
		YogaName:          yoga.Name,
		YogaMeaning:       yoga.Meaning,
		KaranaNumber:      karanaNumber,
		KaranaName:        karanaName,
		KaranaType:        karanaType,
		SiderealSunDegree: math.Round(sunDeg*100) / 100,
	}
}

type nakshatraAyurveda struct {
	Dosha          string
	DoshaQualities string
	Gana           string
	YoniTotem      string
	YoniAnimal     string
	Nadi           string
}

// Canonical Ayurvedic and Totemic classifications for the 27 Nakshatras.
var nakshatraAyurvedaTable = map[int]nakshatraAyurveda{
	1:  {"Vata", "Dynamic, kinetic, initiating, swift", "Deva", "Horse (Ashwa)", "Horse", "Adi"},
	2:  {"Pitta", "Transformative, fiery, catalytic, intense", "Manushya", "Elephant (Gaja)", "Elephant", "Madhya"},
	3:  {"Kapha", "Grounding, cohesive, radiant, sustaining", "Rakshasa", "Sheep / Goat (Mesha)", "Goat", "Antya"},
	4:  {"Kapha", "Fertile, nurturing, stable, magnetic", "Manushya", "Serpent (Sarpa)", "Serpent", "Antya"},
	5:  {"Pitta", "Inquisitive, searching, adaptable, agile", "Deva", "Serpent (Sarpa)", "Serpent", "Madhya"},
	6:  {"Vata", "Electrifying, storm-like, sharp, penetrating", "Manushya", "Dog (Shwana)", "Dog", "Adi"},
	7:  {"Vata", "Restorative, buoyant, expansive, benevolent", "Deva", "Cat (Marjara)", "Cat", "Adi"},
	8:  {"Pitta", "Nourishing, solidifying, disciplined, vital", "Deva", "Sheep / Goat (Mesha)", "Goat", "Madhya"},
	9:  {"Kapha", "Mystical, hypnotic, binding, deep", "Rakshasa", "Cat (Marjara)", "Cat", "Antya"},
	10: {"Kapha", "Regal, ancestral, authoritative, grounded", "Rakshasa", "Rat / Mouse (Mushaka)", "Rat", "Antya"},
	11: {"Pitta", "Creative, prosperous, vibrant, joyful", "Manushya", "Rat / Mouse (Mushaka)", "Rat", "Madhya"},
	12: {"Vata", "Generous, structured, luminous, discerning", "Manushya", "Cow / Bull (Go)", "Cow", "Adi"},
	13: {"Vata", "Dexterous, artisanal, agile, precise", "Deva", "Buffalo (Mahisha)", "Buffalo", "Adi"},
	14: {"Pitta", "Architectural, brilliant, faceted, bold", "Rakshasa", "Tiger (Vyaghra)", "Tiger", "Madhya"},
	15: {"Vata", "Independent, airborne, intuitive, fluid", "Deva", "Buffalo (Mahisha)", "Buffalo", "Antya"},
	16: {"Pitta", "Goal-oriented, triumphant, determined, intense", "Rakshasa", "Tiger (Vyaghra)", "Tiger", "Antya"},
	17: {"Vata", "Devotional, resilient, harmonizing, noble", "Deva", "Deer / Hare (Mriga)", "Deer", "Madhya"},
	18: {"Pitta", "Protective, commanding, sovereign, keen", "Rakshasa", "Deer / Hare (Mriga)", "Deer", "Adi"},
	19: {"Vata", "Root-penetrating, unearthing, radical, raw", "Rakshasa", "Dog (Shwana)", "Dog", "Adi"},
	20: {"Pitta", "Invincible, purifying, flowing, proud", "Manushya", "Monkey (Vanara)", "Monkey", "Madhya"},
	21: {"Kapha", "Enduring, victorious, principled, anchored", "Manushya", "Mongoose (Nakula)", "Mongoose", "Antya"},
	22: {"Kapha", "Receptive, acoustic, wise, persevering", "Deva", "Monkey (Vanara)", "Monkey", "Antya"},
	23: {"Pitta", "Resonant, rhythmical, abundant, martial", "Rakshasa", "Lioness (Simha)", "Lion", "Madhya"},
	24: {"Vata", "Cosmic, meditative, healing, veil-piercing", "Rakshasa", "Horse (Ashwa)", "Horse", "Adi"},
	25: {"Vata", "Ascetic, transformative, fiery-spiritual", "Manushya", "Lion (Simha)", "Lion", "Adi"},
	26: {"Pitta", "Deep-sea, stabilizing, profound, sacrificial", "Manushya", "Cow / Bull (Go)", "Cow", "Madhya"},
	27: {"Kapha", "Nourishing, protective, guiding, boundless", "Deva", "Elephant (Gaja)", "Elephant", "Antya"},
}

// ResolveAyurvedicConstitution maps a Janma Nakshatra index (1-27) to its biological and character triad.
func ResolveAyurvedicConstitution(nakshatraIndex int) AyurvedicConstitution {
	if nakshatraIndex < 1 || nakshatraIndex > 27 {
		nakshatraIndex = 1
	}
	entry, ok := nakshatraAyurvedaTable[nakshatraIndex]
	if !ok {
		entry = nakshatraAyurvedaTable[1]
	}

	return AyurvedicConstitution{
		Dosha:          entry.Dosha,
		DoshaQualities: entry.DoshaQualities,
		Gana:           entry.Gana,
		YoniTotem:      entry.YoniTotem,
		YoniAnimal:     entry.YoniAnimal,
		Nadi:           entry.Nadi,
	}
}
