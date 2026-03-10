<!-- CLASSIFICATION: UNCLASSIFIED -->

# RTSA Demo Guide — Setup, Run, and Showcase (v2.0)

> **Classification**: UNCLASSIFIED
> **Audience**: New users, presenters, and evaluators
> **Version**: 2.0 — covers all v2.0 features including the premium dashboard suite, raw sensor
> streaming, unified event timeline, alert assignment, sensor health monitoring, NATO exchange,
> and analyst forensics workflows.
> **Goal**: Run a reliable RTSA demo with minimal setup friction, a clear storytelling flow, and
> complete coverage of all seventeen use cases (UC001–UC017).

---

## 1) What this guide gives you

- A **from-scratch setup path** (manual + automated, sections 2–4)
- **One-command demo launchers** for every demo scenario (section 5)
- A **role-by-role showcase script** with talk tracks, actions, and expected outcomes (section 6)
- **Seed data reference** and scenario descriptions (section 7)
- **Presenter checklists** and troubleshooting shortcuts (sections 8–9)

---

## 2) Prerequisites (manual checks)

Before running any demo, verify these prerequisites:

| Requirement           | Minimum Version | Check Command            |
| --------------------- | --------------- | ------------------------ |
| Docker Engine         | 24.x            | `docker --version`       |
| Docker Compose Plugin | 2.x             | `docker compose version` |
| Go toolchain          | 1.22+           | `go version`             |
| Node.js               | 20 LTS+         | `node --version`         |
| pnpm                  | 9.x             | `pnpm --version`         |
| Rust toolchain        | 1.77+           | `rustc --version`        |
| wasm-pack             | 0.12+           | `wasm-pack --version`    |
| `grpcurl` (optional)  | 1.8+            | `grpcurl --version`      |
| `curl`                | any             | `curl --version`         |

> **Browser requirement**: WebGPU requires **Chrome 113+** or **Edge 113+**. Firefox Nightly
> with `dom.webgpu.enabled` may work but is not officially supported.

Clone the repository if you have not already:

```bash
git clone https://github.com/arvinddhasmana/RTSA_VS_Opus.git
cd RTSA_VS_Opus
```

---

## 3) First-Time Environment Setup

Run the automated setup script once on a fresh machine. This installs required tooling, generates
development mTLS certificates, and validates Docker Compose file syntax:

```bash
bash scripts/setup/setup-dev.sh
```

The script will:

1. Verify all tool prerequisites are present.
2. Generate self-signed mTLS certificates for inter-service communication under `deploy/certs/`.
3. Create the `.env.dev` file from the template with safe local defaults.
4. Validate that all Docker Compose files parse correctly.

---

## 4) Manual Startup Path (Full Control)

Use this path when you need precise control over each step, or when the automated demo script
fails and you need to diagnose the issue.

### 4.1 Start the platform

```bash
# Start infrastructure (Redpanda, ClickHouse, Observability) and all RTSA microservices
docker compose \
  -f deploy/docker-compose.yml \
  -f deploy/docker-compose.services.yml \
  up -d --wait
```

Wait until all containers reach `healthy` status. This takes approximately 45–90 seconds on
first run due to image pulls.

### 4.2 Initialize platform dependencies

Run these once after first platform start, or after a `--volumes` teardown:

```bash
# Create all Redpanda topics (sensors.*, tracks.*, alerts.*, feedback.*, audit.*)
bash scripts/dev/init-topics.sh

# Create ClickHouse database, tables, materialized views, and seed reference data
bash scripts/dev/init-clickhouse.sh
```

The ClickHouse initialization script creates:

- **Core tables**: `tracks_fused`, `anomaly_detections`, `sensor_observations`,
  `operator_feedback`, `audit_log`
- **v2.0 materialized views**: `mv_active_tracks_by_domain` (10-second granularity),
  `mv_sensor_throughput_5min` (rolling 5-min observation rates),
  `mv_alert_ack_latency` (time-to-acknowledge by severity)

### 4.3 Verify platform health

```bash
bash scripts/dev/health-check.sh
```

All services should show `[✓]` before proceeding. Critical services: `redpanda`, `clickhouse`,
`svc-track`, `svc-alert`, `svc-query`, `svc-fusion-engine`, `svc-anomaly-detection`,
`svc-feedback`.

Verify the WebGPU COP is reachable:

```bash
curl -sf http://localhost:5173/ > /dev/null && echo "[✓] web-cop-gpu: http://localhost:5173" \
  || echo "[✗] web-cop-gpu not responding on port 5173"
```

### 4.4 Seed demo data (optional but recommended for UI demos)

Seed representative historical data into ClickHouse so dashboards render with realistic track
histories, timelines, and metrics before any live feed starts:

```bash
bash scripts/demo/seed-demo-data.sh
```

This script inserts:

