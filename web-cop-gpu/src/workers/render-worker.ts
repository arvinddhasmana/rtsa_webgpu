// CLASSIFICATION: UNCLASSIFIED
// src/workers/render-worker.ts — Render Worker shell
//
// Responsibilities:
//   1. Accept OffscreenCanvas + SAB via postMessage
//   2. Run a 60 Hz render loop (16ms setInterval)
//   3. Read active_track_count from SAB header and log it (proves data flow)
//   4. WebGPU device acquisition deferred to Phase 1
//
// Reference: docs/implementation/v4/phase0_foundation.md F0-7

const RENDER_INTERVAL_MS = 16; // ~60 FPS

const HEADER_OFFSET_ACTIVE_TRACK_COUNT = 0; // uint32 index 0 → byte 0
const HEADER_OFFSET_WRITE_GENERATION = 1;   // uint32 index 1 → byte 4

interface InitMessage {
  type: "init";
  canvas: OffscreenCanvas;
  sab: SharedArrayBuffer;
}

let header: Uint32Array | null = null;
let renderIntervalId: ReturnType<typeof setInterval> | null = null;
let frameCount = 0;

function startRenderLoop(): void {
  if (renderIntervalId !== null) return;

  renderIntervalId = setInterval(() => {
    if (!header) return;

    frameCount++;
    const trackCount = Atomics.load(header, HEADER_OFFSET_ACTIVE_TRACK_COUNT);
    const generation = Atomics.load(header, HEADER_OFFSET_WRITE_GENERATION);

    // Log track count every 60 frames (~1 second) to prove data flow
    if (frameCount % 60 === 0) {
      console.log(
        `[RenderWorker] frame=${frameCount} tracks=${trackCount} gen=${generation}`,
      );
    }

    // TODO(Phase 1): WebGPU render pass goes here
  }, RENDER_INTERVAL_MS);
}

function stopRenderLoop(): void {
  if (renderIntervalId !== null) {
    clearInterval(renderIntervalId);
    renderIntervalId = null;
  }
}

self.addEventListener("message", (event: MessageEvent<InitMessage>) => {
  const msg = event.data;

  if (msg.type === "init") {
    // Store SAB header view (first 4096 bytes as Uint32Array)
    header = new Uint32Array(msg.sab, 0, 4096 / 4);

    // OffscreenCanvas received — WebGPU context will be acquired in Phase 1
    // For Phase 0 we just confirm receipt and start the loop
    console.log(
      `[RenderWorker] Initialised. Canvas: ${msg.canvas.width}×${msg.canvas.height}`,
    );

    startRenderLoop();
  }
});

// Clean up on termination
self.addEventListener("close", () => {
  stopRenderLoop();
});
