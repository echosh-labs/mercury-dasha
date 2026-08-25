#!/usr/bin/env bash
# ==============================================================================
# echosh-labs / mercury-dasha
# Cloud Filestore & VPC Connector Provisioning Script
# NOTE: Used for interim BoltDB NFS volume mounting. Slated for retirement
# once the managed Cloud SQL / Firestore database migration (Phase 2) is active.
# ==============================================================================
set -euo pipefail

PROJECT_ID="${GCP_PROJECT_ID:-$(gcloud config get-value project)}"
REGION="${GCP_REGION:-us-central1}"
ZONE="${GCP_ZONE:-us-central1-a}"
NETWORK="${GCP_NETWORK:-default}"
CONNECTOR_NAME="mercury-vpc-connector"
FILESTORE_INSTANCE="mercury-filestore"
FILESHARE_NAME="mercury_share"

echo "==> Setting up Persistent Storage & Networking for mercury-dasha..."

# 1. Create Serverless VPC Access Connector (if not exists)
if ! gcloud compute networks vpc-access connectors describe "${CONNECTOR_NAME}" --region="${REGION}" --project="${PROJECT_ID}" >/dev/null 2>&1; then
    echo "==> Creating VPC Connector ${CONNECTOR_NAME} in region ${REGION}..."
    gcloud compute networks vpc-access connectors create "${CONNECTOR_NAME}" \
        --region="${REGION}" \
        --range="10.8.0.0/28" \
        --network="${NETWORK}" \
        --min-instances=2 \
        --max-instances=3 \
        --project="${PROJECT_ID}"
else
    echo "==> VPC Connector ${CONNECTOR_NAME} already exists."
fi

# 2. Create Cloud Filestore NFS Instance (if not exists)
if ! gcloud filestore instances describe "${FILESTORE_INSTANCE}" --zone="${ZONE}" --project="${PROJECT_ID}" >/dev/null 2>&1; then
    echo "==> Creating Cloud Filestore Instance ${FILESTORE_INSTANCE}..."
    gcloud filestore instances create "${FILESTORE_INSTANCE}" \
        --zone="${ZONE}" \
        --tier=BASIC_HDD \
        --file-share=name="${FILESHARE_NAME}",capacity=1TB \
        --network=name="${NETWORK}" \
        --project="${PROJECT_ID}"
else
    echo "==> Filestore Instance ${FILESTORE_INSTANCE} already exists."
fi

# 3. Retrieve Filestore IP Address
NFS_IP=$(gcloud filestore instances describe "${FILESTORE_INSTANCE}" --zone="${ZONE}" --project="${PROJECT_ID}" --format="value(networks.ipAddresses[0])")
echo "=================================================================="
echo "Storage Provisioning Complete!"
echo "Filestore IP: ${NFS_IP}"
echo "Fileshare:    /${FILESHARE_NAME}"
echo "VPC Connector: projects/${PROJECT_ID}/locations/${REGION}/connectors/${CONNECTOR_NAME}"
echo ""
echo "Update deploy/service.yaml with:"
echo "  server: ${NFS_IP}"
echo "  path: /${FILESHARE_NAME}"
echo "=================================================================="
