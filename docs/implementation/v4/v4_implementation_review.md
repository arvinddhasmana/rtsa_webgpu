<!-- CLASSIFICATION: UNCLASSIFIED -->

# v4 Implementation Review — WebGPU COP

> **Document**: RTSA v4 Full Implementation Review
> **Version**: 1.0
> **Classification**: UNCLASSIFIED
> **Review Date**: 2026-03-06
> **Reviewer**: Meanest Ever Reviewer (AI Agent)
> **Scope**: All 5 phases — Phase 0 through Phase 4

---

## Executive Summary

The v4 WebGPU COP implementation is **architecturally sound** and follows the `docs/architecture/v1/RTSA_WebGPU_Architecture_v1.md` reference closely. The multi-worker threading model, zero-per-frame-allocation GPU pipeline, SharedArrayBuffer ring buffer, and dual-protocol (WebTransport hot path + gRPC-Web cold path) design are all correctly implemented.

However, there are **blocking issues** that prevent an end-to-end demo from functioning and must be resolved before the system can be demonstrated using the existing simulation scripts. This review is organized in **priority order**: issues that block the demo are listed first, followed by correctness issues, then hardening items.

---

## Review Organisation — Fix Order for AI Developer Agent

Issues are grouped into **7 batches** designed to be fixed sequentially without breaking the system. Each batch is self-contained and testable.

| Batch | Theme | Impact | Issues |
|-------|-------|--------|--------|
| **B1** | Missing `data-testid` attributes — E2E tests fail | E2E tests broken | R-001 to R-004 |
| **B2** | Demo data flow — Mock data writes incorrect `update_epoch_ms` | Tracks appear stale, interpolation broken | R-005 to R-008 |
| **B3** | LOD system not wired into render loop | LOD code exists but is dead code | R-009 to R-010 |
| **B4** | Pick buffer readback race condition + coordinate scaling | Track selection fails intermittently | R-011 to R-013 |
| **B5** | WebTransport ↔ Data Worker integration gaps | Real backend connection non-functional | R-014 to R-017 |
| **B6** | Security & compliance gaps | SDLC policy violations | R-018 to R-022 |
| **B7** | Test coverage & hardening | Test gaps, missing edge cases | R-023 to R-028 |

---

## Batch 1 — Missing `data-testid` Attributes (E2E Tests Broken)

All Playwright E2E tests (`e2e/*.spec.ts`) select elements via `data-testid` attributes, but **no component in `src/components/` emits any `data-testid` attribute**. Every E2E test will fail immediately.

### R-001 — BLOCKING: ClassificationBanner missing `data-testid`

**File**: `web-cop-gpu/src/components/shell/ClassificationBanner.tsx`
**Evidence**: E2E tests use `[data-testid="classification-banner-top"]` (e.g., `e2e/cold-boot.spec.ts:26`, `e2e/helpers.ts:30`)
**Current**: The component renders `<div role="banner" ...>` but has no `data-testid`.
**Required Fix**: Add `data-testid="classification-banner-top"` to the banner `<div>`.

### R-002 — BLOCKING: StatusBar missing `data-testid`

**File**: `web-cop-gpu/src/components/status/StatusBar.tsx`
**Evidence**: E2E uses `[data-testid="status-bar"]` (`e2e/cold-boot.spec.ts:64`) and `[data-testid="fps-display"]`, `[data-testid="latency-display"]` (`e2e/visual-regression.spec.ts:75-76`)
**Required Fix**: Add `data-testid="status-bar"` to the root `<div>`, and `data-testid="fps-display"` / `data-testid="latency-display"` to the respective value elements.

### R-003 — BLOCKING: RoleSelector, DashboardSelector, ConnectionIndicator missing `data-testid`

**Files**:
- `web-cop-gpu/src/components/toolbar/RoleSelector.tsx` — needs `data-testid="role-selector"` on the root `<div>`
- `web-cop-gpu/src/components/toolbar/DashboardSelector.tsx` — needs `data-testid="dashboard-selector"` on the root `<div>`
- `web-cop-gpu/src/components/toolbar/ConnectionIndicator.tsx` — needs `data-testid="connection-indicator"` on the root `<div>`

