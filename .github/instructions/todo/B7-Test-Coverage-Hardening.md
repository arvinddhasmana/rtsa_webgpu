<!-- CLASSIFICATION: UNCLASSIFIED -->

# B7 — Test Coverage and Hardening

> **Batch**: B7 of 7
> **Theme**: Test Coverage Hardening
> **Priority**: WARNING — Workers excluded from coverage; critical paths untested
> **Agent Profile**: `greatest-ever-dev-forworkflow`
> **Source**: v4 Implementation Review R-023 through R-028

---

## Context

This final batch achieves ≥80% line coverage across the `web-cop-gpu/` frontend and validates cross-language serialization correctness. Workers are entirely excluded from current coverage configuration; gRPC service wrappers have no tests; and there is no Go→Rust round-trip integration test. This batch also fixes a device-loss listener leak in the Render Worker.

---

## Issue R-023 — BLOCKING: Workers excluded from Vitest coverage

**File**: `web-cop-gpu/vitest.config.ts`

**Problem**: `exclude: ["src/workers/**", "src/**/*.d.ts"]` — the Data Worker and Render Worker, which contain the most critical runtime code, have **zero test coverage**.

**Required Fix**:

1. Remove `"src/workers/**"` from the `coverage.exclude` array in `vitest.config.ts`.

2. Extract pure logic functions from each worker into separate importable modules so they can be tested without the Worker global context:
   - From `render-worker.ts`: extract any frame-timing logic, SAB read helpers, or state-machine transitions into `web-cop-gpu/src/gpu/render-logic.ts`.
   - From `data-worker.ts`: extract datagram parsing, SAB write helpers, and mock data generation into `web-cop-gpu/src/workers/data-worker-logic.ts`.

3. Write Vitest unit tests for the extracted modules. Mock the `SharedArrayBuffer`, `Worker`, and `WebTransport` globals as needed.

**Target**: After this fix, `pnpm test:coverage` must report ≥80% line coverage across the `src/` tree (excluding `.d.ts` files only).

---

## Issue R-024 — WARNING: No unit tests for gRPC service wrappers

**Files**:

- `web-cop-gpu/src/services/alerts.ts`
- `web-cop-gpu/src/services/feedback.ts`
- `web-cop-gpu/src/services/query.ts`

**Required Fix**: Create `web-cop-gpu/src/services/alerts.test.ts`, `feedback.test.ts`, and `query.test.ts` with Vitest tests that:

1. **`alerts.test.ts`**:
   - Mock the ConnectRPC transport.
   - Test `startAlertStream` emits properly typed alert objects to the designated signal/callback.
   - Test `acknowledgeAlert` calls the gRPC client with the correct `alertId` and `operatorId` parameters.

2. **`feedback.test.ts`**:
   - Test `submitFeedback` validates required inputs (non-empty `trackId`, valid `feedbackType`) before calling gRPC.
   - Test that a successful gRPC response resolves the promise.
   - Test that a gRPC error propagates and does not swallow the error silently.

3. **`query.test.ts`**:
   - Test `fetchTrackDetail` returns a correctly typed `TrackDetail` object mapped from the proto response.
   - Test `searchTracks` handles an empty result set without error.
   - Test `searchTracks` handles pagination fields if present in the proto response.

Read each service file before writing tests. Do not mock internal implementation details — mock only the ConnectRPC transport boundary.

---

## Issue R-025 — WARNING: Device-loss re-init registers duplicate listeners

**File**: `web-cop-gpu/src/workers/render-worker.ts`

**Problem**: When `device.lost` fires, `init(activeCanvas, activeSab)` is called again. Inside `init`, `initGPU` (or `device.ts`) registers another `device.lost` listener on the new device. The old listener from the previous cycle remains attached, creating N duplicate handlers after N device loss events.

**Required Fix**:

