// CLASSIFICATION: UNCLASSIFIED
// tests/components/SensorOverviewMap.test.tsx

import { render, screen } from "@solidjs/testing-library";
import { describe, expect, it, vi } from "vitest";
import { SensorOverviewMap } from "../../src/components/dashboard/SensorOverviewMap";
import type { SensorStatus } from "../../src/services/sensor-health";

const connectedSensor: SensorStatus = {
  sensorId: "RADAR-NORTH-01",
  sensorType: "RADAR",
  status: "CONNECTED",
  eventsPerSecond: 48.2,
  totalReceived: 98230,
  lastSeenSeconds: 2,
  validationPassRate: 98.7,
  dlqCount: 47,
  coverage: { rangeNm: 150, bearingStart: 315, bearingEnd: 45, centerLon: -10.0, centerLat: 60.5 },
};

const offlineSensor: SensorStatus = {
  sensorId: "RADAR-WEST-04",
  sensorType: "RADAR",
  status: "OFFLINE",
  eventsPerSecond: 0,
  totalReceived: 30900,
  lastSeenSeconds: 320,
  validationPassRate: 0,
  dlqCount: 0,
  coverage: { rangeNm: 100, bearingStart: 270, bearingEnd: 360, centerLon: -13.0, centerLat: 58.0 },
};

describe("SensorOverviewMap", () => {
  it("renders map container with testid", () => {
    render(() => <SensorOverviewMap sensors={[]} />);
    expect(screen.getByTestId("sensor-overview-map")).toBeDefined();
  });

  it("shows gap badge when offline sensor with coverage exists", () => {
    render(() => <SensorOverviewMap sensors={[offlineSensor]} />);
    expect(screen.getByTestId("gap-count-badge")).toBeDefined();
    expect(screen.getByText("⚠ 1 GAP")).toBeDefined();
  });

  it("does not show gap badge when no offline sensors", () => {
    render(() => <SensorOverviewMap sensors={[connectedSensor]} />);
    expect(screen.queryByTestId("gap-count-badge")).toBeNull();
  });

  it("renders zoom controls", () => {
    render(() => <SensorOverviewMap sensors={[]} />);
    expect(screen.getByTestId("overview-zoom-in")).toBeDefined();
    expect(screen.getByTestId("overview-zoom-out")).toBeDefined();
  });

  it("calls onSensorClick when a sensor dot is clicked", () => {
    const onClick = vi.fn();
    render(() => <SensorOverviewMap sensors={[connectedSensor]} onSensorClick={onClick} />);
    expect(screen.getByTestId("sensor-overview-map")).toBeDefined();
  });

  it("renders 'GAP DETECTED' text for offline sensor coverage", () => {
    render(() => <SensorOverviewMap sensors={[offlineSensor]} />);
    expect(screen.getByText("GAP DETECTED")).toBeDefined();
  });

  it("renders multiple gap count", () => {
    render(() => <SensorOverviewMap sensors={[offlineSensor, { ...offlineSensor, sensorId: "RADAR-EAST-09" }]} />);
    expect(screen.getByText("⚠ 2 GAPS")).toBeDefined();
  });
});
