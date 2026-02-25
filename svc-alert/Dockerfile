# CLASSIFICATION: UNCLASSIFIED

# ─── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /workspace

# Copy go workspace files
COPY go.work go.work.sum* ./

# Copy module manifests (gen/go and svc-alert)
COPY gen/go/go.mod gen/go/go.sum ./gen/go/
COPY svc-alert/go.mod svc-alert/go.sum* ./svc-alert/

# Download dependencies (cached layer)
WORKDIR /workspace/svc-alert
RUN go mod download

# Copy source code
WORKDIR /workspace
COPY gen/go/ ./gen/go/
COPY svc-alert/ ./svc-alert/

# Build the binary
WORKDIR /workspace/svc-alert
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /out/svc-alert ./cmd/alert/

# ─── Runtime stage ────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/svc-alert /svc-alert

EXPOSE 50051 8081 9090

USER nonroot:nonroot

ENTRYPOINT ["/svc-alert"]
