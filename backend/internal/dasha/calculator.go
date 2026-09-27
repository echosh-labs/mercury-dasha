package dasha

import (
	"errors"
	"fmt"
	"math"
	"time"
)

const (
	SolarYearDays = 365.24219878 // Mean tropical solar year in days
	NakshatraSpan = 360.0 / 27.0 // 13.333333333333334 degrees
	PadaSpan      = NakshatraSpan / 4.0 // 3.3333333333333335 degrees
)

// CalculateFromRequest processes CalculationRequest and computes the full DashaProfile.
func CalculateFromRequest(req CalculationRequest) (*DashaProfile, error) {
	if req.BirthDate == "" {
		return nil, errors.New("birth_date is required (format: YYYY-MM-DD)")
	}

	birthTimeStr := req.BirthTime
	if birthTimeStr == "" {
		birthTimeStr = "12:00"
	}

	dtStr := fmt.Sprintf("%s %s", req.BirthDate, birthTimeStr)
	localTime, err := time.Parse("2006-01-02 15:04", dtStr)
	if err != nil {
		return nil, fmt.Errorf("invalid birth_date or birth_time format: %w", err)
	}

	// Apply timezone offset to get UTC
	offsetDuration := time.Duration(req.TimezoneOffset * float64(time.Hour))
	birthUTC := localTime.Add(-offsetDuration).UTC()

	var moonDeg float64
	if req.MoonDegree != nil {
		moonDeg = *req.MoonDegree
	} else if req.NakshatraIndex > 0 {
		nak, ok := GetNakshatraByIndex(req.NakshatraIndex)
		if !ok {
			return nil, fmt.Errorf("invalid nakshatra_index: %d (must be 1-27)", req.NakshatraIndex)
		}
		pada := req.PadaNumber
		if pada < 1 || pada > 4 {
			pada = 1
		}
		moonDeg = nak.DegreeStart + (float64(pada)-0.5)*PadaSpan
	} else {
		// Ephemeris calculation
		moonDeg = CalculateSiderealMoon(birthUTC)
	}

	targetTime := time.Now().UTC()
	if req.TargetTime != nil && *req.TargetTime != "" {
		if t, err := time.Parse(time.RFC3339, *req.TargetTime); err == nil {
			targetTime = t.UTC()
		}
	}

	profileName := req.Name
	if profileName == "" {
		profileName = fmt.Sprintf("Chart-%s", req.BirthDate)
	}

	prof, err := BuildProfile(profileName, birthUTC, moonDeg, targetTime)
	if err != nil {
		return nil, err
	}
	prof.BirthDate = req.BirthDate
	prof.BirthTime = birthTimeStr
	prof.TimezoneOffset = req.TimezoneOffset
	prof.Latitude = req.Latitude
	prof.Longitude = req.Longitude
	if req.NakshatraIndex > 0 {
		prof.Mode = "nakshatra"
		prof.NakshatraIndex = req.NakshatraIndex
		prof.PadaNumber = req.PadaNumber
	} else {
		prof.Mode = "ephemeris"
	}
	now := time.Now().UTC()
	prof.CreatedAt = now
	prof.UpdatedAt = now

	return prof, nil
}

// BuildSovereignProfile constructs the complete unified SovereignProfile from calculation request.
func BuildSovereignProfile(req CalculationRequest, targetTime time.Time) (*SovereignProfile, error) {
	if req.TargetTime == nil || *req.TargetTime == "" {
		tStr := targetTime.Format(time.RFC3339)
		req.TargetTime = &tStr
	}
	return CalculateFromRequest(req)
}

