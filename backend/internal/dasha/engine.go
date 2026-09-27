package dasha

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/db"
)

// Engine serves as the authoritative orchestrator for the Mercury Dasha domain.
// It unifies celestial cycles, 27 Nakshatras, Hermetic alchemy, and profile persistence.
type Engine struct {
	store db.StorageEngine
}

// NewEngine constructs a new MercuryEngine instance.
func NewEngine(store db.StorageEngine) *Engine {
	return &Engine{
		store: store,
	}
}

// GetOverview returns dynamic Mahadasha cycles, frequencies, and planetary archetypes.
func (e *Engine) GetOverview() map[string]any {
	planetsResp := make([]map[string]any, 0, len(VimshottariPlanets))
	for _, p := range VimshottariPlanets {
		planetsResp = append(planetsResp, map[string]any{
			"id":              p.ID,
			"name":            p.Name,
			"sanskrit_name":   p.SanskritName,
			"duration_years":  p.DurationYears,
			"frequency_hz":    p.RootFrequency,
			"element":         p.Element,
			"chakra":          p.ChakraCenter,
			"color_hex":       p.ColorHex,
			"hermetic_axiom":  p.HermeticAxiom,
			"story_archetype": p.StoryArchetype,
		})
	}

	return map[string]any{
		"engine":            "Mercury Dasha Sovereign Core v2.0",
		"mahadasha_lord":    "Mercury (17-Year Cycle)",
		"total_years":       120,
		"root_frequency_hz": 141.27,
		"planets":           planetsResp,
	}
}

// GetNakshatras returns the catalog of all 27 nakshatras (or filtered by ruling planet),
// hydrating from the BoltDB sovereign storehouse when available.
func (e *Engine) GetNakshatras(planetFilter string) map[string]any {
	filterLower := strings.ToLower(strings.TrimSpace(planetFilter))

	// If no filter is applied and storehouse has catalog, return raw storehouse payload
	if filterLower == "" && e.store != nil {
		if raw, err := e.store.GetNakshatras(); err == nil && len(raw) > 0 {
			var catalog map[string]any
			if err := json.Unmarshal(raw, &catalog); err == nil {
				return catalog
			}
		}
	}

	type NakItem struct {
		ID           string   `json:"id"`
		Name         string   `json:"name"`
		SanskritName string   `json:"sanskrit_name"`
		Range        string   `json:"range"`
		ZodiacSpan   string   `json:"zodiac_span"`
		Frequency    float64  `json:"frequency"`
		FrequencyHz  float64  `json:"frequency_hz"`
		RulingPlanet string   `json:"ruling_planet"`
		Deity        string   `json:"deity"`
		Symbol       string   `json:"symbol"`
		Quality      string   `json:"quality"`
		Padas        []Pada   `json:"padas"`
	}

	nakshatrasList := make([]NakItem, 0, len(Nakshatras))
	mercuryList := make([]NakItem, 0, 3)

	for _, n := range Nakshatras {
		item := NakItem{
			ID:           n.ID,
			Name:         n.Name,
			SanskritName: n.SanskritName,
			Range:        n.ZodiacSpan,
			ZodiacSpan:   n.ZodiacSpan,
			Frequency:    n.FrequencyHz,
			FrequencyHz:  n.FrequencyHz,
			RulingPlanet: string(n.RulingPlanet),
			Deity:        n.Deity,
			Symbol:       n.Symbol,
			Quality:      n.Quality,
			Padas:        n.Padas,
		}

		if n.RulingPlanet == PlanetMercury {
			mercuryList = append(mercuryList, item)
		}

		if filterLower != "" && string(n.RulingPlanet) != filterLower {
			continue
		}
		nakshatrasList = append(nakshatrasList, item)
	}

	return map[string]any{
		"ruling_planet": filterLower,
		"count":         len(nakshatrasList),
		"total":         len(Nakshatras),
		"nakshatras":    nakshatrasList,
		"mercury_ruled": mercuryList,
	}
}

// GetPlanets returns the 9 planetary lords with their complete durations and attributes.
func (e *Engine) GetPlanets() map[string]any {
	return map[string]any{
		"total_years": 120,
		"planets":     VimshottariPlanets,
	}
}