**Evidence**: `e2e/cold-boot.spec.ts:72`, `e2e/role-access.spec.ts:15,23,43,59`, `e2e/helpers.ts:40`, `e2e/reconnection.spec.ts:16`

### R-004 — BLOCKING: AlertSidebar, FeedbackForm, SearchOverlay, AppShell toolbar missing `data-testid`

**Files**:
- `web-cop-gpu/src/components/panels/AlertSidebar.tsx` — needs `data-testid="alert-sidebar"` on root
- `web-cop-gpu/src/components/panels/FeedbackForm.tsx` — needs `data-testid="feedback-form"` on modal container
- `web-cop-gpu/src/components/search/SearchOverlay.tsx` — needs `data-testid="search-overlay"` on overlay container
- `web-cop-gpu/src/components/shell/AppShell.tsx` — needs `data-testid="app-toolbar"` on the left toolbar `<div>`

**Evidence**: `e2e/alerts-feedback.spec.ts:18,32,54,71`, `e2e/visual-regression.spec.ts:103`, `e2e/role-access.spec.ts:71`

**Verification**: After fixing, run `pnpm test:e2e` to confirm all E2E selectors resolve.

---

## Batch 2 — Demo Data Flow: Mock Data Timestamp & Interpolation Issues

### R-005 — BLOCKING: Mock data `update_epoch_ms` truncated to 32 bits — interpolation drift

**File**: `web-cop-gpu/src/gpu/mock-data.ts` (line ~97)
**Evidence**: `trackData.setUint32(base + 0x2c, now >>> 0, true)` — uses `Date.now()` which is ~1.7 trillion. The `>>> 0` truncation to unsigned 32-bit loses the high bits. In the interpolation shader, `current_time_ms` is also `u32`, but the WGSL subtraction `i32(uniforms.current_time_ms) - i32(track.update_epoch_ms)` computes the delta. Both sides are truncated the same way, so the **delta** is correct IF they wrap consistently.
**Actual Issue**: The `update_epoch_ms` in mock data and the `current_time_ms` uniform MUST both use the same truncation (`& 0xFFFF_FFFF` or `>>> 0`). Currently `mock-data.ts` uses `now >>> 0` and `uniforms.ts` uses `currentTimeMs >>> 0`. This is **consistent** — but the `writeUniforms` call passes `performance.now() | 0` (from render-worker.ts line ~128: `const now = performance.now() | 0`).
**The Problem**: `performance.now()` returns **milliseconds since time origin** (a small number, few million at most), while `Date.now()` returns **Unix epoch ms** (1.7 trillion). These are different time bases! The WGSL interpolation shader computes `i32(current_time_ms) - i32(update_epoch_ms)` which will produce a **massive negative delta**, clamped to 0 by `clamp(f32(raw_dt_ms), 0.0, MAX_DR_S * 1000.0)`. This means **dead-reckoning interpolation is completely broken** — all tracks appear frozen at their last known position.

**Required Fix**: Use the same time base in both locations. Either:
- (a) Change `mock-data.ts` to write `(performance.now() | 0) >>> 0` as `update_epoch_ms`, OR
- (b) Change `render-worker.ts` to pass `(Date.now() & 0xFFFFFFFF)` as `currentTimeMs`

Option (b) is preferred because it matches the Go serializer which uses `time.Now().UnixMilli()`.

**Impact**: Without this fix, dead-reckoning interpolation is dead code. Tracks don't move smoothly between server updates.

### R-006 — BLOCKING: Data Worker mock mode also uses `Date.now()` but Render Worker loop overwrites SAB anyway

