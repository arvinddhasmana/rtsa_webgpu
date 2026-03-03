// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/SensorHealthDashboard.test.tsx

import { act, fireEvent, render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

// ── Static mocks ────────────────────────────────────────
vi.mock("../../hooks/useSensorHealth", () => ({
  useSensorHealth: vi.fn(),
}));

vi.mock("../../components/map/MapView", () => ({
  MapView: () => <div data-testid="mock-map">Map</div>,
}));

// Pre-populate the store with test sensors BEFORE importing the component
import { useSensorHealthStore } from "../../stores/sensorHealthStore";
import type { SensorStatus } from "../../types/sensor";

const makeSensor = (overrides: Partial<SensorStatus> = {}): SensorStatus => ({
  sensorId: "RADAR-01",
  sensorType: "RADAR",
  connected: true,
  connectionStatus: "connected",
  totalReceived: 1000,
  totalAccepted: 950,
  totalRejected: 50,
  lastObservationTime: new Date(),
  eventsPerSecond: 5.0,
  acceptanceRate: 95.0,
  rateHistory: [3, 4, 5, 5, 5],
  latencyMs: 100,
  ...overrides,
});

const sensor1 = makeSensor({ sensorId: "RADAR-01", connectionStatus: "connected" });
const sensor2 = makeSensor({
  sensorId: "EW-STATION-01",
  sensorType: "EW",
  connectionStatus: "degraded",
  totalRejected: 5,
});
const sensor3 = makeSensor({
  sensorId: "AIS-COAST-01",
  sensorType: "AIS",
  connectionStatus: "disconnected",
  connected: false,
  totalRejected: 0,
});

import { SensorHealthDashboard } from "../../components/layout/SensorHealthDashboard";

describe("SensorHealthDashboard", () => {
  beforeEach(() => {
    useSensorHealthStore.getState().clearAll();
    useSensorHealthStore.setState({ isLoading: false });
    useSensorHealthStore.getState().upsertSensors([sensor1, sensor2, sensor3]);
  });

  it("renders the dashboard container", () => {
    render(<SensorHealthDashboard />);
    expect(screen.getByTestId("sensor-health-dashboard")).toBeInTheDocument();
  });

  it("renders the Sensor Grid tab by default", () => {
    render(<SensorHealthDashboard />);
    expect(screen.getByTestId("tab-sensor-grid")).toBeInTheDocument();
    expect(screen.getByTestId("sensor-table")).toBeInTheDocument();
  });

  it("renders KPI tiles", () => {
    render(<SensorHealthDashboard />);
    expect(screen.getByTestId("kpi-active")).toBeInTheDocument();
    expect(screen.getByTestId("kpi-degraded")).toBeInTheDocument();
    expect(screen.getByTestId("kpi-throughput")).toBeInTheDocument();
    expect(screen.getByTestId("kpi-latency")).toBeInTheDocument();
  });

  it("shows correct active count in KPI tile", () => {
    render(<SensorHealthDashboard />);
    // sensor1 and sensor2 both have connected=true and recent lastObservationTime
    // → both are re-derived as 'connected' by the store. sensor3 (connected=false) → 'disconnected'
    const activeTile = screen.getByTestId("kpi-active");
    expect(within(activeTile).getByText("2")).toBeInTheDocument();
  });

  it("shows degraded+offline count in KPI tile", () => {
    render(<SensorHealthDashboard />);
    const degradedTile = screen.getByTestId("kpi-degraded");
    // sensor3 has connected=false → disconnected. sensor2 has connected=true → re-derived 'connected'
    // So degraded+offline = 1 (only sensor3)
    expect(within(degradedTile).getByText("1")).toBeInTheDocument();
  });

  it("renders a table row for each sensor", () => {
    render(<SensorHealthDashboard />);
    expect(screen.getByTestId("sensor-row-RADAR-01")).toBeInTheDocument();
    expect(screen.getByTestId("sensor-row-EW-STATION-01")).toBeInTheDocument();
    expect(screen.getByTestId("sensor-row-AIS-COAST-01")).toBeInTheDocument();
  });

  it("renders the map", () => {
    render(<SensorHealthDashboard />);
    expect(screen.getByTestId("mock-map")).toBeInTheDocument();
  });

  it("renders the resize handle", () => {
    render(<SensorHealthDashboard />);
    expect(screen.getByTestId("resize-handle")).toBeInTheDocument();
  });

  it("switches to DLQ tab when clicked", () => {
    render(<SensorHealthDashboard />);
    fireEvent.click(screen.getByTestId("tab-dlq"));
    expect(screen.getByTestId("dlq-viewer")).toBeInTheDocument();
  });

  it("switches back to Sensor Grid tab", () => {
    render(<SensorHealthDashboard />);
    fireEvent.click(screen.getByTestId("tab-dlq"));
    fireEvent.click(screen.getByTestId("tab-sensor-grid"));
    expect(screen.getByTestId("sensor-table")).toBeInTheDocument();
  });

  it("expands inline detail when clicking a sensor row", () => {
    render(<SensorHealthDashboard />);
    fireEvent.click(screen.getByTestId("sensor-row-RADAR-01"));
    expect(screen.getByTestId("sensor-detail-RADAR-01")).toBeInTheDocument();
  });

  it("collapses inline detail when clicking the same row again", () => {
    render(<SensorHealthDashboard />);
    fireEvent.click(screen.getByTestId("sensor-row-RADAR-01"));
    expect(screen.getByTestId("sensor-detail-RADAR-01")).toBeInTheDocument();
    fireEvent.click(screen.getByTestId("sensor-row-RADAR-01"));
    expect(screen.queryByTestId("sensor-detail-RADAR-01")).not.toBeInTheDocument();
  });

  it("shows DLQ icon only for sensors with rejections", () => {
    render(<SensorHealthDashboard />);
    // EW-STATION-01 has 5 rejections → should have DLQ icon
    expect(screen.getByTestId("dlq-icon-EW-STATION-01")).toBeInTheDocument();
    // AIS-COAST-01 has 0 rejections → no DLQ icon
    expect(screen.queryByTestId("dlq-icon-AIS-COAST-01")).not.toBeInTheDocument();
  });

  it("opens DLQ popup when clicking DLQ icon", () => {
    render(<SensorHealthDashboard />);
    fireEvent.click(screen.getByTestId("dlq-icon-EW-STATION-01"));
    expect(screen.getByTestId("dlq-popup")).toBeInTheDocument();
  });

  it("resets layout and selection on Escape key", async () => {
    render(<SensorHealthDashboard />);
    // First select a sensor
    fireEvent.click(screen.getByTestId("sensor-row-RADAR-01"));
    expect(screen.getByTestId("sensor-detail-RADAR-01")).toBeInTheDocument();
    // Press Escape
    await act(async () => {
      fireEvent.keyDown(window, { key: "Escape" });
    });
    expect(screen.queryByTestId("sensor-detail-RADAR-01")).not.toBeInTheDocument();
  });

  it("shows loading state when isLoading=true", () => {
    useSensorHealthStore.setState({ isLoading: true });
    render(<SensorHealthDashboard />);
    expect(screen.getByText(/Awaiting sensor telemetry/i)).toBeInTheDocument();
  });
});

describe("SensorHealthDashboard — table sorting", () => {
  beforeEach(() => {
    useSensorHealthStore.getState().clearAll();
    useSensorHealthStore.setState({ isLoading: false });
    useSensorHealthStore.getState().upsertSensors([
      makeSensor({ sensorId: "Z-SENSOR", eventsPerSecond: 1 }),
      makeSensor({ sensorId: "A-SENSOR", eventsPerSecond: 100 }),
    ]);
  });

  it("sorts ascending by sensorId by default", () => {
    render(<SensorHealthDashboard />);
    const rows = screen.getAllByTestId(/sensor-row-/);
    expect(rows[0].getAttribute("data-testid")).toBe("sensor-row-A-SENSOR");
    expect(rows[1].getAttribute("data-testid")).toBe("sensor-row-Z-SENSOR");
  });

  it("sorts descending after clicking sensorId header twice", () => {
    render(<SensorHealthDashboard />);
    const sensorIdHeader = screen.getByRole("columnheader", { name: /Sensor ID/i });
    // Default is already asc, one click → desc
    fireEvent.click(sensorIdHeader);
    const rows = screen.getAllByTestId(/sensor-row-/);
    expect(rows[0].getAttribute("data-testid")).toBe("sensor-row-Z-SENSOR");
  });
});
