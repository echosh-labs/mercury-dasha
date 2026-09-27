# ==============================================================================
# echosh-labs / mercury-dasha
# Multi-stage single-binary build: Next.js static export + Go Engine + bbolt
# ==============================================================================

# ------------------------------------------------------------------------------
# Stage 1: Build Next.js Static Export Bundle
# ------------------------------------------------------------------------------
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

# ------------------------------------------------------------------------------
# Stage 2: Compile High-Performance Static Go Binary
# ------------------------------------------------------------------------------
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app/backend

# Install build dependencies
RUN apk add --no-cache git ca-certificates

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
# Inject compiled Next.js export bundle directly into Go embed target
COPY --from=frontend-builder /app/frontend/out ./cmd/server/frontend_out

# Compile static binary with optimizations (-s -w strips debug symbols)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -extldflags '-static'" \
    -o /app/mercury-dasha ./cmd/server/main.go

# ------------------------------------------------------------------------------
# Stage 3: Ultra-Minimal Production Container (~18MB)
# ------------------------------------------------------------------------------
FROM alpine:3.20

# Add CA certificates for outbound TLS and tzdata
RUN apk --no-cache add ca-certificates tzdata

# Create dedicated non-root user and persistent volume directory
RUN addgroup -S dasha && adduser -S dasha -G dasha \
    && mkdir -p /var/data \
    && chown -R dasha:dasha /var/data

WORKDIR /app
COPY --from=backend-builder /app/mercury-dasha /app/mercury-dasha

# Set runtime defaults for Cloud Run
ENV PORT=8080 \
    BOLT_DB_PATH=/var/data/mercury-dasha.db \
    SERVICE_NAME=mercury-dasha \
    ENV=production

USER dasha
EXPOSE 8080

ENTRYPOINT ["/app/mercury-dasha"]
