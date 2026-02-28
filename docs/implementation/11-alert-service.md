<!-- CLASSIFICATION: UNCLASSIFIED -->

# Module 11 — Alert Service

> **Module**: 11-alert-service
> **Phase**: P3 (Presentation)
> **Dependencies**: Module 02 (protos), Module 03 (shared libraries), Module 08 (anomaly detection produces alerts)
> **Agent**: `@greatest-ever-developer`
> **Estimated Effort**: 3 days

---

## 1. Objective

Implement the Alert Service (`svc-alert`) that consumes anomaly alerts from `alerts.anomaly.*` topics, maintains a priority queue (CRITICAL first, then by time), and exposes gRPC server-streaming to COP Web App clients with severity and type filtering.

**v2.0 Enhancement**: Add the `AssignAlert` unary RPC that allows an operator to assign an alert to another operator for follow-up investigation. This produces an audit event and sets an `assigned_to` field on the in-memory alert.

**Acceptance Criteria**:

- Consumes from `alerts.anomaly.critical`, `alerts.anomaly.elevated`, `alerts.anomaly.watch`
- In-memory priority queue sorted by severity (CRITICAL > ELEVATED > WATCH) then timestamp
- `StreamAlerts` — server-streaming with severity filter
- `AcknowledgeAlert` — marks alert as acknowledged
- `GetAlertDetails` — returns full alert with features
- Classification filtering on all responses
- Metrics: alert rate by severity, unacknowledged count, time-to-acknowledge
- ≥80% line coverage
- **v2.0**: `AssignAlert` — sets `assigned_to` on in-memory alert, produces `alert_assigned` audit event
- **v2.0**: Audit emitter wired to both `AcknowledgeAlert` and `AssignAlert` handlers

---

## 2. Service Structure

```
svc-alert/
├── cmd/
│   └── alert/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── domain/
│   │   ├── alert_queue.go           # Priority queue implementation
│   │   ├── alert_queue_test.go
│   │   ├── acknowledger.go          # Alert acknowledgment logic
│   │   └── acknowledger_test.go
│   ├── consumer/
│   │   ├── alert_consumer.go
│   │   └── alert_consumer_test.go
│   ├── handler/
│   │   ├── stream.go               # StreamAlerts handler
│   │   ├── acknowledge.go          # AcknowledgeAlert handler
│   │   ├── details.go              # GetAlertDetails handler
│   │   └── handler_test.go
│   └── mapper/
│       └── alert_mapper.go
├── go.mod
├── Dockerfile
└── README.md
```

---

## 3. Alert Priority Queue

```go
// CLASSIFICATION: UNCLASSIFIED
package domain

// AlertQueue is a priority queue that orders alerts by:
//   1. Severity (CRITICAL > ELEVATED > WATCH) — higher priority first
//   2. Timestamp (newer first within same severity)
// Thread-safe via mutex.
type AlertQueue struct {
    mu          sync.RWMutex
    alerts      map[string]*QueuedAlert // key: alert_id
    byPriority  []*QueuedAlert          // sorted slice
    maxSize     int                      // default: 10000
    onChange     func(alert *inferencev1.AnomalyAlert)
}

// QueuedAlert wraps an alert with queue metadata.
type QueuedAlert struct {
    Alert        *inferencev1.AnomalyAlert
    QueuedAt     time.Time
    Acknowledged bool
    AckedBy      string
    AckedAt      *time.Time
}

// Enqueue adds an alert to the priority queue.
// If queue is at maxSize, drops lowest priority alert.
// Triggers onChange callback for streaming clients.
func (q *AlertQueue) Enqueue(alert *inferencev1.AnomalyAlert) { /* implementation */ }

// Acknowledge marks an alert as acknowledged by an operator.
func (q *AlertQueue) Acknowledge(alertID, operatorID, comment string) (*time.Time, error) { /* implementation */ }

// Get returns a specific alert by ID.
func (q *AlertQueue) Get(alertID string) (*QueuedAlert, bool) { /* implementation */ }

// GetUnacknowledged returns all unacknowledged alerts in priority order.
func (q *AlertQueue) GetUnacknowledged() []*QueuedAlert { /* implementation */ }

// UnacknowledgedCount returns the count of unacknowledged alerts.
func (q *AlertQueue) UnacknowledgedCount() int { /* implementation */ }

// severityRank returns numeric rank for priority ordering.
// CRITICAL=3, ELEVATED=2, WATCH=1, NORMAL=0
func severityRank(s commonv1.AlertSeverity) int { /* implementation */ }
```

