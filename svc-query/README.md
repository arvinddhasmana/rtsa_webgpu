<!-- CLASSIFICATION: UNCLASSIFIED -->

# svc-query — RTSA Historical Query Service

## Overview

`svc-query` provides historical data access over gRPC for the RTSA platform.
It queries three ClickHouse tables:

| RPC               | Table               | Description                        |
|-------------------|---------------------|------------------------------------|
| `QueryTracks`     | `tracks_fused`      | Fused track history with filters   |
| `QueryAnomalies`  | `anomaly_detections`| Anomaly alert history              |
| `QueryAuditLog`   | `audit_log`         | Immutable audit trail              |

## Security Model

- **Classification filter**: injected server-side from gRPC metadata; client-supplied levels are NEVER trusted
- **mTLS**: all gRPC channels require mutual TLS authentication
- **Guardrails**: max 30-day query range, max 100K rows per query, 30s timeout
- **Audit trail**: every query emits an audit event to `audit.events` topic
- **Parameterized queries**: all ClickHouse queries use positional `?` parameters — no string interpolation

## Configuration

All configuration is via environment variables (no hardcoded secrets):

| Variable                    | Default               | Description                        |
|-----------------------------|-----------------------|------------------------------------|
| `RTSA_CLICKHOUSE_DSN`       | *(required)*          | ClickHouse connection string       |
| `RTSA_REDPANDA_BROKERS`     | *(required)*          | Redpanda broker addresses          |
| `RTSA_TLS_SERVER_CERT`      | *(required)*          | gRPC server TLS certificate        |
| `RTSA_TLS_SERVER_KEY`       | *(required)*          | gRPC server TLS private key        |
| `RTSA_TLS_CA_CERT`          | *(required)*          | CA certificate for mTLS            |
| `RTSA_QUERY_GRPC_PORT`      | `50072`               | gRPC server port                   |
| `RTSA_CLICKHOUSE_DATABASE`  | `rtsa`                | ClickHouse database name           |
| `RTSA_QUERY_MAX_RANGE_DAYS` | `30`                  | Max query time range (days)        |
| `RTSA_QUERY_MAX_ROWS`       | `100000`              | Max rows per query                 |
| `RTSA_QUERY_TIMEOUT_SEC`    | `30`                  | Query execution timeout            |
| `RTSA_QUERY_DEFAULT_PAGE_SIZE` | `100`              | Default pagination page size       |
| `RTSA_QUERY_MAX_PAGE_SIZE`  | `1000`                | Maximum pagination page size       |
| `RTSA_CLICKHOUSE_TLS_CERT`  | *(optional)*          | ClickHouse client TLS cert (mTLS)  |
| `RTSA_CLICKHOUSE_TLS_KEY`   | *(optional)*          | ClickHouse client TLS key          |
| `RTSA_CLICKHOUSE_TLS_CA`    | *(optional)*          | ClickHouse CA cert                 |

## Running

```bash
export RTSA_CLICKHOUSE_DSN="clickhouse://default:dev@clickhouse:9000/rtsa"
export RTSA_REDPANDA_BROKERS="redpanda:19092"
export RTSA_TLS_SERVER_CERT="./certs/dev/server.crt"
export RTSA_TLS_SERVER_KEY="./certs/dev/server.key"
export RTSA_TLS_CA_CERT="./certs/dev/ca.crt"
./svc-query
```

## Testing

```bash
# Unit tests
go test ./... -race -count=1

# Integration tests (requires ClickHouse on localhost:9000)
go test -tags integration ./tests/integration/... -v
```

## Architecture

```
gRPC client
    │
    ▼ mTLS
Handler (orchestration)
    ├── ExtractClearance (from metadata, never from request)
    ├── ValidateTimeRange (guardrail)
    ├── Repository.Query (parameterized SQL + classification filter)
    ├── AuditEmitter.Emit (to audit.events topic)
    └── Return response
```
