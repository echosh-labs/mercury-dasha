---
name: dropbox-workspace-hub
description: Manage, search, inspect, upload, download, and synchronize Dropbox files, storyboards, and metadata within the Antigravity ecosystem. Use when finding Dropbox files, syncing design assets/storyboards into frontend components, downloading remote fixtures, or archiving local artifacts to Dropbox cloud.
---

# Dropbox Workspace Hub Skill

This skill guides Antigravity agents in interacting seamlessly with Dropbox cloud storage across Windows and WSL2 environments. It enables searching remote directories, inspecting file metadata, syncing assets into frontend storyboards, generating streaming preview links, and managing cloud archives.

---

## 🛠️ 1. Quick Reference & CLI Tooling

The skill provides helper scripts located in `.agents/skills/dropbox-workspace-hub/scripts/`:

- **WSL / Bash**: `.agents/skills/dropbox-workspace-hub/scripts/dropbox-tool.sh`
- **Windows PowerShell**: `.agents/skills/dropbox-workspace-hub/scripts/dropbox-tool.ps1`

### Available Commands:
```bash
# Account identity & quota
./dropbox-tool.sh account

# List directory contents
./dropbox-tool.sh list "/MercuryDasha"

# Search files by query / extension
./dropbox-tool.sh search "storyboard" "/MercuryDasha"

# Inspect detailed metadata & media info
./dropbox-tool.sh meta "/MercuryDasha/assets/banner.png"

# Download remote file to local workspace
./dropbox-tool.sh download "/MercuryDasha/assets/logo.svg" "./frontend/public/logo.svg"

# Upload local artifact/mockup to Dropbox
./dropbox-tool.sh upload "./artifacts/mockup.png" "/MercuryDasha/storyboards/mockup.png"

# Generate temporary direct streaming link (valid 4 hours)
./dropbox-tool.sh link "/MercuryDasha/storyboards/mockup.png"
```

---

## 🔍 2. Searching & Finding Files

When requested to locate assets, datasets, or documents in Dropbox:

1. **Find by Keyword / Extension**:
   ```bash
   # Search globally across Dropbox
   wsl -d Ubuntu -e bash -c "./.agents/skills/dropbox-workspace-hub/scripts/dropbox-tool.sh search 'storyboard'"
   
   # Or via PowerShell
   .\.agents\skills\dropbox-workspace-hub\scripts\dropbox-tool.ps1 search "diagram" "/MercuryDasha"
   ```

2. **List Folder Contents**:
   ```bash
   wsl -d Ubuntu -e bash -c "./.agents/skills/dropbox-workspace-hub/scripts/dropbox-tool.sh list '/MercuryDasha/assets'"
   ```

---

## 🎨 3. Frontend Storyboards & Media Asset Integration

When integrating Dropbox assets into the Next.js / React frontend:

### A. Syncing Remote Assets into Local Frontend
Download images, SVGs, or JSON fixtures into `frontend/public/` so they are immediately accessible in components:
```bash
# Download to frontend public directory
./.agents/skills/dropbox-workspace-hub/scripts/dropbox-tool.sh download \
  "/MercuryDasha/storyboards/hero-illustration.png" \
  "./frontend/public/assets/hero-illustration.png"
```
In your React/Next.js component:
```tsx
<img src="/assets/hero-illustration.png" alt="Hero Storyboard" className="rounded-xl shadow-lg" />
```

### B. Embedding Cloud Previews in Antigravity Artifacts
To render a live preview of a remote Dropbox file in an agent artifact (`walkthrough.md` or `storyboard.md`), generate a temporary direct link:
```bash
PREVIEW_URL=$(./.agents/skills/dropbox-workspace-hub/scripts/dropbox-tool.sh link "/MercuryDasha/storyboards/architecture-diagram.png")
```
Embed in markdown artifacts:
```markdown
![Architecture Diagram](${PREVIEW_URL})
```

---

## 📦 4. Archiving Artifacts & Database Snapshots

To export local build outputs, BoltDB snapshots, or generated designs to Dropbox:

```bash
# 1. Take a local snapshot or generate an artifact
curl -s http://localhost:8080/api/v1/backup -o /tmp/mercury-backup.db

# 2. Upload to target Dropbox directory
./.agents/skills/dropbox-workspace-hub/scripts/dropbox-tool.sh upload \
  "/tmp/mercury-backup.db" \
  "/MercuryDasha/backups/mercury-dasha-manual-snapshot.db"
```

---

## 🔒 5. Environment & Permissions Reference

### Authentication Modes:
1. **Offline Refresh Token (Recommended & Configured)**:
   - `DROPBOX_APP_KEY`
   - `DROPBOX_APP_SECRET`
   - `DROPBOX_REFRESH_TOKEN`
   *Never expires.* Both the Go engine and `dropbox-tool.sh` automatically exchange this refresh token for short-lived access tokens behind the scenes.
2. **Short-Lived Access Token (Legacy / Direct Override)**:
   - `DROPBOX_ACCESS_TOKEN` (valid for 4 hours)

### Local File System Engine (Zero API Overhead):
- **WSL2 Local Path**: `/home/justin/Dropbox`
- **Windows Local Path**: `C:\Users\justi\Dropbox`
- **Native Sync Daemon**: Controlled via `/usr/local/bin/dropbox status`.
- Files written directly to `/home/justin/Dropbox` are immediately synced to the Dropbox cloud by the running daemon without consuming API rate limits.

### Required Cloud API Scopes:
- `files.metadata.read` — Required for search, listing, and metadata inspection.
- `files.content.read` — Required for downloading files and generating temporary links.
- `files.content.write` — Required for uploading backups and saving storyboards.
- `account_info.read` — Required for account info and quota queries.
- **Path Standard**: Use forward slashes `/` for Dropbox remote paths (e.g. `/MercuryDasha/storyboards`).

