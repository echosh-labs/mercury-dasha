package dasha_test

import (
	"math"
	"testing"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
)

func TestPlanetsSumTo120Years(t *testing.T) {
	var total float64
	for _, p := range dasha.VimshottariPlanets {
		total += p.DurationYears
	}
	if math.Abs(total-120.0) > 0.001 {
		t.Fatalf("expected planets to sum to 120 years, got %f", total)
	}
	if len(dasha.VimshottariPlanets) != 9 {
		t.Fatalf("expected 9 planets, got %d", len(dasha.VimshottariPlanets))
	}
}

func TestNakshatrasCoverage(t *testing.T) {
	if len(dasha.Nakshatras) != 27 {
		t.Fatalf("expected 27 nakshatras, got %d", len(dasha.Nakshatras))
	}

	for i, n := range dasha.Nakshatras {
		if n.Index != i+1 {
			t.Errorf("nakshatra index mismatch: got %d, expected %d", n.Index, i+1)
		}
		if len(n.Padas) != 4 {
			t.Errorf("nakshatra %s has %d padas, expected 4", n.Name, len(n.Padas))
		}
		if i > 0 {
			prev := dasha.Nakshatras[i-1]
			if math.Abs(prev.DegreeEnd-n.DegreeStart) > 0.0001 {
				t.Errorf("gap between %s (end %f) and %s (start %f)", prev.Name, prev.DegreeEnd, n.Name, n.DegreeStart)
			}
		}
	}

	last := dasha.Nakshatras[26]
	if math.Abs(last.DegreeEnd-360.0) > 0.0001 {
		t.Errorf("last nakshatra end degree should be 360, got %f", last.DegreeEnd)
	}
}

func TestGetNakshatraByDegree(t *testing.T) {
	// 0.0 -> Ashwini, Pada 1
	n, pada, trav := dasha.GetNakshatraByDegree(0.0)
	if n.ID != "ashwini" || pada != 1 || trav != 0.0 {
		t.Errorf("expected ashwini pada 1 for 0.0 deg, got %s pada %d (trav %f)", n.ID, pada, trav)
	}

	// 108.0 -> Ashlesha (starts at 106.6667)
	n2, pada2, _ := dasha.GetNakshatraByDegree(108.0)
	if n2.ID != "ashlesha" {
		t.Errorf("expected ashlesha for 108 deg, got %s (pada %d)", n2.ID, pada2)
	}
}

func TestCalculateDashaTimeline(t *testing.T) {
	req := dasha.CalculationRequest{
		Name:           "Test-Chart",
		BirthDate:      "1990-05-15",
		BirthTime:      "14:30",
		TimezoneOffset: -4.0, // EDT
		NakshatraIndex: 9,    // Ashlesha (Mercury Ruled)
		PadaNumber:     2,
	}

	profile, err := dasha.CalculateFromRequest(req)
	if err != nil {
		t.Fatalf("calculation failed: %v", err)
	}

	if profile.JanmaNakshatra.ID != "ashlesha" {
		t.Errorf("expected Ashlesha, got %s", profile.JanmaNakshatra.ID)
	}
	if profile.StartingLord != dasha.PlanetMercury {
		t.Errorf("expected Mercury as starting lord, got %s", profile.StartingLord)
	}
	if profile.BirthDate != "1990-05-15" {
		t.Errorf("expected BirthDate 1990-05-15, got %s", profile.BirthDate)
	}
	if profile.BirthTime != "14:30" {
		t.Errorf("expected BirthTime 14:30, got %s", profile.BirthTime)
	}
	if profile.TimezoneOffset != -4.0 {
		t.Errorf("expected TimezoneOffset -4.0, got %f", profile.TimezoneOffset)
	}
	if profile.Mode != "nakshatra" {
		t.Errorf("expected Mode nakshatra, got %s", profile.Mode)
	}

	if len(profile.Timeline) != 9 {
		t.Fatalf("expected 9 Mahadashas, got %d", len(profile.Timeline))
	}

	// First Mahadasha should be Mercury
	firstMaha := profile.Timeline[0]
	if firstMaha.Planet != dasha.PlanetMercury {
		t.Errorf("expected first Mahadasha to be Mercury, got %s", firstMaha.Planet)
	}

	// Verify Antardashas exist
	if len(firstMaha.SubPeriods) == 0 {
		t.Errorf("expected Antardashas in first Mahadasha")
	}

	// Verify active snapshot is resolved
	snap := profile.ActiveSnapshot
	if snap.Mahadasha.Planet == "" {
		t.Errorf("active Mahadasha should not be empty")
	}
	if snap.Antardasha.Planet == "" {
		t.Errorf("active Antardasha should not be empty")
	}
	if snap.StoryContext.NarrativePrompt == "" {
		t.Errorf("story context narrative prompt should not be empty")
	}
}

