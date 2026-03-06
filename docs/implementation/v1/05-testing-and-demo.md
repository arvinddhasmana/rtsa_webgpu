<!-- CLASSIFICATION: UNCLASSIFIED -->

# V1 — Testing, Demo Scenarios & Browser E2E

> **Version**: 1.0
> **Category**: Testing Completeness, Demo Readiness, Browser E2E
> **Priority**: P2 — Required for v1 demo and acceptance
> **Depends On**: `01-infrastructure-fixes.md`, `02-ingestion-service-completion.md`
> **Agent**: `@greatest-ever-developer`

---

## Purpose

The v1 milestone must demonstrate a fully working end-to-end RTSA pipeline — from
simulated sensor feeds through ingestion, fusion, anomaly detection, alerting, and
operator interaction in the COP browser UI. This file covers:

1. **Browser E2E tests** with Playwright for `web-cop/`
2. **Two purpose-built demo scenarios** (maritime primary, multi-domain secondary)
3. **Demo launch / stop scripts** for one-command execution
4. **Negative E2E tests** validating error paths (DLQ, anti-poisoning, classification)
5. **Test data completeness** across all use cases

### Traceability

| Requirement                 | Feature           | Use Case      | What This File Covers                            |
| --------------------------- | ----------------- | ------------- | ------------------------------------------------ |
| CR-UI-001 … CR-UI-006       | FEAT-09, FEAT-10  | UC006, UC008  | Browser E2E: map, alerts, detail panel           |
| CR-UI-007                   | FEAT-10           | UC009         | Browser E2E: forensics query                     |
| CR-FB-001, CR-FB-002        | FEAT-11           | UC010         | Browser E2E: feedback form submission            |
| CR-SEC-001 … CR-SEC-004     | FEAT-13           | UC012         | Browser E2E: classification banner, offline mode |
| NFR-PERF-001 … NFR-PERF-004 | —                 | UC013         | Demo scenario load validation                    |
| CR-ING-001 … CR-ING-006     | FEAT-01 … FEAT-03 | UC001 … UC003 | Demo: multi-sensor data flow                     |
| CR-FUS-001 … CR-FUS-004     | FEAT-04           | UC004         | Demo: track fusion visible on map                |
| CR-INF-001 … CR-INF-005     | FEAT-05, FEAT-06  | UC005, UC007  | Demo: anomaly alerts appear in UI                |

---

## Task TEST-01 — Playwright Setup for `web-cop/`

### Context

The COP currently has 23 vitest unit tests but zero browser-level tests. All UI
requirements (CR-UI-\*) need at least one browser test exercising the real rendered DOM.
Playwright is the standard choice for SolidJS + Vite projects.

### Steps

1. **Install Playwright**

   ```bash
   cd web-cop
   npm install -D @playwright/test
   npx playwright install --with-deps chromium
   ```

   Only install Chromium — we don't need Firefox/WebKit for v1.

2. **Create `web-cop/playwright.config.ts`**

   ```typescript
   // CLASSIFICATION: UNCLASSIFIED
   import { defineConfig, devices } from "@playwright/test";

   export default defineConfig({
     testDir: "./e2e",
     timeout: 60_000,
     retries: 1,
     reporter: [["html", { open: "never" }], ["list"]],
     use: {
       baseURL: "http://localhost:5173",
       trace: "on-first-retry",
       screenshot: "only-on-failure",
     },
     projects: [
       {
         name: "chromium",
         use: { ...devices["Desktop Chrome"] },
       },
     ],
     webServer: {
       command: "npm run dev",
       url: "http://localhost:5173",
       reuseExistingServer: !process.env.CI,
       timeout: 30_000,
     },
   });
   ```

3. **Create `web-cop/e2e/` directory** for all Playwright tests.

4. **Add npm scripts to `web-cop/package.json`**

   ```json
   "test:e2e": "playwright test",
   "test:e2e:headed": "playwright test --headed",
   "test:e2e:report": "playwright show-report"
   ```

