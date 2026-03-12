// CLASSIFICATION: UNCLASSIFIED
// tests/components/SensorStatusCard.test.tsx

import { render, screen } from "@solidjs/testing-library";
import { describe, expect, it } from "vitest";
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

describe("SensorStatusCard", () => {
  it("renders sensor ID and type", () => {
    render(() => <SensorStatusCard sensor={mockSensor} />);
    expect(screen.getByText("RADAR-01")).toBeDefined();
    expect(screen.getByText("RADAR")).toBeDefined();
  });

  it("renders status and metrics", () => {
    render(() => <SensorStatusCard sensor={mockSensor} />);
    expect(screen.getByText("CONNECTED")).toBeDefined();
    expect(screen.getByText(/120.5/)).toBeDefined();
    expect(screen.getByText(/99.8/)).toBeDefined();
  });

  it("renders last seen time", () => {
    render(() => <SensorStatusCard sensor={mockSensor} />);
    expect(screen.getByText("5s ago")).toBeDefined();
  });
});
