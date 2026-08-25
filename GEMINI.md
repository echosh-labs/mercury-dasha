# Mercury Dasha Workspace Rules & Context
# echosh-labs / mercury-dasha

## Project Summary
`mercury-dasha` is a single-binary full-stack web service running on Google Cloud Run (gen2). It merges a Next.js 15 frontend with a high-performance Go backend API and an ACID key-value document store.

## Build & Test Commands
- **Backend Tests**: `cd backend && go test -v ./...`
- **Frontend Build Check**: `cd frontend && npm run build`
- **Full Verification**: `make verify`
- **Local Dev Server**: `make dev-backend` (port 8080) and `make dev-frontend` (port 3000)
- **Container Build & Deploy**: `./scripts/deploy-cloudrun.sh`

## Code Conventions
- Store API handlers in `backend/internal/api/` and decouple database operations via `db.StorageEngine` in `backend/internal/db/store.go`.
- Embed compiled Next.js export bundles via Go `embed.FS` in `backend/cmd/server/main.go`.
- Keep the frontend presentation layer lightweight and retrieve all domain data dynamically through REST endpoints (`/api/dasha/*`, `/api/v1/meta/*`).
