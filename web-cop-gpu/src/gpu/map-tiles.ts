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
export const MAP_BACKGROUND_COLOR: GPUColorDict = { r: 0.0, g: 0.0, b: 0.0, a: 0.0 };

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
): GPURenderPassDescriptor {
  return {
    label: "background-pass",
    colorAttachments: [
      {
        view:       colorAttachmentView,
        clearValue: MAP_BACKGROUND_COLOR,
        loadOp:     "clear",
        storeOp:    "store",
      },
    ],
  };
}

/**
 * Render the background layer (Phase 1: solid colour placeholder).
 *
 * The pass clears the canvas to the map background colour.
 * No draw calls are issued — the clear itself is the output.
 */
export function renderBackground(
  encoder: GPUCommandEncoder,
  colorAttachmentView: GPUTextureView,
): void {
  const pass = encoder.beginRenderPass(
    makeBackgroundPassDescriptor(colorAttachmentView),
  );
  // No draw calls — the clear provides the solid background
  pass.end();
}