**File**: `web-cop-gpu/src/workers/data-worker.ts` (line ~172)
**Evidence**: `view.setUint32(0x2c, Date.now() & 0xffffffff, true)` — uses `Date.now()`.
**File**: `web-cop-gpu/src/workers/render-worker.ts` (lines ~129-132)
**Evidence**: The render worker calls `writeMockTracksToSAB(activeSab, MOCK_TRACK_COUNT)` every frame, **overwriting** whatever the Data Worker wrote. In mock/dev mode, **both workers** write to the SAB simultaneously without coordination.
**Required Fix**: In mock mode (no WebTransport URL), only ONE worker should write mock data to the SAB. The Render Worker should read tracks from the SAB, not write them. This is the correct Phase 2 architecture where the Data Worker writes and the Render Worker reads. Currently the Render Worker overwrites Data Worker writes every frame.
**Recommended approach**: Remove the `initMockTracks` / `writeMockTracksToSAB` / `tickMockTracks` calls from the Render Worker when the Data Worker is handling mock data. The Data Worker mock mode should be the sole writer.

### R-007 — WARNING: Trail data in mock records is always static

**File**: `web-cop-gpu/src/gpu/mock-data.ts` (lines ~107-116)
**Evidence**: Trail positions are computed once from the current position and never updated as tracks move. When `tickMockTracks()` advances positions, the trail ring buffer still contains the original positions from `writeMockTracksToSAB`. This means trails are visually incorrect — they should show the track's recent path, not a static offset.
**Required Fix**: Update trail positions in `tickMockTracks()` by shifting the ring buffer and inserting the current position as the newest trail segment.

### R-008 — WARNING: Data Worker mock mode writes only one track per 16ms interval

**File**: `web-cop-gpu/src/workers/data-worker.ts` (line ~182)
**Evidence**: `startMockUpdates()` writes exactly 1 mock record per interval. At 60 Hz, that's 60 tracks/second. To reach 50,000 tracks at 60 Hz, the SAB would take ~14 minutes to fill. The Render Worker's mock data writes 50,000 tracks instantly.
**Required Fix**: Either (a) have the Data Worker batch-write mock tracks (e.g., 1000 per tick), or (b) as recommended in R-006, consolidate mock data generation to a single location with correct semantics.

---

## Batch 3 — LOD System Dead Code

### R-009 — WARNING: LOD system never invoked from render loop

**File**: `web-cop-gpu/src/gpu/lod.ts`
**Evidence**: The `computeLod()` function is fully implemented, but it is never called from `renderer.ts` or `render-worker.ts`. All render passes (trails, halos, labels) always execute regardless of zoom level. At high zoom-out levels with 50k tracks, this wastes GPU time and may prevent hitting the 60 FPS target.
**Required Fix**: Import `computeLod` in `renderer.ts`, call it with `state.camera.scale` and `state.trackCount`, and conditionally skip trail, halo, and label render passes based on the returned `LodFlags`. Also use `LodFlags.maxInstances` to cap the instance count for the icon pass.

### R-010 — WARNING: Frame timer never invoked

**File**: `web-cop-gpu/src/gpu/frame-timer.ts`
**Evidence**: The `FrameTimer` class is implemented but never instantiated or used in the render loop. GPU timestamp queries and per-pass timing are unavailable for performance profiling.
**Required Fix**: Wire `FrameTimer` into the render loop in `render-worker.ts`. At minimum, enable JS-side frame timing (markJsStart/markJsEnd) and report results via the `stats` postMessage to the main thread. GPU timestamp queries can be conditionally enabled when the `timestamp-query` feature is available.

---

## Batch 4 — Pick Buffer Issues

### R-011 — BLOCKING: Pick buffer readback races with render loop

**File**: `web-cop-gpu/src/gpu/pick.ts` (function `readPickPixel`, line ~75)
**Evidence**: `readPickPixel` creates a new command encoder, copies the entire pick texture to the readback buffer, and submits it independently. But the render loop (in `setInterval`) may submit the next frame's commands while `mapAsync` is awaiting, overwriting the pick texture with new data. The readback buffer is also reused without checking if it's still mapped from a previous read.
**Required Fix**:
1. Check if `readbackBuffer` is already mapped before issuing a new copy (guard against double-map errors).
2. Consider using a double-buffered readback approach: two readback buffers alternated each pick request.
3. Alternatively, pause the render loop during pick readback or use a single-pixel copy instead of copying the entire texture.