// GetAlchemy returns the 7 Sacred Metals, 7 Hermetic Axioms, and live alchemical alignment,
// hydrating base catalogs from the BoltDB sovereign storehouse.
func (e *Engine) GetAlchemy(activePlanet PlanetID) map[string]any {
	if activePlanet == "" {
		activePlanet = PlanetMercury
	}

	alignment := ResolveAlchemicalAlignment(activePlanet)

	philosophy := "Hermetic & Alchemical Principles (Kybalion & Magnum Opus)"
	sovereignElem := "Quicksilver (Liquid Mercury / Hydrargyrum)"
	metals := SacredMetals
	axioms := HermeticAxioms
	stages := MagnumOpusStages

	// Hydrate from BoltDB storehouse if present
	if e.store != nil {
		if raw, err := e.store.GetAlchemy(); err == nil && len(raw) > 0 {
			var cat AlchemyCatalog
			if err := json.Unmarshal(raw, &cat); err == nil {
				if cat.Philosophy != "" {
					philosophy = cat.Philosophy
				}
				if cat.SovereignElement != "" {
					sovereignElem = cat.SovereignElement
				}
				if len(cat.SacredMetals) > 0 {
					metals = cat.SacredMetals
				}
				if len(cat.HermeticAxioms) > 0 {
					axioms = cat.HermeticAxioms
				}
				if len(cat.MagnumOpusStages) > 0 {
					stages = cat.MagnumOpusStages
				}
			}
		}
	}

	return map[string]any{
		"philosophy":         philosophy,
		"sovereign_element":  sovereignElem,
		"live_alignment":     alignment,
		"sacred_metals":      metals,
		"hermetic_axioms":    axioms,
		"magnum_opus_stages": stages,
	}
}

// Calculate parses and calculates the complete 120-year Vimshottari progression.
func (e *Engine) Calculate(req CalculationRequest) (*DashaProfile, error) {
	return CalculateFromRequest(req)
}

// GetSovereignProfile retrieves and automatically enriches a sovereign profile from the BoltDB store.
func (e *Engine) GetSovereignProfile(id string) (*SovereignProfile, error) {
	if e.store == nil {
		return nil, db.ErrNotFound
	}
	data, err := e.store.GetProfile(id)
	if err != nil {
		return nil, err
	}

	var prof SovereignProfile
	if err := json.Unmarshal(data, &prof); err != nil {
		return nil, err
	}

	// Guarantee unified matrices are fully populated
	if prof.Astrology.StartingLord == "" || prof.Alchemy.SacredMetal.ID == "" || prof.Astrology.Panchanga.TithiName == "" || prof.Astrology.Lagna.Rashi == "" {
		PopulateUnifiedMatrices(&prof)
	}

	return &prof, nil
}

// SaveSovereignProfile validates, enriches, and persists a sovereign profile to the BoltDB store.
func (e *Engine) SaveSovereignProfile(prof *SovereignProfile) error {
	if e.store == nil || prof == nil {
		return nil
	}
	PopulateUnifiedMatrices(prof)
	prof.UpdatedAt = time.Now().UTC()
	if prof.CreatedAt.IsZero() {
		prof.CreatedAt = prof.UpdatedAt
	}

	data, err := json.Marshal(prof)
	if err != nil {
		return err
	}
	return e.store.SaveProfile(prof.ID, data)
}

