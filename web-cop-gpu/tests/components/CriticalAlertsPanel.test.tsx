// CLASSIFICATION: UNCLASSIFIED
// tests/components/CriticalAlertsPanel.test.tsx

import { fireEvent, render, screen } from "@solidjs/testing-library";
import { describe, expect, it, vi } from "vitest";
import { CriticalAlertsPanel } from "../../src/components/dashboard/CriticalAlertsPanel";
import type { SpatialAlertPayload } from "../../src/signals/spatial-alerts";

const alerts: SpatialAlertPayload[] = [
  {
    alertId: "gap-001",
    sectorId: "NW-4",
    affectedSensorId: "RADAR-WEST-04",
    severity: "CRITICAL",
    description: "Data gap in sector NW-4",
    lastContactUtc: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
    acknowledged: false,
    areaPolygon: [],
  },
  {
    alertId: "gap-002",
    sectorId: "NE-2",
    affectedSensorId: "ISR-01",
    severity: "ELEVATED",
    description: "ISR offline",
    lastContactUtc: new Date(Date.now() - 8 * 60 * 1000).toISOString(),
    acknowledged: false,
    areaPolygon: [],
  },
];

describe("CriticalAlertsPanel", () => {
  it("renders panel container", () => {
    render(() => <CriticalAlertsPanel spatialAlerts={alerts} />);
    expect(screen.getByTestId("critical-alerts-panel")).toBeDefined();
  });

  it("renders all alert items", () => {
    render(() => <CriticalAlertsPanel spatialAlerts={alerts} />);
    expect(screen.getByTestId("alert-item-gap-001")).toBeDefined();
    expect(screen.getByTestId("alert-item-gap-002")).toBeDefined();
  });

  it("shows sensor IDs in alerts", () => {
    render(() => <CriticalAlertsPanel spatialAlerts={alerts} />);
    expect(screen.getByText("RADAR-WEST-04")).toBeDefined();
    expect(screen.getByText("ISR-01")).toBeDefined();
  });

  it("renders 'No active alerts' when empty", () => {
    render(() => <CriticalAlertsPanel spatialAlerts={[]} />);
    expect(screen.getByText("No active alerts")).toBeDefined();
  });

  it("calls onAlertClick when alert is clicked", () => {
    const onClick = vi.fn();
    render(() => <CriticalAlertsPanel spatialAlerts={alerts} onAlertClick={onClick} />);
    fireEvent.click(screen.getByTestId("alert-item-gap-001"));
    expect(onClick).toHaveBeenCalledOnce();
    expect(onClick).toHaveBeenCalledWith("gap-001");
  });

  it("renders custom title when provided", () => {
    render(() => <CriticalAlertsPanel spatialAlerts={[]} title="Active Gaps" />);
    expect(screen.getByText("Active Gaps")).toBeDefined();
  });

  it("shows CRITICAL severity header text", () => {
    render(() => <CriticalAlertsPanel spatialAlerts={[alerts[0]]} />);
    expect(screen.getByText(/\[CRITICAL\] SENSOR OFFLINE/)).toBeDefined();
  });
});
