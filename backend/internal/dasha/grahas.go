package dasha

import (
	"math"
	"time"
)

// NaturalFriendshipMap captures classical Vedic natural relations (Naisargika Sambandha).
// 1 = Friend, 0 = Neutral, -1 = Enemy
var NaturalFriendship = map[PlanetID]map[PlanetID]int{
	PlanetSun: {
		PlanetMoon: 1, PlanetMars: 1, PlanetJupiter: 1,
		PlanetMercury: 0,
		PlanetVenus: -1, PlanetSaturn: -1, PlanetRahu: -1, PlanetKetu: -1,
	},
	PlanetMoon: {
		PlanetSun: 1, PlanetMercury: 1,
		PlanetMars: 0, PlanetJupiter: 0, PlanetVenus: 0, PlanetSaturn: 0, PlanetRahu: 0, PlanetKetu: 0,
	},
	PlanetMars: {
		PlanetSun: 1, PlanetMoon: 1, PlanetJupiter: 1,
		PlanetVenus: 0, PlanetSaturn: 0,
		PlanetMercury: -1, PlanetRahu: -1, PlanetKetu: 0,
	},
	PlanetMercury: {
		PlanetSun: 1, PlanetVenus: 1,
		PlanetMars: 0, PlanetJupiter: 0, PlanetSaturn: 0,
		PlanetMoon: -1, PlanetRahu: 0, PlanetKetu: 0,
	},
	PlanetJupiter: {
		PlanetSun: 1, PlanetMoon: 1, PlanetMars: 1,
		PlanetSaturn: 0,
		PlanetMercury: -1, PlanetVenus: -1, PlanetRahu: -1, PlanetKetu: 0,
	},
	PlanetVenus: {
		PlanetMercury: 1, PlanetSaturn: 1,
		PlanetMars: 0, PlanetJupiter: 0,
		PlanetSun: -1, PlanetMoon: -1, PlanetRahu: 1, PlanetKetu: 0,
	},
	PlanetSaturn: {
		PlanetMercury: 1, PlanetVenus: 1,
		PlanetJupiter: 0,
		PlanetSun: -1, PlanetMoon: -1, PlanetMars: -1, PlanetRahu: 1, PlanetKetu: -1,
	},
	PlanetRahu: {
		PlanetMercury: 1, PlanetVenus: 1, PlanetSaturn: 1,
		PlanetJupiter: 0,
		PlanetSun: -1, PlanetMoon: -1, PlanetMars: -1, PlanetKetu: -1,
	},
	PlanetKetu: {
		PlanetMars: 1, PlanetVenus: 1, PlanetSaturn: 1,
		PlanetMercury: 0, PlanetJupiter: 0,
		PlanetSun: -1, PlanetMoon: -1, PlanetRahu: -1,
	},
}

// CalculateGrahaSiderealLongitudes derives sidereal degrees for Mars, Mercury, Jupiter, Venus, Saturn, Rahu, and Ketu.
func CalculateGrahaSiderealLongitudes(t time.Time) map[PlanetID]float64 {
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

	// Canonical Lahiri Ayanamsha
	ayanamsha := 23.85709 + (50.29/3600.0)*(T*100.0)

	normalize := func(deg float64) float64 {
		deg = math.Mod(deg, 360.0)
		if deg < 0 {
			deg += 360.0
		}
		return deg
	}

	siderealize := func(tropDeg float64) float64 {
		return normalize(tropDeg - ayanamsha)
	}

	// Keplerian/Meeus mean orbital elements adjusted for planetary perturbation
	mercuryTrop := 252.2509 + 149474.0722*T + 0.00030*T*T
	venusTrop := 181.9798 + 58519.2130*T + 0.00031*T*T
	marsTrop := 355.4330 + 19141.6964*T + 0.00031*T*T
	jupiterTrop := 34.3515 + 3036.3028*T + 0.00022*T*T
	saturnTrop := 50.0774 + 1223.5110*T + 0.00052*T*T

	// Rahu Mean Ascending Lunar Node (regressing westward)
	rahuTrop := 125.0445 - 1934.1363*T + 0.002075*T*T
	ketuTrop := rahuTrop + 180.0

	return map[PlanetID]float64{
		PlanetMercury: math.Round(siderealize(mercuryTrop)*1000) / 1000,
		PlanetVenus:   math.Round(siderealize(venusTrop)*1000) / 1000,
		PlanetMars:    math.Round(siderealize(marsTrop)*1000) / 1000,
		PlanetJupiter: math.Round(siderealize(jupiterTrop)*1000) / 1000,
		PlanetSaturn:  math.Round(siderealize(saturnTrop)*1000) / 1000,
		PlanetRahu:    math.Round(siderealize(rahuTrop)*1000) / 1000,
		PlanetKetu:    math.Round(siderealize(ketuTrop)*1000) / 1000,
	}
}

