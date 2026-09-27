# Mercury Dasha: Low-Level Technical Architecture & Agent Specification
**echosh-labs / mercury-dasha**  
*Document Purpose: Sovereign architectural and low-level engineering specification for autonomous agents developing new tools, endpoints, and storage models.*

---

## 1. Environment Substrate & Bicameral Topology

`mercury-dasha` executes across a **Bicameral Dual-Substrate Topology**:

* **Left Hemisphere (Windows 11 Desktop / IDE)**:
  * **Workspace Path**: `c:\Users\justi\Documents\antigravity\wise-bardeen`
  * **Agent Brain & Logs**: `C:\Users\justi\.gemini\antigravity\brain\<conversation-id>`
  * **Explorer Entry Point**: [`Makefile`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/Makefile)
* **Right Hemisphere (WSL2 Ubuntu 24.04 / Canonical Execution)**:
  * **Canonical Git Root**: `/home/justin/code/echosh-labs/mercury-dasha` (part of the `/home/justin/code/echosh-labs` monorepo)
  * **Linux Dropbox Daemon**: `/home/justin/Dropbox` (synchronizes to the cloud with zero rate limits or token expirations)
  * **AGY CLI Brain**: `/home/justin/.gemini/antigravity-cli/brain` and `~/.gemini/antigravity/brain`
  * **Execution Substrate**: All builds, unit tests, linters, and runtime processes must be invoked natively inside WSL2 via `wsl -d Ubuntu -e bash -c "..."`.
* **Corpus Callosum (Federation Hub)**:
  * Unified at `/home/justin/.gemini/bicameral/` and synchronized bidirectionally via `scripts/bicameral-sync.sh` (`make sync`).

---

## 2. The 6-Pillar Architecture & Single-Binary Paradigm

`mercury-dasha` compiles as a single, self-contained Linux ELF binary embedding a compiled Next.js 15 frontend, a Go 1.23 HTTP API, and an embedded BoltDB key-value document store.

```
+-------------------------------------------------------------------------+
|                  Mercury Dasha Sovereign Single Binary                  |
|                        (bin/mercury-dasha)                              |
+-------------------------------------------------------------------------+
|  Go 1.23 HTTP Engine (http.ServeMux / net/http)                         |
|  ├── /healthz, /api/telemetry, /api/v1/system/*                         |
|  ├── /api/dasha/* (Observatory, Nava Grahas & Alchemical Lab)            |
|  ├── /api/v1/stream/pulse (Chrono-Pulse SSE Metronome)                  |
|  ├── /api/v1/characters/* (Character Sanctuary & Video Chronicles)      |
|  ├── /api/v1/index/* (Local Dropbox POSIX Storehouse & HTTP 206)        |
|  ├── /api/v1/sessions/* (Sonic Chronicle & Precision Slicer)            |
|  ├── /api/v1/amra/*, /api/v1/youtube/* (AMRA Treasury & Studio)         |
|  └── /api/v1/meta/*, /api/v1/profiles/* (BoltDB Document Store)         |
+-------------------------------------------------------------------------+
|  Embedded Frontend Assets (//go:embed all:frontend_out/*)               |
|  ├── / (Observatory & Storehouse Command)                               |
|  ├── /alignment (Planetary Hours, Sacred Metals & Hermetic Axioms)      |
|  ├── /characters (Character Sanctuary, 120y Timeline & Video Studio)   |
|  ├── /foundations (Foundations Storytelling Engine)                     |
|  └── /treasury (AMRA Sovereign Treasury & YouTube Studio)               |
+-------------------------------------------------------------------------+
|  Embedded BoltDB (bbolt ACID Document Storehouse)                       |
|  └── Local Persistence: .data/mercury-dasha-dev.db                      |
+-------------------------------------------------------------------------+
|  POSIX Dropbox Direct Sync Engine                                       |
|  └── Local Path: /home/justin/Dropbox (Continuous Linux Daemon)         |
+-------------------------------------------------------------------------+
```

