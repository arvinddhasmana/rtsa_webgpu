// CLASSIFICATION: UNCLASSIFIED
// src/services/capabilities.ts — Browser capability detection

export interface Capabilities {
  webgpu: boolean;
  webtransport: boolean;
  sharedArrayBuffer: boolean;
  offscreenCanvas: boolean;
}

/**
 * Check all capabilities required by the WebGPU COP application.
 * All four capabilities are mandatory — if any is missing, the app
 * must fall back to the degraded notice.
 *
 * Reference: docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md §3.2
 */
export async function checkCapabilities(): Promise<Capabilities> {
  return {
    webgpu: typeof navigator !== "undefined" && "gpu" in navigator,
    webtransport: typeof WebTransport !== "undefined",
    sharedArrayBuffer: typeof SharedArrayBuffer !== "undefined",
    offscreenCanvas: typeof OffscreenCanvas !== "undefined",
  };
}
