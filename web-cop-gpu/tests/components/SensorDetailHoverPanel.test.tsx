// CLASSIFICATION: UNCLASSIFIED
// tests/components/SensorDetailHoverPanel.test.tsx

import { render, screen, waitFor } from "@solidjs/testing-library";
import { describe, expect, it, vi } from "vitest";
import { SensorDetailHoverPanel } from "../../src/components/dashboard/SensorDetailHoverPanel";
import type { SensorStatus } from "../../src/services/sensor-health";

// Mock the gRPC-dependent sensor-health module
vi.mock("../../src/services/sensor-health", async () => {
  return {
    fetchSensorStatuses: vi.fn(async () => []),
    fetchSensorDiagnostic: vi.fn(async (sensor: SensorStatus) => ({
      ...sensor,
      latencyMs: 120,
      throughputHistory: Array.from({ length: 20 }, (_, i) => 40 + i),
      dlqBreakdown: [],
      recentEvents: [],
      subSensors: [],
      healthScore: 90,
      connectionUptimePct: 97.5,
      peakThroughput: 60,
      avgLatencyMs: 120,
      minLatencyMs: 80,
      maxLatencyMs: 300,
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

const sensor: SensorStatus = {
  sensorId: "RADAR-01",
  sensorType: "RADAR",
  status: "CONNECTED",
  eventsPerSecond: 48.2,
  totalReceived: 98230,
  lastSeenSeconds: 2,
  validationPassRate: 98.7,
  dlqCount: 47,
  coverage: { rangeNm: 150, bearingStart: 315, bearingEnd: 45, centerLon: -10.0, centerLat: 60.5 },
};

describe("SensorDetailHoverPanel", () => {
  it("renders nothing when sensor is null (hidden state)", () => {
    render(() => <SensorDetailHoverPanel sensor={null} />);
    expect(screen.queryByTestId("sensor-detail-hover-panel")).toBeNull();
  });

  it("renders panel when sensor is provided", () => {
    render(() => <SensorDetailHoverPanel sensor={sensor} />);
    expect(screen.getByTestId("sensor-detail-hover-panel")).toBeDefined();
  });

  it("shows sensor ID in header", () => {
    render(() => <SensorDetailHoverPanel sensor={sensor} />);
    expect(screen.getByText("RADAR-01")).toBeDefined();
  });

  it("shows sensor type badge", () => {
    render(() => <SensorDetailHoverPanel sensor={sensor} />);
    expect(screen.getByText("RADAR")).toBeDefined();
  });

  it("shows status badge", () => {
    render(() => <SensorDetailHoverPanel sensor={sensor} />);
    expect(screen.getByText("CONNECTED")).toBeDefined();
  });

  it("shows obs/s trend section label", () => {
    render(() => <SensorDetailHoverPanel sensor={sensor} />);
    expect(screen.getByText("Obs/s Trend")).toBeDefined();
  });

  it("shows connection uptime section after data loads", async () => {
    render(() => <SensorDetailHoverPanel sensor={sensor} />);
    await waitFor(() => expect(screen.getByText("Connection Uptime")).toBeDefined(), { timeout: 2000 });
  });

  it("shows coverage zone section when sensor has coverage", () => {
    render(() => <SensorDetailHoverPanel sensor={sensor} />);
    expect(screen.getByText("Coverage Zone")).toBeDefined();
  });

  it("does not show coverage zone when sensor has no coverage", () => {
    const sensorNoCoverage = { ...sensor, coverage: undefined };
    render(() => <SensorDetailHoverPanel sensor={sensorNoCoverage} />);
    expect(screen.queryByText("Coverage Zone")).toBeNull();
  });
});
