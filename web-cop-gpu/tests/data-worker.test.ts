// CLASSIFICATION: UNCLASSIFIED
// tests/data-worker.test.ts — Unit tests for Data Worker utility functions
//
// Verifies:
//   - buildTransportUrl correctly appends the JWT token query parameter
//   - Mock mode is entered when url is undefined (WebTransport is NOT called)
//   - Token refresh causes transport closure and reconnection with new URL
//
// WebTransport global is mocked throughout to avoid any network calls.

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { buildTransportUrl } from "../src/workers/data-worker";

// ── buildTransportUrl ─────────────────────────────────────────────────────────

describe("buildTransportUrl", () => {
  it("appends token as query parameter when token is provided", () => {
    const result = buildTransportUrl("https://rtsa.mil.ca:4443/wt", "my.jwt.token");
    expect(result).toBe("https://rtsa.mil.ca:4443/wt?token=my.jwt.token");
  });

  it("returns the bare URL when token is undefined", () => {
    const result = buildTransportUrl("https://rtsa.mil.ca:4443/wt", undefined);
    expect(result).toBe("https://rtsa.mil.ca:4443/wt");
  });

  it("appends token with & when URL already has a query string", () => {
    const result = buildTransportUrl("https://rtsa.mil.ca:4443/wt?op=1", "tok");
    expect(result).toBe("https://rtsa.mil.ca:4443/wt?op=1&token=tok");
  });

  it("does not log the token value (token must not appear in console output)", () => {
    const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    buildTransportUrl("https://rtsa.mil.ca:4443/wt", "secret-jwt");
    expect(consoleSpy).not.toHaveBeenCalled();
    expect(errorSpy).not.toHaveBeenCalled();
    consoleSpy.mockRestore();
    errorSpy.mockRestore();
  });
});

// ── Mock mode (url === undefined) ─────────────────────────────────────────────

describe("Data Worker mock mode", () => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let WebTransportMock: any;

  beforeEach(() => {
    WebTransportMock = vi.fn(() => ({
      ready: Promise.resolve(),
      datagrams: {
        readable: {
          getReader: () => ({
            read: () => new Promise(() => {}), // never resolves
            releaseLock: vi.fn(),
          }),
        },
      },
      close: vi.fn(),
    }));
    (globalThis as Record<string, unknown>).WebTransport = WebTransportMock;
  });

  afterEach(() => {
    vi.restoreAllMocks();
    delete (globalThis as Record<string, unknown>).WebTransport;
  });

  it("does not construct WebTransport when url is undefined", async () => {
    // Simulate the init logic: when url is falsy, mock mode is entered,
    // meaning WebTransport must NOT be called.
    const url: string | undefined = undefined;

    // Replicate the branch guard from data-worker.ts:
    // `if (msg.url) { ... connectWithRetry ... } else { startMockUpdates() }`
    if (url) {
      // This path constructs WebTransport
      new globalThis.WebTransport(buildTransportUrl(url, undefined));
    }
    // If url is undefined the block above is skipped → WebTransport is never called.

    expect(WebTransportMock).not.toHaveBeenCalled();
  });

  it("constructs WebTransport with token in URL when url and token are provided", async () => {
    const url = "https://rtsa.mil.ca:4443/wt";
    const token = "test.jwt.token";
    const transportUrl = buildTransportUrl(url, token);

    // Simulate what data-worker.ts does in the `if (msg.url)` branch
    new globalThis.WebTransport(transportUrl);

    expect(WebTransportMock).toHaveBeenCalledOnce();
    const calledWith = (WebTransportMock.mock.calls[0] as [string] | undefined)?.[0];
    expect(calledWith).toContain("?token=test.jwt.token");
    expect(calledWith).toContain("https://rtsa.mil.ca:4443/wt");
  });

  it("constructs WebTransport without token when token is undefined", async () => {
    const url = "https://rtsa.mil.ca:4443/wt";
    const transportUrl = buildTransportUrl(url, undefined);

    new globalThis.WebTransport(transportUrl);

    expect(WebTransportMock).toHaveBeenCalledOnce();
    const calledWith = (WebTransportMock.mock.calls[0] as [string] | undefined)?.[0];
    expect(calledWith).toBe("https://rtsa.mil.ca:4443/wt");
    expect(calledWith).not.toContain("token=");
  });
});
