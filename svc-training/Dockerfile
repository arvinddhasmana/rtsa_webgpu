# CLASSIFICATION: UNCLASSIFIED
# Multi-stage Dockerfile for svc-training.
# Build stage uses golang:1.24-alpine; runtime uses distroless nonroot.

# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# Copy go.work and go.work.sum for workspace-aware builds.

# Copy module files first for layer caching.
COPY gen/go/go.mod gen/go/go.sum ./gen/go/
COPY svc-training/go.mod svc-training/go.sum ./svc-training/

# Download dependencies.
WORKDIR /workspace/svc-training
ENV GOWORK=off
RUN go mod download

# Copy all source.
WORKDIR /workspace
COPY gen/go/ ./gen/go/
COPY svc-training/ ./svc-training/

# Build the binary.
WORKDIR /workspace/svc-training
ENV GOWORK=off
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /bin/svc-training \
    ./cmd/server

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /bin/svc-training /svc-training

# Health check port.
EXPOSE 8081

# Run as nonroot (UID 65532 in distroless nonroot image).
USER nonroot:nonroot

ENTRYPOINT ["/svc-training"]
