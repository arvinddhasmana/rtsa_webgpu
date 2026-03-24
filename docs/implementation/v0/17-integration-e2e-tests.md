<!-- CLASSIFICATION: UNCLASSIFIED -->

# Module 17 — Integration & End-to-End Tests

> **Module**: 17-integration-e2e-tests
> **Phase**: P6 (Validation)
> **Dependencies**: ALL prior modules (00–16)
> **Agent**: `@greatest-ever-developer`
> **Estimated Effort**: 5 days

---

## 1. Objective

Implement the full integration and end-to-end test suite that validates the entire RTSA system from sensor ingestion through fusion, anomaly detection, alerting, and UI query — performed against real infrastructure using Docker Compose.

**Acceptance Criteria**:

- Full-pipeline E2E tests: simulator → ingestion → fusion → anomaly → alert → query
- Component integration tests using testcontainers-go (Redpanda + ClickHouse)
- Performance benchmarks with defined thresholds
- Classification propagation validation across the full chain
- All tests run in CI via `make test-integration` and `make test-e2e`
- Test report with pass/fail summary and timing

---

## 2. Test Structure

```
tests/
├── integration/
│   ├── ingestion_test.go          # Ingestion → Redpanda topic validation
│   ├── fusion_test.go             # Ingestion → Fusion → tracks topic
│   ├── anomaly_test.go            # Fusion → Anomaly → alerts topic
│   ├── feedback_test.go           # Feedback submission → trust scoring → topics
│   ├── etl_test.go                # Redpanda → ClickHouse materialization
│   ├── query_test.go              # ClickHouse → Query Service results
│   ├── audit_test.go              # Audit trail completeness
│   ├── classification_test.go     # Classification propagation chain
│   └── testutil/
│       ├── setup.go               # Docker Compose test environment setup
│       ├── helpers.go             # Test assertion helpers
│       └── fixtures.go            # Test data fixtures
├── e2e/
│   ├── full_pipeline_test.go      # End-to-end pipeline test
│   ├── alert_workflow_test.go     # Alert generation → acknowledgment
│   ├── feedback_workflow_test.go  # Feedback → trust → validated topic
│   └── forensics_query_test.go   # Historical query validation
├── benchmark/
│   ├── ingestion_bench_test.go    # Ingestion throughput benchmark
│   ├── fusion_bench_test.go       # Fusion latency benchmark
│   ├── anomaly_bench_test.go      # Anomaly detection latency benchmark
│   └── query_bench_test.go        # Query response time benchmark
├── go.mod
├── docker-compose.test.yml        # Test-specific Docker Compose
├── Makefile
└── README.md
```

---

## 3. Test Environment Setup

### 3.1 Docker Compose Test Configuration

```yaml
# CLASSIFICATION: UNCLASSIFIED
# tests/docker-compose.test.yml
# Minimal test environment — single instance of each infrastructure component
version: "3.9"

services:
  redpanda:
    image: redpandadata/redpanda:v24.1.1
    command:
      - redpanda start
      - --smp 1
      - --memory 512M
      - --mode dev-container
      - --kafka-addr 0.0.0.0:9092
      - --schema-registry-addr 0.0.0.0:8081
    ports:
      - "9092:9092"
      - "8081:8081"
    healthcheck:
      test: ["CMD", "rpk", "cluster", "health"]
      interval: 5s
      timeout: 5s
      retries: 10

  clickhouse:
    image: clickhouse/clickhouse-server:24.3
    ports:
      - "9000:9000"
      - "8123:8123"
    environment:
      CLICKHOUSE_USER: rtsa_test
      CLICKHOUSE_PASSWORD: test_password
      CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT: 1
    volumes:
      - ../deploy/clickhouse/init.sql:/docker-entrypoint-initdb.d/init.sql:ro
    healthcheck:
      test: ["CMD", "clickhouse-client", "--query", "SELECT 1"]
      interval: 5s
      timeout: 5s
      retries: 10
```

### 3.2 Test Setup Helper

