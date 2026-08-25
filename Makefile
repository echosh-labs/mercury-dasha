# ==============================================================================
# echosh-labs / mercury-dasha
# Development & Deployment Automation
# ==============================================================================

.PHONY: help dev-backend dev-frontend build-frontend build-backend docker-build docker-run test clean deploy-cloudrun

PROJECT_ID ?= $(shell gcloud config get-value project 2>/dev/null || echo "your-gcp-project")
REGION ?= us-central1
IMAGE_TAG ?= $(REGION)-docker.pkg.dev/$(PROJECT_ID)/mercury-dasha/engine:latest

help:
	@echo "echosh-labs / mercury-dasha Commands:"
	@echo "  make dev-backend     - Run Go engine locally with temp boltdb"
	@echo "  make dev-frontend    - Run Next.js frontend dev server"
	@echo "  make test            - Run all Go unit & integration tests"
	@echo "  make verify          - Run full backend tests & frontend checks"
	@echo "  make docker-build    - Build multi-stage single-binary Docker container"
	@echo "  make docker-run      - Run container locally with volume persistence"
	@echo "  make deploy-cloudrun - Build, push, and deploy service to Cloud Run"
	@echo "  make clean           - Remove build artifacts & temp databases"

test:
	cd backend && go test -v ./...

verify: test
	cd frontend && npm run build

build-frontend:
	cd frontend && npm install && npm run build
	mkdir -p backend/cmd/server/frontend_out
	cp -r frontend/out/* backend/cmd/server/frontend_out/

build-backend:
	cd backend && CGO_ENABLED=0 go build -o ../bin/mercury-dasha ./cmd/server/main.go

dev-backend:
	mkdir -p .data
	cd backend && BOLT_DB_PATH=../.data/mercury-dasha-dev.db PORT=8080 ENV=development go run ./cmd/server/main.go

dev-frontend:
	cd frontend && npm run dev

docker-build:
	docker build -t $(IMAGE_TAG) .

docker-run:
	docker-compose up --build

deploy-cloudrun: docker-build
	docker push $(IMAGE_TAG)
	gcloud run services replace deploy/service.yaml --region=$(REGION) --project=$(PROJECT_ID)

clean:
	rm -rf bin/ frontend/out frontend/.next .data backend/cmd/server/frontend_out/*
	touch backend/cmd/server/frontend_out/.gitkeep
