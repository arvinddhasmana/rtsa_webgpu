# CLASSIFICATION: UNCLASSIFIED
# Multi-stage Dockerfile for the RTSA test data simulator.
# Stage 1: Build the simulator binary.
FROM golang:1.22-alpine AS builder

WORKDIR /workspace

# Copy go workspace files for dependency resolution.
COPY go.work go.work
COPY gen/go/ gen/go/
COPY tools/simulator/ tools/simulator/

WORKDIR /workspace/tools/simulator

RUN go build -o /simulator ./cmd/simulator/

# Stage 2: Minimal runtime image.
FROM alpine:3.19

LABEL org.opencontainers.image.title="rtsa-simulator" \
      org.opencontainers.image.description="RTSA Test Data Simulator — UNCLASSIFIED" \
      org.opencontainers.image.vendor="RTSA Project"

RUN addgroup -S simulator && adduser -S simulator -G simulator

WORKDIR /app

COPY --from=builder /simulator /app/simulator
COPY tools/simulator/scenarios/ /app/scenarios/

USER simulator

ENTRYPOINT ["/app/simulator"]
CMD ["--scenario", "/app/scenarios/default.yaml"]
