// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/OperatorDashboard.test.tsx

import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { OperatorDashboard } from "../../components/layout/OperatorDashboard";
import { useAlertStore } from "../../stores/alertStore";

// Mock MapView to avoid context/canvas issues in tests
vi.mock("../../components/map/MapView", () => ({
  MapView: () => <div data-testid="mock-map">Map</div>,
}));

describe("OperatorDashboard", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAlertStore.getState().clearAll();
  });

  it("renders the 3-panel dashboard layout", () => {
    render(<OperatorDashboard />);
    expect(screen.getByTestId("operator-dashboard")).toBeTruthy();

    // Map is rendered in background
    expect(screen.getByTestId("mock-map")).toBeTruthy();

    // Left panel (Timeline and Detail)
    expect(screen.getByText("ENTITY TIMELINE")).toBeTruthy();

    // Right panel (Alerts)
    expect(screen.getByTestId("alert-panel")).toBeTruthy();
  });

  it("handles empty alert state gracefully", () => {
    render(<OperatorDashboard />);
    // "No critical alerts" state should be handled implicitly by AlertPanel component
    expect(screen.getByTestId("alert-panel")).toBeTruthy();
  });

  it("renders critical badge when critical alerts exist", () => {
    useAlertStore.getState().addAlert({
      alertId: "A1",
      trackId: "T1",
      anomalyType: "SPEED",
      severity: "CRITICAL",
      confidenceScore: 0.9,
      explanation: "Test",
      features: [],
      classification: "UNCLASSIFIED",
      detectedAt: new Date()
    });

    render(<OperatorDashboard />);
    // Add an assertion for the unacknowledged critical count badge logic if it exists in the operator dashboard
    expect(screen.getByText("1 UNACKNOWLEDGED ALERTS")).toBeTruthy();
  });
});
