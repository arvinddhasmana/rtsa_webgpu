<!-- CLASSIFICATION: UNCLASSIFIED -->

# Module 13 — Audit Service

> **Module**: 13-audit-service
> **Phase**: P2 (Processing)
> **Dependencies**: Module 02 (protos), Module 03 (shared libs), Module 01 (ClickHouse infra)
> **Agent**: `@greatest-ever-developer`
> **Estimated Effort**: 2 days

---

## 1. Objective

Implement the Audit Service (`svc-audit`) — the immutable, append-only audit trail backbone. It consumes all audit events from the `audit.events` topic and provides a query interface. Every state-changing operation across the entire RTSA system emits audit events to this topic. The audit log has **NO TTL** — retention is indefinite per ITSG-33 AU-11.

**Acceptance Criteria**:

- Consumes from `audit.events` topic
- Writes to `audit_log` ClickHouse table (append-only, no updates, no deletes)
- `GetAuditEntry` — retrieve a single audit event by ID
- `StreamAuditLog` — server-streaming with filters (service, event type, actor, time range)
- Classification filtering on all responses
- **No mutation operations** — this service is read-only after ingestion
- Immutability enforced: no UPDATE or DELETE SQL statements allowed anywhere
- ≥80% line coverage

---

## 2. Service Structure

```
svc-audit/
├── cmd/
│   └── audit/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── consumer/
│   │   ├── audit_consumer.go       # Consumes audit.events
│   │   └── audit_consumer_test.go
│   ├── handler/
│   │   ├── get_entry.go            # GetAuditEntry handler
│   │   ├── stream.go               # StreamAuditLog handler
│   │   └── handler_test.go
│   ├── repository/
│   │   ├── clickhouse.go           # ClickHouse client wrapper
│   │   ├── audit_repo.go           # Audit query builder
│   │   └── repo_test.go
│   └── mapper/
│       └── audit_mapper.go         # Proto ↔ ClickHouse row mapping
├── go.mod
├── Dockerfile
└── README.md
```

---

## 3. Configuration

```go
// CLASSIFICATION: UNCLASSIFIED
package config

type Config struct {
    GRPCPort           int    `env:"RTSA_AUDIT_GRPC_PORT" default:"50073"`
    ClickHouseDSN      string `env:"RTSA_CLICKHOUSE_DSN" required:"true"`
    ClickHouseDatabase string `env:"RTSA_CLICKHOUSE_DATABASE" default:"rtsa"`
    RedpandaBrokers    string `env:"RTSA_REDPANDA_BROKERS" required:"true"`
    ConsumerGroup      string `env:"RTSA_AUDIT_CONSUMER_GROUP" default:"svc-audit"`
    ServiceName        string `env:"RTSA_SERVICE_NAME" default:"svc-audit"`
    MaxQueryRangeDays  int    `env:"RTSA_AUDIT_MAX_RANGE_DAYS" default:"90"`
    MaxResultRows      int    `env:"RTSA_AUDIT_MAX_ROWS" default:"10000"`
    QueryTimeoutSec    int    `env:"RTSA_AUDIT_QUERY_TIMEOUT_SEC" default:"30"`
    DefaultPageSize    int    `env:"RTSA_AUDIT_DEFAULT_PAGE_SIZE" default:"100"`
    OTelEndpoint       string `env:"RTSA_OTEL_ENDPOINT" default:"otel-collector:4317"`
    TLSCertFile        string `env:"RTSA_TLS_CERT_FILE" required:"true"`
    TLSKeyFile         string `env:"RTSA_TLS_KEY_FILE" required:"true"`
    TLSCAFile          string `env:"RTSA_TLS_CA_FILE" required:"true"`
}
```

---

## 4. Audit Consumer

```go
// CLASSIFICATION: UNCLASSIFIED
package consumer

// AuditConsumer consumes from the audit.events topic and writes
// each record to the audit_log ClickHouse table.
//
// Processing guarantees:
//   - At-least-once delivery (Redpanda consumer group)
//   - Deduplication via audit_id (INSERT with ON DUPLICATE skip)
//   - Batch inserts for efficiency (batch of 500, flush every 2s)
//
// IMMUTABILITY RULE: Only INSERT operations. Never UPDATE, DELETE, or ALTER.
type AuditConsumer struct {
    consumer    *redpanda.Consumer
    repo        *repository.AuditRepository
    batchSize   int           // default: 500
    flushPeriod time.Duration // default: 2s
}

// Start begins consuming audit events.
// Flow per message:
//   1. Deserialize AuditEvent protobuf
//   2. Validate required fields (audit_id, service_id, event_type, event_time)
//   3. Add to batch buffer
//   4. When batch is full OR flush timer fires → batch INSERT into ClickHouse
//   5. Commit offsets after successful insert
//
// On deserialization failure: log error, skip message (do NOT DLQ audit events)
// On ClickHouse failure: retry with exponential backoff (1s, 2s, 4s, 8s, max 30s)
func (c *AuditConsumer) Start(ctx context.Context) error { /* implementation */ }
```

---

## 5. Audit Repository

