# CLASSIFICATION: UNCLASSIFIED

# ─── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# Copy go workspace files

# Copy module manifests
COPY gen/go/go.mod gen/go/go.sum ./gen/go/
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY svc-audit/go.mod svc-audit/go.sum* ./svc-audit/

# Download dependencies (cached layer)
WORKDIR /workspace/svc-audit
ENV GOWORK=off
RUN go mod download

# Copy source code
WORKDIR /workspace
COPY gen/go/ ./gen/go/
COPY pkg/ ./pkg/
COPY svc-audit/ ./svc-audit/

# Build the binary
WORKDIR /workspace/svc-audit
ENV GOWORK=off
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /out/svc-audit ./cmd/audit/

# ─── Runtime stage ────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/svc-audit /svc-audit

EXPOSE 8083 9093 50073

USER nonroot:nonroot

ENTRYPOINT ["/svc-audit"]