5. **Create `web-cop/e2e/helpers.ts`** — shared utilities:
   - `waitForMapLoad()` — wait for WebGPU canvas to render
   - `waitForGrpcConnection()` — wait for connection indicator to show "connected"
   - `mockGrpcStream()` — intercept gRPC-Web calls and return canned responses
   - `injectTestTrack(page, track)` — inject a track into a SolidJS signal via
     `page.evaluate()`

### Acceptance Criteria

- [ ] `npx playwright test --list` shows available tests
- [ ] Playwright config points to Vite dev server on 5173
- [ ] Helper utilities are importable from all test files
- [ ] No Playwright tests fail due to setup issues (when backend is mocked)

---

## Task TEST-02 — Browser E2E Test Suite (10+ Tests)

### Context

Each browser E2E test should either mock the gRPC backend via route interception or
run against the real Docker Compose stack. For v1, **mock mode** is primary (faster,
CI-friendly); real-backend mode is a BONUS.

### Test List

Create one test file per feature area under `web-cop/e2e/`:

#### `web-cop/e2e/map.spec.ts` — Map Rendering (CR-UI-001, CR-UI-002)

| Test Name                              | Assertion                                            |
| -------------------------------------- | ---------------------------------------------------- |
| `map renders and shows canvas`         | WebGPU canvas element visible, non-zero dimensions   |
| `track appears on map after injection` | Inject 1 track → marker/icon appears at coordinates  |
| `track moves on update`                | Inject track, update position → marker moves         |

#### `web-cop/e2e/alerts.spec.ts` — Alert Display (CR-UI-003, CR-INF-001)

| Test Name                            | Assertion                                                     |
| ------------------------------------ | ------------------------------------------------------------- |
| `alerts panel shows incoming alert`  | Inject alert → alert row visible with severity badge          |
| `alert filtering by severity works`  | Inject HIGH + LOW alerts → filter to HIGH → only HIGH visible |
| `alert acknowledgment updates state` | Click acknowledge → alert row shows "acknowledged" state      |

#### `web-cop/e2e/detail-panel.spec.ts` — Track Detail (CR-UI-004)

| Test Name                           | Assertion                                                                  |
| ----------------------------------- | -------------------------------------------------------------------------- |
| `clicking track opens detail panel` | Click map marker → detail panel slides in with track ID, type, coordinates |
| `detail panel shows anomaly scores` | Inject track with anomaly score > 0 → anomaly section visible              |

#### `web-cop/e2e/feedback.spec.ts` — Operator Feedback (CR-FB-001, CR-FB-002)

| Test Name                                 | Assertion                                          |
| ----------------------------------------- | -------------------------------------------------- |
| `feedback form submits successfully`      | Fill form → submit → success toast or confirmation |
| `feedback form validates required fields` | Submit empty → validation errors shown             |

#### `web-cop/e2e/forensics.spec.ts` — Historical Query (CR-UI-007)

| Test Name                         | Assertion                                           |
| --------------------------------- | --------------------------------------------------- |
| `forensics query returns results` | Enter time range + filter → results table populated |

#### `web-cop/e2e/classification.spec.ts` — Classification & Security (CR-SEC-001)

| Test Name                                   | Assertion                                                     |
| ------------------------------------------- | ------------------------------------------------------------- |
| `classification banner is always visible`   | Banner exists at top of page with correct classification text |
| `classification banner has correct styling` | Banner background colour matches classification level         |

#### `web-cop/e2e/offline.spec.ts` — Offline / Degraded Mode (CR-SEC-004)

| Test Name                                              | Assertion                                                   |
| ------------------------------------------------------ | ----------------------------------------------------------- |
| `connection indicator shows disconnected when offline` | Intercept and block gRPC → indicator turns red/disconnected |
| `cached data remains visible when offline`             | Load tracks → go offline → tracks still rendered on map     |

### Test Pattern

Each `.spec.ts` file should follow this structure:

```typescript
// CLASSIFICATION: UNCLASSIFIED
import { test, expect } from "@playwright/test";
import { waitForMapLoad, injectTestTrack } from "./helpers";

test.describe("Map Rendering", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await waitForMapLoad(page);
  });

  test("map renders and shows canvas", async ({ page }) => {
    const canvas = page.locator("canvas");
    await expect(canvas).toBeVisible();
    const box = await canvas.boundingBox();
    expect(box!.width).toBeGreaterThan(0);
    expect(box!.height).toBeGreaterThan(0);
  });
});
```

