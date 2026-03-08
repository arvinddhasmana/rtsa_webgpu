// CLASSIFICATION: UNCLASSIFIED
// src/services/alerts.test.ts — Unit tests for AlertService gRPC wrapper
//
// Tests:
//   - startAlertStream emits properly typed alert objects via updateAlerts
//   - startAlertStream handles AbortError gracefully (expected cancellation)
//   - startAlertStream calls setAlertStreamHealthy(false) on unexpected errors
//   - acknowledgeAlert calls gRPC client with correct alertId and operatorId
//   - acknowledgeAlert propagates gRPC errors without swallowing them

import { describe, it, expect, vi, beforeEach } from "vitest";

// ── Mock ConnectRPC client at the transport boundary ──────────────────────────

const mockStreamAlerts = vi.hoisted(() => vi.fn());
const mockAcknowledgeAlert = vi.hoisted(() => vi.fn());

vi.mock("@connectrpc/connect", () => ({
  createPromiseClient: () => ({
    streamAlerts: mockStreamAlerts,
    acknowledgeAlert: mockAcknowledgeAlert,
  }),
}));

vi.mock("./grpc-client", () => ({ transport: {} }));

// Mock the generated service descriptor (value is unused by the mock client)
vi.mock("@gen/rtsa/inference/v1/alert_service_connect.js", () => ({
  AlertService: {},
}));

// Mock enum values used by the service
vi.mock("@gen/rtsa/common/v1/types_pb.js", () => ({
  ClassificationLevel: { UNCLASSIFIED: 1 },
  FeedbackType: {},
  EntityType: {},
  HostileClassification: {},
  TrackStatus: {},
}));

// Mock the signals to observe calls without triggering SolidJS reactivity
const mockUpdateAlerts = vi.hoisted(() => vi.fn());
const mockAcknowledgeAlertLocally = vi.hoisted(() => vi.fn());
vi.mock("../signals/alerts", () => ({
  updateAlerts: mockUpdateAlerts,
  acknowledgeAlertLocally: mockAcknowledgeAlertLocally,
}));

const mockSetAlertStreamHealthy = vi.hoisted(() => vi.fn());
vi.mock("../signals/connection", () => ({
  setAlertStreamHealthy: mockSetAlertStreamHealthy,
  setWtConnected: vi.fn(),
  setGrpcConnected: vi.fn(),
  setConnecting: vi.fn(),
}));

import { startAlertStream, acknowledgeAlert } from "./alerts";

// ── startAlertStream ──────────────────────────────────────────────────────────