1. Read `web-cop-gpu/src/workers/render-worker.ts` and `web-cop-gpu/src/gpu/device.ts` before editing.
2. Move the `device.lost` handler registration **exclusively** into the Render Worker's `init()` function. Do not register it in `device.ts` (or any shared module).
3. Before registering the new handler on device re-init, ensure any prior `device.lost` listener from the previous device is not re-triggered (the old device's promise is already settled; this is handled automatically by the spec — but confirm there is no explicit `addEventListener`-style registration that would accumulate).
4. Add a comment explaining this invariant.

---

## Issue R-026 — WARNING: `TrackRecord` struct size verified correct — no action needed

**Finding from review**: The `TrackRecord` WGSL struct in `interpolation.wgsl` is correctly sized at 128 bytes with correct alignment. No fix required.

**Action**: Add a block comment above the struct definition in `web-cop-gpu/src/shaders/interpolation.wgsl` confirming this analysis for future reviewers:

```wgsl
// TrackRecord layout (128 bytes, verified 2026-03-06):
// Offset 0x00: lat, lon, heading, speed, altitude (5×f32 = 20 bytes)
// Offset 0x14: track_id, update_epoch_ms, alert_flags, threat_level,
//              icon_type, trail_head_idx, trail_count (7×u32 = 28 bytes)
// Offset 0x30: trail (array<vec4<f32>, 5> = 80 bytes, 16-byte aligned ✓)
// Total: 128 bytes
```

---

## Issue R-027 — WARNING: No Go→Rust round-trip integration test

**Location**: `tests/integration/`

**Problem**: Go serializer tests verify Go→binary, and Rust tests verify binary→fields, but no test verifies the two are compatible at matching byte offsets.

**Required Fix**: Create `tests/integration/serializer_roundtrip_test.go` (or a shell-based test that invokes both):

1. Write a Go test that serializes a known `TrackUpdate` struct (with deterministic field values) using `pkg/flatbuf/Serializer.Serialize()` and writes the 128-byte output to a temp file or `testdata/roundtrip_track.bin`.

2. Write a Rust test in `wasm-transforms/` (or a new `tests/integration/` Rust crate) that reads `roundtrip_track.bin` and verifies every field value matches the known input.

3. Alternatively, implement as a single Go test that:
   - Serializes the record.
   - Invokes the Wasm decoder CLI (if one exists) via `exec.Command` and asserts the decoded JSON matches.

The integration test must be runnable via `go test ./tests/integration/... -run TestSerializerRoundtrip`.

---

## Issue R-028 — WARNING: Security audit E2E test gaps

**File**: `web-cop-gpu/e2e/security-audit.spec.ts`

**Required Fix**: Review this file and ensure the following assertions exist. Add any that are missing:

1. **CSP header present**: The response from the dev/prod server includes a `Content-Security-Policy` header. Playwright can intercept the page load response and assert on headers.

2. **No `eval()` in bundle**: Intercept the JavaScript bundle response and assert it contains no calls to `eval(` or `new Function(`.

3. **No hardcoded external URLs**: Assert the bundle contains no references to `http://` or `https://` URLs pointing to external services (only `localhost` and env-var-derived URLs are permitted).

4. **Wasm module integrity**: If a `<link rel="modulepreload">` or `<script>` tag loads the Wasm wrapper, assert it also specifies an `integrity` attribute (SRI).

5. **COOP/COEP headers**: The page response must include `Cross-Origin-Opener-Policy: same-origin` and `Cross-Origin-Embedder-Policy: require-corp` (required for `SharedArrayBuffer` availability).

If any assertion fails, **fix the root cause** (e.g., add/correct the CSP header in the Vite dev server config or production server config), do not just remove the assertion.

---

## Implementation Rules

1. Read every file before editing. Do not add helpers or abstractions beyond what is needed for the tests.
2. Tests must not use `console.log` — use Vitest's `expect` assertions only.
3. Mocks must be scoped to the test file — do not use global mocks via `vitest.config.ts` unless already established.
4. All new test files must include `// CLASSIFICATION: UNCLASSIFIED` as the first line.
5. Follow `docs/sdlc_guidelines/05_testing/` guidelines.
6. Go code must use `-race` flag in test commands.

---

## Verification Steps

1. `cd web-cop-gpu && pnpm test:coverage` — line coverage must be ≥80% for `src/` (excluding `.d.ts`).
2. `go test ./... -race -coverprofile=coverage.out && go tool cover -func=coverage.out` — all Go tests pass with race detector.
3. `go test ./tests/integration/... -run TestSerializerRoundtrip` — round-trip test passes.
4. `cd web-cop-gpu && pnpm test:e2e` — security audit E2E spec passes without modifications to assertions.
5. No duplicate device-loss listeners: simulate two consecutive `device.lost` events in a test and verify the init function is called exactly once per event (not 2×, 4×, etc.).

---

## PR Instructions

- PR title: `test: add worker coverage, gRPC service tests, round-trip test, device-loss fix, security E2E (B7)`
- Label: `ai-orchestrator`
- After the PR is merged, move this file from `.github/instructions/todo/B7-Test-Coverage-Hardening.md` to `.github/instructions/done/B7-Test-Coverage-Hardening.md`.