- 48 hours of simulated fused track history across all 5 domains (Air, Surface, Subsurface,
  Land, Cyber)
- Anomaly detection records including CRITICAL and ELEVATED alerts
- Operator feedback submissions with trust scores
- Audit log entries for track lifecycle and alert acknowledgment
- Sensor observation samples for all 6 sensor types
- NATO exchange records for the NATO Liaison demo

### 4.5 Start a scenario

```bash
bash scripts/demo/run-maritime-demo.sh          # Maritime domain — UC002, UC006, UC008, UC009
bash scripts/demo/run-multi-domain-demo.sh      # All domains — UC002–UC009, UC012
bash scripts/demo/run-fusion-dashboard-demo.sh  # Fusion Dashboard deep-dive — UC016
bash scripts/demo/run-operator-ui-demo.sh       # Operator UI + Timeline — UC010, UC012, UC013
bash scripts/demo/run-sensor-health-demo.sh     # Sensor Health + Coverage — UC017
bash scripts/demo/run-nato-exchange-demo.sh     # NATO outbound/inbound — UC014, UC015
bash scripts/demo/run-analyst-forensics-demo.sh # Forensics + Intel Search — UC013
```

### 4.6 Stop and clean up

```bash
bash scripts/demo/stop-demo.sh --live-feed    # Stop simulators only, keep services running
bash scripts/demo/stop-demo.sh --containers   # Stop all containers
bash scripts/demo/stop-demo.sh --volumes      # Full teardown including Docker volumes
```

---

## 5) Quick-Start Commands (Automated Path)

The `run-demo.sh` entrypoint handles infrastructure startup, initialization, optional seeding,
and scenario launch in a single command.

| Demo Scenario            | Command                                           | Duration | Use Cases Covered                 |
| ------------------------ | ------------------------------------------------- | -------- | --------------------------------- |
| **Maritime** (primary)   | `bash scripts/demo/run-demo.sh maritime`          | 20 min   | UC002, UC006, UC008, UC009, UC010 |
| **Multi-Domain** (broad) | `bash scripts/demo/run-demo.sh multi-domain`      | 30 min   | UC002–UC009, UC012, UC016         |
| **Fusion Dashboard**     | `bash scripts/demo/run-demo.sh fusion-dashboard`  | 15 min   | UC016, UC008, UC012               |
| **Operator UI**          | `bash scripts/demo/run-demo.sh operator-ui`       | 15 min   | UC010, UC012, UC013               |
| **Sensor Health**        | `bash scripts/demo/run-demo.sh sensor-health`     | 10 min   | UC017, UC001                      |
| **NATO Exchange**        | `bash scripts/demo/run-demo.sh nato-exchange`     | 15 min   | UC014, UC015                      |
| **Analyst Forensics**    | `bash scripts/demo/run-demo.sh analyst-forensics` | 20 min   | UC013, UC010, UC011               |
| **Full Suite**           | `bash scripts/demo/run-demo.sh full-suite`        | 60 min   | UC001–UC017                       |

### Common flags

```bash
bash scripts/demo/run-demo.sh maritime --setup            # Run setup-dev.sh first
bash scripts/demo/run-demo.sh maritime --seed             # Seed ClickHouse demo data before scenario
bash scripts/demo/run-demo.sh maritime --dry-run          # Print commands without executing
bash scripts/demo/run-demo.sh maritime --stop-on-complete # Auto-teardown when scenario ends
```

---

## 6) Role-by-Role Showcase Script

This section is written for the human presenter. Each sub-section covers one role and one demo
scenario, with a recommended talk track, specific UI actions, and the expected outcomes.

**Setup before presenting**: Run `bash scripts/demo/run-demo.sh multi-domain --seed` in a
terminal. Open `http://localhost:5173` in **Chrome 113+** or **Edge 113+** (WebGPU required).
Have the terminal visible separately.

---

### 6.1 Scene 1 — System Bootstrap and Health (2 min)

**Role**: Any (no role selection needed)
**Use Cases**: UC001

**Talk track**:

> "RTSA is an event-driven microservices platform. Before any sensor data arrives, the platform
> bootstraps its messaging backbone (Redpanda), its analytics store (ClickHouse), and fifteen Go
> microservices — each communicating exclusively over mTLS-authenticated gRPC channels."

**Actions**:

1. In a terminal: `bash scripts/dev/health-check.sh`
2. Point to each `[✓]` service in the output.
3. Highlight the Redpanda topic families: `sensors.*`, `tracks.*`, `alerts.*`, `feedback.*`.

**Expected output**:

```
[✓] Redpanda broker: localhost:19092 — HEALTHY (12 topics created)
[✓] ClickHouse: localhost:8123 — HEALTHY (rtsa database, 5 tables, 3 materialized views)
[✓] svc-radar-ingestion — HEALTHY
[✓] svc-ew-ingestion — HEALTHY
[✓] svc-fusion-engine — HEALTHY
[✓] svc-anomaly-detection — HEALTHY
[✓] svc-track — HEALTHY (dual consumer groups: track-service, track-svc-sensor-stream)
[✓] svc-alert — HEALTHY
[✓] svc-query — HEALTHY
[✓] svc-feedback — HEALTHY
[✓] svc-audit — HEALTHY
```

---

### 6.2 Scene 2 — Data Ingestion Across All Sensor Domains (3 min)

**Role**: Sensor Operator → **Sensor Health** dashboard (auto-selected on role login)
**Use Cases**: UC002, UC003, UC004, UC005, UC006, UC007, UC017

**Talk track**:

> "RTSA ingests from six sensor domains simultaneously. As Sensor Operator, I have a dedicated
> dashboard showing which sensors are active, how fast they are producing, and what their
> geographic coverage looks like. I can spot a degraded sensor before it creates a gap in the
> fused track picture."

**Actions**:

1. Open `http://localhost:5173`. From the Role Selector dropdown, choose **Sensor Operator**.
2. The **Sensor Health** dashboard loads automatically.
3. Walk through the left sensor status panel: `sensor_id`, type icon, connection status
   (green dot = connected), events/sec, acceptance rate, and last observation timestamp.
4. Click a radar sensor card to centre the map on its position.
5. Point out the fan-sector coverage overlay on the map (translucent blue arc).
6. In a terminal: `bash scripts/demo/run-sensor-health-demo.sh`
7. Watch one sensor card turn amber/red as events/sec degrades.

**Expected outcome**:

- Sensor cards update every 10 seconds via `IngestionService.ListSensorStatuses` bulk RPC
  (new in v2.0).
- Coverage overlays show Radar fan sectors (blue), EW arcs (amber), ISR polygons (green) —
  populated from the new `SensorCoverage` geometry fields on `SensorStatusResponse`.
- The degraded sensor shows 0 events/sec and its coverage arc disappears from the map.

---

### 6.3 Scene 3 — Multi-Source Fusion (2 min)

**Role**: Operations Commander → **Fusion** dashboard (default Level-2 view)
**Use Cases**: UC008, UC016

**Talk track**:

> "Now switching to Operations Commander. The Fusion Dashboard is the default view for this role.
> This makes the multi-source fusion algorithm transparent — you can see both raw pre-fusion
> sensor observations AND correlated fused tracks on the same map simultaneously."

**Actions**:

1. Switch role to **Operations Commander**.
2. The **Fusion Dashboard** loads as the default Level-2 view automatically.
3. On the map, identify the distinct icon types:
   - Fused tracks (filled circles): coloured by hostile class (red/amber/blue/green).
   - Raw radar plots (light-blue diamonds): pre-fusion observations from `StreamSensorObservations`.
   - EW/SIGINT detections (amber triangles): signal intercepts.
   - AIS positions (green circles): maritime transponder positions.
   - ELINT detections (purple squares): emitter observations.
4. Point to the **Fusion Side Panel** on the right:
   - Active tracks by domain with colour-coded dots.
   - Average confidence score.
   - Confidence histogram: High (>=0.8) green, Medium (0.6–0.79) blue, Low amber, Tentative red.
   - Scrollable active track list sorted by confidence — click any track to open detail panel.
5. Collapse the Fusion Side Panel using the chevron (triangle) button. Watch it animate closed.

**Expected outcome**:

- Raw sensor icons arrive via `StreamSensorObservations` gRPC server-streaming — new v2.0 RPC.
- The `track-svc-sensor-stream` consumer group (new in v2.0) reads from all `sensors.*` topics.
- Side panel live-updates domain counts and confidence histograms from the same stream.

---

### 6.4 Scene 4 — Anomaly Detection and Alert Workflow (3 min)

**Role**: Operations Commander → **Operator UI** dashboard (Level-2 tab switch)
**Use Cases**: UC009, UC010, UC012

**Talk track**:

> "The Operator UI dashboard is the tactical command view. The map is in the background with a
> blur effect. Critical alerts and a chronological event timeline are on glass-morphism panels
> in the foreground. When the AI flags a suspicious track, the operator has four choices:
> Inspect, Confirm, Reject, or Assign."

**Actions**:

1. Click the **Operator** tab in the Level-2 dashboard selector (toolbar).
2. Note the map is blurred in background with frosted-glass panels overlaid.
3. In the Alert Panel (left): identify CRITICAL (red-pulsing) and ELEVATED (amber) alerts.
4. On a CRITICAL alert, point out the four buttons: `[Inspect]` `[Confirm]` `[Reject]` `[Assign]`.
5. Click **[Inspect]** — Entity Detail Panel opens with:
   - Track identity, position, speed, heading.
   - **Source Attribution**: per-sensor confidence breakdown table (new in v2.0).
   - **Entity Timeline**: interleaved track state changes, anomaly events, and feedback from
     `GetEventTimeline` RPC — new in v2.0, queries 4 ClickHouse tables via UNION ALL.
   - **Feedback Form**: radio buttons for type, justification text area.
