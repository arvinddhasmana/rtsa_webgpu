// CLASSIFICATION: UNCLASSIFIED
// tests/components/AlertSidebar.test.tsx

import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { AlertSidebar } from "../../src/components/panels/AlertSidebar";
import { setAlerts } from "../../src/signals/alerts";
import type { AlertPayload } from "../../src/workers/shared-protocol";

// Mock the alerts service to avoid real gRPC calls
vi.mock("../../src/services/alerts", () => ({
  acknowledgeAlert: vi.fn().mockResolvedValue(undefined),
  startAlertStream: vi.fn().mockReturnValue({ abort: vi.fn() }),
}));

const mockAlerts: AlertPayload[] = [
  {
    alertId: "alert-1",
    trackId: "track-aaa",
    severity: "CRITICAL",
    description: "Hostile track detected",
    detectedAtMs: Date.now(),
    acknowledged: false,
  },
  {
    alertId: "alert-2",
    trackId: "track-bbb",
    severity: "WATCH",
    description: "Unusual course change",
    detectedAtMs: Date.now() - 5000,
    acknowledged: true,
  },
];

afterEach(() => {
  setAlerts([]);
});

describe("AlertSidebar", () => {
  it("renders ALERTS header", () => {
    render(() => <AlertSidebar />);
    expect(screen.getByText("ALERTS")).toBeDefined();
  });

  it("shows no active alerts message when list is empty", () => {
    render(() => <AlertSidebar />);
    expect(screen.getByText("No active alerts")).toBeDefined();
  });

  it("renders alert list when alerts are present", () => {
    setAlerts(mockAlerts);
    render(() => <AlertSidebar />);
    expect(screen.getByText("Hostile track detected")).toBeDefined();
    expect(screen.getByText("Unusual course change")).toBeDefined();
  });

  it("shows alert count badge", () => {
    setAlerts(mockAlerts);
    render(() => <AlertSidebar />);
    expect(screen.getByText("2")).toBeDefined();
  });

  it("shows Ack button for unacknowledged alerts", () => {
    setAlerts(mockAlerts);
    render(() => <AlertSidebar />);
    const ackButtons = screen.getAllByLabelText("Acknowledge alert");
    expect(ackButtons.length).toBe(1); // only alert-1 is unacknowledged
  });

  it("does not show Ack button for acknowledged alerts", () => {
    setAlerts([{ ...mockAlerts[1], acknowledged: true }]);
    render(() => <AlertSidebar />);
    const ackButtons = screen.queryAllByLabelText("Acknowledge alert");
    expect(ackButtons.length).toBe(0);
  });
});
