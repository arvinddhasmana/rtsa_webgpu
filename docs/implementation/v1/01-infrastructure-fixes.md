<!-- CLASSIFICATION: UNCLASSIFIED -->

# v1 Module 01 — Infrastructure Fixes

> **Module**: 01-infrastructure-fixes
> **Phase**: P0 (Foundation — must be done first)
> **Dependencies**: None
> **Agent**: `@greatest-ever-developer`
> **Estimated Effort**: 2 days
> **Traceability**: UC001, FEAT-01, FEAT-03, CR-SEC-002, BUG-01..BUG-06

---

## 1. Objective

Fix 6 critical and medium-severity infrastructure configuration bugs discovered during the full-system audit. These bugs **block the end-to-end demo pipeline** — Envoy cannot reach backend services, ETL pipelines read from non-existent topics, and ClickHouse inserts fail due to table/column mismatches.

**Authority rule**: The service source code (Go constants, Redpanda Connect YAML configs) is authoritative for topic and table names. The init scripts (`init-topics.sh`, `init-clickhouse.sh`) and Envoy configs must be aligned to match them.

---

## 2. Task 1: Fix Envoy Upstream Ports (BUG-01)

### Problem

In `deploy/envoy/envoy-dev.yaml`, all 5 backend cluster endpoints use **host-mapped ports** instead of the **container-internal port**. In Docker networking, containers communicate via internal ports.

| Cluster            | Current (Wrong)      | Correct              |
| ------------------ | -------------------- | -------------------- |
| `track_service`    | `svc-track:50070`    | `svc-track:50051`    |
| `alert_service`    | `svc-alert:50071`    | `svc-alert:50051`    |
| `query_service`    | `svc-query:50072`    | `svc-query:50051`    |
| `feedback_service` | `svc-feedback:50062` | `svc-feedback:50051` |
| `audit_service`    | `svc-audit:50073`    | `svc-audit:50051`    |

### Instructions

1. Open `deploy/envoy/envoy-dev.yaml`
2. In the `clusters:` section, find each cluster's `load_assignment.endpoints[0].lb_endpoints[0].endpoint.address.socket_address.port_value`
3. Change all 5 port values to `50051`
4. Repeat for `deploy/envoy/envoy.yaml` (production config) — same pattern applies

### Files to Modify

- `deploy/envoy/envoy-dev.yaml` — lines ~148 (track), ~163 (alert), ~178 (query), ~193 (feedback), ~208 (audit)
- `deploy/envoy/envoy.yaml` — corresponding cluster sections

### Validation

```bash
# Verify no non-50051 ports remain in cluster endpoints
grep -A 20 "cluster_name:" deploy/envoy/envoy-dev.yaml | grep "port_value:"
# Expected: all lines show port_value: 50051
```

---

## 3. Task 2: Unify Redpanda Topic Names (BUG-02)

### Problem

`scripts/dev/init-topics.sh` creates topics using a different naming convention than the service code and Redpanda Connect ETL configs expect.

