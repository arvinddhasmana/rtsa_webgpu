// CLASSIFICATION: UNCLASSIFIED
// src/gpu/renderer.ts — Per-frame render pipeline orchestration
//
// Implements the full 10-step per-frame sequence:
//   1. Read SAB → writeBuffer (track data upload)
//   2. Compute: interpolation
//   3. Compute: culling (resets draw_args.instance_count to 0)
//   4. Render: background (map tiles — clear)
//   5. Render: trail lines (loadOp: load)
//   6. Render: track icons (loadOp: load)
//   7. Render: alert halos (loadOp: load)
//   8. Render: SDF labels  (loadOp: load)
//   9. Render: pick buffer  (separate target)
//  10. Present frame (submit command buffer)
//
// Reference: docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md §5.1

import { RECORD_SIZE, TRACK_DATA_OFFSET } from "../services/sab";
import { AtlasTextures } from "./atlas";
import { BindGroups } from "./bind-groups";
import { GPUBuffers, MAX_TRACKS } from "./buffers";
import { CoverageManager } from "./coverage";
import { FrameTimer } from "./frame-timer";
import { computeLod } from "./lod";
import { renderBackground } from "./map-tiles";
import { PickResources } from "./pick";
import { AllPipelines } from "./pipelines";
import { makeViewProjection, writeUniforms } from "./uniforms";

/** Pre-allocated draw args reset data — zero per-frame heap alloc.
 *  vertex_count=4 for quad-strip icons, halos, pick; instance_count reset to 0 before culling. */
const DRAW_ARGS_RESET = new Uint32Array([
  4, // vertex_count (quad strip: 4 verts)
  0, // instance_count (reset to 0 before each culling pass)
  0, // first_vertex
  0, // first_instance
]);

export interface RenderState {
  device:     GPUDevice;
  context:    GPUCanvasContext;
  format:     GPUTextureFormat;
  buffers:    GPUBuffers;
  bindGroups: BindGroups;
  pipelines:  AllPipelines;
  atlas:      AtlasTextures;
  pick:       PickResources;
  sab:        SharedArrayBuffer;
  canvas:     OffscreenCanvas;

  /** Current number of live tracks (read from SAB header each frame). */
  trackCount: number;

  /** Current active dashboard mode */
  dashboard: "sensor" | "commander" | "analytics" | "health" | "coverage";

  /** Camera state for view-projection matrix */
  camera: {
    centerLon: number;
    centerLat: number;
    scale:     number;
  };

  /** Frame timing */
  lastFrameTime: number;

  /** Number of raw observations uploaded this frame. */
  observationCount: number;

  /** Frame timer for GPU timestamp queries and JS wall-clock measurement. */
  frameTimer: FrameTimer;

  /** Sensor coverage and gap renderer */
  coverage: CoverageManager;

  /** Selection state for pathing and drill-down */
  selectedTrackIdHash: number;

  /** Map style: 0: Standard, 1: HD */
  mapStyle: 0 | 1;
}

/**
 * Execute one complete render frame.
 *
 * Zero per-frame allocation: all buffers, bind groups, and pipelines
 * are pre-allocated at init. Only `writeBuffer` is called to update data.
 *
 * Reference: webgpu_guidelines.md §5.1 (per-frame pipeline)
 */
let _frameNum = 0;

