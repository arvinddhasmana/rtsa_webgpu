// CLASSIFICATION: UNCLASSIFIED
// src/gpu/render-logic.ts — Pure render-worker logic extracted for unit testing
//
// Contains stateless helper functions and constants shared by render-worker.ts.
// These functions have no dependency on Worker globals (self, OffscreenCanvas, etc.)
// and can therefore be imported and tested in the jsdom Vitest environment.
//
// Reference: docs/implementation/v4/phase1_core_rendering.md

/** Target render interval matching 60 Hz. */
export const RENDER_INTERVAL_MS = 16;

/**
 * Number of consecutive renderFrame errors that triggers a permanent stop
 * of the render loop and a status:ready=false notification to the main thread.
 */
export const RENDER_ERROR_THRESHOLD = 5;

/** Payload shape for a status message sent back to the main thread. */
export interface RenderStatusMessage {
  type: "status";
  ready: boolean;
  error?: string;
}

/**
 * Build a status message indicating the render worker has encountered a fatal
 * error. Pure function — no side effects.
 */
export function makeErrorStatus(error: string): RenderStatusMessage {
  return { type: "status", ready: false, error };
}

/**
 * Compute the instantaneous frames-per-second value from a frame-to-frame
 * delta time in milliseconds.
 *
 * Returns 0 when dt ≤ 0 to avoid division by zero or negative results.
 */
export function computeFps(dtMs: number): number {
  return dtMs > 0 ? Math.round(1000 / dtMs) : 0;
}

/**
 * Decide whether the stats throttle counter has reached the flush threshold.
 * Returns true when the counter (1-indexed) is a multiple of 60, i.e. once
 * per second at a 60 Hz render interval.
 */
export function shouldFlushStats(counter: number): boolean {
  return counter >= 60;
}
