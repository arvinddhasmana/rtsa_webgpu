// CLASSIFICATION: UNCLASSIFIED
// tests/gpu/frame-timer.test.ts — Unit tests for FrameTimer

import { describe, it, expect, vi, beforeAll } from "vitest";
import { FrameTimer, PASS_LABELS } from "../../src/gpu/frame-timer";

// ── Stub WebGPU globals required by FrameTimer ──────────────────────────────
// GPUBufferUsage and GPUMapMode are not defined in jsdom. Provide minimal stubs
// that match the numeric values used in the WebGPU spec.
beforeAll(() => {
  if (typeof globalThis.GPUBufferUsage === "undefined") {
    Object.defineProperty(globalThis, "GPUBufferUsage", {
      value: {
        QUERY_RESOLVE: 0x0200,
        COPY_SRC:      0x0004,
        COPY_DST:      0x0008,
        MAP_READ:      0x0001,
        STORAGE:       0x0080,
        UNIFORM:       0x0040,
        VERTEX:        0x0020,
        INDEX:         0x0010,
        INDIRECT:      0x0100,
      },
      configurable: true,
    });
  }
  if (typeof globalThis.GPUMapMode === "undefined") {
    Object.defineProperty(globalThis, "GPUMapMode", {
      value: { READ: 1, WRITE: 2 },
      configurable: true,
    });
  }
});

// ── Minimal GPUDevice mock ─────────────────────────────────────────────────
function makeDevice(timestampSupported = false): GPUDevice {
  const features = new Set<string>();
  if (timestampSupported) {
    features.add("timestamp-query");
    features.add("timestamp-query-inside-passes");
  }
  return {
    features: {
      has: (f: string) => features.has(f),
    },
    createQuerySet: vi.fn(() => ({ destroy: vi.fn() })),
    createBuffer: vi.fn(() => ({
      destroy: vi.fn(),
      mapAsync: vi.fn().mockResolvedValue(undefined),
      getMappedRange: vi.fn(() => new BigUint64Array(PASS_LABELS.length).buffer),
      unmap: vi.fn(),
    })),
  } as unknown as GPUDevice;
}

describe("FrameTimer", () => {
  it("reports isSupported=false when device lacks timestamp-query", () => {
    const device = makeDevice(false);
    const timer = new FrameTimer(device);
    expect(timer.isSupported).toBe(false);
    expect(timer.gpuQuerySet).toBeNull();
  });

  it("reports isSupported=true when device has timestamp-query features", () => {
    const device = makeDevice(true);
    const timer = new FrameTimer(device);
    expect(timer.isSupported).toBe(true);
    expect(timer.gpuQuerySet).not.toBeNull();
  });

  it("markJsStart / markJsEnd return elapsed ms ≥ 0", () => {
    const device = makeDevice(false);
    const timer = new FrameTimer(device);
    timer.markJsStart();
    const elapsed = timer.markJsEnd();
    expect(elapsed).toBeGreaterThanOrEqual(0);
  });

  it("smoothed timings start at zero", () => {
    const device = makeDevice(false);
    const timer = new FrameTimer(device);
    expect(timer.smoothed.totalFrameMs).toBe(0);
    expect(timer.smoothed.mainThreadMs).toBe(0);
  });

  it("readbackAsync updates smoothed.mainThreadMs when GPU unsupported", async () => {
    const device = makeDevice(false);
    const timer = new FrameTimer(device);
    await timer.readbackAsync(5.0);
    expect(timer.smoothed.mainThreadMs).toBeCloseTo(5.0);
    expect(timer.smoothed.totalFrameMs).toBeCloseTo(5.0);
  });

  it("readbackAsync accumulates rolling average", async () => {
    const device = makeDevice(false);
    const timer = new FrameTimer(device);
    await timer.readbackAsync(4.0);
    await timer.readbackAsync(8.0);
    // Average of 4 and 8 = 6
    expect(timer.smoothed.mainThreadMs).toBeCloseTo(6.0);
  });

  it("resolveTimestamps is a no-op when unsupported", () => {
    const device = makeDevice(false);
    const timer = new FrameTimer(device);
    const encoder = { resolveQuerySet: vi.fn(), copyBufferToBuffer: vi.fn() } as unknown as GPUCommandEncoder;
    // Should not throw
    expect(() => timer.resolveTimestamps(encoder)).not.toThrow();
    expect(encoder.resolveQuerySet).not.toHaveBeenCalled();
  });

  it("destroy does not throw when unsupported", () => {
    const device = makeDevice(false);
    const timer = new FrameTimer(device);
    expect(() => timer.destroy()).not.toThrow();
  });

  it("destroy does not throw when supported", () => {
    const device = makeDevice(true);
    const timer = new FrameTimer(device);
    expect(() => timer.destroy()).not.toThrow();
  });

  it("PASS_LABELS has correct length", () => {
    expect(PASS_LABELS.length).toBe(11);
  });

  it("PASS_LABELS starts with frame_start and ends with frame_end", () => {
    expect(PASS_LABELS[0]).toBe("frame_start");
    expect(PASS_LABELS[PASS_LABELS.length - 1]).toBe("frame_end");
  });

  it("readbackAsync falls back to CPU-only timings when mapAsync rejects", async () => {
    // Build a device mock whose createBuffer returns a buffer that rejects mapAsync
    const features = new Set(["timestamp-query", "timestamp-query-inside-passes"]);
    const failDevice = {
      features: { has: (f: string) => features.has(f) },
      createQuerySet: vi.fn(() => ({ destroy: vi.fn() })),
      createBuffer: vi.fn(() => ({
        destroy: vi.fn(),
        mapAsync: vi.fn().mockRejectedValue(new DOMException("Device lost", "AbortError")),
        getMappedRange: vi.fn(),
        unmap: vi.fn(),
      })),
    } as unknown as GPUDevice;

    const timer = new FrameTimer(failDevice);
    // Should not throw despite mapAsync rejecting; falls back to CPU timings
    await expect(timer.readbackAsync(7.5)).resolves.toBeUndefined();
    expect(timer.smoothed.mainThreadMs).toBeCloseTo(7.5);
    // getMappedRange must NOT be called after a rejected mapAsync
    const buf = (failDevice.createBuffer as ReturnType<typeof vi.fn>).mock.results[0]?.value;
    if (buf) {
      expect(buf.getMappedRange).not.toHaveBeenCalled();
    }
  });
});