// BuildProfile constructs the complete 120-year Dasha timeline and resolves active snapshot.
func BuildProfile(name string, birthUTC time.Time, moonDeg float64, targetTime time.Time) (*DashaProfile, error) {
	// Normalize degree to [0, 360)
	for moonDeg < 0 {
		moonDeg += 360
	}
	for moonDeg >= 360 {
		moonDeg -= 360
	}

	janmaNak, padaNum, traversed := GetNakshatraByDegree(moonDeg)
	balanceFraction := (NakshatraSpan - traversed) / NakshatraSpan
	if balanceFraction < 0 {
		balanceFraction = 0
	}
	if balanceFraction > 1 {
		balanceFraction = 1
	}

	startingLordID := janmaNak.RulingPlanet
	startingLord := PlanetsByID[startingLordID]
	balanceYears := startingLord.DurationYears * balanceFraction

	// Build full 120-year timeline starting at birth
	timeline := generate120YearTimeline(birthUTC, startingLordID, balanceFraction)

	// Resolve active period snapshot
	snapshot := ResolveActiveSnapshot(timeline, targetTime)

	profileID := fmt.Sprintf("dasha-%d-%s", birthUTC.Unix(), janmaNak.ID)

	prof := &DashaProfile{
		ID:             profileID,
		Name:           name,
		BirthTimeUTC:   birthUTC,
		MoonDegree:     math.Round(moonDeg*1000) / 1000,
		JanmaNakshatra: janmaNak,
		JanmaPada:      padaNum,
		BalanceYears:   math.Round(balanceYears*1000) / 1000,
		StartingLord:   startingLordID,
		ActiveSnapshot: snapshot,
		Timeline:       timeline,
	}

	// Populate unified astrological and alchemical matrices
	PopulateUnifiedMatrices(prof)

	return prof, nil
}