### R-012 — WARNING: Pick texture coordinates don't account for DPR correctly

**File**: `web-cop-gpu/src/gpu/pick.ts` (line ~82)
**Evidence**: `const px = Math.floor((canvasX / 2))` — this divides by 2 because the pick texture is half-resolution. But `canvasX` and `canvasY` come from the main thread's `handleCanvasClick` which already applies `devicePixelRatio`: `const x = Math.round((e.clientX - rect.left) * devicePixelRatio)`. The pick texture dimensions are `canvasWidth / 2` and `canvasHeight / 2` where `canvasWidth` is the physical pixel width. So dividing `(canvasX)` by 2 is correct for mapping physical pixels → half-res pick texture. **This is OK** — no fix needed, but add a comment clarifying the coordinate flow.

### R-013 — WARNING: Pick readback copies entire texture for single pixel

**File**: `web-cop-gpu/src/gpu/pick.ts` (lines ~84-92)
**Evidence**: The entire pick texture (`w × h × 4 bytes`) is copied to the readback buffer for every click. For a 1920×1080 canvas, that's 960×540 × 4 = ~2 MB per click. This is wasteful.
**Required Fix**: Copy only a single pixel (or small tile) using the `origin` and `size` parameters of `copyTextureToBuffer`. Change:
```typescript
encoder.copyTextureToBuffer(
  { texture: pick.texture, origin: { x: px, y: py } },
  { buffer: pick.readbackBuffer, bytesPerRow: 256 },
  { width: 1, height: 1 },
);
```
This reduces the copy to 4 bytes + alignment overhead. It also allows reducing the readback buffer size to 256 bytes (one aligned row).

---

## Batch 5 — WebTransport Integration Gaps

### R-014 — BLOCKING: Data Worker never receives WebTransport URL

**File**: `web-cop-gpu/src/App.tsx` (lines ~210-211)
**Evidence**: `const dataInit: DataInitMessage = { type: "init", sab };` — the `url` field is never provided. This means the Data Worker always enters mock mode, never connecting to the Go WebTransport server.
**Required Fix**: Read the WebTransport URL from the Vite environment variable `VITE_WEBTRANSPORT_URL` and pass it to the Data Worker:
```typescript
const dataInit: DataInitMessage = {
  type: "init",
  sab,
  url: import.meta.env.VITE_WEBTRANSPORT_URL,
};
```
When the env var is undefined (dev mode without backend), the Data Worker correctly falls back to mock mode.

### R-015 — BLOCKING: No JWT token in WebTransport URL

**File**: `web-cop-gpu/src/workers/data-worker.ts` (line ~261)
**Evidence**: `const transport = new WebTransport(url)` — connects without any authentication token. The Go server (`pkg/webtransport/server.go:handleSession`) requires a JWT token in the `token` query parameter and returns 401 without it.
**Required Fix**: Before connecting, the Data Worker must acquire a JWT token via the gRPC cold path. This requires:
1. A new gRPC service call (or the Data Worker requests a token from the main thread via postMessage, and the main thread calls the auth service)
2. Append the token to the URL: `new WebTransport(`${url}?token=${jwt}`)`
3. Implement token refresh when the token nears expiry (5-min TTL per `auth.go`)

### R-016 — WARNING: Data Worker Wasm import path may not resolve in production build

