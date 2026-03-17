// CLASSIFICATION: UNCLASSIFIED
// src/services/alerts.ts — AlertService gRPC calls
//
// Handles streaming alerts subscription and alert acknowledgement via gRPC cold path.
// Reference: docs/implementation/v4/phase3_ui_interaction.md §5

import { createPromiseClient } from "@connectrpc/connect";
import { ClassificationLevel } from "@gen/rtsa/common/v1/types_pb.js";
import { AlertService } from "@gen/rtsa/inference/v1/alert_service_connect.js";
import { acknowledgeAlertLocally, alerts, updateAlerts } from "../signals/alerts";
import { setAlertStreamHealthy } from "../signals/connection";
import type { AlertPayload } from "../workers/shared-protocol";
import { transport } from "./grpc-client";

const client = createPromiseClient(AlertService, transport);

type E2EAlertMocks = {
  assignAlert?: (request: {
    alertId: string;
    assignerOperatorId: string;
    assigneeOperatorId: string;
    comment: string;
  }) => Promise<{
    success: boolean;
    assignedAt?: { seconds: bigint | number };
  }>;
};

function getE2EAlertMocks(): E2EAlertMocks | undefined {
  const maybeGlobal = globalThis as typeof globalThis & {
    __RTSA_E2E_MOCKS__?: E2EAlertMocks;
  };
  return maybeGlobal.__RTSA_E2E_MOCKS__;
}

export interface AssignAlertParams {
  alertId: string;
  assignerOperatorId: string;
  assigneeOperatorId: string;
  comment?: string;
}

export interface AssignAlertResult {
  success: boolean;
  assignedAtMs: number;
}

/** Build and validate assign request payload for AlertService.AssignAlert. */
export function buildAssignAlertRequest(params: AssignAlertParams): {
  alertId: string;
  assignerOperatorId: string;
  assigneeOperatorId: string;
  comment: string;
} {
  const alertId = params.alertId.trim();
  const assignerOperatorId = params.assignerOperatorId.trim();
  const assigneeOperatorId = params.assigneeOperatorId.trim();
  const comment = (params.comment ?? "").trim();

  if (!alertId) throw new Error("alertId is required");
  if (!assignerOperatorId) throw new Error("assignerOperatorId is required");
  if (!assigneeOperatorId) throw new Error("assigneeOperatorId is required");
  if (assigneeOperatorId === assignerOperatorId) {
    throw new Error(
      "assigneeOperatorId must be different from assignerOperatorId",
    );
  }

  return {
    alertId,
    assignerOperatorId,
    assigneeOperatorId,
    comment,
  };
}

/**
 * Start streaming alerts from the AlertService over gRPC-Web.
 * Drives the `alerts` signal when new alerts arrive.
 * Returns an AbortController to cancel the stream.
 */
export function startAlertStream(): AbortController {
  const ac = new AbortController();

  if (import.meta.env.VITE_MOCK_ALERTS === "true" || (import.meta.env.DEV && !import.meta.env.VITE_API_GATEWAY_URL)) {
    startMockAlertStream(ac);
    return ac;
  }

  void (async () => {
    try {
      const accumulated: AlertPayload[] = [];
      const stream = client.streamAlerts(
        { clearanceLevel: ClassificationLevel.UNCLASSIFIED },
        { signal: ac.signal },
      );

      for await (const alert of stream) {
        const payload: AlertPayload = {
          alertId: alert.alertId,
          trackId: alert.trackId,
          severity: mapSeverity(alert.severity),
          description: alert.explanation,
          detectedAtMs: alert.detectedAt
            ? Number(alert.detectedAt.seconds) * 1000
            : Date.now(),
          acknowledged: alert.acknowledged,
        };

        // Upsert: replace existing alert with same ID or add new
        const idx = accumulated.findIndex((a) => a.alertId === payload.alertId);
        if (idx >= 0) {
          accumulated[idx] = payload;
        } else {
          accumulated.push(payload);
        }
        // Keep latest 200 alerts
        if (accumulated.length > 200) {
          accumulated.splice(0, accumulated.length - 200);
        }
        updateAlerts([...accumulated]);
      }
    } catch (err) {
      // AbortError is expected when the stream is cancelled
      if (err instanceof Error && err.name !== "AbortError") {
        // Signal stream unhealthy so UI can surface the failure without console.* in production.
        setAlertStreamHealthy(false);
        if (import.meta.env.DEV) {
          console.error("[AlertService] Stream error:", err);
        }
      }
    }
  })();

  return ac;
}

