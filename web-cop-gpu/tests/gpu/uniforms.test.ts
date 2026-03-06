// CLASSIFICATION: UNCLASSIFIED
// tests/gpu/uniforms.test.ts — Unit tests for uniform buffer encoding
//
// Verifies that makeViewProjection produces a valid orthographic matrix
// and that the uniform struct byte layout matches the WGSL struct.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/wgsl_shader_standards.md §6

import { describe, it, expect, vi } from "vitest";
import { makeViewProjection, UNIFORM_BYTES, writeUniforms } from "../../src/gpu/uniforms";

describe("makeViewProjection", () => {
  it("produces a 16-element Float32Array", () => {
    const vp = makeViewProjection(1920, 1080, 0, 0, 2);
    expect(vp).toBeInstanceOf(Float32Array);
    expect(vp.length).toBe(16);
  });

  it("identity case: center (0,0), scale 1, square canvas", () => {
    const vp = makeViewProjection(100, 100, 0, 0, 1);
    // Column-major orthographic: scale on diagonal, no translation
    expect(vp[0]).toBeCloseTo(1, 5);  // sx
    expect(vp[5]).toBeCloseTo(1, 5);  // sy
    expect(vp[12]).toBeCloseTo(0, 5); // tx = -centerLon * sx = 0
    expect(vp[13]).toBeCloseTo(0, 5); // ty = -centerLat * sy = 0
    expect(vp[15]).toBe(1);           // homogeneous row
  });

  it("scale doubles with scale=2", () => {
    const vp1 = makeViewProjection(100, 100, 0, 0, 1);
    const vp2 = makeViewProjection(100, 100, 0, 0, 2);
    expect(vp2[0]).toBeCloseTo(vp1[0]! * 2, 5);
    expect(vp2[5]).toBeCloseTo(vp1[5]! * 2, 5);
  });

  it("aspect ratio: wider canvas reduces sx relative to sy", () => {
    const vp = makeViewProjection(1920, 1080, 0, 0, 1);
    // sx = scale / aspect = 1 / (1920/1080) ≈ 0.5625
    // sy = scale = 1
    expect(vp[0]!).toBeLessThan(vp[5]!);
  });

  it("non-zero centre shifts translation column", () => {
    const lon = 0.5;
    const lat = 0.3;
    const vp  = makeViewProjection(100, 100, lon, lat, 1);
    // tx = -centerLon * sx  (sx = sy = 1 for square canvas)
    expect(vp[12]).toBeCloseTo(-lon, 5);
    expect(vp[13]).toBeCloseTo(-lat, 5);
  });

  it("handles zero canvas height gracefully (no NaN)", () => {
    const vp = makeViewProjection(100, 0, 0, 0, 1);
    for (const v of vp) {
      expect(isNaN(v)).toBe(false);
    }
  });
});

describe("UNIFORM_BYTES layout", () => {
  it("is exactly 80 bytes", () => {
    expect(UNIFORM_BYTES).toBe(80);
  });

  it("is 16-byte aligned (required by WGSL)", () => {
    expect(UNIFORM_BYTES % 16).toBe(0);
  });
});

describe("writeUniforms", () => {
  it("calls device.queue.writeBuffer with 80-byte data", () => {
    const writeBuffer = vi.fn();
    const mockDevice  = { queue: { writeBuffer } } as unknown as GPUDevice;
    const mockBuffer  = {} as GPUBuffer;
    const vp          = makeViewProjection(1920, 1080, 0, 0, 2);

    writeUniforms(mockDevice, mockBuffer, vp, 1920, 1080, 1000, 500);

    expect(writeBuffer).toHaveBeenCalledOnce();
    const [buf, offset, data] = writeBuffer.mock.calls[0] as [GPUBuffer, number, ArrayBuffer];
    expect(buf).toBe(mockBuffer);
    expect(offset).toBe(0);
    expect(data.byteLength).toBe(UNIFORM_BYTES);
  });

  it("encodes trackCount at byte offset 76 (u32 index 19)", () => {
    const writeBuffer = vi.fn();
    const mockDevice  = { queue: { writeBuffer } } as unknown as GPUDevice;
    const mockBuffer  = {} as GPUBuffer;
    const vp          = makeViewProjection(1920, 1080, 0, 0, 2);

    writeUniforms(mockDevice, mockBuffer, vp, 1920, 1080, 0, 12345);

    const [, , data] = writeBuffer.mock.calls[0] as [GPUBuffer, number, ArrayBuffer];
    const u32View    = new Uint32Array(data);
    expect(u32View[19]).toBe(12345); // track_count at byte 76
  });

  it("encodes viewportSize at byte offsets 64 (float16) and 68 (float17)", () => {
    const writeBuffer = vi.fn();
    const mockDevice  = { queue: { writeBuffer } } as unknown as GPUDevice;
    const mockBuffer  = {} as GPUBuffer;
    const vp          = makeViewProjection(800, 600, 0, 0, 1);

    writeUniforms(mockDevice, mockBuffer, vp, 800, 600, 0, 0);

    const [, , data] = writeBuffer.mock.calls[0] as [GPUBuffer, number, ArrayBuffer];
    const f32View    = new Float32Array(data);
    expect(f32View[16]).toBe(800); // viewport_size.x
    expect(f32View[17]).toBe(600); // viewport_size.y
  });
});
