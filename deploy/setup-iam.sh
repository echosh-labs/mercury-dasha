#!/usr/bin/env bash
# ==============================================================================
# echosh-labs / mercury-dasha
# Least-Privilege IAM Provisioning Script for Cloud Run Engine
# ==============================================================================
set -euo pipefail

# Configuration Defaults
PROJECT_ID="${GCP_PROJECT_ID:-$(gcloud config get-value project)}"
REGION="${GCP_REGION:-us-central1}"
SA_NAME="mercury-dasha-sa"
SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"

echo "==> Configuring IAM for echosh-labs / mercury-dasha on project: ${PROJECT_ID}"

# 1. Enable Required Google Cloud APIs
echo "==> Enabling required GCP APIs..."
gcloud services enable \
    run.googleapis.com \
    artifactregistry.googleapis.com \
    file.googleapis.com \
    vpcaccess.googleapis.com \
    compute.googleapis.com \
    logging.googleapis.com \
    monitoring.googleapis.com \
    --project="${PROJECT_ID}"

# 2. Create Service Account if not exists
if ! gcloud iam service-accounts describe "${SA_EMAIL}" --project="${PROJECT_ID}" >/dev/null 2>&1; then
    echo "==> Creating dedicated Service Account: ${SA_EMAIL}"
    gcloud iam service-accounts create "${SA_NAME}" \
        --display-name="mercury-dasha Engine Runtime SA (echosh-labs)" \
        --project="${PROJECT_ID}"
else
    echo "==> Service Account ${SA_EMAIL} already exists."
fi

# 3. Assign Least-Privilege IAM Roles
ROLES=(
    "roles/logging.logWriter"         # Cloud Logging
    "roles/monitoring.metricWriter"   # Cloud Monitoring
    "roles/vpcaccess.user"            # VPC Access for NFS Filestore mounting
    "roles/storage.objectViewer"      # Optional: Read bucket configs / backups
)

echo "==> Binding IAM roles to ${SA_EMAIL}..."
for ROLE in "${ROLES[@]}"; do
    echo "    - Assigning ${ROLE}"
    gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
        --member="serviceAccount:${SA_EMAIL}" \
        --role="${ROLE}" \
        --condition=None \
        --quiet
done

# 4. Optional: Allow unauthenticated public access if service is public
# (Uncomment below if deploying as a public web dashboard)
# gcloud run services add-iam-policy-binding mercury-dasha \
#     --region="${REGION}" \
#     --member="allUsers" \
#     --role="roles/run.invoker" \
#     --project="${PROJECT_ID}"

echo "==> [mercury-dasha] IAM Setup Completed Successfully!"