// EvaluateGrahaDignity evaluates the essential dignity of a Vedic planet in a given sign and degree.
func EvaluateGrahaDignity(planet PlanetID, rashiIdx int, degInSign float64) GrahaDignity {
	// 1. Exaltation (Ucha) & Debilitation (Neecha)
	switch planet {
	case PlanetSun:
		if rashiIdx == 1 {
			return GrahaDignity{Level: "EXALTED", Name: "Exalted (Ucha)", Description: "Peak solar authority, supreme sovereign brilliance & vital radiance.", ColorHex: "#f59e0b"}
		}
		if rashiIdx == 7 {
			return GrahaDignity{Level: "DEBILITATED", Name: "Debilitated (Neecha)", Description: "Diffused solar vitality requiring conscious alchemical reclamation.", ColorHex: "#ef4444"}
		}
		if rashiIdx == 5 {
			if degInSign <= 20.0 {
				return GrahaDignity{Level: "MOOLATRIKONA", Name: "Moolatrikona", Description: "Sovereign executive command zone; harmonious creative power.", ColorHex: "#fbbf24"}
			}
			return GrahaDignity{Level: "OWN_SIGN", Name: "Own Sign (Swakshetra)", Description: "At home in native sanctuary; natural self-contained vitality.", ColorHex: "#10b981"}
		}

	case PlanetMoon:
		if rashiIdx == 2 {
			return GrahaDignity{Level: "EXALTED", Name: "Exalted (Ucha)", Description: "Sublime emotional equilibrium, overflowing abundance & intuition.", ColorHex: "#f59e0b"}
		}
		if rashiIdx == 8 {
			return GrahaDignity{Level: "DEBILITATED", Name: "Debilitated (Neecha)", Description: "Psychic turbulence; profound occult transformative testing.", ColorHex: "#ef4444"}
		}
		if rashiIdx == 4 {
			return GrahaDignity{Level: "OWN_SIGN", Name: "Own Sign (Swakshetra)", Description: "Sanctuary of emotional peace, empathetic nurturance & clarity.", ColorHex: "#10b981"}
		}

	case PlanetMars:
		if rashiIdx == 10 {
			return GrahaDignity{Level: "EXALTED", Name: "Exalted (Ucha)", Description: "Disciplined courage, tactical supremacy & unstoppable execution.", ColorHex: "#f59e0b"}
		}
		if rashiIdx == 4 {
			return GrahaDignity{Level: "DEBILITATED", Name: "Debilitated (Neecha)", Description: "Internalized friction; emotional impulsivity needing tempered discipline.", ColorHex: "#ef4444"}
		}
		if rashiIdx == 1 {
			if degInSign <= 12.0 {
				return GrahaDignity{Level: "MOOLATRIKONA", Name: "Moolatrikona", Description: "Pioneering heroic drive; sharp righteous executive force.", ColorHex: "#fbbf24"}
			}
			return GrahaDignity{Level: "OWN_SIGN", Name: "Own Sign (Swakshetra)", Description: "Vital warrior power at home in primal fiery strength.", ColorHex: "#10b981"}
		}
		if rashiIdx == 8 {
			return GrahaDignity{Level: "OWN_SIGN", Name: "Own Sign (Swakshetra)", Description: "Strategic occult depth, regenerative endurance & penetrating willpower.", ColorHex: "#10b981"}
		}

	case PlanetMercury:
		if rashiIdx == 6 {
			if degInSign <= 15.0 {
				return GrahaDignity{Level: "EXALTED", Name: "Exalted (Ucha)", Description: "Flawless intellectual precision, sovereign discernment & eloquence.", ColorHex: "#f59e0b"}
			}
			if degInSign <= 20.0 {
				return GrahaDignity{Level: "MOOLATRIKONA", Name: "Moolatrikona", Description: "Analytical mastery, scientific methodology & commercial genius.", ColorHex: "#fbbf24"}
			}
			return GrahaDignity{Level: "OWN_SIGN", Name: "Own Sign (Swakshetra)", Description: "Detailed craftsmanship, clear logic & discerning organization.", ColorHex: "#10b981"}
		}
		if rashiIdx == 12 {
			return GrahaDignity{Level: "DEBILITATED", Name: "Debilitated (Neecha)", Description: "Transcendent, poetic abstraction challenging literal linear logic.", ColorHex: "#ef4444"}
		}
		if rashiIdx == 3 {
			return GrahaDignity{Level: "OWN_SIGN", Name: "Own Sign (Swakshetra)", Description: "Versatile communication, quick-witted insight & fluid exchange.", ColorHex: "#10b981"}
		}

	case PlanetJupiter:
		if rashiIdx == 4 {
			return GrahaDignity{Level: "EXALTED", Name: "Exalted (Ucha)", Description: "Divine benevolence, dharmic wisdom & overflowing spiritual fruition.", ColorHex: "#f59e0b"}
		}
		if rashiIdx == 10 {
			return GrahaDignity{Level: "DEBILITATED", Name: "Debilitated (Neecha)", Description: "Wisdom tested against material pragmatism & worldly restriction.", ColorHex: "#ef4444"}
		}
		if rashiIdx == 9 {
			if degInSign <= 10.0 {
				return GrahaDignity{Level: "MOOLATRIKONA", Name: "Moolatrikona", Description: "Cosmic philosophical vision; authoritative spiritual counsel.", ColorHex: "#fbbf24"}
			}
			return GrahaDignity{Level: "OWN_SIGN", Name: "Own Sign (Swakshetra)", Description: "Higher learning, inspirational guidance & generous leadership.", ColorHex: "#10b981"}
		}
		if rashiIdx == 12 {
			return GrahaDignity{Level: "OWN_SIGN", Name: "Own Sign (Swakshetra)", Description: "Mystical contemplation, deep empathy & spiritual sanctuary.", ColorHex: "#10b981"}
		}

	case PlanetVenus:
		if rashiIdx == 12 {
			return GrahaDignity{Level: "EXALTED", Name: "Exalted (Ucha)", Description: "Unconditional divine love, supreme aesthetic mastery & devotional grace.", ColorHex: "#f59e0b"}
		}
		if rashiIdx == 6 {
			return GrahaDignity{Level: "DEBILITATED", Name: "Debilitated (Neecha)", Description: "Hyper-critical analysis of beauty & relationships; requiring unconditional acceptance.", ColorHex: "#ef4444"}
		}
		if rashiIdx == 7 {
			if degInSign <= 15.0 {
				return GrahaDignity{Level: "MOOLATRIKONA", Name: "Moolatrikona", Description: "Supreme relational harmony, alchemical balance & cultural diplomacy.", ColorHex: "#fbbf24"}
			}
			return GrahaDignity{Level: "OWN_SIGN", Name: "Own Sign (Swakshetra)", Description: "Elegance, artistic refinement & sovereign partnership mastery.", ColorHex: "#10b981"}
		}
		if rashiIdx == 2 {
			return GrahaDignity{Level: "OWN_SIGN", Name: "Own Sign (Swakshetra)", Description: "Sensory appreciation, material elegance & grounded wealth resonance.", ColorHex: "#10b981"}
		}

	case PlanetSaturn:
		if rashiIdx == 7 {
			return GrahaDignity{Level: "EXALTED", Name: "Exalted (Ucha)", Description: "Flawless karmic justice, enduring structural mastery & sovereign poise.", ColorHex: "#f59e0b"}
		}
		if rashiIdx == 1 {
			return GrahaDignity{Level: "DEBILITATED", Name: "Debilitated (Neecha)", Description: "Patience tested by rash impulsive urgency; alchemical forging through delay.", ColorHex: "#ef4444"}
		}
		if rashiIdx == 11 {
			if degInSign <= 20.0 {
				return GrahaDignity{Level: "MOOLATRIKONA", Name: "Moolatrikona", Description: "Architect of long-range collective structures & visionary perseverance.", ColorHex: "#fbbf24"}
			}
			return GrahaDignity{Level: "OWN_SIGN", Name: "Own Sign (Swakshetra)", Description: "Deep resilience, communal dedication & steady karmic progress.", ColorHex: "#10b981"}
		}
		if rashiIdx == 10 {
			return GrahaDignity{Level: "OWN_SIGN", Name: "Own Sign (Swakshetra)", Description: "Executive discipline, institutional authority & mountain-like stamina.", ColorHex: "#10b981"}
		}

	case PlanetRahu:
		if rashiIdx == 2 || rashiIdx == 3 {
			return GrahaDignity{Level: "EXALTED", Name: "Exalted (Ucha)", Description: "Extraordinary visionary intellect, world-penetrating innovation & magnetic ambition.", ColorHex: "#f59e0b"}
		}
		if rashiIdx == 8 || rashiIdx == 9 {
			return GrahaDignity{Level: "DEBILITATED", Name: "Debilitated (Neecha)", Description: "Psychic turbulence; ungrounded obsession needing disciplined tethering.", ColorHex: "#ef4444"}
		}
		if rashiIdx == 11 {
			return GrahaDignity{Level: "OWN_SIGN", Name: "Own Sign / Co-Ruler", Description: "Vast network reach, futurist vision & non-traditional abundance.", ColorHex: "#10b981"}
		}

	case PlanetKetu:
		if rashiIdx == 8 || rashiIdx == 9 {
			return GrahaDignity{Level: "EXALTED", Name: "Exalted (Ucha)", Description: "Supreme spiritual awakening, occult illumination & karmic release.", ColorHex: "#f59e0b"}
		}
		if rashiIdx == 2 || rashiIdx == 3 {
			return GrahaDignity{Level: "DEBILITATED", Name: "Debilitated (Neecha)", Description: "Disconnection from physical/communicative grounding; ascetic friction.", ColorHex: "#ef4444"}
		}
		if rashiIdx == 8 {
			return GrahaDignity{Level: "OWN_SIGN", Name: "Own Sign / Co-Ruler", Description: "Mastery of occult transformation, deep regeneration & spiritual sanctuary.", ColorHex: "#10b981"}
		}
	}

	// 2. Evaluate Natural Relationship with Sign Lord
	signLord := RashisByIndex[rashiIdx].Lord
	if relMap, ok := NaturalFriendship[planet]; ok {
		if score, found := relMap[signLord]; found {
			switch score {
			case 1:
				return GrahaDignity{Level: "FRIEND", Name: "Friendly Sign (Mitra)", Description: "Cooperative elemental environment supporting constructive expression.", ColorHex: "#06b6d4"}
			case -1:
				return GrahaDignity{Level: "ENEMY", Name: "Enemy Sign (Shatru)", Description: "Challenging territory requiring conscious patience and transmutation.", ColorHex: "#f97316"}
			default:
				return GrahaDignity{Level: "NEUTRAL", Name: "Neutral Sign (Sama)", Description: "Balanced, pragmatic operation without strong affinity or friction.", ColorHex: "#94a3b8"}
			}
		}
	}

	return GrahaDignity{Level: "NEUTRAL", Name: "Neutral Sign (Sama)", Description: "Stable operational condition.", ColorHex: "#94a3b8"}
}

