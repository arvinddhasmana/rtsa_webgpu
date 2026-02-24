# CLASSIFICATION: UNCLASSIFIED
FROM golang:1.22-alpine AS builder
WORKDIR /workspace
COPY go.mod go.sum* ./
COPY . .
RUN go build -o /ais-ingestion ./cmd/ais-ingestion

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /ais-ingestion /ais-ingestion
EXPOSE 50055
ENTRYPOINT ["/ais-ingestion"]