6. Click **[Confirm]** on a second alert. Success toast appears; alert marked with checkmark.
   This calls `FeedbackService.SubmitFeedback` with `CONFIRM_ANOMALY`.
7. Click **[Reject]** on a third alert. This calls `SubmitFeedback` with `REJECT_ANOMALY`.
8. Click **[Assign]** on a fourth alert. Enter operator ID `OP-007` in the popover.
   This calls `AlertService.AssignAlert` — new v2.0 RPC — and produces an audit event.
9. Show the **Event Timeline** on the right — all events in chronological order with correlation
   markers where multiple sources reference the same entity within 60 seconds.

**Expected outcome**:

- `[Confirm]` and `[Reject]` trigger `FeedbackService.SubmitFeedback` gRPC.
- `[Assign]` triggers `AlertService.AssignAlert` (new v2.0) and writes an audit event to
  `audit.events` Redpanda topic, which Redpanda Connect ETLs to ClickHouse `audit_log`.
- Entity Timeline via `GetEventTimeline` shows events from `tracks_fused`,
  `anomaly_detections`, `operator_feedback`, and `audit_log` — all ordered by `event_time`.

---

### 6.5 Scene 5 — Multi-Domain Common Operating Picture (3 min)

**Role**: Operations Commander → **Multi-Domain** dashboard (Level-2 tab switch)
**Use Cases**: UC012, UC008, UC009

**Talk track**:

> "For a wide-area operational picture across all five domains simultaneously — Air, Surface,
> Subsurface, Land, and Cyber — we use the Multi-Domain Dashboard. It is deliberately
> maximised: all side panels collapsed, the map fills every pixel, with domain KPIs floating
> as a minimal overlay."

**Actions**:

1. Click the **Multi-Domain** tab in the Level-2 dashboard selector.
2. All panels collapse. The map fills the screen.
3. Point to the **Domain Metrics Overlay** (floating, top-left of map):
   - Air, Surface, Subsurface, Land, Cyber cards — each with track count, observation rate
     (obs/sec), and hostile count.
4. Point out the **sensor coverage overlays** on the map:
   - Blue fan sectors = radar coverage (from `SensorCoverage` geometry via `ListSensorStatuses`).
   - Amber circles = EW/SIGINT arcs.
   - Green polygons = ISR swath coverage.
5. Click the **Layers** button (bottom-right of map). Toggle Sensor Coverage off and on.
6. Click the collapsed Alert strip at the bottom to expand the full alert panel.

**Expected outcome**:

- Domain counts update from both `StreamTracks` (fused) and `StreamSensorObservations` (raw).
- Coverage overlays rendered from `ListSensorStatuses` `SensorCoverage` geometry — new v2.0.
- Layer toggle wires to UI visibility signals controlling WebGPU render-layer visibility.

---

### 6.6 Scene 6 — Intelligence Analyst Forensics (3 min)

**Role**: Intelligence Analyst → **Forensics** (auto-selected), then **Intel Search**
**Use Cases**: UC013, UC010, UC011

**Talk track**:

> "The Intelligence Analyst role provides historical investigation tools. Forensics is the
> default Level-2 view — the analyst queries ClickHouse for historical tracks, anomaly
> detections, and sensor observations with time-range and spatial filters."

**Actions**:

1. Switch role to **Intelligence Analyst**.
2. The **Forensics** panel opens on the right alongside the map.
3. Set time range to last 4 hours. Filter by entity type: **Surface**.
   Map updates via `QueryService.QueryTracks`.
4. Click **Intel Search** tab from Level-2 selector.
5. Type vessel MMSI `123456789` (seeded). Search results show all correlated observations.
6. Click a result — Entity Detail Panel opens with the full 72-hour event timeline via
   `GetEventTimeline`.
7. In the Feedback Form, select "Confirm Hostile" and submit. The returned trust score displays.

**Expected outcome**:

- Forensics queries hit `QueryService.QueryTracks` and `QueryService.QueryAnomalies` with
  ClickHouse parameterized SQL and classification filtering.
- The event timeline for MMSI `123456789` shows: detection, pattern deviation, ELINT intercept,
  operator feedback — all interleaved chronologically.

---

### 6.7 Scene 7 — Security Officer Audit and Oversight (2 min)

**Role**: Security Officer → **Audit and Feedback** dashboard (auto-selected)
**Use Cases**: UC010, UC011, UC013

**Talk track**:

> "Every operator action generates an immutable audit event via the Audit Service. The Security
> Officer sees all submissions, trust scores, and anti-poisoning guard decisions — the
> accountability layer that makes human-in-the-loop feedback trustworthy."

