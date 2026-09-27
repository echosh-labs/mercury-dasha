package dasha

import (
	"fmt"
	"math"
	"time"
)

// RashiDefinition represents one of the 12 classical sidereal zodiac signs.
type RashiDefinition struct {
	Index        int      `json:"index"`         // 1 to 12
	Name         string   `json:"name"`          // e.g. "Mesha (Aries)"
	SanskritName string   `json:"sanskrit_name"` // e.g. "Mesha"
	EnglishName  string   `json:"english_name"`  // e.g. "Aries"
	Lord         PlanetID `json:"lord"`          // Ruling planet ID
	LordName     string   `json:"lord_name"`     // Display name
	Element      string   `json:"element"`       // Fire, Earth, Air, Water
	Symbol       string   `json:"symbol"`        // Zodiac glyph
}

// Classical 12 Sidereal Rashis
var RashisByIndex = map[int]RashiDefinition{
	1:  {Index: 1, Name: "Mesha (Aries)", SanskritName: "Mesha", EnglishName: "Aries", Lord: PlanetMars, LordName: "Mars (Mangala)", Element: "Fire", Symbol: "♈"},
	2:  {Index: 2, Name: "Vrishabha (Taurus)", SanskritName: "Vrishabha", EnglishName: "Taurus", Lord: PlanetVenus, LordName: "Venus (Shukra)", Element: "Earth", Symbol: "♉"},
	3:  {Index: 3, Name: "Mithuna (Gemini)", SanskritName: "Mithuna", EnglishName: "Gemini", Lord: PlanetMercury, LordName: "Mercury (Budha)", Element: "Air", Symbol: "♊"},
	4:  {Index: 4, Name: "Karka (Cancer)", SanskritName: "Karka", EnglishName: "Cancer", Lord: PlanetMoon, LordName: "Moon (Chandra)", Element: "Water", Symbol: "♋"},
	5:  {Index: 5, Name: "Simha (Leo)", SanskritName: "Simha", EnglishName: "Leo", Lord: PlanetSun, LordName: "Sun (Surya)", Element: "Fire", Symbol: "♌"},
	6:  {Index: 6, Name: "Kanya (Virgo)", SanskritName: "Kanya", EnglishName: "Virgo", Lord: PlanetMercury, LordName: "Mercury (Budha)", Element: "Earth", Symbol: "♍"},
	7:  {Index: 7, Name: "Tula (Libra)", SanskritName: "Tula", EnglishName: "Libra", Lord: PlanetVenus, LordName: "Venus (Shukra)", Element: "Air", Symbol: "♎"},
	8:  {Index: 8, Name: "Vrishchika (Scorpio)", SanskritName: "Vrishchika", EnglishName: "Scorpio", Lord: PlanetMars, LordName: "Mars (Mangala)", Element: "Water", Symbol: "♏"},
	9:  {Index: 9, Name: "Dhanu (Sagittarius)", SanskritName: "Dhanu", EnglishName: "Sagittarius", Lord: PlanetJupiter, LordName: "Jupiter (Guru)", Element: "Fire", Symbol: "♐"},
	10: {Index: 10, Name: "Makara (Capricorn)", SanskritName: "Makara", EnglishName: "Capricorn", Lord: PlanetSaturn, LordName: "Saturn (Shani)", Element: "Earth", Symbol: "♑"},
	11: {Index: 11, Name: "Kumbha (Aquarius)", SanskritName: "Kumbha", EnglishName: "Aquarius", Lord: PlanetSaturn, LordName: "Saturn (Shani)", Element: "Air", Symbol: "♒"},
	12: {Index: 12, Name: "Meena (Pisces)", SanskritName: "Meena", EnglishName: "Pisces", Lord: PlanetJupiter, LordName: "Jupiter (Guru)", Element: "Water", Symbol: "♓"},
}

// Classical 12 Bhavas (Houses)
var BhavaMetadata = []struct {
	Number       int
	Name         string
	SanskritName string
	Domain       string
}{
	{1, "Tanu (1st)", "Tanu", "Self, Physical Vitality, Sovereignty & Outer Form"},
	{2, "Dhana (2nd)", "Dhana", "Wealth, Material Resources, Speech, Values & Family"},
	{3, "Sahaja (3rd)", "Sahaja", "Courage, Direct Initiative, Communication & Creative Drive"},
	{4, "Sukha (4th)", "Sukha", "Emotional Sanctuary, Inner Peace, Home & Foundational Ground"},
	{5, "Putra (5th)", "Putra", "Creative Genius, Sovereign Intelligence, Dharma & Divine Insight"},
	{6, "Ari (6th)", "Ari", "Daily Mastery, Problem Resolution, Healing & Dynamic Resilience"},
	{7, "Yuvati (7th)", "Yuvati", "Sovereign Partnerships, Conscious Complementarity & The Mirror"},
	{8, "Randhra (8th)", "Randhra", "Occult Insight, Transformation, Regeneration & Hidden Power"},
	{9, "Bhagya (9th)", "Bhagya", "Higher Wisdom, Sovereign Philosophy, Cosmic Grace & Guru"},
	{10, "Karma (10th)", "Karma", "Executive Mastery, Public Works, Sovereign Stature & Career"},
	{11, "Labha (11th)", "Labha", "Divine Abundance, Overflowing Fruition, Community & Aspirations"},
	{12, "Vyaya (12th)", "Vyaya", "Spiritual Transcendence, Solitude, Subconscious Vision & Liberation"},
}