```go
// CLASSIFICATION: UNCLASSIFIED
package testutil

import (
    "context"
    "testing"
    "time"
)

// TestEnv holds references to test infrastructure.
type TestEnv struct {
    RedpandaBrokers  string
    ClickHouseDSN    string
    SchemaRegistryURL string
    ctx              context.Context
    cancel           context.CancelFunc
}

// SetupTestEnv creates a complete test environment.
// Uses testcontainers-go for isolated, reproducible test infrastructure.
//
// Starts:
//   - Redpanda (Kafka-compatible, Schema Registry)
//   - ClickHouse with RTSA schema initialized
//   - Creates all 21 Redpanda topics
//
// Build tag: //go:build integration
// Environment guard: RTSA_INTEGRATION_TESTS=true
func SetupTestEnv(t *testing.T) *TestEnv {
    t.Helper()
    if os.Getenv("RTSA_INTEGRATION_TESTS") != "true" {
        t.Skip("Skipping integration test: set RTSA_INTEGRATION_TESTS=true")
    }
    // ... start containers, create topics, init schema
    return &TestEnv{}
}

// Teardown cleans up test infrastructure.
func (e *TestEnv) Teardown() {
    e.cancel()
    // ... stop containers
}

// WaitForTopicMessages waits until at least `count` messages appear on a topic.
// Timeout: configurable, default 30s.
func (e *TestEnv) WaitForTopicMessages(t *testing.T, topic string, count int, timeout time.Duration) []*kgo.Record {
    // ... consume and wait
}

// QueryClickHouse executes a query and returns rows.
func (e *TestEnv) QueryClickHouse(t *testing.T, query string, args ...interface{}) *sql.Rows {
    // ... execute query
}
```

---

## 4. Integration Tests

### 4.1 Ingestion Integration Test

```go
// CLASSIFICATION: UNCLASSIFIED
//go:build integration

package integration

// TestRadarIngestionToTopic validates:
//   1. Send radar observation via gRPC to ingestion service
//   2. Verify message appears on sensors.radar.tracks topic
//   3. Verify message headers are set (rtsa-classification, rtsa-source-service, etc.)
//   4. Verify protobuf deserialization of message value
//   5. Verify audit event generated on audit.events topic
func TestRadarIngestionToTopic(t *testing.T) { /* ... */ }

// TestInvalidObservationToDLQ validates:
//   1. Send observation with invalid coordinates (lat=999)
//   2. Verify message appears on sensors.radar.dlq topic
//   3. Verify original message is preserved in DLQ
//   4. Verify rejection metric incremented
func TestInvalidObservationToDLQ(t *testing.T) { /* ... */ }

// TestMultiSensorIngestion validates:
//   1. Send observations to all 6 sensor ingestion services
//   2. Verify each appears on correct topic
//   3. Verify classification headers propagated
func TestMultiSensorIngestion(t *testing.T) { /* ... */ }
```

### 4.2 Fusion Integration Test

```go
// CLASSIFICATION: UNCLASSIFIED
//go:build integration

package integration

// TestSensorToFusedTrack validates:
//   1. Produce 3 radar observations for same entity (close position, same type)
//   2. Wait for fusion engine to correlate
//   3. Verify fused track appears on tracks.fused.surface topic
//   4. Verify track.source_count >= 2
//   5. Verify Kalman-filtered position between observations
func TestSensorToFusedTrack(t *testing.T) { /* ... */ }

// TestTrackMerging validates:
//   1. Create 2 tracks with overlapping positions
//   2. Produce observation that correlates both (score ≥ 0.85)
//   3. Verify one track gets MERGED status
//   4. Verify surviving track has combined source count
func TestTrackMerging(t *testing.T) { /* ... */ }

// TestStaleTrackTimeout validates:
//   1. Create a track with observations
//   2. Stop sending updates
//   3. Wait 60+ seconds
//   4. Verify track status changes to STALE on topic
//   5. Wait 5+ minutes
//   6. Verify track status changes to DROPPED
func TestStaleTrackTimeout(t *testing.T) { /* ... */ }

// TestClassificationPropagation validates:
//   1. Send observation with CLASSIFICATION_PROTECTED_B
//   2. Send observation for same entity with CLASSIFICATION_SECRET
//   3. Verify fused track classification = SECRET (MAX rule)
func TestClassificationPropagation(t *testing.T) { /* ... */ }
```

### 4.3 Anomaly Detection Integration Test

```go
// CLASSIFICATION: UNCLASSIFIED
//go:build integration

package integration

// TestSpeedAnomalyDetection validates:
//   1. Create fused track with normal speed (12 knots, 10 updates)
//   2. Send update with anomalous speed (50 knots)
//   3. Verify alert appears on alerts.anomaly.* topic
//   4. Verify anomaly_type = SPEED
//   5. Verify confidence > 0.5
//   6. Verify explanation is human-readable
func TestSpeedAnomalyDetection(t *testing.T) { /* ... */ }

// TestAISManipulationDetection validates:
//   1. Create track with radar observations at position A
//   2. Send AIS observation at position A + 1.0 NM offset
//   3. Verify alert with anomaly_type = AIS_MANIPULATION
//   4. Verify alert references correct track_id
func TestAISManipulationDetection(t *testing.T) { /* ... */ }

// TestAlertSeverityRouting validates:
//   1. Generate anomaly with confidence ≥ 0.90
//   2. Verify alert on alerts.anomaly.critical topic
//   3. Generate anomaly with confidence 0.70-0.89
//   4. Verify alert on alerts.anomaly.elevated topic
//   5. Generate anomaly with confidence 0.50-0.69
//   6. Verify alert on alerts.anomaly.watch topic
func TestAlertSeverityRouting(t *testing.T) { /* ... */ }
```