describe("startAlertStream", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("returns an AbortController", () => {
    mockStreamAlerts.mockReturnValue(
      (async function* () {
        // empty stream
      })(),
    );
    const ac = startAlertStream();
    expect(ac).toBeInstanceOf(AbortController);
    ac.abort();
  });

  it("calls updateAlerts with a correctly typed AlertPayload for each streamed alert", async () => {
    const detectedAtSeconds = BigInt(1_700_000_000);

    mockStreamAlerts.mockReturnValue(
      (async function* () {
        yield {
          alertId: "alert-001",
          trackId: "track-abc",
          severity: 4, // CRITICAL
          explanation: "Hostile track detected",
          detectedAt: { seconds: detectedAtSeconds },
          acknowledged: false,
        };
      })(),
    );

    startAlertStream();

    // Wait for the async generator to complete
    await new Promise<void>((resolve) => setTimeout(resolve, 20));

    expect(mockUpdateAlerts).toHaveBeenCalledOnce();
    const [payloads] = mockUpdateAlerts.mock.calls[0] as [unknown[]];
    expect(payloads).toHaveLength(1);
    const payload = payloads[0] as Record<string, unknown>;
    expect(payload["alertId"]).toBe("alert-001");
    expect(payload["trackId"]).toBe("track-abc");
    expect(payload["severity"]).toBe("CRITICAL");
    expect(payload["description"]).toBe("Hostile track detected");
    expect(payload["acknowledged"]).toBe(false);
    expect(typeof payload["detectedAtMs"]).toBe("number");
  });

  it("maps severity 3 to ELEVATED", async () => {
    mockStreamAlerts.mockReturnValue(
      (async function* () {
        yield {
          alertId: "alert-002",
          trackId: "track-xyz",
          severity: 3,
          explanation: "Elevated threat",
          detectedAt: null,
          acknowledged: false,
        };
      })(),
    );

    startAlertStream();
    await new Promise<void>((resolve) => setTimeout(resolve, 20));

    const [payloads] = mockUpdateAlerts.mock.calls[0] as [unknown[]];
    const payload = payloads[0] as Record<string, unknown>;
    expect(payload["severity"]).toBe("ELEVATED");
  });

  it("upserts an alert with the same alertId instead of duplicating", async () => {
    let callCount = 0;
    mockStreamAlerts.mockReturnValue(
      (async function* () {
        // First emission
        yield {
          alertId: "dupe-001",
          trackId: "t1",
          severity: 2,
          explanation: "Initial",
          detectedAt: null,
          acknowledged: false,
        };
        // Second emission — same alertId, updated description
        yield {
          alertId: "dupe-001",
          trackId: "t1",
          severity: 2,
          explanation: "Updated",
          detectedAt: null,
          acknowledged: true,
        };
      })(),
    );

    startAlertStream();
    await new Promise<void>((resolve) => setTimeout(resolve, 20));

    // updateAlerts is called twice; the second call should still have only 1 entry
    expect(mockUpdateAlerts).toHaveBeenCalledTimes(2);
    const secondCall = mockUpdateAlerts.mock.calls[1] as [unknown[]];
    expect(secondCall[0]).toHaveLength(1);
    const payload = secondCall[0][0] as Record<string, unknown>;
    expect(payload["description"]).toBe("Updated");
    expect(payload["acknowledged"]).toBe(true);
    callCount++;
    expect(callCount).toBe(1);
  });

  it("does not call setAlertStreamHealthy on AbortError", async () => {
    const abortError = new DOMException("Aborted", "AbortError");
    mockStreamAlerts.mockReturnValue(
      (async function* () {
        throw abortError;
      })(),
    );

    startAlertStream();
    await new Promise<void>((resolve) => setTimeout(resolve, 20));

    expect(mockSetAlertStreamHealthy).not.toHaveBeenCalled();
  });

  it("calls setAlertStreamHealthy(false) on unexpected stream errors", async () => {
    mockStreamAlerts.mockReturnValue(
      (async function* () {
        throw new Error("Network failure");
      })(),
    );

    startAlertStream();
    await new Promise<void>((resolve) => setTimeout(resolve, 20));

    expect(mockSetAlertStreamHealthy).toHaveBeenCalledWith(false);
  });
});

// ── acknowledgeAlert ──────────────────────────────────────────────────────────

describe("acknowledgeAlert", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls acknowledgeAlertLocally for optimistic update before gRPC call", async () => {
    mockAcknowledgeAlert.mockResolvedValue({});

    await acknowledgeAlert("alert-1", "op-42");

    expect(mockAcknowledgeAlertLocally).toHaveBeenCalledWith("alert-1");
  });

  it("calls gRPC client with correct alertId and operatorId parameters", async () => {
    mockAcknowledgeAlert.mockResolvedValue({});

    await acknowledgeAlert("alert-99", "operator-007", "confirmed hostile");

    expect(mockAcknowledgeAlert).toHaveBeenCalledWith({
      alertId: "alert-99",
      operatorId: "operator-007",
      comment: "confirmed hostile",
    });
  });

  it("defaults to empty comment when comment is not provided", async () => {
    mockAcknowledgeAlert.mockResolvedValue({});

    await acknowledgeAlert("a1", "op1");

    const [args] = mockAcknowledgeAlert.mock.calls[0] as [Record<string, unknown>];
    expect(args["comment"]).toBe("");
  });

  it("propagates gRPC errors without swallowing them", async () => {
    const grpcError = new Error("gRPC UNAVAILABLE");
    mockAcknowledgeAlert.mockRejectedValue(grpcError);

    await expect(acknowledgeAlert("a1", "op1")).rejects.toThrow("gRPC UNAVAILABLE");
  });
});