### Acceptance Criteria

- [ ] Minimum 10 browser E2E tests implemented
- [ ] All tests pass with mocked gRPC backend: `npm run test:e2e`
- [ ] Tests cover: map, alerts, detail panel, feedback, forensics, classification, offline
- [ ] Test names clearly reference the CR or UC they validate
- [ ] Classification header present in every test file

---

## Task TEST-03 — Maritime Demo Scenario

### Context

The primary demo showcases a realistic Canadian maritime surveillance scenario centred on
the Halifax approaches. It must run for **20 minutes** and exercise the full pipeline:
ingestion → fusion → anomaly detection → alerting → COP display → operator feedback.

### Scenario File

Create `tools/simulator/scenarios/maritime-demo.yaml`:

```yaml
# CLASSIFICATION: UNCLASSIFIED
# Maritime Demo — Primary v1 demo scenario
# Duration: 20 minutes, Halifax approaches
# Exercises: UC001–UC010
name: "Maritime Demo — Halifax Approaches"
description: >
  Primary v1 demo: 30 surface vessels in Halifax approaches including
  commercial shipping, fishing fleet, coast guard patrol, and 3 hostile
  actors performing AIS spoofing and route deviation.
seed: 20240601
duration_minutes: 20

entities:
  surface:
    count: 30
    hostile_ratio: 0.10
    friendly_ratio: 0.40
    neutral_ratio: 0.35
    unknown_ratio: 0.15
    speed_range: [5, 22]
    patterns:
      straight_line: 0.45 # commercial shipping lanes
      patrol: 0.25 # coast guard
      loitering: 0.20 # fishing fleet
      random: 0.10 # hostile evasion

sensors:
  radar:
    count: 2
    sensor_ids: ["RADAR-DEMO-001", "RADAR-DEMO-002"]
    update_interval_ms: 1000
    coverage_nm: 120
  ais:
    count: 1
    sensor_ids: ["AIS-DEMO-001"]
    update_interval_ms: 2000
  ew:
    count: 1
    sensor_ids: ["EW-DEMO-001"]
    update_interval_ms: 3000

anomalies:
  injection_rate: 0.15
  types:
    speed: 0.30
    route_deviation: 0.30
    ais_manipulation: 0.25
    behavioral: 0.15

operational_area:
  center: { lat: 44.65, lon: -63.57 } # Halifax harbour entrance
  radius_nm: 100
  exclusion_zones:
    - name: "Halifax Inner Harbour"
      center: { lat: 44.65, lon: -63.57 }
      radius_nm: 3
```

### Expected Demo Flow

| Minute | What Happens                                        | What the Audience Sees                               |
| ------ | --------------------------------------------------- | ---------------------------------------------------- |
| 0–2    | Simulator starts, 30 vessels appear                 | Map populates with tracks around Halifax             |
| 2–5    | Fusion engine correlates radar + AIS                | Track count consolidates; fused icons replace raw    |
| 5–8    | Anomaly detection triggers on 3 hostile vessels     | Alert panel highlights speed / AIS anomalies         |
| 8–12   | Operator investigates, submits feedback on 2 tracks | Detail panel → feedback form → success confirmation  |
| 12–15  | More anomalies injected; alert queue grows          | Alert priority ordering visible; acknowledgment flow |
| 15–18  | Forensics query for last 15 min of hostile track    | Forensics panel shows historical positions           |
| 18–20  | Wind-down; pipeline metrics shown in Grafana        | Grafana dashboards show throughput, latency          |

### Acceptance Criteria

- [ ] `tools/simulator/cmd/server/main.go` loads `maritime-demo.yaml` without error
- [ ] 30 surface entities generate radar + AIS + EW observations
- [ ] Anomaly injection rate of 15% produces at least 3 alerts in 20 minutes
- [ ] All sensor types produce to correct topics
- [ ] Scenario YAML passes schema validation (matches existing scenario format)

---

## Task TEST-04 — Multi-Domain Demo Scenario

### Context

