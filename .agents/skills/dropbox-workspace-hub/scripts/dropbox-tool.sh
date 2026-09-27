#!/usr/bin/env bash
# ==============================================================================
# echosh-labs / dropbox-workspace-hub
# CLI Utility for Dropbox Search, Metadata, Download, Upload & Share Links
# ==============================================================================
set -euo pipefail

# Auto-detect and source .env if present
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
for candidate in "${SCRIPT_DIR}/../../.." "${SCRIPT_DIR}/../.." "${SCRIPT_DIR}/.." "$PWD"; do
  if [ -f "${candidate}/.env" ]; then
    set -a
    source "${candidate}/.env"
    set +a
    break
  fi
done

CACHE_FILE="/tmp/.dropbox_token_cache_${USER:-default}"

refresh_token_from_oauth() {
  local app_key="${DROPBOX_APP_KEY:-}"
  local app_secret="${DROPBOX_APP_SECRET:-}"
  local refresh_token="${DROPBOX_REFRESH_TOKEN:-}"

  if [ -z "${refresh_token}" ] || [ -z "${app_key}" ] || [ -z "${app_secret}" ]; then
    return 1
  fi

  local resp
  resp=$(curl -s -X POST https://api.dropboxapi.com/oauth2/token \
    -d grant_type=refresh_token \
    -d refresh_token="${refresh_token}" \
    -u "${app_key}:${app_secret}")

  local new_token
  new_token=$(echo "${resp}" | grep -o '"access_token": "[^"]*' | cut -d'"' -f4 || true)

  if [ -n "${new_token}" ]; then
    local now
    now=$(date +%s)
    echo "${new_token}:${now}" > "${CACHE_FILE}"
    chmod 600 "${CACHE_FILE}" 2>/dev/null || true
    echo "${new_token}"
    return 0
  fi

  return 1
}

get_access_token() {
  # 1. If explicit access token is provided and not expired, use it
  if [ -n "${DROPBOX_ACCESS_TOKEN:-}" ]; then
    echo "${DROPBOX_ACCESS_TOKEN}"
    return 0
  fi

  # 2. Check local disk cache for cached short-lived token (< 3 hours old)
  if [ -f "${CACHE_FILE}" ]; then
    local cached_data cached_token cached_time now diff
    cached_data=$(cat "${CACHE_FILE}" 2>/dev/null || true)
    cached_token="${cached_data%%:*}"
    cached_time="${cached_data##*:}"
    now=$(date +%s)
    diff=$((now - cached_time))
    if [ "${diff}" -lt 10800 ] && [ -n "${cached_token}" ]; then
      echo "${cached_token}"
      return 0
    fi
  fi

  # 3. Request fresh token from refresh_token
  local fresh
  if fresh=$(refresh_token_from_oauth); then
    echo "${fresh}"
    return 0
  fi

  echo "Error: Neither DROPBOX_ACCESS_TOKEN nor (DROPBOX_REFRESH_TOKEN + DROPBOX_APP_KEY + DROPBOX_APP_SECRET) is set." >&2
  exit 1
}

TOKEN=$(get_access_token)

API_URL="https://api.dropboxapi.com/2"
CONTENT_URL="https://content.dropboxapi.com/2"

pretty_json() {
  if command -v jq &>/dev/null; then
    jq .
  elif command -v python3 &>/dev/null; then
    python3 -m json.tool 2>/dev/null || cat
  else
    cat
  fi
}

usage() {
  cat <<EOF
Dropbox Workspace Hub CLI
Usage: $(basename "$0") <command> [arguments]

Commands:
  account                     Get current account information and quota
  list [path]                 List files and folders (default: "")
  search <query> [path]       Search files by name or query
  meta <path>                 Get detailed metadata for a file or folder
  download <remote> <local>   Download a remote Dropbox file to local path
  upload <local> <remote>     Upload a local file to remote Dropbox path
  link <remote>               Generate a temporary direct download/preview URL

Examples:
  ./dropbox-tool.sh account
  ./dropbox-tool.sh list "/MercuryDasha"
  ./dropbox-tool.sh search "backup" "/MercuryDasha"
  ./dropbox-tool.sh meta "/MercuryDasha/backups/latest.db"
  ./dropbox-tool.sh download "/MercuryDasha/assets/banner.png" "./frontend/public/banner.png"
  ./dropbox-tool.sh upload "./artifacts/mockup.png" "/MercuryDasha/storyboards/mockup.png"
  ./dropbox-tool.sh link "/MercuryDasha/storyboards/mockup.png"
EOF
  exit 1
}

