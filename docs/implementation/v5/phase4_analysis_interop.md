<!-- CLASSIFICATION: UNCLASSIFIED -->

# Phase 4 — Analysis & Interoperability Validation

> **Use Cases**: UC013 (Historical Query & Forensics), UC014 (NATO Outbound), UC015 (NATO Inbound)
> **Prerequisites**: Phase 3 complete — all UI and dashboard tests passing
> **Common Guidelines**: See [00_common_guidelines.md](00_common_guidelines.md)
> **Classification**: UNCLASSIFIED
> **Date**: 2026-03-10

---

## 1. Objectives

- Verify historical query service returns correct results with time-range, spatial, and classification filters
- Verify `GetEventTimeline` UNION ALL across 4 ClickHouse tables
- Verify materialized views for dashboard KPIs
- Verify NATO outbound track nomination + classification guard
- Verify NATO inbound track reception + classification mapping
- Run all related unit, integration, E2E, and benchmark tests — zero silent failures
- Compile all issues into Issue Log, then fix in batch

---

## 2. UC013 — Historical Query & Forensics

**Spec**: `docs/business/usecases/UC013_historical_query.md`
**Feature**: FEAT-14 (Historical Analysis), FEAT-20 (Unified Event Timeline)
**Requirements**: CR-HIS-001 through CR-HIS-009

### 2.1 Code Review Checklist

| # | Review Item | File(s) | Requirement |
|---|-------------|---------|-------------|
| 1 | `QueryTracks` handler: time-range + entity_type + classification filters | `svc-query/internal/handler/` | CR-HIS-002 |
| 2 | `QueryAnomalies` handler: time-range + severity + classification filters | handler | CR-HIS-002 |
| 3 | `GetEventTimeline` handler: UNION ALL across `tracks_fused`, `anomaly_detections`, `operator_feedback`, `audit_log` | handler / SQL | CR-HIS-008 |
| 4 | GetEventTimeline classification filter applied to ALL 4 UNION ALL branches | SQL query | CR-SEC-001 |
| 5 | Events ordered by `event_time ASC` | SQL ORDER BY | CR-HIS-008 |
| 6 | Max events configurable (default 200, max 1000) | handler config | CR-HIS-008 |
| 7 | Simple queries complete within 500ms (CR-HIS-006) | measured via tests | NFR |
| 8 | ClickHouse `tracks_fused` table stores all fused track updates | ETL pipeline | CR-HIS-001 |
| 9 | ClickHouse `anomaly_detections` stores scored anomaly events | ETL pipeline | CR-HIS-001 |
| 10 | ClickHouse `sensor_observations` stores raw sensor events | ETL pipeline | CR-HIS-001 |
| 11 | ClickHouse `operator_feedback` stores submitted feedback | ETL pipeline | CR-HIS-001 |
| 12 | ClickHouse `audit_log` append-only (no TTL) per ITSG-33 AU-11 | table schema | CR-HIS-005 |
| 13 | Materialized view `mv_active_tracks_by_domain` at 10s granularity | ClickHouse DDL | CR-HIS-009 |
| 14 | Materialized view `mv_sensor_throughput_5min` rolling 5-min rate | ClickHouse DDL | CR-HIS-009 |
| 15 | Materialized view `mv_alert_ack_latency` by severity | ClickHouse DDL | CR-HIS-009 |
| 16 | Redpanda Connect ETL pipelines: tracks → ClickHouse, audit → ClickHouse | pipeline configs | CR-HIS-001 |
| 17 | 90-day sensor data retention (DC), 2-year audit retention | table TTL settings | CR-HIS-004/005 |
| 18 | Aggregation queries complete within 5s | test assertions | CR-HIS-007 |

### 2.2 Tests to Run

