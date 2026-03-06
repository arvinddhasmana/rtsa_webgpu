// CLASSIFICATION: UNCLASSIFIED
// src/gpu/atlas.ts — Icon and SDF font atlas GPU texture management
//
// Atlas textures are immutable at runtime — baked at build time.
// At init, placeholder textures (solid colour) are created and uploaded to
// the GPU; in production these would be replaced with the actual baked assets.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md §6

/** Icon atlas dimensions: 2048×2048 RGBA (16 MB VRAM). */
export const ICON_ATLAS_WIDTH  = 2048;
export const ICON_ATLAS_HEIGHT = 2048;

/** SDF font atlas dimensions: 2048×1024 R8 (2 MB VRAM). */
export const SDF_ATLAS_WIDTH  = 2048;
export const SDF_ATLAS_HEIGHT = 1024;

export interface AtlasTextures {
  iconAtlas: GPUTexture;
  sdfAtlas:  GPUTexture;
  iconSampler: GPUSampler;
  sdfSampler:  GPUSampler;
}

/**
 * Create GPU textures for the icon atlas and SDF font atlas.
 *
 * In Phase 1, placeholder solid-colour data fills the atlas so the pipeline
 * can be exercised end-to-end. Phase 4 hardening will replace this with the
 * actual baked NATO APP-6 icon atlas and msdf-atlas-gen output.
 *
 * Rule: textures are created once, never regenerated at runtime
 * (webgpu_guidelines.md §6.3, rule 3).
 */
export function createAtlasTextures(device: GPUDevice): AtlasTextures {
  // --- Icon Atlas (RGBA8Unorm) ---
  const iconAtlas = device.createTexture({
    label:  "icon-atlas",
    size:   { width: ICON_ATLAS_WIDTH, height: ICON_ATLAS_HEIGHT },
    format: "rgba8unorm",
    usage:  GPUTextureUsage.TEXTURE_BINDING | GPUTextureUsage.COPY_DST,
  });

  // Upload placeholder data: white with full alpha → all icons show as white squares
  const iconPixels = new Uint8Array(ICON_ATLAS_WIDTH * ICON_ATLAS_HEIGHT * 4).fill(255);
  device.queue.writeTexture(
    { texture: iconAtlas },
    iconPixels,
    { bytesPerRow: ICON_ATLAS_WIDTH * 4 },
    { width: ICON_ATLAS_WIDTH, height: ICON_ATLAS_HEIGHT },
  );

  // --- SDF Atlas (R8Unorm) ---
  const sdfAtlas = device.createTexture({
    label:  "sdf-atlas",
    size:   { width: SDF_ATLAS_WIDTH, height: SDF_ATLAS_HEIGHT },
    format: "r8unorm",
    usage:  GPUTextureUsage.TEXTURE_BINDING | GPUTextureUsage.COPY_DST,
  });

  // Placeholder SDF data: value 0.75 → field value above the 0.5 threshold, renders as solid
  const sdfPixels = new Uint8Array(SDF_ATLAS_WIDTH * SDF_ATLAS_HEIGHT).fill(192); // 0.75 × 255
  device.queue.writeTexture(
    { texture: sdfAtlas },
    sdfPixels,
    { bytesPerRow: SDF_ATLAS_WIDTH },
    { width: SDF_ATLAS_WIDTH, height: SDF_ATLAS_HEIGHT },
  );

  // --- Samplers ---
  const iconSampler = device.createSampler({
    label:        "icon-sampler",
    magFilter:    "linear",
    minFilter:    "linear",
    mipmapFilter: "linear",
    addressModeU: "clamp-to-edge",
    addressModeV: "clamp-to-edge",
  });

  const sdfSampler = device.createSampler({
    label:        "sdf-sampler",
    magFilter:    "linear",
    minFilter:    "linear",
    mipmapFilter: "linear",
    addressModeU: "clamp-to-edge",
    addressModeV: "clamp-to-edge",
  });

  return { iconAtlas, sdfAtlas, iconSampler, sdfSampler };
}

/**
 * Destroy atlas GPU textures. Call on device loss before re-initialisation.
 */
export function destroyAtlasTextures(textures: AtlasTextures): void {
  textures.iconAtlas.destroy();
  textures.sdfAtlas.destroy();
}