func TestEphemerisCalculation(t *testing.T) {
	// Test known astronomical epoch: 2000-01-01 12:00 UTC
	tEpoch := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
	deg := dasha.CalculateSiderealMoon(tEpoch)

	if deg < 0 || deg >= 360 {
		t.Fatalf("sidereal moon longitude out of range: %f", deg)
	}

	nak, pada, _ := dasha.GetNakshatraByDegree(deg)
	t.Logf("Sidereal Moon at J2000: %.3f° (%s, Pada %d)", deg, nak.Name, pada)
}

func TestSovereignProfileUnification(t *testing.T) {
	req := dasha.CalculationRequest{
		Name:           "Sovereign Genesis",
		BirthDate:      "1992-06-15",
		BirthTime:      "12:00",
		TimezoneOffset: -4.0,
		Latitude:       43.1594,
		Longitude:      -79.2469,
	}
	prof, err := dasha.CalculateFromRequest(req)
	if err != nil {
		t.Fatalf("failed to calculate sovereign profile: %v", err)
	}

	// 1. Verify Natal Astrology Matrix
	if prof.Astrology.JanmaNakshatra.ID == "" {
		t.Errorf("expected populated JanmaNakshatra in Astrology matrix")
	}
	if prof.Astrology.StartingLord == "" {
		t.Errorf("expected StartingLord in Astrology matrix")
	}
	if prof.Astrology.ElementalTattva == "" {
		t.Errorf("expected ElementalTattva in Astrology matrix")
	}

	// 2. Verify Natal Alchemical Constitution
	if prof.Alchemy.SacredMetal.ID == "" {
		t.Errorf("expected SacredMetal in Alchemy matrix")
	}
	if prof.Alchemy.GoverningAxiom.Title == "" {
		t.Errorf("expected GoverningAxiom in Alchemy matrix")
	}
	if prof.Alchemy.MagnumOpusStage.Name == "" {
		t.Errorf("expected MagnumOpusStage in Alchemy matrix")
	}
	if prof.Alchemy.ChakraAnchor == "" {
		t.Errorf("expected ChakraAnchor in Alchemy matrix")
	}

	// 3. Verify Active Alchemical State
	if prof.ActiveAlchemy.MahadashaMetal.ID == "" {
		t.Errorf("expected MahadashaMetal in ActiveAlchemy")
	}
	if prof.ActiveAlchemy.AntardashaMetal.ID == "" {
		t.Errorf("expected AntardashaMetal in ActiveAlchemy")
	}

	// 4. Verify Timeline Summary
	if len(prof.TimelineSummary) != 9 {
		t.Errorf("expected 9 Mahadashas in TimelineSummary, got %d", len(prof.TimelineSummary))
	}

	// 5. Verify Symbiotic Resonance
	// Test when transit matches birth lord
	resJanma := dasha.ResolveSymbioticResonance(prof, prof.StartingLord)
	if !resJanma.IsJanmaResonant {
		t.Errorf("expected IsJanmaResonant to be true")
	}
	if resJanma.ResonanceTier != "SOVEREIGN_JANMA_CONJUNCTION" {
		t.Errorf("expected SOVEREIGN_JANMA_CONJUNCTION tier, got %s", resJanma.ResonanceTier)
	}

	// Test when transit matches Mahadasha lord
	mahaLord := prof.ActiveSnapshot.Mahadasha.Planet
	resMaha := dasha.ResolveSymbioticResonance(prof, mahaLord)
	if !resMaha.IsMahadashaResonant {
		t.Errorf("expected IsMahadashaResonant to be true")
	}

	// 6. Verify Vedic Panchanga Matrix
	panchanga := prof.Astrology.Panchanga
	if panchanga.Vara == "" || panchanga.VaraLord == "" {
		t.Errorf("expected non-empty Vara and VaraLord in Panchanga, got '%s'/'%s'", panchanga.Vara, panchanga.VaraLord)
	}
	if panchanga.TithiNumber < 1 || panchanga.TithiNumber > 30 || panchanga.TithiName == "" {
		t.Errorf("expected valid Tithi in Panchanga, got %d (%s)", panchanga.TithiNumber, panchanga.TithiName)
	}
	if panchanga.YogaNumber < 1 || panchanga.YogaNumber > 27 || panchanga.YogaName == "" {
		t.Errorf("expected valid Yoga in Panchanga, got %d (%s)", panchanga.YogaNumber, panchanga.YogaName)
	}
	if panchanga.KaranaNumber < 1 || panchanga.KaranaNumber > 60 || panchanga.KaranaName == "" {
		t.Errorf("expected valid Karana in Panchanga, got %d (%s)", panchanga.KaranaNumber, panchanga.KaranaName)
	}
	if panchanga.SiderealSunDegree <= 0 || panchanga.SiderealSunDegree >= 360 {
		t.Errorf("expected valid SiderealSunDegree in Panchanga, got %f", panchanga.SiderealSunDegree)
	}

	// 7. Verify Ayurvedic Constitution
	ayurveda := prof.Astrology.Ayurveda
	if ayurveda.Dosha == "" || ayurveda.Gana == "" || ayurveda.YoniTotem == "" || ayurveda.Nadi == "" {
		t.Errorf("expected complete AyurvedicConstitution, got %+v", ayurveda)
	}

	// 8. Verify Natal Lagna and Tripod of Embodiment
	lagna := prof.Astrology.Lagna
	if lagna.Rashi == "" || lagna.Lord == "" || lagna.Nakshatra.ID == "" || lagna.Pada < 1 || lagna.Pada > 4 {
		t.Errorf("expected complete NatalLagna, got %+v", lagna)
	}
	tripod := prof.Astrology.Tripod
	if tripod.Surya.HouseNumber < 1 || tripod.Surya.HouseNumber > 12 || tripod.Surya.Rashi == "" {
		t.Errorf("expected valid Surya placement, got %+v", tripod.Surya)
	}
	if tripod.Chandra.HouseNumber < 1 || tripod.Chandra.HouseNumber > 12 || tripod.Chandra.Rashi == "" {
		t.Errorf("expected valid Chandra placement, got %+v", tripod.Chandra)
	}
	if len(tripod.Bhavas) != 12 {
		t.Errorf("expected 12 Bhavas, got %d", len(tripod.Bhavas))
	}
	if len(tripod.Bhavas[0].Luminaries) == 0 {
		t.Errorf("expected Lagna in Bhava 1 luminaries")
	}

	// 9. Verify 9 Classical Grahas
	if len(prof.Astrology.Grahas) != 9 {
		t.Errorf("expected 9 Grahas in prof.Astrology.Grahas, got %d", len(prof.Astrology.Grahas))
	}
}

