// CLASSIFICATION: UNCLASSIFIED
// src/workers/render-worker.ts — Render Worker (Phase 1: Full WebGPU implementation)
//
// Responsibilities:
//   1. Accept OffscreenCanvas + SAB via postMessage
//   2. Initialise WebGPU device, buffers, pipelines, atlases
//   3. Seed the SAB with mock track data (Phase 1 — Phase 2 will provide real data)
//   4. Run a 60 Hz render loop executing the full GPU pipeline
//   5. Handle pick buffer read-back on "select_track" messages
//   6. Handle device loss with re-initialisation
//
// Reference: docs/implementation/v4/phase1_core_rendering.md
//            docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md

import { createAtlasTextures, destroyAtlasTextures } from "../gpu/atlas";
import { createBindGroups } from "../gpu/bind-groups";
import { allocateBuffers, destroyBuffers } from "../gpu/buffers";
import { initGPU } from "../gpu/device";
import { FrameTimer } from "../gpu/frame-timer";
import { initMockTracks, MOCK_TRACK_COUNT, tickMockTracks, writeMockTracksToSAB } from "../gpu/mock-data";
import { createPickResources, destroyPickResources, readPickPixel } from "../gpu/pick";
import { createPipelines } from "../gpu/pipelines";
import {
    computeFps,
    makeErrorStatus,
    RENDER_ERROR_THRESHOLD,
    RENDER_INTERVAL_MS,
    shouldFlushStats,
    type RenderStatusMessage,
} from "../gpu/render-logic";
import { renderFrame, type RenderState } from "../gpu/renderer";
import type { RenderStatsMessage } from "./shared-protocol";

/** Messages accepted by the Render Worker */
interface InitMessage {
  type: "init";
  canvas: OffscreenCanvas;
  sab: SharedArrayBuffer;
  /**
   * When true, the Data Worker is the sole SAB writer.
   * The Render Worker must NOT call mock-data functions in this mode.
   */
  dataWorkerActive?: boolean;
}

interface ResizeMessage {
  type: "resize";
  width: number;
  height: number;
}

interface SelectTrackMessage {
  type: "select_track";
  x: number;
  y: number;
}

/** Messages sent back to the main thread */
interface PickedMessage {
  type: "picked";
  trackIdHash: number;
  x: number;
  y: number;
}

// StatusMessage is re-exported from render-logic as RenderStatusMessage.
// Use the imported type alias for all status message creation.
type StatusMessage = RenderStatusMessage;

type InboundMessage = InitMessage | ResizeMessage | SelectTrackMessage;

// RENDER_INTERVAL_MS and RENDER_ERROR_THRESHOLD are imported from render-logic.ts.

let renderState: RenderState | null = null;
let renderIntervalId: ReturnType<typeof setInterval> | null = null;
let lastTickTime = performance.now();
let activeCanvas: OffscreenCanvas | null = null;
let activeSab: SharedArrayBuffer | null = null;
/** True when the Data Worker is the sole SAB writer — Render Worker skips mock-data. */
let activeDataWorker = false;
/** Stats postMessage throttle counter (reset every 60 frames ≈ 1 second). */
let statsCounter = 0;
/** Tracks consecutive renderFrame errors to detect a failed render pipeline. */
let renderFrameErrorCount = 0;

// makeErrorStatus is imported from render-logic.ts.

/**
 * Tear down existing GPU resources before re-initialisation.
 * All GPU objects become invalid after device loss.
 */
function teardown(): void {
  stopRenderLoop();
  renderFrameErrorCount = 0;
  if (renderState) {
    renderState.frameTimer.destroy();
    destroyAtlasTextures(renderState.atlas);
    destroyPickResources(renderState.pick);
    destroyBuffers(renderState.buffers);
    // Do NOT call device.destroy() here — on device loss the device is already gone.
    renderState = null;
  }
}

/**
 * Initialise the full WebGPU rendering stack.
 * On device loss this is called again with the same canvas and SAB.
 */