// PopulateUnifiedMatrices derives NatalAstrology, NatalAlchemy, ActiveAlchemicalState, and TimelineSummary.
func PopulateUnifiedMatrices(prof *DashaProfile) {
	if prof == nil {
		return
	}

	// If profile was saved with only raw birth parameters, automatically compute the ephemeris
	if prof.MoonDegree == 0 && prof.JanmaNakshatra.ID == "" && prof.BirthDate != "" {
		calcReq := CalculationRequest{
			Name:           prof.Name,
			BirthDate:      prof.BirthDate,
			BirthTime:      prof.BirthTime,
			TimezoneOffset: prof.TimezoneOffset,
			Latitude:       prof.Latitude,
			Longitude:      prof.Longitude,
			NakshatraIndex: prof.NakshatraIndex,
			PadaNumber:     prof.PadaNumber,
		}
		if fresh, err := CalculateFromRequest(calcReq); err == nil && fresh != nil {
			prof.BirthTimeUTC = fresh.BirthTimeUTC
			prof.MoonDegree = fresh.MoonDegree
			prof.JanmaNakshatra = fresh.JanmaNakshatra
			prof.JanmaPada = fresh.JanmaPada
			prof.StartingLord = fresh.StartingLord
			prof.BalanceYears = fresh.BalanceYears
			prof.ActiveSnapshot = fresh.ActiveSnapshot
			prof.Timeline = fresh.Timeline
		}
	}

	// Ensure BirthTimeUTC is populated if we have BirthDate and BirthTimeUTC is zero
	if prof.BirthTimeUTC.IsZero() && prof.BirthDate != "" {
		bTime := prof.BirthTime
		if bTime == "" {
			bTime = "12:00"
		}
		if lt, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%s %s", prof.BirthDate, bTime)); err == nil {
			offsetDuration := time.Duration(prof.TimezoneOffset * float64(time.Hour))
			prof.BirthTimeUTC = lt.Add(-offsetDuration).UTC()
		}
	}

	// Calculate Sidereal Sun and Panchanga if BirthTimeUTC is known
	var siderealSun float64
	var panchanga VedicPanchanga
	if !prof.BirthTimeUTC.IsZero() {
		siderealSun = CalculateSiderealSun(prof.BirthTimeUTC)
		panchanga = ResolvePanchanga(prof.BirthTimeUTC, prof.MoonDegree, siderealSun)
	}

	// Resolve Ayurvedic Constitution from Janma Nakshatra
	nakIdx := prof.JanmaNakshatra.Index
	if nakIdx == 0 && prof.NakshatraIndex > 0 {
		nakIdx = prof.NakshatraIndex
	}
	ayurveda := ResolveAyurvedicConstitution(nakIdx)

	// Derive 9 Classical Grahas, Tripod of Embodiment (Lagna, Surya, Chandra), and 12 Bhavas
	var tripod TripodOfEmbodiment
	var grahas []GrahaPlacement
	if !prof.BirthTimeUTC.IsZero() {
		lat := prof.Latitude
		lon := prof.Longitude
		if lat == 0 && lon == 0 {
			lat = 43.1594
			lon = -79.2469
		}
		lagnaDeg := CalculateSiderealLagna(prof.BirthTimeUTC, lat, lon)
		lagna := ResolveNatalLagna(lagnaDeg)
		grahas = ResolveAllGrahas(prof.BirthTimeUTC, prof.MoonDegree, siderealSun, lagna.RashiIndex)
		tripod = ResolveTripodAndBhavasWithGrahas(prof.BirthTimeUTC, lat, lon, prof.MoonDegree, siderealSun, grahas)
	}

	// 1. Natal Astrology
	var pada Pada
	if prof.JanmaPada >= 1 && prof.JanmaPada <= len(prof.JanmaNakshatra.Padas) {
		pada = prof.JanmaNakshatra.Padas[prof.JanmaPada-1]
	}
	startingLord := PlanetsByID[prof.StartingLord]

	prof.Astrology = NatalAstrology{
		SiderealMoonDegree: prof.MoonDegree,
		SiderealSunDegree:  siderealSun,
		JanmaNakshatra:     prof.JanmaNakshatra,
		JanmaPada:          prof.JanmaPada,
		PadaDetail:         pada,
		StartingLord:       prof.StartingLord,
		StartingLordName:   startingLord.Name,
		BalanceYears:       prof.BalanceYears,
		ElementalTattva:    startingLord.Element,
		Panchanga:          panchanga,
		Ayurveda:           ayurveda,
		Lagna:              tripod.Lagna,
		Tripod:             tripod,
		Grahas:             grahas,
	}

	// 2. Natal Alchemy
	metal, ok := MetalsByPlanet[prof.StartingLord]
	if !ok {
		metal = MetalsByPlanet[PlanetMercury]
	}
	axiom, ok := AxiomsByPlanet[prof.StartingLord]
	if !ok {
		axiom = AxiomsByPlanet[PlanetMercury]
	}
	var stage MagnumOpusStage
	for _, s := range MagnumOpusStages {
		if s.Planet == prof.StartingLord {
			stage = s
			break
		}
	}
	if stage.Name == "" {
		stage = MagnumOpusStages[2] // Separation
	}

	prof.Alchemy = NatalAlchemy{
		SacredMetal:         metal,
		GoverningAxiom:      axiom,
		MagnumOpusStage:     stage,
		ChakraAnchor:        startingLord.ChakraCenter,
		ResonantFrequencyHz: startingLord.RootFrequency,
		QuicksilverAffinity: metal.QuicksilverAffinity,
		AlchemicalMotto:     fmt.Sprintf("Solve et Coagula: Transmuting %s into sovereign conscious embodiment.", metal.Name),
	}

	// 3. Active Alchemical State
	mahaPlanet := prof.ActiveSnapshot.Mahadasha.Planet
	mahaMetal, ok := MetalsByPlanet[mahaPlanet]
	if !ok {
		mahaMetal = MetalsByPlanet[PlanetMercury]
	}
	mahaAxiom, ok := AxiomsByPlanet[mahaPlanet]
	if !ok {
		mahaAxiom = AxiomsByPlanet[PlanetMercury]
	}

	antarPlanet := prof.ActiveSnapshot.Antardasha.Planet
	antarMetal, ok := MetalsByPlanet[antarPlanet]
	if !ok {
		antarMetal = MetalsByPlanet[PlanetMercury]
	}
	antarAxiom, ok := AxiomsByPlanet[antarPlanet]
	if !ok {
		antarAxiom = AxiomsByPlanet[PlanetMercury]
	}

	prof.ActiveAlchemy = ActiveAlchemicalState{
		MahadashaMetal:      mahaMetal,
		MahadashaAxiom:      mahaAxiom,
		AntardashaMetal:     antarMetal,
		AntardashaAxiom:     antarAxiom,
		TransmutationVessel: fmt.Sprintf("Vessel of %s modulated by %s", mahaMetal.Name, antarMetal.Name),
	}

	// 4. Enrich Full 120-Year Timeline Hierarchy & Active Snapshot with Alchemical Metaphysics
	if len(prof.Timeline) > 0 {
		prof.Timeline = EnrichTimelineWithAlchemicalData(prof.Timeline, prof.BirthTimeUTC)

		summary := make([]DashaPeriodSummary, 0, len(prof.Timeline))
		for _, m := range prof.Timeline {
			summary = append(summary, DashaPeriodSummary{
				Level:        m.Level,
				Planet:       m.Planet,
				PlanetName:   m.PlanetName,
				SanskritName: m.SanskritName,
				ColorHex:     m.ColorHex,
				DurationDays: m.DurationDays,
				StartDate:    m.StartDate,
				EndDate:      m.EndDate,
				SacredMetal:  m.SacredMetal,
				MetalSymbol:  m.MetalSymbol,
				AgeRange:     FormatAgeRange(m.AgeStart, m.AgeEnd),
			})
		}
		prof.TimelineSummary = summary
	}

	// Enrich active snapshot periods with alchemical & age markers
	EnrichPeriodWithAlchemicalData(&prof.ActiveSnapshot.Mahadasha, prof.BirthTimeUTC)
	EnrichPeriodWithAlchemicalData(&prof.ActiveSnapshot.Antardasha, prof.BirthTimeUTC)
	EnrichPeriodWithAlchemicalData(&prof.ActiveSnapshot.Pratyantardasha, prof.BirthTimeUTC)

	// 5. Populate First-Class Top-Level Projections
	prof.NakshatraName = prof.JanmaNakshatra.Name
	prof.LagnaRashi = prof.Astrology.Lagna.Rashi
	prof.SunRashi = prof.Astrology.Tripod.Surya.Rashi
	prof.MoonRashi = prof.Astrology.Tripod.Chandra.Rashi
	prof.NatalMetal = prof.Alchemy.SacredMetal.Name
	prof.AlchemicalMotto = prof.Alchemy.AlchemicalMotto
	prof.Dosha = prof.Astrology.Ayurveda.Dosha
	prof.Gana = prof.Astrology.Ayurveda.Gana
	prof.Tithi = prof.Astrology.Panchanga.TithiName
	prof.YoniTotem = prof.Astrology.Ayurveda.YoniTotem
	prof.ElementalTattva = prof.Astrology.ElementalTattva
	prof.MagnumOpusStage = prof.Alchemy.MagnumOpusStage.Name
	prof.ChakraAnchor = prof.Alchemy.ChakraAnchor
	prof.ResonantFrequencyHz = prof.Alchemy.ResonantFrequencyHz
}

