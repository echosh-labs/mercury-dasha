#!/usr/bin/env bash
set -euo pipefail

# Load from .env if present
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
if [ -f "${ROOT_DIR}/.env" ]; then
  # export non-comment lines
  set -a
  source "${ROOT_DIR}/.env"
  set +a
fi

APP_KEY="${DROPBOX_APP_KEY:-}"
APP_SECRET="${DROPBOX_APP_SECRET:-}"
REFRESH_TOKEN="${DROPBOX_REFRESH_TOKEN:-}"

echo "=== 1. Refreshing Token from Dropbox OAuth ==="
TOKEN_RESP=$(curl -s https://api.dropboxapi.com/oauth2/token \
  -d grant_type=refresh_token \
  -d refresh_token="${REFRESH_TOKEN}" \
  -u "${APP_KEY}:${APP_SECRET}")

ACCESS_TOKEN=$(echo "${TOKEN_RESP}" | grep -o '"access_token": "[^"]*' | cut -d'"' -f4)

if [ -z "${ACCESS_TOKEN}" ]; then
  echo "Failed to obtain access token: ${TOKEN_RESP}" >&2
  exit 1
fi

echo "✓ Obtained fresh access token (starts with: ${ACCESS_TOKEN:0:15}...)"

echo -e "\n=== 2. Testing Dropbox API: users/get_current_account ==="
curl -s -X POST https://api.dropboxapi.com/2/users/get_current_account \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" | python3 -m json.tool

echo -e "\n=== 3. Testing Dropbox API: files/list_folder (Root) ==="
curl -s -X POST https://api.dropboxapi.com/2/files/list_folder \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"path": "", "limit": 5}' | python3 -m json.tool
