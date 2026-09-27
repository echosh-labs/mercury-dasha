# Specification & To-Do: Character Profile Schema & BoltDB Database Integration

## 1. Overview & Architectural Purpose
This module provides the core data contracts, astronomical correlation algorithms, and persistent BoltDB storage for the **Foundations Character Chronicle**. 

Instead of treating user data as flat biographical notes, this engine models human experience as a contiguous string of historical life events, where each event is dynamically bound to its exact **Vimshottari Dasha signature** (Mahadasha, Antardasha, Pratyantardasha), active alchemical sacred metal, and governing Hermetic axiom.

---

## 2. Detailed Data Contracts (`backend/internal/character/types.go`)

### 2.1 `CharacterProfile`
The sovereign root document that links biographical identity to the natal master timeline:
```go
package character

import (
	"time"
)

type CharacterProfile struct {
	ID                 string               `json:"id"`                    // Canonical key, e.g. "character:sovereign-genesis"
	NatalProfileID     string               `json:"natal_profile_id"`      // Link to "profile:sovereign-genesis"
	Name               string               `json:"name"`                  // Character or user label (e.g. "Sovereign Genesis")
	ArchetypalTitle    string               `json:"archetypal_title"`      // Derived mythic title (e.g. "The Quicksilver Sovereign")
	BirthDate          string               `json:"birth_date"`            // Local birth date (YYYY-MM-DD)
	BirthTime          string               `json:"birth_time"`            // Local birth time (HH:MM)
	TimezoneOffset     float64              `json:"timezone_offset"`       // Offset in hours (e.g. -4.0 for EDT)
	BirthTimeUTC       time.Time            `json:"birth_time_utc"`        // Canonical UTC epoch

	// Natal Astrological Imprint
	JanmaNakshatra     string               `json:"janma_nakshatra"`       // e.g. "Dhanishta"
	JanmaPada          int                  `json:"janma_pada"`            // 1 - 4
	MoonDegree         float64              `json:"moon_degree"`           // Sidereal Moon degree (0 - 360)
	StartingLord       string               `json:"starting_lord"`         // Planet ruling first cycle (e.g. "mars")

	// Alchemical & Harmonic Alignment
	PrimarySacredMetal string               `json:"primary_sacred_metal"`  // e.g. "quicksilver", "iron"
	GoverningAxiom     string               `json:"governing_axiom"`       // Primary Hermetic principle from Kybalion
	ChakraAnchor       string               `json:"chakra_anchor"`         // Energy center axis (e.g. "Muladhara -> Vishuddha")

	// Contiguous Chronicle & Traits
	ChronicleEvents    []ChronicleEvent     `json:"chronicle_events"`      // Chronologically ordered life events
	TraitMatrix        CharacterTraitMatrix `json:"trait_matrix"`          // Synthesized character traits

	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
}
```

### 2.2 `ChronicleEvent`
The discrete biographical atom representing a single formative experience:
```go
type MilestoneCategory string

const (
	MilestoneOriginImprint       MilestoneCategory = "origin_imprint"       // Birth circumstances, parental separation, infant environment
	MilestoneSeparationOrdeal    MilestoneCategory = "separation_ordeal"    // Physical, familial, or geographic separation
	MilestoneRiteOfPassage       MilestoneCategory = "rite_of_passage"      // Initiation, coming of age, vocational awakening
	MilestoneCrucibleTrial       MilestoneCategory = "crucible_trial"       // Severe loss, illness, confrontation, endurance test
	MilestoneCreativeSynthesis   MilestoneCategory = "creative_synthesis"   // Magnum opus, inventions, major artistic/philosophical works
	MilestoneKarmicConsolidation MilestoneCategory = "karmic_consolidation" // Structural mastery, ethical architecture, late-stage elder wisdom
)

type ChronicleEvent struct {
	ID             string                `json:"id"`                    // "event-{timestamp}-{short_uuid}"
	CharacterID    string                `json:"character_id"`          // "character:sovereign-genesis"
	Title          string                `json:"title"`                 // e.g. "Separation at Birth: The Maternal Hospitalization"
	EventDate      time.Time             `json:"event_date"`            // Exact historical calendar date/time (e.g. 1971-10-03)
	AgeAtEvent     float64               `json:"age_at_event_years"`    // Fractional age at event (e.g. 0.0055 years)
	Category       MilestoneCategory     `json:"category"`
	ImpactRating   int                   `json:"impact_rating"`         // Formative impact: -5 (deep trauma/crucible) to +5 (transcendental illumination)
	NarrativeText  string                `json:"narrative_text"`        // Story text recounted by user

	// Voice Chronicle Audio Reference (Optional, populated when recorded via audio pipeline)
	VoiceSessionID string                `json:"voice_session_id,omitempty"`
	VoiceSliceID   string                `json:"voice_slice_id,omitempty"`
	VoiceAudioURL  string                `json:"voice_audio_url,omitempty"`

	// Calculated Astrological Signature at EventDate
	AstroSignature AstrologicalSignature `json:"astrological_signature"`

	// Foundations Storytelling Synthesis
	AlchemicalStage string               `json:"alchemical_stage"`      // Calcinatio, Solutio, Separatio, Putrefactio, etc.
	CoreThemes      []string             `json:"core_themes"`           // Extracted narrative tags
	KeyLearnings    string               `json:"key_learnings,omitempty"`

	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
}
```

