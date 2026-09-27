package chrono

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/dasha"
)

// HoraGuidance defines daily living correspondences and activities for each planet.
type HoraGuidance struct {
	PlanetID             dasha.PlanetID `json:"planet_id"`
	SuitableActivities   []string       `json:"suitable_activities"`
	UnsuitableActivities []string       `json:"unsuitable_activities"`
	BriefApplication     string         `json:"brief_application"`
}

// PlanetaryHoraGuidances maps each classical planet to authentic daily applications.
var PlanetaryHoraGuidances = map[dasha.PlanetID]HoraGuidance{
	dasha.PlanetSun: {
		PlanetID: dasha.PlanetSun,
		SuitableActivities: []string{
			"Executive decisions and leadership proclamations",
			"Seeking favor from superiors, authorities, or patrons",
			"Public presentations, keynotes, and broadcasts",
			"Creative self-expression and establishing personal authority",
			"Vitality, outdoor sunlight exposure, and posture alignment",
		},
		UnsuitableActivities: []string{
			"Clandestine actions, secrecy, or remaining concealed",
			"Submissive negotiations or yielding sovereignty",
			"Entering into long-term servitude or subordinate contracts",
		},
		BriefApplication: "Prime for leadership, executive clarity, public presentation, and sovereign creative authority.",
	},
	dasha.PlanetMoon: {
		PlanetID: dasha.PlanetMoon,
		SuitableActivities: []string{
			"Intuitive reflection, journaling, and dream synthesis",
			"Culinary arts, hospitality, and nurturing relationships",
			"Public empathy, audience engagement, and community gathering",
			"Fluid transitions, short voyages, and water-related activities",
			"Emotional healing and subconscious integration",
		},
		UnsuitableActivities: []string{
			"Rigid, unyielding legal commitments or immutable contracts",
			"Aggressive confrontations or surgical strikes",
			"Heavy physical endurance testing under acute mental stress",
		},
		BriefApplication: "Favorable for intuition, empathetic communication, creative receptivity, and subconscious reflection.",
	},
	dasha.PlanetMars: {
		PlanetID: dasha.PlanetMars,
		SuitableActivities: []string{
			"Vigorous debugging, refactoring stubborn bugs, and eliminating technical debt",
			"Physical athletics, strength training, and martial arts",
			"Decisive execution, courage, and cutting away dead weight",
			"Engineering with metals, forge work, and high-energy machinery",
			"Confronting obstacles and executing demanding manual operations",
		},
		UnsuitableActivities: []string{
			"Delicate peace talks, fragile diplomatic treaties, or marriage proposals",
			"Calm contemplative meditation requiring still serenity",
			"Starting gentle creative endeavors requiring soft aesthetics",
		},
		BriefApplication: "Ideal for decisive action, breaking inertia, vigorous debugging, and overcoming obstacles.",
	},
	dasha.PlanetMercury: {
		PlanetID: dasha.PlanetMercury,
		SuitableActivities: []string{
			"Writing essays, books, documentation, and external dispatches",
			"Architectural coding, algorithmic design, and intellectual synthesis",
			"Commercial transactions, contract review, and business negotiations",
			"Publishing across multi-channel networks, social media, and newsletters",
			"Learning complex systems, scientific study, and mathematical inquiry",
		},
		UnsuitableActivities: []string{
			"Passive surrender or unstructured daydreaming without intent",
			"Impulsive commitments made without reading the fine print",
			"Emotional confrontations lacking logical balance",
		},
		BriefApplication: "Superb for writing, technical architecture, contract synthesis, communications, and outbound dispatches.",
	},
	dasha.PlanetJupiter: {
		PlanetID: dasha.PlanetJupiter,
		SuitableActivities: []string{
			"Grand strategic planning, vision casting, and long-range roadmaps",
			"Philosophical writing, higher spiritual study, and mentoring others",
			"Financial investments, wealth allocation, and charitable philanthropy",
			"Legal affairs, institutional agreements, and ethical governance",
			"Launching expansive initiatives with generous horizons",
		},
		UnsuitableActivities: []string{
			"Micro-managing minutiae or obsessing over trivial bureaucratic details",
			"Severe deprivation, extreme fasting, or stinginess",
			"Short-sighted transactional bargaining",
		},
		BriefApplication: "Auspicious for strategic planning, philosophical synthesis, wisdom publishing, and financial expansion.",
	},
	dasha.PlanetVenus: {
		PlanetID: dasha.PlanetVenus,
		SuitableActivities: []string{
			"Aesthetic UI/UX styling, graphic design, and artistic refinement",
			"Musical composition, audio engineering, and acoustic resonance",
			"Diplomacy, reconciliation, social networking, and romance",
			"Purchasing fine instruments, beautiful garments, and luxury items",
			"Creating harmonious environments and synthesizing opposing viewpoints",
		},
		UnsuitableActivities: []string{
			"Harsh disciplinary measures or austere punitive actions",
			"Aggressive confrontations or high-friction disputes",
			"Dry bureaucratic accounting without aesthetic consideration",
		},
		BriefApplication: "Exceptional for visual styling, artistic design, musical composition, diplomacy, and harmonious synthesis.",
	},
	dasha.PlanetSaturn: {
		PlanetID: dasha.PlanetSaturn,
		SuitableActivities: []string{
			"Foundational system architecture and enduring infrastructure design",
			"Disciplined deep focus, solitary research, and contemplation of mortality/time",
			"Organizing archives, auditing data stores, and eliminating bloat",
			"Honoring obligations, paying off debts, and long-term commitments",
			"Soil work, construction, mining, and grounding meditation",
		},
		UnsuitableActivities: []string{
			"Hasty execution, spontaneous celebration, or gambling",
			"Launching frivolous ephemeral trends requiring instant virality",
			"Demanding immediate rapid feedback from untested systems",
		},
		BriefApplication: "Mastery through discipline: foundational architecture, deep solitary focus, endurance, and structural order.",
	},
}

