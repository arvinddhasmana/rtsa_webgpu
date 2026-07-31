# CLASSIFICATION: UNCLASSIFIED
# Multi-stage build for svc-webtransport (RTSA hot-path fan-out service).
# Runtime: gcr.io/distroless/static-debian12:nonroot (no shell, minimal surface).
# Feature: FEAT-13 Situational Awareness UI (hot path)

# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /workspace

# Copy module files first for dependency-download layer caching.
COPY gen/go/go.mod gen/go/go.sum ./gen/go/
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY svc-webtransport/go.mod svc-webtransport/go.sum ./svc-webtransport/

WORKDIR /workspace/svc-webtransport
ENV GOWORK=off
RUN go mod download

# Copy source.
WORKDIR /workspace
COPY gen/go/ ./gen/go/
COPY pkg/ ./pkg/
COPY svc-webtransport/ ./svc-webtransport/

# Build the static binary (no CGO).
WORKDIR /workspace/svc-webtransport
ENV GOWORK=off
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-w -s -extldflags=-static -X main.version=${VERSION:-container}" \
    -trimpath \
    -o /out/svc-webtransport \
    ./cmd/webtransport

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

LABEL org.opencontainers.image.title="svc-webtransport"
LABEL org.opencontainers.image.classification="UNCLASSIFIED"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /out/svc-webtransport /svc-webtransport

# 4443/udp: WebTransport (QUIC)  •  8081: health  •  9090: metrics
EXPOSE 4443/udp 8081 9090

USER nonroot:nonroot

ENTRYPOINT ["/svc-webtransport"]
