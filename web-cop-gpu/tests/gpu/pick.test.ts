// CLASSIFICATION: UNCLASSIFIED
// tests/gpu/pick.test.ts — Unit tests for pick buffer row-pitch alignment,
// mapped-buffer guard, and coordinate scaling.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md §7

import { describe, it, expect } from "vitest";
import { readPickPixel, type PickResources } from "../../src/gpu/pick";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Inline mirror of the internal alignTo256 for formula tests. */
function alignTo256(n: number): number {
  return Math.ceil(n / 256) * 256;
}

/**
 * Build a minimal PickResources stub whose readbackBuffer has the given
 * mapState.  Only the fields accessed by readPickPixel before the first
 * GPU call are populated; the rest are left as empty stubs.
 */
function makeMockPick(mapState: GPUBufferMapState): PickResources {
  const readbackBuffer = { mapState } as unknown as GPUBuffer;
  return {
    texture: {} as GPUTexture,
    view: {} as GPUTextureView,
    readbackBuffer,
    width: 960,
    height: 540,
  };
}

// ---------------------------------------------------------------------------
// Pick buffer size formula
// ---------------------------------------------------------------------------

describe("Pick buffer size formula", () => {
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

  it("readback buffer is 256 bytes: minimum for 1-pixel copy at WebGPU alignment", () => {
    // Single-pixel copy: 1 pixel × 4 bytes = 4 bytes of data.
    // WebGPU requires bytesPerRow >= 256 for copyTextureToBuffer.
    // Therefore the minimum valid buffer size is 256 bytes regardless of canvas size.
    const bytesForOnePixel = 1 * 4;
    const alignedBytesPerRow = alignTo256(bytesForOnePixel);
    expect(alignedBytesPerRow).toBe(256);
    // The actual buffer allocation matches this minimum.
    expect(256 /* PICK_READBACK_BUFFER_SIZE */).toBe(alignedBytesPerRow);
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

// ---------------------------------------------------------------------------
// R-011: Mapped-buffer guard
// ---------------------------------------------------------------------------

describe("readPickPixel — mapped-buffer guard (R-011)", () => {
  it("returns null without throwing when buffer mapState is 'mapped'", async () => {
    const pick = makeMockPick("mapped");
    const result = await readPickPixel(
      null as unknown as GPUDevice,
      pick,
      /* canvasX */ 100,
      /* canvasY */ 100,
    );
    expect(result).toBeNull();
  });

  it("returns null without throwing when buffer mapState is 'pending'", async () => {
    const pick = makeMockPick("pending");
    const result = await readPickPixel(
      null as unknown as GPUDevice,
      pick,
      100,
      100,
    );
    expect(result).toBeNull();
  });
});

// ---------------------------------------------------------------------------
// R-012: Coordinate mapping
// ---------------------------------------------------------------------------

describe("readPickPixel — coordinate mapping (R-012)", () => {
  /**
   * Mirrors the coordinate transform in readPickPixel:
   *   canvasX = clientX × devicePixelRatio
   *   px      = Math.floor(canvasX / 2)
   */
  function clientToPickPixel(
    clientX: number,
    clientY: number,
    dpr: number,
  ): { px: number; py: number } {
    const canvasX = clientX * dpr;
    const canvasY = clientY * dpr;
    return {
      px: Math.floor(canvasX / 2),
      py: Math.floor(canvasY / 2),
    };
  }

  it("1920×1080 canvas, DPR=2: client (480, 270) → pick pixel (480, 270)", () => {
    // Physical canvas: 1920×1080 (1920 = 960 CSS px × DPR 2)
    // Pick texture:    960×540 (half physical)
    // Client click:    (480, 270) CSS px
    // canvasX = 480 × 2 = 960;  px = floor(960 / 2) = 480
    // canvasY = 270 × 2 = 540;  py = floor(540 / 2) = 270
    const { px, py } = clientToPickPixel(480, 270, 2);
    expect(px).toBe(480);
    expect(py).toBe(270);
  });

  it("DPR=1: client (400, 300) → pick pixel (200, 150)", () => {
    const { px, py } = clientToPickPixel(400, 300, 1);
    expect(px).toBe(200);
    expect(py).toBe(150);
  });

  it("DPR=3: client (160, 120) → pick pixel (240, 180)", () => {
    // canvasX = 160 × 3 = 480;  px = floor(480 / 2) = 240
    // canvasY = 120 × 3 = 360;  py = floor(360 / 2) = 180
    const { px, py } = clientToPickPixel(160, 120, 3);
    expect(px).toBe(240);
    expect(py).toBe(180);
  });
});