// SubHora represents one of the 7 non-linear sub-divisions of a Planetary Hour governed by Oblique Ascension.
type SubHora struct {
	Index             int            `json:"index"` // 1 to 7
	PlanetID          dasha.PlanetID `json:"planet_id"`
	PlanetName        string         `json:"planet_name"`
	SanskritName      string         `json:"sanskrit_name"`
	ColorHex          string         `json:"color_hex"`
	MetalSymbol       string         `json:"metal_symbol"`
	SacredMetal       string         `json:"sacred_metal"`
	StartTime         time.Time      `json:"start_time"`
	EndTime           time.Time      `json:"end_time"`
	DurationMinutes   float64        `json:"duration_minutes"`
	LagnaArcStart     float64        `json:"lagna_arc_start"`
	LagnaArcEnd       float64        `json:"lagna_arc_end"`
	LagnaArcTraversed float64        `json:"lagna_arc_traversed"`
	AsuCount          float64        `json:"asu_count"` // Vedic Asu duration (1 Asu = 4s = 1')
	Active            bool           `json:"active"`
}

// PlanetaryHour captures a dynamically scaled astronomical temporal hour.
type PlanetaryHour struct {
	Index                  int            `json:"index"`                  // 1 to 24
	Diurnal                bool           `json:"diurnal"`                // true for 1-12, false for 13-24
	Phase                  string         `json:"phase"`                  // "morning_diurnal", "afternoon_diurnal", "nocturnal"
	PlanetID               dasha.PlanetID `json:"planet_id"`
	PlanetName             string         `json:"planet_name"`
	SanskritName           string         `json:"sanskrit_name"`
	StartTime              time.Time      `json:"start_time"`
	EndTime                time.Time      `json:"end_time"`
	DurationMinutes        float64        `json:"duration_minutes"`
	SacredMetal            string         `json:"sacred_metal"`
	MetalSymbol            string         `json:"metal_symbol"`
	Element                string         `json:"element"`
	ChakraCenter           string         `json:"chakra_center"`
	ColorHex               string         `json:"color_hex"`
	HermeticAxiom          string         `json:"hermetic_axiom"`
	MagnumOpusStage        string         `json:"magnum_opus_stage"`
	SuitableActivities     []string       `json:"suitable_activities"`
	UnsuitableActivities   []string       `json:"unsuitable_activities"`
	BriefApplication       string         `json:"brief_application"`
	SubHoras               []SubHora      `json:"sub_horas,omitempty"`
	ActiveSubHora          *SubHora       `json:"active_sub_hora,omitempty"`
	CurrentLagnaDegree     float64        `json:"current_lagna_degree"`
	LocalApparentSolarTime string         `json:"local_apparent_solar_time"`
	EquationOfTimeMinutes  float64        `json:"equation_of_time_minutes"`
	Active                 bool           `json:"active"`
}

