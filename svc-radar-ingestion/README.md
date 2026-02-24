<!-- CLASSIFICATION: UNCLASSIFIED -->

# svc-radar-ingestion

Radar Sensor Ingestion Service for the RTSA (Real-Time Situational Awareness & Risk Assessment) platform.

## Overview

This service is the **reference implementation** for all RTSA sensor ingestion services. It receives radar track reports via gRPC, validates, normalizes, enriches, and produces them to Redpanda.

## Architecture

```
gRPC Client → IngestionService
                  ↓
            RadarValidator     ← Rejects to DLQ if invalid
                  ↓
            RadarNormalizer    ← Trims/normalizes fields
                  ↓
            Enricher           ← Adds observation_id, metadata, checks classification
                  ↓
       ObservationProducer     → sensors.radar.tracks (Redpanda)
                  ↓
         AuditEmitter          → audit.events (Redpanda)
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `RTSA_SERVICE_NAME` | (required) | Service name |
| `RTSA_GRPC_PORT` | `50051` | gRPC listen port |
| `RTSA_REDPANDA_BROKERS` | `localhost:9092` | Comma-separated broker addresses |
| `RTSA_RADAR_OUTPUT_TOPIC` | `sensors.radar.tracks` | Output Redpanda topic |
| `RTSA_RADAR_DLQ_TOPIC` | `dlq.sensors.radar` | DLQ Redpanda topic |
| `RTSA_MAX_CLASSIFICATION` | `UNCLASSIFIED` | Maximum data classification level |
| `RTSA_RADAR_MAX_FUTURE_OFFSET` | `300` | Max future timestamp offset (seconds) |
| `RTSA_RADAR_MAX_PAST_OFFSET` | `86400` | Max past timestamp offset (seconds) |

## Running

```bash
RTSA_SERVICE_NAME=svc-radar-ingestion \
RTSA_REDPANDA_BROKERS=localhost:9092 \
RTSA_REDPANDA_TLS_ENABLED=false \
go run ./cmd/radar-ingestion
```

## Testing

```bash
# Unit tests
go test ./...

# Integration tests (requires Docker)
RTSA_INTEGRATION_TESTS=true go test -tags integration ./...
```