async function init(offscreen: OffscreenCanvas, sabBuf: SharedArrayBuffer, dataWorkerActive: boolean): Promise<void> {
  activeCanvas = offscreen;
  activeSab    = sabBuf;
  activeDataWorker = dataWorkerActive;

  teardown();

  const { device, context, format } = await initGPU(offscreen);

  // Instantiate FrameTimer once per device (zero per-frame allocations).
  // Enables GPU timestamp queries when the device supports the feature;
  // gracefully falls back to JS wall-clock timing when unsupported.
  const frameTimer = new FrameTimer(device);

  // Register device.lost handler exclusively here in init().
  // Do NOT register this handler anywhere else (e.g. in device.ts):
  // each call to init() produces a new GPUDevice whose lost promise is a
  // one-shot — it settles exactly once. Registering multiple .then() calls
  // on the same device's lost promise is harmless (each fires once), but
  // registering the handler in a shared module would accumulate O(N) handlers
  // across N re-init cycles. Keeping the registration here ensures exactly
  // one re-init callback per device, with no stale references.
  device.lost.then((info: GPUDeviceLostInfo) => {
    if (info.reason === "destroyed") return;
    if (import.meta.env.DEV) {
      console.warn(`[RenderWorker] Device lost (${info.reason}), re-initialising…`);
    }
    if (activeCanvas && activeSab) {
      init(activeCanvas, activeSab, activeDataWorker).catch((err: unknown) => {
        if (import.meta.env.DEV) {
          console.error("[RenderWorker] Re-init failed:", err);
        }
      });
    }
  });

  // Allocate all GPU resources (zero per-frame allocation after this)
  const buffers    = allocateBuffers(device);
  const atlas      = createAtlasTextures(device);
  const pick       = createPickResources(device, offscreen.width, offscreen.height);
  const pipelines  = createPipelines(device, format);
  const bindGroups = createBindGroups(device, pipelines, buffers, atlas, pick);

  // Seed SAB with mock data only when no Data Worker is providing data.
  // When activeDataWorker is true, the Data Worker is the sole SAB writer.
  if (!activeDataWorker) {
    initMockTracks(MOCK_TRACK_COUNT);
    writeMockTracksToSAB(sabBuf, MOCK_TRACK_COUNT);
  }

  renderState = {
    device,
    context,
    format,
    buffers,
    bindGroups,
    pipelines,
    atlas,
    pick,
    sab: sabBuf,
    canvas: offscreen,
    trackCount: MOCK_TRACK_COUNT,
    camera: {
      centerLon: 0,
      centerLat: 0,
      scale:     2.0,
    },
    lastFrameTime: performance.now(),
    frameTimer,
  };

  startRenderLoop();

  const status: StatusMessage = { type: "status", ready: true };
  self.postMessage(status);

  if (import.meta.env.DEV) {
    console.log(
      `[RenderWorker] Initialised. Canvas: ${offscreen.width}×${offscreen.height}, ` +
      `tracks: ${MOCK_TRACK_COUNT}, format: ${format}`,
    );
  }
}