// ResolveSymbioticResonance evaluates topocentric harmonic resonance between a profile and the active transit hora planet.
func ResolveSymbioticResonance(prof *DashaProfile, transitPlanet PlanetID) SymbioticResonance {
	if transitPlanet == "" {
		transitPlanet = PlanetMercury
	}

	transitMetal, ok := MetalsByPlanet[transitPlanet]
	if !ok {
		transitMetal = MetalsByPlanet[PlanetMercury]
	}
	transitAxiom, ok := AxiomsByPlanet[transitPlanet]
	if !ok {
		transitAxiom = AxiomsByPlanet[PlanetMercury]
	}
	transitName := string(transitPlanet)
	if p, ok := PlanetsByID[transitPlanet]; ok {
		transitName = p.Name
	}

	isJanma := transitPlanet == prof.StartingLord
	isMaha := transitPlanet == prof.ActiveSnapshot.Mahadasha.Planet
	isAntar := transitPlanet == prof.ActiveSnapshot.Antardasha.Planet

	tier := "NEUTRAL_TRANSIT"
	guidance := fmt.Sprintf("Ambient cosmic alignment under %s (%s). Harmonize inner intention with the natural tide of time.", transitName, transitMetal.Name)

	if isJanma && isMaha {
		tier = "SOVEREIGN_JANMA_CONJUNCTION"
		guidance = fmt.Sprintf("Peak Sovereign Resonance: The live %s celestial hora directly aligns with both your birth lord and active Mahadasha era. Maximum catalytic empowerment.", transitName)
	} else if isJanma {
		tier = "SOVEREIGN_JANMA_CONJUNCTION"
		guidance = fmt.Sprintf("Natal Alignment: The live %s hora aligns with your birth Janma ruler (%s). Vitality, authentic will, and foundational essence activated.", transitName, prof.Astrology.StartingLordName)
	} else if isMaha {
		tier = "MAHADASHA_HARMONIC"
		guidance = fmt.Sprintf("Mahadasha Harmonic: The live %s hora resonates with your governing %s era. Strategic endeavors, overarching vision, and major works favored.", transitName, prof.ActiveSnapshot.Mahadasha.PlanetName)
	} else if isAntar {
		tier = "ANTARDASHA_HARMONIC"
		guidance = fmt.Sprintf("Antardasha Harmonic: The live %s hora activates your active %s sub-chapter. Tactical execution, focused problem solving, and breakthroughs favored.", transitName, prof.ActiveSnapshot.Antardasha.PlanetName)
	}

	return SymbioticResonance{
		TransitHoraPlanet:     transitPlanet,
		TransitHoraPlanetName: transitName,
		TransitMetal:          transitMetal,
		TransitAxiom:          transitAxiom,
		IsJanmaResonant:       isJanma,
		IsMahadashaResonant:   isMaha,
		IsAntardashaResonant:  isAntar,
		ResonanceTier:         tier,
		TransmutationGuidance: guidance,
	}
}

