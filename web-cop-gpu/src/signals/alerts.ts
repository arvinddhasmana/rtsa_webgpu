// CLASSIFICATION: UNCLASSIFIED
// src/signals/alerts.ts — Alert list signal
//
// The Data Worker pushes alerts via postMessage; this module surfaces them as
// SolidJS signals for reactive consumption by the AlertSidebar component.
// Reference: docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md §3.3

import { createSignal } from "solid-js";
import type { AlertPayload } from "../workers/shared-protocol";

/** The current live alert list (most-recent first). */
export const [alerts, setAlerts] = createSignal<AlertPayload[]>([]);

/** Replace the full alert list (called when Data Worker emits alerts_updated). */
export function updateAlerts(incoming: AlertPayload[]): void {
  // Most-recent first; CRITICAL before others within same ms
  const sorted = [...incoming].sort((a, b) => {
    const sevOrder: Record<AlertPayload["severity"], number> = {
      CRITICAL: 0,
      ELEVATED: 1,
      WATCH: 2,
      NORMAL: 3,
      UNSPECIFIED: 4,
    };
    const sevDiff = sevOrder[a.severity] - sevOrder[b.severity];
    if (sevDiff !== 0) return sevDiff;
    return b.detectedAtMs - a.detectedAtMs;
  });
  setAlerts(sorted);
}

/** Mark an alert as acknowledged locally (optimistic update). */
export function acknowledgeAlertLocally(alertId: string): void {
  setAlerts((prev) =>
    prev.map((a) => (a.alertId === alertId ? { ...a, acknowledged: true } : a)),
  );
}
