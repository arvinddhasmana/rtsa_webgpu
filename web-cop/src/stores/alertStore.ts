// CLASSIFICATION: UNCLASSIFIED
// src/stores/alertStore.ts

import { create } from "zustand";
import { AnomalyAlert } from "../types/alert";

function severityRank(s: string): number {
  switch (s) {
    case "CRITICAL":
      return 3;
    case "ELEVATED":
      return 2;
    case "WATCH":
      return 1;
    default:
      return 0;
  }
}

interface AlertState {
  alerts: Map<string, AnomalyAlert>;
  acknowledgedIds: Set<string>;
  minSeverityFilter: string;

  addAlert: (alert: AnomalyAlert) => void;
  acknowledgeAlert: (alertId: string) => void;
  setMinSeverityFilter: (severity: string) => void;
  clearAll: () => void;

  getUnacknowledgedAlerts: () => AnomalyAlert[];
  getCriticalCount: () => number;
  getFilteredAlerts: () => AnomalyAlert[];
}

export const useAlertStore = create<AlertState>((set, get) => ({
  alerts: new Map(),
  acknowledgedIds: new Set(),
  minSeverityFilter: "WATCH",

  addAlert: (alert) =>
    set((state) => {
      const newAlerts = new Map(state.alerts);
      newAlerts.set(alert.alertId, alert);
      return { alerts: newAlerts };
    }),

  acknowledgeAlert: (alertId) =>
    set((state) => {
      const newAcked = new Set(state.acknowledgedIds);
      newAcked.add(alertId);
      return { acknowledgedIds: newAcked };
    }),

  setMinSeverityFilter: (severity) => set({ minSeverityFilter: severity }),
  clearAll: () => set({ alerts: new Map(), acknowledgedIds: new Set() }),

  getUnacknowledgedAlerts: () =>
    Array.from(get().alerts.values())
      .filter((a) => !get().acknowledgedIds.has(a.alertId))
      .sort((a, b) => severityRank(b.severity) - severityRank(a.severity)),

  getCriticalCount: () =>
    Array.from(get().alerts.values()).filter(
      (a) =>
        a.severity === "CRITICAL" && !get().acknowledgedIds.has(a.alertId)
    ).length,

  getFilteredAlerts: () => {
    const minRank = severityRank(get().minSeverityFilter);
    return Array.from(get().alerts.values())
      .filter((a) => severityRank(a.severity) >= minRank)
      .sort((a, b) => severityRank(b.severity) - severityRank(a.severity));
  },
}));
