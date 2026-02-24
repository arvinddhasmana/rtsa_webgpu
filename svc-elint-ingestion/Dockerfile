# CLASSIFICATION: UNCLASSIFIED
FROM golang:1.22-alpine AS builder
WORKDIR /workspace
COPY go.mod go.sum* ./
COPY . .
RUN go build -o /elint-ingestion ./cmd/elint-ingestion

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /elint-ingestion /elint-ingestion
EXPOSE 50053
ENTRYPOINT ["/elint-ingestion"]
