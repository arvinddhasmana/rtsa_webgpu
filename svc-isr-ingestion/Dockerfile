# CLASSIFICATION: UNCLASSIFIED
FROM golang:1.22-alpine AS builder
WORKDIR /workspace
COPY go.mod go.sum* ./
COPY . .
RUN go build -o /isr-ingestion ./cmd/isr-ingestion

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /isr-ingestion /isr-ingestion
EXPOSE 50054
ENTRYPOINT ["/isr-ingestion"]
