# CLASSIFICATION: UNCLASSIFIED
FROM golang:1.24-alpine AS builder
WORKDIR /workspace
COPY gen/go/go.mod gen/go/go.sum* ./gen/go/
COPY pkg/go.mod pkg/go.sum* ./pkg/
COPY svc-elint-ingestion/go.mod svc-elint-ingestion/go.sum* ./svc-elint-ingestion/
WORKDIR /workspace/svc-elint-ingestion
ENV GOWORK=off
RUN go mod download
WORKDIR /workspace
COPY gen/go/ ./gen/go/
COPY pkg/ ./pkg/
COPY svc-elint-ingestion/ ./svc-elint-ingestion/
WORKDIR /workspace/svc-elint-ingestion
ENV GOWORK=off
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /elint-ingestion ./cmd/elint-ingestion
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.25 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /elint-ingestion /elint-ingestion
COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe
EXPOSE 50053
ENTRYPOINT ["/elint-ingestion"]
