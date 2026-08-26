# mercury-dasha (Mercury Stack)
*Part of the **echosh-labs** ecosystem*

`mercury-dasha` is a unified, single-binary application engine running on **Google Cloud Run (gen2)**. It merges:
- **Next.js Frontend**: Static export (`output: 'export'`) compiled directly into the Go binary.
- **Go Engine Backend**: Ultra-low-overhead HTTP multiplexer with REST/WebSocket capabilities.
- **bbolt Embedded KV Store**: Sub-millisecond ACID JSON document store with zero external database dependencies.
- **Persistent Volume Mount**: Cloud Filestore (NFS) via Serverless VPC Access for POSIX file locking (`flock`/`mmap`) with single-writer guarantees (`maxScale: 1`).

---

## 🏛 Architecture

```
+-------------------------------------------------------------+
|                      Google Cloud Run                       |
|                                                             |
|  +---------------------+        +------------------------+  |
|  |   Next.js Frontend  |        |    Go Engine Backend   |  |
|  |   (Static Export)   | <----> |  (REST / WebSocket API)|  |
|  +---------------------+        +------------------------+  |
|                                              |              |
|                                              v              |
|                                    +---------------------+  |
|                                    |  bbolt Storage DB   |  |
|                                    |  (/var/data/app.db) |  |
|                                    +---------------------+  |
+----------------------------------------------|--------------+
                                               v
                             +-----------------------------------+
                             |     Persistent NFS Filestore      |
                             |  • POSIX file locking (flock)     |
                             |  • Serverless VPC Access          |
                             +-----------------------------------+
```

---

## 🚀 Quick Start (Local Development)

### 1. Verify & Test
```bash
make verify
```
Runs Go backend unit & API integration tests alongside Next.js production build checks.

### 2. Run Backend Engine
```bash
make dev-backend
```
The Go engine will initialize a local test database at `.data/mercury-dasha-dev.db` and start serving on `http://localhost:8080`.

### 3. Run Next.js Frontend
```bash
make dev-frontend
```
Next.js will start on `http://localhost:3000` and proxy API calls to port 8080.

### 4. Run with Docker Compose (Full Stack Single-Binary)
```bash
make docker-run
```
Visits `http://localhost:8080` with volume persistence across container restarts.

---

## 📡 API Reference

| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `/healthz` | `GET` | Service telemetry, Go runtime allocations, and storage stats |
| `/api/telemetry` | `GET` | Live uptime, memory consumption (MB), GC cycles, and goroutine count |
| `/api/dasha/overview` | `GET` | Planetary mahadasha cycles, frequencies, and correspondences |
| `/api/dasha/nakshatras` | `GET` | Mercury-ruled nakshatras (Ashlesha, Jyeshtha, Revati) |
| `/api/dasha/alchemy` | `GET` | Hermetic axioms and quicksilver correspondences |
| `/api/v1/meta` | `GET` | List document keys with optional `?prefix=` search |
| `/api/v1/meta/{key}` | `GET` | Retrieve raw JSON document |
| `/api/v1/meta/{key}` | `POST/PUT` | Validate and atomically commit JSON document |
| `/api/v1/meta/{key}` | `DELETE` | Remove document from `dasha_meta` bucket |
| `/api/v1/backup` | `GET` | Download live binary snapshot of the Bolt database |

---

## ☁️ Cloud Run Deployment & Artifact Registry

Container images are published to **Google Artifact Registry** (`us-central1-docker.pkg.dev/${PROJECT_ID}/mercury-dasha/engine:latest`).

### Step 1: Provision IAM Service Account
```bash
export GCP_PROJECT_ID="your-project-id"
export GCP_REGION="us-central1"

./deploy/setup-iam.sh
```
This creates `mercury-dasha-sa@${GCP_PROJECT_ID}.iam.gserviceaccount.com` with least-privilege roles (`logging.logWriter`, `monitoring.metricWriter`, `vpcaccess.user`).

### Step 2: Provision Filestore & VPC Connector (Interim)
```bash
./deploy/setup-storage.sh
```
Note the output Filestore IP and update `deploy/service.yaml`.

### Step 3: Deploy Service
```bash
make deploy-cloudrun
```

---

## 🗺️ Roadmap & Parked Architectural Decisions (TODO)

### 1. The Tale of Two Frontends: Unified Dossier vs. Dedicated Status App
There are two distinct Next.js frontend codebases in the `echosh-labs` ecosystem:
- **Frontend A (`echosh-labs.com`)**: The multi-route public compendium & dossier (`/`, `/compendium`, `/foundations`, `/axis-mundi`, `/martial-arts`, `/echosh`, `/archive`, `/services`, `/treasury`).
- **Frontend B (`mercury-dasha/frontend`)**: The focused, single-page developer dashboard for planetary telemetry, Mahadasha cycles, and the BoltDB `MetaEditor`.

#### Current Deployment Behavior:
- When built via the ecosystem root deployment pipeline, `echosh-labs.com` static exports are compiled into `backend/cmd/server/frontend_out/` and embedded into the Go single-binary, serving the rich multi-route Dossier on Cloud Run.
- When built via the standalone multi-stage `Dockerfile` inside `mercury-dasha`, it compiles and embeds the minimal telemetry dashboard from `mercury-dasha/frontend/`.

#### Parked Decisions (To Be Selected):
- [ ] **Option 1 (Unified Production Dossier)**: Keep `mercury-dasha` as the single-binary engine serving the full `echosh-labs.com` multi-route dossier, with the Go backend powering `/api/telemetry`, `/api/dasha/*`, and `/api/v1/meta/*`.
- [ ] **Option 2 (Standalone Admin Engine)**: Deploy `mercury-dasha` exclusively with the minimal status dashboard and `MetaEditor` as an internal operational service.
- [ ] **Option 3 (Route Embedding)**: Embed the minimal status dashboard and `MetaEditor` directly as a route (e.g. `/services` or `/status`) inside `echosh-labs.com`, reconciling both into a single unified frontend repository.

---

### 2. Dynamic "Thin Frontend Shell" & Managed Cloud Database Storehouse
- [ ] **Pure Presentation Shell**: Decouple the frontend from embedded/static domain data. All astrological cycles, alchemical constants, and user/agent metadata will be served dynamically via the Go API with client-side caching (SWR / TanStack Query).
- [ ] **Database Storehouse Migration**: Migrate from single-writer BoltDB (`maxScale: 1`) to a managed cloud database:
  - **Option A**: Cloud SQL (PostgreSQL) using connection pooling and `golang-migrate`.
  - **Option B**: Cloud Firestore / Supabase for serverless document persistence.
- [ ] **Cloud Run Auto-Scaling**: Lift the single-instance constraint to allow seamless horizontal scaling (`maxScale: 10+`) and multi-region deployment.
- [ ] **Storage Bucket Integration**: Utilize Cloud Storage (GCS) buckets for binary backups and file uploads via signed URLs.
