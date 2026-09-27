# Mercury Dasha Workspace Rules & Canonical Agent Guidelines
# echosh-labs / mercury-dasha

## 🌌 1. Environment Substrate & Single Source of Truth
- **Canonical Repository Root**: `/home/justin/code/echosh-labs/mercury-dasha` (within the `/home/justin/code/echosh-labs` monorepo).
- **Host Execution Substrate**: WSL2 Ubuntu (`wsl -d Ubuntu -e bash -c "..."`). All builds, tests, and processes MUST be executed natively within WSL2.
- **Windows File Paths (`C:\...`)**: Non-canonical artifacts. Never treat Windows paths as the sovereign target.
- **Local Storage Engine**: `/home/justin/Dropbox` (POSIX). The official Linux Dropbox daemon runs continuously (`python3 ~/dropbox.py status` -> `Up to date`), automatically synchronizing local disk writes to the Dropbox cloud with zero API rate limits or token expirations.
- **Remote Cloud API Fallback**: Permanent OAuth2 auto-refresh integration using App Key, App Secret, and offline Refresh Token configured in `.env`.

---

## 🏛️ 2. Architectural Structure: The 6-Pillar Single Binary
`mercury-dasha` is a sovereign single-binary full-stack service merging a compiled Next.js 15 frontend with a high-performance Go 1.23 API and an embedded ACID key-value document store (`bbolt`).

1. **Dasha Observatory (`internal/dasha`)**:
   - Astronomical Sidereal Moon ephemeris calculator using Meeus lunar orbital anomaly with Lahiri Ayanamsha subtraction.
   - Authoritative dataset for all 27 Nakshatras, 108 Padas, and 9 Vimshottari Grahas.
   - 120-year 3-tier timeline hierarchy (Mahadasha -> Antardasha -> Pratyantardasha) with elapsed balance calculation at birth.
   - Dynamic Foundations StoryContext seed generation.
2. **Alchemical Laboratory (`internal/dasha/alchemy.go`)**:
   - The 7 Sacred Metals with Quicksilver/Mercury prominence as the volatile-to-fixed bridge.
   - The 7 Hermetic Axioms from *The Kybalion*.
   - The 7 Magnum Opus transformation stages and dynamic planetary alignment resolver.
3. **Chrono-Pulse SSE Metronome (`internal/chrono`)**:
   - Low-overhead Server-Sent Events ticker broadcasting live planetary hora, active sacred metal, governing axiom, and story archetype every 2 seconds (`/api/v1/stream/pulse`).
4. **Dropbox Sovereign Storehouse (`internal/indexer`, `internal/dropbox`)**:
   - High-throughput POSIX crawler with smart code noise filter (ignoring node_modules, .git, venv).
   - BoltDB persistent index tracking 85,119 documents (35,009 books, 44,098 code files, 3,656 audio tracks, 1,430 videos, 926 text notes).
   - In-app media streaming via `/api/v1/index/content` with HTTP 206 Range seeking for audio and video.
   - Cloud backup snapshot synchronizer to `/MercuryDasha/backups`.
5. **Agentic Mission Control (`internal/api/handlers.go`)**:
   - System runtime telemetry (RAM, Goroutines, DB pages, Uptime).
   - In-app Agent Brain and conversation report reader (`walkthrough.md`, `implementation_plan.md`).
   - Interactive live API catalog runner.
6. **AMRA Sovereign Treasury & Studio (`internal/amra`, `internal/youtube`)**:
   - Rooted in the Vedic principle of *Āmra* (आम्र—sacred fruition, *karma-phala*, and the overflowing *Pūrṇa Kumbha* of divine abundance and immortality).
   - Immutable financial audit ledger (`BucketAmraLedger`) with SHA-256 idempotency deduplication.
   - YouTube Sovereign Uploader, automated build-pipeline step, and YouTube Analytics v2 feedback loop with monetization tracking.
   - Unified ecosystem revenue calculation synthesizing SaaS subscriptions with digital media accruals into sovereign treasury reserves.

---

## 🛠️ 3. Build, Test & Verification Commands (Run in WSL)
- **Backend Tests**: `cd backend && go test -v ./...`
- **Frontend Build**: `cd frontend && npm run build`
- **Compile Single Binary**: `cd backend && CGO_ENABLED=0 go build -o ../bin/mercury-dasha ./cmd/server/main.go`
- **Full Verification**: `make verify`
- **Local Dev Server**: `PORT=8080 BOLT_DB_PATH=.data/mercury-dasha-dev.db ./bin/mercury-dasha`

---

## 🔒 4. Code & Repository Conventions
- Store API handlers in `backend/internal/api/` and decouple database operations via `db.StorageEngine` in `backend/internal/db/store.go`.
- Embed compiled Next.js export bundles via Go `embed.FS` in `backend/cmd/server/main.go`.
- Keep `.gitignore` strictly enforced: never track `.data/`, `*.db`, `bin/`, `*.exe`, `node_modules/`, `dist/`, or `.env`.

---

## 🧠 5. Bicameral Brain & Instruction Provenance Protocol
Antigravity operates in a dual-substrate bicameral topology:
- **Left Hemisphere (Windows Host IDE)**:
  - Substrate: Windows 11 Desktop Antigravity IDE
  - Brain Path: `C:\Users\justi\.gemini\antigravity\brain` (accessible in WSL via `/mnt/c/Users/justi/.gemini/antigravity/brain`)
  - Primary Rules: User Global Rules (`<RULE[user_global]>`) + Workspace `GEMINI.md`
- **Right Hemisphere (WSL2 Ubuntu AGY CLI)**:
  - Substrate: WSL2 Ubuntu Native POSIX Shell
  - Brain Path: `/home/justin/.gemini/antigravity-cli/brain` (and `~/.gemini/antigravity/brain`)
  - Primary Rules: Monorepo Root `AGENTS.md` + Workspace `GEMINI.md`
- **Corpus Callosum (Federation Hub)**:
  - Unified at `/home/justin/.gemini/bicameral/` and synchronized via `scripts/bicameral-sync.sh` (`make sync`).
  - Agents in either hemisphere must maintain provenance awareness: recognize whether executing on Windows or WSL, and always ground execution against the canonical root: `/home/justin/code/echosh-labs/mercury-dasha`.

### 🔄 5.1 The "Synchronize & Go" First-Touch Protocol (MANDATORY ON SESSION START)
Whenever an agent begins working in this repository (first user prompt or after context refresh):
1. **Identify Substrate**: Recognize immediately whether executing from Windows Host IDE (`wise-bardeen`) or WSL2 Native POSIX (`/home/justin/code/echosh-labs/mercury-dasha`).
2. **Execute Bicameral Synchronization Proactively**: Run `make sync` (or `./scripts/bicameral-sync.sh`) without prompting or waiting for user instructions. This:
   - Harmonizes brain transcripts across Left (`/mnt/c/Users/justi/.gemini/antigravity/brain`) and Right (`~/.gemini/antigravity-cli/brain`).
   - Compiles `/home/justin/.gemini/bicameral/manifest.json`.
   - Bidirectionally synchronizes codebase files between the WSL canonical repo and Windows IDE workspace.
3. **Inspect Latest Cross-Hemisphere Work**: Examine the top session of the *other* hemisphere from the sync script output or `manifest.json`. Recall recent accomplishments, artifacts (`walkthrough.md`, `implementation_plan.md`), and tests completed on the other side before proposing new changes.
4. **Autonomous Grounding**: Never ask the user "Should I synchronize?". Execute this synchronization proactively as the very first grounding action.


