// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/stores/alertStore.test.ts

import { describe, it, expect, beforeEach } from "vitest";
import { useAlertStore } from "../../stores/alertStore";
import { AnomalyAlert } from "../../types/alert";

function makeAlert(overrides: Partial<AnomalyAlert> = {}): AnomalyAlert {
  return {
    alertId: "ALT-001",
    trackId: "TRK-001",
    anomalyType: "SPEED",
    severity: "WATCH",
    confidenceScore: 0.8,
    explanation: "Speed anomaly detected",
    features: [],
    classification: "UNCLASSIFIED",
    detectedAt: new Date(),
    ...overrides,
  };
}

describe("AlertStore", () => {
  beforeEach(() => {
    useAlertStore.getState().clearAll();
    useAlertStore.getState().setMinSeverityFilter("WATCH");
  });

  it("T03: addAlert adds alert to store", () => {
    const alert = makeAlert({ alertId: "A1" });
    useAlertStore.getState().addAlert(alert);
    expect(useAlertStore.getState().alerts.size).toBe(1);
    expect(useAlertStore.getState().alerts.get("A1")).toEqual(alert);
  });

  it("T04: acknowledgeAlert adds alertId to acknowledgedIds", () => {
    useAlertStore.getState().addAlert(makeAlert({ alertId: "A1" }));
    useAlertStore.getState().acknowledgeAlert("A1");
    expect(useAlertStore.getState().acknowledgedIds.has("A1")).toBe(true);
  });

  it("T04: acknowledged alerts excluded from getUnacknowledgedAlerts", () => {
    useAlertStore.getState().addAlert(makeAlert({ alertId: "A1" }));
    useAlertStore.getState().addAlert(makeAlert({ alertId: "A2" }));
    useAlertStore.getState().acknowledgeAlert("A1");
    const unacked = useAlertStore.getState().getUnacknowledgedAlerts();
    expect(unacked.map((a) => a.alertId)).not.toContain("A1");
    expect(unacked.map((a) => a.alertId)).toContain("A2");
  });

  it("setMinSeverityFilter updates the filter", () => {
    useAlertStore.getState().setMinSeverityFilter("CRITICAL");
    expect(useAlertStore.getState().minSeverityFilter).toBe("CRITICAL");
  });

  it("T03: getFilteredAlerts filters by severity", () => {
    useAlertStore.getState().addAlert(makeAlert({ alertId: "W1", severity: "WATCH" }));
    useAlertStore.getState().addAlert(makeAlert({ alertId: "E1", severity: "ELEVATED" }));
    useAlertStore.getState().addAlert(makeAlert({ alertId: "C1", severity: "CRITICAL" }));

    useAlertStore.getState().setMinSeverityFilter("ELEVATED");
    const filtered = useAlertStore.getState().getFilteredAlerts();
    expect(filtered.map((a) => a.alertId)).not.toContain("W1");
    expect(filtered.map((a) => a.alertId)).toContain("E1");
    expect(filtered.map((a) => a.alertId)).toContain("C1");
  });

  it("getFilteredAlerts sorts CRITICAL first", () => {
    useAlertStore.getState().addAlert(makeAlert({ alertId: "W1", severity: "WATCH" }));
    useAlertStore.getState().addAlert(makeAlert({ alertId: "C1", severity: "CRITICAL" }));
    useAlertStore.getState().addAlert(makeAlert({ alertId: "E1", severity: "ELEVATED" }));

    useAlertStore.getState().setMinSeverityFilter("WATCH");
    const filtered = useAlertStore.getState().getFilteredAlerts();
    expect(filtered[0].severity).toBe("CRITICAL");
  });

  it("getCriticalCount counts unacknowledged critical alerts", () => {
    useAlertStore.getState().addAlert(makeAlert({ alertId: "C1", severity: "CRITICAL" }));
    useAlertStore.getState().addAlert(makeAlert({ alertId: "C2", severity: "CRITICAL" }));
    useAlertStore.getState().acknowledgeAlert("C1");

    expect(useAlertStore.getState().getCriticalCount()).toBe(1);
  });

  it("clearAll resets all state", () => {
    useAlertStore.getState().addAlert(makeAlert());
    useAlertStore.getState().acknowledgeAlert("ALT-001");
    useAlertStore.getState().clearAll();

    expect(useAlertStore.getState().alerts.size).toBe(0);
    expect(useAlertStore.getState().acknowledgedIds.size).toBe(0);
  });

  it("getUnacknowledgedAlerts returns empty when all acknowledged", () => {
    useAlertStore.getState().addAlert(makeAlert({ alertId: "A1" }));
    useAlertStore.getState().acknowledgeAlert("A1");
    expect(useAlertStore.getState().getUnacknowledgedAlerts()).toHaveLength(0);
  });
});