Secondary demo showcases the full multi-domain capability: surface + air + subsurface +
cyber. Duration: **30 minutes**. Exercises all sensor types and all ingestion services.

### Scenario File

Create `tools/simulator/scenarios/multi-domain-demo.yaml`:

```yaml
# CLASSIFICATION: UNCLASSIFIED
# Multi-Domain Demo — Secondary v1 demo scenario
# Duration: 30 minutes, North Atlantic multi-domain
# Exercises: UC001–UC010 across all sensor domains
name: "Multi-Domain Demo — North Atlantic"
description: >
  Secondary v1 demo: 25 surface + 15 air + 5 subsurface + 10 cyber
  entities across the North Atlantic. Exercises all 6 sensor types
  and all ingestion services.
seed: 20240602
duration_minutes: 30

entities:
  surface:
    count: 25
    hostile_ratio: 0.12
    friendly_ratio: 0.40
    neutral_ratio: 0.32
    unknown_ratio: 0.16
    speed_range: [5, 22]
    patterns:
      straight_line: 0.45
      patrol: 0.25
      loitering: 0.20
      random: 0.10
  air:
    count: 15
    hostile_ratio: 0.20
    friendly_ratio: 0.45
    neutral_ratio: 0.20
    unknown_ratio: 0.15
    speed_range: [150, 550]
    altitude_range: [1000, 15000]
    patterns:
      straight_line: 0.50
      patrol: 0.35
      random: 0.15
  subsurface:
    count: 5
    hostile_ratio: 0.20
    friendly_ratio: 0.40
    neutral_ratio: 0.20
    unknown_ratio: 0.20
    speed_range: [5, 15]
    depth_range: [50, 400]
    patterns:
      straight_line: 0.60
      patrol: 0.40

sensors:
  radar:
    count: 3
    sensor_ids: ["RADAR-MD-001", "RADAR-MD-002", "RADAR-MD-003"]
    update_interval_ms: 1000
    coverage_nm: 150
  ais:
    count: 1
    sensor_ids: ["AIS-MD-001"]
    update_interval_ms: 2000
  ew:
    count: 2
    sensor_ids: ["EW-MD-001", "EW-MD-002"]
    update_interval_ms: 2000
  elint:
    count: 1
    sensor_ids: ["ELINT-MD-001"]
    update_interval_ms: 3000
  isr:
    count: 1
    sensor_ids: ["ISR-MD-001"]
    update_interval_ms: 5000
  cyber:
    count: 1
    sensor_ids: ["CYBER-MD-001"]
    update_interval_ms: 5000
    ioc_rate: 3

anomalies:
  injection_rate: 0.20
  types:
    speed: 0.25
    route_deviation: 0.25
    ais_manipulation: 0.20
    behavioral: 0.20
    proximity: 0.10

operational_area:
  center: { lat: 47.0, lon: -55.0 } # Grand Banks / North Atlantic
  radius_nm: 200
  exclusion_zones:
    - name: "St. John's Harbour"
      center: { lat: 47.56, lon: -52.71 }
      radius_nm: 5
```

### Acceptance Criteria

- [ ] Scenario loads without error
- [ ] All 6 sensor types generate observations
- [ ] 45 entities + 10 cyber IOCs produce across all sensor topics
- [ ] Anomaly injection rate of 20% triggers alerts across multiple domains
- [ ] Air and subsurface entity patterns (altitude, depth) are correctly configured
- [ ] Scenario YAML passes schema validation

---

## Task TEST-05 — Demo Launch / Stop Scripts

### Context

Demo execution must be one-command. Create `scripts/demo/` with launch and stop scripts
that orchestrate Docker Compose + simulator.

### `scripts/demo/run-maritime-demo.sh`

```bash
#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# Launch the maritime demo scenario (20 minutes)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=== RTSA Maritime Demo — Starting infrastructure ==="
cd "$PROJECT_ROOT"
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml up -d --wait

echo "=== Waiting for services to stabilise (30s) ==="
sleep 30

echo "=== Running init scripts ==="
bash scripts/dev/init-topics.sh
bash scripts/dev/init-clickhouse.sh

echo "=== Starting simulator with maritime-demo scenario ==="
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml \
  run --rm simulator \
  /app/simulator --scenario /scenarios/maritime-demo.yaml

echo "=== Maritime demo complete ==="
```

