<!-- CLASSIFICATION: UNCLASSIFIED -->

# Simulated Feed Data in Development, Testing, and Demo

> **Classification**: UNCLASSIFIED
> **Module**: Data Simulation & Test Infrastructure
> **Last Updated**: March 2026
> **Applies To**: `tools/simulator/`, `scripts/demo/`, `scripts/dev/`, `tests/`

---

## Table of Contents

1. [Overview](#overview)
2. [Simulator Architecture](#simulator-architecture)
3. [Scenario Configurations](#scenario-configurations)
4. [Sensor Generator Breakdown](#sensor-generator-breakdown)
5. [Demo Workflow](#demo-workflow)
6. [Test Automation Integration](#test-automation-integration)
7. [Static Seed Data Strategy](#static-seed-data-strategy)
8. [Data Flow & Dependencies](#data-flow--dependencies)
9. [Reproducibility & Determinism](#reproducibility--determinism)
10. [Common Tasks](#common-tasks)

---

## Overview

RTSA uses a **dual-layer data generation strategy** for realistic synthetic testing without operational data:

| Layer               | Source                                            | Purpose                                                          | Scope                                                      |
| ------------------- | ------------------------------------------------- | ---------------------------------------------------------------- | ---------------------------------------------------------- |
| **Live Streaming**  | `tools/simulator/` (Go gRPC)                      | Real-time synthetic observations during demo & integration tests | All 6 sensor types; configurable entity counts & anomalies |
| **Historical Seed** | `scripts/demo/seed-demo-data.sh` (ClickHouse SQL) | Static forensic data for analytics queries & forensic dashboard  | 6 ClickHouse tables; 48-hour track history                 |

### Key Properties

- **All data is synthetic** — No real operational information
- **Deterministic** — Seed-based RNG ensures reproducible scenarios
- **Multi-domain** — Surface, Air, Subsurface + Cyber + ISR entities
- **Anomaly-aware** — Injected behavioral anomalies for AD testing
- **NATO-compatible** — STANAG 5516 data exchange scenarios included

---

## Simulator Architecture

### 6 Sensor Type Generators

The simulator generates observations for all RTSA sensor channels:

```
┌─────────────────────────────────────────────────────────────┐
│                      Simulator Engine                       │
│                   (tools/simulator/cmd/)                    │
├──────────────┬──────────────┬──────────────┬────────────────┤
│   Radar      │    EW/SIGINT │   ELINT/COMINT│ ISR           │
│  Port 50051  │  Port 50052  │  Port 50053   │ Port 50054    │
├──────────────┼──────────────┼──────────────┬────────────────┤
│   AIS/BFT    │    Cyber     │                                │
│  Port 50055  │  Port 50056  │                                │
└──────────────┴──────────────┴────────────────────────────────┘
```

### Entity Domain Generators

| Domain         | Entity Types                     | Typical Count | Data Source                                         |
| -------------- | -------------------------------- | ------------- | --------------------------------------------------- |
| **Surface**    | Ships, Vessels, Surface craft    | 3–200         | Simulated positions with maritime movement patterns |
| **Air**        | Aircraft, UAVs, Helicopters      | 2–100         | Radar altitude tracks                               |
| **Subsurface** | Submarines, Underwater platforms | 1–50          | ISR/sonobuoy observations                           |
| **Cyber**      | Network assets, DNS queries      | 5–50          | Anomalous traffic patterns                          |

### 5 Movement Patterns

Each entity generates observations following one of these kinematic models:

1. **Straight-Line**: Constant velocity transit (shipping lanes)
2. **Patrol**: Rectangular perimeter pattern (coastal surveillance)
3. **Loitering**: Circular orbit around waypoint (hover/orbit)
4. **Zigzag**: Evasion pattern (suspicious behavior)
5. **Random**: Brownian walk (general clutter)

### 5 Anomaly Types

Injected into observations at configurable rates (default 5%, stress tests 10–30%):

| Type                 | Detection Signal                      | Example                              |
| -------------------- | ------------------------------------- | ------------------------------------ |
| **Speed**            | Velocity spike 2× nominal             | Aircraft at 1500 knots (unrealistic) |
| **Route Deviation**  | Sigma deviation from baseline         | Ship off-course > 5 NM               |
| **AIS Manipulation** | Radar ↔ AIS position mismatch         | True position offset 0.5–2.0 NM      |
| **Behavioral**       | Unusual acceleration/heading          | Vessel erratic course changes        |
| **Proximity**        | Closest point of approach < threshold | Ships converging < 0.5 NM            |

---

## Scenario Configurations

All scenarios are YAML-based and located in `tools/simulator/scenarios/`. Each defines entity counts, anomaly rates, duration, and movement parameters.

### Scenario Listing

#### 1. **default.yaml** — Standard Demo

- **Purpose**: General-purpose demo; balanced mix of all sensor types
- **Duration**: Infinite (run until stopped)
- **Update Interval**: 1 second
- **Node Count**: 35 total entities
- **Sensor Generators**:
  - Radar: 7
  - EW/SIGINT: 5
  - ELINT/COMINT: 3
  - ISR: 4
  - AIS/BFT: 10
  - Cyber: 1
- **Anomaly Rate**: 5%
- **Movement Patterns**: Mixed (Patrol, Loitering, Transit)
- **Use Cases**:
  - Initial system bring-up
  - Health check before major tests
  - Baseline performance verification

#### 2. **maritime-demo.yaml** — Blue Water Scenario

- **Purpose**: Maritime-focused operations; surface + subsurface vessels
- **Duration**: 20 minutes
- **Update Interval**: 1.5 seconds
- **Node Count**: 28 total entities
- **Sensor Generators**:
  - Radar: 6
  - EW/SIGINT: 4
  - ELINT/COMINT: 2
  - ISR: 3
  - AIS/BFT: 12
  - Cyber: 1
- **Anomaly Rate**: 8%
- **Movement Patterns**: Patrol (shipping lanes), Transit, Loitering
- **Use Cases**:
  - Naval operations demo
  - Surface engagement scenarios
  - Shipping corridor monitoring
- **Operational Area**: Mid-Atlantic (43–47°N, 55–65°W)

#### 3. **multi-domain-demo.yaml** — Integrated Operations

- **Purpose**: All domains simultaneously (air, surface, subsurface, cyber)
- **Duration**: 25 minutes
- **Update Interval**: 1 second
- **Node Count**: 52 total entities
- **Sensor Generators**:
  - Radar: 9
  - EW/SIGINT: 6
  - ELINT/COMINT: 5
  - ISR: 7
  - AIS/BFT: 15
  - Cyber: 2
- **Anomaly Rate**: 7%
- **Movement Patterns**: All (Patrol, Loitering, Zigzag, Random)
- **Anomaly Distribution**:
  - Speed: 2 entities
  - Route Deviation: 2 entities
  - AIS Manipulation: 1 entity
  - Behavioral: 2 entities
  - Proximity: 0 entities
- **Use Cases**:
  - Multi-domain operations center demo
  - Fusion Engine stress testing
  - Analyst training (complex scenario)

#### 4. **sensor-health-demo.yaml** — Health Monitoring

- **Purpose**: Test Sensor Health Dashboard; verify all 6 ingestion service health endpoints respond
- **Duration**: 15 minutes
- **Update Interval**: 2 seconds
- **Node Count**: 18 total entities
- **Sensor Generators**:
  - Radar: 3
  - EW/SIGINT: 2
  - ELINT/COMINT: 2
  - ISR: 2
  - AIS/BFT: 5
  - Cyber: 1
- **Anomaly Rate**: 0% (clean data for health test)
- **Movement Patterns**: Straight-Line, Patrol (predictable)
- **Use Cases**:
  - E2E test: Sensor Health Dashboard (`e2e/sensor-health.spec.ts`)
  - CI/CD health verification
  - Development iteration cycle

#### 5. **anomaly-demo.yaml** — Anomaly Showcase

- **Purpose**: Demonstrate Anomaly Detection capabilities; high anomaly concentration
- **Duration**: 18 minutes
- **Update Interval**: 800 ms (faster updates)
- **Node Count**: 22 total entities
- **Sensor Generators**:
  - Radar: 5
  - EW/SIGINT: 4
  - ELINT/COMINT: 3
  - ISR: 3
  - AIS/BFT: 6
  - Cyber: 1
- **Anomaly Rate**: 30% (intentional high rate for demonstration)
- **Anomaly Trigger Points** (time-based):
  - t=3 min: Speed anomaly spike
  - t=7 min: Route deviations begin
  - t=12 min: AIS manipulation activates
  - t=16 min: Behavioral anomalies appear
- **Use Cases**:
  - Anomaly Detection service (Module 08) testing
  - Alert triggering validation
  - Stakeholder demonstrations

#### 6. **stress.yaml** — High-Volume Load Test

- **Purpose**: System capacity validation; stress test all components
- **Duration**: 30 minutes (or abort on CPU overload)
- **Update Interval**: 500 ms (2× normal frequency)
- **Node Count**: 350 total entities
- **Sensor Generators**:
  - Radar: 200
  - EW/SIGINT: 50
  - ELINT/COMINT: 30
  - ISR: 40
  - AIS/BFT: 25
  - Cyber: 5
- **Anomaly Rate**: 10%
- **Update Throughput**: ~5,000 observations/second sustained
- **Use Cases**:
  - Load testing ingestion services
  - Redpanda throughput validation
  - WebGPU rendering stress (>50k tracks)
  - ClickHouse ingestion pipeline capacity
- **Performance Targets**:
  - Maintain E2E latency < 16 ms (50k tracks @ 60 FPS)
  - Keep main thread CPU < 20%

#### 7. **nato-exchange.yaml** — STANAG 5516 Interop

- **Purpose**: NATO data exchange scenario; NFFI/MIPS adapters validation
- **Duration**: 22 minutes
- **Update Interval**: 1 second
- **Node Count**: 30 total entities (NATO-aligned callsigns)
- **Sensor Generators**:
  - Radar: 8
  - EW/SIGINT: 3
  - ELINT/COMINT: 2
  - ISR: 5
  - AIS/BFT: 10
  - Cyber: 2
- **Anomaly Rate**: 5%
- **Data Schema**: STANAG 5516 track format (J3PF compliance)
- **Use Cases**:
  - NATO adapter testing (`svc-nato-adapter`)
  - Allied force integration
  - Multi-nation fusion scenarios

---

## Sensor Generator Breakdown

### Summary Table: Sensor Generators by Scenario

| Scenario                    | Radar | EW/SIGINT | ELINT/COMINT | ISR | AIS/BFT | Cyber | **Total** |
| --------------------------- | ----- | --------- | ------------ | --- | ------- | ----- | --------- |
| **default.yaml**            | 7     | 5         | 3            | 4   | 10      | 1     | 30        |
| **maritime-demo.yaml**      | 6     | 4         | 2            | 3   | 12      | 1     | 28        |
| **multi-domain-demo.yaml**  | 9     | 6         | 5            | 7   | 15      | 2     | 44        |
| **sensor-health-demo.yaml** | 3     | 2         | 2            | 2   | 5       | 1     | 15        |
| **anomaly-demo.yaml**       | 5     | 4         | 3            | 3   | 6       | 1     | 22        |
| **stress.yaml**             | 200   | 50        | 30           | 40  | 25      | 5     | **350**   |
| **nato-exchange.yaml**      | 8     | 3         | 2            | 5   | 10      | 2     | 30        |

### Sensor Type Distribution Analysis

**Radar** (most common):

- Represents primary air defense & surface detection
- 200 instances in stress test (dominates at scale)
- 3–9 instances in standard scenarios

**EW/SIGINT** (electronic warfare):

- Signals intelligence complement to radar
- 2–6 instances balanced across scenarios
- Critical for multi-domain fusion

**ELINT/COMINT** (communications intelligence):

- Lower volume, high-value intelligence
- 2–5 instances per scenario
- Behavioral anomaly indicators

**ISR** (intelligence, surveillance, reconnaissance):

- Real-time surveillance streams
- 2–7 instances; scaling with multi-domain scenarios
- Primary anomaly detection trigger

**AIS/BFT** (automatic identification / Blue Force Tracking):

- Maritime vessel identification
- Highest civilian sensor count (6–15 instances)
- AIS manipulation anomalies concentrated here

**Cyber** (network threat):

- Network anomaly detection
- Smallest generator (1–5 instances)
- Growing importance for integrated cyber-physical scenarios

---

## Data Flow & Dependencies

### Observation Stream

```
  Simulator Engine (tools/simulator/)
           ↓ (gRPC stream ← 6 endpoints)
  ┌────────┴────────┬────────────┬────────────┐
  ↓                 ↓            ↓            ↓
[Radar]        [EW/SIGINT]   [ELINT]       [ISR]
Ingestion      Ingestion     Ingestion     Ingestion
:50051         :50052        :50053        :50054
  ↓                 ↓            ↓            ↓
  └────────┬────────┴────────────┴─────┬─────┘
           ↓ (Protobuf records)         ↓
      [Redpanda Topics]           [Async Aggregation]
      (rtsa.sensor.radar, etc.)
           ↓
  ┌────────┴────────┐
  ↓                 ↓
[Track Fusion]  [Anomaly Detection]
(svc-fusion)    (svc-anomaly-detection)
  ↓                 ↓
  └────────┬────────┘
           ↓ (Fused Tracks + Alerts)
      [Redpanda Topics]
      (rtsa.tracks.fused, rtsa.alerts)
           ↓
  ┌────────┴───────────────────────┐
  ↓                                 ↓
[ClickHouse]                   [WebTransport]
(batch ETL)                    (real-time COP)
  ↓                                 ↓
[Analytics/Forensics]         [Browser: SolidJS/WebGPU]
```

---

## Demo Workflow

### Quick Start: Run a Demo Scenario

```bash
cd /home/arvind/workspace/rtsa_webgpu

# Start 31-container stack
docker compose -f deploy/docker-compose.yml \
               -f deploy/docker-compose.services.yml \
               up -d

# Wait for Redpanda to be ready
docker compose -f deploy/docker-compose.yml \
               -f deploy/docker-compose.services.yml \
               exec redpanda /opt/redpanda/bin/rpk cluster info

# Run multi-domain demo (25 min, all sensor types)
bash scripts/demo/run-demo.sh multi-domain --seed --stop-on-complete

# OR run sensor-health demo (for Playwright E2E tests)
bash scripts/demo/run-demo.sh sensor-health --seed --stop-on-complete
```

### Demo Script Structure

**Location**: `scripts/demo/`

| Script              | Purpose                                                                       |
| ------------------- | ----------------------------------------------------------------------------- |
| `_common.sh`        | Shared utilities: docker compose wrapper, demo config parser, logging         |
| `run-demo.sh`       | Main entry point; parses scenario name, calls seed data + simulator + sensors |
| `seed-demo-data.sh` | ClickHouse seeding (historical data)                                          |
| `init-topics.sh`    | Redpanda topic initialization                                                 |
| `stop-demo.sh`      | Cleanup: stops containers, optionally clears volumes                          |

### Typical Demo Invocation

```bash
# 1. Download & initialize satellite imagery
bash scripts/demo/init-map-layer.sh

# 2. Load ClickHouse seed data (48-hour forensic history)
bash scripts/demo/seed-demo-data.sh

# 3. Launch simulator container with scenario
docker compose ... run --rm simulator \
  --scenario /app/scenarios/multi-domain-demo.yaml \
  --seed 12345

# 4. (Optional) Inject feedback samples for training
docker compose ... run --rm simulator \
  --scenario /app/scenarios/anomaly-demo.yaml \
  --seed 67890 \
  --inject-feedback
```

### Browser Validation

After demo starts, visit **http://localhost:5173**:

- **Health Dashboard**: Should show 6 rows (one per ingestion service)
- **Map View**: Displays real-time tracks colored by track type
- **Anomaly Dashboard**: Shows triggered alerts from Anomaly Detection service
- **Forensic Query**: ClickHouse historical data queries on Analytics tab

---

## Test Automation Integration

### Unit Tests

```bash
cd /home/arvind/workspace/rtsa_webgpu/tools/simulator
go test -v -cover ./...
```

**Coverage**: Entity manager, movement generators, anomaly injection, YAML parser

### Integration Tests

```bash
cd /home/arvind/workspace/rtsa_webgpu
make test-integration
```

**Flow**:

1. Spin up 31-container stack
2. Inject simulator stream for 5 minutes
3. Query ClickHouse/Redpanda to verify:
   - Observation counts match expected entity × update frequency
   - Anomaly injection rate ≈ configured
   - Track fusion produces merged output
   - Alerts triggered for anomaly types

### E2E Tests (Playwright)

```bash
cd /home/arvind/workspace/rtsa_webgpu/web-cop-gpu
BASE_URL=http://localhost:5173 \
npx playwright test e2e/sensor-health.spec.ts \
  --project=chromium \
  --reporter=list
```

**Test Cases**:

1. ✓ **Sensor Health Dashboard visible** — DOM renders 6 service health rows
2. ✓ **Status filtering** — Click "Online" filter; verify only 6 row filtered
3. ✓ **Sidebar collapse/expand** — UI responsiveness check
4. ✓ **Dashboard switching** — Health ↔ Map views work
5. ✓ **Screenshot capture** — Proof-of-work: `e2e/snapshots/sensor-health-dashboard.png` (163 KB typical)

**What's Tested**:

- CSP header allows gRPC-Web to `http://localhost:8443` (dev mode)
- Multi-service fan-out: 6 parallel `ListSensorStatuses` calls with headers
- Envoy routing via `x-ingestion-target` header
- UI rendering of aggregated sensor status

### Performance/Benchmark Tests

```bash
cd /home/arvind/workspace/rtsa_webgpu/tests
make benchmark
# Runs stress.yaml; measures:
# - Redpanda throughput (msg/sec)
# - ClickHouse ETL latency (ms)
# - WebGPU frame time (16 ms target @ 60 FPS)
# - Main thread CPU utilization (< 20% target)
```

---

## Static Seed Data Strategy

### ClickHouse Seeding

**File**: `scripts/demo/seed-demo-data.sh`

Creates static forensic dataset for analytics queries without live simulator.

#### Table: `sensor_observations`

Schema: sensor_type, observation_timestamp, entity_id, latitude, longitude, velocity, signal_strength, ...

**Seeded Data**:

- 11 representative observations across all 6 sensor types
- Timestamps: ISO 8601 with 48-hour window
- Coordinates: Mid-Atlantic AOR (43–47°N, 55–65°W)

```sql
-- Example record: Radar detection
INSERT INTO sensor_observations VALUES (
  'RADAR',
  toDateTime('2026-03-11 14:30:00'),
  'SHIP-001',
  44.250,
  -59.830,
  12.5,  -- knots
  85,    -- dBm
  'low_confidence'
);
```

#### Table: `tracks_fused`

Schema: track_id, source_tracks, fusion_confidence, heading, speed, position_history, ...

**Seeded Data**: 10 fused tracks with multi-sensor correlation history

**Key Track**: `TRK-0002` (MMSI 123456789 — fictional)

- Simulates suspicious maritime vessel
- Movement: Erratic course changes hours 20–24 of observation window
- Sensor sources: Radar, AIS, ISR (demonstrating multi-modal fusion)
- Forensic analysis example for analyst dashboard

#### Table: `anomaly_detections`

Schema: alert_id, anomaly_type, detected_timestamp, affected_track_id, severity, confidence, ...

**Seeded Data**: 10 alerts across 5 anomaly types

- Speed anomalies: 2 alerts
- Route deviations: 3 alerts
- AIS manipulation: 2 alerts
- Behavioral: 2 alerts
- Proximity: 1 alert

#### Table: `operator_feedback`

Schema: feedback_id, alert_id, operator_assessment, corrected_label, timestamp, ...

**Seeded Data**: 5 feedback records

- Operator corrections to classifier
- Ground truth for retraining pipeline
- ForeignKey references to `anomaly_detections`

#### Table: `audit_log`

Schema: event_id, event_type, source_service, affected_entity, timestamp, details, ...

**Seeded Data**: 12 audit entries

- Service startup/shutdown events
- Data ingestion milestones
- Alert firing and feedback submission
- Track fusion updates

#### Table: `nato_allied_tracks`

Schema: nato_track_id, callsign, nation_source, track_data_json, received_timestamp, ...

**Seeded Data**: 5 NATO STANAG 5516 tracks

- Multi-nation data exchange example
- Link to domestic tracks (cross-reference exercise)
- NATO adapter (`svc-nato-adapter`) validation

---

## Reproducibility & Determinism

### Seed-Based RNG

All scenarios use deterministic seed-based random number generation:

```bash
# Reproducible run: same seed → identical observations
go run ./cmd/simulator/ \
  --scenario scenarios/multi-domain-demo.yaml \
  --seed 42

# Different run: different seed → different tracks but same structure
go run ./cmd/simulator/ \
  --scenario scenarios/multi-domain-demo.yaml \
  --seed 99
```

**Implementation**:

- Go's `math/rand` seeded with `--seed` value
- Deterministic entity position updates (same Δt & seed → same position)
- Anomaly injection uses seeded RNG to select which entities get anomalies

### YAML Scenario Versioning

Each scenario file is **immutable** once committed:

- Entity counts defined in YAML (not randomized)
- Movement parameters fixed per scenario
- Anomaly injection points time-based (not stochastic)

**Consequence**: Same scenario + same seed = byte-for-byte identical observations (if running on identical hardware/timing).

### Validation Workflow

1. **Check scenario definition**: `cat tools/simulator/scenarios/<name>.yaml`
2. **Run with known seed**: `--seed 12345`
3. **Capture output**: `docker logs --follow rtsa-simulator | tee simulation-run-12345.log`
4. **Verify observation sequence**: Parse log; extract entity positions
5. **Compare to baseline**: `diff simulation-run-12345.log baseline-run-12345.log` (should match)

---

## Common Tasks

### Task 1: Run Health Check Demo (Quick, 10 min)

```bash
cd /home/arvind/workspace/rtsa_webgpu

# Start system
docker compose -f deploy/docker-compose.yml \
               -f deploy/docker-compose.services.yml up -d

# Run sensor-health scenario (15 min run, auto-stops)
bash scripts/demo/run-demo.sh sensor-health --seed --stop-on-complete

# Verify in browser: http://localhost:5173 (Sensor Health Dashboard)
# Expected: 6 rows (Radar, EW, ELINT, ISR, AIS, Cyber) with CONNECTED status
```

### Task 2: Run Multi-Domain Demo (Full workflow, 30 min)

```bash
# Start system + multi-domain demo (25 min + overhead)
bash scripts/demo/run-demo.sh multi-domain --seed --stop-on-complete

# View Operator UI: http://localhost:5173
# - 52 entities across all domains
# - Real-time track updates
# - 7% anomaly rate

# Query analytics (after demo completes):
docker compose ... exec clickhouse clickhouse-client
  > SELECT count() FROM tracks_fused WHERE created_at > now() - INTERVAL 25 MINUTE;
```

### Task 3: Run Anomaly Detection Showcase (15 min)

```bash
bash scripts/demo/run-demo.sh anomaly-demo --seed --stop-on-complete

# During run, view Anomaly dashboard: http://localhost:5173/anomalies
# - 30% anomaly rate (vs 5% normal)
# - 22 entities with seeded anomalies
# - Watch alerts fire in real-time

# Post-run: query anomaly_detections table
docker compose ... exec clickhouse clickhouse-client
  > SELECT anomaly_type, count() FROM anomaly_detections
      WHERE detected_at > now() - INTERVAL 20 MINUTE
      GROUP BY anomaly_type ORDER BY count DESC;
```

### Task 4: Stress Test (30 min, System Capacity Validation)

```bash
# Prerequisites: 64 GB RAM, 8+ CPU cores recommended
bash scripts/demo/run-demo.sh stress --seed --stop-on-complete

# Monitor during run:
# Terminal 1: Docker resource usage
docker stats --no-stream

# Terminal 2: Redpanda throughput
docker compose ... exec redpanda \
  /opt/redpanda/bin/rpk topic describe rtsa.sensor.radar

# Terminal 3: WebGPU performance (if browser open)
# - Open DevTools → Performance → Record
# - Monitor frame rate (target: 60 FPS @ 50k tracks)
# - Monitor main thread CPU (target: < 20%)
```

### Task 5: Validate Scenario Determinism

```bash
# Run 1: Seed 42
bash scripts/demo/run-demo.sh multi-domain --seed 42 \
  --dry-run > run-1-seed42.log

# Run 2: Same seed
bash scripts/demo/run-demo.sh multi-domain --seed 42 \
  --dry-run > run-2-seed42.log

# Verify outputs match
diff run-1-seed42.log run-2-seed42.log
# Expected: No diff (bitwise identical observations)

# Run 3: Different seed
bash scripts/demo/run-demo.sh multi-domain --seed 99 \
  --dry-run > run-3-seed99.log

# Compare structure
diff <(head -20 run-1-seed42.log) <(head -20 run-3-seed99.log)
# Expected: Same entity count lines but different trajectory values
```

### Task 6: Playwright E2E Validation

```bash
cd /home/arvind/workspace/rtsa_webgpu/web-cop-gpu

# Prerequisite: system running with sensor-health demo
BASE_URL=http://localhost:5173 \
npx playwright test \
  e2e/sensor-health.spec.ts \
  --project=chromium \
  --reporter=html

# View results
npx playwright show-report
# - 5 tests should pass
# - Screenshot: `/snapshots/sensor-health-dashboard.png`
```

---

## Environment Variables

### Simulator Configuration (via Docker Compose)

```bash
# In docker-compose.services.yml or run command:
SIM_RADAR_ENDPOINT=svc-radar-ingestion:50051
SIM_EW_ENDPOINT=svc-ew-ingestion:50052
SIM_ELINT_ENDPOINT=svc-elint-ingestion:50053
SIM_ISR_ENDPOINT=svc-isr-ingestion:50054
SIM_AIS_ENDPOINT=svc-ais-ingestion:50055
SIM_CYBER_ENDPOINT=svc-cyber-ingestion:50056

SIM_SURFACE_ENTITIES=20
SIM_AIR_ENTITIES=10
SIM_SUB_ENTITIES=5
SIM_UPDATE_INTERVAL_MS=1000
SIM_DURATION_MINUTES=0        # 0 = infinite
SIM_ANOMALY_RATE=0.05
SIM_RANDOM_SEED=0             # 0 = random; >0 = deterministic
SIM_SCENARIO_FILE=</app/scenarios/multi-domain-demo.yaml>
```

### Frontend Configuration

```bash
# In docker-compose.services.yml web-cop-gpu build args:
VITE_API_GATEWAY_URL=http://localhost:8443
VITE_MAP_TILE_URL=https://tiles.openstreetmap.org/{z}/{x}/{y}.png
BUILD_ENV=development              # or 'production'
```

---

## References

- **Simulator README**: `tools/simulator/README.md`
- **V1 Architecture**: `docs/architecture/v1/RTSA_WebGPU_Architecture_v1.md`
- **Scenario Files**: `tools/simulator/scenarios/`
- **Demo Scripts**: `scripts/demo/`
- **Test Suite**: `tests/` (Unit, Integration, E2E, Benchmark)
- **ClickHouse Init**: `scripts/demo/seed-demo-data.sh`

---

**Classification**: UNCLASSIFIED
**Last Reviewed**: March 2026
**Maintainer**: RTSA Development Team
