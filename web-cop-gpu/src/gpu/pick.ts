// CLASSIFICATION: UNCLASSIFIED
// src/gpu/pick.ts — Pick buffer texture and async readback management
//
// Manages the secondary render target used for O(1) track selection.
// The pick texture is rendered in a separate render pass; on click events
// the pixel at (x, y) is read back asynchronously to identify the track.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md §7

export interface PickResources {
  /** The R32Uint pick texture. */
  texture: GPUTexture;
  /** View of the pick texture used as render attachment. */
  view: GPUTextureView;
  /**
   * Read-back buffer (MAP_READ | COPY_DST) for async CPU readback.
   * Sized to 256 bytes (single-pixel copy, minimum WebGPU row-pitch alignment).
   */
  readbackBuffer: GPUBuffer;
  /** Pixel width of the pick texture. */
  width: number;
  /** Pixel height of the pick texture. */
  height: number;
}

/**
 * Create the pick texture and its CPU readback buffer.
 * The pick texture is half the canvas resolution to save VRAM and bandwidth.
 *
 * Reference: webgpu_guidelines.md §7.2
 */
export function createPickResources(
  device: GPUDevice,
  canvasWidth: number,
  canvasHeight: number,
): PickResources {
  // Half resolution
  const w = Math.max(1, Math.floor(canvasWidth / 2));
  const h = Math.max(1, Math.floor(canvasHeight / 2));

  const texture = device.createTexture({
    label:  "pick-texture",
    size:   { width: w, height: h },
    format: "r32uint",
    usage:  GPUTextureUsage.RENDER_ATTACHMENT | GPUTextureUsage.COPY_SRC,
  });

  const view = texture.createView({ label: "pick-texture-view" });

  // Readback buffer: sized to 256 bytes — enough for a single-pixel copy.
  // WebGPU requires bytesPerRow >= 256; reading one pixel only ever needs 4 bytes,
  // so 256 bytes satisfies the alignment constraint with minimal VRAM usage.
  const readbackBuffer = device.createBuffer({
    label: "pick-readback",
    size:  256,
    usage: GPUBufferUsage.MAP_READ | GPUBufferUsage.COPY_DST,
  });

  return { texture, view, readbackBuffer, width: w, height: h };
}

/**
 * Destroy pick resources. Call on canvas resize or device loss.
 */
export function destroyPickResources(pick: PickResources): void {
  pick.texture.destroy();
  pick.readbackBuffer.destroy();
}

/**
 * Issue a GPU→CPU copy of a single pick-texture pixel, then read back the
 * track_id_hash at the given canvas coordinates.
 *
 * This is async and must NOT block the render loop (webgpu_guidelines.md §7.2).
 * Returns 0 if the click hits the background (no track).
 * Returns null if a previous readback is still in flight (click is silently dropped).
 */
export async function readPickPixel(
  device: GPUDevice,
  pick: PickResources,
  canvasX: number,
  canvasY: number,
): Promise<number | null> {
  // R-011: Guard against re-entrancy — drop the click if the buffer is still mapped.
  if (pick.readbackBuffer.mapState !== "unmapped") {
    if (import.meta.env.DEV) {
      console.warn("[pick] readback buffer still in flight — click dropped");
    }
    return null;
  }

  // Pixel coordinate pipeline:
  // e.clientX  →  × devicePixelRatio  →  canvasX (physical pixels)
  // Pick texture is half-resolution: pickW = canvasWidth / 2
  // Therefore: px = Math.floor(canvasX / 2)  ✓
  const px = Math.floor(canvasX / 2);
  const py = Math.floor(canvasY / 2);

  if (px < 0 || py < 0 || px >= pick.width || py >= pick.height) {
    return 0;
  }

  // R-013: Copy only the single pixel at (px, py) instead of the full texture.
  // bytesPerRow must be >= 256 per WebGPU spec; copying 1 pixel (4 bytes) fits in 256.
  const encoder = device.createCommandEncoder({ label: "pick-readback-encoder" });
  encoder.copyTextureToBuffer(
    { texture: pick.texture, origin: { x: px, y: py, z: 0 }, mipLevel: 0 },
    { buffer: pick.readbackBuffer, offset: 0, bytesPerRow: 256 },
    { width: 1, height: 1, depthOrArrayLayers: 1 },
  );
  device.queue.submit([encoder.finish()]);

  await pick.readbackBuffer.mapAsync(GPUMapMode.READ);
  const data = new Uint32Array(pick.readbackBuffer.getMappedRange(0, 4));
  const trackIdHash = data[0] ?? 0;
  pick.readbackBuffer.unmap();

  return trackIdHash;
}