// findLordIndex returns the 0-8 index in VimshottariPlanets.
func findLordIndex(id PlanetID) int {
	for i, p := range VimshottariPlanets {
		if p.ID == id {
			return i
		}
	}
	return 0
}

// generate120YearTimeline builds the Mahadasha, Antardasha, and Pratyantardasha sequence.
func generate120YearTimeline(birthUTC time.Time, startLord PlanetID, startBalanceFraction float64) []DashaPeriod {
	startIdx := findLordIndex(startLord)
	timeline := make([]DashaPeriod, 0, 9)

	currentTime := birthUTC

	for i := 0; i < 9; i++ {
		planetIdx := (startIdx + i) % 9
		planet := VimshottariPlanets[planetIdx]

		var durYears float64
		if i == 0 {
			durYears = planet.DurationYears * startBalanceFraction
		} else {
			durYears = planet.DurationYears
		}

		durDays := durYears * SolarYearDays
		startTime := currentTime
		endTime := startTime.Add(time.Duration(durDays * 24 * float64(time.Hour)))
		currentTime = endTime

		// Generate Antardashas (sub-periods)
		antardashas := generateAntardashas(planet, startTime, endTime, i == 0, startBalanceFraction)

		timeline = append(timeline, DashaPeriod{
			Level:        LevelMahadasha,
			Planet:       planet.ID,
			PlanetName:   planet.Name,
			SanskritName: planet.SanskritName,
			ColorHex:     planet.ColorHex,
			DurationDays: durDays,
			StartDate:    startTime,
			EndDate:      endTime,
			SubPeriods:   antardashas,
		})
	}

	return timeline
}

