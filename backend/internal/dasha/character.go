package dasha

import (
	"fmt"
	"time"
)

// CharacterChronicleRef is a lightweight metadata reference to an audiovisual timeline manifest.
type CharacterChronicleRef struct {
	ID          string    `json:"id"`           // e.g. "chronicle-sovereign-genesis-mars"
	Title       string    `json:"title"`        // e.g. "Mars-Jupiter Epoch Chronicle"
	DashaLevel  string    `json:"dasha_level"`  // "Mahadasha" | "Antardasha"
	Planet      PlanetID  `json:"planet"`
	DurationSec float64   `json:"duration_sec"`
	Orientation string    `json:"orientation"`  // "16:9", "9:16", "1:1"
	CreatedAt   time.Time `json:"created_at"`
}

// EnrichPeriodWithAlchemicalData populates the alchemical, vibrational, and age bounds for a DashaPeriod.
func EnrichPeriodWithAlchemicalData(p *DashaPeriod, birthTime time.Time) {
	if p == nil {
		return
	}

	// 1. Resolve planet data from PlanetsByID
	if planetData, exists := PlanetsByID[p.Planet]; exists {
		p.PlanetName = planetData.Name
		p.SanskritName = planetData.SanskritName
		p.ColorHex = planetData.ColorHex
		p.StoryArchetype = planetData.StoryArchetype
		p.FrequencyHz = planetData.RootFrequency
		p.HermeticAxiom = planetData.HermeticAxiom
	}

	// 2. Resolve Sacred Metal from MetalsByPlanet
	if metal, metalExists := MetalsByPlanet[p.Planet]; metalExists {
		p.SacredMetal = metal.Name
		p.MetalSymbol = metal.Symbol
	}

	// 3. Compute Age Bounds
	if !birthTime.IsZero() {
		if !p.StartDate.IsZero() {
			ageStart := p.StartDate.Sub(birthTime).Hours() / (24.0 * 365.25)
			if ageStart < 0 {
				ageStart = 0
			}
			p.AgeStart = ageStart
		}
		if !p.EndDate.IsZero() {
			ageEnd := p.EndDate.Sub(birthTime).Hours() / (24.0 * 365.25)
			if ageEnd < 0 {
				ageEnd = 0
			}
			p.AgeEnd = ageEnd
		}
	}

	// 4. Recursively enrich sub-periods (Antardashas and Pratyantardashas)
	for i := range p.SubPeriods {
		EnrichPeriodWithAlchemicalData(&p.SubPeriods[i], birthTime)
	}
}

// EnrichTimelineWithAlchemicalData enriches the entire 120-year Dasha hierarchy.
func EnrichTimelineWithAlchemicalData(timeline []DashaPeriod, birthTime time.Time) []DashaPeriod {
	for i := range timeline {
		EnrichPeriodWithAlchemicalData(&timeline[i], birthTime)
	}
	return timeline
}

// FormatAgeRange creates a concise human-readable string for character age during an epoch.
func FormatAgeRange(ageStart, ageEnd float64) string {
	return fmt.Sprintf("Age %.1f – %.1f", ageStart, ageEnd)
}
