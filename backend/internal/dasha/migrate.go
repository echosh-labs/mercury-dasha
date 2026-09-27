package dasha

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/echosh-labs/mercury-dasha/internal/db"
)

// MigrateProfiles scans the BoltDB profiles bucket, normalizes profile keys,
// enriches legacy profiles with Natal Astrology, Natal Alchemy, and Active Alchemical states,
// and saves the unified schema back to the database.
func MigrateProfiles(store db.StorageEngine) error {
	if store == nil {
		return nil
	}

	keys, err := store.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list profiles for migration: %w", err)
	}

	if len(keys) == 0 {
		log.Println("[Migration] No profiles found to migrate in BoltDB.")
		return nil
	}

	migratedCount := 0
	for _, k := range keys {
		raw, err := store.GetProfile(k)
		if err != nil || len(raw) == 0 {
			continue
		}

		var prof DashaProfile
		if err := json.Unmarshal(raw, &prof); err != nil {
			log.Printf("[Migration] Skipping corrupt profile %s: %v", k, err)
			continue
		}

		// Normalize key
		cleanID := strings.TrimSpace(k)
		if !strings.HasPrefix(cleanID, "profile:") {
			cleanID = "profile:" + cleanID
		}
		if prof.ID == "" || !strings.HasPrefix(prof.ID, "profile:") {
			prof.ID = cleanID
		}

		// Check if migration is needed (missing astrological or alchemical matrices)
		needsEnrichment := prof.Astrology.StartingLord == "" || prof.Alchemy.SacredMetal.ID == "" || len(prof.TimelineSummary) == 0

		if needsEnrichment {
			PopulateUnifiedMatrices(&prof)
		}

		// Ensure topocentric coordinates are anchored if this is sovereign-genesis
		if strings.Contains(cleanID, "sovereign-genesis") {
			if prof.Latitude == 0 && prof.Longitude == 0 {
				prof.Latitude = 43.1594
				prof.Longitude = -79.2469
				prof.LocationName = "St. Catharines, ON"
				prof.Timezone = "America/Toronto"
				prof.TimezoneOffset = -4.0
			}
		}

		// Save updated unified profile
		updatedBytes, err := json.Marshal(prof)
		if err != nil {
			log.Printf("[Migration] Failed to marshal unified profile %s: %v", prof.ID, err)
			continue
		}

		if err := store.SaveProfile(cleanID, updatedBytes); err != nil {
			log.Printf("[Migration] Failed to save unified profile %s: %v", cleanID, err)
			continue
		}

		// Remove old un-prefixed key if it was migrated
		if k != cleanID {
			_ = store.DeleteProfile(k)
		}

		migratedCount++
		log.Printf("[Migration] Successfully unified profile: %s (Natal Metal: %s, Axiom: %s)", cleanID, prof.Alchemy.SacredMetal.Name, prof.Alchemy.GoverningAxiom.Title)
	}

	log.Printf("[Migration] Profile unification completed. %d profiles verified and unified.", migratedCount)
	return nil
}