// CalculateSiderealLagna calculates the topocentric Sidereal Ascendant degree [0, 360) using spherical trigonometry and Lahiri Ayanamsha.
func CalculateSiderealLagna(birthUTC time.Time, lat, lon float64) float64 {
	year := birthUTC.Year()
	month := int(birthUTC.Month())
	day := float64(birthUTC.Day()) + (float64(birthUTC.Hour()) / 24.0) + (float64(birthUTC.Minute()) / 1440.0) + (float64(birthUTC.Second()) / 86400.0)

	if month <= 2 {
		year -= 1
		month += 12
	}
	A := math.Floor(float64(year) / 100.0)
	B := 2.0 - A + math.Floor(A/4.0)
	jd := math.Floor(365.25*float64(year+4716)) + math.Floor(30.6001*float64(month+1)) + day + B - 1524.5

	// Julian centuries from J2000.0
	T := (jd - 2451545.0) / 36525.0

	// Greenwich Mean Sidereal Time in degrees (Meeus formula)
	gmst := 280.46061837 + 360.98564736629*(jd-2451545.0) + 0.000387933*T*T - (T*T*T)/38710000.0
	gmst = math.Mod(gmst, 360.0)
	if gmst < 0 {
		gmst += 360.0
	}

	// Local Sidereal Time = GMST + Longitude (East positive)
	lst := math.Mod(gmst+lon, 360.0)
	if lst < 0 {
		lst += 360.0
	}

	// Obliquity of the Ecliptic eps
	eps := 23.439291 - 0.0130042*T

	deg2rad := math.Pi / 180.0
	rad2deg := 180.0 / math.Pi

	lstRad := lst * deg2rad
	latRad := lat * deg2rad
	epsRad := eps * deg2rad

	// Standard Ascendant spherical formula:
	// y = -cos(LST)
	// x = sin(LST)*cos(eps) + tan(lat)*sin(eps)
	yComp := -math.Cos(lstRad)
	xComp := math.Sin(lstRad)*math.Cos(epsRad) + math.Tan(latRad)*math.Sin(epsRad)

	ascRad := math.Atan2(yComp, xComp)
	ascTrop := ascRad * rad2deg
	for ascTrop < 0 {
		ascTrop += 360.0
	}
	for ascTrop >= 360.0 {
		ascTrop -= 360.0
	}

	// Canonical Lahiri Ayanamsha: 23°51'25.5" at J2000 with 50.29"/yr precession
	ayanamsha := 23.85709 + (50.29/3600.0)*(T*100.0)

	siderealLagna := ascTrop - ayanamsha
	for siderealLagna < 0 {
		siderealLagna += 360.0
	}
	for siderealLagna >= 360.0 {
		siderealLagna -= 360.0
	}

	return math.Round(siderealLagna*1000) / 1000
}

// ResolveNatalLagna resolves the complete NatalLagna matrix from an exact Sidereal Ascendant degree.
func ResolveNatalLagna(lagnaDeg float64) NatalLagna {
	for lagnaDeg < 0 {
		lagnaDeg += 360.0
	}
	for lagnaDeg >= 360.0 {
		lagnaDeg -= 360.0
	}

	rashiIdx := int(math.Floor(lagnaDeg/30.0)) + 1
	if rashiIdx < 1 {
		rashiIdx = 1
	}
	if rashiIdx > 12 {
		rashiIdx = 12
	}

	rashi := RashisByIndex[rashiIdx]
	degInSign := lagnaDeg - float64(rashiIdx-1)*30.0

	risingNak, padaNum, _ := GetNakshatraByDegree(lagnaDeg)

	return NatalLagna{
		Degree:        lagnaDeg,
		Rashi:         rashi.Name,
		RashiSanskrit: rashi.SanskritName,
		RashiIndex:    rashiIdx,
		DegreeInSign:  math.Round(degInSign*100) / 100,
		Lord:          rashi.Lord,
		LordName:      rashi.LordName,
		Nakshatra:     risingNak,
		Pada:          padaNum,
		Symbol:        rashi.Symbol,
		Element:       rashi.Element,
	}
}

