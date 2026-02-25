<!-- CLASSIFICATION: UNCLASSIFIED -->

# svc-training

Training Pipeline Noop Service — reads validated operator feedback and produces noop model candidates.

## Overview

This service implements the feedback→training pipeline as a noop consumer for v1. It:
1. Consumes from `feedback.operator.validated`
2. Logs each feedback message
3. Produces a noop JSON payload to `models.anomaly.candidates`

No real ML training occurs in this stub.

## Requirements Traceability

| Requirement | Feature | Use Case |
|---|---|---|
| CR-FB-003, CR-FB-004 | FEAT-12 | UC014, UC015 |

## Topics

| Direction | Topic | Format |
|---|---|---|
| Consume | `feedback.operator.validated` | Protobuf FeedbackEvent |
| Produce | `models.anomaly.candidates` | JSON `{"model_id":"noop","status":"stub","timestamp":"..."}` |

## Configuration

| Variable | Default | Description |
|---|---|---|
| `RTSA_HEALTH_PORT` | `8081` | Health check HTTP port |
| `INPUT_TOPIC` | `feedback.operator.validated` | Input Redpanda topic |
| `OUTPUT_TOPIC` | `models.anomaly.candidates` | Output Redpanda topic |
| `RTSA_REDPANDA_BROKERS` | `localhost:19092` | Redpanda broker addresses |
| `CONSUMER_GROUP` | `svc-training` | Consumer group ID |
| `RTSA_SERVICE_NAME` | `svc-training` | Service identifier |
| `RTSA_LOG_LEVEL` | `info` | Log level |
| `RTSA_LOG_FORMAT` | `json` | Log format |

## Running locally

```bash
cd svc-training
go run ./cmd/server
```

## Testing

```bash
go test ./...
```
