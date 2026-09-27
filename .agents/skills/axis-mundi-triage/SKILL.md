---
name: axis-mundi-triage
description: >-
  Retrieve, inspect, and integrate Google Keep notes and workspace directives from the local
  Axis Mundi MCP server (http://localhost:8080/mcp). Monitors maintenance status (Pending, Execute,
  Active, Complete) in Manual mode (on-demand user request) or Auto mode (periodic polling via
  schedule timer), triages new voice-ingested notes, and safely integrates them into the codebase.
---

# Axis Mundi Note Retrieval & Workspace Triage Skill

This skill enables Antigravity agents to interact with the local **Axis Mundi MCP Server** running at `http://localhost:8080/mcp`. It manages Google Keep notes captured via Gemini Voice, tracks workflow lifecycle states, and integrates notes into the target codebase at the user's direction.

---

## 🚀 Quick Execution Cheatsheet

| Goal | Command / Tool Call |
| :--- | :--- |
| **Poll Pending Notes** | `python3 .agents/skills/axis-mundi-triage/scripts/poll-notes.py --status Pending` |
| **Search Notes by Keyword** | `python3 .agents/skills/axis-mundi-triage/scripts/poll-notes.py --query "keyword"` |
| **Inspect Note Content** | `python3 .agents/skills/axis-mundi-triage/scripts/fetch-note.py <noteId>` |
| **Search & Inspect Note** | `python3 .agents/skills/axis-mundi-triage/scripts/fetch-note.py --search "keyword"` |
| **Claim Note (Set Active)** | `python3 .agents/skills/axis-mundi-triage/scripts/set-status.py <noteId> Active` |
| **Finalize (Set Complete)** | `python3 .agents/skills/axis-mundi-triage/scripts/set-status.py <noteId> Complete` |

---

## 🛠️ Operational Modes

### 🔹 Mode 1: Manual Mode (Interactive On-Demand)
Use this mode when the user explicitly asks:
* *"Check Axis Mundi for new notes"*
* *"Fetch the recent note about boltyaml and implement it"*
* *"What's in my Keep notes registry?"*

#### Step-by-Step Procedure:
1. **Poll & Discover**:
   Run `scripts/poll-notes.py` with `--status Pending` or `--query <keyword>`:
   ```bash
   python3 .agents/skills/axis-mundi-triage/scripts/poll-notes.py --type keep --status Pending
   ```
2. **Inspect & Triage**:
   Retrieve the complete note payload, checklist state, and timestamps:
   ```bash
   python3 .agents/skills/axis-mundi-triage/scripts/fetch-note.py <note_id>
   ```
3. **Claim Task**:
   Before modifying code, update the status to `Active` (alerts the TUI via SSE):
   ```bash
   python3 .agents/skills/axis-mundi-triage/scripts/set-status.py <note_id> Active
   ```
4. **Codebase Integration**:
   - Locate the relevant source file or create new packages/modules as specified in the note.
   - Implement the required types, functions, or architectural refactors.
   - Run verification checks (e.g. `npm run typecheck`, `npm test`, `go test ./...`).
5. **Mark Complete**:
   Once tests pass and changes are verified:
   ```bash
   python3 .agents/skills/axis-mundi-triage/scripts/set-status.py <note_id> Complete
   ```

---

### 🔹 Mode 2: Auto Mode (Periodic Polling via Cron / Schedule)
Use this mode when the user requests automated background monitoring or polling for incoming notes.

#### Step-by-Step Procedure:
1. **Setup Background Schedule**:
   Call the `schedule` tool with a recurring 5-minute or 10-minute cron expression:
   ```json
   {
     "CronExpression": "*/10 * * * *",
     "Prompt": "Check Axis Mundi MCP for newly ingested Pending notes and report any new directives."
   }
   ```
2. **Handle Inbound Notification**:
   When the schedule notification arrives, run `scripts/poll-notes.py --status Pending`.
3. **Report to User**:
   If new pending items exist, present a concise bulleted summary:
   * Title & Note ID
   * Ingestion timestamp
   * Snippet / Directive scope
   * Ask the user if they would like to proceed with code integration.

---

## 📋 Status Workflow & Transitions

Valid workflow statuses in Axis Mundi:
* **`Pending`**: Default state for newly ingested notes from voice/Keep.
* **`Execute`**: Approved for agent execution.
* **`Active`**: Agent is currently working on the codebase implementation.
* **`Blocked`**: Waiting on external dependency or user clarification.
* **`Review`**: Implementation complete, awaiting user sign-off or test execution.
* **`Complete`**: Integrated into codebase and verified.
* **`Error`**: Integration or test verification failed.

See [status-lifecycle.md](./references/status-lifecycle.md) for full state transition rules.
