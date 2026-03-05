// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/hooks/useFeedback.test.ts

import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { feedbackClient } from "../../api/feedback-client";
import { useFeedback } from "../../hooks/useFeedback";
import { useAuthStore } from "../../stores/authStore";

vi.mock("../../api/feedback-client", () => ({
  feedbackClient: {
    submitFeedback: vi.fn(),
  },
}));

describe("useFeedback", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.getState().setOperator({
      id: "op-001",
      name: "Test Operator",
      unit: "TEST",
      clearance: "PROTECTED_B",
      roles: ["OPERATOR"],
    });
  });

  it("initializes with idle state", () => {
    const { result } = renderHook(() => useFeedback());
    expect(result.current.status).toBe("idle");
    expect(result.current.trustScore).toBeNull();
    expect(result.current.error).toBeNull();
  });

  it("transitions to submitting then accepted on score >= 0.8", async () => {
    (feedbackClient.submitFeedback as any).mockResolvedValue({
      trustScore: 0.85,
      accepted: true,
    });

    const { result } = renderHook(() => useFeedback());

    let promise: Promise<void>;

    act(() => {
      promise = result.current.submit("TRK-1", "ALT-1", "CONFIRM_HOSTILE", "Test notes");
    });

    expect(result.current.status).toBe("submitting");

    await act(async () => {
      await promise;
    });

    expect(result.current.status).toBe("accepted");
    expect(result.current.trustScore).toBe(0.85);
  });

  it("transitions to under_review on score >= 0.4 and < 0.8", async () => {
    (feedbackClient.submitFeedback as any).mockResolvedValue({
      trustScore: 0.6,
      accepted: true,
    });

    const { result } = renderHook(() => useFeedback());

    await act(async () => {
      await result.current.submit("TRK-1", "ALT-1", "RECLASSIFY", "Test review");
    });

    expect(result.current.status).toBe("under_review");
    expect(result.current.trustScore).toBe(0.6);
  });

  it("transitions to rejected state on API error", async () => {
    (feedbackClient.submitFeedback as any).mockRejectedValue(new Error("Engine rejected feedback"));

    const { result } = renderHook(() => useFeedback());

    await act(async () => {
      await result.current.submit("TRK-1", "ALT-1", "CONFIRM_HOSTILE", "Test notes");
    });

    expect(result.current.status).toBe("rejected");
    expect(result.current.error).toBeTruthy();
    expect(result.current.error?.message).toBe("Engine rejected feedback");
  });
});
