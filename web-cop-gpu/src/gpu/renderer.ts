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

import { GPUBuffers, MAX_TRACKS } from "./buffers";
import { BindGroups } from "./bind-groups";
import { AllPipelines } from "./pipelines";
import { PickResources } from "./pick";
import { AtlasTextures } from "./atlas";
import { renderBackground } from "./map-tiles";
import { writeUniforms, makeViewProjection } from "./uniforms";
import { TRACK_DATA_OFFSET, RECORD_SIZE } from "../services/sab";

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

  /** Camera state for view-projection matrix */
  camera: {
    centerLon: number;
    centerLat: number;
    scale:     number;
  };

  /** Frame timing */
  lastFrameTime: number;
}

/**
 * Execute one complete render frame.
 *
 * Zero per-frame allocation: all buffers, bind groups, and pipelines
 * are pre-allocated at init. Only `writeBuffer` is called to update data.
 *
 * Reference: webgpu_guidelines.md §5.1 (per-frame pipeline)
 */
export function renderFrame(state: RenderState): void {
  const { device, context, buffers, bindGroups, pipelines, pick, sab } = state;

  const now = Date.now() & 0xffffffff;
  state.lastFrameTime = now;

  // 1. Read active_track_count from SAB header (Atomics.load for safety)
  const header = new Uint32Array(sab, 0, 4096 / 4);
  const trackCount = Math.min(
    Atomics.load(header, 0), // HEADER_OFFSET_ACTIVE_TRACK_COUNT
    MAX_TRACKS,
  );
  state.trackCount = trackCount;

  if (trackCount === 0) return; // Nothing to render

  // 2. Upload track data from SAB to GPU storage buffer (one writeBuffer per frame)
  // SharedArrayBuffer is accepted by writeBuffer per the WebGPU spec;
  // the type assertion works around an over-narrow @webgpu/types definition.
  const sabTrackView = new Uint8Array(sab, TRACK_DATA_OFFSET, trackCount * RECORD_SIZE);
  device.queue.writeBuffer(
    buffers.trackStorage,
    0,
    sabTrackView.buffer as unknown as ArrayBuffer,
    sabTrackView.byteOffset,
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
  );

  // Reset indirect draw args (instance_count = 0) before culling
  device.queue.writeBuffer(buffers.drawArgs, 0, DRAW_ARGS_RESET);

  // Obtain the current swap chain texture
  const colorTexture = context.getCurrentTexture();
  const colorView    = colorTexture.createView({ label: "color-view" });

  // Build command buffer
  const encoder = device.createCommandEncoder({ label: "frame-encoder" });

  // 3. Compute: Interpolation
  {
    const workgroups = Math.ceil(trackCount / 256);
    const pass = encoder.beginComputePass({ label: "interpolation-pass" });
    pass.setPipeline(pipelines.compute.interpolation);
    pass.setBindGroup(0, bindGroups.interpolation.g0);
    pass.setBindGroup(1, bindGroups.interpolation.g1);
    pass.dispatchWorkgroups(workgroups);
    pass.end();
  }

  // 4. Compute: Culling (writes visible_indices and draw_args.instance_count)
  {
    const workgroups = Math.ceil(trackCount / 256);
    const pass = encoder.beginComputePass({ label: "culling-pass" });
    pass.setPipeline(pipelines.compute.culling);
    pass.setBindGroup(0, bindGroups.culling.g0);
    pass.setBindGroup(1, bindGroups.culling.g1);
    pass.dispatchWorkgroups(workgroups);
    pass.end();
  }

  // 5. Render: background (map tiles — loadOp: clear)
  renderBackground(encoder, colorView);

  // 6. Render: trail lines (loadOp: load — composites on top of background)
  {
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
  {
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
    pass.setBindGroup(2, bindGroups.trackIcons.g2);
    // 4 verts per instance, use indirect draw args from culling pass
    pass.drawIndirect(buffers.drawArgs, 0);
    pass.end();
  }

  // 8. Render: alert halos
  {
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

  // 9. Render: SDF labels
  {
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
    // Labels rendered using direct draw (trackCount glyph instances)
    // In Phase 1: 0 glyphs rendered (glyph CPU builder deferred to Phase 3)
    pass.draw(4, 0, 0, 0);
    pass.end();
  }

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
  device.queue.submit([encoder.finish()]);
}