**File**: `web-cop-gpu/src/workers/data-worker.ts` (line ~88)
**Evidence**: `const mod = await import(/* @vite-ignore */ new URL("../../wasm-decoder/pkg/wasm_decoder.js", import.meta.url).href)` — this relative path works if the Rust Wasm module is compiled and the `pkg/` directory exists. The `wasm-decoder` directory is outside `src/`, so Vite may not bundle it into the production build.
**Required Fix**: Either:
1. Configure Vite to include the Wasm output in the build (copy plugin), OR
2. Move the Wasm output to a location under `src/` or `public/`, OR
3. Add a build step that copies `wasm-decoder/pkg/` → `public/wasm/` and update the import path.
**Verification**: Run `pnpm build` and confirm the Wasm module is included in `dist/`.

### R-017 — WARNING: No WebTransport fallback to gRPC-Web streaming

**Evidence**: The v1 architecture specifies that when WebTransport is blocked by enterprise proxies, the system should fall back to gRPC-Web streaming for the hot path. Currently, if `WebTransport(url)` fails, the Data Worker retries with exponential backoff indefinitely but never falls back to gRPC-Web.
**Required Fix**: After N consecutive WebTransport failures (e.g., 5), switch the Data Worker to a gRPC-Web streaming mode that receives track updates via `server-stream` and writes them to the SAB. This is a new feature, not a bug fix, but is critical for enterprise deployment scenarios.

---

## Batch 6 — Security & Compliance Gaps

### R-018 — BLOCKING: `alertFlags` always 0 in Go serializer

**File**: `pkg/flatbuf/serializer.go` (line ~119)
**Evidence**: `alertFlags := uint32(0) // Phase 3 will wire alert state; keep zero for now`
**Impact**: In a real backend demo, no track will ever show an alert halo because alert_flags is hardcoded to 0. The Alert Sidebar will only be populated by gRPC streaming (cold path), but the GPU visual indicator (pulsing halos on the canvas) will never appear for any track.
**Required Fix**: Wire the alert state from the `TrackUpdate` proto message. If the proto doesn't have an alert field yet, at minimum derive it from `threat_level >= 4` (Suspect/Hostile) as a temporary heuristic so the demo shows halos.

### R-019 — WARNING: `hostileClassToThreatLevel` mapping inconsistent with WGSL

**File**: `pkg/flatbuf/serializer.go` (lines ~185-199)
**Evidence**: The Go serializer maps `HOSTILE → 5`, `SUSPECT → 4`, `NEUTRAL → 3`, `FRIENDLY → 2`, `PENDING → 1`. But the WGSL `threat_color` function in `trail.wgsl` maps: `0=grey, 1=blue, 2=green, 3=amber, 4=orange, 5=red`. This means **Friendly (2) = green, Neutral (3) = amber** — amber for neutral is unexpected; STANAG typically uses green for neutral. Similarly, the `halos.wgsl` colour mapping shows `5=red, 4=orange, 3=amber, default=blue` which means **Friendly tracks get blue halos** even though they shouldn't have alert halos at all.
**Required Fix**: Document the threat level colour mapping explicitly in the layout.go constants, and verify it matches the STANAG APP-6 standard colour codes used in the WGSL shaders. Ensure Friendly=blue/green, Neutral=green, Suspect=amber/yellow, Hostile=red.

### R-020 — WARNING: `console.log` statements in production code

**Files**: Multiple files contain `console.log`, `console.warn`, `console.error` calls:
- `web-cop-gpu/src/workers/render-worker.ts` — lines 95, 108, 141, 168, etc.
- `web-cop-gpu/src/workers/data-worker.ts` — line ~87 (Wasm load failure log)
- `web-cop-gpu/src/services/alerts.ts` — line ~66

**SDLC Rule**: `solidjs_standards.md §7.1` — No `console.log` in production code.
**Required Fix**: Replace with structured logging or remove. For Worker code where structured logging isn't available, gate behind a `__DEV__` compile-time flag or use `import.meta.env.DEV` conditionals.

### R-021 — WARNING: FeedbackForm uses hardcoded `operatorId: "operator"`

**File**: `web-cop-gpu/src/components/panels/FeedbackForm.tsx` (line ~50)
**Evidence**: `operatorId: "operator"` — the operator identity is hardcoded rather than derived from the authenticated session.
**Required Fix**: Source the `operatorId` from the JWT claims (same token used for WebTransport auth). Pass it through a signal or context provider to the feedback form.