### 2.3 `AstrologicalSignature`
The exact cosmological and alchemical state computed for the specific historical date of the event:
```go
type AstrologicalSignature struct {
	HistoricalDate       time.Time `json:"historical_date"`

	// Vimshottari 3-Tier Dasha Hierarchy Active on this Date
	MahadashaLord        string    `json:"mahadasha_lord"`         // e.g. "mars" (Iron / Forge / Vital Will)
	AntardashaLord       string    `json:"antardasha_lord"`        // e.g. "saturn" (Lead / Boundary / Solitude & Separation)
	PratyantardashaLord  string    `json:"pratyantardasha_lord"`   // e.g. "mercury" (Quicksilver / Cognitive Imprint)
	MahaCycleProgress    float64   `json:"maha_cycle_progress"`    // Percent of Mahadasha completed at this date
	AntarCycleProgress   float64   `json:"antar_cycle_progress"`   // Percent of Antardasha completed at this date

	// Alchemical Metal Interplay
	PrimaryMetal         string    `json:"primary_metal"`          // e.g. "Iron (Mars)"
	CatalystMetal        string    `json:"catalyst_metal"`         // e.g. "Lead (Saturn)"
	AlchemicalRole       string    `json:"alchemical_role"`        // Dynamic description of the elemental tension

	// Hermetic Axiom & Archetypal Seed
	GoverningAxiom       string    `json:"governing_axiom"`        // Principle of Cause and Effect / Polarities
	ChakraAxis           string    `json:"chakra_axis"`            // Muladhara (Root) -> Ajna
	ThematicPhase        string    `json:"thematic_phase"`         // "Mangala Mahadasha, governed by Shani Antardasha"
	ArchetypalConflict   string    `json:"archetypal_conflict"`    // e.g. "Dynamic Forge Will vs. Cold Solitary Constraint"
	NarrativePrompt      string    `json:"narrative_prompt"`       // Foundations story prompt synthesized for this event
}
```

### 2.4 `CharacterTraitMatrix`
Synthesized traits dynamically updated as chronicle events are added or updated:
```go
type CharacterTraitMatrix struct {
	ResilienceScore    float64           `json:"resilience_score"`    // Derived from Saturn / Mars crucible trials
	IntuitiveDepth     float64           `json:"intuitive_depth"`     // Derived from Moon / Ketu reflective phases
	VolatileAgility    float64           `json:"volatile_agility"`    // Derived from Mercury / Rahu pivots
	ArchitectonicWill  float64           `json:"architectonic_will"`  // Derived from Sun / Jupiter expansion chapters

	DominantArchetypes []string          `json:"dominant_archetypes"` // e.g. ["The Impervious Craftsman", "The Solitary Warden"]
	CrucibleSummary    []CrucibleMetric  `json:"crucible_summary"`    // Distribution of events per governing planet
}

type CrucibleMetric struct {
	Planet     string  `json:"planet"`
	EventCount int     `json:"event_count"`
	MeanImpact float64 `json:"mean_impact"`
}
```

---

## 3. Database Architecture (BoltDB Integration)

### 3.1 Bucket Allocation
In `backend/internal/db/store.go`:
```go
var (
	BucketCharacters      = []byte("characters")
	BucketChronicleEvents = []byte("chronicle_events")
)
```

### 3.2 Key Strategies
1. **Character Profile**:
   - Bucket: `BucketCharacters`
   - Key: `character:{character_id}` (e.g. `character:sovereign-genesis`)
   - Value: Full serialized JSON of `CharacterProfile`.
2. **Chronicle Events**:
   - Bucket: `BucketChronicleEvents`
   - Key: `event:{character_id}:{event_timestamp_rfc3339}_{uuid}`
   - *Rationale*: Key sorting in BoltDB is lexicographical. Prefacing with the event timestamp ensures natural chronological ordering when iterating with cursors!