func TestVedicPanchangaCalculation(t *testing.T) {
	// J2000.0: 2000-01-01 12:00:00 UTC (Saturday / Shanivara)
	j2000 := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
	sunDeg := dasha.CalculateSiderealSun(j2000)
	moonDeg := dasha.CalculateSiderealMoon(j2000)

	// Around J2000, Tropical Sun is ~280.46°. With Lahiri Ayanamsha ~23.857°, Sidereal Sun should be ~256.6° (Sagittarius).
	if sunDeg < 250.0 || sunDeg > 265.0 {
		t.Errorf("expected J2000 Sidereal Sun around 256.6°, got %f°", sunDeg)
	}

	panchanga := dasha.ResolvePanchanga(j2000, moonDeg, sunDeg)
	if panchanga.Vara != "Shanivara (Saturday)" {
		t.Errorf("expected Shanivara (Saturday), got %s", panchanga.Vara)
	}
	if panchanga.VaraLord != "Saturn (Shani)" {
		t.Errorf("expected Saturn (Shani), got %s", panchanga.VaraLord)
	}
	if panchanga.TithiNumber < 1 || panchanga.TithiNumber > 30 {
		t.Errorf("invalid Tithi number: %d", panchanga.TithiNumber)
	}
	if panchanga.YogaNumber < 1 || panchanga.YogaNumber > 27 {
		t.Errorf("invalid Yoga number: %d", panchanga.YogaNumber)
	}
	if panchanga.KaranaNumber < 1 || panchanga.KaranaNumber > 60 {
		t.Errorf("invalid Karana number: %d", panchanga.KaranaNumber)
	}
}

