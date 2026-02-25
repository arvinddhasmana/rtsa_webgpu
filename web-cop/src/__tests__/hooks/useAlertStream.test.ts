// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/hooks/useAlertStream.test.ts

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useAlertStore } from "../../stores/alertStore";
import { useAuthStore } from "../../stores/authStore";

describe("useAlertStream (T14)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAlertStore.getState().clearAll();
    useAlertStore.getState().setMinSeverityFilter("WATCH");
    useAuthStore.getState().setOperator({
      id: "op-001",
      name: "Test",
      unit: "TEST",
      clearance: "PROTECTED_B",
      roles: [],
    });

    global.AbortController = vi.fn().mockImplementation(() => ({
      abort: vi.fn(),
      signal: { aborted: false },
    }));
    global.fetch = vi.fn().mockRejectedValue(new Error("Connection refused"));
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("T14: starts disconnected when backend unavailable", async () => {
    const { useAlertStream } = await import("../../hooks/useAlertStream");

    const { result } = renderHook(() => useAlertStream());

    await act(async () => {
      await new Promise((r) => setTimeout(r, 50));
    });

    expect(result.current.isConnected).toBe(false);
    expect(result.current.error).toBeTruthy();
  });

  it("T14: starts with reconnectAttempts=0", async () => {
    const { useAlertStream } = await import("../../hooks/useAlertStream");
    const { result } = renderHook(() => useAlertStream());

    // Initial state before any async
    expect(result.current.reconnectAttempts).toBe(0);
  });

  it("alert store integrates correctly with severity filter", () => {
    useAlertStore.getState().setMinSeverityFilter("CRITICAL");
    expect(useAlertStore.getState().minSeverityFilter).toBe("CRITICAL");
  });
});
