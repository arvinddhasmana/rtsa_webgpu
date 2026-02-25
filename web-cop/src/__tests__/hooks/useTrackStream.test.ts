// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/hooks/useTrackStream.test.ts

import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useAuthStore } from "../../stores/authStore";
import { useTrackStore } from "../../stores/trackStore";

// Mock fetch for gRPC-Web stream
const mockAbort = vi.fn();
const mockAbortController = {
  abort: mockAbort,
  signal: { aborted: false },
};

describe("useTrackStream (T13)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useTrackStore.getState().clearAll();
    useAuthStore.getState().setOperator({
      id: "op-001",
      name: "Test",
      unit: "TEST",
      clearance: "PROTECTED_B",
      roles: [],
    });

    globalThis.AbortController = vi
      .fn()
      .mockImplementation(
        () => mockAbortController,
      ) as unknown as typeof globalThis.AbortController;
    globalThis.fetch = vi
      .fn()
      .mockRejectedValue(
        new Error("Connection refused"),
      ) as typeof globalThis.fetch;
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("T13: sets isConnected=false on connection error", async () => {
    // Dynamic import after mocking
    const { useTrackStream } = await import("../../hooks/useTrackStream");

    const { result } = renderHook(() => useTrackStream());

    // Wait for initial state
    await act(async () => {
      await new Promise((r) => setTimeout(r, 50));
    });

    expect(result.current.isConnected).toBe(false);
    expect(result.current.error).toBeTruthy();
  });

  it("T13: increments reconnectAttempts on repeated failures", async () => {
    const { useTrackStream } = await import("../../hooks/useTrackStream");

    const { result } = renderHook(() => useTrackStream());

    await act(async () => {
      await new Promise((r) => setTimeout(r, 50));
    });

    expect(result.current.reconnectAttempts).toBeGreaterThanOrEqual(1);
  });
});
