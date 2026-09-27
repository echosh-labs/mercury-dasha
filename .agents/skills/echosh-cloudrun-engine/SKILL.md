---
name: echosh-cloudrun-engine
description: Runbook for developing, building, verifying, and deploying echosh-labs single-binary services (Go API + embedded Next.js static export) on Google Cloud Run.
---

# Echosh-Labs Cloud Run Engine Workflow

This skill outlines the standard development, verification, containerization, and deployment procedures for single-binary services within the `echosh-labs` ecosystem (e.g., `mercury-dasha`, `axis-mundi`).

---

## 1. Local Verification Chain

Always run the full verification pipeline before committing or deploying changes:

```bash
# 1. Run backend unit & integration tests
cd backend && go test -v ./...

# 2. Verify frontend Next.js static export compilation
cd frontend && npm run build

# 3. Full verification target
make verify
```

---

## 2. Multi-Stage Docker Build Architecture

All `echosh-labs` single-binary services use a multi-stage Docker build pipeline:

1. **Stage 1 (`frontend-builder`)**: Node 20 Alpine builds the Next.js static bundle (`output: 'export'`) into `/app/frontend/out`.
2. **Stage 2 (`backend-builder`)**: Go 1.23 Alpine copies `/app/frontend/out` into `cmd/server/frontend_out` (targeted by `//go:embed all:frontend_out/*`) and compiles a static binary (`CGO_ENABLED=0 go build -ldflags="-s -w -extldflags '-static'"`).
3. **Stage 3 (Production Image)**: Alpine 3.20 minimal runtime container (~18MB) with CA certificates, dedicated non-root user, and persistent data mount directory `/var/data`.

---

## 3. Storage Decoupling & Concurrency Rules

1. **Repository Abstraction**:
   - Always encapsulate storage operations behind the `db.StorageEngine` interface.
   - API route handlers in `backend/internal/api/` must never depend directly on concrete database structs.

2. **Concurrency & Locking Guarantees**:
   - When running embedded file-locked databases (BoltDB / SQLite) over NFS mounts in Cloud Run Gen2, enforce:
     ```yaml
     autoscaling.knative.dev/maxScale: "1"
     ```
   - When migrating to managed cloud databases (Cloud SQL PostgreSQL / Firestore), remove the `maxScale: "1"` constraint to enable horizontal auto-scaling (`maxScale: 10+`).

---

## 4. Google Cloud Run Deployment

Deployments target **Google Artifact Registry** and scale-to-zero Cloud Run instances:

```bash
# Set project and region
export GCP_PROJECT_ID="your-project-id"
export GCP_REGION="us-central1"

# Run automated deployment
./scripts/deploy-cloudrun.sh
```