// DashaResonance captures symbiotic alignment between the active Hora and Vimshottari period.
type DashaResonance struct {
	ActiveMahadasha    dasha.PlanetID `json:"active_mahadasha"`
	ActiveAntardasha   dasha.PlanetID `json:"active_antardasha"`
	ResonanceType      string         `json:"resonance_type"`      // "sovereign", "catalytic", "immediate", "harmonic", "neutral"
	ResonanceTitle     string         `json:"resonance_title"`
	Description        string         `json:"description"`
	HarmonicMultiplier float64        `json:"harmonic_multiplier"` // e.g. 2.0 for Sovereign, 1.5 for Catalytic
}

// PlanetaryDaySchedule captures the full 24-hour cycle and active real-time status.
type PlanetaryDaySchedule struct {
	Date                          string          `json:"date"`
	DayLord                       string          `json:"day_lord"`
	DayLordPlanetID               dasha.PlanetID  `json:"day_lord_planet_id"`
	Sunrise                       time.Time       `json:"sunrise"`
	SolarNoon                     time.Time       `json:"solar_noon"`     // True Solar Transit (Hour 6/7 boundary)
	Sunset                        time.Time       `json:"sunset"`
	NextSunrise                   time.Time       `json:"next_sunrise"`
	SolarMidnight                 time.Time       `json:"solar_midnight"` // Anti-meridian passage (Hour 18/19 boundary)
	EquationOfTimeMinutes         float64         `json:"equation_of_time_minutes"`
	LocalApparentSolarTime        string          `json:"local_apparent_solar_time"`
	CurrentLagnaDegree            float64         `json:"current_lagna_degree"`
	DayHourDurationMinutes        float64         `json:"day_hour_duration_minutes"`
	NightHourDurationMinutes      float64         `json:"night_hour_duration_minutes"`
	MorningHourDurationMinutes    float64         `json:"morning_hour_duration_minutes"`   // L_morning
	AfternoonHourDurationMinutes  float64         `json:"afternoon_hour_duration_minutes"` // L_afternoon
	Hours                         []PlanetaryHour `json:"hours"`
	ActiveHour                    PlanetaryHour   `json:"active_hour"`
	ElapsedMinutes                float64         `json:"elapsed_minutes"`
	RemainingMinutes              float64         `json:"remaining_minutes"`
	ProgressPercent               float64         `json:"progress_percent"`
	NextHourPlanet                string          `json:"next_hour_planet"`
	NextHourStartTime             time.Time       `json:"next_hour_start_time"`
	DashaResonance                *DashaResonance `json:"dasha_resonance,omitempty"`
}

type dayScheduleTemplate struct {
	Date                         string
	DayLord                      string
	DayLordPlanetID              dasha.PlanetID
	Sunrise                      time.Time
	SolarNoon                    time.Time
	Sunset                       time.Time
	NextSunrise                  time.Time
	SolarMidnight                time.Time
	EquationOfTimeMinutes        float64
	DayHourDurationMinutes       float64
	NightHourDurationMinutes     float64
	MorningHourDurationMinutes   float64
	AfternoonHourDurationMinutes float64
	Hours                        []PlanetaryHour
}

