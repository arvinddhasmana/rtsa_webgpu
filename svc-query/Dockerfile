# CLASSIFICATION: UNCLASSIFIED

# ─── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /workspace

# Copy go workspace files
COPY go.work go.work.sum* ./

# Copy module manifests
COPY gen/go/go.mod gen/go/go.sum ./gen/go/
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY svc-query/go.mod svc-query/go.sum* ./svc-query/

# Download dependencies (cached layer)
WORKDIR /workspace/svc-query
RUN go mod download

# Copy source code
WORKDIR /workspace
COPY gen/go/ ./gen/go/
COPY pkg/ ./pkg/
COPY svc-query/ ./svc-query/

# Build the binary
WORKDIR /workspace/svc-query
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /out/svc-query ./cmd/query/

# ─── Runtime stage ────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/svc-query /svc-query

EXPOSE 50072 8082 9092

USER nonroot:nonroot

ENTRYPOINT ["/svc-query"]
