// CLASSIFICATION: UNCLASSIFIED
// tests/components/SensorGrid.test.tsx

import { render, screen } from "@solidjs/testing-library";
import { afterEach, describe, expect, it } from "vitest";
import { SensorGrid } from "../../src/components/dashboard/SensorGrid";
import { SensorStatus } from "../../src/services/sensor-health";
import { setSelectedStatuses, setSelectedTypes } from "../../src/signals/sensor-filters";

const mockSensors: SensorStatus[] = [
  {
    sensorId: "RADAR-01",
    sensorType: "RADAR",
    status: "CONNECTED",
    eventsPerSecond: 100,
    totalReceived: 1000,
    lastSeenSeconds: 1,
    validationPassRate: 100,
    dlqCount: 0,
  },
  {
    sensorId: "AIS-02",
    sensorType: "AIS/BFT",
    status: "OFFLINE",
    eventsPerSecond: 0,
    totalReceived: 500,
    lastSeenSeconds: -1,
    validationPassRate: 0,
    dlqCount: 10,
  }
];

afterEach(() => {
  setSelectedStatuses(["CONNECTED", "STALE", "OFFLINE"]);
  setSelectedTypes(["RADAR", "EW/SIGINT", "ELINT/COMINT", "ISR", "AIS/BFT", "CYBER"]);
});

describe("SensorGrid", () => {
  it("renders all sensors by default", () => {
    render(() => <SensorGrid sensors={mockSensors} />);
    expect(screen.getByText("RADAR-01")).toBeDefined();
    expect(screen.getByText("AIS-02")).toBeDefined();
  });

  it("filters sensors by status", () => {
    setSelectedStatuses(["CONNECTED"]);
    render(() => <SensorGrid sensors={mockSensors} />);
    expect(screen.getByText("RADAR-01")).toBeDefined();
    expect(screen.queryByText("AIS-02")).toBeNull();
  });

  it("shows empty state when no sensors match", () => {
    setSelectedStatuses(["STALE"]);
    render(() => <SensorGrid sensors={mockSensors} />);
    expect(screen.getByText("No sensors match the selected filters")).toBeDefined();
  });
});
