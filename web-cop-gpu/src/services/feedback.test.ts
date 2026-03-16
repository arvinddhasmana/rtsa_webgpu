// CLASSIFICATION: UNCLASSIFIED
// src/services/feedback.test.ts — Unit tests for FeedbackService gRPC wrapper
//
// Tests:
//   - submitFeedback validates required inputs before calling gRPC
//   - submitFeedback resolves with correctly mapped result on success
//   - submitFeedback propagates gRPC errors without swallowing them
//   - justification length validation (max 500 chars)

import { beforeEach, describe, expect, it, vi } from "vitest";

// ── Mock ConnectRPC client at the transport boundary ──────────────────────────

const mockSubmitFeedback = vi.hoisted(() => vi.fn());

vi.mock("@connectrpc/connect", () => ({
  createPromiseClient: () => ({
    submitFeedback: mockSubmitFeedback,
  }),
}));

vi.mock("./grpc-client", () => ({ transport: {} }));

vi.mock("@gen/rtsa/feedback/v1/feedback_service_connect.js", () => ({
  FeedbackService: {},
}));

vi.mock("@gen/rtsa/common/v1/types_pb.js", () => ({
  ClassificationLevel: { UNCLASSIFIED: 1 },
  FeedbackType: {
    CONFIRM_HOSTILE: 1,
    CONFIRM_FRIENDLY: 2,
    RECLASSIFY: 3,
    REJECT_ANOMALY: 4,
    CONFIRM_ANOMALY: 5,
  },
  EntityType: {},
  HostileClassification: {},
  TrackStatus: {},
}));

import {
  buildConfirmAlertFeedbackRequest,
  buildRejectAlertFeedbackRequest,
  submitConfirmAlertFeedback,
  submitFeedback,
  submitRejectAlertFeedback,
  type SubmitFeedbackParams,
} from "./feedback";

const baseParams: SubmitFeedbackParams = {
  trackId: "track-001",
  operatorId: "op-007",
  feedbackType: "CONFIRM_HOSTILE",
  justification: "Confirmed by SIGINT cross-reference",
};

// ── Input validation ──────────────────────────────────────────────────────────

describe("submitFeedback — input validation", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("throws when trackId is empty", async () => {
    await expect(
      submitFeedback({ ...baseParams, trackId: "" }),
    ).rejects.toThrow("trackId is required");
    expect(mockSubmitFeedback).not.toHaveBeenCalled();
  });

  it("throws when trackId is only whitespace", async () => {
    await expect(
      submitFeedback({ ...baseParams, trackId: "   " }),
    ).rejects.toThrow("trackId is required");
  });

  it("throws when operatorId is empty", async () => {
    await expect(
      submitFeedback({ ...baseParams, operatorId: "" }),
    ).rejects.toThrow("operatorId is required");
    expect(mockSubmitFeedback).not.toHaveBeenCalled();
  });

  it("throws when justification is empty", async () => {
    await expect(
      submitFeedback({ ...baseParams, justification: "" }),
    ).rejects.toThrow("justification is required");
    expect(mockSubmitFeedback).not.toHaveBeenCalled();
  });

  it("throws when justification exceeds 500 characters", async () => {
    const longJustification = "x".repeat(501);
    await expect(
      submitFeedback({ ...baseParams, justification: longJustification }),
    ).rejects.toThrow("justification too long");
    expect(mockSubmitFeedback).not.toHaveBeenCalled();
  });

  it("accepts justification of exactly 500 characters", async () => {
    mockSubmitFeedback.mockResolvedValue({
      feedbackId: "fb-1",
      trustScore: 0.9,
      validated: true,
    });
    const exactly500 = "y".repeat(500);
    await expect(
      submitFeedback({ ...baseParams, justification: exactly500 }),
    ).resolves.toBeDefined();
    expect(mockSubmitFeedback).toHaveBeenCalledOnce();
  });

  it("trims leading/trailing whitespace from trackId and operatorId", async () => {
    mockSubmitFeedback.mockResolvedValue({
      feedbackId: "fb-2",
      trustScore: 0.8,
      validated: false,
    });

    await submitFeedback({
      ...baseParams,
      trackId: "  track-trim  ",
      operatorId: "  op-trim  ",
    });

    const [args] = mockSubmitFeedback.mock.calls[0] as [
      Record<string, unknown>,
    ];
    expect(args["trackId"]).toBe("track-trim");
    expect(args["operatorId"]).toBe("op-trim");
  });
});

// ── Successful response ───────────────────────────────────────────────────────

