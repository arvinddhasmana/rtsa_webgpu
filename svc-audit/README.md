<!-- CLASSIFICATION: UNCLASSIFIED -->

# svc-audit — Immutable Audit Trail Service

> **Classification**: UNCLASSIFIED
> **Module**: 13 — Audit Service

## Overview

`svc-audit` is the immutable, append-only audit trail backbone. It consumes all audit events from the `audit.events` Redpanda topic and writes them to the `audit_log` ClickHouse table.

The audit log has **NO TTL** — retention is indefinite per ITSG-33 AU-11.

## Immutability Contract

- Only `INSERT` operations in all SQL
- No `UPDATE`, `DELETE`, `ALTER`, or `TRUNCATE` anywhere
- Data is retained indefinitely — never purged programmatically

## Processing Flow

```
Redpanda (audit.events) → Deserialize proto → Validate → Buffer → Batch INSERT → ClickHouse (audit_log)
```

## Configuration

| Environment Variable | Default | Description |
|---|---|---|
| `RTSA_CLICKHOUSE_DSN` | `clickhouse://default:@localhost:9000/rtsa` | ClickHouse connection string |
| `RTSA_REDPANDA_BROKERS` | `localhost:19092` | Redpanda brokers |
| `RTSA_AUDIT_CONSUMER_GROUP` | `svc-audit` | Consumer group ID |
| `RTSA_AUDIT_BATCH_SIZE` | `500` | Batch insert size |
| `RTSA_AUDIT_FLUSH_PERIOD_SEC` | `2` | Flush period in seconds |
| `RTSA_HEALTH_PORT` | `8083` | Health check HTTP port |

## Running

```bash
RTSA_CLICKHOUSE_DSN=clickhouse://default:@localhost:9000/rtsa \
RTSA_REDPANDA_BROKERS=localhost:19092 \
go run ./cmd/audit/
```

## Testing

```bash
go test -race ./...
```
