// CLASSIFICATION: UNCLASSIFIED
// tests/components/SensorDiagnosticView.test.tsx

import { fireEvent, render, screen, waitFor } from "@solidjs/testing-library";
import { describe, expect, it, vi } from "vitest";
import { SensorDiagnosticView } from "../../src/components/dashboard/SensorDiagnosticView";
import type { SensorStatus } from "../../src/services/sensor-health";

// Mock the sensor-filters module so we can spy on setSelectedSensor
vi.mock("../../src/signals/sensor-filters", async () => {
  const { createSignal } = await import("solid-js");
  const [selectedSensor, setSelectedSensor] = createSignal(null);
  return {
    selectedSensor,
    setSelectedSensor: vi.fn(setSelectedSensor),
    selectedStatuses: () => ["CONNECTED", "STALE", "OFFLINE"],
    setSelectedStatuses: vi.fn(),
    selectedTypes: () => [
      "RADAR",
      "EW/SIGINT",
      "ELINT/COMINT",
      "ISR",
      "AIS/BFT",
      "CYBER",
    ],
    setSelectedTypes: vi.fn(),
    sidebarCollapsed: () => false,
    setSidebarCollapsed: vi.fn(),
    toggleStatusFilter: vi.fn(),
    toggleTypeFilter: vi.fn(),
  };
});

const mockSensor: SensorStatus = {
  sensorId: "RADAR-DIAG-01",
  sensorType: "RADAR",
  status: "CONNECTED",
  eventsPerSecond: 85.4,
  totalReceived: 8000,
  lastSeenSeconds: 2,
  validationPassRate: 97.1,
  dlqCount: 35,
};

const LOAD_TIMEOUT = 3000;

describe("SensorDiagnosticView", () => {
  it("renders sensor ID in the header breadcrumb", async () => {
    render(() => <SensorDiagnosticView sensor={mockSensor} />);
    await waitFor(() =>
      expect(screen.getByText("RADAR-DIAG-01")).toBeDefined(),
    );
  });

  it("renders sensor type badge", async () => {
    render(() => <SensorDiagnosticView sensor={mockSensor} />);
    const badges = screen.getAllByText("RADAR");
    expect(badges.length).toBeGreaterThan(0);
  });

  it("has data-testid='sensor-diagnostic-view' on the outer container", () => {
    render(() => <SensorDiagnosticView sensor={mockSensor} />);
    expect(screen.getByTestId("sensor-diagnostic-view")).toBeDefined();
  });

  it("has back button with data-testid='diagnostic-back-btn'", () => {
    render(() => <SensorDiagnosticView sensor={mockSensor} />);
    expect(screen.getByTestId("diagnostic-back-btn")).toBeDefined();
  });

  it("calls setSelectedSensor(null) when back button is clicked", async () => {
    const { setSelectedSensor } =
      await import("../../src/signals/sensor-filters");
    render(() => <SensorDiagnosticView sensor={mockSensor} />);
    const btn = screen.getByTestId("diagnostic-back-btn");
    fireEvent.click(btn);
    expect(setSelectedSensor).toHaveBeenCalledWith(null);
  });

  it("shows DLQ Breakdown section heading after data loads", async () => {
    render(() => <SensorDiagnosticView sensor={mockSensor} />);
    await waitFor(
      () => expect(screen.getByText(/DLQ Breakdown/i)).toBeDefined(),
      { timeout: LOAD_TIMEOUT },
    );
  });

  it("shows Recent Events section heading after data loads", async () => {
    render(() => <SensorDiagnosticView sensor={mockSensor} />);
    await waitFor(
      () => expect(screen.getByText(/Recent Events/i)).toBeDefined(),
      { timeout: LOAD_TIMEOUT },
    );
  });

  it("shows Overall Health Score section after data loads", async () => {
    render(() => <SensorDiagnosticView sensor={mockSensor} />);
    await waitFor(
      () => expect(screen.getByText(/Overall Health Score/i)).toBeDefined(),
      { timeout: LOAD_TIMEOUT },
    );
  });

  it("shows Sensor Parameters section after data loads", async () => {
    render(() => <SensorDiagnosticView sensor={mockSensor} />);
    await waitFor(
      () => expect(screen.getByText(/Sensor Parameters/i)).toBeDefined(),
      { timeout: LOAD_TIMEOUT },
    );
  });

  it("shows Status History section after data loads", async () => {
    render(() => <SensorDiagnosticView sensor={mockSensor} />);
    await waitFor(
      () => expect(screen.getByText(/Status History/i)).toBeDefined(),
      { timeout: LOAD_TIMEOUT },
    );
  });

  it("shows Latency Distribution section after data loads", async () => {
    render(() => <SensorDiagnosticView sensor={mockSensor} />);
    await waitFor(
      () => expect(screen.getByText(/Latency Distribution/i)).toBeDefined(),
      { timeout: LOAD_TIMEOUT },
    );
  });

  it("shows Throughput History section after data loads", async () => {
    render(() => <SensorDiagnosticView sensor={mockSensor} />);
    await waitFor(
      () => expect(screen.getByText(/Throughput History/i)).toBeDefined(),
      { timeout: LOAD_TIMEOUT },
    );
  });

  it("shows Connection Uptime section after data loads", async () => {
    render(() => <SensorDiagnosticView sensor={mockSensor} />);
    await waitFor(
      () => expect(screen.getByText(/Connection Uptime/i)).toBeDefined(),
      { timeout: LOAD_TIMEOUT },
    );
  });

  it("shows quick stats in header (obs/s and DLQ)", async () => {
    render(() => <SensorDiagnosticView sensor={mockSensor} />);
    await waitFor(() => {
      expect(screen.getByText(/obs\/s/i)).toBeDefined();
      expect(screen.getByText(/DLQ/i)).toBeDefined();
    });
  });
});
