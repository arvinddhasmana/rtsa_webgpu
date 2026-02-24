<!-- CLASSIFICATION: UNCLASSIFIED -->

# Module 12 — Query Service & Redpanda Connect ETL

> **Module**: 12-query-service-etl
> **Phase**: P3 (Presentation) + P2 (Processing)
> **Dependencies**: Module 01 (ClickHouse infra), Module 02 (protos), Module 03 (shared libs)
> **Agent**: `@greatest-ever-developer`
> **Estimated Effort**: 5 days

---

## 1. Objective

Implement two coupled subsystems:

1. **Query Service** (`svc-query`): gRPC service for historical queries against ClickHouse, with classification-aware filtering, query guardrails, pagination, and audit logging.
2. **Redpanda Connect ETL Pipelines**: 5 pipeline configurations that materialize Redpanda topic data into ClickHouse tables.

**Acceptance Criteria**:

- `QueryTracks` — parameterized ClickHouse query with classification filter
- `QueryAnomalies` — anomaly history with severity/type filters
- `QueryAuditLog` — audit log query (operator/service scope)
- Classification ceiling injected server-side (never trust client)
- Query guardrails: max 30-day range, max 100K rows, 30s timeout
- Pagination with cursor-based tokens
- Every query generates an audit event
- 5 ETL pipelines: tracks, sensors, alerts, feedback, audit
- ≥80% line coverage

---

## 2. Part A — Query Service

### 2.1 Service Structure

```
svc-query/
├── cmd/
│   └── query/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── domain/
│   │   ├── guardrail.go          # Query guardrail enforcement
│   │   ├── guardrail_test.go
│   │   ├── paginator.go          # Cursor-based pagination
│   │   └── paginator_test.go
│   ├── handler/
│   │   ├── tracks.go             # QueryTracks handler
│   │   ├── anomalies.go          # QueryAnomalies handler
│   │   ├── audit_log.go          # QueryAuditLog handler
│   │   └── handler_test.go
│   ├── repository/
│   │   ├── clickhouse.go         # ClickHouse client wrapper
│   │   ├── tracks_repo.go        # Track query builder
│   │   ├── anomaly_repo.go       # Anomaly query builder
│   │   ├── audit_repo.go         # Audit log query builder
│   │   └── repo_test.go
│   └── security/
│       ├── classification_filter.go  # Server-side classification injection
│       └── classification_filter_test.go
├── go.mod
├── Dockerfile
└── README.md
```

### 2.2 Configuration

```go
// CLASSIFICATION: UNCLASSIFIED
package config

type Config struct {
    GRPCPort           int    `env:"RTSA_QUERY_GRPC_PORT" default:"50072"`
    ClickHouseDSN      string `env:"RTSA_CLICKHOUSE_DSN" required:"true"`
    ClickHouseDatabase string `env:"RTSA_CLICKHOUSE_DATABASE" default:"rtsa"`
    MaxQueryRangeDays  int    `env:"RTSA_QUERY_MAX_RANGE_DAYS" default:"30"`
    MaxResultRows      int    `env:"RTSA_QUERY_MAX_ROWS" default:"100000"`
    QueryTimeoutSec    int    `env:"RTSA_QUERY_TIMEOUT_SEC" default:"30"`
    DefaultPageSize    int    `env:"RTSA_QUERY_DEFAULT_PAGE_SIZE" default:"100"`
    MaxPageSize        int    `env:"RTSA_QUERY_MAX_PAGE_SIZE" default:"1000"`
    RedpandaBrokers    string `env:"RTSA_REDPANDA_BROKERS" required:"true"`
    ServiceName        string `env:"RTSA_SERVICE_NAME" default:"svc-query"`
    OTelEndpoint       string `env:"RTSA_OTEL_ENDPOINT" default:"otel-collector:4317"`
    TLSCertFile        string `env:"RTSA_TLS_CERT_FILE" required:"true"`
    TLSKeyFile         string `env:"RTSA_TLS_KEY_FILE" required:"true"`
    TLSCAFile          string `env:"RTSA_TLS_CA_FILE" required:"true"`
}
```

### 2.3 Classification Filter