### R-022 — WARNING: `AlertSidebar` acknowledgement uses hardcoded `operatorId: "operator"`

**File**: `web-cop-gpu/src/components/panels/AlertSidebar.tsx` (line ~38)
**Evidence**: `await acknowledgeAlert(props.alert.alertId, "operator")`
**Required Fix**: Same as R-021 — source from authenticated session.

---

## Batch 7 — Test Coverage & Hardening

### R-023 — BLOCKING: Workers excluded from Vitest coverage

**File**: `web-cop-gpu/vitest.config.ts` (line ~14)
**Evidence**: `exclude: ["src/workers/**", "src/**/*.d.ts"]` — Workers are explicitly excluded from coverage. This means the Data Worker and Render Worker (the most critical code paths) have **zero test coverage** and can't be verified to meet the 80% target.
**Required Fix**: Remove `src/workers/**` from the coverage exclusion list. Write unit tests for the pure logic in each worker (SAB write functions, datagram processing, mock data generation) by extracting testable functions into separate modules that can be imported without the Worker global context.

### R-024 — WARNING: No unit tests for `services/alerts.ts`, `services/feedback.ts`, `services/query.ts`

**Evidence**: The `tests/` directory has no test files for the gRPC service wrappers. These are the cold-path functions used by the UI for alert streaming, feedback submission, and track queries.
**Required Fix**: Add unit tests that mock the ConnectRPC transport and verify:
- `startAlertStream` — emits correctly typed alerts to the signal
- `acknowledgeAlert` — calls the gRPC client with correct parameters
- `submitFeedback` — validates inputs and maps response
- `fetchTrackDetail` — maps proto response to `TrackDetail` type
- `searchTracks` — handles empty results and pagination

### R-025 — WARNING: Render Worker device-loss re-init re-creates OffscreenCanvas context

**File**: `web-cop-gpu/src/workers/render-worker.ts` (lines ~93-96)
**Evidence**: On device loss, `init(activeCanvas, activeSab)` is called again, which calls `initGPU(offscreen)` → `canvas.getContext("webgpu")`. But the WebGPU spec states that `getContext` returns the **same context** for a given canvas. If the context is already configured, calling `configure` again with a new device should work. However, the `device.lost` handler in `device.ts` (line ~38) registers ANOTHER `device.lost` listener on the new device, and the old listener from `render-worker.ts` is still active. This creates duplicate listeners on each device loss cycle.
**Required Fix**: Ensure `device.lost` is only registered once per init cycle. Move the `device.lost` handler registration to the Render Worker's `init()` function only and do not register it again in `device.ts`.

### R-026 — WARNING: `TrackRecord` WGSL struct size may not match 128 bytes

**File**: `web-cop-gpu/src/shaders/interpolation.wgsl` (lines ~10-23)
**Evidence**: The `TrackRecord` struct has:
- 5 × f32 (20 bytes) + 7 × u32 (28 bytes) + 1 × array<vec4<f32>, 5> (80 bytes) = 128 bytes
- WGSL alignment: `f32` fields are 4-byte aligned, `u32` fields are 4-byte aligned, `vec4<f32>` is 16-byte aligned. The `trail` field starts at offset `48` raw, but WGSL may pad to align `array<vec4<f32>,5>` to 16 bytes. Since `48` is already 16-byte aligned (48 = 3 × 16), this works. Total = 48 + 80 = 128 bytes. **This is correct.**
**No fix needed** — noting this was verified for audit trail.

### R-027 — WARNING: No integration test for Go serializer → Wasm decoder round-trip