CMD="${1:-}"
shift || true

case "${CMD}" in
  account)
    curl -s -X POST "${API_URL}/users/get_current_account" \
      -H "Authorization: Bearer ${TOKEN}" | pretty_json
    ;;

  list)
    DIR_PATH="${1:-}"
    curl -s -X POST "${API_URL}/files/list_folder" \
      -H "Authorization: Bearer ${TOKEN}" \
      -H "Content-Type: application/json" \
      -d "{\"path\": \"${DIR_PATH}\", \"recursive\": false, \"limit\": 100}" | pretty_json
    ;;

  search)
    QUERY="${1:-}"
    DIR_PATH="${2:-}"
    if [ -z "${QUERY}" ]; then
      echo "Error: search requires a query string." >&2
      exit 1
    fi
    curl -s -X POST "${API_URL}/files/search_v2" \
      -H "Authorization: Bearer ${TOKEN}" \
      -H "Content-Type: application/json" \
      -d "{\"query\": \"${QUERY}\", \"options\": {\"path\": \"${DIR_PATH}\", \"max_results\": 25}}" | pretty_json
    ;;

  meta)
    FILE_PATH="${1:-}"
    if [ -z "${FILE_PATH}" ]; then
      echo "Error: meta requires a remote path." >&2
      exit 1
    fi
    curl -s -X POST "${API_URL}/files/get_metadata" \
      -H "Authorization: Bearer ${TOKEN}" \
      -H "Content-Type: application/json" \
      -d "{\"path\": \"${FILE_PATH}\", \"include_media_info\": true}" | pretty_json
    ;;

  download)
    REMOTE_PATH="${1:-}"
    LOCAL_PATH="${2:-}"
    if [ -z "${REMOTE_PATH}" ] || [ -z "${LOCAL_PATH}" ]; then
      echo "Error: download requires <remote_path> and <local_path>." >&2
      exit 1
    fi
    mkdir -p "$(dirname "${LOCAL_PATH}")"
    curl -s -X POST "${CONTENT_URL}/files/download" \
      -H "Authorization: Bearer ${TOKEN}" \
      -H "Dropbox-API-Arg: {\"path\": \"${REMOTE_PATH}\"}" \
      -o "${LOCAL_PATH}"
    echo "✓ Downloaded ${REMOTE_PATH} -> ${LOCAL_PATH}"
    ;;

  upload)
    LOCAL_PATH="${1:-}"
    REMOTE_PATH="${2:-}"
    if [ -z "${LOCAL_PATH}" ] || [ -z "${REMOTE_PATH}" ]; then
      echo "Error: upload requires <local_path> and <remote_path>." >&2
      exit 1
    fi
    if [ ! -f "${LOCAL_PATH}" ]; then
      echo "Error: local file ${LOCAL_PATH} does not exist." >&2
      exit 1
    fi
    curl -s -X POST "${CONTENT_URL}/files/upload" \
      -H "Authorization: Bearer ${TOKEN}" \
      -H "Dropbox-API-Arg: {\"path\": \"${REMOTE_PATH}\", \"mode\": \"overwrite\", \"autorename\": false}" \
      -H "Content-Type: application/octet-stream" \
      --data-binary @"${LOCAL_PATH}" | pretty_json
    ;;

  link)
    REMOTE_PATH="${1:-}"
    if [ -z "${REMOTE_PATH}" ]; then
      echo "Error: link requires a remote path." >&2
      exit 1
    fi
    curl -s -X POST "${API_URL}/files/get_temporary_link" \
      -H "Authorization: Bearer ${TOKEN}" \
      -H "Content-Type: application/json" \
      -d "{\"path\": \"${REMOTE_PATH}\"}" | pretty_json
    ;;

  *)
    usage
    ;;
esac
