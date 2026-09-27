# Axis Mundi Note & Status Lifecycle

This reference outlines the maintenance and triage workflow when integrating Google Keep notes and workspace directives into the codebase.

---

## 1. Status Hierarchy & State Machine

```
   [ Personal Gemini Voice Input / Mobile Google Keep ]
                          │
                          ▼
                     [ PENDING ]  (Ingested, awaiting review)
                          │
                          ├──────────────────────┐
                          ▼                      ▼
                     [ EXECUTE ]            [ BLOCKED ] (Waiting for deps)
             (Queued for code integration)       │
                          │                      ▼
                          ▼                 [ PENDING ]
                      [ ACTIVE ]
             (Agent actively implementing)
                          │
                          ├──────────────────────┐
                          ▼                      ▼
                      [ REVIEW ]              [ ERROR ] (Build/test failure)
              (PR / test verification)           │
                          │                      ▼
                          ▼                  [ ACTIVE ]
                     [ COMPLETE ]
              (Merged, verified & live)
```

---

## 2. Status Definitions

| Status | Meaning | Agent Action |
| :--- | :--- | :--- |
| `Pending` | Newly ingested note from Gemini Voice or Keep | Discoverable via `poll-notes.py --status Pending` |
| `Execute` | Approved for implementation | Priority task to claim |
| `Active` | Agent is currently implementing the code | Agent calls `set_status(id, "Active")` before starting |
| `Blocked` | Dependency or question needs human resolution | Set when external input is needed |
| `Review` | Code written, awaiting user sign-off or test run | Set after files are modified and tests pass |
| `Complete`| Directive fully integrated, tested, and verified | Final state; removes item from active triage |
| `Error` | Execution failed or tests broken | Flags item in TUI for attention |

---

## 3. Two Operational Modes

### 🔹 Mode 1: Manual Mode (Interactive / On-Demand)
The user prompts:
* *"Check Axis Mundi for new notes"*
* *"Retrieve note about boltyaml and integrate it"*
* *"What notes are in Execute status?"*

**Execution Steps**:
1. Run `poll-notes.py --status Pending` (or specific search).
2. Fetch full body with `fetch-note.py <id>`.
3. Set status to `Active`: `set-status.py <id> Active`.
4. Locate target files in the repository and apply edits.
5. Run tests / build verification.
6. Advance status to `Complete`: `set-status.py <id> Complete`.

### 🔹 Mode 2: Auto Mode (Periodic Polling via Cron/Schedule)
To run automated background checks:
1. Schedule a recurring cron notification using the `schedule` tool:
   ```json
   {
     "CronExpression": "*/10 * * * *",
     "Prompt": "Poll Axis Mundi MCP for new Pending/Execute notes and report status."
   }
   ```
2. When the timer triggers, execute `poll-notes.py --status Pending`.
3. If new directives are found, alert the user with the summary and request confirmation to integrate.
