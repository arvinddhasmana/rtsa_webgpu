// CLASSIFICATION: UNCLASSIFIED
// tests/components/SensorFleetList.test.tsx

import { fireEvent, render, screen } from "@solidjs/testing-library";
import { describe, expect, it, vi } from "vitest";
import { SensorFleetList } from "../../src/components/dashboard/SensorFleetList";
import type { SensorStatus } from "../../src/services/sensor-health";

const sensors: SensorStatus[] = [
  {
    sensorId: "RADAR-01",
    sensorType: "RADAR",
    status: "CONNECTED",
    eventsPerSecond: 48.2,
    totalReceived: 10000,
    lastSeenSeconds: 2,
    validationPassRate: 98.7,
    dlqCount: 12,
  },
  {
    sensorId: "AIS-PORT-01",
    sensorType: "AIS/BFT",
    status: "OFFLINE",
    eventsPerSecond: 0,
    totalReceived: 5000,
    lastSeenSeconds: 320,
    validationPassRate: 0,
    dlqCount: 0,
  },
];

describe("SensorFleetList", () => {
  it("renders all sensors", () => {
    render(() => <SensorFleetList sensors={sensors} />);
    expect(screen.getByTestId("fleet-row-RADAR-01")).toBeDefined();
    expect(screen.getByTestId("fleet-row-AIS-PORT-01")).toBeDefined();
  });

  it("renders sensor IDs", () => {
    render(() => <SensorFleetList sensors={sensors} />);
    expect(screen.getByText("RADAR-01")).toBeDefined();
    expect(screen.getByText("AIS-PORT-01")).toBeDefined();
  });

  it("highlights selected sensor row", () => {
    render(() => <SensorFleetList sensors={sensors} selectedSensorId="RADAR-01" />);
    const row = screen.getByTestId("fleet-row-RADAR-01");
    expect(row.getAttribute("style")).toContain("rgba(59, 130, 246, 0.12)");
  });

  it("calls onSensorSelect when row is clicked", () => {
    const onSelect = vi.fn();
    render(() => <SensorFleetList sensors={sensors} onSensorSelect={onSelect} />);
    fireEvent.click(screen.getByTestId("fleet-row-RADAR-01"));
    expect(onSelect).toHaveBeenCalledOnce();
    expect(onSelect).toHaveBeenCalledWith(sensors[0]);
  });

  it("calls onSensorHover on mouse enter and null on leave", () => {
    const onHover = vi.fn();
    render(() => <SensorFleetList sensors={sensors} onSensorHover={onHover} />);
    fireEvent.mouseEnter(screen.getByTestId("fleet-row-RADAR-01"));
    expect(onHover).toHaveBeenCalledWith(sensors[0]);
    fireEvent.mouseLeave(screen.getByTestId("fleet-row-RADAR-01"));
    expect(onHover).toHaveBeenCalledWith(null);
  });

  it("renders data-testid fleet-list container", () => {
    render(() => <SensorFleetList sensors={[]} />);
    expect(screen.getByTestId("sensor-fleet-list")).toBeDefined();
  });

  it("renders compact mode without obs rate text", () => {
    render(() => <SensorFleetList sensors={sensors} compact />);
    expect(screen.queryByText(/\/s/)).toBeNull();
  });

  it("shows obs rate in non-compact mode", () => {
    render(() => <SensorFleetList sensors={sensors} />);
    expect(screen.getByText(/48\.2 \/s/)).toBeDefined();
  });
});