export function renderFrame(state: RenderState): void {
  const { device, context, buffers, bindGroups, pipelines, pick, sab } = state;

  const now = Date.now() & 0xffffffff;
  state.lastFrameTime = now;
  _frameNum++;

  // Push error scope on first few frames to detect silent GPU validation errors
  const debugFrame = import.meta.env.DEV && _frameNum <= 3;
  if (debugFrame) device.pushErrorScope('validation');

  // 1. Read active_track_count from SAB header (Atomics.load for safety)
  const header = new Uint32Array(sab, 0, 4096 / 4);
  const trackCount = Math.min(
    Atomics.load(header, 0), // HEADER_OFFSET_ACTIVE_TRACK_COUNT
    MAX_TRACKS,
  );
  state.trackCount = trackCount;

  // Compute LOD flags for this frame based on camera scale and track count.
  // Must happen before render pass dispatch so conditional passes are skipped correctly.
  const lod = computeLod(state.camera.scale, trackCount);

  // 2. Upload track data from SAB to GPU storage buffer (one writeBuffer per frame)
  // SharedArrayBuffer is accepted by writeBuffer per the WebGPU spec;
  // the type assertion works around an over-narrow @webgpu/types definition.
  const sabTrackView = new Uint8Array(sab, TRACK_DATA_OFFSET, trackCount * RECORD_SIZE);
  device.queue.writeBuffer(
    buffers.trackStorage,
    0,
    // @ts-expect-error WebGPU fully supports SharedArrayBuffer views
    sabTrackView,
    0,
    trackCount * RECORD_SIZE,
  );

  // Build and upload uniforms (view-projection, viewport, time, track_count)
  const viewProj = makeViewProjection(
    state.canvas.width,
    state.canvas.height,
    state.camera.centerLon,
    state.camera.centerLat,
    state.camera.scale,
  );
  writeUniforms(
    device,
    buffers.uniform,
    viewProj,
    state.canvas.width,
    state.canvas.height,
    now,
    trackCount,
    state.dashboard,
    state.selectedTrackIdHash,
    state.mapStyle,
  );

  // Reset indirect draw args (instance_count = 0) before culling
  device.queue.writeBuffer(buffers.drawArgs, 0, DRAW_ARGS_RESET);

  // Obtain the current swap chain texture
  const colorTexture = context.getCurrentTexture();
  const colorView    = colorTexture.createView({ label: "color-view" });

  // Build command buffer
  const encoder = device.createCommandEncoder({ label: "frame-encoder" });

  // 3. Compute: Interpolation
  if (trackCount > 0) {
    const workgroups = Math.ceil(trackCount / 256);
    const pass = encoder.beginComputePass({ label: "interpolation-pass" });
    pass.setPipeline(pipelines.compute.interpolation);
    pass.setBindGroup(0, bindGroups.interpolation.g0);
    pass.setBindGroup(1, bindGroups.interpolation.g1);
    pass.dispatchWorkgroups(workgroups);
    pass.end();
  }

  // 4. Compute: Culling (writes visible_indices and draw_args.instance_count)
  if (trackCount > 0) {
    const workgroups = Math.ceil(trackCount / 256);
    const pass = encoder.beginComputePass({ label: "culling-pass" });
    pass.setPipeline(pipelines.compute.culling);
    pass.setBindGroup(0, bindGroups.culling.g0);
    pass.setBindGroup(1, bindGroups.culling.g1);
    pass.dispatchWorkgroups(workgroups);
    pass.end();
  }

  // 5. Render: background (loadOp: clear)
  renderBackground(encoder, colorView);

  // 5.1 Render: Sensor Coverage (loadOp: load)
  {
    const pass = encoder.beginRenderPass({
      label: "coverage-pass",
      colorAttachments: [{
        view: colorView,
        loadOp: "load",
        storeOp: "store",
      }],
    });
    // Use the global uniform bind group (set 0)
    state.coverage.draw(pass, bindGroups.coverage.g0);
    pass.end();
  }

  // 5.2 Render: Raw Observations (loadOp: load)
  if (state.observationCount > 0) {
    const pass = encoder.beginRenderPass({
      label: "observations-pass",
      colorAttachments: [{
        view:    colorView,
        loadOp:  "load",
        storeOp: "store",
      }],
    });
    pass.setPipeline(pipelines.render.observations);
    pass.setBindGroup(0, bindGroups.observations.g0);
    pass.setBindGroup(1, bindGroups.observations.g1);
    // Draw 4 vertices per instance (quad), observationCount instances
    pass.draw(4, state.observationCount, 0, 0);
    pass.end();
  }

  // 6. Render: trail lines (loadOp: load — composites on top of background)
  if (lod.renderTrails) {
    const pass = encoder.beginRenderPass({
      label: "trail-pass",
      colorAttachments: [{
        view:    colorView,
        loadOp:  "load",
        storeOp: "store",
      }],
    });
    pass.setPipeline(pipelines.render.trail);
    pass.setBindGroup(0, bindGroups.trail.g0);
    pass.setBindGroup(1, bindGroups.trail.g1);
    // 24 verts per instance (4 segments × 6 verts), trackCount instances
    pass.draw(24, trackCount, 0, 0);
    pass.end();
  }

  // 7. Render: track icons
  if (trackCount > 0) {
    const pass = encoder.beginRenderPass({
      label: "icons-pass",
      colorAttachments: [{
        view:    colorView,
        loadOp:  "load",
        storeOp: "store",
      }],
    });
    pass.setPipeline(pipelines.render.trackIcons);
    pass.setBindGroup(0, bindGroups.trackIcons.g0);
    pass.setBindGroup(1, bindGroups.trackIcons.g1);
    pass.drawIndirect(buffers.drawArgs, 0);
    pass.end();
  }

  // 8. Render: alert halos
  if (lod.renderHalos) {
    const pass = encoder.beginRenderPass({
      label: "halos-pass",
      colorAttachments: [{
        view:    colorView,
        loadOp:  "load",
        storeOp: "store",
      }],
    });
    pass.setPipeline(pipelines.render.halos);
    pass.setBindGroup(0, bindGroups.halos.g0);
    pass.setBindGroup(1, bindGroups.halos.g1);
    pass.drawIndirect(buffers.drawArgs, 0);
    pass.end();
  }

/*
  // 9. Render: SDF labels
  if (lod.renderLabels) {
    const pass = encoder.beginRenderPass({
      label: "labels-pass",
      colorAttachments: [{
        view:    colorView,
        loadOp:  "load",
        storeOp: "store",
      }],
    });
    pass.setPipeline(pipelines.render.labels);
    pass.setBindGroup(0, bindGroups.labels.g0);
    pass.setBindGroup(1, bindGroups.labels.g1);
    pass.setBindGroup(2, bindGroups.labels.g2);
    // Note: draw(4, 0) is used because glyph builder is deferred to Phase 3.
    // If enabled, this pass currently triggers a validation error on some drivers.
    pass.draw(4, 0, 0, 0);
    pass.end();
  }
*/

  // 10. Render: pick buffer (separate R32Uint target)
  {
    const pass = encoder.beginRenderPass({
      label: "pick-pass",
      colorAttachments: [{
        view:     pick.view,
        clearValue: { r: 0, g: 0, b: 0, a: 0 },
        loadOp:   "clear",
        storeOp:  "store",
      }],
    });
    pass.setPipeline(pipelines.render.pick);
    pass.setBindGroup(0, bindGroups.pick.g0);
    pass.setBindGroup(1, bindGroups.pick.g1);
    pass.drawIndirect(buffers.drawArgs, 0);
    pass.end();
  }

  // 11. Submit all commands and present
  state.frameTimer.resolveTimestamps(encoder);
  device.queue.submit([encoder.finish()]);

  // Pop error scope and log any GPU validation errors
  if (debugFrame) {
    void device.popErrorScope().then((err) => {
      if (err) {
        console.error(`[Renderer] GPU validation error (frame ${_frameNum}):`, err.message);
      } else if (_frameNum === 1) {
        console.log(`[Renderer] Frame 1 submitted with no GPU validation errors.`);
      }
    });
  }
}
