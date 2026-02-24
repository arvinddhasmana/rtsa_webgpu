# CLASSIFICATION: UNCLASSIFIED
FROM golang:1.22-alpine AS builder
WORKDIR /workspace
COPY go.mod go.sum* ./
COPY . .
RUN go build -o /cyber-ingestion ./cmd/cyber-ingestion

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /cyber-ingestion /cyber-ingestion
EXPOSE 50056
ENTRYPOINT ["/cyber-ingestion"]