```bash
# Unit tests
cd svc-query && go test -race -count=1 -v ./...

# Coverage (target: ≥85% for handlers)
cd svc-query && go test -race -coverprofile=coverage.out ./... && \
  go tool cover -func=coverage.out | tail -1

# Integration tests: IT11, IT12, IT13
bash scripts/dev/test-integration.sh
# IT11: Tracks → ClickHouse (100 rows materialized)
# IT12: Audit → ClickHouse (ITSG-33 no TTL)
# IT13: Materialized view aggregations

# Integration: query_test.go
bash scripts/dev/test-integration.sh
# Verify tests/integration/query_test.go passes

# E2E: Forensics query
bash scripts/dev/test-e2e.sh
# Verify tests/e2e/forensics_query_test.go passes

# Benchmark: B04 (Query 100-row p95 ≤ 500ms)
bash scripts/dev/test-bench.sh
```

### 2.3 ClickHouse Schema Verification

```bash
# Verify tables exist
docker exec rtsa-clickhouse clickhouse-client --database rtsa \
  --query "SHOW TABLES"
# Expected: tracks_fused, anomaly_detections, sensor_observations,
#           operator_feedback, audit_log

# Verify materialized views exist
docker exec rtsa-clickhouse clickhouse-client --database rtsa \
  --query "SELECT name FROM system.tables WHERE name LIKE 'mv_%'"
# Expected: mv_active_tracks_by_domain, mv_sensor_throughput_5min, mv_alert_ack_latency

# Verify audit_log has NO TTL (ITSG-33 AU-11)
docker exec rtsa-clickhouse clickhouse-client --database rtsa \
  --query "SHOW CREATE TABLE audit_log" | grep -i ttl
# Expected: no TTL clause

# Verify sensor_observations has 90-day TTL
docker exec rtsa-clickhouse clickhouse-client --database rtsa \
  --query "SHOW CREATE TABLE sensor_observations" | grep -i ttl
```

### 2.4 Expected Outcomes

- Query service unit tests pass with ≥85% handler coverage
- IT11, IT12, IT13 all pass — ETL pipeline verified
- Forensics E2E test passes — historical query returns correct results
- B04: Query 100-row response p95 ≤ 500ms
- Audit log has no TTL; sensor data has 90-day TTL
- All 3 materialized views present and producing correct aggregations

---

## 3. UC014 — NATO Outbound Data Exchange

**Spec**: `docs/business/usecases/UC014_nato_outbound.md`
**Feature**: FEAT-15 (NATO Data Exchange)
**Requirements**: CR-NATO-001 through CR-NATO-005

### 3.1 Code Review Checklist

| # | Review Item | File(s) | Requirement |
|---|-------------|---------|-------------|
| 1 | NominateTracks RPC handler — nominates tracks for NATO sharing | `svc-nato-adapter/internal/handler/` | CR-NATO-001 |
| 2 | Classification guard: blocks tracks above configured release level | handler logic | CR-NATO-005 |
| 3 | STANAG 5516 message formatting (J-Series) | formatter | CR-NATO-001 |
| 4 | NFFI XML message formatting | formatter | CR-NATO-002 |
| 5 | Classification mapping: NATO ↔ GC levels | mapper | CR-NATO-004 |
| 6 | Nomination produces audit event with `resource_type = 'nato_nomination'` | producer | CR-SEC-003 |
| 7 | Revocation workflow — revoke previously nominated track | handler | FEAT-15 |
| 8 | Blocked nomination: clear audit trail with reason | audit event | CR-NATO-005 |

### 3.2 Tests to Run

```bash
# Unit tests
cd svc-nato-adapter && go test -race -count=1 -v ./...

# Coverage
cd svc-nato-adapter && go test -race -coverprofile=coverage.out ./... && \
  go tool cover -func=coverage.out | tail -1

# Serializer round-trip test (if exists)
# Verify STANAG/NFFI message parse → serialize → parse round-trip
```

### 3.3 Expected Outcomes

- NATO adapter unit tests pass with ≥80% coverage
- Classification guard correctly blocks above-ceiling tracks
- STANAG 5516 and NFFI message formatting verified

