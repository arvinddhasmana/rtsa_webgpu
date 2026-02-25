# CLASSIFICATION: UNCLASSIFIED
FROM golang:1.24-alpine AS builder
WORKDIR /workspace
COPY gen/go/go.mod gen/go/go.sum* ./gen/go/
COPY pkg/go.mod pkg/go.sum* ./pkg/
COPY svc-isr-ingestion/go.mod svc-isr-ingestion/go.sum* ./svc-isr-ingestion/
WORKDIR /workspace/svc-isr-ingestion
ENV GOWORK=off
RUN go mod download
WORKDIR /workspace
COPY gen/go/ ./gen/go/
COPY pkg/ ./pkg/
COPY svc-isr-ingestion/ ./svc-isr-ingestion/
WORKDIR /workspace/svc-isr-ingestion
ENV GOWORK=off
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /isr-ingestion ./cmd/isr-ingestion

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /isr-ingestion /isr-ingestion
EXPOSE 50054
ENTRYPOINT ["/isr-ingestion"]
