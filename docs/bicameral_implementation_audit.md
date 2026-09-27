# 🧠 Bicameral Brain Federation Audit & Cross-Agent Hallucination Prevention

> **Substrate Perspective**: Right Hemisphere (`WSL2 Ubuntu AGY CLI`)  
> **Session ID**: `b26e16e4-5730-49f5-93ba-e34ff08f1712`  
> **Federation Hub (Corpus Callosum)**: `/home/justin/.gemini/bicameral/`  
> **Canonical Root**: `/home/justin/code/echosh-labs/mercury-dasha`  
> **Status**: Verified & Synchronized (24 Unified Sessions)

---

## 🏛️ 1. Executive Summary & Audit Overview

This audit evaluates the **Bicameral Brain Federation** implementation originally architected by the Windows-based Antigravity agent (Left Hemisphere, session `65df33eb-aefb-4206-b758-6072bfa69b3a`), reviewed and enhanced by the native WSL2 AGY CLI agent (Right Hemisphere, session `b26e16e4-5730-49f5-93ba-e34ff08f1712`).

The primary mission of the Bicameral Protocol is to **prevent cross-agent hallucinations** that occur when agents operate across two distinct environments (Windows Host Desktop vs. Linux WSL2 POSIX substrate) with isolated brain stores, disparate filesystem paths, and separate instruction chains.

```
═════════════════════════════════════════════════════════════════════════════════
                       BICAMERAL BRAIN TOPOLOGY
═════════════════════════════════════════════════════════════════════════════════
 🪟 LEFT HEMISPHERE (Windows Host IDE)       🐧 RIGHT HEMISPHERE (WSL2 AGY CLI)
 • Environment: Windows 11 Desktop           • Environment: Ubuntu Linux 24.04
 • Brain: C:\Users\justi\.gemini\antigravity • Brain: ~/.gemini/antigravity-cli
 • Active Sessions: 11                       • Active Sessions: 13
 • Rule Origin: User Global + GEMINI.md      • Rule Origin: AGENTS.md + GEMINI.md
                     │                                     │
                     └──────────────────┬──────────────────┘
                                        ▼
                   🏛️ CORPUS CALLOSUM (Federation Hub)
                     Location: ~/.gemini/bicameral/
                     ├── left_hemisphere  -> Windows Brain
                     ├── right_hemisphere -> WSL CLI Brain
                     ├── manifest.json    -> 24-Session Index
                     └── scripts/bicameral-sync.sh
                                        │
             ┌──────────────────────────┴──────────────────────────┐
             ▼                                                     ▼
    ⚙️ Go Backend Engine                                🖥️ Next.js Mission Control
    • /api/v1/agent/bicameral                           • Dual-Hemisphere Banner
    • /api/v1/agent/conversations                       • WIN / WSL Filter Pills
    • /api/v1/agent/artifact                            • Unified Artifact Viewer
═════════════════════════════════════════════════════════════════════════════════
```

---

## 🔬 2. Left & Right Brain Activity Breakdown

The unified manifest now catalogs all **24 historical sessions** across both hemispheres, including their initial prompt intent, step count, and generated artifacts.

### A. Left Hemisphere (Windows Antigravity IDE — 11 Sessions)