---

## 4. StreamAlerts Handler

```go
// CLASSIFICATION: UNCLASSIFIED
package handler

// StreamAlertsHandler implements AlertService.StreamAlerts.
// Flow:
//   1. Parse StreamAlertsRequest filters (min_severity, anomaly_types, entity_types, clearance)
//   2. Send all existing unacknowledged alerts matching filters
//   3. Subscribe to queue onChange notifications
//   4. For each new alert:
//      a. Apply filters (severity, type, entity type, classification)
//      b. If passes → send to client stream
//   5. Continue until disconnect
func (h *StreamAlertsHandler) StreamAlerts(
    req *inferencev1.StreamAlertsRequest,
    stream inferencev1.AlertService_StreamAlertsServer) error { /* implementation */ }
```

---

## 5. Metrics

| Metric                                           | Type      | Labels                     |
| ------------------------------------------------ | --------- | -------------------------- |
| `rtsa_alert_service_alerts_received_total`       | Counter   | `severity`, `anomaly_type` |
| `rtsa_alert_service_alerts_unacknowledged`       | Gauge     | `severity`                 |
| `rtsa_alert_service_stream_clients`              | Gauge     | -                          |
| `rtsa_alert_service_time_to_acknowledge_seconds` | Histogram | `severity`                 |
| `rtsa_alert_service_queue_size`                  | Gauge     | -                          |

---

## 6. Test Scenarios

| #   | Test                                 | Expected                             |
| --- | ------------------------------------ | ------------------------------------ |
| T01 | Enqueue CRITICAL then WATCH          | CRITICAL first in priority order     |
| T02 | Enqueue at max size                  | Lowest priority dropped              |
| T03 | Acknowledge existing alert           | Acknowledged=true, operator recorded |
| T04 | Acknowledge non-existent             | NOT_FOUND error                      |
| T05 | GetUnacknowledged                    | Only unacknowledged returned         |
| T06 | StreamAlerts: min_severity=ELEVATED  | WATCH alerts excluded                |
| T07 | StreamAlerts: filter by anomaly type | Only matching types sent             |
| T08 | StreamAlerts: classification filter  | Higher classified excluded           |
| T09 | GetAlertDetails                      | Full alert with features             |
| T10 | Time-to-acknowledge metric           | Correct duration recorded            |
| T11 | **v2.0** AssignAlert: valid alert    | assigned_to set; audit event produced |
| T12 | **v2.0** AssignAlert: non-existent   | NOT_FOUND error returned             |
| T13 | **v2.0** AssignAlert: audit event    | alert_assigned event in audit stream |

---

## 7. Agent Invocation

```
@greatest-ever-developer Implement Module 11 from docs/implementation/11-alert-service.md

Context:
- Read docs/implementation/00-implementation-overview.md for global conventions
- Read docs/implementation/02-protobuf-schemas.md for AlertService proto
- Read docs/architecture/component_design.md §8.2 for alert service diagram
- Priority queue: CRITICAL > ELEVATED > WATCH, then by timestamp
- StreamAlerts: initial dump of all unacknowledged, then incremental
- Classification filtering is MANDATORY

Deliverables:
1. Complete svc-alert/ with all files
2. Priority queue with severity ordering
3. StreamAlerts with filtering
4. AcknowledgeAlert handler
5. Unit tests (≥80% coverage)
6. Integration tests with testcontainers
```
