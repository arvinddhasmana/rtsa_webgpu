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
  detectedAt: new Date("2026-01-01T12:00:00Z"),
  ...overrides,
});

describe("AlertPanel", () => {
  beforeEach(() => {
    useAlertStore.getState().clearAll();
    useAlertStore.getState().setMinSeverityFilter("WATCH");
  });

  it("renders the ALERTS header", () => {
    render(<AlertPanel />);
    expect(screen.getByText("ALERTS")).toBeTruthy();
  });

  it("shows 'No alerts' when store is empty", () => {
    render(<AlertPanel />);
    expect(screen.getByText("No alerts")).toBeTruthy();
  });

  it("shows unacknowledged count badge when alerts exist", () => {
    useAlertStore.getState().addAlert(makeAlert({ alertId: "A1" }));
    useAlertStore.getState().addAlert(makeAlert({ alertId: "A2" }));
    render(<AlertPanel />);
    expect(screen.getByTestId("unacknowledged-count")).toBeTruthy();
    expect(screen.getByTestId("unacknowledged-count").textContent).toBe("2");
  });

  it("hides badge when all alerts acknowledged", () => {
    useAlertStore.getState().addAlert(makeAlert({ alertId: "A1" }));
    useAlertStore.getState().acknowledgeAlert("A1");
    render(<AlertPanel />);
    expect(screen.queryByTestId("unacknowledged-count")).toBeNull();
  });

  it("renders alert cards for each filtered alert", () => {
    useAlertStore
      .getState()
      .addAlert(makeAlert({ alertId: "A1", severity: "WATCH" }));
    useAlertStore
      .getState()
      .addAlert(makeAlert({ alertId: "A2", severity: "CRITICAL" }));
    render(<AlertPanel />);
    expect(screen.getByTestId("alert-card-A1")).toBeTruthy();
    expect(screen.getByTestId("alert-card-A2")).toBeTruthy();
  });

  it("renders AlertFilter component", () => {
    render(<AlertPanel />);
    expect(screen.getByTestId("alert-filter")).toBeTruthy();
  });

  it("renders at most 120 alerts to keep UI responsive", () => {
    for (let i = 0; i < 130; i += 1) {
      useAlertStore.getState().addAlert(
        makeAlert({
          alertId: `A-${i}`,
          trackId: `TRK-${i}`,
          severity: i % 2 === 0 ? "CRITICAL" : "WATCH",
          detectedAt: new Date(
            `2026-01-01T12:${String(i % 60).padStart(2, "0")}:00Z`,
          ),
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
