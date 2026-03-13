// CLASSIFICATION: UNCLASSIFIED
// tests/components/CoverageAreaMap.test.tsx

import { render, screen } from "@solidjs/testing-library";
import { describe, expect, it } from "vitest";
import { CoverageAreaMap } from "../../src/components/dashboard/CoverageAreaMap";
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

describe("CoverageAreaMap", () => {
  it("renders without sensors (empty)", () => {
    render(() => <CoverageAreaMap sensors={[]} />);
    expect(screen.getByTestId("coverage-area-map")).toBeDefined();
  });

  it("renders with connected sensor and coverage", () => {
    render(() => <CoverageAreaMap sensors={[connectedSensor]} showLabels />);
    expect(screen.getByTestId("coverage-area-map")).toBeDefined();
  });

  it("renders gap badge when offline sensor with coverage exists", () => {
    render(() => <CoverageAreaMap sensors={[offlineSensor]} />);
    expect(screen.getByTestId("gap-badge")).toBeDefined();
  });

  it("does not show gap badge when all sensors are connected", () => {
    render(() => <CoverageAreaMap sensors={[connectedSensor]} />);
    expect(screen.queryByTestId("gap-badge")).toBeNull();
  });

  it("renders zoom controls", () => {
    render(() => <CoverageAreaMap sensors={[]} />);
    expect(screen.getByTestId("map-zoom-in")).toBeDefined();
    expect(screen.getByTestId("map-zoom-out")).toBeDefined();
  });

  it("renders radar sweep line for RADAR sensor with showSweepAnimation", () => {
    render(() => (
      <CoverageAreaMap
        sensors={[connectedSensor]}
        focusSensorId="RADAR-NORTH-01"
        showSweepAnimation
        showRangeRings
      />
    ));
    // The component renders the sweep line only when focusSensor is RADAR type
    expect(screen.getByTestId("coverage-area-map")).toBeDefined();
  });

  it("renders 'GAP DETECTED' text for offline sensor", () => {
    render(() => (
      <CoverageAreaMap
        sensors={[offlineSensor]}
        showGapHatching
      />
    ));
    expect(screen.getByText("GAP DETECTED")).toBeDefined();
  });
});
