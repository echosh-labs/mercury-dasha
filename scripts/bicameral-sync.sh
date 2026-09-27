#!/usr/bin/env bash
# ==============================================================================
# MERCURY DASHA // BICAMERAL BRAIN FEDERATION & CODE SYNC UTILITY
# ==============================================================================
# Links and harmonizes the Left Hemisphere (Windows Antigravity IDE) and
# Right Hemisphere (WSL2 Ubuntu Antigravity CLI) into a unified substrate.
# Synchronizes brain session transcripts, compiles manifest.json, and ensures
# code alignment between the canonical WSL2 repo and Windows IDE workspace.
# ==============================================================================

set -euo pipefail

BOLD='\033[1m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
MAGENTA='\033[0;35m'
NC='\033[0m'

HUB_DIR="${HOME}/.gemini/bicameral"
LEFT_BRAIN_DIR="/mnt/c/Users/justi/.gemini/antigravity/brain"
RIGHT_BRAIN_DIR="${HOME}/.gemini/antigravity-cli/brain"
RIGHT_BRAIN_ALT="${HOME}/.gemini/antigravity/brain"
MANIFEST_FILE="${HUB_DIR}/manifest.json"

CANONICAL_WSL_CODE="/home/justin/code/echosh-labs/mercury-dasha"
WINDOWS_IDE_CODE="/mnt/c/Users/justi/Documents/antigravity/wise-bardeen"

SYNC_CODE=true
for arg in "$@"; do
  case $arg in
    --no-code)
      SYNC_CODE=false
      ;;
  esac
done

echo -e "\n${BOLD}${CYAN}══════════════════════════════════════════════════════════════════${NC}"
echo -e "${BOLD}${CYAN}      BICAMERAL BRAIN & CODEBASE FEDERATION BRIDGE                ${NC}"
echo -e "${BOLD}${CYAN}══════════════════════════════════════════════════════════════════${NC}\n"

# 1. Create Hub Directory
mkdir -p "${HUB_DIR}"

# 2. Check and Link Left Hemisphere (Windows IDE)
echo -e "${BOLD}[Stage 1/4] Linking Left Hemisphere (Windows IDE Desktop):${NC}"
if [ -d "${LEFT_BRAIN_DIR}" ]; then
  ln -sfn "${LEFT_BRAIN_DIR}" "${HUB_DIR}/left_hemisphere"
  LEFT_COUNT=$(find "${LEFT_BRAIN_DIR}" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l)
  echo -e "  • Brain Path: ${CYAN}${LEFT_BRAIN_DIR}${NC}"
  echo -e "  • Status:     ${GREEN}Active & Linked (${LEFT_COUNT} sessions)${NC}"
else
  echo -e "  • Brain Path: ${YELLOW}${LEFT_BRAIN_DIR} (Not accessible)${NC}"
  LEFT_COUNT=0
fi

# 3. Check and Link Right Hemisphere (WSL2 AGY CLI)
echo -e "\n${BOLD}[Stage 2/4] Linking Right Hemisphere (WSL2 Ubuntu CLI):${NC}"
if [ -d "${RIGHT_BRAIN_DIR}" ]; then
  ln -sfn "${RIGHT_BRAIN_DIR}" "${HUB_DIR}/right_hemisphere"
  RIGHT_COUNT=$(find "${RIGHT_BRAIN_DIR}" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l)
  echo -e "  • Brain Path: ${CYAN}${RIGHT_BRAIN_DIR}${NC}"
  echo -e "  • Status:     ${GREEN}Active & Linked (${RIGHT_COUNT} sessions)${NC}"
else
  echo -e "  • Brain Path: ${YELLOW}${RIGHT_BRAIN_DIR} (Not found)${NC}"
  RIGHT_COUNT=0
fi

if [ -d "${RIGHT_BRAIN_ALT}" ]; then
  RIGHT_ALT_COUNT=$(find "${RIGHT_BRAIN_ALT}" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l)
  echo -e "  • Alt WSL:    ${CYAN}${RIGHT_BRAIN_ALT}${NC} (${RIGHT_ALT_COUNT} sessions)"
fi

# 4. Compile Unified JSON Manifest & Extract Cross-Hemisphere Context
echo -e "\n${BOLD}[Stage 3/4] Compiling Bicameral Brain Manifest & Provenance...${NC}"

python3 - << 'EOF'
import os
import json
from datetime import datetime, timezone

left_dir = "/mnt/c/Users/justi/.gemini/antigravity/brain"
right_dir = os.path.expanduser("~/.gemini/antigravity-cli/brain")
right_alt = os.path.expanduser("~/.gemini/antigravity/brain")
manifest_path = os.path.expanduser("~/.gemini/bicameral/manifest.json")

