# mercury-dasha (Mercury Sovereign Stack)
*Single-Binary Sovereign Full-Stack Engine for [echosh-labs](https://echosh-labs.com)*

`mercury-dasha` is an ultra-high-performance single-binary web service and astrological ephemeris engine running natively on sovereign local hardware (WSL2 Ubuntu) with deployment readiness for Google Cloud Run (gen2).

---

## 🏛 6-Pillar Single-Binary Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            mercury-dasha (Go 1.23)                          │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │                    Embedded Next.js 15 Frontend                       │  │
│  │  [/] Observatory & Storehouse     [/alignment] Planetary Hours/Metals │  │
│  │  [/characters] Sanctuary & Video  [/foundations] Storytelling Engine  │  │
│  │  [/treasury] AMRA Sovereign Treasury & YouTube Studio Dock            │  │
│  └──────────────────────────────────┬────────────────────────────────────┘  │
│                                     │ (REST & SSE Stream)                   │
│  ┌──────────────────────────────────▼────────────────────────────────────┐  │
│  │                    High-Performance Go REST Engine                    │  │
│  │  • Meeus Ephemeris Dasha Engine        • Chrono-Pulse SSE Metronome   │  │
│  │  • Vedic Panchanga & Ayurveda          • Natal Lagna & Whole Bhavas   │  │
│  │  • 9 Nava Grahas & Dignities           • Character Timeline Video DSP │  │
│  │  • Hermetic Alchemy & Kybalion         • AMRA Billing & Audit Ledger  │  │
│  │  • Local POSIX Dropbox Streamer        • Mission Control Diagnostics  │  │
│  └──────────────────────────────────┬────────────────────────────────────┘  │
│                                     │                                       │
│  ┌──────────────────────────────────▼────────────────────────────────────┐  │
│  │                    Embedded BoltDB Engine (bbolt)                     │  │
│  │  • dasha_profiles (with characters)   • dasha_meta                    │  │
│  │  • dropbox_index                      • dasha_events                  │  │
│  │  • amra_ledger / amra_subscriptions   • dnd_sessions / audio_slices   │  │
│  │  (ACID Persistent Store: 85,119 files indexed)                        │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────┬───────────────────────────────────────┘
                                      │
                 ┌────────────────────┴────────────────────┐
                 ▼                                         ▼
   ┌───────────────────────────┐             ┌───────────────────────────┐
   │    Local Dropbox Engine   │             │   Dropbox Cloud API v2    │
   │    /home/justin/Dropbox   │             │   (Permanent OAuth2)      │
   │  • Official Linux Daemon  │             │ • Non-expiring refresh    │
   │  • Zero API quota cost    │             │ • Remote backup target    │
   │  • 85,119 files indexed   │             │ • 2.2 TB Pro Tier Quota   │
   └───────────────────────────┘             └───────────────────────────┘
```

---

## 🌟 Sovereign Engine Pillars & First-Class Routes

### 1. Observatory & Storehouse Command (`/`)
- Sidereal Moon astronomical computation using Meeus lunar anomaly orbital model with Lahiri Ayanamsha subtraction ($23.85^\circ$).
- 27 Nakshatras & 108 Padas with Sanskrit names, deities, symbols, and sound syllables.
- 120-Year Vimshottari Dasha 3-tier timeline (Mahadasha $\rightarrow$ Antardasha $\rightarrow$ Pratyantardasha) with elapsed cycle balance resolution.
- High-throughput POSIX crawler indexing `/home/justin/Dropbox` (85,119 files: 35,009 books, 44,098 clean code files, 3,656 audio tracks, 1,430 video recordings, 926 text manuscripts).
- Embedded AMR-NB DSP audio transcoder and HTTP 206 Range streaming.

### 2. Daily Planetary Ephemeris (`/ephemeris`)
- Dedicated, public-facing 24-hour non-linear planetary ephemeris chart.
- Day Lord, solar anchors (Sunrise, True Solar Noon, Apparent Sunset, Solar Midnight), and 24 diurnal & nocturnal Chaldean hours.
- Interactive hour cards with inline alchemical expansion (Sacred Metals, Hermetic Axioms, and Living Directives).
- Standalone daily cosmic weather entry point requiring zero user profiles or registration.

### 3. Planetary Alignment & Deep Inspection (`/alignment`)
- Dedicated full-page route for granular astrological and alchemical analysis.
- 4 strategic metric tiles (Day Lord, Solar Corrections LAST/EoT, Active Live Hora, Active Sub-Hora).
- 7 Non-Linear Sub-Horas calculated via oblique ascension and topocentric Lagna arc traversal ($1\text{ Asu} = 4\text{s}$).
- Guidance matrix: Favorable & Harmonic Endeavors vs. Cautionary & Dissonant Activities.
- Topocentric Observer Calibration drawer with browser geolocation auto-detection and persistent centroid anchors.

### 4. Unified Character Sanctuary (`/characters`)
- Comprehensive character management replacing isolated profiles.
- **Vedic Panchanga Matrix**: The 5 Cosmic Limbs calculated at birth—Vara (Solar Day), Tithi (Lunar Phase), Nakshatra (Mansion), Yoga (Luni-Solar Conjunction), and Karana (Half-Tithi).
- **Ayurvedic Triad**: Primary Dosha (*Vata, Pitta, Kapha*), Gana temperament (*Deva, Manushya, Rakshasa*), Yoni totem animal, and Nadi pulse.
- **Tripod of Embodiment & 12 Bhavas**: Sidereal Lagna (Ascendant), Surya, and Chandra with complete Whole Sign House Temple mappings.
- **9 Classical Nava Grahas & Essential Dignities**: Exact coordinates and 7-tier dignity evaluator (*Param Ucha, Moolatrikona, Swakshetra, Mitra, Sama, Shatru, Neecha*).
- **120-Year Master Timeline**: Enriched with sacred metals, Hermetic axioms, story archetypes, harmonic sound frequencies, and character age bounds.
- **Direct Video Chronicle Bridge**: Auto-compiles video timeline manifests (`internal/timeline.GenerateFromCharacterDasha`) with 4 narrative scenes, resonant audio ducking (-14dB), and live telemetry HUD overlays.

### 5. Foundations Storytelling Engine (`/foundations`)
- Dynamic Foundations narrative seed generator binding active Dasha periods to character arcs, environmental friction, and sovereign questlines.

### 6. AMRA Sovereign Treasury & Studio (`/treasury`)
- **Vedic Philosophy of Āmra (आम्र)**: Anchored in the principle of sacred fruition (*karma-phala*) and the *Pūrṇa Kumbha* of eternal divine abundance.
- **Immutable Financial Ledger**: BoltDB audit ledger (`BucketAmraLedger`) with SHA-256 idempotency deduplication.
- **YouTube Sovereign Uploader**: Resumable chunked uploader with automated pipeline integration (`YouTubeUploadStep`).
- **YouTube Analytics v2**: Real-time channel analytics (views, watch time, CPM, ad revenue) with graceful non-monetized channel handling.
- **Unified Ecosystem Revenue**: Dynamic financial metrics synthesizing SaaS subscriptions with digital media accruals into sovereign treasury reserves.

### 7. Chrono-Pulse SSE Metronome & Mission Control
- 2-second real-time Server-Sent Events ticker (`/api/v1/stream/pulse`) delivering live hora, sacred metal, axiom, and Foundations narrative seed.
- System diagnostics: RAM, Goroutines, DB pages, Uptime, and interactive live API catalog.
- Bicameral Agent Brain reader and instruction synchronizer (`make sync`).

---

## 🚀 Commands & Development (Run in WSL2)

```bash
# Harmonize bicameral brains and codebase substrates
make sync

# Run all Go unit & domain tests (14 packages)
make test

# Static analysis & typechecking
make lint

# Compile sovereign single binary (Next.js export + Go embed)
make build

# Run full unified verification suite (unit, lint, export, 18 smoke probes)
make verify

# Launch sovereign stack locally
make dev
```

---

## 🔑 Environment Configuration (`.env`)

```env
PORT=8080
ENV=development
SERVICE_NAME=mercury-dasha
BOLT_DB_PATH=.data/mercury-dasha-dev.db

# Permanent Offline OAuth2 Refresh Token
DROPBOX_APP_KEY=your-app-key
DROPBOX_APP_SECRET=your-app-secret
DROPBOX_REFRESH_TOKEN=your-refresh-token
DROPBOX_BACKUP_PATH=/MercuryDasha/backups

# Local POSIX Dropbox Root
DROPBOX_LOCAL_PATH=/home/justin/Dropbox
```

---

## 🗺️ Ecosystem Bridges & Axis Mundi MCP

`mercury-dasha` operates as the central ephemeris and media storehouse across the `echosh-labs` monorepo:
1. **Axis Mundi Integration (`POST /mcp`)**: Streamable HTTP MCP server connecting Google Keep directives, Docs, Sheets, and Calendar events.
2. **Gemini Voice Directives**: The `.agents/skills/axis-mundi-triage` pipeline ingests voice-captured directives into Mission Control.
3. **Multi-Service Agentic Workflows**: Exposes Vimshottari calculations and the 85,119-document knowledge storehouse to sibling services (`axis-mundi`, `echosh`, `foundations`, `shaolin`).