// ResolveLuminaryPlacement computes the sign, house, and nakshatra placement of a luminary (Sun or Moon) from Lagna.
func ResolveLuminaryPlacement(degree float64, lagnaRashiIdx int) LuminaryPlacement {
	for degree < 0 {
		degree += 360.0
	}
	for degree >= 360.0 {
		degree -= 360.0
	}

	rashiIdx := int(math.Floor(degree/30.0)) + 1
	if rashiIdx < 1 {
		rashiIdx = 1
	}
	if rashiIdx > 12 {
		rashiIdx = 12
	}

	rashi := RashisByIndex[rashiIdx]
	degInSign := degree - float64(rashiIdx-1)*30.0

	// Whole Sign House: (rashiIdx - lagnaRashiIdx + 12) % 12 + 1
	houseNum := (rashiIdx - lagnaRashiIdx + 12) % 12
	if houseNum == 0 {
		houseNum = 12
	}

	houseName := BhavaMetadata[houseNum-1].Name
	nak, padaNum, _ := GetNakshatraByDegree(degree)

	return LuminaryPlacement{
		Degree:       degree,
		Rashi:        rashi.Name,
		RashiIndex:   rashiIdx,
		DegreeInSign: math.Round(degInSign*100) / 100,
		RashiLord:    rashi.Lord,
		HouseNumber:  houseNum,
		HouseName:    houseName,
		Nakshatra:    nak,
		Pada:         padaNum,
	}
}

// ResolveTripodAndBhavas constructs the unified Tripod of Embodiment (Lagna, Surya, Chandra) and all 12 Bhavas.
func ResolveTripodAndBhavas(birthUTC time.Time, lat, lon, moonDeg, sunDeg float64) TripodOfEmbodiment {
	// Fallback to Niagara / St. Catharines sovereign genesis coordinates if unprovided
	if lat == 0 && lon == 0 {
		lat = 43.1594
		lon = -79.2469
	}

	lagnaDeg := CalculateSiderealLagna(birthUTC, lat, lon)
	lagna := ResolveNatalLagna(lagnaDeg)

	surya := ResolveLuminaryPlacement(sunDeg, lagna.RashiIndex)
	chandra := ResolveLuminaryPlacement(moonDeg, lagna.RashiIndex)

	// Build 12 Bhavas
	bhavas := make([]BhavaDetail, 12)
	for i := 0; i < 12; i++ {
		houseNum := i + 1
		meta := BhavaMetadata[i]

		// Rashi for this house under Whole Sign system
		rashiIdx := (lagna.RashiIndex - 1 + i) % 12 + 1
		rashi := RashisByIndex[rashiIdx]

		var luminaries []string
		if houseNum == 1 {
			luminaries = append(luminaries, fmt.Sprintf("Lagna (%s %.1f°)", lagna.RashiSanskrit, lagna.DegreeInSign))
		}
		if surya.HouseNumber == houseNum {
			luminaries = append(luminaries, fmt.Sprintf("Surya/Sun (%.1f°)", surya.DegreeInSign))
		}
		if chandra.HouseNumber == houseNum {
			luminaries = append(luminaries, fmt.Sprintf("Chandra/Moon (%.1f°)", chandra.DegreeInSign))
		}

		bhavas[i] = BhavaDetail{
			HouseNumber:  houseNum,
			Name:         meta.Name,
			SanskritName: meta.SanskritName,
			Domain:       meta.Domain,
			Rashi:        rashi.Name,
			RashiLord:    rashi.LordName,
			Luminaries:   luminaries,
		}
	}

	return TripodOfEmbodiment{
		Lagna:   lagna,
		Surya:   surya,
		Chandra: chandra,
		Bhavas:  bhavas,
	}
}

// ResolveTripodAndBhavasWithGrahas constructs the unified Tripod of Embodiment with all 9 residing Grahas in the Bhavas.
func ResolveTripodAndBhavasWithGrahas(birthUTC time.Time, lat, lon, moonDeg, sunDeg float64, grahas []GrahaPlacement) TripodOfEmbodiment {
	if lat == 0 && lon == 0 {
		lat = 43.1594
		lon = -79.2469
	}

	lagnaDeg := CalculateSiderealLagna(birthUTC, lat, lon)
	lagna := ResolveNatalLagna(lagnaDeg)

	surya := ResolveLuminaryPlacement(sunDeg, lagna.RashiIndex)
	chandra := ResolveLuminaryPlacement(moonDeg, lagna.RashiIndex)

	bhavas := make([]BhavaDetail, 12)
	for i := 0; i < 12; i++ {
		houseNum := i + 1
		meta := BhavaMetadata[i]

		rashiIdx := (lagna.RashiIndex - 1 + i) % 12 + 1
		rashi := RashisByIndex[rashiIdx]

		var luminaries []string
		if houseNum == 1 {
			luminaries = append(luminaries, fmt.Sprintf("Lagna (%s %.1f°)", lagna.RashiSanskrit, lagna.DegreeInSign))
		}

		for _, g := range grahas {
			if g.HouseNumber == houseNum {
				luminaries = append(luminaries, fmt.Sprintf("%s (%.1f° • %s)", g.Name, g.DegreeInSign, g.Dignity.Name))
			}
		}

		bhavas[i] = BhavaDetail{
			HouseNumber:  houseNum,
			Name:         meta.Name,
			SanskritName: meta.SanskritName,
			Domain:       meta.Domain,
			Rashi:        rashi.Name,
			RashiLord:    rashi.LordName,
			Luminaries:   luminaries,
		}
	}

	return TripodOfEmbodiment{
		Lagna:   lagna,
		Surya:   surya,
		Chandra: chandra,
		Bhavas:  bhavas,
	}
}
