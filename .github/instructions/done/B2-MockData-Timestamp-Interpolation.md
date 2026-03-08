<!-- CLASSIFICATION: UNCLASSIFIED -->

# B2 — Demo Data Flow: Mock Data Timestamp & Interpolation Issues

> **Batch**: B2 of 7
> **Theme**: MockData Timestamp Interpolation
> **Priority**: BLOCKING — Dead-reckoning interpolation is broken; tracks appear frozen
> **Agent Profile**: `greatest-ever-dev-forworkflow`
> **Source**: v4 Implementation Review R-005 through R-008

---

## Context

The GPU interpolation shader in `web-cop-gpu/src/shaders/interpolation.wgsl` computes dead-reckoning by subtracting `update_epoch_ms` (written to the SharedArrayBuffer) from `current_time_ms` (supplied as a uniform). If these two timestamps use **different time bases**, the delta is a large negative number, it clamps to zero, and all tracks appear frozen at their last received position. This batch fixes the time-base mismatch and consolidates mock data writes to a single worker.

---

## Issue R-005 — BLOCKING: Time-base mismatch breaks interpolation

**Problem**: `web-cop-gpu/src/gpu/mock-data.ts` writes `update_epoch_ms` using `Date.now()` (Unix epoch ms, ~1.7 trillion), while `web-cop-gpu/src/workers/render-worker.ts` passes `performance.now() | 0` (ms since time origin, a few million) as `currentTimeMs` to the uniforms. The WGSL shader subtracts them and gets a massive negative delta that clamps to 0.

**Required Fix** (prefer option b to match the Go serializer which uses `time.Now().UnixMilli()`):

In `web-cop-gpu/src/workers/render-worker.ts`, find the line that builds the `currentTimeMs` value for `writeUniforms` and change it from:

```typescript
const now = performance.now() | 0;
```

to:

```typescript
const now = Date.now() & 0xffffffff;
```

Verify that every call to `writeUniforms` / `setCurrentTimeMs` in the render worker uses this `Date.now()` based value.

**Constraint**: The WGSL uniform struct declares `current_time_ms` as `u32`. `Date.now() & 0xFFFFFFFF` produces a safe unsigned 32-bit value with the same wrapping behaviour as the FlatBuffer `update_epoch_ms` field written by the Go serializer (`pkg/flatbuf/serializer.go`).

---

## Issue R-006 — BLOCKING: Render Worker overwrites Data Worker SAB writes every frame

**Files**:

- `web-cop-gpu/src/workers/render-worker.ts`
- `web-cop-gpu/src/workers/data-worker.ts`

**Problem**: In mock/dev mode the Render Worker calls `initMockTracks()` / `writeMockTracksToSAB()` / `tickMockTracks()` **every frame**, overwriting whatever the Data Worker wrote to the SharedArrayBuffer. This violates the Phase 2 architecture where the Data Worker is the sole SAB writer.

**Required Fix**:

1. Remove the `initMockTracks`, `writeMockTracksToSAB`, and `tickMockTracks` calls from the Render Worker's render loop (and init path) when the Data Worker is responsible for mock data.
2. The Render Worker must only **read** from the SAB; the Data Worker must be the **sole writer**.
3. Add a guard: only execute the Render Worker mock-data write path when the Data Worker is explicitly disabled (i.e., no `data-worker.ts` instance was created).

---

## Issue R-007 — WARNING: Trail ring buffer is never updated on tick

**File**: `web-cop-gpu/src/gpu/mock-data.ts`

**Problem**: Trail positions are written once at init (as static offsets from the initial position) and never updated as `tickMockTracks()` advances track positions. Trails display the wrong historical path.

**Required Fix**: In the `tickMockTracks()` function, after updating a track's `lat`, `lon`, and `heading` fields in the SAB, shift the trail ring buffer by one entry and write the **current** position as the newest trail segment. This mimics how real track updates from the backend push new positions into the ring buffer.

---

## Issue R-008 — WARNING: Data Worker mock mode writes only 1 track per 16 ms interval

**File**: `web-cop-gpu/src/workers/data-worker.ts`

**Problem**: `startMockUpdates()` writes exactly 1 mock record per `setInterval` tick. At 60 Hz this yields only 60 tracks/second — it would take ~14 minutes to populate 50,000 track slots.

**Required Fix**: Either:

- (a) Batch-write mock tracks (e.g., write 1,000 records per tick so the SAB reaches target occupancy within a few seconds), **or**
- (b) As recommended in R-006, consolidate mock data generation entirely to one location. If the Render Worker no longer writes mock data, the Data Worker batch approach becomes the definitive source.

Preferred: implement (b) — consolidate all mock data generation inside `data-worker.ts`, writing in batches of 1,000 tracks per tick.

---

## Implementation Rules

1. Read every file before editing. Do not refactor unrelated logic.
2. The WGSL shader (`interpolation.wgsl`) itself does **not** need to change — only the TypeScript time-base.
3. Verify that `Date.now() & 0xFFFFFFFF` is consistent with the `>>> 0` truncation used in `web-cop-gpu/src/gpu/mock-data.ts`. Both produce unsigned 32-bit values; they are equivalent for values under 2^32 ms (year 2106).
4. All files must retain the `// CLASSIFICATION: UNCLASSIFIED` header.
5. Follow `docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md` and `go_standards.md`.

---

## Unit Tests Required

- Add a unit test in `web-cop-gpu/src/gpu/` that:
  1. Calls `writeMockTracksToSAB()` or the equivalent Data Worker batch writer.
  2. Reads back `update_epoch_ms` from the SAB.
  3. Asserts that `Math.abs(Date.now() - readBackValue) < 2000` (within 2 seconds).
- Add a unit test for `tickMockTracks()` that verifies at least one trail entry changes after a tick.

---

## Verification Steps

1. `cd web-cop-gpu && pnpm test` — all unit tests must pass.
2. `cd web-cop-gpu && pnpm dev` — open the browser, observe tracks move smoothly between server update intervals (visible drift/interpolation of icon positions). Tracks must not appear frozen.
3. Open browser DevTools → Console — no errors about SAB write conflicts or race conditions.

---

## PR Instructions

- PR title: `fix(web-cop-gpu): fix time-base mismatch and consolidate mock SAB writes to data-worker (B2)`
- Label: `ai-orchestrator`
- After the PR is merged, move this file from `.github/instructions/todo/B2-MockData-Timestamp-Interpolation.md` to `.github/instructions/done/B2-MockData-Timestamp-Interpolation.md`.