describe("submitFeedback — successful gRPC response", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSubmitFeedback.mockResolvedValue({
      feedbackId: "fb-001",
      trustScore: 0.95,
      validated: true,
    });
  });

  it("resolves with correctly mapped feedbackId, trustScore, validated", async () => {
    const result = await submitFeedback(baseParams);
    expect(result.feedbackId).toBe("fb-001");
    expect(result.trustScore).toBe(0.95);
    expect(result.validated).toBe(true);
  });

  it("calls the gRPC client with the correct feedbackType enum value", async () => {
    await submitFeedback({ ...baseParams, feedbackType: "CONFIRM_HOSTILE" });

    const [args] = mockSubmitFeedback.mock.calls[0] as [
      Record<string, unknown>,
    ];
    // FeedbackType.CONFIRM_HOSTILE = 1 (from mock)
    expect(args["feedbackType"]).toBe(1);
  });

  it("passes optional alertId when provided", async () => {
    await submitFeedback({ ...baseParams, alertId: "alert-99" });

    const [args] = mockSubmitFeedback.mock.calls[0] as [
      Record<string, unknown>,
    ];
    expect(args["alertId"]).toBe("alert-99");
  });

  it("passes undefined alertId when not provided", async () => {
    await submitFeedback(baseParams);

    const [args] = mockSubmitFeedback.mock.calls[0] as [
      Record<string, unknown>,
    ];
    expect(args["alertId"]).toBeUndefined();
  });
});

describe("alert quick-action payload adapters", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSubmitFeedback.mockResolvedValue({
      feedbackId: "fb-quick",
      trustScore: 0.7,
      validated: true,
    });
  });

  it("buildConfirmAlertFeedbackRequest maps to CONFIRM_ANOMALY with default justification", () => {
    const request = buildConfirmAlertFeedbackRequest({
      alertId: "alert-100",
      trackId: "track-100",
      operatorId: "op-100",
    });
    expect(request.feedbackType).toBe("CONFIRM_ANOMALY");
    expect(request.alertId).toBe("alert-100");
    expect(request.trackId).toBe("track-100");
    expect(request.justification).toContain("quick-action");
  });

  it("buildRejectAlertFeedbackRequest maps to REJECT_ANOMALY with default justification", () => {
    const request = buildRejectAlertFeedbackRequest({
      alertId: "alert-101",
      trackId: "track-101",
      operatorId: "op-101",
    });
    expect(request.feedbackType).toBe("REJECT_ANOMALY");
    expect(request.alertId).toBe("alert-101");
    expect(request.trackId).toBe("track-101");
    expect(request.justification).toContain("quick-action");
  });

  it("submitConfirmAlertFeedback sends CONFIRM_ANOMALY enum to gRPC request", async () => {
    await submitConfirmAlertFeedback({
      alertId: "alert-200",
      trackId: "track-200",
      operatorId: "op-200",
      justification: "Confirmed during triage",
    });

    const [args] = mockSubmitFeedback.mock.calls[0] as [
      Record<string, unknown>,
    ];
    expect(args["feedbackType"]).toBe(5);
    expect(args["alertId"]).toBe("alert-200");
    expect(args["justification"]).toBe("Confirmed during triage");
  });

  it("submitRejectAlertFeedback sends REJECT_ANOMALY enum to gRPC request", async () => {
    await submitRejectAlertFeedback({
      alertId: "alert-201",
      trackId: "track-201",
      operatorId: "op-201",
      justification: "False positive",
    });

    const [args] = mockSubmitFeedback.mock.calls[0] as [
      Record<string, unknown>,
    ];
    expect(args["feedbackType"]).toBe(4);
    expect(args["alertId"]).toBe("alert-201");
    expect(args["justification"]).toBe("False positive");
  });

  it("quick-action payload builders validate missing alertId", () => {
    expect(() =>
      buildConfirmAlertFeedbackRequest({
        alertId: "",
        trackId: "track-300",
        operatorId: "op-300",
      }),
    ).toThrow("alertId is required");
  });
});

// ── gRPC error propagation ────────────────────────────────────────────────────

describe("submitFeedback — gRPC error propagation", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("propagates a gRPC UNAVAILABLE error without swallowing it", async () => {
    mockSubmitFeedback.mockRejectedValue(new Error("gRPC UNAVAILABLE"));

    await expect(submitFeedback(baseParams)).rejects.toThrow(
      "gRPC UNAVAILABLE",
    );
  });

  it("propagates a gRPC PERMISSION_DENIED error", async () => {
    mockSubmitFeedback.mockRejectedValue(new Error("gRPC PERMISSION_DENIED"));

    await expect(submitFeedback(baseParams)).rejects.toThrow(
      "gRPC PERMISSION_DENIED",
    );
  });

  it("does not resolve when gRPC rejects", async () => {
    mockSubmitFeedback.mockRejectedValue(new Error("Network error"));

    let resolved = false;
    await submitFeedback(baseParams).then(
      () => {
        resolved = true;
      },
      () => {},
    );
    expect(resolved).toBe(false);
  });
});