| init-topics.sh Creates          | Services/ETL Expect                                                                                              | Match?                |
| ------------------------------- | ---------------------------------------------------------------------------------------------------------------- | --------------------- |
| `sensor.raw.radar`              | `sensors.radar.tracks`                                                                                           | **NO**                |
| `sensor.raw.ew_sigint`          | `sensors.ew.intercepts`                                                                                          | **NO**                |
| `sensor.raw.elint_comint`       | `sensors.elint.detections`                                                                                       | **NO**                |
| `sensor.raw.isr`                | `sensors.isr.observations`                                                                                       | **NO**                |
| `sensor.raw.ais_bft`            | `sensors.ais.positions`                                                                                          | **NO**                |
| `sensor.raw.cyber`              | `sensors.cyber.iocs`                                                                                             | **NO**                |
| `sensor.raw.radar.dlq`          | `dlq.sensors.radar`                                                                                              | **NO**                |
| `sensor.raw.ew_sigint.dlq`      | `dlq.sensors.ew`                                                                                                 | **NO**                |
| `sensor.raw.elint_comint.dlq`   | `dlq.sensors.elint`                                                                                              | **NO**                |
| `sensor.raw.isr.dlq`            | `dlq.sensors.isr`                                                                                                | **NO**                |
| `sensor.raw.ais_bft.dlq`        | `dlq.sensors.ais`                                                                                                | **NO**                |
| `sensor.raw.cyber.dlq`          | `dlq.sensors.cyber`                                                                                              | **NO**                |
| `entity.tracks.fused`           | `tracks.fused.surface`, `tracks.fused.air`, `tracks.fused.subsurface`, `tracks.fused.land`, `tracks.fused.cyber` | **NO** (1 topic vs 5) |
| `inference.anomaly.scores`      | `alerts.anomaly.critical`, `alerts.anomaly.elevated`, `alerts.anomaly.watch`                                     | **NO** (1 topic vs 3) |
| `feedback.operator.submissions` | `feedback.operator.submissions`                                                                                  | YES                   |
| `feedback.operator.validated`   | `feedback.operator.validated`                                                                                    | YES                   |
| `audit.events`                  | `audit.events`                                                                                                   | YES                   |
| `nato.exchange.inbound`         | `nato.exchange.inbound`                                                                                          | YES                   |
| `nato.exchange.outbound`        | `nato.exchange.outbound`                                                                                         | YES                   |

### Instructions

Replace the `create_all_topics()` function body in `scripts/dev/init-topics.sh` with:

```bash
create_all_topics() {
  echo ""
  log_info "Creating Redpanda topics for local development..."
  echo ""

  # ── Sensor validated topics — 24h retention (services produce here) ──
  create_topic "sensors.radar.tracks"       3 86400000  1
  create_topic "sensors.ew.intercepts"      3 86400000  1
  create_topic "sensors.elint.detections"   3 86400000  1
  create_topic "sensors.isr.observations"   3 86400000  1
  create_topic "sensors.ais.positions"      3 86400000  1
  create_topic "sensors.cyber.iocs"         3 86400000  1

  # ── Dead-letter queues for sensor topics ──
  create_topic "dlq.sensors.radar"          1 604800000 1
  create_topic "dlq.sensors.ew"             1 604800000 1
  create_topic "dlq.sensors.elint"          1 604800000 1
  create_topic "dlq.sensors.isr"            1 604800000 1
  create_topic "dlq.sensors.ais"            1 604800000 1
  create_topic "dlq.sensors.cyber"          1 604800000 1

  # ── Fused track topics — 48h retention (per entity type) ──
  create_topic "tracks.fused.surface"       3 172800000 1
  create_topic "tracks.fused.air"           3 172800000 1
  create_topic "tracks.fused.subsurface"    3 172800000 1
  create_topic "tracks.fused.land"          3 172800000 1
  create_topic "tracks.fused.cyber"         3 172800000 1

  # ── Alert topics — 48h retention (per severity) ──
  create_topic "alerts.anomaly.critical"    2 172800000 1
  create_topic "alerts.anomaly.elevated"    2 172800000 1
  create_topic "alerts.anomaly.watch"       2 172800000 1

  # ── Feedback topics — 7 day retention ──
  create_topic "feedback.operator.submissions" 1 604800000 1
  create_topic "feedback.operator.validated"   1 604800000 1

  # ── Audit events — 30 day retention in dev ──
  create_topic "audit.events"               2 2592000000 1

  # ── NATO exchange — 24h retention ──
  create_topic "nato.exchange.inbound"      1 86400000  1
  create_topic "nato.exchange.outbound"     1 86400000  1

  # ── Model lifecycle — 7 day retention ──
  create_topic "models.anomaly.candidates"  1 604800000 1
  create_topic "models.anomaly.published"   1 604800000 1

  echo ""
  log_pass "All topics created/verified."
}
```

### Cross-Reference Check

After updating `init-topics.sh`, verify these topic names match:

1. **svc-radar-ingestion** config defaults: `sensors.radar.tracks`, `dlq.sensors.radar` ✓
2. **svc-ew-ingestion** `ingestion.MustLoad()`: `sensors.ew.intercepts`, `dlq.sensors.ew` ✓
3. **svc-elint-ingestion**: `sensors.elint.detections`, `dlq.sensors.elint` ✓
4. **svc-isr-ingestion**: `sensors.isr.observations`, `dlq.sensors.isr` ✓
5. **svc-ais-ingestion**: `sensors.ais.positions`, `dlq.sensors.ais` ✓
6. **svc-cyber-ingestion**: `sensors.cyber.iocs`, `dlq.sensors.cyber` ✓
7. **svc-track consumer**: `tracks.fused.surface/air/subsurface/land/cyber` ✓
8. **svc-alert consumer**: `alerts.anomaly.critical/elevated/watch` ✓
9. **Redpanda Connect sensors YAML**: `sensors.radar.tracks`, `sensors.ew.intercepts`, etc. ✓
10. **Redpanda Connect tracks YAML**: `tracks.fused.surface/air/subsurface/land/cyber` ✓
11. **Redpanda Connect alerts YAML**: `alerts.anomaly.critical/elevated/watch` ✓

---

## 4. Task 3: Unify ClickHouse Table and Column Names (BUG-03, BUG-04)

### Problem

`scripts/dev/init-clickhouse.sh` creates tables with different names and column schemas than the Redpanda Connect ETL configs expect.

| init-clickhouse.sh Creates | Redpanda Connect Writes To | Match? |
| -------------------------- | -------------------------- | ------ |
| `sensor_events`            | `sensor_observations`      | **NO** |
| `entity_tracks`            | `tracks_fused`             | **NO** |
| `anomaly_scores`           | `anomaly_detections`       | **NO** |
| `feedback_events`          | `operator_feedback`        | **NO** |
| `audit_events`             | `audit_log`                | **NO** |

Additionally, column names differ (e.g., `sensor_events.raw_payload` vs `sensor_observations.metadata_json`; `entity_tracks.hostile_status` vs `tracks_fused.hostile_classification`; etc.).

### Instructions

Replace all 5 `CREATE TABLE` statements in `scripts/dev/init-clickhouse.sh` to match the Redpanda Connect field mappings exactly. The authoritative column names come from the Redpanda Connect YAML `columns:` sections.

#### Table 1: `sensor_observations` (was `sensor_events`)

```sql
CREATE TABLE IF NOT EXISTS sensor_observations
(
    observation_id       String,
    sensor_id            String,
    sensor_type          LowCardinality(String),
    latitude             Float64,
    longitude            Float64,
    altitude_meters      Float64,
    speed_knots          Float64,
    heading_degrees      Float64,
    classification_level LowCardinality(String),
    metadata_json        String,
    event_time           DateTime64(3, 'UTC'),
    ingestion_time       DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_time)
ORDER BY (sensor_type, sensor_id, event_time)
TTL toDateTime(event_time) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;
```

#### Table 2: `tracks_fused` (was `entity_tracks`)

```sql
CREATE TABLE IF NOT EXISTS tracks_fused
(
    track_id               String,
    entity_type            LowCardinality(String),
    hostile_classification LowCardinality(String),
    latitude               Float64,
    longitude              Float64,
    altitude_meters        Float64,
    speed_knots            Float64,
    heading_degrees        Float64,
    confidence_score       Float32,
    source_count           UInt16,
    source_sensors         Array(String),
    classification_level   LowCardinality(String),
    track_status           LowCardinality(String),
    event_time             DateTime64(3, 'UTC'),
    ingestion_time         DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_time)
ORDER BY (entity_type, track_id, event_time)
TTL toDateTime(event_time) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;
```

#### Table 3: `anomaly_detections` (was `anomaly_scores`)

```sql
CREATE TABLE IF NOT EXISTS anomaly_detections
(
    alert_id             String,
    track_id             String,
    anomaly_type         LowCardinality(String),
    severity             LowCardinality(String),
    confidence_score     Float32,
    explanation          String,
    model_version        String,
    classification_level LowCardinality(String),
    event_time           DateTime64(3, 'UTC'),
    ingestion_time       DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_time)
ORDER BY (track_id, event_time)
TTL toDateTime(event_time) + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;
```

#### Table 4: `operator_feedback` (was `feedback_events`)