// generateAntardashas builds the 9 Antardashas within a Mahadasha.
func generateAntardashas(mahaPlanet Planet, mahaStart, mahaEnd time.Time, isFirstMaha bool, firstMahaBalance float64) []DashaPeriod {
	mIdx := findLordIndex(mahaPlanet.ID)
	antardashas := make([]DashaPeriod, 0, 9)

	if !isFirstMaha {
		// Standard full Mahadasha: all 9 Antardashas
		current := mahaStart
		for j := 0; j < 9; j++ {
			aIdx := (mIdx + j) % 9
			aPlanet := VimshottariPlanets[aIdx]

			durYears := (mahaPlanet.DurationYears * aPlanet.DurationYears) / 120.0
			durDays := durYears * SolarYearDays
			start := current
			end := start.Add(time.Duration(durDays * 24 * float64(time.Hour)))
			current = end

			pratyantardashas := generatePratyantardashas(mahaPlanet, aPlanet, start, end)

			antardashas = append(antardashas, DashaPeriod{
				Level:        LevelAntardasha,
				Planet:       aPlanet.ID,
				PlanetName:   aPlanet.Name,
				SanskritName: aPlanet.SanskritName,
				ColorHex:     aPlanet.ColorHex,
				DurationDays: durDays,
				StartDate:    start,
				EndDate:      end,
				SubPeriods:   pratyantardashas,
			})
		}
	} else {
		// First Mahadasha with balance: compute nominal durations and keep the portion after birth
		fullMahaDurDays := mahaPlanet.DurationYears * SolarYearDays
		elapsedNominalDays := fullMahaDurDays * (1.0 - firstMahaBalance)
		nominalMahaStart := mahaStart.Add(-time.Duration(elapsedNominalDays * 24 * float64(time.Hour)))

		current := nominalMahaStart
		for j := 0; j < 9; j++ {
			aIdx := (mIdx + j) % 9
			aPlanet := VimshottariPlanets[aIdx]

			durYears := (mahaPlanet.DurationYears * aPlanet.DurationYears) / 120.0
			durDays := durYears * SolarYearDays
			start := current
			end := start.Add(time.Duration(durDays * 24 * float64(time.Hour)))
			current = end

			// If this Antardasha ended before birth, skip it
			if !end.After(mahaStart) {
				continue
			}

			actualStart := start
			if actualStart.Before(mahaStart) {
				actualStart = mahaStart
			}

			pratyantardashas := generatePratyantardashas(mahaPlanet, aPlanet, actualStart, end)

			actualDurDays := end.Sub(actualStart).Hours() / 24.0

			antardashas = append(antardashas, DashaPeriod{
				Level:        LevelAntardasha,
				Planet:       aPlanet.ID,
				PlanetName:   aPlanet.Name,
				SanskritName: aPlanet.SanskritName,
				ColorHex:     aPlanet.ColorHex,
				DurationDays: actualDurDays,
				StartDate:    actualStart,
				EndDate:      end,
				SubPeriods:   pratyantardashas,
			})
		}
	}

	return antardashas
}

// generatePratyantardashas builds the 9 Pratyantardashas within an Antardasha.
func generatePratyantardashas(mahaPlanet, antarPlanet Planet, antarStart, antarEnd time.Time) []DashaPeriod {
	aIdx := findLordIndex(antarPlanet.ID)
	pratyantars := make([]DashaPeriod, 0, 9)

	totalAntarDays := antarEnd.Sub(antarStart).Hours() / 24.0
	current := antarStart

	for k := 0; k < 9; k++ {
		pIdx := (aIdx + k) % 9
		pPlanet := VimshottariPlanets[pIdx]

		fraction := pPlanet.DurationYears / 120.0
		durDays := totalAntarDays * fraction
		start := current
		end := start.Add(time.Duration(durDays * 24 * float64(time.Hour)))
		current = end

		pratyantars = append(pratyantars, DashaPeriod{
			Level:        LevelPratyantardasha,
			Planet:       pPlanet.ID,
			PlanetName:   pPlanet.Name,
			SanskritName: pPlanet.SanskritName,
			ColorHex:     pPlanet.ColorHex,
			DurationDays: durDays,
			StartDate:    start,
			EndDate:      end,
		})
	}

	return pratyantars
}

