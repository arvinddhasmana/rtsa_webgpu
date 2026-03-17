// CLASSIFICATION: UNCLASSIFIED
// src/gpu/map-tiles.ts — Raster map tile background layer
//
// Manages a tile cache and renders pre-fetched raster tiles as the first
// layer in every frame. Tiles are fetched and uploaded to GPU textures by the
// Render Worker (not the main thread).
//
// In Phase 1, a solid dark background colour is rendered as a placeholder.
// Full tile pyramid fetching and rendering is deferred to Phase 3.
//
// Reference: docs/implementation/v4/phase1_core_rendering.md R1-10

/** Tile configuration */
export const MAP_BACKGROUND_COLOR: GPUColorDict = { r: 0.10, g: 0.15, b: 0.20, a: 1.0 };

/**
 * Create the render pass descriptor for the background (map tile) layer.
 *
 * The background pass uses `loadOp: "clear"` with the map background colour,
 * which clears the canvas before any track layers are drawn.
 * Subsequent render passes use `loadOp: "load"` to composite on top.
 *
 * Reference: webgpu_guidelines.md §5.3
 */
export function makeBackgroundPassDescriptor(
  colorAttachmentView: GPUTextureView,
  mapStyle: number = 0,
): GPURenderPassDescriptor {
  const clearColor = mapStyle === 1
    ? { r: 0.02, g: 0.04, b: 0.08, a: 1.0 } // HD: Deep Ocean/Black
    : MAP_BACKGROUND_COLOR; // Standard: Dark Slate

  return {
    label: "background-pass",
    colorAttachments: [
      {
        view:       colorAttachmentView,
        clearValue: clearColor,
        loadOp:     "clear",
        storeOp:    "store",
      },
    ],
  };
}

import { BindGroups } from "./bind-groups";
import { AllPipelines } from "./pipelines";

/**
 * Render the background layer (Phase 1: solid colour placeholder, Phase 5: procedural topography).
 */
export function renderBackground(
  encoder: GPUCommandEncoder,
  colorAttachmentView: GPUTextureView,
  pipelines: AllPipelines,
  bindGroups: BindGroups,
): void {
  const pass = encoder.beginRenderPass({
    label: "background-pass",
    colorAttachments: [
      {
        view:       colorAttachmentView,
        loadOp:     "clear",
        clearValue: { r: 0, g: 0, b: 0, a: 1 },
        storeOp:    "store",
      },
    ],
  });

  pass.setPipeline(pipelines.render.background);
  pass.setBindGroup(0, bindGroups.background.g0);
  pass.draw(4);
  pass.end();
}