### 4.4 ETL Integration Test

```go
// CLASSIFICATION: UNCLASSIFIED
//go:build integration

package integration

// TestTracksETLToClickHouse validates:
//   1. Produce 100 fused track messages to tracks.fused.surface
//   2. Wait for Redpanda Connect ETL to process (max 30s)
//   3. Query ClickHouse tracks_fused table
//   4. Verify 100 rows present with correct field mapping
//   5. Verify classification_level correctly mapped
func TestTracksETLToClickHouse(t *testing.T) { /* ... */ }

// TestAuditETLToClickHouse validates:
//   1. Produce audit events to audit.events topic
//   2. Wait for ETL processing
//   3. Query ClickHouse audit_log table
//   4. Verify events present with no TTL (check table schema)
func TestAuditETLToClickHouse(t *testing.T) { /* ... */ }

// TestMaterializedViews validates:
//   1. Insert track data for 3 entity types
//   2. Query mv_track_count_by_type
//   3. Verify counts per entity type are correct
//   4. Insert anomaly data for 2 types
//   5. Query mv_anomaly_summary_hourly
//   6. Verify aggregations are correct
func TestMaterializedViews(t *testing.T) { /* ... */ }
```

### 4.5 Classification Chain Test

```go
// CLASSIFICATION: UNCLASSIFIED
//go:build integration

package integration

// TestClassificationEndToEnd validates classification propagation
// across the entire pipeline:
//
//   1. Sensor observation at PROTECTED_B
//      → verify Redpanda header: rtsa-classification=PROTECTED_B
//   2. Second observation at SECRET for same entity
//      → verify fused track classification = SECRET (MAX rule)
//   3. Anomaly alert on that track
//      → verify alert classification = SECRET (inherited)
//   4. Query service with PROTECTED_B caller
//      → verify track NOT returned (clearance too low)
//   5. Query service with SECRET caller
//      → verify track IS returned
//   6. All operations generate audit events
//      → verify audit events carry correct classification
func TestClassificationEndToEnd(t *testing.T) { /* ... */ }
```

---

## 5. End-to-End Tests

### 5.1 Full Pipeline Test

```go
// CLASSIFICATION: UNCLASSIFIED
//go:build e2e

package e2e

// TestFullPipeline validates the complete data flow:
//
// Simulator → Ingestion → Redpanda → Fusion → Redpanda → Anomaly Detection
//   → Redpanda → Alert Service → Query Service (via ClickHouse)
//
// Steps:
//   1. Start simulator with seed=42, 10 surface entities, 5 air, 2 subsurface
//   2. Run for 60 seconds
//   3. Verify at least 10 fused tracks appear on tracks.fused.* topics
//   4. Verify at least 1 anomaly alert (5% injection rate, ~1 of 17 entities)
//   5. Query ClickHouse via Query Service for tracks in time range
//   6. Verify results match expected entity count (±20% tolerance)
//   7. Query audit log — verify >0 audit events per service
//
// Timeout: 5 minutes
func TestFullPipeline(t *testing.T) { /* ... */ }
```

### 5.2 Alert Workflow Test

```go
// CLASSIFICATION: UNCLASSIFIED
//go:build e2e

package e2e

// TestAlertWorkflow validates the operator alert workflow:
//
//   1. Inject speed anomaly via simulator
//   2. Wait for alert on alerts.anomaly.* topic
//   3. Stream alerts via Alert Service StreamAlerts
//   4. Verify alert received by stream client
//   5. Acknowledge alert via AcknowledgeAlert RPC
//   6. Verify alert marked as acknowledged
//   7. Verify time-to-acknowledge metric recorded
//   8. Verify audit event for acknowledgment
func TestAlertWorkflow(t *testing.T) { /* ... */ }
```

### 5.3 Feedback Workflow Test

