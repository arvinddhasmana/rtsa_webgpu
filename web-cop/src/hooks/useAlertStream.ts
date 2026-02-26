// CLASSIFICATION: UNCLASSIFIED
// src/hooks/useAlertStream.ts

import { toJson } from "@bufbuild/protobuf";
import { Code, ConnectError, createClient } from "@connectrpc/connect";
import { AlertService } from "@gen/rtsa/inference/v1/alert_service_pb";
import { AnomalyAlertSchema } from "@gen/rtsa/inference/v1/anomaly_alert_pb";
import { useCallback, useEffect, useRef, useState } from "react";
import { transport } from "../api/grpc-client";
import { useAlertStore } from "../stores/alertStore";
import type { AnomalyAlert } from "../types/alert";
import type {
  AlertSeverity,
  AnomalyType,
  ClassificationLevel,
} from "../types/common";

const BASE_DELAY_MS = 1000;
const MAX_DELAY_MS = 30000;

const alertClient = createClient(AlertService, transport);

function stripPrefix(val: string | undefined, prefix: string): string {
  if (!val) return "";
  return val.startsWith(prefix) ? val.slice(prefix.length) : val;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function protoAlertToLocal(protoAlert: any): AnomalyAlert {
  const j = toJson(AnomalyAlertSchema, protoAlert, {
    alwaysEmitImplicit: true,
  }) as Record<string, unknown>;
  return {
    alertId: (j["alertId"] ?? "") as string,
    trackId: (j["trackId"] ?? "") as string,
    anomalyType: (stripPrefix(j["anomalyType"] as string, "ANOMALY_TYPE_") ||
      "BEHAVIORAL") as AnomalyType,
    severity: (stripPrefix(j["severity"] as string, "ALERT_SEVERITY_") ||
      "WATCH") as AlertSeverity,
    confidenceScore: (j["confidenceScore"] ?? 0) as number,
    explanation: (j["explanation"] ?? "") as string,
    features: ((j["features"] ?? []) as Array<Record<string, unknown>>).map(
      (f) => ({
        featureName: (f["featureName"] ?? "") as string,
        value: (f["value"] ?? 0) as number,
        contributionWeight: (f["contributionWeight"] ?? 0) as number,
      }),
    ),
    classification: (stripPrefix(
      j["classification"] as string,
      "CLASSIFICATION_LEVEL_",
    ) || "UNCLASSIFIED") as ClassificationLevel,
    detectedAt: j["detectedAt"]
      ? new Date(j["detectedAt"] as string)
      : new Date(),
  };
}

interface StreamState {
  isConnected: boolean;
  error: Error | null;
  reconnectAttempts: number;
}

/**
 * useAlertStream — subscribes to real-time anomaly alerts via gRPC-Web server-streaming.
 *
 * Uses the generated AlertService client with @connectrpc/connect-web transport
 * for proper binary protobuf framing and gRPC-Web envelope encoding.
 *
 * Flow:
 *   1. Opens StreamAlerts gRPC-Web stream to svc-alert via Envoy
 *   2. Maps proto AnomalyAlert → local AnomalyAlert type and updates AlertStore
 *   3. On connect error: exponential backoff reconnect (1s, 2s … max 30s)
 */
export function useAlertStream(): StreamState {
  const addAlert = useAlertStore((s) => s.addAlert);

  const [state, setState] = useState<StreamState>({
    isConnected: false,
    error: null,
    reconnectAttempts: 0,
  });

  const abortRef = useRef<AbortController | null>(null);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const attemptsRef = useRef(0);
  const mountedRef = useRef(true);
  const connectedRef = useRef(false);

  const connect = useCallback(() => {
    if (!mountedRef.current) return;

    abortRef.current = new AbortController();
    const signal = abortRef.current.signal;

    void (async () => {
      try {
        for await (const alert of alertClient.streamAlerts({}, { signal })) {
          if (!mountedRef.current) return;

          if (!connectedRef.current) {
            connectedRef.current = true;
            attemptsRef.current = 0;
            setState({ isConnected: true, error: null, reconnectAttempts: 0 });
          }

          addAlert(protoAlertToLocal(alert));
        }
        if (mountedRef.current) throw new Error("Alert stream ended");
      } catch (err) {
        if (!mountedRef.current) return;
        if (err instanceof ConnectError && err.code === Code.Canceled) return;

        connectedRef.current = false;

        setState((s) => ({
          isConnected: false,
          error: err instanceof Error ? err : new Error(String(err)),
          reconnectAttempts: s.reconnectAttempts + 1,
        }));
        attemptsRef.current += 1;

        const delay = Math.min(
          BASE_DELAY_MS * Math.pow(2, attemptsRef.current - 1),
          MAX_DELAY_MS,
        );
        timeoutRef.current = setTimeout(connect, delay);
      }
    })();
  }, [addAlert]);

  useEffect(() => {
    mountedRef.current = true;
    connect();

    return () => {
      mountedRef.current = false;
      connectedRef.current = false;
      abortRef.current?.abort();
      if (timeoutRef.current !== null) clearTimeout(timeoutRef.current);
    };
  }, [connect]);

  return state;
}
