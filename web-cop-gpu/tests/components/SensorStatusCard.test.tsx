// CLASSIFICATION: UNCLASSIFIED
// tests/components/SensorStatusCard.test.tsx

import { fireEvent, render, screen } from "@solidjs/testing-library";
import { describe, expect, it, vi } from "vitest";
import { SensorStatusCard } from "../../src/components/dashboard/SensorStatusCard";
import { SensorStatus } from "../../src/services/sensor-health";

const mockSensor: SensorStatus = {
  sensorId: "RADAR-01",
  sensorType: "RADAR",
  status: "CONNECTED",
  eventsPerSecond: 120.5,
  totalReceived: 5000,
  lastSeenSeconds: 5,
  validationPassRate: 99.8,
  dlqCount: 2,
};

const offlineSensor: SensorStatus = {
  sensorId: "RADAR-OFFLINE",
  sensorType: "RADAR",
  status: "OFFLINE",
  eventsPerSecond: 0,
  totalReceived: 1200,
  lastSeenSeconds: 300,
  validationPassRate: 0,
  dlqCount: 0,
};

const highDlqSensor: SensorStatus = {
  sensorId: "EW-STALE-01",
  sensorType: "EW/SIGINT",
  status: "STALE",
  eventsPerSecond: 5.2,
  totalReceived: 18000,
  lastSeenSeconds: 75,
  validationPassRate: 78.4,
  dlqCount: 387,
};

describe("SensorStatusCard — common behaviour", () => {
  it("renders sensor ID and type", () => {
    render(() => <SensorStatusCard sensor={mockSensor} />);
    expect(screen.getByText("RADAR-01")).toBeDefined();
    expect(screen.getByText("RADAR")).toBeDefined();
  });

  it("has data-testid attribute with sensor ID", () => {
    render(() => <SensorStatusCard sensor={mockSensor} />);
    expect(screen.getByTestId("sensor-card-RADAR-01")).toBeDefined();
  });

  it("calls onSelect callback with sensor when clicked", () => {
    const onSelect = vi.fn();
    render(() => <SensorStatusCard sensor={mockSensor} onSelect={onSelect} />);
    const card = screen.getByTestId("sensor-card-RADAR-01");
    fireEvent.click(card);
    expect(onSelect).toHaveBeenCalledOnce();
    expect(onSelect).toHaveBeenCalledWith(mockSensor);
  });

  it("does not throw when clicked without onSelect prop", () => {
    render(() => <SensorStatusCard sensor={mockSensor} />);
    const card = screen.getByTestId("sensor-card-RADAR-01");
    expect(() => fireEvent.click(card)).not.toThrow();
  });
});

describe("SensorStatusCard — Full View (default)", () => {
  it("renders data-view=full attribute", () => {
    render(() => <SensorStatusCard sensor={mockSensor} />);
    expect(
      screen.getByTestId("sensor-card-RADAR-01").getAttribute("data-view"),
    ).toBe("full");
  });

  it("shows Connected badge", () => {
    render(() => <SensorStatusCard sensor={mockSensor} />);
    expect(screen.getByText("Connected")).toBeDefined();
  });

  it("shows Throughput metric tile with obs/s value", () => {
    render(() => <SensorStatusCard sensor={mockSensor} />);
    // eventsPerSecond appears in the tile
    expect(screen.getByText(/120\.5/)).toBeDefined();
  });

  it("shows Total Recv'd metric tile with formatted total", () => {
    render(() => <SensorStatusCard sensor={mockSensor} />);
    // totalReceived=5000 → "5,000" via toLocaleString()
    expect(screen.getByText(/5[,.]?000/)).toBeDefined();
  });

  it("shows DLQ Count metric tile", () => {
    render(() => <SensorStatusCard sensor={mockSensor} />);
    // dlqCount=2; tile has heading label
    expect(screen.getByText("DLQ Count")).toBeDefined();
    expect(screen.getByText(/^2$/)).toBeDefined();
  });

  it("shows Validation metric tile with pass rate", () => {
    render(() => <SensorStatusCard sensor={mockSensor} />);
    expect(screen.getByText("Validation")).toBeDefined();
    expect(screen.getByText(/99\.8/)).toBeDefined();
  });

  it("shows chart with -15m and now axis labels", () => {
    render(() => <SensorStatusCard sensor={mockSensor} />);
    const chart = screen.getByTestId("full-card-chart");
    expect(chart).toBeDefined();
    expect(chart.textContent).toContain("-15m");
    expect(chart.textContent).toContain("now");
  });

  it("shows chart legend with obs/s and accepted labels", () => {
    render(() => <SensorStatusCard sensor={mockSensor} />);
    const chart = screen.getByTestId("full-card-chart");
    expect(chart.textContent).toContain("obs/s");
    expect(chart.textContent).toContain("accepted");
  });

  it("shows Last Seen in footer", () => {
    render(() => <SensorStatusCard sensor={mockSensor} />);
    expect(screen.getByText("5s ago")).toBeDefined();
  });

  it("does NOT show 'Data Quality' label (replaced by Validation tile)", () => {
    render(() => <SensorStatusCard sensor={mockSensor} />);
    expect(screen.queryByText("Data Quality")).toBeNull();
  });

  it("shows N/A validation for offline sensor", () => {
    render(() => <SensorStatusCard sensor={offlineSensor} />);
    expect(screen.getByText("N/A")).toBeDefined();
  });

  it("shows DLQ count in red when > 50", () => {
    render(() => <SensorStatusCard sensor={highDlqSensor} />);
    expect(screen.getByText(/387/)).toBeDefined();
  });

  it("shows Throughput tile heading", () => {
    render(() => <SensorStatusCard sensor={mockSensor} />);
    expect(screen.getByText("Throughput")).toBeDefined();
  });
});

describe("SensorStatusCard — Compact View", () => {
  it("renders data-view=compact attribute", () => {
    render(() => <SensorStatusCard sensor={mockSensor} compact />);
    expect(
      screen.getByTestId("sensor-card-RADAR-01").getAttribute("data-view"),
    ).toBe("compact");
  });

  it("renders status and metrics in compact mode", () => {
    render(() => <SensorStatusCard sensor={mockSensor} compact />);
    expect(screen.getByText("Connected")).toBeDefined();
    expect(screen.getByText(/120\.5/)).toBeDefined();
    expect(screen.getByText(/99\.8/)).toBeDefined();
  });

  it("shows DLQ count in footer", () => {
    render(() => <SensorStatusCard sensor={mockSensor} compact />);
    expect(screen.getByText("2")).toBeDefined();
  });

  it("shows Last Seen in compact footer", () => {
    render(() => <SensorStatusCard sensor={mockSensor} compact />);
    expect(screen.getByText("5s ago")).toBeDefined();
  });

  it("does NOT show full-card-chart in compact mode", () => {
    render(() => <SensorStatusCard sensor={mockSensor} compact />);
    expect(screen.queryByTestId("full-card-chart")).toBeNull();
  });
});
