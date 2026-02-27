// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/ResultsView.test.tsx

import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ResultsView } from "../../components/forensics/ResultsView";
import { AnomalyAlert } from "../../types/alert";

const makeAlert = (overrides: Partial<AnomalyAlert> = {}): AnomalyAlert => ({
  alertId: "ALT-001",
  trackId: "TRK-001",
  anomalyType: "SPEED",
  severity: "ELEVATED",
  confidenceScore: 0.8,
  explanation: "Speed anomaly",
  features: [],
  classification: "UNCLASSIFIED",
  detectedAt: new Date("2026-01-01T12:00:00Z"),
  ...overrides,
});

describe("ResultsView", () => {
  it("renders results view container", () => {
    render(
      <ResultsView
        tracks={[]}
        alerts={[makeAlert()]}
        totalCount={1}
        classificationCeiling="UNCLASSIFIED"
        onTrackSelect={vi.fn()}
      />
    );
    expect(screen.getByTestId("results-view")).toBeTruthy();
  });

  it("shows result counts", () => {
    render(
      <ResultsView
        tracks={[]}
        alerts={[makeAlert(), makeAlert({ alertId: "A2" })]}
        totalCount={2}
        classificationCeiling="UNCLASSIFIED"
        onTrackSelect={vi.fn()}
      />
    );
    expect(screen.getByText(/2 results/)).toBeTruthy();
  });

  it("renders export CSV button", () => {
    render(
      <ResultsView
        tracks={[]}
        alerts={[]}
        totalCount={0}
        classificationCeiling="PROTECTED_B"
        onTrackSelect={vi.fn()}
      />
    );
    expect(screen.getByTestId("export-csv")).toBeTruthy();
  });

  it("renders alert rows", () => {
    render(
      <ResultsView
        tracks={[]}
        alerts={[makeAlert({ alertId: "A1" }), makeAlert({ alertId: "A2" })]}
        totalCount={2}
        classificationCeiling="UNCLASSIFIED"
        onTrackSelect={vi.fn()}
      />
    );
    fireEvent.click(screen.getByTestId("results-tab-alerts"));
    expect(screen.getByTestId("result-row-A1")).toBeTruthy();
    expect(screen.getByTestId("result-row-A2")).toBeTruthy();
  });

  it("calls onTrackSelect when row clicked", () => {
    const onTrackSelect = vi.fn();
    render(
      <ResultsView
        tracks={[]}
        alerts={[makeAlert({ alertId: "A1", trackId: "TRK-X" })]}
        totalCount={1}
        classificationCeiling="UNCLASSIFIED"
        onTrackSelect={onTrackSelect}
      />
    );
    fireEvent.click(screen.getByTestId("results-tab-alerts"));
    fireEvent.click(screen.getByTestId("result-row-A1"));
    expect(onTrackSelect).toHaveBeenCalledWith("TRK-X");
  });

  it("sorts by severity when clicking SEVERITY header", () => {
    render(
      <ResultsView
        tracks={[]}
        alerts={[
          makeAlert({ alertId: "A1", severity: "WATCH" }),
          makeAlert({ alertId: "A2", severity: "CRITICAL" }),
        ]}
        totalCount={2}
        classificationCeiling="UNCLASSIFIED"
        onTrackSelect={vi.fn()}
      />
    );
    fireEvent.click(screen.getByTestId("results-tab-alerts"));
    fireEvent.click(screen.getByText(/SEVERITY/));
    // Just ensure no crash occurs after sorting
    expect(screen.getByTestId("results-view")).toBeTruthy();
  });
});
