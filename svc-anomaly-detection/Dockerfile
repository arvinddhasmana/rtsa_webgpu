# CLASSIFICATION: UNCLASSIFIED
# Multi-stage Dockerfile for svc-anomaly-detection.
# Build stage uses golang:1.24-alpine; runtime uses distroless nonroot.

# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# Copy go.work and go.work.sum for workspace-aware builds.

# Copy module files first for layer caching.
COPY gen/go/go.mod gen/go/go.sum ./gen/go/
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY svc-anomaly-detection/go.mod svc-anomaly-detection/go.sum ./svc-anomaly-detection/

# Download dependencies.
WORKDIR /workspace/svc-anomaly-detection
ENV GOWORK=off
RUN go mod download

# Copy all source.
WORKDIR /workspace
COPY gen/go/ ./gen/go/
COPY pkg/ ./pkg/
COPY svc-anomaly-detection/ ./svc-anomaly-detection/

# Build the binary.
WORKDIR /workspace/svc-anomaly-detection
ENV GOWORK=off
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /bin/anomaly-detection \
    ./cmd/anomaly-detection

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /bin/anomaly-detection /anomaly-detection

# Health check port.
EXPOSE 8081

# Run as nonroot (UID 65532 in distroless nonroot image).
USER nonroot:nonroot

ENTRYPOINT ["/anomaly-detection"]
