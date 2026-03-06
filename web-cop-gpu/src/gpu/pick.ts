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
  /** Read-back buffer (MAP_READ | COPY_DST) for async CPU readback. */
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

  // Readback buffer: size = w × h × 4 bytes (one u32 per pixel).
  // Row pitch must be aligned to 256 bytes per WebGPU spec.
  const bytesPerRow = alignTo256(w * 4);
  const readbackBuffer = device.createBuffer({
    label: "pick-readback",
    size:  bytesPerRow * h,
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
 * Issue a GPU→CPU copy of the pick texture, then read back the track_id_hash
 * at the given canvas coordinates.
 *
 * This is async and must NOT block the render loop (webgpu_guidelines.md §7.2).
 * Returns 0 if the click hits the background (no track).
 */
export async function readPickPixel(
  device: GPUDevice,
  pick: PickResources,
  canvasX: number,
  canvasY: number,
): Promise<number> {
  // Scale from canvas coordinates to pick texture coordinates
  const px = Math.floor((canvasX / 2));
  const py = Math.floor((canvasY / 2));

  if (px < 0 || py < 0 || px >= pick.width || py >= pick.height) {
    return 0;
  }

  const bytesPerRow = alignTo256(pick.width * 4);

  const encoder = device.createCommandEncoder({ label: "pick-readback-encoder" });
  encoder.copyTextureToBuffer(
    { texture: pick.texture },
    { buffer: pick.readbackBuffer, bytesPerRow, rowsPerImage: pick.height },
    { width: pick.width, height: pick.height },
  );
  device.queue.submit([encoder.finish()]);

  await pick.readbackBuffer.mapAsync(GPUMapMode.READ);
  const data = new Uint32Array(pick.readbackBuffer.getMappedRange());
  const rowStride = bytesPerRow / 4;
  const trackIdHash = data[py * rowStride + px] ?? 0;
  pick.readbackBuffer.unmap();

  return trackIdHash;
}

/** Align a byte count to the next multiple of 256 (WebGPU row-pitch requirement). */
function alignTo256(n: number): number {
  return Math.ceil(n / 256) * 256;
}