```sql
CREATE TABLE IF NOT EXISTS operator_feedback
(
    feedback_id          String,
    track_id             String,
    operator_id          String,
    feedback_type        LowCardinality(String),
    justification        String,
    trust_score          Float32,
    clearance_score      Float32,
    accuracy_score       Float32,
    temporal_score       Float32,
    deviation_score      Float32,
    validated            UInt8 DEFAULT 0,
    classification_level LowCardinality(String),
    event_time           DateTime64(3, 'UTC'),
    ingestion_time       DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_time)
ORDER BY (operator_id, event_time)
SETTINGS index_granularity = 8192;
```

#### Table 5: `audit_log` (was `audit_events`)

```sql
-- ITSG-33: AU-3 — Content of Audit Records
-- NO TTL — immutable audit trail (CR-SEC-003)
CREATE TABLE IF NOT EXISTS audit_log
(
    audit_id             String,
    service_id           LowCardinality(String),
    event_type           LowCardinality(String),
    actor_id             String,
    actor_type           LowCardinality(String),
    resource_type        LowCardinality(String),
    resource_id          String,
    action               LowCardinality(String),
    detail_json          String,
    classification_level LowCardinality(String),
    event_time           DateTime64(3, 'UTC'),
    ingestion_time       DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_time)
ORDER BY (service_id, event_type, event_time)
SETTINGS index_granularity = 8192;
```

### Impact on svc-query and svc-audit

After renaming tables, check if `svc-query` and `svc-audit` repository code references the old table names. If they do, update the SQL query strings to use the new names:

- `svc-query/internal/repository/tracks_repo.go` — update `sensor_events` → `sensor_observations`, `entity_tracks` → `tracks_fused`, `anomaly_scores` → `anomaly_detections`
- `svc-query/internal/repository/anomaly_repo.go` — update `anomaly_scores` → `anomaly_detections`
- `svc-query/internal/repository/audit_repo.go` — update `audit_events` → `audit_log`
- `svc-audit/internal/repository/audit_repo.go` — update `audit_events` → `audit_log`

Run `grep -rn "sensor_events\|entity_tracks\|anomaly_scores\|audit_events\|feedback_events" svc-*/` to find all references.

### Materialized Views

If `init-clickhouse.sh` creates materialized views referencing old table names, update those too. Check the integration tests at `tests/integration/etl_test.go` which validate materialized view queries.

---

## 5. Task 4: Move Envoy to Services Compose (BUG-05)

### Problem

The `envoy` service is declared in `deploy/docker-compose.yml` (infrastructure stack) but its `depends_on` references services (`svc-track`, `svc-alert`, etc.) that only exist in `deploy/docker-compose.services.yml` (application overlay). Running `docker compose -f deploy/docker-compose.yml up` fails.

### Instructions

1. **Remove** the `envoy` service block from `deploy/docker-compose.yml`
2. **Add** the `envoy` service block to `deploy/docker-compose.services.yml`
3. Ensure the `envoy` service's `depends_on` references are valid in the services overlay context
4. Update the `volumes` mount for certs to use the correct relative path from the services compose file

### Validation

```bash
# Infrastructure-only should work without errors
docker compose -f deploy/docker-compose.yml up -d

# Full stack should work with envoy able to reach services
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml up -d
```

---

## 6. Task 5: Generate Missing TypeScript Audit Client (BUG-06)

### Problem

The TypeScript code generation produces `gen/ts/rtsa/audit/v1/audit_event_pb.ts` but is missing:

- `gen/ts/rtsa/audit/v1/audit_service_pb.ts`
- `gen/ts/rtsa/audit/v1/audit_service_connect.ts`

### Instructions

1. Check `buf.gen.yaml` — verify the `out` and `opt` settings for the `@connectrpc/protoc-gen-connect-es` plugin include all proto files including `proto/rtsa/audit/v1/audit_service.proto`
2. Run `buf generate` and verify both files are generated
3. If a path exclusion or glob filter is blocking `audit_service.proto`, remove the filter

### Validation

```bash
buf generate
ls gen/ts/rtsa/audit/v1/
# Expected: audit_event_pb.ts, audit_service_pb.ts, audit_service_connect.ts
```

---

## 7. Task 6: ETL Integration Validation Test

After applying all fixes above, add a quick integration validation to confirm the full data path works.

