<!-- CLASSIFICATION: UNCLASSIFIED -->

# RTSA Integration & End-to-End Tests — Module 17

This directory contains the full integration, end-to-end, and performance benchmark test suite for the RTSA system.

## Structure

```
tests/
├── integration/          # Integration tests (IT01–IT14) using testcontainers-go
│   ├── testutil/         # Shared test utilities (setup, helpers, fixtures)
│   ├── ingestion_test.go # IT01–IT03: Radar ingestion → Redpanda topics
│   ├── fusion_test.go    # IT04–IT07: Fusion engine correlation & classification
│   ├── anomaly_test.go   # IT08–IT10: Anomaly detection & severity routing
│   ├── feedback_test.go  # Feedback trust scoring pipeline
│   ├── etl_test.go       # IT11–IT13: Redpanda → ClickHouse ETL
│   ├── query_test.go     # Query service + classification filter
│   ├── audit_test.go     # Audit trail completeness (ITSG-33 AU-11)
│   └── classification_test.go  # IT14: MANDATORY classification propagation chain
├── e2e/                  # End-to-end tests (E2E01–E2E03) using Docker Compose
│   ├── full_pipeline_test.go   # E2E01: Sim → Ingestion → Fusion → Alert → Query
│   ├── full_pipeline_test.go   # E2E02: Alert generation → acknowledgment workflow
│   ├── feedback_workflow_test.go # E2E03: Feedback → Trust → Validated
│   └── forensics_query_test.go # Historical forensics query validation
├── benchmark/            # Performance benchmarks (B01–B04) with threshold assertions
│   ├── ingestion_bench_test.go # B01: Ingestion ≥ 1000 obs/sec
│   │                           # B02: Fusion p95 ≤ 100ms
│   │                           # B03: Anomaly p95 ≤ 200ms
│   └── query_bench_test.go     # B04: Query 100 rows ≤ 500ms p95
├── docker-compose.test.yml     # Test-specific Docker Compose (Redpanda + ClickHouse)
└── Makefile                    # Test targets
```

## Prerequisites

- Go 1.22+
- Docker (for testcontainers-go and E2E tests)

## Running Tests

### Integration Tests (testcontainers — self-contained)

```bash
RTSA_INTEGRATION_TESTS=true go test -v -tags integration -timeout 10m ./integration/...
```

Or via Make:
```bash
cd tests && make test-integration
```

### End-to-End Tests (requires running services)

```bash
# Start infrastructure
docker compose -f tests/docker-compose.test.yml up -d --wait

# Run E2E tests
RTSA_INTEGRATION_TESTS=true go test -v -tags e2e -timeout 15m ./e2e/...

# Tear down
docker compose -f tests/docker-compose.test.yml down -v
```

Or via Make:
```bash
cd tests && make test-e2e
```

### Performance Benchmarks

```bash
RTSA_INTEGRATION_TESTS=true go test -v -tags integration -bench=. -benchtime=10s -timeout 10m ./benchmark/...
```

Or via Make:
```bash
cd tests && make test-bench
```

## Test Scenarios

| #     | Category       | Test                               | Validates                       |
|-------|----------------|------------------------------------|-------------------------------|
| IT01  | Ingestion      | Radar → topic                      | gRPC → Redpanda                |
| IT02  | Ingestion      | Invalid → DLQ                      | Validation + DLQ routing       |
| IT03  | Ingestion      | All 6 sensors                      | Multi-sensor ingestion         |
| IT04  | Fusion         | 3 obs → fused track                | Correlation + Kalman           |
| IT05  | Fusion         | Track merging                      | Score ≥ 0.85 merge             |
| IT06  | Fusion         | Stale timeout                      | 60s → STALE                   |
| IT07  | Fusion         | Classification MAX                 | PROTECTED_B + SECRET → SECRET  |
| IT08  | Anomaly        | Speed anomaly                      | >3σ detection                  |
| IT09  | Anomaly        | AIS manipulation                   | Position delta >0.5 NM         |
| IT10  | Anomaly        | Severity routing                   | CRITICAL/ELEVATED/WATCH topics |
| IT11  | ETL            | Tracks → ClickHouse                | 100 rows materialized          |
| IT12  | ETL            | Audit → ClickHouse                 | No TTL verified (ITSG-33)      |
| IT13  | ETL            | Materialized views                 | Aggregations correct           |
| **IT14** | **Classification** | **End-to-end chain** | **MANDATORY: Full propagation** |
| E2E01 | Pipeline       | Full pipeline                      | Sim → Alert → Query            |
| E2E02 | Workflow       | Alert workflow                     | Generate → Stream → Acknowledge|
| E2E03 | Workflow       | Feedback workflow                  | Submit → Trust → Validate      |
| B01   | Performance    | Ingestion throughput               | ≥ 1000 obs/sec                 |
| B02   | Performance    | Fusion latency                     | ≤ 100ms p95                    |
| B03   | Performance    | Anomaly latency                    | ≤ 200ms p95                    |
| B04   | Performance    | Query response                     | ≤ 500ms p95 (100 rows)         |

## Critical Requirements

- **IT14 is MANDATORY**: Classification propagation must be validated in every CI run.
- **Benchmark thresholds**: Failing a threshold causes `b.Errorf()` — CI will fail.
- **RTSA_INTEGRATION_TESTS=true**: All integration/E2E tests are guarded by this env var.
- **Build tags**: Integration tests use `//go:build integration`, E2E use `//go:build e2e`.
- **Deterministic**: All test fixtures use seed-based randomness (`NewSeededRand(seed)`).
- **Mid-Atlantic coordinates**: All geographic fixtures use 43–47°N, 55–65°W.

## Compliance

- ITSG-33 AU-11: Audit log immutability validated in IT12
- NATO STANAG 5516: Classification propagation validated in IT07, IT14
- NIST 800-53: Access control enforcement validated in IT14, query tests