```go
// CLASSIFICATION: UNCLASSIFIED
package security

// ClassificationFilter injects server-side classification filtering.
// The caller's clearance level comes from the mTLS certificate (extracted
// by the classification interceptor and placed in gRPC metadata).
//
// NEVER trust client-supplied classification level.
type ClassificationFilter struct{}

// InjectFilter adds a WHERE clause to ensure the caller only sees data
// at or below their clearance level.
//
// It maps ClassificationLevel enum to the ClickHouse Enum8 ordinal:
//   UNCLASSIFIED=1, PROTECTED_A=2, PROTECTED_B=3, PROTECTED_C=4, SECRET=5
//
// Example:
//   Input query:  SELECT * FROM tracks_fused WHERE entity_type = ?
//   Output query: SELECT * FROM tracks_fused WHERE entity_type = ? AND classification_level <= ?
//
// The classification parameter is ALWAYS added as the last parameter.
func (f *ClassificationFilter) InjectFilter(
    query string,
    callerClearance commonv1.ClassificationLevel,
) (string, interface{}) { /* implementation */ }

// ExtractClearance gets the caller's clearance from gRPC context metadata.
// Returns UNCLASSIFIED if not present (deny by default — least privilege).
func ExtractClearance(ctx context.Context) commonv1.ClassificationLevel { /* implementation */ }
```

### 2.4 Query Guardrail

```go
// CLASSIFICATION: UNCLASSIFIED
package domain

// QueryGuardrail enforces safety limits on queries.
type QueryGuardrail struct {
    MaxRangeDays  int // default: 30
    MaxResultRows int // default: 100000
    TimeoutSec    int // default: 30
}

// ValidateTimeRange ensures the query time range does not exceed MaxRangeDays.
// Returns INVALID_ARGUMENT if range is too wide or end < start.
func (g *QueryGuardrail) ValidateTimeRange(start, end *timestamppb.Timestamp) error { /* implementation */ }

// EnforceRowLimit adds LIMIT clause to query.
// If requested limit is 0 or > MaxResultRows, uses MaxResultRows.
func (g *QueryGuardrail) EnforceRowLimit(requestedLimit int) int { /* implementation */ }

// QueryContext returns a context.Context with the configured timeout.
func (g *QueryGuardrail) QueryContext(parent context.Context) (context.Context, context.CancelFunc) { /* implementation */ }
```

### 2.5 Cursor-Based Pagination

```go
// CLASSIFICATION: UNCLASSIFIED
package domain

// PaginationToken encodes cursor state for efficient pagination.
// Encoded as base64(JSON) in page_token field.
type PaginationToken struct {
    LastID        string    `json:"lid"`    // Last row ID seen
    LastTimestamp time.Time `json:"lt"`     // Last row timestamp
    PageSize      int       `json:"ps"`     // Requested page size
}

// EncodePaginationToken serializes the token for response.
func EncodePaginationToken(token *PaginationToken) string { /* base64(JSON) */ }

// DecodePaginationToken deserializes from request.
// Returns nil if empty string (first page).
func DecodePaginationToken(s string) (*PaginationToken, error) { /* implementation */ }

// ApplyPagination modifies query and params for cursor-based pagination.
// Adds: AND (event_time, track_id) > (?, ?) ORDER BY event_time ASC, track_id ASC LIMIT ?
func ApplyPagination(query string, params []interface{}, token *PaginationToken, pageSize int) (string, []interface{}) { /* implementation */ }
```

### 2.6 ClickHouse Repository

```go
// CLASSIFICATION: UNCLASSIFIED
package repository

import "github.com/ClickHouse/clickhouse-go/v2"

// ClickHouseClient wraps the ClickHouse connection.
type ClickHouseClient struct {
    conn clickhouse.Conn
}

// NewClickHouseClient creates a new client with mTLS.
func NewClickHouseClient(cfg *config.Config) (*ClickHouseClient, error) { /* implementation */ }

// TracksRepository handles track queries.
type TracksRepository struct {
    client *ClickHouseClient
    filter *security.ClassificationFilter
    guard  *domain.QueryGuardrail
}

// QueryTracks executes a parameterized track query.
// Filter criteria:
//   - time_range (required)
//   - entity_types (optional, []EntityType)
//   - hostile_classifications (optional, []HostileClassification)
//   - bounding_box (optional, lat/lon bounds)
//   - track_status (optional, []TrackStatus)
//   - min_confidence (optional, float64)
//
// Always injects classification filter.
// Always enforces guardrails.
// Returns paginated results.
func (r *TracksRepository) QueryTracks(
    ctx context.Context,
    req *queryv1.QueryTracksRequest,
) (*queryv1.QueryTracksResponse, error) { /* implementation */ }

// AnomalyRepository handles anomaly queries.
type AnomalyRepository struct {
    client *ClickHouseClient
    filter *security.ClassificationFilter
    guard  *domain.QueryGuardrail
}

// QueryAnomalies executes a parameterized anomaly query.
// Filter criteria:
//   - time_range (required)
//   - anomaly_types (optional)
//   - severity_levels (optional)
//   - track_ids (optional)
//   - min_confidence (optional)
//
// Always injects classification filter.
func (r *AnomalyRepository) QueryAnomalies(
    ctx context.Context,
    req *queryv1.QueryAnomaliesRequest,
) (*queryv1.QueryAnomaliesResponse, error) { /* implementation */ }

// AuditRepository handles audit log queries.
type AuditRepository struct {
    client *ClickHouseClient
    filter *security.ClassificationFilter
    guard  *domain.QueryGuardrail
}

// QueryAuditLog executes a parameterized audit query.
// Filter criteria:
//   - time_range (required)
//   - service_ids (optional)
//   - event_types (optional)
//   - actor_ids (optional)
//   - resource_types (optional)
func (r *AuditRepository) QueryAuditLog(
    ctx context.Context,
    req *queryv1.QueryAuditLogRequest,
) (*queryv1.QueryAuditLogResponse, error) { /* implementation */ }
```