**Actions**:

1. Switch role to **Security Officer**.
2. Point out recent feedback entries from Scene 4 (Confirm, Reject, Assign actions).
3. Each entry shows: operator ID, feedback type, trust score, anti-poisoning decision.
4. Show the `AssignAlert` audit entry: `assigner_id`, `assignee_id`, `alert_id`, timestamp.
5. Note the trust score distribution — higher clearance and accuracy → higher trust weight.

**Expected outcome**:

- All audit entries are immutable in ClickHouse `audit_log` (append-only table).
- Anti-poisoning guard logs show rate-limit and bulk-anomaly detection decisions.

---

### 6.8 Scene 8 — NATO Data Exchange (2 min)

**Role**: NATO Liaison → **NATO Exchange** dashboard (auto-selected)
**Use Cases**: UC014, UC015

**Talk track**:

> "The NATO Liaison role manages interoperability with allied systems via Link 16, NFFI, and
> MIP formats. The NATO Exchange dashboard shows link connectivity, outbound track nominations,
> and inbound allied tracks — all with classification guard enforcement."

**Actions**:

1. Switch role to **NATO Liaison**.
2. NATO Exchange dashboard loads: Link 16 and NFFI connectivity header with green dots.
3. Track Nomination Queue shows tracks awaiting outbound review.
4. Click **[Nominate]** on a surface track — NATO overlay icon appears on the map.
5. Point out a track marked "BLOCKED: classification ceiling exceeded" by the cross-domain guard.
6. Show inbound NATO tracks on map (NATO icon, REL TO label).
7. Click **[Revoke]** on the nominated track to demonstrate bidirectional control.

**Expected outcome**:

- Nomination actions produce audit events with `resource_type = 'nato_nomination'`.
- Classification guard blocks tracks rated above the configured release level.

---

### 6.9 Scene 9 — Close (1 min)

**Talk track**:

> "In ten minutes we demonstrated six sensor domains, AI-driven anomaly detection, multi-source
> fusion with visible raw observations, five purpose-built operator role dashboards, a unified
> four-source event timeline, immutable audit trail, and NATO interoperability — all from a
> single docker compose command."

**Actions**:

1. Run: `bash scripts/demo/stop-demo.sh --volumes`
2. Show the `docker compose down -v` output — clean shutdown.

---

## 7) Scenario and Seed Data Reference

### 7.1 Maritime Demo Scenario

Simulates a 20-minute maritime patrol in the North Atlantic (approx. 60°N, 10°W):

| Event                             | Time Offset | Details                                                           |
| --------------------------------- | ----------- | ----------------------------------------------------------------- |
| Radar detection of USV            | T+0:00      | Unidentified surface vehicle, 12 knots, heading 180 degrees       |
| AIS signal near USV               | T+1:30      | MMSI `987654321`, vessel name `UNKNOWN-AIS` — possible spoofing   |
| EW intercept — encrypted comms    | T+3:00      | Signal on 8.5 GHz, high-power burst, correlated to USV position   |
| Fusion creates TRK-0001           | T+3:15      | Confidence 0.82 — four sensors correlated                         |
| Speed anomaly detected            | T+5:00      | USV acceleration to 28 knots — anomaly score 0.91, CRITICAL alert |
| AIS position delta anomaly        | T+6:30      | AIS position diverges 1.2 NM from radar — AIS_MANIPULATION alert  |
| Second vessel detected (friendly) | T+8:00      | MMSI `123456789`, NATO-tracked, TRK-0002                          |
| Track merge event                 | T+10:00     | TRK-0001 and radar ghost merged (confidence 0.87)                 |
| Operator confirms anomaly         | T+12:00     | Feedback: CONFIRM_ANOMALY, trust score 0.85                       |
| Vessel exits bounding box         | T+18:00     | Track status changes to STALE                                     |

**Use Cases**: UC002, UC006, UC003, UC008, UC009, UC010.

### 7.2 Multi-Domain Demo Scenario

Simulates a 30-minute combined-arms exercise across all five domains:

| Domain     | Entities  | Scenario Details                                                                               |
| ---------- | --------- | ---------------------------------------------------------------------------------------------- |
| Air        | 8 tracks  | 4 friendly aircraft (F-35 signatures), 2 unknown UAVs, 2 suspect cruise-missile-profile tracks |
| Surface    | 12 tracks | 6 allied vessels, 3 merchant ships, 2 unknown fast-movers, 1 potentially spoofed AIS           |
| Subsurface | 2 tracks  | Acoustic contact correlated from 2 ELINT sensors, TENTATIVE classification                     |
| Land       | 5 tracks  | ISR-detected vehicle convoys, one deviating from expected route                                |
| Cyber      | 3 IOCs    | C2 callback domain, lateral movement pattern, data exfiltration indicator                      |

**Use Cases**: UC002–UC009, UC012, UC016.