func TestAyurvedicConstitutionCompleteness(t *testing.T) {
	for i := 1; i <= 27; i++ {
		consti := dasha.ResolveAyurvedicConstitution(i)
		if consti.Dosha == "" {
			t.Errorf("nakshatra %d missing Dosha", i)
		}
		if consti.Dosha != "Vata" && consti.Dosha != "Pitta" && consti.Dosha != "Kapha" {
			t.Errorf("nakshatra %d invalid Dosha %s", i, consti.Dosha)
		}
		if consti.Gana == "" {
			t.Errorf("nakshatra %d missing Gana", i)
		}
		if consti.Gana != "Deva" && consti.Gana != "Manushya" && consti.Gana != "Rakshasa" {
			t.Errorf("nakshatra %d invalid Gana %s", i, consti.Gana)
		}
		if consti.YoniTotem == "" || consti.YoniAnimal == "" {
			t.Errorf("nakshatra %d missing Yoni totem", i)
		}
		if consti.Nadi == "" {
			t.Errorf("nakshatra %d missing Nadi", i)
		}
	}
}

func TestNatalLagnaCalculation(t *testing.T) {
	// J2000.0 at Greenwich (51.48° N, 0° E)
	j2000 := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
	lagnaDeg := dasha.CalculateSiderealLagna(j2000, 51.48, 0.0)

	if lagnaDeg < 0 || lagnaDeg >= 360.0 {
		t.Errorf("invalid Lagna degree: %f", lagnaDeg)
	}

	lagna := dasha.ResolveNatalLagna(lagnaDeg)
	if lagna.RashiIndex < 1 || lagna.RashiIndex > 12 {
		t.Errorf("invalid RashiIndex: %d", lagna.RashiIndex)
	}
	if lagna.Rashi == "" || lagna.Lord == "" {
		t.Errorf("missing Rashi or Lord in NatalLagna: %+v", lagna)
	}
	if lagna.Pada < 1 || lagna.Pada > 4 {
		t.Errorf("invalid Pada: %d", lagna.Pada)
	}
	if lagna.Nakshatra.ID == "" {
		t.Errorf("missing rising Nakshatra in NatalLagna: %+v", lagna)
	}
}

func TestTripodAndBhavasCompleteness(t *testing.T) {
	// St. Catharines coordinates: 43.1594° N, -79.2469° W
	birthUTC := time.Date(1985, 10, 24, 18, 30, 0, 0, time.UTC)
	moonDeg := 305.5 // Dhanishta
	sunDeg := 187.2  // Swati / Libra

	tripod := dasha.ResolveTripodAndBhavas(birthUTC, 43.1594, -79.2469, moonDeg, sunDeg)

	if len(tripod.Bhavas) != 12 {
		t.Fatalf("expected 12 Bhavas, got %d", len(tripod.Bhavas))
	}

	// Verify sequential house numbers 1-12
	for i, bhava := range tripod.Bhavas {
		if bhava.HouseNumber != i+1 {
			t.Errorf("bhava index %d has house number %d, expected %d", i, bhava.HouseNumber, i+1)
		}
		if bhava.Rashi == "" || bhava.Domain == "" || bhava.RashiLord == "" {
			t.Errorf("bhava %d missing fields: %+v", i+1, bhava)
		}
	}

	// Verify Sun and Moon house numbers are within 1 to 12
	if tripod.Surya.HouseNumber < 1 || tripod.Surya.HouseNumber > 12 {
		t.Errorf("invalid Surya house number: %d", tripod.Surya.HouseNumber)
	}
	if tripod.Chandra.HouseNumber < 1 || tripod.Chandra.HouseNumber > 12 {
		t.Errorf("invalid Chandra house number: %d", tripod.Chandra.HouseNumber)
	}

	// House 1 must contain Lagna in Luminaries list
	hasLagna := false
	for _, lum := range tripod.Bhavas[0].Luminaries {
		if len(lum) >= 5 && lum[:5] == "Lagna" {
			hasLagna = true
			break
		}
	}
	if !hasLagna {
		t.Errorf("expected Lagna in Bhava 1 luminaries list, got %v", tripod.Bhavas[0].Luminaries)
	}
}

func TestNavaGrahasCalculation(t *testing.T) {
	birthUTC := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
	sunDeg := dasha.CalculateSiderealSun(birthUTC)
	moonDeg := dasha.CalculateSiderealMoon(birthUTC)
	lagnaDeg := dasha.CalculateSiderealLagna(birthUTC, 51.48, 0.0)
	lagna := dasha.ResolveNatalLagna(lagnaDeg)

	grahas := dasha.ResolveAllGrahas(birthUTC, moonDeg, sunDeg, lagna.RashiIndex)
	if len(grahas) != 9 {
		t.Fatalf("expected 9 Grahas, got %d", len(grahas))
	}

	for _, g := range grahas {
		if g.Degree < 0 || g.Degree >= 360.0 {
			t.Errorf("graha %s invalid degree %f", g.Name, g.Degree)
		}
		if g.RashiIndex < 1 || g.RashiIndex > 12 {
			t.Errorf("graha %s invalid RashiIndex %d", g.Name, g.RashiIndex)
		}
		if g.HouseNumber < 1 || g.HouseNumber > 12 {
			t.Errorf("graha %s invalid HouseNumber %d", g.Name, g.HouseNumber)
		}
		if g.Dignity.Level == "" || g.Dignity.Name == "" {
			t.Errorf("graha %s missing dignity: %+v", g.Name, g.Dignity)
		}
		if g.Nakshatra.ID == "" || g.Pada < 1 || g.Pada > 4 {
			t.Errorf("graha %s invalid nakshatra/pada: %+v", g.Name, g)
		}
	}
}

