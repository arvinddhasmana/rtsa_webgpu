// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/hooks/useEventTimeline.test.ts

import { renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { queryClient } from "../../api/query-client";
import { useEventTimeline } from "../../hooks/useEventTimeline";

// Mock the grpc Client
vi.mock("../../api/query-client", () => ({
  queryClient: {
    getEventTimeline: vi.fn(),
  },
}));

describe("useEventTimeline", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("returns empty initial state for no trackId", () => {
    const { result } = renderHook(() => useEventTimeline(null, "ALL"));

    expect(result.current.events).toEqual([]);
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBe(null);
    expect(result.current.refreshing).toBe(false);
  });

  it("fetches events when trackId is provided", async () => {
    const mockEvents = [
      { eventId: "E1", trackId: "TRK-1", type: "track", timestamp: new Date(), sourceId: "SYS" }
    ];

    (queryClient.getEventTimeline as any).mockResolvedValue({ events: mockEvents });

    const { result } = renderHook(() => useEventTimeline("TRK-1", "ALL"));

    // Initial loading state
    expect(result.current.loading).toBe(true);

    // Wait for fetch to complete
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.events).toHaveLength(1);
    expect(result.current.error).toBe(null);
  });

  it("clears events when trackId becomes null", async () => {
    const mockEvents = [
      { eventId: "E1", trackId: "TRK-1", type: "track", timestamp: new Date(), sourceId: "SYS" }
    ];

    (queryClient.getEventTimeline as any).mockResolvedValue({ events: mockEvents });

    const { result, rerender } = renderHook(
      ({ id, type }) => useEventTimeline(id, type),
      { initialProps: { id: "TRK-1", type: "ALL" as const } }
    );

    await waitFor(() => {
      expect(result.current.events).toHaveLength(1);
    });

    // Rerender with null trackId
    rerender({ id: null, type: "ALL" });

    expect(result.current.events).toEqual([]);
    expect(result.current.loading).toBe(false);
  });

  it("handles API errors gracefully", async () => {
    const errorMsg = "Network error";
    (queryClient.getEventTimeline as any).mockRejectedValue(new Error(errorMsg));

    const { result } = renderHook(() => useEventTimeline("TRK-1", "ALL"));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toBeTruthy();
    expect(result.current.error).toBe("Timeline unavailable");
  });
});