### 2.7 Query Builder Pattern

All queries MUST be parameterized. **No string concatenation of user input into SQL.**

```go
// CLASSIFICATION: UNCLASSIFIED
// Example: TracksRepository.QueryTracks builds a query like:
//
//   SELECT track_id, entity_type, hostile_classification, latitude, longitude,
//          altitude_meters, speed_knots, heading_degrees, confidence_score,
//          source_count, source_sensors, classification_level, track_status, event_time
//   FROM tracks_fused
//   WHERE event_time >= ? AND event_time <= ?
//     AND entity_type IN (?, ?)
//     AND classification_level <= ?
//   ORDER BY event_time ASC, track_id ASC
//   LIMIT ?
//
// Using positional parameters ONLY. Never string interpolation.
```

### 2.8 Handler — QueryTracks

```go
// CLASSIFICATION: UNCLASSIFIED
package handler

// QueryTracksHandler implements QueryService.QueryTracks.
// Flow:
//   1. Extract caller clearance from context (classification interceptor)
//   2. Validate request: time_range required, non-empty
//   3. Apply guardrails (time range, row limit)
//   4. Decode page_token if present
//   5. Build parameterized query with classification filter
//   6. Execute against ClickHouse with timeout context
//   7. Map rows to QueryTracksResponse
//   8. Build next_page_token if more results
//   9. Emit audit event (query type, filters, row count, caller)
//  10. Return response
func (h *Handler) QueryTracks(
    ctx context.Context,
    req *queryv1.QueryTracksRequest,
) (*queryv1.QueryTracksResponse, error) { /* implementation */ }
```

### 2.9 Audit Emission

Every query execution MUST emit an audit event:

```go
// CLASSIFICATION: UNCLASSIFIED
// Audit event for each query
auditEmitter.Emit(ctx, audit.AuditParams{
    EventType:    "query_executed",
    ResourceType: "tracks",        // or "anomalies", "audit_log"
    ResourceID:   "",              // N/A for queries
    Action:       "QUERY",
    Details: map[string]interface{}{
        "query_type":     "QueryTracks",
        "time_range":     timeRange,
        "filters":        filterSummary,
        "result_count":   resultCount,
        "page_size":      pageSize,
        "operator_clearance": clearanceLevel,
    },
})
```

### 2.10 Metrics

| Metric                                          | Type      | Labels                 |
| ----------------------------------------------- | --------- | ---------------------- |
| `rtsa_query_service_queries_total`              | Counter   | `query_type`, `status` |
| `rtsa_query_service_query_duration_seconds`     | Histogram | `query_type`           |
| `rtsa_query_service_rows_returned`              | Histogram | `query_type`           |
| `rtsa_query_service_guardrail_rejections_total` | Counter   | `reason`               |

---

## 3. Part B — Redpanda Connect ETL Pipelines

### 3.1 Overview

5 Redpanda Connect pipelines materialize streaming data into ClickHouse:

| Pipeline | Source Topics                    | Target Table          | Consumer Group                  |
| -------- | -------------------------------- | --------------------- | ------------------------------- |
| Tracks   | `tracks.fused.*` (5 topics)      | `tracks_fused`        | `rpconnect-clickhouse-tracks`   |
| Sensors  | `sensors.*` (7 topics)           | `sensor_observations` | `rpconnect-clickhouse-sensors`  |
| Alerts   | `alerts.anomaly.*` (3 topics)    | `anomaly_detections`  | `rpconnect-clickhouse-alerts`   |
| Feedback | `feedback.operator.*` (2 topics) | `operator_feedback`   | `rpconnect-clickhouse-feedback` |
| Audit    | `audit.events`                   | `audit_log`           | `rpconnect-clickhouse-audit`    |