var (
	schedTemplateCacheMu sync.RWMutex
	schedTemplateCache   = make(map[string]dayScheduleTemplate)
)

// CalculateDaySchedule generates the authoritative non-linear 24-hour schedule for reference time t.
// Strictly adheres to True Solar Noon meridian transit between Hour 6 and Hour 7,
// and derives non-linear sub-horas from Oblique Ascension Lagna arc rising rates.
// Implements thread-safe template caching for maximal real-time data efficiency.
func CalculateDaySchedule(t time.Time, coords SolarCoordinates, activeProfile *dasha.DashaProfile) PlanetaryDaySchedule {
	// 1. Compute astronomical solar times
	todaySolar := CalculateSolarTimes(t, coords)

	var scheduleDate time.Time
	var sunrise, noon, sunset, nextSunrise, solarMidnight time.Time
	var morningHourDur, afternoonHourDur, nightHourDur time.Duration
	var eot float64

	// The astronomical day strictly begins at Local Apparent Sunrise.
	// If t is before today's sunrise, t belongs to the nocturnal hours of yesterday!
	if t.Before(todaySolar.Sunrise) {
		prevDay := t.AddDate(0, 0, -1)
		prevSolar := CalculateSolarTimes(prevDay, coords)
		scheduleDate = prevDay
		sunrise = prevSolar.Sunrise
		noon = prevSolar.SolarNoon
		sunset = prevSolar.Sunset
		nextSunrise = todaySolar.Sunrise
		solarMidnight = prevSolar.SolarMidnight
		eot = prevSolar.EquationOfTimeMinutes

		morningHourDur = prevSolar.MorningHourDuration
		afternoonHourDur = prevSolar.AfternoonHourDuration
		nightHourDur = nextSunrise.Sub(sunset) / 12
	} else {
		scheduleDate = t
		sunrise = todaySolar.Sunrise
		noon = todaySolar.SolarNoon
		sunset = todaySolar.Sunset
		nextSunrise = todaySolar.NextSunrise
		solarMidnight = todaySolar.SolarMidnight
		eot = todaySolar.EquationOfTimeMinutes

		morningHourDur = todaySolar.MorningHourDuration
		afternoonHourDur = todaySolar.AfternoonHourDuration
		nightHourDur = todaySolar.NightHourDuration
	}

	dateStr := scheduleDate.Format("2006-01-02")
	cacheKey := fmt.Sprintf("%s_%.3f_%.3f", dateStr, coords.Latitude, coords.Longitude)

	schedTemplateCacheMu.RLock()
	tmpl, found := schedTemplateCache[cacheKey]
	schedTemplateCacheMu.RUnlock()

	if !found {
		// 2. Day Lord determined by weekday of the local sunrise timestamp
		sunriseWeekday := sunrise.Weekday()
		dayLordPlanetID := DayRulers[sunriseWeekday]

		startIdx := 0
		for i, p := range ChaldeanOrder {
			if p == dayLordPlanetID {
				startIdx = i
				break
			}
		}

		rawHours := make([]PlanetaryHour, 24)
		for h := 0; h < 24; h++ {
			isDiurnal := h < 12
			horaPlanetID := ChaldeanOrder[(startIdx+h)%7]
			planet := dasha.PlanetsByID[horaPlanetID]
			alchem := dasha.ResolveAlchemicalAlignment(horaPlanetID)
			guidance := PlanetaryHoraGuidances[horaPlanetID]

			var start, end time.Time
			var phase string
			var durMins float64

			if h < 6 {
				phase = "morning_diurnal"
				durMins = morningHourDur.Minutes()
				start = sunrise.Add(time.Duration(float64(h) * durMins * float64(time.Minute)))
				if h == 5 {
					end = noon
				} else {
					end = sunrise.Add(time.Duration(float64(h+1) * durMins * float64(time.Minute)))
				}
			} else if h < 12 {
				phase = "afternoon_diurnal"
				durMins = afternoonHourDur.Minutes()
				afternoonIdx := h - 6
				if afternoonIdx == 0 {
					start = noon
				} else {
					start = noon.Add(time.Duration(float64(afternoonIdx) * durMins * float64(time.Minute)))
				}
				if h == 11 {
					end = sunset
				} else {
					end = noon.Add(time.Duration(float64(afternoonIdx+1) * durMins * float64(time.Minute)))
				}
			} else {
				phase = "nocturnal"
				durMins = nightHourDur.Minutes()
				nightIdx := h - 12
				if nightIdx == 0 {
					start = sunset
				} else {
					start = sunset.Add(time.Duration(float64(nightIdx) * durMins * float64(time.Minute)))
				}
				if h == 23 {
					end = nextSunrise
				} else {
					end = sunset.Add(time.Duration(float64(nightIdx+1) * durMins * float64(time.Minute)))
				}
			}

			subPartitions := PartitionHoraByObliqueAscension(start, end, coords.Latitude, coords.Longitude)
			subHoras := make([]SubHora, 7)
			horaLordIdx := 0
			for idx, p := range ChaldeanOrder {
				if p == horaPlanetID {
					horaLordIdx = idx
					break
				}
			}

			for s := 0; s < 7; s++ {
				subPlanetID := ChaldeanOrder[(horaLordIdx+s)%7]
				subPlanet := dasha.PlanetsByID[subPlanetID]
				subAlchem := dasha.ResolveAlchemicalAlignment(subPlanetID)

				subHoras[s] = SubHora{
					Index:             s + 1,
					PlanetID:          subPlanetID,
					PlanetName:        subPlanet.Name,
					SanskritName:      subPlanet.SanskritName,
					ColorHex:          subPlanet.ColorHex,
					MetalSymbol:       subAlchem.ActiveMetal.Symbol,
					SacredMetal:       subAlchem.ActiveMetal.Name,
					StartTime:         subPartitions[s].StartTime,
					EndTime:           subPartitions[s].EndTime,
					DurationMinutes:   subPartitions[s].DurationMinutes,
					LagnaArcStart:     subPartitions[s].LagnaArcStart,
					LagnaArcEnd:       subPartitions[s].LagnaArcEnd,
					LagnaArcTraversed: subPartitions[s].LagnaArcTraversed,
					AsuCount:          subPartitions[s].AsuCount,
				}
			}

			rawHours[h] = PlanetaryHour{
				Index:                 h + 1,
				Diurnal:               isDiurnal,
				Phase:                 phase,
				PlanetID:              horaPlanetID,
				PlanetName:            planet.Name,
				SanskritName:          planet.SanskritName,
				StartTime:             start,
				EndTime:               end,
				DurationMinutes:       math.Round(durMins*10) / 10,
				SacredMetal:           alchem.ActiveMetal.Name,
				MetalSymbol:           alchem.ActiveMetal.Symbol,
				Element:               planet.Element,
				ChakraCenter:          planet.ChakraCenter,
				ColorHex:              planet.ColorHex,
				HermeticAxiom:         planet.HermeticAxiom,
				MagnumOpusStage:       alchem.MagnumOpusStage.Name,
				SuitableActivities:    guidance.SuitableActivities,
				UnsuitableActivities:  guidance.UnsuitableActivities,
				BriefApplication:      guidance.BriefApplication,
				SubHoras:              subHoras,
				EquationOfTimeMinutes: eot,
			}
		}

		dayHourDur := (noon.Sub(sunrise) + sunset.Sub(noon)) / 12
		tmpl = dayScheduleTemplate{
			Date:                         dateStr,
			DayLord:                      dasha.PlanetsByID[dayLordPlanetID].Name,
			DayLordPlanetID:              dayLordPlanetID,
			Sunrise:                      sunrise,
			SolarNoon:                    noon,
			Sunset:                       sunset,
			NextSunrise:                  nextSunrise,
			SolarMidnight:                solarMidnight,
			EquationOfTimeMinutes:        eot,
			DayHourDurationMinutes:       dayHourDur.Minutes(),
			NightHourDurationMinutes:     nightHourDur.Minutes(),
			MorningHourDurationMinutes:   morningHourDur.Minutes(),
			AfternoonHourDurationMinutes: afternoonHourDur.Minutes(),
			Hours:                        rawHours,
		}

		schedTemplateCacheMu.Lock()
		if len(schedTemplateCache) > 100 {
			schedTemplateCache = make(map[string]dayScheduleTemplate)
		}
		schedTemplateCache[cacheKey] = tmpl
		schedTemplateCacheMu.Unlock()
	}

	// 3. Compute real-time Local Apparent Solar Time and instantaneous Ascendant (Lagna)
	apparentSolarTime := ConvertToLocalApparentSolarTime(t, coords.Longitude)
	apparentSolarTimeStr := apparentSolarTime.Format("15:04:05")
	currentLagna := CalculateAscendant(t, coords.Latitude, coords.Longitude)
	roundedLagna := math.Round(currentLagna*100) / 100

	// 4. Instantiated hours with real-time dynamic active states
	hours := make([]PlanetaryHour, 24)
	var activeHour PlanetaryHour
	activeFound := false

	for h := 0; h < 24; h++ {
		hCopy := tmpl.Hours[h]
		hCopy.CurrentLagnaDegree = roundedLagna
		hCopy.LocalApparentSolarTime = apparentSolarTimeStr

		isActive := false
		if !activeFound {
			if (t.Equal(hCopy.StartTime) || t.After(hCopy.StartTime)) && t.Before(hCopy.EndTime) {
				isActive = true
				activeFound = true
			} else if h == 23 && t.After(hCopy.EndTime) {
				isActive = true
				activeFound = true
			}
		}
		hCopy.Active = isActive

		if isActive {
			subs := make([]SubHora, len(hCopy.SubHoras))
			copy(subs, hCopy.SubHoras)
			for s := range subs {
				subActive := false
				if (t.Equal(subs[s].StartTime) || t.After(subs[s].StartTime)) && t.Before(subs[s].EndTime) {
					subActive = true
				} else if s == len(subs)-1 && t.After(subs[s].EndTime) {
					subActive = true
				}
				subs[s].Active = subActive
				if subActive {
					hCopy.ActiveSubHora = &subs[s]
				}
			}
			hCopy.SubHoras = subs
			activeHour = hCopy
		}

		hours[h] = hCopy
	}

	// 6. Active Hour Fallback & Progress Metrics
	if !activeFound && len(hours) > 0 {
		activeHour = hours[0]
		hours[0].Active = true
	}

	elapsed := t.Sub(activeHour.StartTime).Minutes()
	if elapsed < 0 {
		elapsed = 0
	}
	remaining := activeHour.EndTime.Sub(t).Minutes()
	if remaining < 0 {
		remaining = 0
	}

	progressPct := 0.0
	if activeHour.DurationMinutes > 0 {
		progressPct = math.Min(100.0, math.Max(0.0, (elapsed/activeHour.DurationMinutes)*100.0))
	}

	nextIdx := activeHour.Index % 24
	nextHourPlanet := hours[nextIdx].PlanetName
	nextHourStart := activeHour.EndTime

	// 7. Resolve Symbiotic Dasha Resonance (if DashaProfile is supplied)
	var resonance *DashaResonance
	if activeProfile != nil {
		mahaLord := activeProfile.ActiveSnapshot.Mahadasha.Planet
		antarLord := activeProfile.ActiveSnapshot.Antardasha.Planet
		pratyantarLord := activeProfile.ActiveSnapshot.Pratyantardasha.Planet

		horaPlanet := activeHour.PlanetID

		if horaPlanet == mahaLord {
			resonance = &DashaResonance{
				ActiveMahadasha:    mahaLord,
				ActiveAntardasha:   antarLord,
				ResonanceType:      "sovereign",
				ResonanceTitle:     "Sovereign Resonance (Mahadasha Alignment)",
				Description:        fmt.Sprintf("Current %s Hora matches your governing Mahadasha lord (%s). Prime window for macrocosmic destiny alignment and life-defining initiatives.", activeHour.PlanetName, mahaLord),
				HarmonicMultiplier: 2.0,
			}
		} else if horaPlanet == antarLord {
			resonance = &DashaResonance{
				ActiveMahadasha:    mahaLord,
				ActiveAntardasha:   antarLord,
				ResonanceType:      "catalytic",
				ResonanceTitle:     "Catalytic Resonance (Antardasha Alignment)",
				Description:        fmt.Sprintf("Current %s Hora matches your active Antardasha sub-period (%s). High-velocity momentum for active creative projects and tangible progress.", activeHour.PlanetName, antarLord),
				HarmonicMultiplier: 1.6,
			}
		} else if horaPlanet == pratyantarLord {
			resonance = &DashaResonance{
				ActiveMahadasha:    mahaLord,
				ActiveAntardasha:   antarLord,
				ResonanceType:      "immediate",
				ResonanceTitle:     "Immediate Resonance (Pratyantardasha Alignment)",
				Description:        fmt.Sprintf("Current %s Hora matches your active Pratyantardasha lord (%s). Favorable for micro-tasks and immediate tactical breakthroughs.", activeHour.PlanetName, pratyantarLord),
				HarmonicMultiplier: 1.3,
			}
		} else {
			resonance = &DashaResonance{
				ActiveMahadasha:    mahaLord,
				ActiveAntardasha:   antarLord,
				ResonanceType:      "harmonic",
				ResonanceTitle:     "Harmonic Integration",
				Description:        fmt.Sprintf("%s Hora operates under the broader umbrella of your %s Mahadasha.", activeHour.PlanetName, mahaLord),
				HarmonicMultiplier: 1.0,
			}
		}
	}

	return PlanetaryDaySchedule{
		Date:                         scheduleDate.Format("2006-01-02"),
		DayLord:                      tmpl.DayLord,
		DayLordPlanetID:              tmpl.DayLordPlanetID,
		Sunrise:                      tmpl.Sunrise,
		SolarNoon:                    tmpl.SolarNoon,
		Sunset:                       tmpl.Sunset,
		NextSunrise:                  tmpl.NextSunrise,
		SolarMidnight:                tmpl.SolarMidnight,
		EquationOfTimeMinutes:        tmpl.EquationOfTimeMinutes,
		LocalApparentSolarTime:       apparentSolarTimeStr,
		CurrentLagnaDegree:           math.Round(currentLagna*100) / 100,
		DayHourDurationMinutes:       math.Round(tmpl.DayHourDurationMinutes*10) / 10,
		NightHourDurationMinutes:     math.Round(tmpl.NightHourDurationMinutes*10) / 10,
		MorningHourDurationMinutes:   math.Round(tmpl.MorningHourDurationMinutes*10) / 10,
		AfternoonHourDurationMinutes: math.Round(tmpl.AfternoonHourDurationMinutes*10) / 10,
		Hours:                        hours,
		ActiveHour:                   activeHour,
		ElapsedMinutes:               math.Round(elapsed*10) / 10,
		RemainingMinutes:             math.Round(remaining*10) / 10,
		ProgressPercent:              math.Round(progressPct*10) / 10,
		NextHourPlanet:               nextHourPlanet,
		NextHourStartTime:            nextHourStart,
		DashaResonance:               resonance,
	}
}
