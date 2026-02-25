# CLASSIFICATION: UNCLASSIFIED
FROM golang:1.24-alpine AS builder
WORKDIR /workspace
COPY gen/go/go.mod gen/go/go.sum* ./gen/go/
COPY pkg/go.mod pkg/go.sum* ./pkg/
COPY svc-ew-ingestion/go.mod svc-ew-ingestion/go.sum* ./svc-ew-ingestion/
WORKDIR /workspace/svc-ew-ingestion
ENV GOWORK=off
RUN go mod download
WORKDIR /workspace
COPY gen/go/ ./gen/go/
COPY pkg/ ./pkg/
COPY svc-ew-ingestion/ ./svc-ew-ingestion/
WORKDIR /workspace/svc-ew-ingestion
ENV GOWORK=off
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /ew-ingestion ./cmd/ew-ingestion

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /ew-ingestion /ew-ingestion
EXPOSE 50052
ENTRYPOINT ["/ew-ingestion"]