```go
// CLASSIFICATION: UNCLASSIFIED
package repository

// AuditRepository provides read and write access to the audit_log table.
//
// IMMUTABILITY CONTRACT:
//   - The only write operation is BatchInsert (INSERT INTO)
//   - No UPDATE, DELETE, ALTER, or TRUNCATE operations exist in this package
//   - This contract MUST be verified in code review
type AuditRepository struct {
    client *ClickHouseClient
}

// BatchInsert inserts a batch of audit events.
// Uses ClickHouse batch API for efficiency.
// Duplicate audit_ids are silently ignored (idempotent).
func (r *AuditRepository) BatchInsert(
    ctx context.Context,
    events []*auditv1.AuditEvent,
) error { /* implementation */ }

// GetEntry retrieves a single audit event by audit_id.
// Classification filter is injected.
func (r *AuditRepository) GetEntry(
    ctx context.Context,
    auditID string,
    callerClearance commonv1.ClassificationLevel,
) (*auditv1.AuditEvent, error) { /* implementation */ }

// QueryAuditLog performs a filtered, paginated query.
// All queries are parameterized.
// Classification filter is always injected.
//
// Filter criteria:
//   - time_range (required)
//   - service_ids (optional, []string)
//   - event_types (optional, []string)
//   - actor_ids (optional, []string)
//   - actor_type (optional, ActorType)
//   - resource_types (optional, []string)
//   - actions (optional, []string)
//
// Sort: event_time DESC (most recent first)
func (r *AuditRepository) QueryAuditLog(
    ctx context.Context,
    req *auditv1.QueryAuditLogRequest,
    callerClearance commonv1.ClassificationLevel,
    pageToken *domain.PaginationToken,
    pageSize int,
) ([]*auditv1.AuditEvent, *domain.PaginationToken, error) { /* implementation */ }
```

---

## 6. gRPC Handlers

```go
// CLASSIFICATION: UNCLASSIFIED
package handler

// GetAuditEntry returns a single audit event by ID.
// Flow:
//   1. Extract caller clearance from gRPC context
//   2. Query ClickHouse with classification filter
//   3. Return event or NOT_FOUND
func (h *Handler) GetAuditEntry(
    ctx context.Context,
    req *auditv1.GetAuditEntryRequest,
) (*auditv1.GetAuditEntryResponse, error) { /* implementation */ }

// StreamAuditLog streams filtered audit events.
// Flow:
//   1. Extract caller clearance from gRPC context
//   2. Validate time_range (required, max 90 days)
//   3. Execute paginated query, stream each page
//   4. Send each event as a stream message
//   5. Continue until all matching events sent
//
// NOTE: This queries ClickHouse (historical), not real-time Redpanda.
// For real-time audit monitoring, consume audit.events topic directly.
func (h *Handler) StreamAuditLog(
    req *auditv1.StreamAuditLogRequest,
    stream auditv1.AuditService_StreamAuditLogServer,
) error { /* implementation */ }
```

---

## 7. Immutability Enforcement

The audit trail is the compliance backbone. The following rules MUST be enforced:

1. **No UPDATE statements** in any audit service code
2. **No DELETE statements** in any audit service code
3. **No ALTER TABLE** in any audit service code
4. **No TRUNCATE** in any audit service code
5. **ClickHouse table has NO TTL** — data is retained indefinitely
6. **Audit events are never modified** after creation — only appended
7. **Static analysis check**: `grep -rn "UPDATE\|DELETE\|ALTER\|TRUNCATE" svc-audit/` must return zero matches (excluding comments and test helper teardown)

---

## 8. Metrics

| Metric                                             | Type      | Labels                     |
| -------------------------------------------------- | --------- | -------------------------- |
| `rtsa_audit_service_events_consumed_total`         | Counter   | `event_type`, `service_id` |
| `rtsa_audit_service_batch_insert_duration_seconds` | Histogram | -                          |
| `rtsa_audit_service_batch_size`                    | Histogram | -                          |
| `rtsa_audit_service_queries_total`                 | Counter   | `query_type`, `status`     |
| `rtsa_audit_service_consumer_lag`                  | Gauge     | `partition`                |

---

## 9. Test Scenarios

| #   | Test                                      | Expected                            |
| --- | ----------------------------------------- | ----------------------------------- |
| T01 | BatchInsert 100 events                    | All 100 visible in ClickHouse       |
| T02 | BatchInsert duplicate audit_id            | Idempotent — no error, no duplicate |
| T03 | GetEntry existing ID                      | Returns full event                  |
| T04 | GetEntry non-existent ID                  | NOT_FOUND                           |
| T05 | GetEntry classification filter            | Higher classified excluded          |
| T06 | StreamAuditLog by service_id              | Only events from that service       |
| T07 | StreamAuditLog by event_type              | Only matching event types           |
| T08 | StreamAuditLog by actor_id                | Only events by that actor           |
| T09 | StreamAuditLog time range exceeds 90 days | INVALID_ARGUMENT                    |
| T10 | StreamAuditLog pagination                 | Correct cursor traversal            |
| T11 | Consumer processes malformed message      | Logged and skipped, no crash        |
| T12 | Consumer ClickHouse unavailable           | Retry with backoff                  |

---

## 10. Agent Invocation

```
@greatest-ever-developer Implement Module 13 from docs/implementation/13-audit-service.md

Context:
- Read docs/implementation/00-implementation-overview.md for global conventions
- Read docs/implementation/02-protobuf-schemas.md for AuditEvent proto
- Read docs/architecture/data_architecture.md §5.1 for audit_log DDL
- Read docs/architecture/data_architecture.md §6 for retention policy

Deliverables:
1. Complete svc-audit/ with all files
2. AuditConsumer with batch inserts
3. Immutable repository (INSERT only, no UPDATE/DELETE)
4. GetAuditEntry and StreamAuditLog handlers
5. Classification filtering on all queries
6. Unit tests (≥80% coverage)
7. Integration tests with testcontainers (ClickHouse + Redpanda)

CRITICAL:
- audit_log has NO TTL — data retained indefinitely per ITSG-33 AU-11
- IMMUTABILITY: Only INSERT. No UPDATE, DELETE, ALTER, TRUNCATE anywhere
- All queries MUST be parameterized
- Deduplication by audit_id
```
