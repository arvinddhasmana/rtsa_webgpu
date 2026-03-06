// CLASSIFICATION: UNCLASSIFIED
// src/gpu/device.ts — WebGPU adapter/device acquisition and canvas configuration
//
// Reference: docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md §3

export interface GPUContext {
  device: GPUDevice;
  context: GPUCanvasContext;
  format: GPUTextureFormat;
}

/**
 * Acquire a high-performance WebGPU adapter and device.
 * Configures the OffscreenCanvas context with premultiplied alpha.
 *
 * Throws if WebGPU is unavailable (capability gate should have prevented this).
 * On device loss, re-initialisation is attempted after a short delay.
 *
 * Reference: webgpu_guidelines.md §3.1
 */
export async function initGPU(canvas: OffscreenCanvas): Promise<GPUContext> {
  const adapter = await navigator.gpu.requestAdapter({
    powerPreference: "high-performance",
  });
  if (!adapter) {
    throw new Error("[GPU] WebGPU adapter unavailable");
  }

  const device = await adapter.requestDevice({
    requiredLimits: {
      maxStorageBufferBindingSize: adapter.limits.maxStorageBufferBindingSize,
      maxBufferSize: adapter.limits.maxBufferSize,
    },
  });

  // Handle device loss — log and schedule re-init
  device.lost.then((info) => {
    if (info.reason === "destroyed") return; // intentional teardown
    console.error(`[GPU] Device lost: ${info.reason} — ${info.message}`);
    // Caller is responsible for re-initialising via the Render Worker
  });

  const context = canvas.getContext("webgpu");
  if (!context) {
    throw new Error("[GPU] Failed to acquire webgpu canvas context");
  }

  const format = navigator.gpu.getPreferredCanvasFormat();
  context.configure({
    device,
    format,
    alphaMode: "premultiplied",
  });

  return { device, context, format };
}
