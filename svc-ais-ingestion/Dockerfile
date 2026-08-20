# CLASSIFICATION: UNCLASSIFIED
FROM golang:1.24-alpine AS builder
ARG TARGETARCH
WORKDIR /workspace
COPY gen/go/go.mod gen/go/go.sum* ./gen/go/
COPY pkg/go.mod pkg/go.sum* ./pkg/
COPY svc-ais-ingestion/go.mod svc-ais-ingestion/go.sum* ./svc-ais-ingestion/
WORKDIR /workspace/svc-ais-ingestion
ENV GOWORK=off
RUN go mod download
WORKDIR /workspace
COPY gen/go/ ./gen/go/
COPY pkg/ ./pkg/
COPY svc-ais-ingestion/ ./svc-ais-ingestion/
WORKDIR /workspace/svc-ais-ingestion
ENV GOWORK=off
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o /ais-ingestion ./cmd/ais-ingestion
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.25 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-${TARGETARCH} && \
    chmod +x /bin/grpc_health_probe

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /ais-ingestion /ais-ingestion
COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe
EXPOSE 50055
ENTRYPOINT ["/ais-ingestion"]
