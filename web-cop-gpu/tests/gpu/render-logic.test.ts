// CLASSIFICATION: UNCLASSIFIED
// tests/gpu/render-logic.test.ts — Unit tests for pure render-worker logic
//
// Verifies:
//   - makeErrorStatus returns correctly typed status message
//   - computeFps handles normal, zero, and negative delta times
//   - shouldFlushStats triggers at correct counter values
//   - Constants have expected values

import { describe, it, expect } from "vitest";
import {
  makeErrorStatus,
  computeFps,
  shouldFlushStats,
  RENDER_INTERVAL_MS,
  RENDER_ERROR_THRESHOLD,
} from "../../src/gpu/render-logic";

describe("makeErrorStatus", () => {
  it("returns type=status with ready=false", () => {
    const msg = makeErrorStatus("GPU pipeline failure");
    expect(msg.type).toBe("status");
    expect(msg.ready).toBe(false);
  });

  it("includes the error string", () => {
    const msg = makeErrorStatus("Device lost");
    expect(msg.error).toBe("Device lost");
  });

  it("handles empty error string", () => {
    const msg = makeErrorStatus("");
    expect(msg.error).toBe("");
    expect(msg.ready).toBe(false);
  });

  it("does not mutate input or shared state", () => {
    const msg1 = makeErrorStatus("error A");
    const msg2 = makeErrorStatus("error B");
    expect(msg1.error).toBe("error A");
    expect(msg2.error).toBe("error B");
  });
});

describe("computeFps", () => {
  it("computes fps correctly for 16ms frame time (≈60 fps)", () => {
    expect(computeFps(16)).toBe(63); // Math.round(1000/16)
  });

  it("computes fps correctly for 33ms frame time (≈30 fps)", () => {
    expect(computeFps(33)).toBe(30); // Math.round(1000/33)
  });

  it("returns 0 when dt is 0", () => {
    expect(computeFps(0)).toBe(0);
  });

  it("returns 0 when dt is negative", () => {
    expect(computeFps(-1)).toBe(0);
  });

  it("returns a non-negative integer", () => {
    for (const dt of [8, 16, 32, 100, 500]) {
      const fps = computeFps(dt);
      expect(fps).toBeGreaterThanOrEqual(0);
      expect(Number.isInteger(fps)).toBe(true);
    }
  });
});

describe("shouldFlushStats", () => {
  it("returns false for counter < 60", () => {
    for (let i = 1; i < 60; i++) {
      expect(shouldFlushStats(i)).toBe(false);
    }
  });

  it("returns true when counter reaches 60", () => {
    expect(shouldFlushStats(60)).toBe(true);
  });

  it("returns true when counter exceeds 60", () => {
    expect(shouldFlushStats(61)).toBe(true);
    expect(shouldFlushStats(120)).toBe(true);
  });

  it("returns false for counter of 0", () => {
    expect(shouldFlushStats(0)).toBe(false);
  });
});

describe("constants", () => {
  it("RENDER_INTERVAL_MS is 16 (~60 FPS)", () => {
    expect(RENDER_INTERVAL_MS).toBe(16);
  });

  it("RENDER_ERROR_THRESHOLD is 5", () => {
    expect(RENDER_ERROR_THRESHOLD).toBe(5);
  });
});
