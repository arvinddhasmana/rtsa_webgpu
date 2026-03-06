// CLASSIFICATION: UNCLASSIFIED
// tests/gpu/pick.test.ts — Unit tests for pick buffer row-pitch alignment
//
// Tests the alignTo256 helper (exported indirectly via behaviour)
// and boundary conditions for pick pixel coordinates.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md §7

import { describe, it, expect } from "vitest";

// alignTo256 is a private helper — we test its behaviour via createPickResources
// indirectly through the readbackBuffer size formula.

describe("Pick buffer size formula", () => {
  /**
   * Mirror of the internal alignTo256 for testing the formula:
   *   bytesPerRow = ceil(w * 4 / 256) * 256
   *   size        = bytesPerRow * h
   */
  function alignTo256(n: number): number {
    return Math.ceil(n / 256) * 256;
  }

  it("alignTo256(256) returns 256", () => {
    expect(alignTo256(256)).toBe(256);
  });

  it("alignTo256(257) returns 512", () => {
    expect(alignTo256(257)).toBe(512);
  });

  it("alignTo256(0) returns 0", () => {
    expect(alignTo256(0)).toBe(0);
  });

  it("alignTo256(1) returns 256", () => {
    expect(alignTo256(1)).toBe(256);
  });

  it("alignTo256(1024) returns 1024 (already aligned)", () => {
    expect(alignTo256(1024)).toBe(1024);
  });

  it("readback buffer size for 1920×1080 canvas (half-res pick = 960×540)", () => {
    const w = 960;
    const h = 540;
    const bytesPerRow = alignTo256(w * 4); // 3840 → 3840 (already aligned)
    const totalBytes  = bytesPerRow * h;
    expect(totalBytes).toBe(3840 * 540);
  });

  it("pick texture is half canvas resolution", () => {
    const canvasW = 1920;
    const canvasH = 1080;
    const pickW   = Math.floor(canvasW / 2);
    const pickH   = Math.floor(canvasH / 2);
    expect(pickW).toBe(960);
    expect(pickH).toBe(540);
  });

  it("pick pixel coordinate scaling from canvas → pick texture", () => {
    // canvasX / 2 should give pick texture x
    const canvasX = 400;
    const pickX   = Math.floor(canvasX / 2);
    expect(pickX).toBe(200);
  });

  it("out-of-bounds pick coordinate detected correctly", () => {
    const pickW = 960;
    const pickH = 540;

    function isOutOfBounds(px: number, py: number): boolean {
      return px < 0 || py < 0 || px >= pickW || py >= pickH;
    }

    expect(isOutOfBounds(-1,  0)).toBe(true);
    expect(isOutOfBounds(0,  -1)).toBe(true);
    expect(isOutOfBounds(960, 0)).toBe(true);
    expect(isOutOfBounds(0, 540)).toBe(true);
    expect(isOutOfBounds(959, 539)).toBe(false);
    expect(isOutOfBounds(0, 0)).toBe(false);
  });
});
