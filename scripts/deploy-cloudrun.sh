#!/usr/bin/env bash
# ==============================================================================
# echosh-labs / mercury-dasha
# Cloud Run Deployment Script (Google Cloud Build + Artifact Registry)
# ==============================================================================

set -euo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null 2>&1 && pwd )"
cd "$DIR"

PROJECT_ID=$(gcloud config get-value project 2>/dev/null || echo "${GCP_PROJECT_ID:-echosh-labs-prod}")
REGION="${GCP_REGION:-us-central1}"
REPO_NAME="${ARTIFACT_REPO:-mercury-dasha}"
SERVICE_NAME="${SERVICE_NAME:-mercury-dasha}"
IMAGE_NAME="engine"
IMAGE_URI="${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/${IMAGE_NAME}:latest"

echo "=========================================================="
echo "🚀 Deploying Mercury Dasha (Unified Engine + UI) to Cloud Run"
echo "Project: $PROJECT_ID"
echo "Region:  $REGION"
echo "Service: $SERVICE_NAME"
echo "Image:   $IMAGE_URI"
echo "=========================================================="

# 1. Run local Go test suite
echo "🧪 [1/4] Running local Go test suite..."
(cd backend && go test -v ./...)

# 2. Check / Create Artifact Registry Repository
echo "📦 [2/4] Verifying Artifact Registry repository '$REPO_NAME'..."
if ! gcloud artifacts repositories describe "$REPO_NAME" --location="$REGION" --project="$PROJECT_ID" >/dev/null 2>&1; then
    echo "Creating Artifact Registry repository '$REPO_NAME'..."
    gcloud artifacts repositories create "$REPO_NAME" \
        --repository-format=docker \
        --location="$REGION" \
        --description="Docker repository for Mercury Dasha images" \
        --project="$PROJECT_ID"
else
    echo "  ✔ Artifact Registry repository '$REPO_NAME' is active."
fi

# 3. Build and push image using Google Cloud Build
echo "🔨 [3/4] Building multi-stage container image via Google Cloud Build..."
gcloud builds submit --tag "$IMAGE_URI" --project "$PROJECT_ID" .

# 4. Deploy to Cloud Run with scale-to-zero ($0.00 idle cost)
echo "⚡ [4/4] Deploying to Cloud Run service '$SERVICE_NAME'..."
gcloud run deploy "$SERVICE_NAME" \
    --image "$IMAGE_URI" \
    --region "$REGION" \
    --project "$PROJECT_ID" \
    --allow-unauthenticated \
    --memory=512Mi \
    --cpu=1 \
    --min-instances=0 \
    --max-instances=1 \
    --concurrency=250 \
    --port=8080

# 5. Live Endpoint Verification
SERVICE_URL=$(gcloud run services describe "$SERVICE_NAME" --region "$REGION" --project "$PROJECT_ID" --format 'value(status.url)' 2>/dev/null || echo "")

if [ -n "$SERVICE_URL" ]; then
    echo "=========================================================="
    echo " ✅ Mercury Dasha Cloud Run deployment complete!"
    echo " 🌐 Live Service URL: $SERVICE_URL"
    echo "=========================================================="
    echo "🩺 Testing endpoints..."
    TOKEN=$(gcloud auth print-identity-token 2>/dev/null || echo "")
    
    if [ -n "$TOKEN" ]; then
        curl -s -H "Authorization: Bearer $TOKEN" "$SERVICE_URL/" | grep -q -i "html" && echo "  ✔ UI Check PASSED: Next.js Frontend is serving at $SERVICE_URL/" || echo "  ⚠️ UI endpoint check did not return expected HTML."
        curl -s -H "Authorization: Bearer $TOKEN" "$SERVICE_URL/healthz" | grep -q "UP" && echo "  ✔ API Check PASSED: Go backend is UP and operational!" || echo "  ⚠️ Service health check did not return expected UP state."
    else
        curl -s "$SERVICE_URL/" | grep -q -i "html" && echo "  ✔ UI Check PASSED: Next.js Frontend is serving at $SERVICE_URL/" || echo "  ⚠️ UI endpoint check did not return expected HTML."
        curl -s "$SERVICE_URL/healthz" | grep -q "UP" && echo "  ✔ API Check PASSED: Go backend is UP and operational!" || echo "  ⚠️ Service health check did not return expected UP state."
    fi
fi