```go
// CLASSIFICATION: UNCLASSIFIED
//go:build e2e

package e2e

// TestFeedbackWorkflow validates the operator feedback loop:
//
//   1. Create a fused track (via ingestion + fusion)
//   2. Submit CONFIRM_HOSTILE feedback via Feedback Service
//   3. Verify feedback appears on feedback.operator.submissions topic
//   4. Verify trust score computed (trust ≥ 0)
//   5. If trust ≥ 0.5: verify on feedback.operator.validated topic
//   6. Verify feedback stored in ClickHouse via ETL
//   7. Query feedback via Query Service
//   8. Verify audit trail for feedback submission
func TestFeedbackWorkflow(t *testing.T) { /* ... */ }
```

---

## 6. Performance Benchmarks

### 6.1 Benchmark Thresholds

| Metric                                     | Threshold     | Measurement                    |
| ------------------------------------------ | ------------- | ------------------------------ |
| Ingestion throughput (single service)      | ≥1000 obs/sec | Radar ingestion gRPC           |
| Fusion latency (observation → fused track) | ≤100ms p95    | Topic timestamp delta          |
| Anomaly detection latency (track → alert)  | ≤200ms p95    | Topic timestamp delta          |
| ETL latency (Redpanda → ClickHouse)        | ≤5s p99       | Ingestion time check           |
| Query response (100 rows)                  | ≤500ms p95    | gRPC response time             |
| Query response (10K rows)                  | ≤2s p95       | gRPC response time             |
| Alert streaming latency                    | ≤50ms p95     | Alert publish → client receive |
| Track streaming latency                    | ≤50ms p95     | Track publish → client receive |

### 6.2 Benchmark Tests

```go
// CLASSIFICATION: UNCLASSIFIED
//go:build integration

package benchmark

// BenchmarkIngestionThroughput measures the maximum
// ingestion rate for the radar ingestion service.
//
// Setup: Start radar ingestion service with testcontainer Redpanda
// Execute: Send 10000 observations as fast as possible
// Measure: observations/second
// Threshold: Must exceed 1000 obs/sec
func BenchmarkIngestionThroughput(b *testing.B) { /* ... */ }

// BenchmarkFusionLatency measures the time from
// sensor observation publish to fused track publish.
//
// Setup: Start fusion engine with testcontainer Redpanda
// Execute: Publish 1000 observations, measure track output time
// Measure: p50, p95, p99 latency
// Threshold: p95 ≤ 100ms
func BenchmarkFusionLatency(b *testing.B) { /* ... */ }

// BenchmarkAnomalyLatency measures the time from
// fused track publish to anomaly alert publish.
//
// Setup: Start anomaly detection with testcontainer Redpanda
// Execute: Publish 1000 tracks with 10% anomalous
// Measure: p50, p95, p99 latency
// Threshold: p95 ≤ 200ms
func BenchmarkAnomalyLatency(b *testing.B) { /* ... */ }

// BenchmarkQueryPerformance measures ClickHouse query
// response time through the Query Service.
//
// Setup: Insert 100K tracks into ClickHouse
// Execute: Run 100 queries with varied filters
// Measure: p50, p95, p99 response time
// Threshold: 100 rows ≤ 500ms p95, 10K rows ≤ 2s p95
func BenchmarkQueryPerformance(b *testing.B) { /* ... */ }
```

---

## 7. Test Data Fixtures

```go
// CLASSIFICATION: UNCLASSIFIED
package testutil

// MidAtlanticPosition returns a random position in the operational area.
// Lat: 43.0–47.0, Lon: -65.0–-55.0
func MidAtlanticPosition(rng *rand.Rand) *commonv1.Position { /* ... */ }

// ValidRadarObservation returns a fully valid radar observation.
func ValidRadarObservation() *ingestionv1.SensorObservation { /* ... */ }

// ValidAISObservation returns a fully valid AIS observation.
func ValidAISObservation() *ingestionv1.SensorObservation { /* ... */ }

// FusedTrackFixture returns a valid fused track for testing.
func FusedTrackFixture(entityType commonv1.EntityType) *entityv1.FusedTrack { /* ... */ }

// AnomalyAlertFixture returns a valid anomaly alert for testing.
func AnomalyAlertFixture(severity commonv1.AlertSeverity) *inferencev1.AnomalyAlert { /* ... */ }

// OperatorFeedbackFixture returns a valid operator feedback for testing.
func OperatorFeedbackFixture(feedbackType feedbackv1.FeedbackType) *feedbackv1.OperatorFeedback { /* ... */ }

// AuditEventFixture returns a valid audit event for testing.
func AuditEventFixture(eventType, service string) *auditv1.AuditEvent { /* ... */ }
```

---

## 8. Makefile Targets