### Instructions

Add a test or script that:

1. Produces a test message to `sensors.radar.tracks` with valid Redpanda Connect-expected JSON structure
2. Waits up to 10 seconds for a row to appear in ClickHouse `sensor_observations` table
3. Validates the row column values match the produced message

This can be added to `tests/integration/etl_test.go` as an additional test case, or as a standalone validation script at `scripts/dev/validate-etl.sh`.

---

## 8. Test Scenarios

| #   | Test                                                      | Expected                                   | Module Task |
| --- | --------------------------------------------------------- | ------------------------------------------ | ----------- |
| T01 | Envoy health check after port fix                         | All 5 backend clusters reachable           | Task 1      |
| T02 | Produce to `sensors.radar.tracks`                         | Message consumed by Redpanda Connect       | Task 2      |
| T03 | All topics from init-topics.sh match service configs      | `rpk topic list` shows all expected topics | Task 2      |
| T04 | ETL: sensor → ClickHouse `sensor_observations`            | Row appears with correct columns           | Task 3      |
| T05 | ETL: track → ClickHouse `tracks_fused`                    | Row appears with correct columns           | Task 3      |
| T06 | ETL: alert → ClickHouse `anomaly_detections`              | Row appears with correct columns           | Task 3      |
| T07 | ETL: feedback → ClickHouse `operator_feedback`            | Row appears with correct columns           | Task 3      |
| T08 | ETL: audit → ClickHouse `audit_log` (no TTL)              | Row appears, no TTL on table               | Task 3      |
| T09 | `docker compose -f docker-compose.yml up` without errors  | Infra-only starts clean                    | Task 4      |
| T10 | `buf generate` produces TS audit service                  | 3 files in `gen/ts/rtsa/audit/v1/`         | Task 5      |
| T11 | Full stack health check via `make health-check`           | All services report healthy                | All tasks   |
| T12 | svc-query SQL queries work against new table names        | Queries return results                     | Task 3      |
| T13 | svc-audit repo works against `audit_log` table            | Audit insert + query work                  | Task 3      |
| T14 | Existing integration tests pass (`make integration-test`) | All IT01-IT14 pass                         | All tasks   |

---

## 9. Agent Invocation

```
@greatest-ever-developer Implement v1 Module 01 from docs/implementation/v1/01-infrastructure-fixes.md

Context:
- Read docs/implementation/v1/00-v1-overview.md for v1 scope and traceability
- Read docs/implementation/00-implementation-overview.md for global conventions
- CRITICAL: These are configuration fixes, not new service development
- The service source code is AUTHORITATIVE — init scripts and Envoy configs must match
- Topic names: services use sensors.radar.tracks, tracks.fused.*, alerts.anomaly.* (NOT sensor.raw.*)
- ClickHouse tables: Redpanda Connect uses sensor_observations, tracks_fused, anomaly_detections, operator_feedback, audit_log
- All 5 Envoy cluster endpoints must use port 50051 (container port, NOT host-mapped ports)
- After fixing, run existing integration tests to confirm nothing breaks
- Check svc-query and svc-audit repository code for hardcoded old table names

Files to modify:
1. deploy/envoy/envoy-dev.yaml — fix 5 port_value entries
2. deploy/envoy/envoy.yaml — fix 5 port_value entries (if same pattern)
3. scripts/dev/init-topics.sh — replace create_all_topics() function
4. scripts/dev/init-clickhouse.sh — replace all 5 CREATE TABLE statements
5. deploy/docker-compose.yml — remove envoy service block
6. deploy/docker-compose.services.yml — add envoy service block
7. buf.gen.yaml — fix if audit_service.proto excluded
8. svc-query/internal/repository/*.go — update table names if needed
9. svc-audit/internal/repository/*.go — update table names if needed

Deliverables:
1. All 6 bugs fixed (BUG-01 through BUG-06)
2. init-topics.sh creates correct topics matching service code
3. init-clickhouse.sh creates correct tables matching ETL configs
4. Envoy routes to port 50051 on all clusters
5. buf generate produces TS audit service files
6. All existing tests continue to pass
7. Full stack starts successfully with make docker-up-all
```
