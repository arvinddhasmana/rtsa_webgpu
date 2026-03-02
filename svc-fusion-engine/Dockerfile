# CLASSIFICATION: UNCLASSIFIED
# Multi-stage build for svc-fusion-engine
# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# Copy workspace and module files first (layer caching)
COPY gen/go/go.mod gen/go/go.sum ./gen/go/
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY svc-fusion-engine/go.mod svc-fusion-engine/go.sum ./svc-fusion-engine/

# Download dependencies
WORKDIR /workspace/svc-fusion-engine
ENV GOWORK=off
RUN go mod download

# Copy source
WORKDIR /workspace
COPY gen/go/ ./gen/go/
COPY pkg/ ./pkg/
COPY svc-fusion-engine/ ./svc-fusion-engine/

RUN GRPC_HEALTH_PROBE_VERSION=v0.4.25 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

# Build the binary
WORKDIR /workspace/svc-fusion-engine
ENV GOWORK=off
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /fusion-engine \
    ./cmd/fusion-engine

# Stage 2: Runtime (distroless for minimal attack surface)
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe

COPY --from=builder /fusion-engine /fusion-engine

EXPOSE 8081

ENTRYPOINT ["/fusion-engine"]
