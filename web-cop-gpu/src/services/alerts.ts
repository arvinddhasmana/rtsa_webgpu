// CLASSIFICATION: UNCLASSIFIED
// src/services/alerts.ts — AlertService gRPC calls
//
// Handles streaming alerts subscription and alert acknowledgement via gRPC cold path.
// Reference: docs/implementation/v4/phase3_ui_interaction.md §5

import { createPromiseClient } from "@connectrpc/connect";
import { AlertService } from "@gen/rtsa/inference/v1/alert_service_connect.js";
import { transport } from "./grpc-client";
import { ClassificationLevel } from "@gen/rtsa/common/v1/types_pb.js";
import type { AlertPayload } from "../workers/shared-protocol";
import { updateAlerts, acknowledgeAlertLocally } from "../signals/alerts";
import { setAlertStreamHealthy } from "../signals/connection";

const client = createPromiseClient(AlertService, transport);

/**
 * Start streaming alerts from the AlertService over gRPC-Web.
 * Drives the `alerts` signal when new alerts arrive.
 * Returns an AbortController to cancel the stream.
 */
export function startAlertStream(): AbortController {
  const ac = new AbortController();

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
