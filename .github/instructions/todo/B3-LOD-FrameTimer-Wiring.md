<!-- CLASSIFICATION: UNCLASSIFIED -->

# B3 — LOD System and Frame Timer: Dead Code Wiring

> **Batch**: B3 of 7
> **Theme**: LOD FrameTimer Wiring
> **Priority**: WARNING — LOD and frame timing exist but are never called; GPU overhead not managed
> **Agent Profile**: `greatest-ever-dev-forworkflow`
> **Source**: v4 Implementation Review R-009 through R-010

---

## Context

Two fully-implemented subsystems — the Level-of-Detail (LOD) system (`web-cop-gpu/src/gpu/lod.ts`) and the frame timer (`web-cop-gpu/src/gpu/frame-timer.ts`) — are dead code: they are never imported or invoked from the render loop. Without LOD, all 50,000 tracks render trail lines, halos, and labels at every zoom level, which can prevent the 60 FPS target. Without the frame timer, the StatusBar FPS display is estimated rather than measured.

---

## Issue R-009 — WARNING: LOD system never invoked from render loop

**Files to read first**:

- `web-cop-gpu/src/gpu/lod.ts` — understand `LodFlags` type and `computeLod(scale, trackCount)` signature
- `web-cop-gpu/src/gpu/renderer.ts` — locate the render pass dispatch functions
- `web-cop-gpu/src/workers/render-worker.ts` — locate where `renderer.ts` functions are called each frame

**Required Fix**:

1. Import `computeLod` and `LodFlags` from `web-cop-gpu/src/gpu/lod.ts` into `web-cop-gpu/src/gpu/renderer.ts` (or `render-worker.ts` — place it in whichever file has direct access to `state.camera.scale` and `state.trackCount`).

2. At the start of each frame (before dispatching GPU render passes), call:

   ```typescript
   const lod = computeLod(state.camera.scale, state.trackCount);
   ```

3. Use `lod.showTrails` to conditionally skip the trail render pass.
4. Use `lod.showHalos` to conditionally skip the halo render pass.
5. Use `lod.showLabels` to conditionally skip the label render pass.
6. Use `lod.maxInstances` to cap the instance count for the icon draw call (do not draw more than `lod.maxInstances` track icons per frame).

**Constraint**: Do not alter the LOD thresholds in `lod.ts` — only wire the call site. Read the `LodFlags` interface to understand the boolean flags and `maxInstances` field before editing.

---

## Issue R-010 — WARNING: Frame timer never invoked

**Files to read first**:

- `web-cop-gpu/src/gpu/frame-timer.ts` — understand the `FrameTimer` class API: `markJsStart()`, `markJsEnd()`, `getStats()`, and any GPU timestamp query methods
- `web-cop-gpu/src/workers/render-worker.ts` — locate the frame loop (`requestAnimationFrame` or `setInterval` callback) and the `stats` postMessage

**Required Fix**:

1. Instantiate `FrameTimer` once during Render Worker init (not per-frame to comply with the zero-per-frame-allocation rule):

   ```typescript
   const frameTimer = new FrameTimer();
   ```

2. In the frame loop:
   - Call `frameTimer.markJsStart()` at the beginning of the JS frame handler.
   - Call `frameTimer.markJsEnd()` immediately before submitting GPU commands.

3. GPU timestamp queries: enable them conditionally when the `timestamp-query` WebGPU feature is available on the device. Do not hard-fail if unavailable.

4. After computing frame stats, include timer data in the existing `stats` postMessage sent to the main thread so the StatusBar can display accurate FPS:

   ```typescript
   self.postMessage({ type: "stats", fps: frameTimer.getStats().fps, ... });
   ```

5. Do not allocate any new objects inside the frame loop. Ensure `FrameTimer` uses pre-allocated buffers internally (verify in `frame-timer.ts` before wiring).

---

## Implementation Rules

1. Read every target file before editing.
2. LOD and FrameTimer must be instantiated once at init, not per frame.
3. Zero per-frame heap allocations — verify no `new` expressions are added in the render loop.
4. All files must retain the `// CLASSIFICATION: UNCLASSIFIED` header.
5. Follow `docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md` and `wgsl_shader_standards.md`.

---

## Unit Tests Required

- In `web-cop-gpu/src/gpu/lod.test.ts` (create if it doesn't exist):
  - Test that `computeLod` with a small `scale` (fully zoomed out) returns `showTrails = false` and `showLabels = false`.
  - Test that `computeLod` with a large `scale` (fully zoomed in) returns `showTrails = true` and `showLabels = true`.
  - Test that `maxInstances` is always ≤ `trackCount`.

- In `web-cop-gpu/src/gpu/frame-timer.test.ts` (create if it doesn't exist):
  - Test that `markJsStart()` followed by `markJsEnd()` and `getStats()` returns a positive `fps` or `frameDurationMs` value.

---

## Verification Steps

1. `cd web-cop-gpu && pnpm test` — all unit tests must pass.
2. `cd web-cop-gpu && pnpm dev` — open the browser:
   - Zoom fully out (camera scale minimum) → trail lines and labels must disappear from the canvas.
   - Zoom fully in (camera scale maximum) → trail lines and labels must reappear.
   - The StatusBar FPS counter must reflect measured frame timing, not a hardcoded estimate.
3. Open DevTools → Performance tab — confirm no large per-frame GC spikes introduced by this batch.

---

## PR Instructions

- PR title: `feat(web-cop-gpu): wire LOD system and FrameTimer into render loop (B3)`
- Label: `ai-orchestrator`
- After the PR is merged, move this file from `.github/instructions/todo/B3-LOD-FrameTimer-Wiring.md` to `.github/instructions/done/B3-LOD-FrameTimer-Wiring.md`.