sessions = []

def scan_hemisphere(directory, hemisphere_name):
    if not os.path.isdir(directory):
        return
    for item in os.listdir(directory):
        full_path = os.path.join(directory, item)
        if not os.path.isdir(full_path):
            continue
        
        artifacts = []
        total_size = 0
        try:
            for f in os.listdir(full_path):
                fp = os.path.join(full_path, f)
                if os.path.isfile(fp):
                    total_size += os.path.getsize(fp)
                    if f.endswith(".md") and not ".metadata." in f:
                        artifacts.append(f)
            
            mtime = os.path.getmtime(full_path)
            sessions.append({
                "id": item,
                "hemisphere": hemisphere_name,
                "path": full_path,
                "mtime": mtime,
                "date": datetime.fromtimestamp(mtime, tz=timezone.utc).isoformat(),
                "size_bytes": total_size,
                "artifacts": sorted(artifacts)
            })
        except Exception:
            pass

scan_hemisphere(left_dir, "left")
scan_hemisphere(right_dir, "right")
if os.path.isdir(right_alt):
    scan_hemisphere(right_alt, "right")

# Sort descending by modification time
sessions.sort(key=lambda s: s["mtime"], reverse=True)

left_sessions = [s for s in sessions if s["hemisphere"] == "left"]
right_sessions = [s for s in sessions if s["hemisphere"] == "right"]

manifest = {
    "generated_at": datetime.now(timezone.utc).isoformat(),
    "left_hemisphere": {
        "name": "Windows Antigravity IDE",
        "path": left_dir,
        "exists": os.path.isdir(left_dir),
        "sessions_count": len(left_sessions),
        "rules_origin": "Windows Host User Global + Workspace GEMINI.md",
        "latest_session": left_sessions[0] if left_sessions else None
    },
    "right_hemisphere": {
        "name": "WSL2 Ubuntu AGY CLI",
        "path": right_dir,
        "exists": os.path.isdir(right_dir),
        "sessions_count": len(right_sessions),
        "rules_origin": "WSL POSIX Environment + Monorepo AGENTS.md",
        "latest_session": right_sessions[0] if right_sessions else None
    },
    "total_sessions": len(sessions),
    "sessions": sessions
}

os.makedirs(os.path.dirname(manifest_path), exist_ok=True)
with open(manifest_path, "w", encoding="utf-8") as f:
    json.dump(manifest, f, indent=2)

print(f"  ✔ Manifest compiled: {manifest_path} ({len(sessions)} total sessions)")
if left_sessions:
    ls = left_sessions[0]
    print(f"  • Latest Left  (Windows): {ls['id']} ({ls['date'][:19]}) | Artifacts: {ls['artifacts']}")
if right_sessions:
    rs = right_sessions[0]
    print(f"  • Latest Right (WSL2):    {rs['id']} ({rs['date'][:19]}) | Artifacts: {rs['artifacts']}")
EOF

# 5. Synchronize Codebase Between Canonical WSL2 Root and Windows IDE Workspace
echo -e "\n${BOLD}[Stage 4/4] Synchronizing Codebase Substrates...${NC}"
if [ "$SYNC_CODE" = true ] && [ -d "$CANONICAL_WSL_CODE" ] && [ -d "$WINDOWS_IDE_CODE" ]; then
  # 5a. Canonical WSL -> Windows IDE (update newer or missing files)
  rsync -au \
    --exclude='.git' \
    --exclude='node_modules' \
    --exclude='.next' \
    --exclude='.data' \
    --exclude='bin' \
    --exclude='dist' \
    --exclude='cmd/server/dist' \
    "${CANONICAL_WSL_CODE}/" "${WINDOWS_IDE_CODE}/"

  # 5b. Windows IDE -> Canonical WSL (capture files created from Windows side, e.g. docs/artifacts)
  rsync -au \
    --exclude='.git' \
    --exclude='node_modules' \
    --exclude='.next' \
    --exclude='.data' \
    --exclude='bin' \
    --exclude='dist' \
    --exclude='cmd/server/dist' \
    "${WINDOWS_IDE_CODE}/" "${CANONICAL_WSL_CODE}/"

  echo -e "  ✔ Codebase synchronized bidirectionally between WSL canonical root and Windows IDE."
else
  echo -e "  • Code synchronization skipped (${SYNC_CODE})."
fi

echo -e "\n${BOLD}${GREEN}✔ Bicameral Brain & Codebase Synchronization Complete!${NC}\n"