| Session ID | Steps | Artifacts | Intent & Core Activity |
| :--- | :---: | :--- | :--- |
| `65df33eb` | 1,281 | `implementation_plan.md`, `walkthrough.md` | **Bicameral Architecture & Core Restoration**: Restored 85,119 BoltDB documents, recovered OAuth2 refresh token, built Mission Control console, created initial `bicameral-sync.sh`, implemented unified `test.sh`. |
| `b79dfce4` | 379 | `implementation_plan.md`, `walkthrough.md` | **Dropbox Crawler & BoltDB Indexer**: Designed local POSIX Dropbox crawler (`/home/justin/Dropbox`), HTTP 206 Range media streaming, and BoltDB key-value schema. |
| `be36ae2f` | 394 | `implementation_plan.md`, `walkthrough.md` | **Dropbox OAuth2 Token Integration**: Configured offline refresh token flow, cloud backup upload endpoint (`POST /api/v1/dropbox/backup`), and credentials safety. |
| `4e609184` | 453 | `cleanup_plan.md` | **Dead Code & Script Archival**: Cleaned obsolete temporary scripts, removed unused demo files, sanitized repository root. |
| `7cf3072f` | 178 | — | **Concise Agent Protocol**: Refined agent instructions and communication parameters for fast, direct code generation. |
| `a4b89643` | 137 | `archive_route_plan.md` | **Ecosystem Archive Route**: Designed the showcase archive routes in `echosh-labs.com`. |
| `d5396a73` | 2,196 | `walkthrough.md` | **Mercury Dasha Frontend Grounding**: Initial Next.js 15 single-binary frontend implementation with Web Audio synthesizer. |
| `1f7a152e` | 1,714 | `setup_guide.md` | **Project Genesis & Single-Binary Architecture**: Initial prototyping of Go + embedded Next.js static export. |
| *Earlier* | <10 | — | Initial testing sessions (`688c2854`, `6ab2364d`, `820a11b2`). |

### B. Right Hemisphere (WSL2 Ubuntu AGY CLI — 13 Sessions)

| Session ID | Steps | Artifacts | Intent & Core Activity |
| :--- | :---: | :--- | :--- |
| `b26e16e4` *(Current)* | 105+ | `bicameral_implementation_audit.md` | **Bicameral Validation & Hallucination Elimination**: Audited Left Brain implementation, eliminated hardcoded session IDs, added transcript prompt extraction, verified 6-stage test suite. |
| `9524e105` | 146 | — | **Monorepo Audit & Master Protocol**: Restructured `/home/justin/code/echosh-labs` root, authored canonical monorepo `AGENTS.md` and root `Makefile`. |
| `230b6b19` | 1,546 | — | **Multi-Project Unification**: Built Shaolin martial arts video scraper pipeline, Foundations storyboard engine, and unified dossier routing. |
| `16ea4790` | 753 | — | **Codebase Style & Architecture Refactor**: Standardized Go repository interfaces and eliminated tight coupling between route handlers and storage. |
| `79e65483` | 583 | — | **Cloud Run Deployment Pipeline**: Set up Docker multi-stage builds and Google Artifact Registry scripts. |
| `7c5337a7` | 116 | — | **Shaolin Video Harvester**: Go scraper initialization for video ingestion. |
| `42c157b4` | 91 | — | **Database Engine Abstraction**: Storage interface design for potential PostgreSQL / Cloud SQL migration. |
| *Earlier* | <40 | — | Early exploratory sessions (`1a987384`, `7a34efcb`, `8bd7577b`, `3cfb0fbd`, `214540f2`, `40c48472`). |

---

## 🛡️ 3. How the Bicameral Bridge Prevents Cross-Agent Hallucinations

Cross-agent hallucination occurs when an AI agent in one environment makes false assumptions about the state, files, or capabilities of the project due to missing contextual feedback from another environment.

The Bicameral Protocol prevents hallucinations across **four specific failure vectors**:

### Vector 1: Substrate Path Confusion (`C:\...` vs `/home/justin/...`)
- **The Risk**: Windows agents attempt to run Windows paths (`c:\Users\justi\...`) inside Linux containers or WSL shell scripts, causing "file not found" errors or broken build targets.
- **The Fix**: Canonical grounding enforced in both `GEMINI.md` and `AGENTS.md`. The sovereign root is always `/home/justin/code/echosh-labs/mercury-dasha`. When a Windows agent runs commands, it must wrap them via `wsl -d Ubuntu -e bash -c "..."`.