### Single-Binary Compilation Pipeline:
1. `cd frontend && npm run build`: Generates static export in `frontend/out`.
2. `mkdir -p backend/cmd/server/frontend_out && cp -r frontend/out/* backend/cmd/server/frontend_out/`: Syncs web build into the Go embed folder.
3. `cd backend && CGO_ENABLED=0 go build -o ../bin/mercury-dasha ./cmd/server/main.go`: Compiles the sovereign Linux binary.
4. At runtime, [`backend/cmd/server/main.go`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/cmd/server/main.go) mounts `frontend_out` via `fs.Sub(frontendFS, "frontend_out")`. Client-side page navigation is served directly as static HTML (`/characters/index.html`, `/alignment/index.html`, etc.).

---

## 3. Storage Layer & Interface Decoupling (`internal/db`)

Database operations are decoupled behind the [`StorageEngine`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/db/store.go#L33-L51) interface in [`backend/internal/db/store.go`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/db/store.go). Concrete BoltDB logic is isolated in [`backend/internal/db/boltdb.go`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/db/boltdb.go).

### BoltDB Bucket Schema
The store initializes and manages 11 persistent buckets:

| Bucket Byte Name | Constant Name | Payload Structure / Purpose |
| :--- | :--- | :--- |
| `"dasha_profiles"` | `BucketProfiles` | Serialized `DashaProfile` JSON structs keyed by profile ID (supports `profile:` prefix). |
| `"dropbox_index"` | `BucketDropboxIndex` | Serialized `IndexEntry` documents for indexed POSIX files. |
| `"dropbox_stats"` | `BucketDropboxStats` | Cached category counts (`text`, `audio`, `video`, `code`, `books`). |
| `"dasha_meta"` | `BucketMeta` | Freeform JSON key-value document store (`/api/v1/meta/{key}`). |
| `"dasha_events"` | `BucketEvents` | Chronological event logs. |
| `"dnd_sessions"` | `BucketSessions` | Serialized `AudioSession` records with markers and slice references. |
| `"audio_slices"` | `BucketAudioSlices` | Sub-clip metadata carved from master audio sessions. |
| `"system"` | `BucketSystem` | Persistent configuration key-values. |
| `"amra_idempotency"` | `BucketAmraIdempotency` | Payment and event transaction deduplication keys. |
| `"amra_subscriptions"`| `BucketAmraSubscriptions` | Subscription state records. |
| `"amra_ledger"` | `BucketAmraLedger` | Financial and ledger audit trail entries. |

---

## 4. Domain Subsystems & Mathematical Formulations

### 4.1 Dasha Observatory (`internal/dasha`)
* **Meeus Lunar Ephemeris Calculation**:
  * [`CalculateMoonPosition(t time.Time)`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/dasha/calculator.go) calculates the Moon's geometric longitude using Jean Meeus' truncated trigonometric series.
  * **Lahiri Ayanamsha**: Converted to Sidereal coordinates ($0.0^\circ - 360.0^\circ$) by subtracting Lahiri Ayanamsha:
    $$\text{Ayanamsha} = 23.85^\circ + \left(\frac{\text{Year} - 2000.0 + \frac{\text{DayOfYear}}{365.25}}{3600.0} \times 50.29''\right)$$
* **Zodiac Partitioning**:
  * $360^\circ$ circle divided into **27 Nakshatras** ($13^\circ 20'$ / 800 arcminutes each).
  * Each Nakshatra is divided into **4 Padas** ($3^\circ 20'$ / 200 arcminutes each), yielding 108 Padas total.
* **Vimshottari 120-Year Cycle**:
  $$\text{Ketu (7y)} \to \text{Venus (20y)} \to \text{Sun (6y)} \to \text{Moon (10y)} \to \text{Mars (7y)} \to \text{Rahu (18y)} \to \text{Jupiter (16y)} \to \text{Saturn (19y)} \to \text{Mercury (17y)}$$
* **Birth Balance Calculation**:
  $$\text{Elapsed Arc} = \theta_{\text{moon}} - \theta_{\text{nakshatra\_start}}$$
  $$\text{Elapsed Fraction} = \frac{\text{Elapsed Arc}}{13.333333^\circ}$$
  $$\text{Balance Years} = \text{PlanetDuration} \times (1 - \text{Elapsed Fraction})$$

---

### 4.2 Vedic Panchanga & Ayurvedic Triad (`internal/dasha/panchanga.go`)
Resolves the 5 Cosmic Limbs (*Pancha-Anga*) and Ayurvedic constitution at the exact birth timestamp:
1. **Vara (Solar Day & Planetary Lord)**: Day of week mapped to governing Graha ($0=\text{Sun}, 1=\text{Moon}, \dots, 6=\text{Saturn}$).
2. **Tithi (Lunar Phase)**:
   $$\text{AngleDiff} = (\theta_{\text{moon}} - \theta_{\text{sun}}) \pmod{360^\circ}$$
   $$\text{Tithi} = \left\lfloor \frac{\text{AngleDiff}}{12^\circ} \right\rfloor + 1$$
   - Shukla Paksha (Waxing, Tithis 1–15) & Krishna Paksha (Waning, Tithis 16–30).
3. **Nakshatra**: Active lunar mansion (1–27) and pada (1–4).
4. **Yoga (Luni-Solar Conjunction)**:
   $$\text{SumAngle} = (\theta_{\text{moon}} + \theta_{\text{sun}}) \pmod{360^\circ}$$
   $$\text{Yoga} = \left\lfloor \frac{\text{SumAngle}}{13.333333^\circ} \right\rfloor + 1$$
   - 27 classical yogas (e.g. *Vishkumbha*, *Ayushman*, *Siddhi*).
5. **Karana (Half-Tithi)**:
   $$\text{Karana} = \left\lfloor \frac{\text{AngleDiff}}{6^\circ} \right\rfloor + 1$$
   - 60 Karanas categorized into 7 repeating cycles (*Bava*, *Balava*, *Kaulava*, *Taitila*, *Gara*, *Vanija*, *Vishti*) and 4 fixed (*Shakuni*, *Chatushpada*, *Naga*, *Kimstughna*).
6. **Ayurvedic Triad**:
   - **Dosha**: Primary bio-energy (*Vata*, *Pitta*, *Kapha*) derived from nakshatra element.
   - **Gana**: Temperament (*Deva* / Divine, *Manushya* / Human, *Rakshasa* / Primal).
   - **Yoni Totem**: Animal archetype governing instinct and primal compatibility.
   - **Nadi**: Subtle physiological pulse (*Aadi*, *Madhya*, *Antya*).

---

### 4.3 Natal Lagna, Surya & Whole Sign Bhavas (`internal/dasha/lagna.go`)
* **Sidereal Ascendant (Lagna)**:
  Computed from geographic coordinates (Latitude $\phi$, Longitude $\lambda$) and Local Sidereal Time (LST):
  $$\text{RAMC} = \text{GMST} + \lambda$$
  $$\tan(\theta_{\text{tropical\_lagna}}) = \frac{-\cos(\text{RAMC})}{\sin(\text{RAMC})\cos(\epsilon) + \tan(\phi)\sin(\epsilon)}$$
  $$\theta_{\text{sidereal\_lagna}} = (\theta_{\text{tropical\_lagna}} - \text{Ayanamsha}) \pmod{360^\circ}$$
* **Tripod of Embodiment**:
  - **Lagna (Soul Incarnation / Physical Form)**: 1st House Anchor.
  - **Surya (Core Identity / Ego / Divine Spark)**: Sidereal Sun position.
  - **Chandra (Mind / Emotional Perception / Memory)**: Sidereal Moon position.
* **Whole Sign House Temple (12 Bhavas)**:
  - House 1 begins at the sign boundary containing the Sidereal Lagna.
  - Houses 1 through 12 correspond strictly to 12 consecutive zodiac rashis.
  - Each Bhava resolves Sanskrit title (*Tanu*, *Dhana*, *Sahaja*, etc.), governing sign, sign lord, and residing luminaries.

---

### 4.4 Nava Grahas & 7-Tier Essential Dignities (`internal/dasha/grahas.go`)
Computes topocentric sidereal longitudes for all 9 classical Grahas (*Sun, Moon, Mars, Mercury, Jupiter, Venus, Saturn, Rahu, Ketu*) and evaluates planetary condition:
* **7 Dignity Levels**:
  1. **Param Ucha (Exaltation)**: Peak degree of maximum potency (e.g. Sun at $10^\circ$ Aries, Jupiter at $5^\circ$ Cancer).
  2. **Moolatrikona (Root Trine)**: Prime operational sign and degree bracket.
  3. **Swakshetra (Own Sign / Domicile)**: Planet residing in its sovereign rulership sign.
  4. **Mitra (Friendly Sign)**: Natural and temporal friend relationship.
  5. **Sama (Neutral Sign)**: Balanced disposition.
  6. **Shatru (Enemy Sign)**: Inimical sign environment.
  7. **Neecha (Debilitation)**: Lowest degree of vibrational constriction (e.g. Sun at $10^\circ$ Libra, Mars at $28^\circ$ Cancer).

---

### 4.5 Character Sanctuary & Timeline Video Chronicles (`internal/timeline/from_character.go`)
* **Unified Character Persona**: Persisted in BoltDB `BucketProfiles` under keys `profile:<id>`, encapsulating complete astrology, ayurveda, tripod, nava grahas, and 120-year enriched Vimshottari timeline.
* **Alchemical Timeline Enrichment**: Each Dasha period is tagged with sacred metal, Hermetic axiom, story archetype, sound frequency, and character age bounds:
  $$\text{AgeStart} = \frac{T_{\text{start}} - T_{\text{birth}}}{365.25 \times 24 \text{ hrs}}$$
* **Direct Timeline-to-Video Engine Bridge**:
  - [`GenerateFromCharacterDasha`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/timeline/from_character.go) compiles a bespoke video generation timeline.
  - Configures canvas ($1920 \times 1080$ landscape or $1080 \times 1920$ portrait).
  - Automatically synthesizes ChronoBinding, audio track with resonant frequency bed, 4 narrative scene blocks (Genesis, Threshold, Crucible, Sovereignty), and live telemetry HUD overlay.
  - Dispatches to video generation worker and links resulting asset in character chronicles (`POST /api/v1/characters/{id}/chronicle`).

---

### 4.6 Alchemical Laboratory (`internal/dasha/alchemy.go`)
Resolves planetary configurations into the Western Hermetic and Alchemical traditions:
* **The 7 Sacred Metals & Kybalion Axioms**:
  1. **Gold (Sun)**: The Principle of Mentalism (*"The All is Mind; The Universe is Mental."*)
  2. **Silver (Moon)**: The Principle of Correspondence (*"As above, so below; as below, so above."*)
  3. **Quicksilver / Mercury (Mercury)**: The Principle of Vibration (*"Nothing rests; everything moves; everything vibrates."*)
  4. **Copper (Venus)**: The Principle of Polarity (*"Everything is dual; opposites are identical in nature, but different in degree."*)
  5. **Iron (Mars)**: The Principle of Rhythm (*"Everything flows, out and in; the pendulum-swing manifests in everything."*)
  6. **Tin (Jupiter)**: The Principle of Cause and Effect (*"Every Cause has its Effect; every Effect has its Cause."*)
  7. **Lead (Saturn)**: The Principle of Gender (*"Gender is in everything; everything has its Masculine and Feminine Principles."*)
* **7 Magnum Opus Stages**:
  $$\text{Calcination (Saturn)} \to \text{Dissolution (Jupiter)} \to \text{Separation (Mars)} \to \text{Conjunction (Venus)} \to \text{Fermentation (Mercury)} \to \text{Distillation (Moon)} \to \text{Coagulation (Sun)}$$

---

### 4.7 Chrono-Pulse SSE Metronome (`internal/chrono`)
* **Chaldean Planetary Hora Calculation**: Computes the 24 diurnal and nocturnal planetary hours based on solar sunrise/sunset and geographic latitude/longitude.
* **Low-Overhead SSE Broadcaster**: [`Metronome`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/chrono/metronome.go) runs a background ticker every 2 seconds emitting live hora, active sacred metal, governing axiom, story archetype, and Go memory metrics.

---

### 4.8 Local Storehouse & AMR Audio Codec (`internal/indexer`, `internal/amr`)
* **POSIX Crawler**: High-throughput asynchronous scanner targeting `/home/justin/Dropbox`.
* **Smart Noise Filter**: Automatically skips `.git`, `node_modules`, `venv`, `dist`, `.next`, and build artifacts.
* **Embedded AMR Audio Transcoder**: Zero-dependency pure-Go AMR-NB DSP engine (`internal/amr`) transcoding BlackBerry/GSM `.amr` recordings to 16-bit 8kHz PCM `.wav` on the fly with disk caching and HTTP 206 Range streaming.

---

### 4.9 AMRA Sovereign Treasury & Studio (`internal/amra`, `internal/youtube`)
* **Vedic Principle of Āmra (आम्र)**: Sacred fruition (*karma-phala*) and the *Pūrṇa Kumbha* of divine abundance.
* **Immutable Financial Ledger**: Persistent audit ledger (`BucketAmraLedger`) with SHA-256 idempotency deduplication.
* **YouTube Sovereign Uploader & Analytics**: Automated pipeline step (`YouTubeUploadStep`) and YouTube Analytics v2 feedback loop.
* **Unified Ecosystem Revenue**: Synthesizes SaaS MRR with YouTube media accruals into sovereign treasury reserves.

---

## 5. Complete REST & SSE API Matrix

All routes registered in [`backend/internal/api/router.go`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/api/router.go):

| Category | Method | Path | Description / Payload |
| :--- | :--- | :--- | :--- |
| **Telemetry** | `GET` | `/healthz` | Service health status and Go runtime version. |
| **Telemetry** | `GET` | `/api/telemetry` | Go runtime memory allocations, goroutines, and GC stats. |
| **System** | `GET` | `/api/v1/system/overview` | Full system telemetry, substrate mapping, DB page allocations. |
| **System** | `GET` | `/api/v1/system/routes` | Interactive API documentation catalog. |
| **Bicameral** | `GET` | `/api/v1/agent/bicameral` | Dual-hemisphere status, session counts, instruction chain info. |
| **Bicameral** | `GET` | `/api/v1/agent/conversations`| Enumerate conversation sessions across Windows and WSL brains. |
| **Bicameral** | `GET` | `/api/v1/agent/artifact` | View raw markdown text (`?id=<session_id>&name=walkthrough.md`). |
| **Chrono** | `GET` | `/api/v1/stream/pulse` | SSE 2-second heartbeat stream emitting hora and system metrics. |
| **Dasha** | `POST`| `/api/dasha/calculate` | Calculate 120-year Vimshottari hierarchy from birth ephemeris. |
| **Dasha** | `GET` | `/api/dasha/overview` | Vimshottari system overview and 9 planetary resonance frequencies. |
| **Dasha** | `GET` | `/api/dasha/nakshatras` | List all 27 Vedic Nakshatras with deities, symbols, and padas. |
| **Dasha** | `GET` | `/api/dasha/planets` | List 9 Vimshottari Grahas with cycle years and chakra centers. |
| **Dasha** | `GET` | `/api/dasha/hora` | Active planetary hora, sacred metal, axiom, and archetype. |
| **Dasha** | `GET` | `/api/dasha/alchemy` | The 7 Sacred Metals, 7 Hermetic Axioms, and Magnum Opus stages. |
| **Characters**| `GET` | `/api/v1/characters` | Enumerate all stored character profiles with summary metrics. |
| **Characters**| `POST`| `/api/v1/characters` | Create new character profile with complete astrological math. |
| **Characters**| `GET` | `/api/v1/characters/{id}` | Retrieve comprehensive character profile by ID. |
| **Characters**| `PUT` | `/api/v1/characters/{id}` | Update character attributes and recalculate timeline. |
| **Characters**| `DELETE`| `/api/v1/characters/{id}` | Delete stored character persona. |
| **Characters**| `GET` | `/api/v1/characters/{id}/timeline` | Retrieve enriched 120-year Dasha hierarchy. |
| **Characters**| `POST`| `/api/v1/characters/{id}/chronicle` | Compile video timeline manifest from active character epoch. |
| **Characters**| `GET` | `/api/v1/characters/{id}/chronicles` | Enumerate generated audiovisual chronicle manifests. |
| **Profiles** | `GET` | `/api/v1/profiles` | List legacy Dasha profiles. |
| **Profiles** | `POST`| `/api/v1/profiles/{id}` | Store or update computed Dasha profile. |
| **Profiles** | `GET` | `/api/v1/profiles/{id}` | Retrieve stored Dasha profile by ID. |
| **Profiles** | `DELETE`| `/api/v1/profiles/{id}`| Delete stored Dasha profile. |
| **AMRA** | `GET` | `/api/v1/amra/plans` | Retrieve active sovereign subscription plans. |
| **AMRA** | `GET` | `/api/v1/amra/metrics` | Unified ecosystem revenue metrics (SaaS + YouTube). |
| **AMRA** | `GET` | `/api/v1/amra/ledger` | Immutable financial audit ledger records. |
| **AMRA** | `GET` | `/api/v1/amra/gcloud/billing` | Google Cloud infrastructure burn rate and cost forecast. |
| **YouTube** | `GET` | `/api/v1/youtube/status` | Sovereign uploader authorization and channel status. |
| **YouTube** | `GET` | `/api/v1/youtube/analytics` | Channel performance, views, retention, and monetization. |
| **Storehouse**| `GET` | `/api/v1/index/status` | Current crawl status and file counts across categories. |
| **Storehouse**| `POST`| `/api/v1/index/scan` | Trigger asynchronous crawl of `/home/justin/Dropbox`. |
| **Storehouse**| `GET` | `/api/v1/index/search` | Full-text and metadata search (`?q=&cat=&ext=&limit=`). |
| **Storehouse**| `GET` | `/api/v1/index/content` | HTTP 206 Partial Content / Range media streaming. |
| **Sessions** | `GET` | `/api/v1/sessions` | List recorded audio sessions (`?campaign=`). |
| **Sessions** | `POST`| `/api/v1/sessions` | Initialize audio session record on disk and in BoltDB. |
| **Sessions** | `GET` | `/api/v1/sessions/{id}` | Retrieve session details with markers and audio slices. |
| **Sessions** | `POST`| `/api/v1/sessions/{id}/chunk`| Append raw PCM chunk and update WAV header in-place. |
| **Sessions** | `POST`| `/api/v1/sessions/{id}/slice`| Carve precision WAV slice. |
| **Sessions** | `GET` | `/api/v1/sessions/{id}/stream`| Stream session WAV audio with HTTP 206 seeking. |
| **BoltDB** | `GET` | `/api/v1/meta` | List all document keys in the `dasha_meta` bucket. |
| **BoltDB** | `GET` | `/api/v1/meta/{key}` | Retrieve raw JSON document. |
| **BoltDB** | `POST`| `/api/v1/meta/{key}` | Put or overwrite JSON document. |
| **BoltDB** | `DELETE`| `/api/v1/meta/{key}` | Delete JSON document. |
| **BoltDB** | `GET` | `/api/v1/backup` | Stream raw BoltDB snapshot as a downloadable binary. |
| **Dropbox** | `GET` | `/api/v1/dropbox/status` | Check cloud OAuth2 status and token validity. |
| **Dropbox** | `POST`| `/api/v1/dropbox/backup` | Upload BoltDB backup snapshot to `/MercuryDasha/backups`. |

---

## 6. How to Extend & Code New Tools: Agent Runbook

When an incoming agent is tasked with adding a tool, endpoint, or data model, follow this protocol:

### Step 1: Add or Modify Data Models
* If adding database tables/records, create domain structs in their respective package (e.g. `backend/internal/dasha/types.go` or `backend/internal/session/model.go`).
* If persistence is required, add a bucket constant to `backend/internal/db/boltdb.go` and include it in `Open()` bucket initialization.
* Add repository methods to the `StorageEngine` interface in `backend/internal/db/store.go`, and implement them on `*Store` in `backend/internal/db/boltdb.go`.

### Step 2: Implement Controller Handlers
* Place HTTP handler functions in `backend/internal/api/`.
* Attach handlers as methods on `(h *Handler)`. Use `h.store` for database operations, `h.engine` for Dasha math, `h.sessions` for audio, and `h.indexer` for filesystem tasks.
* Enforce proper HTTP status codes, JSON content headers, and error responses.

### Step 3: Register Routes & Update Documentation
* Register the endpoint in `backend/internal/api/router.go`. Go 1.22+ method routing is supported (`mux.HandleFunc("GET /api/v1/...", handler)`).
* Register the new endpoint in the documentation array inside `APICatalogHandler` in `backend/internal/api/handlers_system.go` so it appears automatically in the live Mission Control interface.

### Step 4: Verification & Build Commands
Always verify in WSL2 Ubuntu using the workspace commands:

```bash
# 1. Run unit tests
cd backend && go test -v ./...

# 2. Run static analysis
cd backend && go vet ./...

# 3. Compile the unified binary
make build

# 4. Run the full unified test suite
./test.sh
```

---

## 7. Key File Reference Links for Antigravity Agents

* **Root Automation Makefile**: [`Makefile`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/Makefile)
* **Unified Test Runner**: [`test.sh`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/test.sh)
* **Bicameral Sync Script**: [`scripts/bicameral-sync.sh`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/scripts/bicameral-sync.sh)
* **Server Entry Point**: [`backend/cmd/server/main.go`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/cmd/server/main.go)
* **Storage Engine Interface**: [`backend/internal/db/store.go`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/db/store.go)
* **BoltDB Store**: [`backend/internal/db/boltdb.go`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/db/boltdb.go)
* **Vedic Panchanga & Ayurveda**: [`backend/internal/dasha/panchanga.go`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/dasha/panchanga.go)
* **Natal Lagna & 12 Bhavas**: [`backend/internal/dasha/lagna.go`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/dasha/lagna.go)
* **Nava Grahas & Essential Dignities**: [`backend/internal/dasha/grahas.go`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/dasha/grahas.go)
* **Character Sanctuary Domain**: [`backend/internal/dasha/character.go`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/dasha/character.go)
* **Timeline Video Compiler**: [`backend/internal/timeline/from_character.go`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/timeline/from_character.go)
* **AMRA Sovereign Treasury**: [`backend/internal/amra/engine.go`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/amra/engine.go)
* **YouTube Uploader & Studio**: [`backend/internal/youtube/uploader.go`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/youtube/uploader.go)
* **HTTP API Router**: [`backend/internal/api/router.go`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/api/router.go)
* **Character API Handlers**: [`backend/internal/api/handlers_character.go`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/api/handlers_character.go)
* **Mission Control Handlers**: [`backend/internal/api/handlers_system.go`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/api/handlers_system.go)
