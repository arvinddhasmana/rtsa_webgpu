<!-- CLASSIFICATION: UNCLASSIFIED -->

# Phase 1 — Foundation & Sensor Ingestion Validation

> **Use Cases**: UC001 (System Initialization), UC002–UC007 (Sensor Ingestion × 6 types), UC008 (Multi-Source Fusion)
> **Prerequisites**: None — this is the first phase
> **Common Guidelines**: See [00_common_guidelines.md](00_common_guidelines.md)
> **Classification**: UNCLASSIFIED
> **Date**: 2026-03-10

---

## 1. Objectives

- Verify the platform bootstraps correctly (Docker stack, Redpanda topics, ClickHouse schema)
- Verify all 6 sensor ingestion services validate, normalize, and produce to correct Redpanda topics
- Verify invalid data is routed to dead-letter queue (DLQ)
- Verify the fusion engine correlates multi-sensor data into unified entity tracks
- Verify classification propagation follows MAX rule
- Run all related unit, integration, E2E, and benchmark tests — zero silent failures

---

## 2. UC001 — System Initialization & Platform Bootstrap

**Spec**: `docs/business/usecases/UC001_system_initialization.md`
**Feature**: FEAT-01 (Platform Infrastructure), FEAT-02 (Security Framework), FEAT-03 (Event Streaming Backbone)
**Requirements**: CR-SEC-002, NFR-AVAIL-001, NFR-AVAIL-003

### 2.1 Code Review Checklist

| # | Review Item | File(s) |
|---|-------------|---------|
| 1 | Docker Compose YAML parses without errors | `deploy/docker-compose.yml`, `deploy/docker-compose.services.yml` |
| 2 | All 14 Go services listed in Compose | `deploy/docker-compose.services.yml` |
| 3 | Health check script covers ALL services (Redpanda, ClickHouse, all svc-*) | `scripts/dev/health-check.sh` |
| 4 | init-topics creates ALL topic families: `sensors.*`, `tracks.*`, `alerts.*`, `feedback.*`, `audit.*` | `scripts/dev/init-topics.sh` |
| 5 | init-clickhouse creates core tables: `tracks_fused`, `anomaly_detections`, `sensor_observations`, `operator_feedback`, `audit_log` | `scripts/dev/init-clickhouse.sh` |
| 6 | init-clickhouse creates v2.0 materialized views: `mv_active_tracks_by_domain`, `mv_sensor_throughput_5min`, `mv_alert_ack_latency` | `scripts/dev/init-clickhouse.sh` |
| 7 | setup-dev.sh validates all prerequisites (Docker, Go, Node, Rust, wasm-pack) | `scripts/setup/setup-dev.sh` |
| 8 | mTLS certificates for inter-service communication referenced in Compose | `deploy/`, `certs/` |

### 2.2 Tests to Run

```bash
# 1. Validate Docker Compose syntax
docker compose -f deploy/docker-compose.yml config --quiet
docker compose -f deploy/docker-compose.services.yml config --quiet

# 2. Start full stack
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml up -d --wait

# 3. Initialize
bash scripts/dev/init-topics.sh
bash scripts/dev/init-clickhouse.sh

# 4. Verify
bash scripts/dev/health-check.sh
# Expected: ALL services show [✓]
```

### 2.3 Expected Outcomes

- All Docker Compose files parse without syntax errors
- All services reach `healthy` status within 90 seconds
- Health check shows `[✓]` for every service
- Redpanda has ≥12 topics created
- ClickHouse has 5 tables + 3 materialized views in `rtsa` database

---

## 3. UC002–UC007 — Sensor Ingestion (All 6 Types)

**Spec**: `docs/business/usecases/UC002_radar_ingestion.md` through `UC007_cyber_threat_ingestion.md`
**Feature**: FEAT-04 through FEAT-09
**Requirements**: CR-ING-001 through CR-ING-012

### 3.1 Service Matrix

| UC | Sensor Type | Service Directory | Redpanda Topic |
|----|-------------|-------------------|----------------|
| UC002 | Radar | `svc-radar-ingestion/` | `sensors.radar.validated` |
| UC003 | EW/SIGINT | `svc-ew-ingestion/` | `sensors.ew.validated` |
| UC004 | ELINT/COMINT | `svc-elint-ingestion/` | `sensors.elint.validated` |
| UC005 | ISR Metadata | `svc-isr-ingestion/` | `sensors.isr.validated` |
| UC006 | AIS/BFT | `svc-ais-ingestion/` | `sensors.ais.validated` |
| UC007 | Cyber Threat | `svc-cyber-ingestion/` | `sensors.cyber.validated` |

### 3.2 Code Review Checklist (per service)