### 7.3 Fusion Dashboard Scenario

Sensors start sequentially to make the fusion process visible:

| Phase                   | Time | Description                              |
| ----------------------- | ---- | ---------------------------------------- |
| Radar only              | T+0  | Raw plots visible, no fused tracks yet   |
| EW/SIGINT added         | T+2  | Intercepts appear near radar plots       |
| AIS added               | T+4  | Maritime positions near EW contacts      |
| First fused track       | T+6  | Confidence >=0.85, 3-sensor correlation  |
| ISR corroboration       | T+8  | Confidence increases to 0.93             |
| Conflicting observation | T+10 | Confidence drops, tentative flag appears |

**Use Cases**: UC016, UC008, UC012.

### 7.4 Operator UI Scenario

| Phase              | Time | Description                            |
| ------------------ | ---- | -------------------------------------- |
| Pre-seeded alerts  | T+0  | 5 CRITICAL + 10 ELEVATED alerts loaded |
| New CRITICAL alert | T+2  | USV speed anomaly                      |
| New ELEVATED alert | T+5  | Route deviation                        |
| Track merge event  | T+6  | Timeline shows merge marker            |
| New CRITICAL alert | T+9  | AIS spoofing detection                 |
| Alert assignment   | T+10 | AssignAlert RPC called, badge visible  |
| Alert closed       | T+13 | Feedback-confirmed alert resolved      |

**Use Cases**: UC010, UC012, UC013.

### 7.5 Sensor Health Scenario

| Phase                     | Time | Description                              |
| ------------------------- | ---- | ---------------------------------------- |
| All sensors healthy       | T+0  | Green status, nominal events/sec         |
| RADAR-NORTH-01 degrades   | T+2  | Events/sec drops 80%, card turns amber   |
| EW-STATION-03 disconnects | T+5  | Card turns red, coverage arc disappears  |
| Sensor degradation alert  | T+6  | Alert appears in alert panel             |
| Both sensors recover      | T+8  | Cards return to green, coverage restored |

**Use Cases**: UC017, UC001.

### 7.6 NATO Exchange Scenario

| Phase                  | Time | Description                             |
| ---------------------- | ---- | --------------------------------------- |
| Pre-seeded NATO tracks | T+0  | 5 inbound allied tracks (REL TO FVEY)   |
| Outbound nominations   | T+0  | 3 organic tracks nominated              |
| Partner acknowledgment | T+3  | Nominated tracks show "Acknowledged"    |
| New inbound track      | T+5  | NATO icon appears on map                |
| Classification block   | T+8  | Cross-domain guard rejects one outbound |
| Track revocation       | T+12 | Nominated track recalled                |

**Use Cases**: UC014, UC015.

### 7.7 Analyst Forensics Scenario

| Phase                 | Time | Description                                    |
| --------------------- | ---- | ---------------------------------------------- |
| Time-range query      | T+0  | Last 48 hours, surface domain, MMSI 123456789  |
| Entity timeline opens | T+2  | 4-hour anomalous activity window in history    |
| Intel Search          | T+5  | Search by callsign VIPER-01 — air track found  |
| Feedback submission   | T+8  | Trust score returned and displayed             |
| Model retrain log     | T+12 | Batch threshold reached, retrain shown (UC011) |

**Use Cases**: UC013, UC010, UC011.

### 7.8 Full Suite Scenario

Runs all scenarios in sequence with 5-minute transition periods.
Covers UC001–UC017. Duration: approximately 60 minutes.

### 7.9 Seed Data Summary

**Sensors registered (12 total)**:

| Sensor ID        | Type         | Position       | Coverage                 |
| ---------------- | ------------ | -------------- | ------------------------ |
| `RADAR-NORTH-01` | RADAR        | 60.5°N, 8.2°W  | 150 NM, sector 000°–120° |
| `RADAR-SOUTH-02` | RADAR        | 55.3°N, 12.1°W | 120 NM, sector 090°–270° |
| `EW-STATION-01`  | EW_SIGINT    | 59.1°N, 9.8°W  | 200 NM omnidirectional   |
| `EW-STATION-03`  | EW_SIGINT    | 56.7°N, 11.3°W | 180 NM omnidirectional   |
| `ELINT-ARRAY-01` | ELINT_COMINT | 58.4°N, 7.6°W  | 300 NM omnidirectional   |
| `ISR-UAV-ALPHA`  | ISR          | Variable       | 50x50 NM swath polygon   |
| `ISR-SAT-BRAVO`  | ISR          | Orbital        | 200x200 NM swath polygon |
| `AIS-COAST-01`   | AIS_BFT      | 61.0°N, 10.5°W | VHF range, 40 NM         |
| `AIS-COAST-02`   | AIS_BFT      | 54.9°N, 13.0°W | VHF range, 40 NM         |
| `CYBER-FEED-01`  | CYBER        | N/A            | N/A                      |
| `NATO-LINK16-01` | AIS_BFT      | 58.0°N, 6.0°W  | NATO Link 16 terminal    |
| `BFT-GROUND-01`  | AIS_BFT      | 57.5°N, 9.0°W  | Blue Force Tracker       |