### 3.2 File Structure

```
deploy/etl/
├── tracks-pipeline.yaml
├── sensors-pipeline.yaml
├── alerts-pipeline.yaml
├── feedback-pipeline.yaml
├── audit-pipeline.yaml
└── README.md
```

### 3.3 Pipeline: Tracks → ClickHouse

```yaml
# CLASSIFICATION: UNCLASSIFIED
# deploy/etl/tracks-pipeline.yaml
# Materializes fused tracks from Redpanda to ClickHouse
input:
  kafka_franz:
    seed_brokers: ["${REDPANDA_BROKERS}"]
    topics:
      - tracks.fused.surface
      - tracks.fused.air
      - tracks.fused.subsurface
      - tracks.fused.land
      - tracks.fused.cyber
    consumer_group: "rpconnect-clickhouse-tracks"
    start_from_oldest: true
    tls:
      enabled: true
      root_cas_file: /certs/ca.crt
      cert_file: /certs/client.crt
      key_file: /certs/client.key

pipeline:
  processors:
    - protobuf:
        operator: from_string
        message: rtsa.entity.v1.FusedTrack
        import_paths: ["/proto"]
    - mapping: |
        root.track_id = this.track_id
        root.entity_type = match this.entity_type {
          1 => "SURFACE", 2 => "AIR", 3 => "SUBSURFACE",
          4 => "LAND", 5 => "CYBER", _ => "UNSPECIFIED"
        }
        root.hostile_classification = match this.hostile_class {
          1 => "HOSTILE", 2 => "FRIENDLY", 3 => "NEUTRAL",
          4 => "UNKNOWN", _ => "UNSPECIFIED"
        }
        root.latitude = this.estimated_position.latitude
        root.longitude = this.estimated_position.longitude
        root.altitude_meters = this.estimated_position.altitude_meters.or(null)
        root.speed_knots = this.estimated_position.speed_knots.or(null)
        root.heading_degrees = this.estimated_position.heading_degrees.or(null)
        root.confidence_score = this.confidence_score
        root.source_count = this.source_count
        root.source_sensors = this.sources.map_each(s -> s.sensor_id)
        root.classification_level = match this.classification {
          1 => "UNCLASSIFIED", 2 => "PROTECTED_A", 3 => "PROTECTED_B",
          4 => "PROTECTED_C", 5 => "SECRET", _ => "UNCLASSIFIED"
        }
        root.track_status = match this.status {
          1 => "ACTIVE", 2 => "STALE", 3 => "DROPPED",
          4 => "MERGED", _ => "ACTIVE"
        }
        root.event_time = this.updated_at

output:
  sql_insert:
    driver: clickhouse
    dsn: "clickhouse://${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}@${CLICKHOUSE_HOST}:9440/rtsa?secure=true"
    table: tracks_fused
    columns:
      - track_id
      - entity_type
      - hostile_classification
      - latitude
      - longitude
      - altitude_meters
      - speed_knots
      - heading_degrees
      - confidence_score
      - source_count
      - source_sensors
      - classification_level
      - track_status
      - event_time
    batching:
      count: 1000
      period: "5s"
    max_in_flight: 4
```

### 3.4 Pipeline: Sensors → ClickHouse

```yaml
# CLASSIFICATION: UNCLASSIFIED
# deploy/etl/sensors-pipeline.yaml
input:
  kafka_franz:
    seed_brokers: ["${REDPANDA_BROKERS}"]
    topics:
      - sensors.radar.tracks
      - sensors.ew.intercepts
      - sensors.elint.detections
      - sensors.isr.observations
      - sensors.ais.positions
      - sensors.cyber.iocs
      - sensors.nato.link16
    consumer_group: "rpconnect-clickhouse-sensors"
    start_from_oldest: true
    tls:
      enabled: true
      root_cas_file: /certs/ca.crt
      cert_file: /certs/client.crt
      key_file: /certs/client.key

pipeline:
  processors:
    - protobuf:
        operator: from_string
        message: rtsa.ingestion.v1.SensorObservation
        import_paths: ["/proto"]
    - mapping: |
        root.observation_id = this.observation_id
        root.sensor_id = this.sensor_id
        root.sensor_type = match this.sensor_type {
          1 => "RADAR", 2 => "EW_SIGINT", 3 => "ELINT_COMINT",
          4 => "ISR", 5 => "AIS_BFT", 6 => "CYBER", 7 => "NATO",
          _ => "UNSPECIFIED"
        }
        root.latitude = this.position.latitude.or(null)
        root.longitude = this.position.longitude.or(null)
        root.altitude_meters = this.position.altitude_meters.or(null)
        root.speed_knots = this.position.speed_knots.or(null)
        root.heading_degrees = this.position.heading_degrees.or(null)
        root.classification_level = match this.classification {
          1 => "UNCLASSIFIED", 2 => "PROTECTED_A", 3 => "PROTECTED_B",
          4 => "PROTECTED_C", 5 => "SECRET", _ => "UNCLASSIFIED"
        }
        root.metadata_json = this.metadata.string()
        root.event_time = this.observation_time

output:
  sql_insert:
    driver: clickhouse
    dsn: "clickhouse://${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}@${CLICKHOUSE_HOST}:9440/rtsa?secure=true"
    table: sensor_observations
    columns:
      - observation_id
      - sensor_id
      - sensor_type
      - latitude
      - longitude
      - altitude_meters
      - speed_knots
      - heading_degrees
      - classification_level
      - metadata_json
      - event_time
    batching:
      count: 1000
      period: "5s"
    max_in_flight: 4
```

