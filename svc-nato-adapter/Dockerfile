# CLASSIFICATION: UNCLASSIFIED
# Multi-stage Dockerfile for svc-nato-adapter.
# Build stage uses golang:1.22-alpine; runtime uses distroless nonroot.

# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /workspace

# Copy go.work and go.work.sum for workspace-aware builds.
COPY go.work go.work.sum ./

# Copy module files first for layer caching.
COPY gen/go/go.mod gen/go/go.sum ./gen/go/
COPY svc-nato-adapter/go.mod svc-nato-adapter/go.sum ./svc-nato-adapter/

# Download dependencies.
RUN go work sync && go mod download

# Copy all source.
COPY gen/go/ ./gen/go/
COPY svc-nato-adapter/ ./svc-nato-adapter/

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /bin/svc-nato-adapter \
    ./svc-nato-adapter/cmd/server

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /bin/svc-nato-adapter /svc-nato-adapter

# Health check port and gRPC port.
EXPOSE 50051 8081

# Run as nonroot (UID 65532 in distroless nonroot image).
USER nonroot:nonroot

ENTRYPOINT ["/svc-nato-adapter"]