**Pre-seeded tracks (10 total)**:

| Track ID   | Domain     | Classification | Confidence | Notes                                   |
| ---------- | ---------- | -------------- | ---------- | --------------------------------------- |
| `TRK-0001` | Surface    | Suspect        | 0.82       | 4-sensor fusion, speed anomaly CRITICAL |
| `TRK-0002` | Surface    | Friendly       | 0.95       | MMSI `123456789`, NATO-tracked          |
| `TRK-0003` | Air        | Unknown        | 0.67       | 2-sensor fusion, tentative              |
| `TRK-0004` | Air        | Friendly       | 0.98       | Callsign `VIPER-01`, NATO Link 16       |
| `TRK-0005` | Subsurface | Tentative      | 0.51       | Acoustic correlation only               |
| `TRK-0006` | Land       | Unknown        | 0.73       | ISR vehicle convoy                      |
| `TRK-0007` | Cyber      | N/A            | 0.88       | IOC: `malicious-c2.example.com`         |
| `TRK-0008` | Surface    | Hostile        | 0.91       | Speed anomaly CRITICAL, active alert    |
| `TRK-0009` | Air        | Suspect        | 0.76       | Route deviation ELEVATED, active alert  |
| `TRK-0010` | Surface    | Unknown        | 0.83       | AIS spoofing CRITICAL, active alert     |

**Pre-seeded active alerts**:

- 3 CRITICAL: TRK-0001, TRK-0008, TRK-0010
- 4 ELEVATED: TRK-0003, TRK-0009, TRK-0005, TRK-0006

**Pre-seeded NATO records**:

- 5 inbound allied tracks with `REL TO FVEY` classification
- 3 organic tracks in outbound nomination queue

---

## 8) Presenter Checklist

### Pre-Demo Checklist (15 min before presenting)

- [ ] Docker daemon is running: `docker info` produces no errors
- [ ] Ports 5173, 9092, 8123, and 8080 are free: `lsof -i :5173`
- [ ] `bash scripts/demo/run-demo.sh multi-domain --seed` completes successfully
- [ ] `http://localhost:5173` loads with RTSA classification banner at top and bottom
- [ ] Browser is **Chrome 113+** or **Edge 113+** (WebGPU required)
- [ ] Role Selector shows all 5 roles: Operations Commander, Intelligence Analyst, Security
      Officer, Sensor Operator, NATO Liaison
- [ ] Operations Commander role shows Fusion, Multi-Domain, and Operator tabs
- [ ] At least one active track is visible on the Fusion Dashboard map
- [ ] Second terminal tab ready with: `bash scripts/demo/stop-demo.sh --volumes` (do not run yet)
- [ ] Browser extensions that block gRPC-Web are disabled
- [ ] Projector zoom set to 110–125% for audience readability

### During-Demo Recovery

| Symptom                          | Fix                                                      |
| -------------------------------- | -------------------------------------------------------- |
| UI shows "Connecting..." forever | `docker logs rtsa-gateway --tail 20` — check mTLS        |
| No tracks on map                 | `docker ps \| grep simulator` — check scenario container |
| Alert panel empty                | `docker logs rtsa-svc-alert --tail 20`                   |
| Sensor Health cards not loading  | `docker logs rtsa-svc-radar-ingestion --tail 20`         |
| Fusion Side Panel shows zero     | `docker logs rtsa-svc-track --tail 30`                   |
| GetEventTimeline returns empty   | `bash scripts/dev/init-clickhouse.sh`                    |
| Browser CORS error               | `docker restart rtsa-gateway`                            |
| Map blank / WebGPU not available | Use Chrome 113+ or Edge 113+; check `chrome://gpu`       |

---

## 9) Troubleshooting Reference

### 9.1 Platform fails to start

```bash
docker compose -f deploy/docker-compose.yml config --quiet
docker compose -f deploy/docker-compose.yml ps --format "table {{.Name}}\t{{.Status}}"
docker compose -f deploy/docker-compose.yml up 2>&1 | grep -E "(ERROR|error|fail|FAIL)"
```

### 9.2 Redpanda topics missing

```bash
docker exec rtsa-redpanda rpk topic list
bash scripts/dev/init-topics.sh
```

### 9.3 ClickHouse schema or v2.0 materialized views missing

```bash
docker exec rtsa-clickhouse clickhouse-client --database rtsa --query "SHOW TABLES"
docker exec rtsa-clickhouse clickhouse-client --database rtsa \
  --query "SELECT name FROM system.tables WHERE name LIKE 'mv_%'"
bash scripts/dev/init-clickhouse.sh
```