// ResolveActiveSnapshot determines the current active Mahadasha, Antardasha, and Pratyantardasha.
func ResolveActiveSnapshot(timeline []DashaPeriod, refTime time.Time) ActivePeriodSnapshot {
	var activeMaha DashaPeriod
	var activeAntar DashaPeriod
	var activePratyantar DashaPeriod

	// Default fallback if outside timeline range
	if len(timeline) > 0 {
		activeMaha = timeline[0]
		if len(activeMaha.SubPeriods) > 0 {
			activeAntar = activeMaha.SubPeriods[0]
			if len(activeAntar.SubPeriods) > 0 {
				activePratyantar = activeAntar.SubPeriods[0]
			}
		}
	}

	for _, m := range timeline {
		if (refTime.Equal(m.StartDate) || refTime.After(m.StartDate)) && refTime.Before(m.EndDate) {
			activeMaha = m
			for _, a := range m.SubPeriods {
				if (refTime.Equal(a.StartDate) || refTime.After(a.StartDate)) && refTime.Before(a.EndDate) {
					activeAntar = a
					for _, p := range a.SubPeriods {
						if (refTime.Equal(p.StartDate) || refTime.After(p.StartDate)) && refTime.Before(p.EndDate) {
							activePratyantar = p
							break
						}
					}
					break
				}
			}
			break
		}
	}

	// Calculations for elapsed / remaining
	mahaTotal := activeMaha.EndDate.Sub(activeMaha.StartDate).Hours() / 24.0
	mahaElapsed := refTime.Sub(activeMaha.StartDate).Hours() / 24.0
	if mahaElapsed < 0 {
		mahaElapsed = 0
	}
	if mahaElapsed > mahaTotal {
		mahaElapsed = mahaTotal
	}
	mahaRemain := mahaTotal - mahaElapsed
	mahaPct := 0.0
	if mahaTotal > 0 {
		mahaPct = (mahaElapsed / mahaTotal) * 100.0
	}

	antarTotal := activeAntar.EndDate.Sub(activeAntar.StartDate).Hours() / 24.0
	antarElapsed := refTime.Sub(activeAntar.StartDate).Hours() / 24.0
	if antarElapsed < 0 {
		antarElapsed = 0
	}
	if antarElapsed > antarTotal {
		antarElapsed = antarTotal
	}
	antarRemain := antarTotal - antarElapsed
	antarPct := 0.0
	if antarTotal > 0 {
		antarPct = (antarElapsed / antarTotal) * 100.0
	}

	mPlanet := PlanetsByID[activeMaha.Planet]
	aPlanet := PlanetsByID[activeAntar.Planet]

	storyCtx := StoryContext{
		ActiveArchetype:    fmt.Sprintf("%s / %s", mPlanet.Name, aPlanet.Name),
		ThematicPhase:      fmt.Sprintf("%s Mahadasha, governed by %s Antardasha", mPlanet.SanskritName, aPlanet.SanskritName),
		AlchemicalElement:  fmt.Sprintf("%s & %s", mPlanet.Element, aPlanet.Element),
		ChakraFocus:        fmt.Sprintf("%s -> %s", mPlanet.ChakraCenter, aPlanet.ChakraCenter),
		ResonantFrequency:  mPlanet.RootFrequency,
		HermeticPrinciples: []string{mPlanet.HermeticAxiom, aPlanet.HermeticAxiom},
		NarrativePrompt: fmt.Sprintf(
			"The protagonist navigates the major chapter of %s (%s) under the immediate catalyst of %s (%s). Tension emerges between %s and %s.",
			mPlanet.Name, mPlanet.StoryArchetype, aPlanet.Name, aPlanet.StoryArchetype, mPlanet.Element, aPlanet.Element,
		),
	}

	return ActivePeriodSnapshot{
		ReferenceTime:      refTime,
		Mahadasha:          activeMaha,
		Antardasha:         activeAntar,
		Pratyantardasha:    activePratyantar,
		MahaElapsedDays:    math.Round(mahaElapsed*10) / 10,
		MahaRemainingDays:  math.Round(mahaRemain*10) / 10,
		MahaPercentDone:    math.Round(mahaPct*100) / 100,
		AntarElapsedDays:   math.Round(antarElapsed*10) / 10,
		AntarRemainingDays: math.Round(antarRemain*10) / 10,
		AntarPercentDone:   math.Round(antarPct*100) / 100,
		StoryContext:       storyCtx,
	}
}

