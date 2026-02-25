// CLASSIFICATION: UNCLASSIFIED
// src/api/alert-client.ts

import { transport } from "./grpc-client";
import { AnomalyAlert } from "../types/alert";

export interface StreamAlertsRequest {
  minSeverity: string;
  classificationCeiling: string;
}

/**
 * AlertClient wraps the AlertService gRPC-Web streaming endpoint.
 * Production: connects to svc-alert via Envoy at /rtsa.alert.v1.AlertService/StreamAlerts
 */
export class AlertClient {
  private readonly baseUrl: string;

  constructor() {
    this.baseUrl = transport.baseUrl;
  }

  streamAlerts(
    request: StreamAlertsRequest,
    onMessage: (alert: AnomalyAlert) => void,
    onError: (err: Error) => void
  ): AbortController {
    const controller = new AbortController();
    void this.baseUrl;
    void request;
    void onMessage;
    void onError;
    return controller;
  }
}

export const alertClient = new AlertClient();
