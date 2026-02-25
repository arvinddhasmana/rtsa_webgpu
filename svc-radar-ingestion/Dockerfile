# CLASSIFICATION: UNCLASSIFIED
# Multi-stage build for svc-radar-ingestion

# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Copy workspace and module files first for layer caching
COPY gen/go/go.mod gen/go/go.sum gen/go/
COPY pkg/go.mod pkg/go.sum pkg/
COPY svc-radar-ingestion/go.mod svc-radar-ingestion/go.sum svc-radar-ingestion/

# Download dependencies
WORKDIR /build/svc-radar-ingestion
ENV GOWORK=off
RUN go mod download

# Copy source files
WORKDIR /build
COPY gen/go/ gen/go/
COPY pkg/ pkg/
COPY svc-radar-ingestion/ svc-radar-ingestion/

# Install grpc_health_probe for Docker health checks
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.25 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

# Build the binary
WORKDIR /build/svc-radar-ingestion
ENV GOWORK=off
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" \
    -o /app/radar-ingestion \
    ./cmd/radar-ingestion

# Stage 2: Runtime (distroless)
FROM gcr.io/distroless/static-debian12:nonroot

# Metadata
LABEL org.opencontainers.image.title="svc-radar-ingestion"
LABEL org.opencontainers.image.description="RTSA Radar Sensor Ingestion Service"
LABEL org.opencontainers.image.vendor="RTSA"
LABEL classification="UNCLASSIFIED"

# Copy binary and health probe
COPY --from=builder /app/radar-ingestion /radar-ingestion
COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe

# Run as non-root (distroless nonroot image)
USER nonroot:nonroot

# gRPC port
EXPOSE 50051

ENTRYPOINT ["/radar-ingestion"]
