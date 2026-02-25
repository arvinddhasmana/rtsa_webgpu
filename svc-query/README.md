<!-- CLASSIFICATION: UNCLASSIFIED -->

# svc-query — Historical Query Service

> **Classification**: UNCLASSIFIED
> **Module**: 12 — Query Service

## Overview

`svc-query` provides a gRPC interface for historical queries against ClickHouse. It implements the `QueryService` proto with classification-aware filtering, query guardrails, cursor-based pagination, and audit logging.

## gRPC Endpoints

| RPC | Description | Deadline |
|---|---|---|
| `QueryTracks` | Historical fused track data | 30s |
| `QueryAnomalies` | Anomaly detection history | 30s |
| `QueryAuditLog` | Audit log query | 30s |

## Features

- **Classification filtering**: Server-side injection — never trust client clearance
- **Query guardrails**: Max 30-day range, max 100K rows, 30s timeout
- **Cursor-based pagination**: Efficient page traversal via `page_token`
- **Parameterized queries only**: SQL injection prevention

## Configuration

| Environment Variable | Default | Description |
|---|---|---|
| `RTSA_QUERY_GRPC_PORT` | `50072` | gRPC server port |
| `RTSA_CLICKHOUSE_DSN` | `clickhouse://default:@localhost:9000/rtsa` | ClickHouse connection string |
| `RTSA_QUERY_MAX_RANGE_DAYS` | `30` | Maximum query time range |
| `RTSA_QUERY_MAX_ROWS` | `100000` | Maximum result rows |
| `RTSA_QUERY_TIMEOUT_SEC` | `30` | Query timeout in seconds |
| `RTSA_QUERY_DEFAULT_PAGE_SIZE` | `100` | Default page size |
| `RTSA_REDPANDA_BROKERS` | `localhost:19092` | Redpanda brokers |

## Running

```bash
RTSA_CLICKHOUSE_DSN=clickhouse://default:@localhost:9000/rtsa \
RTSA_REDPANDA_BROKERS=localhost:19092 \
go run ./cmd/query/
```

## Testing

```bash
go test -race ./...
```