**Evidence**: The Go serializer tests (`pkg/flatbuf/serializer_test.go`) verify Go → binary record, and the Rust tests verify binary record → field accessors. But there is no **cross-language round-trip test** that verifies the Go-serialized bytes are correctly decoded by the Rust Wasm module at matching offsets.
**Required Fix**: Add a test in `tests/integration/` that:
1. Serializes a known TrackUpdate via Go `pkg/flatbuf/Serializer.Serialize()`
2. Writes the 128-byte output to a file
3. Has a Rust test that reads the same file and verifies all field values match
This catches offset misalignment between Go and Rust.

### R-028 — WARNING: E2E `security-audit.spec.ts` may have uncovered gaps

**File**: `web-cop-gpu/e2e/security-audit.spec.ts`
**Evidence**: Not yet reviewed in detail. Ensure this test file verifies:
- CSP headers are present and correct
- No `eval()` or `Function()` calls in the bundle
- No hardcoded URLs to external services
- Wasm module integrity
**Required Fix**: Review and run the security audit E2E test. Fix any failures.

---

## Non-Blocking Observations (Address Before Next Release)

### O-001 — Inline styles throughout all components

All UI components use inline `style={{}}` objects. While this works, it bypasses any future design system or CSS extraction. Consider migrating to CSS modules or a utility-class system in a follow-up.

### O-002 — `alertFlags` in mock data: only ~5% of tracks have alerts

**File**: `web-cop-gpu/src/gpu/mock-data.ts` (line ~53)
**Evidence**: `alertFlags: Math.random() < 0.05 ? 1 : 0` — only 5% of mock tracks have alert halos. For demo purposes, consider increasing this to ~15-20% to make the halo animation more visually impactful.

### O-003 — Classification banner rendered only at top, not bottom

**File**: `web-cop-gpu/src/components/shell/AppShell.tsx`
**Evidence**: The `AppShell` renders `<ClassificationBanner />` at the top but the bottom banner is missing. The SDLC security classification standard requires classification banners at **both** top and bottom of the viewport.
**Required Fix**: Add a second `<ClassificationBanner />` at the bottom of the AppShell, outside the main content row.

### O-004 — `FeedbackForm` uses `FEEDBACK_OPTIONS.map()` (creates new array each render)

**File**: `web-cop-gpu/src/components/panels/FeedbackForm.tsx` (line ~113)
**Evidence**: `.map()` inside JSX creates a new array on each render. Use `<For each={FEEDBACK_OPTIONS}>` for better SolidJS performance since `FEEDBACK_OPTIONS` is a static array.

### O-005 — DegradedNotice `missing.map()` should use `<For>`

**File**: `web-cop-gpu/src/components/DegradedNotice.tsx` (line ~41)
**Evidence**: Same pattern as O-004 — the `.map()` in JSX should use SolidJS `<For>`.

### O-006 — RoleSelector `ROLES.map()` should use `<For>`

**File**: `web-cop-gpu/src/components/toolbar/RoleSelector.tsx` (line ~46)
**Evidence**: Same `.map()` in JSX pattern.

---

## Fix Execution Order — Step-by-Step for AI Developer Agent

### Step 1: Batch 1 (R-001 through R-004)
- Add all missing `data-testid` attributes to components
- **Test**: `cd web-cop-gpu && pnpm test` (unit tests should still pass)
- **Test**: `cd web-cop-gpu && pnpm test:e2e` (E2E should find elements)

### Step 2: Batch 2 (R-005 through R-008)
- Fix the time base mismatch between `performance.now()` and `Date.now()`
- Consolidate mock data writes to a single worker (Data Worker)
- Update trail data on tick
- **Test**: `cd web-cop-gpu && pnpm test` (unit tests)
- **Verify**: Open dev server (`pnpm dev`), confirm tracks appear and move smoothly

### Step 3: Batch 3 (R-009 through R-010)
- Wire LOD system into render loop
- Wire frame timer (at least JS-side timing)
- **Test**: Open dev server, zoom in/out and verify trails/labels appear/disappear
- **Verify**: StatusBar shows accurate FPS