// CalculateSiderealMoon computes the Moon's sidereal longitude (0° - 360°) using Lahiri Ayanamsha.
func CalculateSiderealMoon(t time.Time) float64 {
	// 1. Julian Day Calculation
	year := t.Year()
	month := int(t.Month())
	day := float64(t.Day()) + (float64(t.Hour())/24.0) + (float64(t.Minute())/1440.0) + (float64(t.Second())/86400.0)

	if month <= 2 {
		year -= 1
		month += 12
	}

	a := math.Floor(float64(year) / 100.0)
	b := 2.0 - a + math.Floor(a/4.0)

	jd := math.Floor(365.25*float64(year+4716)) + math.Floor(30.6001*float64(month+1)) + day + b - 1524.5

	// 2. Julian Centuries from J2000.0
	T := (jd - 2451545.0) / 36525.0

	// 3. Moon's Mean Elements (in degrees)
	// Moon's mean longitude L0
	L0 := 218.3164477 + 481267.88128*T - 0.0015786*T*T + (T*T*T)/538841.0
	// Mean elongation of the Moon D
	D := 297.8501921 + 445267.1142*T - 0.0018819*T*T + (T*T*T)/545868.0
	// Sun's mean anomaly M
	M := 357.5291092 + 35999.05029*T - 0.0001536*T*T + (T*T*T)/24490000.0
	// Moon's mean anomaly M'
	Mp := 134.9633964 + 477198.8675*T + 0.0087414*T*T + (T*T*T)/69699.0
	// Moon's argument of latitude F
	F := 93.2720950 + 483202.0175*T - 0.0036539*T*T - (T*T*T)/3526000.0

	degToRad := math.Pi / 180.0

	// Periodic lunar perturbations (Meeus standard terms)
	dL := 6.289 * math.Sin(Mp*degToRad)
	dL += 1.274 * math.Sin((2.0*D-Mp)*degToRad)
	dL += 0.658 * math.Sin(2.0*D*degToRad)
	dL += 0.214 * math.Sin(2.0*Mp*degToRad)
	dL -= 0.186 * math.Sin(M*degToRad)
	dL -= 0.114 * math.Sin(2.0*F*degToRad)
	dL += 0.059 * math.Sin((2.0*D-M)*degToRad)
	dL += 0.057 * math.Sin((2.0*D-Mp-M)*degToRad)

	tropicalMoon := L0 + dL

	// 4. Lahiri Ayanamsha (standard Indian astronomical reference)
	// Base Lahiri at J2000.0: ~23.85709 degrees, rate ~50.29 arcseconds per year
	ayanamsha := 23.85709 + (50.29/3600.0)*(T*100.0)

	// 5. Sidereal Moon
	siderealMoon := tropicalMoon - ayanamsha

	// Normalize into [0, 360)
	for siderealMoon < 0 {
		siderealMoon += 360.0
	}
	for siderealMoon >= 360.0 {
		siderealMoon -= 360.0
	}

	return siderealMoon
}