### 9.4 Verify v2.0 gRPC RPCs

```bash
# StreamSensorObservations (new v2.0)
grpcurl -plaintext -d '{"clearance_level": 1}' \
  localhost:8080 rtsa.entity.v1.TrackService/StreamSensorObservations

# GetEventTimeline (new v2.0)
grpcurl -plaintext -d '{"track_id": "TRK-0001", "max_events": 10}' \
  localhost:8080 rtsa.query.v1.QueryService/GetEventTimeline

# ListSensorStatuses (new v2.0)
grpcurl -plaintext -d '{"active_within_seconds": 300}' \
  localhost:8080 rtsa.ingestion.v1.IngestionService/ListSensorStatuses

# AssignAlert (new v2.0)
grpcurl -plaintext \
  -d '{"alert_id": "ALT-001", "assigner_operator_id": "OP-001", "assignee_operator_id": "OP-007"}' \
  localhost:8080 rtsa.inference.v1.AlertService/AssignAlert
```

### 9.5 Reset demo to clean state

```bash
bash scripts/demo/stop-demo.sh --volumes
bash scripts/demo/run-demo.sh multi-domain --setup --seed
```

### 9.6 View service logs

```bash
# All services
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml logs -f

# Single service
docker logs rtsa-svc-fusion-engine -f --tail 50

# Filter for specific track
docker logs rtsa-svc-track -f | grep "TRK-0001"

# Check StreamSensorObservations handler
docker logs rtsa-svc-track --tail 200 | grep "StreamSensorObservations"
```

---

## 10) Architecture Reference (Quick-Read for Presenters)

Data flow exercised by all demo scenarios:

```
External Sensors --gRPC--> Ingestion Services --> Redpanda sensors.* topics
                                                          |
                                                          v
                                                    Fusion Engine
                                                    (Kalman filter,
                                                     correlation scoring)
                                                          |
                                          tracks.fused.* Redpanda topics
                                               /                  \
                                    Anomaly Detection           Track Service
                                    (AI/ML inference)           (v2.0: dual consumer groups)
                                         |                           |
                                   alerts.* topics           gRPC-Web streams:
                                         |                   - StreamTracks (fused)
                                   Alert Service             - StreamSensorObservations (raw)
                                   (v2.0: AssignAlert)            |
                                         |                   Browser UI
                                   Redpanda Connect          (5 role views, 8 dashboard types)
                                   (ETL to ClickHouse)
                                         |
                                   ClickHouse OLAP
                                   (GetEventTimeline UNION ALL,
                                    3 new materialized views)
```

**v2.0 additions visible in demo**:

| Feature                      | Data Path                                                                                 | Visible In                                |
| ---------------------------- | ----------------------------------------------------------------------------------------- | ----------------------------------------- |
| `StreamSensorObservations`   | `sensors.*` -> `track-svc-sensor-stream` consumer group -> gRPC-Web                       | Fusion Dashboard raw sensor icons         |
| `GetEventTimeline`           | ClickHouse UNION ALL: `tracks_fused + anomaly_detections + operator_feedback + audit_log` | Operator UI timeline, Entity Detail Panel |
| `SensorCoverage` geometry    | `ListSensorStatuses` bulk RPC -> WebGPU coverage render pass                              | Multi-Domain overlays, Sensor Health map  |
| `AssignAlert`                | `svc-alert` handler -> `audit.events` Redpanda topic                                      | Operator UI, Security Officer audit log   |
| `mv_active_tracks_by_domain` | AggregatingMergeTree, 10-second granularity                                               | Domain Metrics Overlay, Fusion Side Panel |
| `mv_sensor_throughput_5min`  | Rolling 5-min observation rate by sensor_type                                             | Multi-Domain sensor throughput bar        |
| `mv_alert_ack_latency`       | Time-to-acknowledge distribution by severity                                              | Security Officer dashboard metrics        |
| Two-Level RBAC Shell         | `uiStore.activeDashboardView`, `DashboardSelector`, `MainLayout` router                   | All 5 roles, Level-2 tabs                 |
| Glassmorphism design system  | CSS custom properties, `.glass-panel`, NVG theme                                          | All dashboards                            |

---

## 11) Demo Script Version History

| Version | Date       | Changes                                                                                                                                                                                                                                                                                                                                             |
| ------- | ---------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1.0     | 2026-01-15 | Initial maritime and multi-domain scenarios                                                                                                                                                                                                                                                                                                         |
| 2.0     | 2026-02-28 | Added Fusion Dashboard, Operator UI, Sensor Health, NATO Exchange, and Analyst Forensics scenarios; added `seed-demo-data.sh`; added all new role-specific demo scripts; updated presenter script for all 5 operator roles and 8 demo scenarios; added v2.0 RPC verification commands; expanded troubleshooting; added complete seed data reference |
