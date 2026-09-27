# 🪐 mercury-dasha — Sovereign Full-Stack Ephemeris & Media Observatory
> A high-performance single-binary web service merging Jean Meeus astronomical orbital models, Vedic Vimshottari Dasha calculations, Hermetic alchemy, and local media streaming on **Port 8080**.

---

## 🌟 The Vibe
`mercury-dasha` is an observatory engine bridging ancient celestial computation with modern full-stack web engineering. It computes the sidereal Moon and planetary positions using astronomical algorithms, projects 120-year Vimshottari Dasha planetary periods, maps Vedic Panchanga (the 5 cosmic limbs) and Ayurvedic temperaments, and acts as a local media streaming hub with Server-Sent Events (SSE) telemetry metronomes.

---

## 🚀 60-Second Quickstart

```bash
# 1. Clone the repository
git clone https://github.com/echosh-labs/mercury-dasha.git
cd mercury-dasha

# 2. Setup your local environment (Optional: works 100% offline out-of-the-box)
cp .env.example .env

# 3. Boot the full-stack engine
make dev
```

Once running:
- **Observatory Portal**: Open [http://localhost:8080](http://localhost:8080)
- **Daily Planetary Ephemeris**: [http://localhost:8080/ephemeris](http://localhost:8080/ephemeris)
- **Planetary Alignment Inspection**: [http://localhost:8080/alignment](http://localhost:8080/alignment)
- **Character Sanctuary**: [http://localhost:8080/characters](http://localhost:8080/characters)
- **Health Diagnostics**: `GET http://localhost:8080/health`

> **Runs Fully Offline:**  
> All astronomical ephemeris algorithms, Dasha timelines, Vedic Panchanga charts, and alchemical correspondence matrices run completely offline with pure Go mathematics and embedded BoltDB. Zero cloud or Dropbox accounts are needed to explore and use the system!

---

## 🗺️ Interactive Tour & Web Routes

Explore the 6 core pillars of the observatory:

1. **Observatory & Media Storehouse (`/`)**:
   - **Sidereal Moon Computation**: Real-time lunar anomaly orbital model with Lahiri Ayanamsha subtraction ($23.85^\circ$).
   - **27 Nakshatras & 108 Padas**: Sanskrit names, ruling deities, geometric symbols, and acoustic syllables.
   - **Local Storehouse Crawler**: Stream local audio, video, books, and manuscripts with HTTP 206 Range seeking.

2. **Daily Planetary Ephemeris (`/ephemeris`)**:
   - **24 Chaldean Hours**: Non-linear diurnal & nocturnal planetary hours calculated from sunrise and sunset.
   - **Alchemical Expansions**: Associated sacred metals, Hermetic axioms from *The Kybalion*, and daily directives.

3. **Planetary Alignment Inspection (`/alignment`)**:
   - **Sub-Hora Calculations**: Granular 7-tier sub-horas calculated via oblique ascension ($1\text{ Asu} = 4\text{s}$).
   - **Guidance Matrix**: Strategic evaluation of harmonic endeavors vs. dissonant actions for the active planetary hour.

4. **Character Sanctuary & Timelines (`/characters`)**:
   - **Vedic Panchanga Matrix**: Calculates Vara (Solar Day), Tithi (Lunar Phase), Nakshatra (Mansion), Yoga, and Karana at birth.
   - **Ayurvedic Triad**: Primary Dosha (*Vata, Pitta, Kapha*), Gana temperament, Yoni totem, and Nadi pulse.
   - **120-Year Vimshottari Timeline**: Full multi-tier planetary progression (*Mahadasha*, *Antardasha*, *Pratyantardasha*).

5. **Foundations Storytelling Engine (`/foundations`)**:
   - Dynamic narrative synthesizer linking active astrological periods to character arcs and creative storytelling quests.

6. **AMRA Sovereign Treasury Dock (`/treasury`)**:
   - Real-time double-entry financial ledger, YouTube revenue analytics, and SaaS subscription tiers.

---

## 🏗️ Technical Architecture

```
mercury-dasha/
├── backend/                  # Go 1.24 REST API & Astronomical Engine
│   ├── cmd/server/           # Application entrypoint & embedded Next.js static asset server
│   └── internal/
│       ├── dasha/            # Meeus lunar ephemeris, Lahiri Ayanamsha, & Vimshottari algorithms
│       ├── chrono/           # 24 Chaldean planetary hours & solar anchor calculations
│       ├── api/              # REST route handlers & Server-Sent Events (SSE) metronome
│       ├── db/               # Embedded BoltDB persistent key-value store (bbolt)
│       └── dropbox/          # Optional local POSIX / cloud media crawler
├── frontend/                 # Next.js 15 (App Router + TailwindCSS)
│   ├── app/                  # Observatory, Alignment, Ephemeris, Characters, Treasury routes
│   └── components/           # Real-time celestial charts, HUD controls, audio player
├── scripts/                  # Bicameral synchronization and development utilities
└── Makefile                  # Build, test, and verification automation
```

---

## ⚙️ Environment Configuration

```env
PORT=8080
ENV=development
BOLT_DB_PATH=.data/mercury-dasha-dev.db

# Optional Dropbox Media Sync (Leave blank to run offline)
DROPBOX_LOCAL_PATH=/path/to/local/media
DROPBOX_APP_KEY=your-app-key
DROPBOX_APP_SECRET=your-app-secret
DROPBOX_REFRESH_TOKEN=your-refresh-token

# Optional YouTube Studio Integration
YOUTUBE_CLIENT_ID=your-youtube-client-id
YOUTUBE_CLIENT_SECRET=your-youtube-client-secret
YOUTUBE_REDIRECT_URL=http://localhost:8080/api/v1/youtube/auth/callback
```

---

## 🛠️ Verification & Build Commands

| Command | Action |
| :--- | :--- |
| `make dev` | Starts Next.js development server and Go backend on Port 8080. |
| `make test` | Runs unit tests across all 14 Go packages (ephemeris, Panchanga, API, BoltDB). |
| `make build` | Compiles Next.js static export and bundles into a standalone single Go binary (`bin/mercury-dasha`). |
| `make verify` | Executes complete end-to-end verification (lint, unit tests, export, and smoke probes). |

---

## 📄 License
Dual-licensed under the AGPL-3.0 and commercial enterprise licensing from [echoSH labs](https://echosh-labs.com).