### `scripts/demo/run-multi-domain-demo.sh`

Same structure but uses `multi-domain-demo.yaml`.

### `scripts/demo/stop-demo.sh`

```bash
#!/usr/bin/env bash
# CLASSIFICATION: UNCLASSIFIED
# Stop all RTSA demo services and clean up volumes
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

cd "$PROJECT_ROOT"
echo "=== Stopping all RTSA services ==="
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml down -v
echo "=== Demo stopped and volumes cleaned ==="
```

### Acceptance Criteria

- [ ] `bash scripts/demo/run-maritime-demo.sh` starts infrastructure, runs simulator, exits cleanly
- [ ] `bash scripts/demo/stop-demo.sh` tears down everything including volumes
- [ ] All scripts have `set -euo pipefail` and classification headers
- [ ] Scripts are executable (`chmod +x`)

---

## Task TEST-06 — Negative E2E Tests

### Context

The existing E2E suite (`tests/e2e/`) covers happy-path flows. v1 requires negative
tests validating error handling, DLQ routing, anti-poisoning rejection, and
classification boundary enforcement.

### Test Additions to `tests/e2e/`

#### `tests/e2e/negative_test.go`

```go
// CLASSIFICATION: UNCLASSIFIED
//go:build e2e

package e2e
```

| Test Name                                | What It Validates                                                                              | Key Assertion                                                  |
| ---------------------------------------- | ---------------------------------------------------------------------------------------------- | -------------------------------------------------------------- |
| `TestNeg01_MalformedSensorDLQ`           | Publish invalid protobuf to `sensors.radar.tracks` → message routes to DLQ topic               | DLQ topic `sensors.radar.dlq` receives the message within 30s  |
| `TestNeg02_AntiPoisoningRejectsFeedback` | Submit feedback with trust score 0.0 → `svc-feedback` rejects                                  | No message appears on `feedback.operator.validated`            |
| `TestNeg03_ClassificationViolation`      | Send request with classification level above user clearance → gRPC returns `PERMISSION_DENIED` | gRPC status code is `PERMISSION_DENIED`                        |
| `TestNeg04_OversizedPayloadRejected`     | Send sensor observation > 1MB → ingestion service rejects                                      | gRPC status code is `RESOURCE_EXHAUSTED` or `INVALID_ARGUMENT` |

Each test must:

- Use `skipE2E(t)` guard
- Have `//go:build e2e` build tag
- Timeout within 2 minutes
- Clean up any test messages after assertion

### Acceptance Criteria

- [ ] At least 4 negative E2E tests implemented
- [ ] All negative tests pass against running stack
- [ ] Tests validate DLQ, anti-poisoning, classification, and payload limits
- [ ] Classification header present in test file

---

## Task TEST-07 — Test Data Completeness Audit

### Context

Ensure that every use case (UC001–UC015) has at least one test path exercising it. This
task is an audit + gap-fill — not necessarily new test files, but ensuring traceability.

### Use Case Coverage Matrix

