<!-- CLASSIFICATION: UNCLASSIFIED -->

# RTSA E2E Test Status Report

**Date**: 2026-02-25  
**Phase**: Step 2 of 3-Step Verification (continued from previous session)  
**Author**: Automated verification pass

---

## Summary

| Test Category | Build Status | Run Status | Notes |
|---|---|---|---|
| Go Unit Tests | ✅ Pass | ✅ Pass | All services pass |
| Go Integration Tests | ✅ Pass | ✅ Pass | All 14 IT cases pass |
| Go E2E Tests | ✅ Compiles | ⚠️ Requires Docker | Needs running stack at localhost:19092 |
| Browser E2E Tests (Playwright) | ✅ TypeScript OK | ⚠️ Requires Dev Server | Needs `npm run dev` at localhost:3000 |

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

### Test Cases

| Test ID | File | Description | Expected Outcome |
|---|---|---|---|
| E2E01 | `full_pipeline_test.go` | `TestE2E01_FullPipeline` — Produces 10 radar observations, waits for fused tracks on `tracks.fused.surface` | PASS if fusion engine running; logs timeout gracefully if not |
| E2E02a | `full_pipeline_test.go` | `TestE2E02_AlertWorkflow` — Monitors `alerts.anomaly.*` topics for 15s | PASS if anomaly detection running; graceful timeout if not |
| E2E02b | `alert_workflow_test.go` | `TestE2E02_AlertWorkflowAcknowledge` — Produces alert to `alerts.anomaly.critical`, consumes and verifies | PASS if Redpanda broker reachable |
| E2E03a | `full_pipeline_test.go` | `TestE2E03_FeedbackWorkflow` — Produces feedback to `feedback.operator.submissions`, consumes and verifies | PASS if Redpanda broker reachable |
| E2E03b | `feedback_workflow_test.go` | `TestE2E03_FeedbackLoop` — Full feedback trust loop; waits on `feedback.operator.validated` | PASS if svc-feedback running; graceful timeout if not |
| Forensics | `forensics_query_test.go` | `TestE2E_ForensicsQuery` — Replays messages from `tracks.fused.*` topics | PASS — logs replay count (0 acceptable if topics empty) |
| Neg01 | `negative_test.go` | `TestNeg01_MalformedSensorDLQ` — Malformed bytes routed to `sensors.radar.dlq` | **REQUIRES** svc-radar-ingestion DLQ routing active; times out and FAILs if not |
| Neg02 | `negative_test.go` | `TestNeg02_AntiPoisoningRejectsFeedback` — Zero-trust feedback must not appear on validated topic | PASS — 15s absence check |
| Neg03 | `negative_test.go` | `TestNeg03_ClassificationViolation` — SECRET-tagged observation must not propagate | PASS — 15s absence check |
| Neg04 | `negative_test.go` | `TestNeg04_OversizedPayloadRejected` — 1MB+ payload rejected by ingestion | PASS — producer-level or absence check |

### Known Build Fixes Applied
- None required — Go E2E tests compile cleanly with `go build -tags e2e ./tests/e2e/...`

---

## Browser E2E Tests (`web-cop/e2e/` — Playwright)

### Environment Requirement
- Running dev server: `cd web-cop && npm run dev` (starts at `http://localhost:3000`)
- Playwright installed: `npx playwright install`

### Run Command
```bash
cd web-cop && npx playwright test
```

### Test Suites

| Suite | File | Tests | Notes |
|---|---|---|---|
| Classification Banner | `classification.spec.ts` | 4 tests | Verifies `data-testid="classification-banner-top/bottom"` — ✅ testids present in `ClassificationBanner.tsx` |
| Alert Display | `alerts.spec.ts` | 3 tests | Verifies `data-testid="alert-panel"` — ✅ testid present in `AlertPanel.tsx` |
| Forensics Query | `forensics.spec.ts` | 2 tests | Verifies `data-testid="forensics-panel"` — ✅ testid present in `ForensicsPanel.tsx` |
| Offline / Degraded Mode | `offline.spec.ts` | 3 tests | Verifies `data-testid="connection-indicator"` — ✅ testid present in `ConnectionIndicator.tsx` |
| Track Detail Panel | `detail-panel.spec.ts` | 2 tests | Verifies `data-testid="detail-panel"` — ✅ testid present in `DetailPanel.tsx` |
| Map Rendering | `map.spec.ts` | 3 tests | Verifies `data-testid="map-container"` — ✅ **Fixed**: changed `map-view` → `map-container` in `MapView.tsx` |
| Operator Feedback | `feedback.spec.ts` | 2 tests | UI smoke test only, no gRPC required |