| # | Review Item | Expected |
|---|-------------|----------|
| 1 | gRPC handler `IngestSensorEvent` validates all required fields | Coordinates, timestamps, sensor_id, sensor_type validated |
| 2 | Invalid data produces to DLQ topic with error reason | `sensors.<type>.dlq` or equivalent |
| 3 | Valid data normalizes to Protobuf and produces to validated topic | `sensors.<type>.validated` |
| 4 | handler_test.go exists with table-driven tests | Happy path + invalid inputs |
| 5 | domain/*_test.go exists for validation logic | ≥90% coverage target |
| 6 | Classification field propagated from sensor event | CR-ING-007 |
| 7 | Shared ingestion handler from `pkg/ingestion/` used correctly | Common handler pattern |

### 3.3 Tests to Run

```bash
# Unit tests — per service
for svc in svc-radar-ingestion svc-ew-ingestion svc-elint-ingestion \
           svc-isr-ingestion svc-ais-ingestion svc-cyber-ingestion; do
  echo "=== Testing $svc ==="
  (cd "$svc" && go test -race -count=1 -v ./...)
done

# Unit tests with coverage — per service
for svc in svc-radar-ingestion svc-ew-ingestion svc-elint-ingestion \
           svc-isr-ingestion svc-ais-ingestion svc-cyber-ingestion; do
  echo "=== Coverage: $svc ==="
  (cd "$svc" && go test -race -coverprofile=coverage.out ./... && \
   go tool cover -func=coverage.out | tail -1)
done

# Shared pkg/ingestion tests
cd pkg && go test -race -count=1 -v ./ingestion/...

# Integration tests: IT01, IT02, IT03
bash scripts/dev/test-integration.sh
# Verify IT01 (Radar→topic), IT02 (Invalid→DLQ), IT03 (All 6 sensors)

# Benchmark: B01 (Ingestion ≥ 1,000 obs/sec)
bash scripts/dev/test-bench.sh
```

### 3.4 Expected Outcomes

- All 6 services compile and all unit tests pass
- Each service achieves ≥80% line coverage (≥90% for domain/validation logic)
- IT01, IT02, IT03 pass — sensor data flows through Redpanda correctly
- B01: Ingestion throughput ≥ 1,000 obs/sec

---

## 4. UC008 — Multi-Source Fusion

**Spec**: `docs/business/usecases/UC008_multi_source_fusion.md`
**Feature**: FEAT-10 (Multi-Source Data Fusion)
**Requirements**: CR-FUS-001 through CR-FUS-007

### 4.1 Code Review Checklist

| # | Review Item | File(s) |
|---|-------------|---------|
| 1 | Gating criteria match spec: spatial (<10km air, <5km surface), temporal (<30s), speed (±50%), heading (±45°) | `svc-fusion-engine/internal/domain/` |
| 2 | Correlation scoring weights match spec: position=0.35, kinematics=0.25, sensor affinity=0.15, historical=0.25 | `svc-fusion-engine/internal/domain/` |
| 3 | Entity type assigned correctly (AIR/SURFACE/SUBSURFACE/LAND/SPACE/CYBER) per CR-FUS-002 | domain logic |
| 4 | Hostile status assigned correctly (UNKNOWN/PENDING/FRIENDLY/NEUTRAL/HOSTILE/SUSPECT) per CR-FUS-003 | domain logic |
| 5 | Confidence score calculated per CR-FUS-004 | domain logic |
| 6 | Track merge flow: two tracks → one when correlation confirms same entity | domain logic |
| 7 | Track stale timeout: no updates for config duration → STALE/LOST | domain logic |
| 8 | Classification = MAX of contributing sensors (CR-FUS-007 / IT07) | domain logic |
| 9 | Audit events produced for track creation, merge, split | producer code |
| 10 | Output fused track contains all required fields from UC008 §7 | proto/track types |
| 11 | Unit tests cover: happy path, merge, split, stale, classification propagation | `*_test.go` files |

### 4.2 Tests to Run

```bash
# Unit tests
cd svc-fusion-engine && go test -race -count=1 -v ./...

# Coverage (target: ≥90% for domain logic)
cd svc-fusion-engine && go test -race -coverprofile=coverage.out ./... && \
  go tool cover -func=coverage.out | tail -1

# Integration tests: IT04, IT05, IT06, IT07
bash scripts/dev/test-integration.sh
# Verify: IT04 (3obs→fused track), IT05 (track merging), IT06 (stale timeout), IT07 (classification MAX)

# Benchmark: B02 (Fusion p95 ≤ 100ms)
bash scripts/dev/test-bench.sh
```

### 4.3 Expected Outcomes

- Fusion engine domain logic ≥90% line coverage
- IT04–IT07 all pass
- B02: Fusion latency p95 ≤ 100ms
- Classification propagation follows MAX rule (IT07 passing confirms this)

---

## 5. E2E Pipeline Validation (Full Phase 1 Flow)

After all use cases above are individually reviewed:

```bash
# E2E01: Full pipeline — Sim → Ingestion → Fusion → Alert → Query
bash scripts/dev/test-e2e.sh
# Verify E2E01 passes: sensor simulator data flows through ingestion,
# fusion produces tracks, anomaly detection scores them, query returns them
```

### Expected Outcomes

- E2E01 passes — complete data flow from sensor simulator to query service
- No silent test failures (check `test-results/e2e/failures.txt`)

---

## 6. Issue Log

_Issues found during Phase 1 review are recorded here and fixed in batch after all UC001–UC008 reviews are complete._

| # | UC | Severity | File(s) | Issue Description | Requirement |
|---|----|----------|---------|-------------------|-------------|
| 1 | UC002-UC007 | IMPROVEMENT | `svc-*-ingestion/` | Unit test coverage is ~65%, below target | Target ≥80% |
| 2 | UC008 | IMPROVEMENT | `svc-fusion-engine/` | Unit test coverage is 72.7%, below target | Target ≥90% |


---

## 7. Batch Fix Execution

After the Issue Log is populated:

1. Fix all BLOCKING issues first
2. Fix WARNING issues
3. Fix IMPROVEMENT items
4. Re-run ALL tests from sections 2–5 to confirm no regressions

---

## 8. Phase 1 Completion Criteria

- [ ] UC001: Docker stack starts, health check passes, topics/tables created
- [ ] UC002–UC007: All 6 ingestion services pass unit tests with ≥80% coverage
- [ ] UC002–UC007: IT01, IT02, IT03 pass
- [ ] UC008: Fusion engine passes unit tests with ≥90% domain coverage
- [ ] UC008: IT04, IT05, IT06, IT07 pass
- [ ] E2E01 passes (full pipeline)
- [ ] B01 (ingestion throughput) and B02 (fusion latency) pass
- [ ] All Issue Log items resolved
- [ ] No silent test failures in any stage
