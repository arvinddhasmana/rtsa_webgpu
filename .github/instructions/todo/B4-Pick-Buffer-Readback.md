<!-- CLASSIFICATION: UNCLASSIFIED -->

# B4 — Pick Buffer Race Condition and Coordinate Scaling

> **Batch**: B4 of 7
> **Theme**: Pick Buffer Readback
> **Priority**: BLOCKING — Track selection fails intermittently due to GPU readback race
> **Agent Profile**: `greatest-ever-dev-forworkflow`
> **Source**: v4 Implementation Review R-011 through R-013

---

## Context

The pick buffer (`web-cop-gpu/src/gpu/pick.ts`) enables track selection by click: a GPU render pass writes track IDs into a half-resolution texture, and on click, the pixel at the cursor position is read back from the GPU to the CPU. Currently, this readback races with the render loop and copies the entire texture (up to 2 MB) for every click. This batch fixes the race condition and reduces the copy to a single pixel.

---

## Issue R-011 — BLOCKING: Pick buffer readback races with render loop

**File**: `web-cop-gpu/src/gpu/pick.ts` (function `readPickPixel`)

**Problem**: `readPickPixel` creates a new command encoder, copies the pick texture to the readback buffer, and submits it. Meanwhile the render loop (driven by `setInterval` or `requestAnimationFrame`) may submit the next frame's GPU commands while `mapAsync` is still pending. The readback buffer is also reused without checking whether it is still mapped from a previous request, which causes a WebGPU validation error.

**Required Fix**:

1. Add a guard at the top of `readPickPixel`:

   ```typescript
   if (pick.readbackBuffer.mapState !== "unmapped") {
     return null; // previous read still in flight — drop this click
   }
   ```

2. Implement a **double-buffered readback** approach: keep two readback buffers (`readbackBufferA`, `readbackBufferB`) and alternate which one is used per pick request. This prevents stalls when two clicks arrive before the first GPU read completes.

3. Alternatively (simpler, acceptable for v4): use the single-buffer guard above and log a warning in DEV mode when a click is dropped.

Choose option 1 (single-buffer guard) as the minimum fix. Option 2 (double-buffered) is preferred for correctness but either is acceptable.

---

## Issue R-012 — WARNING: Add coordinate comment (no code change required)

**File**: `web-cop-gpu/src/gpu/pick.ts` (line ~82)

**Action**: Add a comment above the coordinate calculation to document the pixel coordinate pipeline so future developers do not accidentally double-scale or remove the `/2` factor:

```typescript
// Pixel coordinate pipeline:
// e.clientX  →  × devicePixelRatio  →  canvasX (physical pixels)
// Pick texture is half-resolution: pickW = canvasWidth / 2
// Therefore: px = Math.floor(canvasX / 2)  ✓
const px = Math.floor(canvasX / 2);
const py = Math.floor(canvasY / 2);
```

This is a **comment-only change**. Do not modify any logic.

---

## Issue R-013 — WARNING: Pick readback copies entire texture instead of single pixel

**File**: `web-cop-gpu/src/gpu/pick.ts` (lines ~84–92)

**Problem**: The entire pick texture (`canvasWidth/2 × canvasHeight/2 × 4 bytes`, up to ~2 MB) is copied for every click. Only the single pixel at `(px, py)` is needed.

**Required Fix**: Change the `copyTextureToBuffer` call to copy only one pixel:

```typescript
encoder.copyTextureToBuffer(
  { texture: pick.texture, origin: { x: px, y: py, z: 0 }, mipLevel: 0 },
  { buffer: pick.readbackBuffer, offset: 0, bytesPerRow: 256 },
  { width: 1, height: 1, depthOrArrayLayers: 1 },
);
```

`bytesPerRow` must be at least 256 bytes (WebGPU alignment requirement). The readback buffer size can be reduced to 256 bytes; update the buffer allocation size if it was previously sized to the full texture.

After `mapAsync` resolves, read only the first 4 bytes:

```typescript
const data = new Uint32Array(pick.readbackBuffer.getMappedRange(0, 4));
const trackId = data[0];
pick.readbackBuffer.unmap();
```

---

## Implementation Rules

1. Read `web-cop-gpu/src/gpu/pick.ts` in full before editing.
2. Zero new per-frame allocations — the command encoder for pick readback is created on-demand per click (not per frame), which is acceptable.
3. The readback buffer must be sized to accommodate the `bytesPerRow` alignment (256 bytes minimum); update `GPUBufferDescriptor.size` accordingly if you reduce it.
4. All files must retain the `// CLASSIFICATION: UNCLASSIFIED` header.
5. Follow `docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md`.

---

## Unit Tests Required

- In `web-cop-gpu/src/gpu/pick.test.ts` (create if it doesn't exist):
  - Test that calling `readPickPixel` when the readback buffer is already mapped returns `null` (or equivalent "drop" result) without throwing.
  - Test coordinate mapping: for a canvas of `1920 × 1080` with DPR=2, a click at client `(480, 270)` should map to pick texture pixel `(480, 270)` (physical `960 × 540` / half-res = `480 × 270`).

---

## Verification Steps

1. `cd web-cop-gpu && pnpm test` — all unit tests must pass.
2. `cd web-cop-gpu && pnpm dev` — click rapidly on track icons to trigger concurrent pick requests. The TrackDetailPanel must open correctly (or silently drop the second click) — it must **not** throw a WebGPU validation error in the console.
3. Open DevTools → Performance — confirm no 2 MB buffer copies on every click after the fix.

---

## PR Instructions

- PR title: `fix(web-cop-gpu): fix pick buffer readback race and reduce copy to single pixel (B4)`
- Label: `ai-orchestrator`
- After the PR is merged, move this file from `.github/instructions/todo/B4-Pick-Buffer-Readback.md` to `.github/instructions/done/B4-Pick-Buffer-Readback.md`.
