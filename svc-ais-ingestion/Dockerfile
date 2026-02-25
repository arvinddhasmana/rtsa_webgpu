# CLASSIFICATION: UNCLASSIFIED
FROM golang:1.24-alpine AS builder
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
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /ais-ingestion ./cmd/ais-ingestion

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /ais-ingestion /ais-ingestion
EXPOSE 50055
ENTRYPOINT ["/ais-ingestion"]
