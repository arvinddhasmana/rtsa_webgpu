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

import { initGPU }           from "../gpu/device";
import { allocateBuffers, destroyBuffers } from "../gpu/buffers";
import { createPipelines }   from "../gpu/pipelines";
import { createBindGroups }  from "../gpu/bind-groups";
import { createPickResources, destroyPickResources, readPickPixel } from "../gpu/pick";
import { createAtlasTextures, destroyAtlasTextures } from "../gpu/atlas";
import { renderFrame, type RenderState } from "../gpu/renderer";
import { initMockTracks, writeMockTracksToSAB, tickMockTracks, MOCK_TRACK_COUNT } from "../gpu/mock-data";

/** Messages accepted by the Render Worker */
interface InitMessage {
  type: "init";
  canvas: OffscreenCanvas;
  sab: SharedArrayBuffer;
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

interface StatusMessage {
  type: "status";
  ready: boolean;
  error?: string;
}

type InboundMessage = InitMessage | ResizeMessage | SelectTrackMessage;

const RENDER_INTERVAL_MS = 16; // ~60 FPS

let renderState: RenderState | null = null;
let renderIntervalId: ReturnType<typeof setInterval> | null = null;
let lastTickTime = performance.now();
let canvas: OffscreenCanvas | null = null;
let sab: SharedArrayBuffer | null = null;

/**
 * Initialise the full WebGPU rendering stack.
 * On device loss, this is called again with the same canvas and SAB.
 */
async function init(offscreen: OffscreenCanvas, sabBuf: SharedArrayBuffer): Promise<void> {
  canvas = offscreen;
  sab    = sabBuf;

  // Tear down any existing state before re-init
  if (renderState) {
    stopRenderLoop();
    destroyAtlasTextures(renderState.pipelines as unknown as Parameters<typeof destroyAtlasTextures>[0]);
    destroyPickResources(renderState.pick);
    destroyBuffers(renderState.buffers);
  }

  const { device, context, format } = await initGPU(offscreen);

  // Re-init on device loss (not destroyed)
  device.lost.then((info) => {
    if (info.reason === "destroyed") return;
    console.warn(`[RenderWorker] Device lost (${info.reason}), re-initialising…`);
    stopRenderLoop();
    if (canvas && sab) {
      init(canvas, sab).catch((err: unknown) => {
        console.error("[RenderWorker] Re-init failed:", err);
      });
    }
  });

  // Allocate all GPU resources (zero per-frame allocation after this)
  const buffers    = allocateBuffers(device);
  const atlas      = createAtlasTextures(device);
  const pick       = createPickResources(device, offscreen.width, offscreen.height);
  const pipelines  = createPipelines(device, format);
  const bindGroups = createBindGroups(device, pipelines, buffers, atlas, pick);

  // Seed SAB with mock data (Phase 1 — until real WebTransport data arrives)
  initMockTracks(MOCK_TRACK_COUNT);
  writeMockTracksToSAB(sabBuf, MOCK_TRACK_COUNT);

  renderState = {
    device,
    context,
    format,
    buffers,
    bindGroups,
    pipelines,
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
  };

  startRenderLoop();

  const status: StatusMessage = { type: "status", ready: true };
  self.postMessage(status);

  console.log(
    `[RenderWorker] Initialised. Canvas: ${offscreen.width}×${offscreen.height}, ` +
    `tracks: ${MOCK_TRACK_COUNT}, format: ${format}`,
  );
}

function startRenderLoop(): void {
  if (renderIntervalId !== null) return;

  renderIntervalId = setInterval(() => {
    if (!renderState) return;

    const now = performance.now();
    const dt  = now - lastTickTime;
    lastTickTime = now;

    // Animate mock tracks and re-write to SAB
    tickMockTracks(dt);
    if (sab) {
      writeMockTracksToSAB(sab, MOCK_TRACK_COUNT);
    }

    try {
      renderFrame(renderState);
    } catch (err) {
      console.error("[RenderWorker] renderFrame error:", err);
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
      init(msg.canvas, msg.sab).catch((err: unknown) => {
        console.error("[RenderWorker] Init failed:", err);
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
        // Re-create pick resources for new size
        destroyPickResources(renderState.pick);
        renderState.pick = createPickResources(
          renderState.device,
          msg.width,
          msg.height,
        );
        // Re-configure canvas context
        renderState.context.configure({
          device: renderState.device,
          format: renderState.format,
          alphaMode: "premultiplied",
        });
      }
      break;
    }

    case "select_track": {
      if (renderState) {
        readPickPixel(renderState.device, renderState.pick, msg.x, msg.y)
          .then((trackIdHash) => {
            const picked: PickedMessage = {
              type: "picked",
              trackIdHash,
              x: msg.x,
              y: msg.y,
            };
            self.postMessage(picked);
          })
          .catch((err: unknown) => {
            console.error("[RenderWorker] Pick readback error:", err);
          });
      }
      break;
    }

    default:
      break;
  }
});

// Clean up on termination
self.addEventListener("close", () => {
  stopRenderLoop();
  if (renderState) {
    renderState.device.destroy();
  }
});