---

## 4. UC015 — NATO Inbound Data Exchange

**Spec**: `docs/business/usecases/UC015_nato_inbound.md`
**Feature**: FEAT-15 (NATO Data Exchange)
**Requirements**: CR-NATO-001 through CR-NATO-005

### 4.1 Code Review Checklist

| # | Review Item | File(s) | Requirement |
|---|-------------|---------|-------------|
| 1 | Inbound STANAG 5516 message parser | `svc-nato-adapter/internal/` | CR-NATO-001 |
| 2 | Inbound NFFI XML parser | parser | CR-NATO-002 |
| 3 | Inbound classification mapping: NATO → GC level | mapper | CR-NATO-004 |
| 4 | Inbound tracks marked with NATO source + REL TO label | track metadata | CR-NATO-001 |
| 5 | Inbound tracks integrated into fusion engine or displayed separately | track flow | FEAT-15 |
| 6 | Cross-domain guard enforced at inbound boundary | guard logic | CR-NATO-005 |

### 4.2 Tests to Run

```bash
# Same service — verify inbound-specific tests
cd svc-nato-adapter && go test -race -count=1 -v -run "Inbound" ./...
```

### 4.3 Expected Outcomes

- Inbound STANAG/NFFI parsing tested
- Classification mapping NATO → GC verified
- REL TO labels correctly applied to inbound tracks

---

## 5. Serializer Round-Trip Test

**Reference**: v4 Implementation Review R-027

| # | Review Item | File(s) |
|---|-------------|---------|
| 1 | Go→Rust serializer round-trip test exists | `tests/integration/serializer_roundtrip_test.go` |
| 2 | Go `Serializer.Serialize()` output matches Rust `wasm-decoder` field extraction | round-trip test |
| 3 | All 128-byte record fields match between Go and Rust at correct offsets | round-trip test |

```bash
# Run serializer round-trip test
bash scripts/dev/test-integration.sh
# Verify tests/integration/serializer_roundtrip_test.go passes
```

---

## 6. Classification Propagation Chain (IT14 — MANDATORY)

Although tested in integration suite, explicitly verify here as it crosses UC013/UC014/UC015 boundaries:

```bash
# IT14: MANDATORY classification propagation chain
bash scripts/dev/test-integration.sh
# Verify IT14 passes — full classification propagation from sensor → fusion → anomaly → query → NATO

# Neg03: Classification violation rejected
bash scripts/dev/test-e2e.sh
# Verify Neg03 passes
```

---

## 7. Issue Log

_Issues found during Phase 4 review are recorded here and fixed in batch._

| # | UC | Severity | File(s) | Issue Description | Requirement |
|---|----|----------|---------|-------------------|-------------|
| | | | | _(to be populated during execution)_ | |

---

## 8. Batch Fix Execution

After the Issue Log is populated:

1. Fix all BLOCKING issues first
2. Fix WARNING issues
3. Fix IMPROVEMENT items
4. Re-run ALL tests from sections 2–6 to confirm no regressions

---

## 9. Phase 4 Completion Criteria

- [ ] UC013: Query service passes unit tests with ≥85% handler coverage
- [ ] UC013: IT11, IT12, IT13 pass (ETL pipeline verified)
- [ ] UC013: GetEventTimeline UNION ALL returns correct results
- [ ] UC013: Forensics E2E test passes
- [ ] UC013: B04 passes (query latency threshold)
- [ ] UC013: Materialized views producing correct aggregations
- [ ] UC013: Audit log has no TTL (ITSG-33 AU-11)
- [ ] UC014: NATO outbound nomination + classification guard verified
- [ ] UC015: NATO inbound parsing + classification mapping verified
- [ ] IT14: MANDATORY classification propagation chain passes
- [ ] Neg03: Classification violation rejection verified
- [ ] Serializer round-trip test passes
- [ ] All Issue Log items resolved
- [ ] No silent test failures in any stage