func TestPlanetaryDignities(t *testing.T) {
	// 1. Sun exalted in Aries (Sign 1)
	sunExalted := dasha.EvaluateGrahaDignity(dasha.PlanetSun, 1, 10.0)
	if sunExalted.Level != "EXALTED" {
		t.Errorf("expected Sun in Aries to be EXALTED, got %s", sunExalted.Level)
	}

	// 2. Sun debilitated in Libra (Sign 7)
	sunDebil := dasha.EvaluateGrahaDignity(dasha.PlanetSun, 7, 10.0)
	if sunDebil.Level != "DEBILITATED" {
		t.Errorf("expected Sun in Libra to be DEBILITATED, got %s", sunDebil.Level)
	}

	// 3. Jupiter exalted in Cancer (Sign 4)
	jupExalted := dasha.EvaluateGrahaDignity(dasha.PlanetJupiter, 4, 5.0)
	if jupExalted.Level != "EXALTED" {
		t.Errorf("expected Jupiter in Cancer to be EXALTED, got %s", jupExalted.Level)
	}

	// 4. Saturn exalted in Libra (Sign 7)
	satExalted := dasha.EvaluateGrahaDignity(dasha.PlanetSaturn, 7, 20.0)
	if satExalted.Level != "EXALTED" {
		t.Errorf("expected Saturn in Libra to be EXALTED, got %s", satExalted.Level)
	}

	// 5. Mars debilitated in Cancer (Sign 4)
	marsDebil := dasha.EvaluateGrahaDignity(dasha.PlanetMars, 4, 28.0)
	if marsDebil.Level != "DEBILITATED" {
		t.Errorf("expected Mars in Cancer to be DEBILITATED, got %s", marsDebil.Level)
	}
}

func TestEnrichTimelineWithAlchemicalData(t *testing.T) {
	birthTime := time.Date(1988, 10, 15, 6, 30, 0, 0, time.UTC)
	timeline := []dasha.DashaPeriod{
		{
			Planet:    dasha.PlanetMars,
			StartDate: birthTime,
			EndDate:   birthTime.AddDate(7, 0, 0),
			SubPeriods: []dasha.DashaPeriod{
				{
					Planet:    dasha.PlanetJupiter,
					StartDate: birthTime,
					EndDate:   birthTime.AddDate(1, 0, 0),
				},
			},
		},
	}

	enriched := dasha.EnrichTimelineWithAlchemicalData(timeline, birthTime)
	if len(enriched) != 1 {
		t.Fatalf("expected 1 enriched period, got %d", len(enriched))
	}

	p := enriched[0]
	if p.SacredMetal != "Iron (Ferrum)" {
		t.Errorf("expected Mars sacred metal to be Iron (Ferrum), got %s", p.SacredMetal)
	}
	if p.HermeticAxiom == "" {
		t.Errorf("expected non-empty Hermetic Axiom for Mars")
	}
	if p.StoryArchetype == "" {
		t.Errorf("expected non-empty Story Archetype for Mars")
	}
	if p.AgeStart != 0 {
		t.Errorf("expected AgeStart 0, got %f", p.AgeStart)
	}
	if p.AgeEnd < 6.9 || p.AgeEnd > 7.1 {
		t.Errorf("expected AgeEnd around 7, got %f", p.AgeEnd)
	}

	sub := p.SubPeriods[0]
	if sub.SacredMetal != "Tin (Stannum)" {
		t.Errorf("expected Jupiter sacred metal to be Tin (Stannum), got %s", sub.SacredMetal)
	}

	ageStr := dasha.FormatAgeRange(p.AgeStart, p.AgeEnd)
	if ageStr != "Age 0.0 – 7.0" {
		t.Errorf("expected 'Age 0.0 – 7.0', got '%s'", ageStr)
	}
}

