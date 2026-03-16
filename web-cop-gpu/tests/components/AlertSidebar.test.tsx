// CLASSIFICATION: UNCLASSIFIED
// tests/components/AlertSidebar.test.tsx

import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@solidjs/testing-library";
import AlertSidebar from "../../src/components/panels/AlertSidebar";
import { setAlerts } from "../../src/signals/alerts";
import { setOperatorId } from "../../src/signals/auth";
import type { AlertPayload } from "../../src/workers/shared-protocol";

// Mock the alerts service to avoid real gRPC calls
const acknowledgeAlertMock = vi.fn().mockResolvedValue(undefined);

vi.mock("../../src/services/alerts", () => ({
  acknowledgeAlert: (...args: unknown[]) => acknowledgeAlertMock(...args),
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
  setOperatorId("anonymous");
  acknowledgeAlertMock.mockClear();
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

  it("has data-testid alert-sidebar on root element", () => {
    render(() => <AlertSidebar />);
    expect(screen.getByTestId("alert-sidebar")).toBeDefined();
  });

  it("calls acknowledgeAlert with operatorId from auth signal, not hardcoded string", async () => {
    setOperatorId("op-sentinel-2");
    setAlerts(mockAlerts);
    render(() => <AlertSidebar />);

    const ackBtn = screen.getByLabelText("Acknowledge alert");
    fireEvent.click(ackBtn);

    await waitFor(() => expect(acknowledgeAlertMock).toHaveBeenCalledTimes(1));
    expect(acknowledgeAlertMock).toHaveBeenCalledWith("alert-1", "op-sentinel-2");
    expect(acknowledgeAlertMock).not.toHaveBeenCalledWith("alert-1", "operator");
  });

  it("falls back to anonymous operatorId when no token has been acquired", async () => {
    // Default operatorId is "anonymous"
    setAlerts(mockAlerts);
    render(() => <AlertSidebar />);

    fireEvent.click(screen.getByLabelText("Acknowledge alert"));
    await waitFor(() => expect(acknowledgeAlertMock).toHaveBeenCalledTimes(1));

    expect(acknowledgeAlertMock).toHaveBeenCalledWith("alert-1", "anonymous");
    expect(acknowledgeAlertMock).not.toHaveBeenCalledWith("alert-1", "operator");
  });
});
