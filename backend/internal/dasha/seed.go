package dasha

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/echosh-labs/mercury-dasha/internal/db"
)

// AlchemyCatalog represents the static knowledge base of the Alchemical Laboratory.
type AlchemyCatalog struct {
	Philosophy       string            `json:"philosophy"`
	SovereignElement string            `json:"sovereign_element"`
	SacredMetals     []SacredMetal     `json:"sacred_metals"`
	HermeticAxioms   []HermeticAxiom   `json:"hermetic_axioms"`
	MagnumOpusStages []MagnumOpusStage `json:"magnum_opus_stages"`
}

// SeedSovereignStorehouse idempotently populates BoltDB buckets for Nakshatras and Alchemy.
func SeedSovereignStorehouse(store db.StorageEngine) error {
	if store == nil {
		return fmt.Errorf("storage engine is nil")
	}

	// 1. Seed Nakshatras Catalog if missing
	nakData, err := store.GetNakshatras()
	if err != nil || len(nakData) == 0 {
		catalog := map[string]any{
			"count":      len(Nakshatras),
			"total":      len(Nakshatras),
			"nakshatras": Nakshatras,
		}
		data, err := json.Marshal(catalog)
		if err != nil {
			return fmt.Errorf("failed to serialize nakshatras catalog: %w", err)
		}
		if err := store.SaveNakshatras(data); err != nil {
			return fmt.Errorf("failed to persist nakshatras to bolt storehouse: %w", err)
		}
		log.Printf("Storehouse: Seeded %d Nakshatras and 108 Padas into BoltDB (dasha_nakshatras)", len(Nakshatras))
	}

	// 2. Seed Alchemical Laboratory Catalog if missing
	alchemyData, err := store.GetAlchemy()
	if err != nil || len(alchemyData) == 0 {
		catalog := AlchemyCatalog{
			Philosophy:       "Spagyric Transmutation and Hermetic Ephemeris Architecture",
			SovereignElement: "Quicksilver (Mercury / Budha) — The Universal Volatile-to-Fixed Bridge",
			SacredMetals:     SacredMetals,
			HermeticAxioms:   HermeticAxioms,
			MagnumOpusStages: MagnumOpusStages,
		}
		data, err := json.Marshal(catalog)
		if err != nil {
			return fmt.Errorf("failed to serialize alchemy catalog: %w", err)
		}
		if err := store.SaveAlchemy(data); err != nil {
			return fmt.Errorf("failed to persist alchemy to bolt storehouse: %w", err)
		}
		log.Printf("Storehouse: Seeded 7 Sacred Metals and 7 Hermetic Axioms into BoltDB (dasha_alchemy)")
	}

	return nil
}
