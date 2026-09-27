#!/usr/bin/env bash
set -e

# Load from .env if present
if [ -f "./.env" ]; then
  set -a
  source "./.env"
  set +a
fi

CLI="./.agents/skills/dropbox-workspace-hub/scripts/dropbox-tool.sh"

echo "=== 1. Testing Account Info ==="
$CLI account

echo -e "\n=== 2. Testing Upload ==="
echo '{"engine": "mercury-dasha", "status": "operational", "features": ["search", "content_retrieval", "backup"], "timestamp": "2026-08-26"}' > /tmp/mercury-dasha-cloud-init.json
$CLI upload /tmp/mercury-dasha-cloud-init.json /MercuryDasha/meta/mercury-dasha-cloud-init.json

echo -e "\n=== 3. Testing Metadata Inspection ==="
$CLI meta /MercuryDasha/meta/mercury-dasha-cloud-init.json

echo -e "\n=== 4. Testing Search ==="
$CLI search "mercury-dasha"

echo -e "\n=== 5. Testing Temporary Link Generation ==="
$CLI link /MercuryDasha/meta/mercury-dasha-cloud-init.json

echo -e "\n=== 6. Testing Download & Content Retrieval ==="
$CLI download /MercuryDasha/meta/mercury-dasha-cloud-init.json /tmp/downloaded-test.json
cat /tmp/downloaded-test.json
echo ""

echo -e "\n=== 7. Testing Folder Listing ==="
$CLI list "/MercuryDasha/meta"