function startMockAlertStream(ac: AbortController) {
  const generate = (id: string, severity: AlertSeverity, desc: string, track: string) => ({
    alertId: id,
    severity,
    description: desc,
    trackId: track,
    detectedAtMs: Date.now() - Math.random() * 3600000,
    acknowledged: Math.random() > 0.7,
  });

  const initialAlerts: AlertPayload[] = [
    generate("mock-1", "CRITICAL", "[MOCK] Unidentified signature detected - Sector 4", "track-v-101"),
    generate("mock-2", "ELEVATED", "[MOCK] AIS/Radar mismatch - Commercial Vessel", "track-v-202"),
    generate("mock-3", "WATCH", "[MOCK] Fast move detected - Boundary crossing", "track-v-303"),
    generate("mock-4", "CRITICAL", "[MOCK] Electronic interference localized - Carrier group", "track-v-404"),
    generate("mock-5", "NORMAL", "[MOCK] Routine port arrival - Tanker Alpha", "track-v-505"),
  ];

  updateAlerts(initialAlerts);

  const interval = setInterval(() => {
    if (ac.signal.aborted) {
      clearInterval(interval);
      return;
    }
    const newAlert = generate(
      `mock-${Date.now()}`,
      Math.random() > 0.8 ? "CRITICAL" : "ELEVATED",
      "[MOCK] Real-time sensor anomaly detected",
      `track-v-${Math.floor(Math.random() * 1000)}`
    );
    updateAlerts([newAlert, ...accumulated_alerts_logic_here()]);
  }, 30000);
}

// Helper to get current alerts to append to (simplified for mock)
function accumulated_alerts_logic_here(): AlertPayload[] {
  return alerts();
}

/** Acknowledge an alert via gRPC and update the local signal optimistically. */
export async function acknowledgeAlert(
  alertId: string,
  operatorId: string,
  comment = "",
): Promise<void> {
  // Optimistic update
  acknowledgeAlertLocally(alertId);
  try {
    await client.acknowledgeAlert({ alertId, operatorId, comment });
  } catch (err) {
    if (import.meta.env.DEV) {
      console.error("[AlertService] Acknowledge failed:", err);
    }
    throw err;
  }
}

/** Assign an alert to another operator via AlertService.AssignAlert. */
export async function assignAlert(
  params: AssignAlertParams,
): Promise<AssignAlertResult> {
  const request = buildAssignAlertRequest(params);
  const response = getE2EAlertMocks()?.assignAlert
    ? await getE2EAlertMocks()!.assignAlert!(request)
    : await client.assignAlert(request);
  return {
    success: response.success,
    assignedAtMs: response.assignedAt
      ? Number(response.assignedAt.seconds) * 1000
      : Date.now(),
  };
}

type AlertSeverity = AlertPayload["severity"];

function mapSeverity(v: number): AlertSeverity {
  // AlertSeverity proto enum: 0=UNSPECIFIED, 1=NORMAL, 2=WATCH, 3=ELEVATED, 4=CRITICAL
  const map: Record<number, AlertSeverity> = {
    0: "UNSPECIFIED",
    1: "NORMAL",
    2: "WATCH",
    3: "ELEVATED",
    4: "CRITICAL",
  };
  return map[v] ?? "UNSPECIFIED";
}
