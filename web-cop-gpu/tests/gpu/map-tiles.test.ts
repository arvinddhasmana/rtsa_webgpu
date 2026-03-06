// CLASSIFICATION: UNCLASSIFIED
// tests/gpu/map-tiles.test.ts — Unit tests for map tile background layer
//
// Verifies the render pass descriptor structure returned by
// makeBackgroundPassDescriptor.

import { describe, it, expect } from "vitest";
import {
  makeBackgroundPassDescriptor,
  MAP_BACKGROUND_COLOR,
} from "../../src/gpu/map-tiles";

describe("makeBackgroundPassDescriptor", () => {
  const mockView = {} as GPUTextureView;

  it("returns a descriptor with one color attachment", () => {
    const desc = makeBackgroundPassDescriptor(mockView);
    expect(desc.colorAttachments).toHaveLength(1);
  });

  it("uses loadOp: clear on the first attachment", () => {
    const desc = makeBackgroundPassDescriptor(mockView);
    const att  = (desc.colorAttachments as GPURenderPassColorAttachment[])[0]!;
    expect(att.loadOp).toBe("clear");
  });

  it("uses storeOp: store on the first attachment", () => {
    const desc = makeBackgroundPassDescriptor(mockView);
    const att  = (desc.colorAttachments as GPURenderPassColorAttachment[])[0]!;
    expect(att.storeOp).toBe("store");
  });

  it("sets clearValue to the map background colour", () => {
    const desc = makeBackgroundPassDescriptor(mockView);
    const att  = (desc.colorAttachments as GPURenderPassColorAttachment[])[0]!;
    expect(att.clearValue).toEqual(MAP_BACKGROUND_COLOR);
  });

  it("binds the provided texture view", () => {
    const desc = makeBackgroundPassDescriptor(mockView);
    const att  = (desc.colorAttachments as GPURenderPassColorAttachment[])[0]!;
    expect(att.view).toBe(mockView);
  });

  it("background colour channels are in valid 0–1 range", () => {
    const col = MAP_BACKGROUND_COLOR as GPUColorDict;
    expect(col.r).toBeGreaterThanOrEqual(0);
    expect(col.r).toBeLessThanOrEqual(1);
    expect(col.g).toBeGreaterThanOrEqual(0);
    expect(col.g).toBeLessThanOrEqual(1);
    expect(col.b).toBeGreaterThanOrEqual(0);
    expect(col.b).toBeLessThanOrEqual(1);
    expect(col.a).toBe(1.0);
  });
});
