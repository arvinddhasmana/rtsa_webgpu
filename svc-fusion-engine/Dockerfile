# CLASSIFICATION: UNCLASSIFIED
# Multi-stage build for svc-fusion-engine
# Stage 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /workspace

# Copy workspace and module files first (layer caching)
COPY go.work go.work.sum ./
COPY gen/go/go.mod gen/go/go.sum ./gen/go/
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY svc-fusion-engine/go.mod svc-fusion-engine/go.sum ./svc-fusion-engine/

# Download dependencies
WORKDIR /workspace/svc-fusion-engine
RUN go mod download

# Copy source
WORKDIR /workspace
COPY gen/go/ ./gen/go/
COPY pkg/ ./pkg/
COPY svc-fusion-engine/ ./svc-fusion-engine/

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /fusion-engine \
    ./svc-fusion-engine/cmd/fusion-engine

# Stage 2: Runtime (distroless for minimal attack surface)
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /fusion-engine /fusion-engine

EXPOSE 8081

ENTRYPOINT ["/fusion-engine"]