function startRenderLoop(): void {
  if (renderIntervalId !== null) return;

  renderIntervalId = setInterval(() => {
    if (!renderState) return;

    // R-010: Mark JS frame start for wall-clock measurement.
    renderState.frameTimer.markJsStart();

    const now = performance.now();
    const dt  = now - lastTickTime;
    lastTickTime = now;

    // Animate mock tracks only when Render Worker is the sole SAB writer.
    // When the Data Worker is active it is the sole writer; the Render Worker
    // must only read from the SAB.
    if (!activeDataWorker) {
      tickMockTracks(dt);
      if (activeSab) {
        writeMockTracksToSAB(activeSab, MOCK_TRACK_COUNT);
      }
    }

    try {
      renderFrame(renderState);
      // Reset consecutive error counter on each successful frame.
      renderFrameErrorCount = 0;
    } catch (err) {
      renderFrameErrorCount++;
      if (import.meta.env.DEV) {
        console.error("[RenderWorker] renderFrame error:", err);
      }
      // After repeated failures the render pipeline is likely broken.
      // Stop the loop and notify the main thread so the failure is observable.
      if (renderFrameErrorCount >= RENDER_ERROR_THRESHOLD) {
        stopRenderLoop();
        const errMsg = err instanceof Error ? err.message : String(err);
        self.postMessage(makeErrorStatus(
          `Render pipeline failed after ${RENDER_ERROR_THRESHOLD} consecutive errors: ${errMsg}`,
        ));
        return;
      }
    }

    // R-010: Mark JS frame end (after GPU command submission inside renderFrame).
    const jsMs = renderState.frameTimer.markJsEnd();

    // R-010: Async GPU timestamp readback — non-blocking, updates smoothed averages.
    // No new allocations here; FrameTimer uses pre-allocated buffers internally.
    void renderState.frameTimer.readbackAsync(jsMs);

    // Throttle stats postMessage to once per second (~60 frames).
    statsCounter++;
    if (shouldFlushStats(statsCounter)) {
      statsCounter = 0;
      // Use actual frame-to-frame interval for fps — more accurate than JS work duration.
      const fps = computeFps(dt);
      const statsMsg: RenderStatsMessage = {
        type: "stats",
        fps,
        trackCount: renderState.trackCount,
        visibleCount: renderState.trackCount,
      };
      self.postMessage(statsMsg);
    }
  }, RENDER_INTERVAL_MS);
}

function stopRenderLoop(): void {
  if (renderIntervalId !== null) {
    clearInterval(renderIntervalId);
    renderIntervalId = null;
  }
}

self.addEventListener("message", (event: MessageEvent<InboundMessage>) => {
  const msg = event.data;

  switch (msg.type) {
    case "init": {
      init(msg.canvas, msg.sab, msg.dataWorkerActive ?? false).catch((err: unknown) => {
        if (import.meta.env.DEV) {
          console.error("[RenderWorker] Init failed:", err);
        }
        const status: StatusMessage = {
          type:  "status",
          ready: false,
          error: err instanceof Error ? err.message : String(err),
        };
        self.postMessage(status);
      });
      break;
    }

    case "resize": {
      if (renderState) {
        renderState.canvas.width  = msg.width;
        renderState.canvas.height = msg.height;
        // Re-create pick resources for new canvas size
        destroyPickResources(renderState.pick);
        renderState.pick = createPickResources(
          renderState.device,
          msg.width,
          msg.height,
        );
        // Re-configure canvas context for new size
        renderState.context.configure({
          device:    renderState.device,
          format:    renderState.format,
          alphaMode: "premultiplied",
        });
      }
      break;
    }

    case "select_track": {
      if (renderState) {
        readPickPixel(renderState.device, renderState.pick, msg.x, msg.y)
          .then((trackIdHash) => {
            if (trackIdHash === null) return;
            const picked: PickedMessage = {
              type: "picked",
              trackIdHash,
              x: msg.x,
              y: msg.y,
            };
            self.postMessage(picked);
          })
          .catch((err: unknown) => {
            if (import.meta.env.DEV) {
              console.error("[RenderWorker] Pick readback error:", err);
            }
            // Notify the main thread so the failure is observable in production.
            self.postMessage(makeErrorStatus(
              `Pick readback failed: ${err instanceof Error ? err.message : String(err)}`,
            ));
          });
      }
      break;
    }

    default:
      break;
  }
});

// Clean up on Worker termination
self.addEventListener("close", () => {
  stopRenderLoop();
  if (renderState) {
    renderState.frameTimer.destroy();
    destroyAtlasTextures(renderState.atlas);
    destroyPickResources(renderState.pick);
    destroyBuffers(renderState.buffers);
    renderState.device.destroy();
    renderState = null;
  }
});