### 3.5 Pipeline: Alerts → ClickHouse

```yaml
# CLASSIFICATION: UNCLASSIFIED
# deploy/etl/alerts-pipeline.yaml
input:
  kafka_franz:
    seed_brokers: ["${REDPANDA_BROKERS}"]
    topics:
      - alerts.anomaly.critical
      - alerts.anomaly.elevated
      - alerts.anomaly.watch
    consumer_group: "rpconnect-clickhouse-alerts"
    start_from_oldest: true
    tls:
      enabled: true
      root_cas_file: /certs/ca.crt
      cert_file: /certs/client.crt
      key_file: /certs/client.key

pipeline:
  processors:
    - protobuf:
        operator: from_string
        message: rtsa.inference.v1.AnomalyAlert
        import_paths: ["/proto"]
    - mapping: |
        root.alert_id = this.alert_id
        root.track_id = this.track_id
        root.anomaly_type = match this.anomaly_type {
          1 => "SPEED", 2 => "ROUTE_DEVIATION", 3 => "AIS_MANIPULATION",
          4 => "BEHAVIORAL", 5 => "TEMPORAL", 6 => "PROXIMITY",
          _ => "UNSPECIFIED"
        }
        root.severity = match this.severity {
          1 => "NORMAL", 2 => "WATCH", 3 => "ELEVATED", 4 => "CRITICAL",
          _ => "NORMAL"
        }
        root.confidence_score = this.confidence_score
        root.explanation = this.explanation
        root.model_version = this.model_version
        root.classification_level = match this.classification {
          1 => "UNCLASSIFIED", 2 => "PROTECTED_A", 3 => "PROTECTED_B",
          4 => "PROTECTED_C", 5 => "SECRET", _ => "UNCLASSIFIED"
        }
        root.event_time = this.detected_at

output:
  sql_insert:
    driver: clickhouse
    dsn: "clickhouse://${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}@${CLICKHOUSE_HOST}:9440/rtsa?secure=true"
    table: anomaly_detections
    columns:
      - alert_id
      - track_id
      - anomaly_type
      - severity
      - confidence_score
      - explanation
      - model_version
      - classification_level
      - event_time
    batching:
      count: 500
      period: "5s"
    max_in_flight: 2
```

### 3.6 Pipeline: Feedback → ClickHouse

