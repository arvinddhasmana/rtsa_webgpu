// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/AlertPanel.test.tsx

import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { AlertPanel } from "../../components/alerts/AlertPanel";
import { useAlertStore } from "../../stores/alertStore";
import { AnomalyAlert } from "../../types/alert";

const makeAlert = (overrides: Partial<AnomalyAlert> = {}): AnomalyAlert => ({
  alertId: "ALT-001",
  trackId: "TRK-001",
  anomalyType: "SPEED",
  severity: "WATCH",
  confidenceScore: 0.8,
  explanation: "Test",
  features: [],
  classification: "UNCLASSIFIED",
  // Make the date recent so it doesn't get filtered out by default 24h timeRangeFilter
  detectedAt: new Date(Date.now() - 1000 * 60 * 60),
  ...overrides,
});

describe("AlertPanel", () => {
  beforeEach(() => {
    useAlertStore.getState().clearAll();
    useAlertStore.getState().setMinSeverityFilter("WATCH");
  });

  it("renders the ALERTS header", () => {
    // There's no longer just an "ALERTS" text header, but rather tab buttons.
    render(<AlertPanel />);
    expect(screen.getByRole('button', { name: /triage queue/i })).toBeTruthy();
  });

  it("shows 'Queue is clear...' when store is empty in pending tab", () => {
    render(<AlertPanel />);
    expect(screen.getByText("Queue is clear. No unacknowledged alerts match filters.")).toBeTruthy();
  });

  it("shows unacknowledged count badge when alerts exist", () => {
    useAlertStore.getState().addAlert(makeAlert({ alertId: "A1" }));
    useAlertStore.getState().addAlert(makeAlert({ alertId: "A2" }));
    render(<AlertPanel />);
    expect(screen.getByTestId("unacknowledged-count")).toBeTruthy();
    expect(screen.getByTestId("unacknowledged-count").textContent).toBe("2");
  });

  it("hides badge when all alerts acknowledged (in header)", () => {
    useAlertStore.getState().addAlert(makeAlert({ alertId: "A1" }));
    useAlertStore.getState().acknowledgeAlert("A1");
    render(<AlertPanel />);
    // Testing unacknowledged-count badge removal
    expect(screen.queryByTestId("unacknowledged-count")).toBeNull();
  });

  it("renders alert cards for each filtered alert in active tab", () => {
    // Add one unacknowledged, one acknowledged
    useAlertStore.getState().addAlert(makeAlert({ alertId: "A1", severity: "WATCH" }));
    useAlertStore.getState().addAlert(makeAlert({ alertId: "A2", severity: "CRITICAL" }));

    // Default tab is QUEUE, shows unacknowledged
    render(<AlertPanel />);
    expect(screen.getByTestId("alert-card-A1")).toBeTruthy();
    expect(screen.getByTestId("alert-card-A2")).toBeTruthy();

    // Acknowledge one, should disappear from QUEUE and appear in HISTORY
    fireEvent.click(screen.getByTestId("alert-card-A1")); // Acknowledges A1

    expect(screen.queryByTestId("alert-card-A1")).toBeNull();
    expect(screen.getByTestId("alert-card-A2")).toBeTruthy();

    // Switch to HISTORY tab
    fireEvent.click(screen.getByRole('button', { name: /history/i }));

    expect(screen.getByTestId("alert-card-A1")).toBeTruthy();
    // A2 isn't acknowledged yet.
    expect(screen.queryByTestId("alert-card-A2")).toBeNull();
  });

  it("renders AlertFilter component", () => {
    render(<AlertPanel />);
    expect(screen.getByTestId("alert-filter")).toBeTruthy();
  });

  it("renders at most 120 alerts to keep UI responsive", () => {
    const now = Date.now();
    for (let i = 0; i < 130; i += 1) {
      useAlertStore.getState().addAlert(
        makeAlert({
          alertId: `A-${i}`,
          trackId: `TRK-${i}`,
          severity: i % 2 === 0 ? "CRITICAL" : "WATCH",
          // Recent date so it's not filtered out
          detectedAt: new Date(now - i * 1000),
        }),
      );
    }

    render(<AlertPanel />);

    const renderedCards = document.querySelectorAll(
      '[data-testid^="alert-card-"]',
    );
    expect(renderedCards.length).toBe(120);
  });
});

describe("AlertFilter", () => {
  beforeEach(() => {
    useAlertStore.getState().clearAll();
    useAlertStore.getState().setMinSeverityFilter("WATCH");
  });

  it("renders severity filter buttons", () => {
    render(<AlertPanel />);
    expect(screen.getByTestId("filter-watch")).toBeTruthy();
    expect(screen.getByTestId("filter-elevated")).toBeTruthy();
    expect(screen.getByTestId("filter-critical")).toBeTruthy();
  });

  it("clicking CRITICAL filter updates store", () => {
    render(<AlertPanel />);
    fireEvent.click(screen.getByTestId("filter-critical"));
    expect(useAlertStore.getState().minSeverityFilter).toBe("CRITICAL");
  });
});
