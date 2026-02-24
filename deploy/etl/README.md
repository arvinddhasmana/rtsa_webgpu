<!-- CLASSIFICATION: UNCLASSIFIED -->

# deploy/etl — Redpanda Connect ETL Pipelines

## Overview

Five Redpanda Connect pipelines that materialise streaming event data from Redpanda
topics into ClickHouse tables for historical analysis.

| Pipeline File             | Source Topics (count)         | Target Table          | TTL       |
|---------------------------|-------------------------------|-----------------------|-----------|
| `tracks-pipeline.yaml`    | `tracks.fused.*` (5)          | `tracks_fused`        | 90 days   |
| `sensors-pipeline.yaml`   | `sensors.*` (7)               | `sensor_observations` | 90 days   |
| `alerts-pipeline.yaml`    | `alerts.anomaly.*` (3)        | `anomaly_detections`  | 365 days  |
| `feedback-pipeline.yaml`  | `feedback.operator.*` (2)     | `operator_feedback`   | 730 days  |
| `audit-pipeline.yaml`     | `audit.events` (1)            | `audit_log`           | **NONE**  |

## CRITICAL: Audit Log Retention

The `audit_log` table has **NO TTL**. Audit records are retained indefinitely per
ITSG-33 Section A.12.4 accountability requirements. **Never** add TTL or configure
automatic deletion for audit records.

## Environment Variables

Each pipeline uses these environment variables (no hardcoded credentials):

| Variable                  | Description                           |
|---------------------------|---------------------------------------|
| `REDPANDA_BROKERS`        | Comma-separated broker addresses      |
| `CLICKHOUSE_HOST`         | ClickHouse hostname                   |
| `CLICKHOUSE_PORT`         | ClickHouse port (default: 9000)       |
| `CLICKHOUSE_USER`         | ClickHouse username                   |
| `CLICKHOUSE_PASSWORD`     | ClickHouse password (from secrets)    |
| `CLICKHOUSE_DATABASE`     | Database name (default: rtsa)         |
| `PROMETHEUS_PUSHGATEWAY`  | Prometheus pushgateway address        |

## Running

```bash
# Start all pipelines with Docker Compose
docker compose -f deploy/docker-compose.yml up redpanda-connect-tracks
docker compose -f deploy/docker-compose.yml up redpanda-connect-sensors
# ... etc

# Or run a single pipeline directly:
redpanda-connect run deploy/etl/tracks-pipeline.yaml
```

## Security Notes

- All topics are consumed with TLS (cert files mounted at `/certs/`)
- Classification level is mapped from protobuf enum to ClickHouse Enum8 string
- No raw sensor payloads are stored in `metadata_json` — only structured metadata
- Audit pipeline never logs message payloads (may contain classified data)