```yaml
# CLASSIFICATION: UNCLASSIFIED
# deploy/etl/feedback-pipeline.yaml
input:
  kafka_franz:
    seed_brokers: ["${REDPANDA_BROKERS}"]
    topics:
      - feedback.operator.submissions
      - feedback.operator.validated
    consumer_group: "rpconnect-clickhouse-feedback"
    start_from_oldest: true
    tls:
      enabled: true
      root_cas_file: /certs/ca.crt
      cert_file: /certs/client.crt
      key_file: /certs/client.key

pipeline:
  processors:
    - protobuf:
        operator: from_string
        message: rtsa.feedback.v1.OperatorFeedback
        import_paths: ["/proto"]
    - mapping: |
        root.feedback_id = this.feedback_id
        root.track_id = this.track_id
        root.operator_id = this.operator_id
        root.feedback_type = match this.feedback_type {
          1 => "CONFIRM_HOSTILE", 2 => "CONFIRM_FRIENDLY",
          3 => "RECLASSIFY", 4 => "REJECT_ANOMALY",
          5 => "CONFIRM_ANOMALY", _ => "UNSPECIFIED"
        }
        root.justification = this.justification
        root.trust_score = this.trust_score
        root.clearance_score = this.trust_breakdown.clearance_score
        root.accuracy_score = this.trust_breakdown.accuracy_score
        root.temporal_score = this.trust_breakdown.temporal_score
        root.deviation_score = this.trust_breakdown.deviation_score
        # Determine validation status based on source topic
        root.validated = if meta("kafka_topic").has_prefix("feedback.operator.validated") { 1 } else { 0 }
        root.classification_level = match this.classification {
          1 => "UNCLASSIFIED", 2 => "PROTECTED_A", 3 => "PROTECTED_B",
          4 => "PROTECTED_C", 5 => "SECRET", _ => "UNCLASSIFIED"
        }
        root.event_time = this.submitted_at

output:
  sql_insert:
    driver: clickhouse
    dsn: "clickhouse://${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}@${CLICKHOUSE_HOST}:9440/rtsa?secure=true"
    table: operator_feedback
    columns:
      - feedback_id
      - track_id
      - operator_id
      - feedback_type
      - justification
      - trust_score
      - clearance_score
      - accuracy_score
      - temporal_score
      - deviation_score
      - validated
      - classification_level
      - event_time
    batching:
      count: 500
      period: "5s"
    max_in_flight: 2
```

### 3.7 Pipeline: Audit → ClickHouse

```yaml
# CLASSIFICATION: UNCLASSIFIED
# deploy/etl/audit-pipeline.yaml
# CRITICAL: audit_log has NO TTL — retained indefinitely per ITSG-33
input:
  kafka_franz:
    seed_brokers: ["${REDPANDA_BROKERS}"]
    topics:
      - audit.events
    consumer_group: "rpconnect-clickhouse-audit"
    start_from_oldest: true
    tls:
      enabled: true
      root_cas_file: /certs/ca.crt
      cert_file: /certs/client.crt
      key_file: /certs/client.key

pipeline:
  processors:
    - protobuf:
        operator: from_string
        message: rtsa.audit.v1.AuditEvent
        import_paths: ["/proto"]
    - mapping: |
        root.audit_id = this.audit_id
        root.service_id = this.service_id
        root.event_type = this.event_type
        root.actor_id = this.actor_id
        root.actor_type = match this.actor_type {
          1 => "SERVICE", 2 => "OPERATOR", 3 => "SYSTEM", _ => "SERVICE"
        }
        root.resource_type = this.resource_type
        root.resource_id = this.resource_id
        root.action = this.action
        root.detail_json = this.detail_json
        root.classification_level = match this.classification {
          1 => "UNCLASSIFIED", 2 => "PROTECTED_A", 3 => "PROTECTED_B",
          4 => "PROTECTED_C", 5 => "SECRET", _ => "UNCLASSIFIED"
        }
        root.event_time = this.event_time

output:
  sql_insert:
    driver: clickhouse
    dsn: "clickhouse://${CLICKHOUSE_USER}:${CLICKHOUSE_PASSWORD}@${CLICKHOUSE_HOST}:9440/rtsa?secure=true"
    table: audit_log
    columns:
      - audit_id
      - service_id
      - event_type
      - actor_id
      - actor_type
      - resource_type
      - resource_id
      - action
      - detail_json
      - classification_level
      - event_time
    batching:
      count: 500
      period: "5s"
    max_in_flight: 2
```

---

## 4. ClickHouse DDL

The following DDL must be applied during infrastructure setup (Module 01 `init-clickhouse.sh`). Reproduced here as reference for the query service:

### 4.1 Core Tables