// ListSovereignProfileSummaries returns lightweight summary projections for all saved profiles.
func (e *Engine) ListSovereignProfileSummaries() ([]ProfileSummary, error) {
	if e.store == nil {
		return []ProfileSummary{}, nil
	}
	keys, err := e.store.ListProfiles()
	if err != nil {
		return nil, err
	}

	summaries := make([]ProfileSummary, 0, len(keys))
	for _, k := range keys {
		raw, err := e.store.GetProfile(k)
		if err != nil {
			continue
		}
		var prof SovereignProfile
		if err := json.Unmarshal(raw, &prof); err != nil {
			continue
		}
		if prof.Astrology.StartingLord == "" || prof.Astrology.Panchanga.TithiName == "" || prof.Astrology.Lagna.Rashi == "" {
			PopulateUnifiedMatrices(&prof)
		}

		natalMetal := prof.Alchemy.SacredMetal.Name
		if natalMetal == "" {
			natalMetal = string(prof.StartingLord)
		}
		activeMaha := prof.ActiveSnapshot.Mahadasha.PlanetName
		if activeMaha == "" {
			activeMaha = string(prof.ActiveSnapshot.Mahadasha.Planet)
		}

		summaries = append(summaries, ProfileSummary{
			ID:             prof.ID,
			Name:           prof.Name,
			BirthDate:      prof.BirthDate,
			LocationName:   prof.LocationName,
			NakshatraName:  prof.JanmaNakshatra.Name,
			NakshatraIndex: prof.JanmaNakshatra.Index,
			PadaNumber:     prof.JanmaPada,
			StartingLord:   string(prof.StartingLord),
			NatalMetal:     natalMetal,
			ActiveMaha:     activeMaha,
			Dosha:               prof.Astrology.Ayurveda.Dosha,
			Gana:                prof.Astrology.Ayurveda.Gana,
			Tithi:               prof.Astrology.Panchanga.TithiName,
			LagnaRashi:          prof.Astrology.Lagna.Rashi,
			SunRashi:            prof.Astrology.Tripod.Surya.Rashi,
			MoonRashi:           prof.Astrology.Tripod.Chandra.Rashi,
			YoniTotem:           prof.Astrology.Ayurveda.YoniTotem,
			ElementalTattva:     prof.Astrology.ElementalTattva,
			AlchemicalMotto:     prof.Alchemy.AlchemicalMotto,
			MagnumOpusStage:     prof.Alchemy.MagnumOpusStage.Name,
			ChakraAnchor:        prof.Alchemy.ChakraAnchor,
			ResonantFrequencyHz: prof.Alchemy.ResonantFrequencyHz,
			UpdatedAt:           prof.UpdatedAt,
		})
	}
	return summaries, nil
}

// GetProfileResonance computes real-time topocentric resonance between a saved profile and the active transit hora planet.
func (e *Engine) GetProfileResonance(id string, transitPlanet PlanetID) (*SymbioticResonance, error) {
	prof, err := e.GetSovereignProfile(id)
	if err != nil {
		return nil, err
	}
	res := ResolveSymbioticResonance(prof, transitPlanet)
	return &res, nil
}

// GetDetailedTimeline returns the full 3-tier 120-year timeline for a saved profile,
// generating it deterministically if omitted from the compact stored record.
func (e *Engine) GetDetailedTimeline(id string) ([]DashaPeriod, error) {
	prof, err := e.GetSovereignProfile(id)
	if err != nil {
		return nil, err
	}
	if len(prof.Timeline) > 0 {
		return prof.Timeline, nil
	}

	// Recompute full timeline from birth parameters
	calcReq := CalculationRequest{
		Name:           prof.Name,
		BirthDate:      prof.BirthDate,
		BirthTime:      prof.BirthTime,
		TimezoneOffset: prof.TimezoneOffset,
		NakshatraIndex: prof.NakshatraIndex,
		PadaNumber:     prof.PadaNumber,
	}
	fresh, err := CalculateFromRequest(calcReq)
	if err != nil {
		return nil, err
	}
	return fresh.Timeline, nil
}

// SaveProfile persists a chart profile to the BoltDB store.
func (e *Engine) SaveProfile(id string, data []byte) error {
	if e.store == nil {
		return nil
	}
	return e.store.SaveProfile(id, data)
}

// GetProfile retrieves a saved chart profile from the BoltDB store.
func (e *Engine) GetProfile(id string) ([]byte, error) {
	if e.store == nil {
		return nil, db.ErrNotFound
	}
	return e.store.GetProfile(id)
}

// ListProfiles returns all saved chart profile IDs.
func (e *Engine) ListProfiles() ([]string, error) {
	if e.store == nil {
		return []string{}, nil
	}
	return e.store.ListProfiles()
}

// DeleteProfile removes a profile from the BoltDB store.
func (e *Engine) DeleteProfile(id string) error {
	if e.store == nil {
		return nil
	}
	return e.store.DeleteProfile(id)
}
