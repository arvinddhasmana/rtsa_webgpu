<!-- CLASSIFICATION: UNCLASSIFIED -->

# svc-alert — Alert Service

## Overview

`svc-alert` is a gRPC service that consumes anomaly alerts from Redpanda topics,
maintains a thread-safe in-memory priority queue ordered by severity (CRITICAL first),
and streams alerts in real-time to authorised COP Web App clients.

---

## Architecture

```
Redpanda (alerts.anomaly.critical / elevated / watch)
        │
        ▼
  AlertConsumer (protobuf deserialisation)
        │
        ▼
  AlertQueue (priority sorted: CRITICAL > ELEVATED > WATCH > NORMAL, then newer first)
        │
        ├──► StreamAlerts (gRPC server-streaming, classification-filtered)
        ├──► AcknowledgeAlert (gRPC unary)
        └──► GetAlertDetails (gRPC unary)
```

---

## gRPC API

| RPC               | Type              | Description                                   |
|-------------------|-------------------|-----------------------------------------------|
| `StreamAlerts`    | Server-streaming  | Real-time alert stream with severity/type/classification filters |
| `AcknowledgeAlert`| Unary             | Mark an alert as acknowledged by an operator  |
| `GetAlertDetails` | Unary             | Retrieve a full alert (including features)    |

### Classification Filtering (MANDATORY)

All RPCs enforce classification filtering. Alerts with a `classification` level
exceeding the `clearance_level` in the request are **never** returned.

---

## Configuration

All configuration is via environment variables (prefix: `RTSA_`):

| Variable                       | Default                              | Description                     |
|--------------------------------|--------------------------------------|---------------------------------|
| `RTSA_GRPC_PORT`               | `50051`                              | gRPC server port                |
| `RTSA_HEALTH_PORT`             | `8081`                               | Health check HTTP port          |
| `RTSA_METRICS_PORT`            | `9090`                               | Prometheus metrics port         |
| `RTSA_REDPANDA_BROKERS`        | `localhost:19092`                    | Redpanda broker(s), CSV         |
| `RTSA_ALERT_CONSUMER_GROUP`    | `svc-alert`                          | Kafka consumer group ID         |
| `RTSA_ALERT_TOPICS`            | `alerts.anomaly.critical,alerts.anomaly.elevated,alerts.anomaly.watch` | Topics to consume |
| `RTSA_ALERT_MAX_QUEUE_SIZE`    | `10000`                              | Max alerts in memory            |
| `RTSA_TLS_CA_CERT`             | `./certs/dev/ca.crt`                 | TLS CA certificate path         |
| `RTSA_TLS_SERVER_CERT`         | `./certs/dev/server.crt`             | TLS server certificate path     |
| `RTSA_TLS_SERVER_KEY`          | `./certs/dev/server.key`             | TLS server key path             |
| `RTSA_LOG_LEVEL`               | `info`                               | Log level (debug/info/warn/error)|
| `RTSA_SERVICE_NAME`            | `svc-alert`                          | Service name for telemetry      |

---

## Prometheus Metrics

| Metric                                           | Type      | Labels                     |
|--------------------------------------------------|-----------|----------------------------|
| `rtsa_alert_service_alerts_received_total`       | Counter   | `severity`, `anomaly_type` |
| `rtsa_alert_service_alerts_unacknowledged`       | Gauge     | `severity`                 |
| `rtsa_alert_service_stream_clients`              | Gauge     | —                          |
| `rtsa_alert_service_time_to_acknowledge_seconds` | Histogram | `severity`                 |
| `rtsa_alert_service_queue_size`                  | Gauge     | —                          |

---

## Priority Queue

Alerts are ordered by:
1. **Severity** (CRITICAL=3 > ELEVATED=2 > WATCH=1 > NORMAL=0)
2. **Detected-at timestamp** (newer first within same severity)

When the queue reaches `MaxQueueSize`, the lowest-priority alert is evicted.

---

## Running Locally

```bash
# Start dependencies
docker compose up -d redpanda

# Run the service
RTSA_REDPANDA_BROKERS=localhost:19092 go run ./cmd/alert/
```

---

## Testing

```bash
cd svc-alert
go test ./... -race -count=1 -coverprofile=coverage.out
go tool cover -func=coverage.out | tail -5
```

---

## Health Check

```
GET http://localhost:8081/healthz
→ 200 OK
```

---

## Dependencies

| Dependency                          | Purpose                    |
|-------------------------------------|----------------------------|
| `google.golang.org/grpc`            | gRPC server implementation |
| `google.golang.org/protobuf`        | Protobuf serialisation     |
| `github.com/prometheus/client_golang`| Metrics exposition        |
| `github.com/twmb/franz-go`          | Redpanda consumer client   |
| `github.com/arvinddhasmana/RTSA_VS_Opus/gen/go` | Generated proto types |
