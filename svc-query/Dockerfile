# CLASSIFICATION: UNCLASSIFIED
# Multi-stage build: builder (golang) -> distroless runtime
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /workspace

# Copy workspace files
COPY go.work go.work.sum* ./
COPY gen/go/ gen/go/
COPY pkg/ pkg/
COPY svc-query/ svc-query/

# Download dependencies
RUN cd svc-query && go mod download

# Build the binary (CGO disabled for distroless compatibility)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /workspace/bin/svc-query \
    ./svc-query/cmd/query/...

# Runtime stage — distroless for minimal attack surface
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /workspace/bin/svc-query /svc-query

# Run as nonroot user (UID 65532)
USER nonroot:nonroot

EXPOSE 50072

ENTRYPOINT ["/svc-query"]