### Step 4: Batch 4 (R-011 through R-013)
- Fix pick buffer readback race and single-pixel copy
- **Test**: `pnpm dev`, click on a track icon, verify the TrackDetailPanel opens
- **Test**: Add unit test for pick buffer coordinate mapping

### Step 5: Batch 5 (R-014 through R-017)
- Wire WebTransport URL env var into Data Worker init
- Implement JWT token acquisition for WebTransport
- Fix Wasm module bundling for production build
- **Test**: `pnpm build` — verify Wasm module is in `dist/`
- **Test**: With backend running (`scripts/cop-dev/start-backend.sh`), confirm WebTransport connects

### Step 6: Batch 6 (R-018 through R-022)
- Wire alert flags in Go serializer
- Fix threat level colour mappings
- Remove/gate console.log calls
- Wire operator identity from session
- **Test**: `go test ./pkg/flatbuf/... -race`
- **Test**: `go test ./pkg/webtransport/... -race`
- **Test**: `pnpm lint` (no console.log warnings)

### Step 7: Batch 7 (R-023 through R-028)
- Remove workers from coverage exclusion, add worker logic tests
- Add gRPC service wrapper tests
- Fix device-loss duplicate listener
- Add Go→Rust round-trip integration test
- Review and run security E2E test
- **Test**: `pnpm test:coverage` — verify ≥80% line coverage
- **Test**: `go test ./... -race -coverprofile=coverage.out`

### Step 8: Non-blocking observations (O-001 through O-006)
- Add bottom classification banner
- Fix `.map()` → `<For>` in SolidJS components
- Increase mock alert percentage for demo impact
- **Test**: `pnpm test && pnpm test:e2e`

---

## End-to-End Demo Verification Checklist

After all batches are resolved, verify the following demo scenarios work:

| # | Scenario | Verification |
|---|----------|-------------|
| 1 | Cold boot in Chrome/Edge | App loads, classification banner top/bottom, capability gate passes |
| 2 | 50k mock tracks render at 60 FPS | StatusBar shows ≥55 FPS, track icons visible on dark background |
| 3 | Track selection via click | Click a track icon → TrackDetailPanel opens with hash → gRPC fetch |
| 4 | Alert halos pulse on alerted tracks | ~5% of tracks show pulsing coloured circles |
| 5 | Trail lines behind moving tracks | Trails fade from recent (bright) to old (dim) |
| 6 | Role switching | Switch to Ops Commander → alert sidebar appears |
| 7 | Alert acknowledgement | Click Ack on an alert → UI updates optimistically |
| 8 | Feedback submission | Select track → Submit Feedback → form closes after success |
| 9 | Ctrl+K search | Opens search overlay, type track ID prefix → results appear |
| 10 | Degraded mode | Disable WebGPU in DevTools → degraded notice with missing features |
| 11 | Backend WebTransport (with `start-backend.sh`) | Connection indicator shows green, latency displayed |
| 12 | Connection loss recovery | Block WebTransport → indicator shows red → restore → reconnects |
| 13 | LOD zoom levels | Zoom out → trails/labels disappear → zoom in → they reappear |
| 14 | Security headers | DevTools Network tab shows COOP, COEP, CSP headers |
| 15 | Run all demo scripts | `bash scripts/demo/run-demo.sh maritime --seed` completes |

---

## Summary Statistics

| Category | Count |
|----------|-------|
| BLOCKING issues | 10 |
| WARNING issues | 18 |
| Non-blocking observations | 6 |
| Total items | 34 |

**Verdict**: The implementation cannot be merged or demonstrated in its current state due to 10 BLOCKING issues. The most critical are the missing `data-testid` attributes (E2E tests entirely broken), the time base mismatch (interpolation broken), and the missing WebTransport URL passthrough (real backend connection impossible).

**Positive findings**: The architecture is solid, buffer management is correct (zero per-frame allocation), the WGSL shaders are well-structured with consistent struct layouts, the SAB ring buffer implementation is clean, and the Go backend (WebTransport server, FlatBuffer serializer, authentication, priority shedding) is production-quality.