### Vector 2: Brain Amnesia ("Did that work ever happen?")
- **The Risk**: A WSL agent inspecting its local `~/.gemini/antigravity-cli/brain` cannot see that a Windows agent previously wrote a 10,000-line implementation plan or recovered 85,119 files in BoltDB, leading it to waste time recreating existing systems or claiming they do not exist.
- **The Fix**: The federation bridge symlinks both brain directories into `~/.gemini/bicameral/` and compiles an authoritative `manifest.json`. The Go backend exposes `/api/v1/agent/conversations` and `/api/v1/agent/artifact`, enabling any agent or dashboard to read historical transcripts and artifacts from either brain.

### Vector 3: Hardcoded Session Drift (Session Stagnation)
- **The Risk**: In the initial Windows implementation, `ActiveSession` was hardcoded to `"65df33eb-aefb-4206-b758-6072bfa69b3a"`, causing the frontend and downstream agents to hallucinate that `65df33eb` was permanently running even when a new session was actively modifying code.
- **The Fix**: Replaced hardcoded IDs in both `handlers_system.go` and `agentic-console.tsx` with **dynamic mtime evaluation**. The backend scans both brain directories and tags the most recently updated session as `active_session`.

### Vector 4: Transcript Summarization & Prompt Transparency
- **The Risk**: Sessions executed in the CLI typically do not output standalone `.md` artifacts into the conversation root, leading the frontend to show empty states or assume the session had no output.
- **The Fix**: Added automated transcript parsing to `extractTranscriptSummary()` and `scripts/bicameral-sync.sh`. The first `<USER_REQUEST>` and step count are parsed from `.system_generated/logs/transcript.jsonl`, giving immediate visibility into what was done in every session across both hemispheres.

---

## ⚙️ 4. Verification Suite Results

The unified test runner (`./test.sh`) was executed to confirm complete system integrity across all 6 stages:

```
══════════════════════════════════════════════════════════════════
🎉 ALL SUITES PASSED CLEANLY! TOTAL TIME: 20s
   • Go Backend Tests:       10 Packages Passed
   • Go Linter (vet):        Zero Warnings
   • Frontend Typecheck:     TypeScript 0 Errors
   • Next.js Static Export:  ✓ Exporting (3/3)
   • Frontend Tests:         5/5 Tests Passed
   • Single Binary:          bin/mercury-dasha Ready
══════════════════════════════════════════════════════════════════
Ephemeral Live Smoke Probes (Port 18080):
  • [GET] /healthz:                  PASS (HTTP 200 UP)
  • [GET] /api/v1/system/overview:   PASS (service: mercury-dasha)
  • [GET] /api/v1/system/routes:      PASS (20 registered routes)
  • [GET] /api/v1/agent/conversations: PASS (24 bicameral sessions)
  • [GET] /api/v1/agent/bicameral:   PASS (dynamic active session)
  • [GET] /api/dasha/calculate:      PASS (Vimshottari calculation)
  • [GET] / (Next.js embedded SPA):  PASS (HTTP 200 OK)
```

---

## 📋 5. Summary Checklist of Applied Hardening

- [x] **`scripts/bicameral-sync.sh`**: Upgraded to extract prompt summaries and step counts from `.system_generated/logs/transcript.jsonl`, fixed `datetime.now(timezone.utc)` deprecation, dynamically sets `active_session`.
- [x] **`backend/internal/api/handlers_system.go`**:
  - Added `PromptSummary` and `StepCount` to `ConversationEntry`.
  - Added `extractTranscriptSummary()` buffered streaming scanner.
  - Dynamically calculates `ActiveSession` from directory modtimes instead of hardcoding `65df33eb`.
- [x] **`frontend/components/agentic-console.tsx`**:
  - Replaced hardcoded `isCurrent` check with `conv.id === bicameral?.active_session`.
  - Added prompt summary preview and step count to conversation cards.
  - Added dynamic empty state distinguishing Left vs. Right brain sessions.
- [x] **Verification**: Zero TypeScript errors, zero Go vet warnings, 10/10 Go packages passing, 5/5 frontend tests passing, ephemeral HTTP probes passing 100%.
