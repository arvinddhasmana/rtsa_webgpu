# CLASSIFICATION: UNCLASSIFIED
# Multi-stage build for svc-track
# Runtime: gcr.io/distroless/static-debian12:nonroot (no shell, minimal attack surface)
# Feature: FEAT-13 Situational Awareness UI

# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder
ARG TARGETARCH

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /workspace

# Copy go workspace files

# Copy module files for dependency resolution
COPY gen/go/go.mod gen/go/go.sum ./gen/go/
COPY svc-track/go.mod svc-track/go.sum ./svc-track/

# Download dependencies (cached layer)
WORKDIR /workspace/svc-track
ENV GOWORK=off
RUN go mod download

# Copy source files
WORKDIR /workspace
COPY gen/go/ ./gen/go/
COPY svc-track/ ./svc-track/

RUN GRPC_HEALTH_PROBE_VERSION=v0.4.25 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-${TARGETARCH} && \
    chmod +x /bin/grpc_health_probe

# Build the binary (static, no CGO)
WORKDIR /workspace/svc-track
ENV GOWORK=off
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build \
    -ldflags="-w -s -extldflags=-static" \
    -trimpath \
    -o /out/svc-track \
    ./cmd/track

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe

# Copy CA certificates for outbound TLS connections
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /out/svc-track /svc-track

# Expose ports
EXPOSE 50051 8081 9090

# Run as non-root (uid=65532 in distroless nonroot)
USER nonroot:nonroot

ENTRYPOINT ["/svc-track"]
