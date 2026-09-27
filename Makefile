# ==============================================================================
# echosh-labs / mercury-dasha
# Sovereign Local & Cloud Run Automation
# ==============================================================================

.PHONY: help dev build test verify clean deploy-cloudrun run docker-build docker-run sync clear-port free-port verify-port fix-portproxy

PORT ?= 8080
BOLT_DB_PATH ?= .data/mercury-dasha-dev.db
PROJECT_ID ?= $(shell gcloud config get-value project 2>/dev/null || echo "your-gcp-project")
REGION ?= us-central1
IMAGE_TAG ?= $(REGION)-docker.pkg.dev/$(PROJECT_ID)/mercury-dasha/engine:latest

help:
	@echo "echosh-labs / mercury-dasha Sovereign Commands:"
	@echo "  make sync         - Harmonize bicameral brain manifest and codebase substrates"
	@echo "  make clear-port   - Identify and clear contentious processes on port $(PORT)"
	@echo "  make free-port    - Alias for clear-port"
	@echo "  make verify-port  - Verify cross-substrate port accessibility on port $(PORT)"
	@echo "  make fix-portproxy- Align Windows netsh portproxy with WSL2 networking mode"
	@echo "  make dev          - Run single-binary engine locally in WSL on port $(PORT)"
	@echo "  make build        - Compile Next.js export, sync embed, and build Linux binary"
	@echo "  make test         - Run Go backend tests and frontend test suites"
	@echo "  make lint         - Run Go vet and frontend TypeScript linters"
	@echo "  make verify       - Run full unified test suite (./test.sh)"
	@echo "  make docker-build - Build multi-stage single-binary Docker container"
	@echo "  make clean        - Remove build outputs while preserving .data/"

sync:
	./scripts/bicameral-sync.sh

test:
	@echo "🧪 Running Go backend tests..."
	cd backend && go test -v ./...
	@echo "🧪 Running frontend test suite..."
	cd frontend && npm test

lint:
	@echo "🔍 Running Go vet..."
	cd backend && go vet ./...
	@echo "🔍 Running frontend typecheck..."
	cd frontend && npm run lint

build-frontend:
	cd frontend && npm run build
	rm -rf backend/cmd/server/frontend_out
	cp -r frontend/out backend/cmd/server/frontend_out

build-backend:
	mkdir -p bin
	cd backend && CGO_ENABLED=0 go build -o ../bin/mercury-dasha ./cmd/server/main.go

build: build-frontend build-backend
	@echo "✔ Sovereign Linux binary ready at: bin/mercury-dasha"

verify:
	./test.sh

clear-port:
	@./scripts/clear-port.sh $(PORT)

free-port: clear-port

verify-port:
	@./scripts/verify-port-forwarding.sh $(PORT)

fix-portproxy:
	@./scripts/remedy-portproxy.sh $(PORT)

dev: clear-port build
	mkdir -p .data
	PORT=$(PORT) BOLT_DB_PATH=$(BOLT_DB_PATH) ./bin/mercury-dasha

run: clear-port
	PORT=$(PORT) BOLT_DB_PATH=$(BOLT_DB_PATH) ./bin/mercury-dasha

docker-build:
	docker build -t $(IMAGE_TAG) .

deploy-cloudrun: docker-build
	docker push $(IMAGE_TAG)
	gcloud run services replace deploy/service.yaml --region=$(REGION) --project=$(PROJECT_ID)

clean:
	rm -rf bin/ frontend/out frontend/.next backend/cmd/server/frontend_out/*
	touch backend/cmd/server/frontend_out/.gitkeep
