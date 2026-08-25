#!/usr/bin/env bash
# ==============================================================================
# echosh-labs / mercury-dasha
# Live API Demonstration & Verification Script
# ==============================================================================
set -euo pipefail

TARGET_URL="${1:-http://localhost:8080}"
echo "=================================================================="
echo "🎯 Testing mercury-dasha Engine at: ${TARGET_URL}"
echo "=================================================================="

# 1. Health & Database Engine Telemetry
echo -e "\n1. [GET] /healthz -> Checking Service Telemetry & BoltDB Stats"
curl -s -X GET "${TARGET_URL}/healthz" | jq . || curl -s -X GET "${TARGET_URL}/healthz"

# 2. Write a JSON Document
echo -e "\n\n2. [POST] /api/v1/meta/agent:dasha:config -> Writing JSON Document"
curl -s -X POST "${TARGET_URL}/api/v1/meta/agent:dasha:config" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "dasha-01",
    "version": "1.0.0",
    "framework": "echosh-labs",
    "model_settings": {
      "temperature": 0.2,
      "max_tokens": 4096
    },
    "tags": ["production", "mercury", "engine"]
  }' | jq . || true

# 3. Write a Second Document
echo -e "\n3. [POST] /api/v1/meta/system:cluster:01 -> Writing Cluster Metadata"
curl -s -X POST "${TARGET_URL}/api/v1/meta/system:cluster:01" \
  -H "Content-Type: application/json" \
  -d '{
    "region": "us-central1",
    "max_scale": 1,
    "storage": "filestore-nfs"
  }' | jq . || true

# 4. List All Keys in BoltDB
echo -e "\n4. [GET] /api/v1/meta -> Listing All Document Keys"
curl -s -X GET "${TARGET_URL}/api/v1/meta" | jq . || curl -s -X GET "${TARGET_URL}/api/v1/meta"

# 5. Query Specific Document by Key
echo -e "\n\n5. [GET] /api/v1/meta/agent:dasha:config -> Fetching Document"
curl -s -X GET "${TARGET_URL}/api/v1/meta/agent:dasha:config" | jq . || curl -s -X GET "${TARGET_URL}/api/v1/meta/agent:dasha:config"

# 6. Prefix Search
echo -e "\n\n6. [GET] /api/v1/meta?prefix=agent: -> Prefix Search"
curl -s -X GET "${TARGET_URL}/api/v1/meta?prefix=agent:" | jq . || curl -s -X GET "${TARGET_URL}/api/v1/meta?prefix=agent:"

# 7. Live Database Snapshot Download
echo -e "\n\n7. [GET] /api/v1/backup -> Verifying Live Snapshot Download"
HTTP_CODE=$(curl -s -o /tmp/mercury-test-backup.db -w "%{http_code}" "${TARGET_URL}/api/v1/backup")
if [ "${HTTP_CODE}" -eq 200 ]; then
  FILE_SIZE=$(ls -lh /tmp/mercury-test-backup.db | awk '{print $5}')
  echo "✅ Success: Live snapshot downloaded (${FILE_SIZE})"
  rm -f /tmp/mercury-test-backup.db
else
  echo "❌ Snapshot failed with HTTP status ${HTTP_CODE}"
fi

echo -e "\n=================================================================="
echo "🎉 Verification Complete! Open in browser: ${TARGET_URL}"
echo "=================================================================="