```sql
-- tracks_fused
CREATE TABLE IF NOT EXISTS tracks_fused (
    track_id String,
    entity_type Enum8('UNSPECIFIED'=0, 'SURFACE'=1, 'AIR'=2, 'SUBSURFACE'=3, 'LAND'=4, 'CYBER'=5),
    hostile_classification Enum8('UNSPECIFIED'=0, 'HOSTILE'=1, 'FRIENDLY'=2, 'NEUTRAL'=3, 'UNKNOWN'=4),
    latitude Float64,
    longitude Float64,
    altitude_meters Nullable(Float64),
    speed_knots Nullable(Float64),
    heading_degrees Nullable(Float64),
    confidence_score Float64,
    source_count UInt8,
    source_sensors Array(String),
    classification_level Enum8('UNCLASSIFIED'=1, 'PROTECTED_A'=2, 'PROTECTED_B'=3, 'PROTECTED_C'=4, 'SECRET'=5),
    track_status Enum8('ACTIVE'=1, 'STALE'=2, 'DROPPED'=3, 'MERGED'=4),
    event_time DateTime64(3, 'UTC'),
    ingestion_time DateTime64(3, 'UTC') DEFAULT now64(3)
) ENGINE = MergeTree()
  PARTITION BY toYYYYMMDD(event_time)
  ORDER BY (entity_type, track_id, event_time)
  TTL event_time + INTERVAL 90 DAY
  SETTINGS index_granularity = 8192;

-- sensor_observations
CREATE TABLE IF NOT EXISTS sensor_observations (
    observation_id String,
    sensor_id String,
    sensor_type Enum8('UNSPECIFIED'=0, 'RADAR'=1, 'EW_SIGINT'=2, 'ELINT_COMINT'=3, 'ISR'=4, 'AIS_BFT'=5, 'CYBER'=6, 'NATO'=7),
    latitude Nullable(Float64),
    longitude Nullable(Float64),
    altitude_meters Nullable(Float64),
    speed_knots Nullable(Float64),
    heading_degrees Nullable(Float64),
    classification_level Enum8('UNCLASSIFIED'=1, 'PROTECTED_A'=2, 'PROTECTED_B'=3, 'PROTECTED_C'=4, 'SECRET'=5),
    metadata_json String,
    event_time DateTime64(3, 'UTC'),
    ingestion_time DateTime64(3, 'UTC') DEFAULT now64(3)
) ENGINE = MergeTree()
  PARTITION BY (sensor_type, toYYYYMMDD(event_time))
  ORDER BY (sensor_type, sensor_id, event_time)
  TTL event_time + INTERVAL 90 DAY
  SETTINGS index_granularity = 8192;

-- anomaly_detections
CREATE TABLE IF NOT EXISTS anomaly_detections (
    alert_id String,
    track_id String,
    anomaly_type Enum8('UNSPECIFIED'=0, 'SPEED'=1, 'ROUTE_DEVIATION'=2, 'AIS_MANIPULATION'=3, 'BEHAVIORAL'=4, 'TEMPORAL'=5, 'PROXIMITY'=6),
    severity Enum8('NORMAL'=1, 'WATCH'=2, 'ELEVATED'=3, 'CRITICAL'=4),
    confidence_score Float64,
    explanation String,
    model_version String,
    classification_level Enum8('UNCLASSIFIED'=1, 'PROTECTED_A'=2, 'PROTECTED_B'=3, 'PROTECTED_C'=4, 'SECRET'=5),
    event_time DateTime64(3, 'UTC'),
    ingestion_time DateTime64(3, 'UTC') DEFAULT now64(3)
) ENGINE = MergeTree()
  PARTITION BY toYYYYMMDD(event_time)
  ORDER BY (anomaly_type, track_id, event_time)
  TTL event_time + INTERVAL 365 DAY
  SETTINGS index_granularity = 8192;

-- operator_feedback
CREATE TABLE IF NOT EXISTS operator_feedback (
    feedback_id String,
    track_id String,
    operator_id String,
    feedback_type Enum8('UNSPECIFIED'=0, 'CONFIRM_HOSTILE'=1, 'CONFIRM_FRIENDLY'=2, 'RECLASSIFY'=3, 'REJECT_ANOMALY'=4, 'CONFIRM_ANOMALY'=5),
    justification String,
    trust_score Float64,
    clearance_score Float64,
    accuracy_score Float64,
    temporal_score Float64,
    deviation_score Float64,
    validated UInt8,
    classification_level Enum8('UNCLASSIFIED'=1, 'PROTECTED_A'=2, 'PROTECTED_B'=3, 'PROTECTED_C'=4, 'SECRET'=5),
    event_time DateTime64(3, 'UTC'),
    ingestion_time DateTime64(3, 'UTC') DEFAULT now64(3)
) ENGINE = MergeTree()
  PARTITION BY toYYYYMMDD(event_time)
  ORDER BY (operator_id, track_id, event_time)
  TTL event_time + INTERVAL 730 DAY
  SETTINGS index_granularity = 8192;

-- audit_log (NO TTL — indefinite per ITSG-33)
CREATE TABLE IF NOT EXISTS audit_log (
    audit_id String,
    service_id String,
    event_type String,
    actor_id String,
    actor_type Enum8('SERVICE'=1, 'OPERATOR'=2, 'SYSTEM'=3),
    resource_type String,
    resource_id String,
    action String,
    detail_json String,
    classification_level Enum8('UNCLASSIFIED'=1, 'PROTECTED_A'=2, 'PROTECTED_B'=3, 'PROTECTED_C'=4, 'SECRET'=5),
    event_time DateTime64(3, 'UTC'),
    ingestion_time DateTime64(3, 'UTC') DEFAULT now64(3)
) ENGINE = MergeTree()
  PARTITION BY toYYYYMMDD(event_time)
  ORDER BY (service_id, event_type, event_time)
  SETTINGS index_granularity = 8192;
```

