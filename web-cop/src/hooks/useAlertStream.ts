// CLASSIFICATION: UNCLASSIFIED
// src/hooks/useAlertStream.ts

import { createPromiseClient } from "@connectrpc/connect";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  AlertSeverity as GenAlertSeverity,
  AnomalyType as GenAnomalyType,
  ClassificationLevel as GenClassificationLevel,
} from "../../../gen/ts/rtsa/common/v1/types_pb";
import { AlertService } from "../../../gen/ts/rtsa/inference/v1/alert_service_connect";
import { StreamAlertsRequest } from "../../../gen/ts/rtsa/inference/v1/alert_service_pb";
import { AnomalyAlert as ProtoAnomalyAlert } from "../../../gen/ts/rtsa/inference/v1/anomaly_alert_pb";
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

function cleanEnum(val: string, prefix: string): string {
  if (!val) return val;
  return val.replace(prefix, "");
}

function protoAlertToLocal(a: ProtoAnomalyAlert): AnomalyAlert {
  return {
    alertId: a.alertId,
    trackId: a.trackId,
    anomalyType: (cleanEnum(GenAnomalyType[a.anomalyType], "ANOMALY_TYPE_") ||
      "BEHAVIORAL") as AnomalyType,
    severity: (cleanEnum(GenAlertSeverity[a.severity], "ALERT_SEVERITY_") ||
      "WATCH") as AlertSeverity,
    confidenceScore: a.confidenceScore,
    explanation: a.explanation,
    features: (a.features || []).map((f) => ({
      featureName: f.featureName,
      value: f.value,
      contributionWeight: f.contributionWeight,
    })),
    classification: (cleanEnum(
      GenClassificationLevel[a.classification],
      "CLASSIFICATION_LEVEL_",
    ) || "UNCLASSIFIED") as ClassificationLevel,
    detectedAt: a.detectedAt ? a.detectedAt.toDate() : new Date(),
  };
}

interface StreamState {
  isConnected: boolean;
  error: Error | null;
  reconnectAttempts: number;
}

export function useAlertStream(): StreamState {
  const addAlert = useAlertStore((s) => s.addAlert);

  const [state, setState] = useState<StreamState>({
    isConnected: false,
    error: null,
    reconnectAttempts: 0,
  });

  // Proper refs to avoid stale closures and infinite loops
  const isMountedRef = useRef(true);
  const reconnectAttemptsRef = useRef(0);

  const abortControllerRef = useRef<AbortController | null>(null);
  const retryTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const connect = useCallback(() => {
    if (!isMountedRef.current) return;

    if (retryTimeoutRef.current) {
      clearTimeout(retryTimeoutRef.current);
      retryTimeoutRef.current = null;
    }

    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
    abortControllerRef.current = new AbortController();

    // Create new client instance
    const client = createPromiseClient(AlertService, transport);

    (async () => {
      try {
        const stream = client.streamAlerts(new StreamAlertsRequest(), {
          signal: abortControllerRef.current?.signal,
        });

        let firstMessageReceived = false;

        for await (const alert of stream) {
          if (!isMountedRef.current) break;

          // On first successful message, mark as fully connected and reset retry
          if (!firstMessageReceived) {
            firstMessageReceived = true;
            reconnectAttemptsRef.current = 0;
            setState({ isConnected: true, error: null, reconnectAttempts: 0 });
          }

          addAlert(protoAlertToLocal(alert));
        }
      } catch (err: any) {
        if (!isMountedRef.current) return;
        if (err.name === "AbortError") return;

        console.error("Alert stream error:", err);

        setState({
          isConnected: false,
          error: err,
          reconnectAttempts: reconnectAttemptsRef.current + 1,
        });

        reconnectAttemptsRef.current++;

        const delay = Math.min(
          BASE_DELAY_MS * Math.pow(1.5, reconnectAttemptsRef.current),
          MAX_DELAY_MS,
        );

        retryTimeoutRef.current = setTimeout(connect, delay);
      }
    })();
  }, [addAlert]);

  useEffect(() => {
    isMountedRef.current = true;
    connect();

    return () => {
      isMountedRef.current = false;
      if (abortControllerRef.current) abortControllerRef.current.abort();
      if (retryTimeoutRef.current) clearTimeout(retryTimeoutRef.current);
    };
  }, []);

  return state;
}