### 3.3 Storage Engine Interface Methods
```go
type CharacterStore interface {
	SaveCharacter(profile *CharacterProfile) error
	GetCharacter(id string) (*CharacterProfile, error)
	ListCharacters() ([]string, error)
	DeleteCharacter(id string) error

	AddChronicleEvent(characterID string, event *ChronicleEvent) error
	GetChronicleEvents(characterID string) ([]ChronicleEvent, error)
	GetChronicleEvent(characterID string, eventID string) (*ChronicleEvent, error)
	UpdateChronicleEvent(characterID string, event *ChronicleEvent) error
	DeleteChronicleEvent(characterID string, eventID string) error
}
```

---

## 4. Astrological Signature Resolver Algorithm (`backend/internal/character/engine.go`)

### Resolution Logic:
1. Fetch the user's natal `DashaProfile` via `natal_profile_id` (e.g. `profile:sovereign-genesis`).
2. Verify that `event_date >= birth_time_utc`.
3. Pass `profile.Timeline` and `event_date` into `dasha.ResolveActiveSnapshot(timeline, eventDate)`.
4. Extract:
   - `snapshot.Mahadasha.Planet` (e.g. `mars`)
   - `snapshot.Antardasha.Planet` (e.g. `saturn`)
   - `snapshot.Pratyantardasha.Planet` (e.g. `mercury`)
   - Cycle elapsed & remaining percentages.
5. Map planetary rulers to Sacred Metals in `dasha.SacredMetals`:
   - Mars -> Iron (Crucible Hammer, Courage, Forge Fire)
   - Saturn -> Lead (Lead / Saturnian Architect, Boundary, Solitude, Weight of Reality)
   - Mercury -> Quicksilver (Neural bridge, memory encoding)
6. Map to Hermetic Axioms from `dasha.HermeticAxioms`.
7. Generate Foundations `NarrativePrompt` and assign `AlchemicalStage`.

---

## 5. API Endpoints Contract (`backend/internal/api/handlers_character.go`)

| Endpoint | Method | Description |
|---|---|---|
| `/api/v1/characters` | `GET` | List all saved character profile identifiers |
| `/api/v1/characters/{id}` | `GET` | Retrieve character profile, trait matrix, and chronicle events |
| `/api/v1/characters/{id}` | `POST` / `PUT` | Save or update character profile metadata |
| `/api/v1/characters/{id}` | `DELETE` | Delete character profile and cascade-delete its chronicle events |
| `/api/v1/characters/{id}/events` | `POST` | Ingest life event, auto-resolve Dasha signature, recalculate trait matrix |
| `/api/v1/characters/{id}/events/{event_id}` | `GET` | Retrieve single event with full signature breakdown |
| `/api/v1/characters/{id}/events/{event_id}` | `PUT` | Update event narrative, date, impact, or category |
| `/api/v1/characters/{id}/events/{event_id}` | `DELETE` | Remove event and recalibrate trait matrix |
| `/api/v1/characters/{id}/signature` | `GET` | Query param `?date=YYYY-MM-DD`: Preview astrological signature before saving event |

---

## 6. Actionable Implementation Checklist (To-Do)

- [ ] **Step 1: Create Types Package** (`backend/internal/character/types.go`):
  - Implement `CharacterProfile`, `ChronicleEvent`, `AstrologicalSignature`, `CharacterTraitMatrix`.
  - Add JSON struct tags and validation helpers.
- [ ] **Step 2: Build Astrological Signature Resolver** (`backend/internal/character/engine.go`):
  - Implement `ResolveAstrologicalSignature(natalProfile *dasha.DashaProfile, eventDate time.Time) AstrologicalSignature`.
  - Implement `CalculateTraitMatrix(events []ChronicleEvent) CharacterTraitMatrix`.
- [ ] **Step 3: BoltDB Storage Implementation** (`backend/internal/db/boltdb_character.go`):
  - Initialize `BucketCharacters` and `BucketChronicleEvents`.
  - Implement CRUD with ACID transaction handling and chronological sorting.
- [ ] **Step 4: API Handler Implementation** (`backend/internal/api/handlers_character.go`):
  - Wire HTTP route handlers into `backend/internal/api/router.go`.
- [ ] **Step 5: Automated Testing**:
  - Unit tests verifying historical date resolution (`1971-10-03` -> Mars-Saturn-Mercury).
  - BoltDB integration tests for event insertion, retrieval, and trait matrix recalibration.