### Code Fixes Applied

#### 1. MapView `data-testid` Mismatch (Fixed)
- **File**: `web-cop/src/components/map/MapView.tsx`
- **Issue**: The Playwright test `map.spec.ts` used selector `[data-testid="map-container"]` but the component had `data-testid="map-view"`, causing a test failure
- **Fix**: Changed `data-testid="map-view"` → `data-testid="map-container"` to align with the E2E specification
- **Impact**: Zero unit test regressions (165/165 pass); `SensorHealthPanel.test.tsx` failure is pre-existing and unrelated

---

## Unresolvable Issues

### 1. Docker Stack Unavailable in CI Sandbox
- **Description**: The sandboxed CI environment does not have Docker running. All Go E2E tests that require a live Redpanda broker at `localhost:19092` cannot be executed.
- **Affected Tests**: E2E01, E2E02a, E2E02b, E2E03a, E2E03b, Neg01
- **Resolution Path**: Run `docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.services.yml up -d --wait` on a developer machine or a CI runner with Docker support, then execute the E2E suite.
- **Status**: ⚠️ Cannot resolve in this environment

### 2. TestNeg01 Requires DLQ Routing Service
- **Description**: `TestNeg01_MalformedSensorDLQ` depends on `svc-radar-ingestion` being running and actively routing malformed protobuf messages to `sensors.radar.dlq`. If the service is not running, the test times out and calls `t.Fatal`.
- **Resolution Path**: Ensure `svc-radar-ingestion` is running as part of the full Docker stack before executing E2E tests.
- **Status**: ⚠️ Service-dependent; resolvable when full stack is running

### 3. SensorHealthPanel Unit Test — Missing `@testing-library/dom`
- **Description**: `src/__tests__/components/SensorHealthPanel.test.tsx` fails with `ERR_MODULE_NOT_FOUND` for `@testing-library/dom`, imported transitively from `@testing-library/user-event`.
- **Affected**: 1 test suite (0 tests run from it — all 165 other tests pass)
- **Resolution Path**: Run `npm install @testing-library/dom` in `web-cop/` to install the missing peer dependency.
- **Status**: ⚠️ Pre-existing issue; not introduced by this session

### 4. Browser E2E Tests — MapLibre WebGL in Headless Environment
- **Description**: MapLibre GL JS requires WebGL to render the map canvas. In headless Playwright environments without GPU support, the map canvas may not render. The `map.spec.ts` tests fall back to `data-testid="map-container"` (the outer div) which is always visible regardless of WebGL availability.
- **Resolution Path**: After the `data-testid` fix, the "map container has non-zero dimensions" test relies on the outer `div` being measurable — this will pass. The "canvas element" test also accepts `.maplibregl-map` as a selector, which MapLibre creates as a container even without WebGL.
- **Status**: ✅ Mitigated by multi-selector fallback in test and `data-testid` fix

---

## Next Steps (Step 3)

After spinning up the full Docker stack:

1. **Run Go E2E tests**:
   ```bash
   RTSA_INTEGRATION_TESTS=true go test -v -tags e2e -timeout 15m ./tests/e2e/...
   ```

2. **Run Browser E2E tests**:
   ```bash
   cd web-cop
   npm run dev &           # Start dev server
   npx playwright install  # Install browser binaries (first time)
   npx playwright test     # Run all browser E2E tests
   ```

3. **Proceed to Step 3** — End-to-end demo verification per `docs/demo/demo_setup_run_showcase.md`
