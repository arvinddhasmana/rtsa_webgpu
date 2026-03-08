// CLASSIFICATION: UNCLASSIFIED
// tests/gpu/lod.test.ts — Unit tests for the LOD system

import { describe, it, expect } from "vitest";
import { computeLod, LOD_LEVELS } from "../../src/gpu/lod";

describe("computeLod", () => {
  it("returns level=full at high scale (zoomed in)", () => {
    const flags = computeLod(1.0, 50_000);
    expect(flags.level).toBe("full");
    expect(flags.renderTrails).toBe(true);
    expect(flags.renderHalos).toBe(true);
    expect(flags.renderLabels).toBe(true);
    expect(flags.maxInstances).toBe(50_000);
  });

  it("returns level=full at exactly the FULL threshold", () => {
    const flags = computeLod(LOD_LEVELS.FULL, 1_000);
    expect(flags.level).toBe("full");
  });

  it("returns level=medium between FULL and MEDIUM thresholds", () => {
    const flags = computeLod(0.2, 5_000);
    expect(flags.level).toBe("medium");
    expect(flags.renderTrails).toBe(false);
    expect(flags.renderHalos).toBe(true);
    expect(flags.renderLabels).toBe(true); // track count ≤ 10k
    expect(flags.maxInstances).toBe(5_000);
  });

  it("caps medium maxInstances at 20k when track count > 20k", () => {
    const flags = computeLod(0.2, 50_000);
    expect(flags.level).toBe("medium");
    expect(flags.maxInstances).toBe(20_000);
  });

  it("disables labels in medium when trackCount > 10k", () => {
    const flags = computeLod(0.2, 15_000);
    expect(flags.level).toBe("medium");
    expect(flags.renderLabels).toBe(false);
  });

  it("returns level=minimal at very low scale", () => {
    const flags = computeLod(0.001, 50_000);
    expect(flags.level).toBe("minimal");
    expect(flags.renderTrails).toBe(false);
    expect(flags.renderHalos).toBe(false);
    expect(flags.renderLabels).toBe(false);
    expect(flags.maxInstances).toBe(10_000);
  });

  it("caps minimal maxInstances at 10k", () => {
    const flags = computeLod(0.005, 50_000);
    expect(flags.maxInstances).toBe(10_000);
  });

  it("handles zero tracks at all LOD levels", () => {
    expect(computeLod(1.0, 0).maxInstances).toBe(0);
    expect(computeLod(0.2, 0).maxInstances).toBe(0);
    expect(computeLod(0.001, 0).maxInstances).toBe(0);
  });

  it("LOD_LEVELS constants are ordered (FULL > MEDIUM > MINIMAL)", () => {
    expect(LOD_LEVELS.FULL).toBeGreaterThan(LOD_LEVELS.MEDIUM);
    expect(LOD_LEVELS.MEDIUM).toBeGreaterThan(LOD_LEVELS.MINIMAL);
  });

  it("maxInstances is always ≤ trackCount across all LOD levels", () => {
    const trackCounts = [0, 1, 5_000, 10_000, 20_000, 50_000];
    const scales = [0.001, 0.01, 0.1, 0.5, 1.0, 2.0];
    for (const tc of trackCounts) {
      for (const s of scales) {
        const flags = computeLod(s, tc);
        expect(flags.maxInstances).toBeLessThanOrEqual(tc);
      }
    }
  });
});
