// CLASSIFICATION: UNCLASSIFIED
// tests/capabilities.test.ts — Unit tests for the browser capability gate
//
// Verifies that checkCapabilities() returns correct booleans for each
// mocked navigator/globalThis state.

import { describe, it, expect, vi, afterEach } from "vitest";

// We import dynamically inside each test so vi.stubGlobal takes effect.
afterEach(() => {
  vi.unstubAllGlobals();
});

describe("checkCapabilities", () => {
  it("returns all true when all APIs are present", async () => {
    vi.stubGlobal("navigator", {
      gpu: {},
    });
    vi.stubGlobal("WebTransport", class {});
    vi.stubGlobal("SharedArrayBuffer", class {});
    vi.stubGlobal("OffscreenCanvas", class {});

    const { checkCapabilities } = await import("../src/services/capabilities");
    const caps = await checkCapabilities();

    expect(caps.webgpu).toBe(true);
    expect(caps.webtransport).toBe(true);
    expect(caps.sharedArrayBuffer).toBe(true);
    expect(caps.offscreenCanvas).toBe(true);
  });

  it("returns webgpu=false when navigator.gpu is absent", async () => {
    vi.stubGlobal("navigator", {});
    vi.stubGlobal("WebTransport", class {});
    vi.stubGlobal("SharedArrayBuffer", class {});
    vi.stubGlobal("OffscreenCanvas", class {});

    const { checkCapabilities } = await import("../src/services/capabilities");
    const caps = await checkCapabilities();

    expect(caps.webgpu).toBe(false);
    expect(caps.webtransport).toBe(true);
    expect(caps.sharedArrayBuffer).toBe(true);
    expect(caps.offscreenCanvas).toBe(true);
  });

  it("returns webtransport=false when WebTransport is absent", async () => {
    vi.stubGlobal("navigator", { gpu: {} });
    // Do not stub WebTransport — leave it undefined
    vi.stubGlobal("WebTransport", undefined);
    vi.stubGlobal("SharedArrayBuffer", class {});
    vi.stubGlobal("OffscreenCanvas", class {});

    const { checkCapabilities } = await import("../src/services/capabilities");
    const caps = await checkCapabilities();

    expect(caps.webgpu).toBe(true);
    expect(caps.webtransport).toBe(false);
  });

  it("returns sharedArrayBuffer=false when SharedArrayBuffer is absent", async () => {
    vi.stubGlobal("navigator", { gpu: {} });
    vi.stubGlobal("WebTransport", class {});
    vi.stubGlobal("SharedArrayBuffer", undefined);
    vi.stubGlobal("OffscreenCanvas", class {});

    const { checkCapabilities } = await import("../src/services/capabilities");
    const caps = await checkCapabilities();

    expect(caps.sharedArrayBuffer).toBe(false);
  });

  it("returns offscreenCanvas=false when OffscreenCanvas is absent", async () => {
    vi.stubGlobal("navigator", { gpu: {} });
    vi.stubGlobal("WebTransport", class {});
    vi.stubGlobal("SharedArrayBuffer", class {});
    vi.stubGlobal("OffscreenCanvas", undefined);

    const { checkCapabilities } = await import("../src/services/capabilities");
    const caps = await checkCapabilities();

    expect(caps.offscreenCanvas).toBe(false);
  });

  it("returns all false when no APIs are present", async () => {
    vi.stubGlobal("navigator", {});
    vi.stubGlobal("WebTransport", undefined);
    vi.stubGlobal("SharedArrayBuffer", undefined);
    vi.stubGlobal("OffscreenCanvas", undefined);

    const { checkCapabilities } = await import("../src/services/capabilities");
    const caps = await checkCapabilities();

    expect(caps.webgpu).toBe(false);
    expect(caps.webtransport).toBe(false);
    expect(caps.sharedArrayBuffer).toBe(false);
    expect(caps.offscreenCanvas).toBe(false);
  });
});