### 4.2 Materialized Views

```sql
-- Track count by type (real-time aggregation)
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_track_count_by_type
ENGINE = SummingMergeTree() ORDER BY (entity_type, hour)
AS SELECT
    entity_type,
    toStartOfHour(event_time) AS hour,
    uniqExact(track_id) AS unique_tracks,
    count() AS total_observations
FROM tracks_fused
GROUP BY entity_type, hour;

-- Anomaly summary hourly
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_anomaly_summary_hourly
ENGINE = SummingMergeTree() ORDER BY (anomaly_type, severity, hour)
AS SELECT
    anomaly_type,
    severity,
    toStartOfHour(event_time) AS hour,
    count() AS alert_count,
    avg(confidence_score) AS avg_confidence,
    uniqExact(track_id) AS affected_tracks
FROM anomaly_detections
GROUP BY anomaly_type, severity, hour;
```

---

## 5. Test Scenarios

### 5.1 Query Service Tests

| #   | Test                                                | Expected                   |
| --- | --------------------------------------------------- | -------------------------- |
| T01 | QueryTracks with valid time range                   | Returns matching tracks    |
| T02 | QueryTracks: range exceeds 30 days                  | INVALID_ARGUMENT error     |
| T03 | QueryTracks: end before start                       | INVALID_ARGUMENT error     |
| T04 | QueryTracks: classification filter                  | Only sees data ≤ clearance |
| T05 | QueryTracks: SECRET caller queries PROTECTED_B data | Results returned           |
| T06 | QueryTracks: PROTECTED_B caller queries SECRET data | Excluded from results      |
| T07 | QueryTracks: entity type filter                     | Only matching types        |
| T08 | QueryTracks: bounding box filter                    | Only positions in box      |
| T09 | QueryTracks: pagination first page                  | Returns page + token       |
| T10 | QueryTracks: pagination second page                 | Continues from cursor      |
| T11 | QueryAnomalies: severity filter                     | Only matching severities   |
| T12 | QueryAnomalies: anomaly type filter                 | Only matching types        |
| T13 | QueryAuditLog: service_id filter                    | Only from that service     |
| T14 | Query generates audit event                         | Event emitted to Redpanda  |
| T15 | Row limit enforcement                               | Max 100K rows returned     |
| T16 | Query timeout enforcement                           | Context cancelled at 30s   |

### 5.2 Integration Tests

| #    | Test                               | Expected                 |
| ---- | ---------------------------------- | ------------------------ |
| IT01 | Insert tracks → query via service  | End-to-end ClickHouse    |
| IT02 | Classification injection verified  | SQL contains filter      |
| IT03 | Pagination across 1000+ rows       | Correct cursor traversal |
| IT04 | Concurrent query load (10 clients) | No data races            |

---

## 6. Agent Invocation

```
@greatest-ever-developer Implement Module 12 from docs/implementation/12-query-service-etl.md

Context:
- Read docs/implementation/00-implementation-overview.md for global conventions
- Read docs/implementation/02-protobuf-schemas.md for QueryService proto
- Read docs/architecture/data_architecture.md for ClickHouse DDL and ETL config
- Read docs/architecture/component_design.md §8.3 for query service diagram

Deliverables:
Part A - Query Service (svc-query/):
1. ClickHouse client with mTLS
2. Classification filter (server-side injection)
3. Query guardrails (30-day range, 100K rows, 30s timeout)
4. Cursor-based pagination
5. QueryTracks, QueryAnomalies, QueryAuditLog handlers
6. Audit event emission for every query
7. Unit tests (≥80% coverage)
8. Integration tests with testcontainers (ClickHouse + Redpanda)

Part B - ETL Pipelines (deploy/etl/):
1. tracks-pipeline.yaml (5 topics → tracks_fused)
2. sensors-pipeline.yaml (7 topics → sensor_observations)
3. alerts-pipeline.yaml (3 topics → anomaly_detections)
4. feedback-pipeline.yaml (2 topics → operator_feedback)
5. audit-pipeline.yaml (1 topic → audit_log)

CRITICAL:
- All ClickHouse queries MUST be parameterized (no string concatenation)
- Classification filter MUST be injected server-side
- Audit log table has NO TTL (indefinite retention per ITSG-33)
- Use clickhouse-go/v2 driver
```
