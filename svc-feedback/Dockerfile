# CLASSIFICATION: UNCLASSIFIED

# ──────────────────────────────────────────────────────────────────────────────
# Stage 1: Build
# Multi-stage build keeps the final image small and removes build toolchain.
# ──────────────────────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

# Install security updates.
RUN apk add --no-cache ca-certificates tzdata git

WORKDIR /workspace

# Copy workspace and module files first for layer caching.
COPY gen/go/go.mod gen/go/go.sum ./gen/go/
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY svc-feedback/go.mod svc-feedback/go.sum ./svc-feedback/

# Download dependencies (cached layer).
WORKDIR /workspace/svc-feedback
ENV GOWORK=off
RUN go mod download

# Copy source.
WORKDIR /workspace
COPY gen/go/ ./gen/go/
COPY pkg/ ./pkg/
COPY svc-feedback/ ./svc-feedback/

# Build the binary with security flags.
# CGO_ENABLED=0 for static binary.
# -trimpath removes local path information.
# -ldflags "-s -w" strips debug symbols for smaller binary.
WORKDIR /workspace/svc-feedback
ENV GOWORK=off
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /bin/svc-feedback \
    ./cmd/feedback

# ──────────────────────────────────────────────────────────────────────────────
# Stage 2: Runtime
# Use distroless for minimal attack surface (no shell, no package manager).
# ──────────────────────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

# Copy trusted CA certificates for TLS.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled binary.
COPY --from=builder /bin/svc-feedback /svc-feedback

# Non-root user (distroless nonroot uses uid 65532).
USER nonroot:nonroot

# Expose gRPC and health ports.
EXPOSE 50051 8081

ENTRYPOINT ["/svc-feedback"]
