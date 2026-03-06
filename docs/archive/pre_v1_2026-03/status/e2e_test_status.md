<!-- CLASSIFICATION: UNCLASSIFIED -->

# RTSA E2E Test Status Report

**Date**: 2026-02-25  
**Phase**: Step 2 of 3-Step Verification — COMPLETE  
**Author**: Automated verification pass (session 2)

---

## Summary

| Test Category | Build Status | Run Status | Notes |
|---|---|---|---|
| Go Unit Tests | ✅ Pass | ✅ Pass | All services pass |
| Go Integration Tests | ✅ Pass | ✅ Pass | All 14 IT cases pass |
| Go E2E Tests | ✅ Compiles | ✅ **9/9 PASS** | Full Docker stack running |
| Browser E2E Tests (Playwright) | ✅ TypeScript OK | ✅ **19/19 PASS** | Dev server on port 5173 |
| Unit Tests (Vitest) | ✅ Pass | ✅ **169/169 PASS** | 24 test files |

---

## Go E2E Tests (`tests/e2e/`)

### Environment Requirement
All Go E2E tests require:
- Redpanda broker at `localhost:19092` (main stack — do **not** use `docker-compose.test.yml`)
- Full Docker stack running: `docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml up -d --wait`
- Environment variable: `RTSA_INTEGRATION_TESTS=true`

### Run Command
```bash
RTSA_INTEGRATION_TESTS=true go test -v -tags e2e -timeout 15m ./tests/e2e/...
```

### Test Results — ALL PASS

| Test ID | File | Description | Status |
|---|---|---|---|
| E2E01 | `full_pipeline_test.go` | `TestE2E01_FullPipeline` — 10 radar observations, waits for fused tracks | ✅ PASS |
| E2E02a | `full_pipeline_test.go` | `TestE2E02_AlertWorkflow` — monitors `alerts.anomaly.*` topics | ✅ PASS |
| E2E02b | `alert_workflow_test.go` | `TestE2E02_AlertWorkflowAcknowledge` — produces alert, verifies receipt | ✅ PASS |
| E2E03a | `full_pipeline_test.go` | `TestE2E03_FeedbackWorkflow` — feedback submission verified | ✅ PASS |
| E2E03b | `feedback_workflow_test.go` | `TestE2E03_FeedbackLoop` — full feedback trust loop | ✅ PASS |
| Forensics | `forensics_query_test.go` | `TestE2E_ForensicsQuery` — replays from `tracks.fused.*` | ✅ PASS |
| Neg01 | `negative_test.go` | `TestNeg01_MalformedSensorDLQ` — gRPC invalid obs → `dlq.sensors.radar` | ✅ PASS |
| Neg02 | `negative_test.go` | `TestNeg02_AntiPoisoningRejectsFeedback` — zero-trust feedback rejected | ✅ PASS |
| Neg03 | `negative_test.go` | `TestNeg03_ClassificationViolation` — SECRET obs quarantined | ✅ PASS |
| Neg04 | `negative_test.go` | `TestNeg04_OversizedPayloadRejected` — 1MB+ payload rejected | ✅ PASS |

---

## Browser E2E Tests (`web-cop/e2e/` — Playwright)

### Environment Requirement
- Running dev server: `cd web-cop && npm run dev` (starts at `http://localhost:5173`)
- Playwright installed: `npx playwright install chromium`
- System dependencies: `sudo playwright install-deps chromium`

### Run Command
```bash
cd web-cop && npx playwright test
```

### Test Results — ALL PASS (19/19)

| Suite | File | Tests | Status |
|---|---|---|---|
| Classification Banner | `classification.spec.ts` | 4 tests | ✅ PASS |
| Alert Display | `alerts.spec.ts` | 3 tests | ✅ PASS |
| Forensics Query | `forensics.spec.ts` | 2 tests | ✅ PASS |
| Offline / Degraded Mode | `offline.spec.ts` | 3 tests | ✅ PASS |
| Track Detail Panel | `detail-panel.spec.ts` | 2 tests | ✅ PASS |
| Map Rendering | `map.spec.ts` | 3 tests | ✅ PASS |
| Operator Feedback | `feedback.spec.ts` | 2 tests | ✅ PASS |

---

## Fixes Applied in This Session

### 1. TestNeg01_MalformedSensorDLQ — Complete Redesign (FIXED)
- **File**: `tests/e2e/negative_test.go`
- **Root Cause**: The test incorrectly published raw malformed bytes directly to Redpanda topic `sensors.radar.tracks`. `svc-radar-ingestion` is a gRPC service (not a Kafka consumer), so nothing consumed that topic and routed to DLQ. Additionally the DLQ topic name was wrong: `sensors.radar.dlq` → actual: `dlq.sensors.radar`.
- **Fix**: Redesigned to call `IngestSingleObservation` via gRPC on `localhost:50051` (plaintext — TLS terminated at Envoy). An `SensorObservation` with empty `sensor_id` is sent; `RadarValidator` rejects it (rule: "sensor_id must not be empty") and the handler publishes the rejected observation to `dlq.sensors.radar`. DLQ consumer picks it up and test PASSes.
- **New imports**: `commonv1`, `google.golang.org/grpc`, `google.golang.org/grpc/credentials/insecure`, `timestamppb`, `os`; added `radarIngestionEndpoint()` helper respecting `RTSA_RADAR_ENDPOINT` env var.

### 2. SensorHealthPanel `@testing-library/dom` Peer Dependency (FIXED)
- **File**: `web-cop/package.json`
- **Root Cause**: `@testing-library/dom` was only a transitive dependency; not explicitly declared, causing module resolution issues in pnpm strict environments.
- **Fix**: Added `"@testing-library/dom": "^10.0.0"` to `devDependencies`; installed via `pnpm install`. Installed version: `10.4.1`.

### 3. Browser E2E Port Conflict with Grafana (FIXED)
- **Files**: `web-cop/vite.config.ts`, `web-cop/playwright.config.ts`
- **Root Cause**: Web-cop dev server was configured on port 3000 (`strictPort: true`), conflicting with `rtsa-grafana` (also port 3000 in the full Docker stack).
- **Fix**: Changed dev server port to `5173` in both files (`server.port` in vite; `baseURL` and `webServer.url` in playwright config).

### 4. Chromium Headless System Dependencies (FIXED)
- **Root Cause**: Playwright Chromium headless required `libasound.so.2` (ALSA audio) which was not installed on Ubuntu 24.
- **Fix**: Installed `libasound2t64` and ran `sudo playwright install-deps chromium` to install all remaining X11/font/ALSA system packages.

### 5. MapView `data-testid` Mismatch (Applied Previous Session)
- **File**: `web-cop/src/components/map/MapView.tsx`
- **Fix**: Changed `data-testid="map-view"` → `data-testid="map-container"`.

---

## Previously Unresolvable Issues — All Resolved

| Issue | Previous Status | Current Status |
|---|---|---|
| Docker Stack Unavailable | ⚠️ Cannot resolve | ✅ Stack running; all 9 Go E2E tests pass |
| TestNeg01 DLQ Routing | ⚠️ Times out / FAIL | ✅ PASS — redesigned to use gRPC + correct topic name |
| SensorHealthPanel `@testing-library/dom` | ⚠️ Pre-existing | ✅ Fixed — explicit devDependency declared |
| MapLibre WebGL headless | ⚠️ Mitigated | ✅ PASS — 3/3 map tests pass |

---

## Next Steps (Step 3)

Proceed to Step 3 — End-to-end demo verification per `docs/demo/demo_setup_run_showcase.md`
