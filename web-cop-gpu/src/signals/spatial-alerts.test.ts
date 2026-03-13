// CLASSIFICATION: UNCLASSIFIED
// src/signals/spatial-alerts.test.ts — Unit tests for the spatial alerts signal
//
// Tests:
//   - Initial state: spatialAlerts() is empty, activeSpatialAlertId() is null
//   - setSpatialAlerts replaces the alert list
//   - setActiveSpatialAlertId sets and clears the active alert
//   - Multiple alerts are stored in the order provided

import { describe, it, expect, beforeEach } from "vitest";

// Import the module under test.
// Using dynamic import to get a fresh module state per describe block.
import {
  spatialAlerts,
  setSpatialAlerts,
  activeSpatialAlertId,
  setActiveSpatialAlertId,
  type SpatialAlertPayload,
} from "./spatial-alerts";

const makeAlert = (overrides: Partial<SpatialAlertPayload> = {}): SpatialAlertPayload => ({
  alertId: "gap-001",
  sectorId: "NW-4",
  affectedSensorId: "RADAR-07",
  severity: "CRITICAL",
  description: "Data gap in sector NW-4",
  lastContactUtc: "2026-03-13T14:00:00.000Z",
  acknowledged: false,
  areaPolygon: [
    { lat: 51.0, lon: -6.0 },
    { lat: 52.0, lon: -6.0 },
    { lat: 52.0, lon: -4.0 },
    { lat: 51.0, lon: -4.0 },
  ],
  ...overrides,
});

describe("spatialAlerts signal", () => {
  beforeEach(() => {
    // Reset to empty state before each test
    setSpatialAlerts([]);
    setActiveSpatialAlertId(null);
  });

  it("starts with an empty alert list", () => {
    expect(spatialAlerts()).toHaveLength(0);
  });

  it("setSpatialAlerts replaces the alert list with provided alerts", () => {
    const alerts = [makeAlert({ alertId: "gap-001" }), makeAlert({ alertId: "gap-002", sectorId: "SE-1" })];
    setSpatialAlerts(alerts);
    expect(spatialAlerts()).toHaveLength(2);
    expect(spatialAlerts()[0].alertId).toBe("gap-001");
    expect(spatialAlerts()[1].alertId).toBe("gap-002");
  });

  it("setSpatialAlerts with empty array clears the list", () => {
    setSpatialAlerts([makeAlert()]);
    expect(spatialAlerts()).toHaveLength(1);
    setSpatialAlerts([]);
    expect(spatialAlerts()).toHaveLength(0);
  });

  it("alert payload preserves all fields", () => {
    const alert = makeAlert({
      alertId: "gap-xyz",
      sectorId: "NW-4",
      affectedSensorId: "RADAR-07",
      severity: "ELEVATED",
      description: "Partial data gap in sector NW-4",
      lastContactUtc: "2026-03-13T14:00:00.000Z",
      acknowledged: true,
      areaPolygon: [
        { lat: 50.0, lon: -7.0 },
        { lat: 51.0, lon: -7.0 },
      ],
    });
    setSpatialAlerts([alert]);
    const stored = spatialAlerts()[0];
    expect(stored.alertId).toBe("gap-xyz");
    expect(stored.sectorId).toBe("NW-4");
    expect(stored.affectedSensorId).toBe("RADAR-07");
    expect(stored.severity).toBe("ELEVATED");
    expect(stored.description).toBe("Partial data gap in sector NW-4");
    expect(stored.acknowledged).toBe(true);
    expect(stored.areaPolygon).toHaveLength(2);
  });

  it("stores alerts in the exact order provided", () => {
    const alerts: SpatialAlertPayload[] = [
      makeAlert({ alertId: "gap-A", severity: "WATCH" }),
      makeAlert({ alertId: "gap-B", severity: "CRITICAL" }),
      makeAlert({ alertId: "gap-C", severity: "ELEVATED" }),
    ];
    setSpatialAlerts(alerts);
    expect(spatialAlerts().map((a) => a.alertId)).toEqual(["gap-A", "gap-B", "gap-C"]);
  });
});

describe("activeSpatialAlertId signal", () => {
  beforeEach(() => {
    setActiveSpatialAlertId(null);
  });

  it("starts as null", () => {
    expect(activeSpatialAlertId()).toBeNull();
  });

  it("setActiveSpatialAlertId sets the active alert ID", () => {
    setActiveSpatialAlertId("gap-001");
    expect(activeSpatialAlertId()).toBe("gap-001");
  });

  it("setActiveSpatialAlertId can be cleared back to null", () => {
    setActiveSpatialAlertId("gap-001");
    setActiveSpatialAlertId(null);
    expect(activeSpatialAlertId()).toBeNull();
  });

  it("setActiveSpatialAlertId replaces a previous value", () => {
    setActiveSpatialAlertId("gap-001");
    setActiveSpatialAlertId("gap-002");
    expect(activeSpatialAlertId()).toBe("gap-002");
  });
});