| Use Case                       | Unit Test                   | Integration Test | E2E Test | Browser E2E                     | Demo Scenario   |
| ------------------------------ | --------------------------- | ---------------- | -------- | ------------------------------- | --------------- |
| UC001 — Sensor Ingestion       | svc-radar-ingestion tests   | IT01, IT02       | E2E01    | `map.spec.ts`                   | Both            |
| UC002 — Multi-Sensor Fusion    | svc-fusion-engine tests     | IT04, IT05       | E2E01    | `map.spec.ts`                   | Both            |
| UC003 — AIS Correlation        | svc-ais-ingestion tests     | IT02             | E2E01    | `map.spec.ts`                   | Both            |
| UC004 — Track Management       | svc-track tests             | IT05             | E2E01    | `detail-panel.spec.ts`          | Both            |
| UC005 — Anomaly Detection      | svc-anomaly-detection tests | IT06, IT07       | E2E01    | `alerts.spec.ts`                | Both            |
| UC006 — COP Display            | web-cop vitest              | —                | —        | `map.spec.ts`, `alerts.spec.ts` | Both            |
| UC007 — Alert Management       | svc-alert tests             | IT08             | E2E02    | `alerts.spec.ts`                | Both            |
| UC008 — Track Detail           | web-cop vitest              | —                | —        | `detail-panel.spec.ts`          | Both            |
| UC009 — Forensic Query         | svc-query tests             | IT09             | E2E03    | `forensics.spec.ts`             | Both            |
| UC010 — Operator Feedback      | svc-feedback tests          | IT10             | E2E02    | `feedback.spec.ts`              | Both            |
| UC011 — NATO Export            | svc-nato-adapter tests      | —                | —        | —                               | —               |
| UC012 — Security Enforcement   | pkg/classification tests    | IT11             | —        | `classification.spec.ts`        | Both            |
| UC013 — Performance Under Load | benchmark tests             | —                | —        | —                               | Stress scenario |
| UC014 — Anti-Poisoning         | svc-feedback tests          | IT10             | Neg02    | —                               | Both            |
| UC015 — Model Retraining       | svc-training tests          | —                | —        | —                               | —               |

### Gap Analysis

The following gaps must be filled during implementation:

1. **UC006 / UC008**: No integration-level test that starts the COP against a real
   backend. The browser E2E tests (TEST-02) will close this gap.
2. **UC011**: NATO adapter is noop — unit test only is sufficient for v1.
3. **UC015**: Training pipeline is noop — unit test only is sufficient for v1.
4. **UC013**: Existing `tests/benchmark/` tests cover throughput. The stress scenario
   (`tools/simulator/scenarios/stress.yaml`) exercises sustained load. Confirm the
   benchmark thresholds match NFR requirements:
   - NFR-PERF-001: Ingestion latency < 100ms p99
   - NFR-PERF-002: Fusion cycle < 500ms
   - NFR-PERF-003: Alert generation < 2s from anomaly detection
   - NFR-PERF-004: Query response < 3s for 24h window

### Acceptance Criteria

- [ ] Every UC has at least one test type covering it (see matrix)
- [ ] No UC is completely untested
- [ ] Benchmark thresholds in test code match NFR values

---

## Agent Invocation

```
@greatest-ever-developer Implement all tasks in docs/implementation/v1/05-testing-and-demo.md

Read the full task file first. Execute tasks in order:

TEST-01: Playwright Setup
1. Install @playwright/test in web-cop/.
2. Create playwright.config.ts with Chromium only, Vite dev server.
3. Create e2e/ directory with helpers.ts.
4. Add npm scripts to package.json.

TEST-02: Browser E2E Tests
1. Read web-cop/src/components/ to understand existing component structure.
2. Create 7 spec files under web-cop/e2e/ with minimum 10 tests total.
3. Use mock gRPC approach for CI-friendly execution.
4. Ensure each spec file has classification header.
5. Run `npm run test:e2e` and fix any failures.

TEST-03: Maritime Demo Scenario
1. Read existing tools/simulator/scenarios/ for YAML schema format.
2. Create maritime-demo.yaml with 30 surface vessels, Halifax area, 20 min.
3. Validate YAML loads by running simulator in dry-run if supported.

TEST-04: Multi-Domain Demo Scenario
1. Create multi-domain-demo.yaml with 25 surface + 15 air + 5 subsurface + 10 cyber.
2. Must include all 6 sensor types.
3. Validate YAML format matches existing scenarios.

TEST-05: Demo Scripts
1. Create scripts/demo/ directory.
2. Create run-maritime-demo.sh, run-multi-domain-demo.sh, stop-demo.sh.
3. Make all scripts executable (chmod +x).
4. Validate scripts parse correctly (bash -n).

TEST-06: Negative E2E Tests
1. Read tests/e2e/ for existing patterns and helpers.
2. Create tests/e2e/negative_test.go with 4 negative tests.
3. Follow existing build-tag and skipE2E patterns.

TEST-07: Test Data Audit
1. Cross-reference UC001-UC015 against all test files.
2. Report any remaining gaps.
3. Ensure benchmark thresholds match NFR requirements.

Commit each task separately with message prefix "v1(testing):".
```