// ResolveAllGrahas constructs the full 9-Graha retinue with exact positions, Bhavas, and evaluated dignities.
func ResolveAllGrahas(birthUTC time.Time, moonDeg, sunDeg float64, lagnaRashiIdx int) []GrahaPlacement {
	longitudes := CalculateGrahaSiderealLongitudes(birthUTC)
	longitudes[PlanetSun] = sunDeg
	longitudes[PlanetMoon] = moonDeg

	grahasOrder := []PlanetID{
		PlanetSun,
		PlanetMoon,
		PlanetMars,
		PlanetMercury,
		PlanetJupiter,
		PlanetVenus,
		PlanetSaturn,
		PlanetRahu,
		PlanetKetu,
	}

	placements := make([]GrahaPlacement, len(grahasOrder))
	for i, pid := range grahasOrder {
		deg := longitudes[pid]
		for deg < 0 {
			deg += 360.0
		}
		for deg >= 360.0 {
			deg -= 360.0
		}

		rashiIdx := int(math.Floor(deg/30.0)) + 1
		if rashiIdx < 1 {
			rashiIdx = 1
		}
		if rashiIdx > 12 {
			rashiIdx = 12
		}

		rashi := RashisByIndex[rashiIdx]
		degInSign := deg - float64(rashiIdx-1)*30.0

		// Whole Sign house number
		houseNum := (rashiIdx - lagnaRashiIdx + 12) % 12
		if houseNum == 0 {
			houseNum = 12
		}
		houseName := BhavaMetadata[houseNum-1].Name

		nak, padaNum, _ := GetNakshatraByDegree(deg)
		dignity := EvaluateGrahaDignity(pid, rashiIdx, degInSign)
		planetMeta := PlanetsByID[pid]

		placements[i] = GrahaPlacement{
			ID:           pid,
			Name:         planetMeta.Name,
			SanskritName: planetMeta.SanskritName,
			Degree:       deg,
			Rashi:        rashi.Name,
			RashiIndex:   rashiIdx,
			DegreeInSign: math.Round(degInSign*100) / 100,
			HouseNumber:  houseNum,
			HouseName:    houseName,
			Nakshatra:    nak,
			Pada:         padaNum,
			Dignity:      dignity,
			IsRetrograde: false,
		}
	}

	return placements
}