```makefile
# CLASSIFICATION: UNCLASSIFIED
# tests/Makefile

.PHONY: test-integration test-e2e test-bench test-all

# Run integration tests (requires Docker)
test-integration:
	RTSA_INTEGRATION_TESTS=true go test -v -tags integration -timeout 10m ./integration/...

# Run end-to-end tests (requires full Docker Compose stack)
test-e2e:
	docker compose -f docker-compose.test.yml up -d --wait
	RTSA_INTEGRATION_TESTS=true go test -v -tags e2e -timeout 15m ./e2e/...
	docker compose -f docker-compose.test.yml down -v

# Run performance benchmarks
test-bench:
	RTSA_INTEGRATION_TESTS=true go test -v -tags integration -bench=. -benchtime=10s -timeout 10m ./benchmark/...

# Run all tests
test-all: test-integration test-e2e test-bench
```

---

## 9. CI Integration

```yaml
# CLASSIFICATION: UNCLASSIFIED
# .github/workflows/integration-tests.yml (reference — not created by this module)
# Integration tests run on PR merge to main

# Expected CI steps:
#   1. Checkout code
#   2. Start Docker Compose test infrastructure
#   3. Build all services
#   4. Run integration tests: make test-integration
#   5. Run E2E tests: make test-e2e
#   6. Run benchmarks: make test-bench
#   7. Upload test report
#   8. Fail if any test fails or benchmark threshold exceeded
```

---

## 10. Test Scenarios Summary

| #     | Category       | Test                 | Validates                       |
| ----- | -------------- | -------------------- | ------------------------------- |
| IT01  | Ingestion      | Radar → topic        | gRPC → Redpanda                 |
| IT02  | Ingestion      | Invalid → DLQ        | Validation + DLQ routing        |
| IT03  | Ingestion      | All 6 sensors        | Multi-sensor ingestion          |
| IT04  | Fusion         | 3 obs → fused track  | Correlation + Kalman            |
| IT05  | Fusion         | Track merging        | Score ≥ 0.85 merge              |
| IT06  | Fusion         | Stale timeout        | 60s → STALE, 5m → DROPPED       |
| IT07  | Fusion         | Classification MAX   | PROTECTED_B + SECRET → SECRET   |
| IT08  | Anomaly        | Speed anomaly        | >3σ detection                   |
| IT09  | Anomaly        | AIS manipulation     | Position delta >0.5 NM          |
| IT10  | Anomaly        | Severity routing     | CRITICAL/ELEVATED/WATCH topics  |
| IT11  | ETL            | Tracks → ClickHouse  | 100 rows materialized           |
| IT12  | ETL            | Audit → ClickHouse   | No TTL verified                 |
| IT13  | ETL            | Materialized views   | Aggregations correct            |
| IT14  | Classification | End-to-end chain     | Full classification propagation |
| E2E01 | Pipeline       | Full pipeline        | Sim → Alert → Query             |
| E2E02 | Workflow       | Alert workflow       | Generate → Stream → Acknowledge |
| E2E03 | Workflow       | Feedback workflow    | Submit → Trust → Validate       |
| B01   | Performance    | Ingestion throughput | ≥1000 obs/sec                   |
| B02   | Performance    | Fusion latency       | ≤100ms p95                      |
| B03   | Performance    | Anomaly latency      | ≤200ms p95                      |
| B04   | Performance    | Query response       | ≤500ms p95 (100 rows)           |

---

## 11. Agent Invocation

```
@greatest-ever-developer Implement Module 17 from docs/implementation/17-integration-e2e-tests.md

Context:
- Read docs/implementation/00-implementation-overview.md for all module interfaces
- Read ALL prior module docs (01-16) for service APIs and topic schemas
- Read docs/architecture/data_architecture.md for topic map and ClickHouse DDL
- Read docs/sdlc_guidelines/05_testing/testing_strategy.md for test requirements

Deliverables:
1. tests/ directory with complete structure
2. testutil/ package with setup helpers and fixtures
3. Integration tests (IT01-IT14) using testcontainers-go
4. E2E tests (E2E01-E2E03) using Docker Compose
5. Performance benchmarks (B01-B04) with threshold assertions
6. docker-compose.test.yml for test infrastructure
7. Makefile with test targets

CRITICAL:
- All integration tests guarded by RTSA_INTEGRATION_TESTS=true
- All integration tests use //go:build integration tag
- All E2E tests use //go:build e2e tag
- Classification propagation test is MANDATORY (IT14)
- Benchmark thresholds MUST be asserted (fail if exceeded)
- Test fixtures use Mid-Atlantic coordinates (43-47°N, 55-65°W)
- Tests must be deterministic (use seed-based randomness)
```
