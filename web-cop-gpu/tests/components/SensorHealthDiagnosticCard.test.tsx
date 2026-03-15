// CLASSIFICATION: UNCLASSIFIED
// tests/components/SensorHealthDiagnosticCard.test.tsx

import { fireEvent, render, screen, waitFor } from "@solidjs/testing-library";
import { describe, expect, it, vi } from "vitest";
import { SensorHealthDiagnosticCard } from "../../src/components/dashboard/SensorHealthDiagnosticCard";
import type { SensorStatus } from "../../src/services/sensor-health";

vi.mock("../../src/services/sensor-health", async () => {
  return {
    fetchSensorStatuses: vi.fn(async () => []),
    fetchSensorDiagnostic: vi.fn(async (sensor: SensorStatus) => ({
      ...sensor,
      latencyMs: 85,
      throughputHistory: Array.from({ length: 20 }, () => 48),
      dlqBreakdown: [],
      recentEvents: [],
      subSensors: [],
      healthScore: 88,
      connectionUptimePct: 97.5,
      peakThroughput: 60,
      avgLatencyMs: 85,
      minLatencyMs: 45,
      maxLatencyMs: 200,
      statusHistory: [],
      rangeNm: 150,
      position: { lat: 60.5, lon: -10.0 },
      bearingStart: 315,
      bearingEnd: 45,
      scanRateRpm: 6,
      frequencyBandGhz: 9.4,
      dlqReasons: [],
      connectivityEvents: [],
      uptimePercent: 97.5,
      obsPerSecHistory: Array.from({ length: 60 }, () => 48),
    })),
    sensorTypeLabel: vi.fn((t: unknown) => String(t)),
  };
});

vi.mock("../../src/signals/sensor-filters", async () => {
  const { createSignal } = await import("solid-js");
  const [selectedSensor, setSelectedSensor] = createSignal(null);
  return {
    selectedSensor,
    setSelectedSensor: vi.fn(setSelectedSensor),
    selectedStatuses: () => ["CONNECTED", "STALE", "OFFLINE"],
    setSelectedStatuses: vi.fn(),
    selectedTypes: () => ["RADAR", "EW/SIGINT", "ELINT/COMINT", "ISR", "AIS/BFT", "CYBER"],
    setSelectedTypes: vi.fn(),
    sidebarCollapsed: () => false,
    setSidebarCollapsed: vi.fn(),
    toggleStatusFilter: vi.fn(),
    toggleTypeFilter: vi.fn(),
    cardView: () => "full",
    setCardView: vi.fn(),
  };
});

const sensor: SensorStatus = {
  sensorId: "RADAR-DIAG-CARD-01",
  sensorType: "RADAR",
  status: "CONNECTED",
  eventsPerSecond: 48.2,
  totalReceived: 98230,
  lastSeenSeconds: 2,
  validationPassRate: 98.7,
  dlqCount: 5,
};

describe("SensorHealthDiagnosticCard", () => {
  it("renders with data-testid='sensor-health-diagnostic-card'", () => {
    render(() => <SensorHealthDiagnosticCard sensor={sensor} />);
    expect(screen.getByTestId("sensor-health-diagnostic-card")).toBeDefined();
  });

  it("shows sensor ID in header", () => {
    render(() => <SensorHealthDiagnosticCard sensor={sensor} />);
    expect(screen.getByText("RADAR-DIAG-CARD-01")).toBeDefined();
  });

  it("shows sensor type", () => {
    render(() => <SensorHealthDiagnosticCard sensor={sensor} />);
    expect(screen.getByText("RADAR")).toBeDefined();
  });

  it("shows sensor status and health widgets", async () => {
    render(() => <SensorHealthDiagnosticCard sensor={sensor} />);
    expect(screen.getByText("CONNECTED")).toBeDefined();
    await waitFor(() => expect(screen.getByText("NOMINAL")).toBeDefined(), { timeout: 2000 });
  });

  it("shows throughput metric", async () => {
    render(() => <SensorHealthDiagnosticCard sensor={sensor} />);
    expect(await screen.findByText("Throughput", {}, { timeout: 2000 })).toBeDefined();
    await waitFor(() => expect(screen.getByText(/48\.2/)).toBeDefined(), { timeout: 2000 });
  });

  it("shows DLQ count metric", async () => {
    render(() => <SensorHealthDiagnosticCard sensor={sensor} />);
    expect(await screen.findByText("DLQ Count", {}, { timeout: 2000 })).toBeDefined();
    await waitFor(() => expect(screen.getByText("5")).toBeDefined(), { timeout: 2000 });
  });

  it("shows health score after data loads", async () => {
    render(() => <SensorHealthDiagnosticCard sensor={sensor} />);
    await waitFor(() => expect(screen.getByText("88")).toBeDefined(), { timeout: 2000 });
    await waitFor(() => expect(screen.getByText("NOMINAL")).toBeDefined(), { timeout: 2000 });
  });

  it("has close button that calls onClose", () => {
    const onClose = vi.fn();
    render(() => <SensorHealthDiagnosticCard sensor={sensor} onClose={onClose} />);
    const closeBtn = screen.getByTestId("diagnostic-card-close-btn");
    fireEvent.click(closeBtn);
    expect(onClose).toHaveBeenCalledOnce();
  });

  it("has View Full Diagnostics button", async () => {
    render(() => <SensorHealthDiagnosticCard sensor={sensor} />);
    await waitFor(() => expect(screen.getByTestId("diagnostic-card-view-full-btn")).toBeDefined(), { timeout: 2000 });
  });

  it("clicking View Full Diagnostics calls setSelectedSensor", async () => {
    const { setSelectedSensor } = await import("../../src/signals/sensor-filters");
    render(() => <SensorHealthDiagnosticCard sensor={sensor} />);
    const viewBtn = await screen.findByTestId("diagnostic-card-view-full-btn", {}, { timeout: 2000 });
    fireEvent.click(viewBtn);
    await waitFor(() => expect(setSelectedSensor).toHaveBeenCalledWith(sensor), { timeout: 2000 });
  });

  it("shows DEGRADED label for health score 70", async () => {
    // Override the diagnostic mock for this test
    const { fetchSensorDiagnostic } = await import("../../src/services/sensor-health");
    (fetchSensorDiagnostic as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ...sensor,
      healthScore: 70,
      connectionUptimePct: 80,
      avgLatencyMs: 200,
      throughputHistory: [],
      dlqBreakdown: [],
      recentEvents: [],
      subSensors: [],
      peakThroughput: 0,
      minLatencyMs: 100,
      maxLatencyMs: 400,
      statusHistory: [],
      rangeNm: null,
      position: null,
      bearingStart: null,
      bearingEnd: null,
      scanRateRpm: null,
      frequencyBandGhz: null,
      dlqReasons: [],
      connectivityEvents: [],
      uptimePercent: 80,
      obsPerSecHistory: [],
      latencyMs: 200,
      dlqCount: 35,
    });
    render(() => <SensorHealthDiagnosticCard sensor={sensor} />);
    await waitFor(() => expect(screen.getByText("DEGRADED")).toBeDefined(), { timeout: 2000 });
  });

  it("renders OPS and timeline sections", async () => {
    render(() => <SensorHealthDiagnosticCard sensor={sensor} />);
    await waitFor(() => expect(screen.getByText(/Observation Rate/i)).toBeDefined(), { timeout: 2000 });
    expect(screen.getByText(/Diagnostic Event Timeline/i)).toBeDefined();
  });
});
