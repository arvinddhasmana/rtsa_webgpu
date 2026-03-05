// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/AlertCard.test.tsx

import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { AlertCard } from "../../components/alerts/AlertCard";
import { useAlertStore } from "../../stores/alertStore";
import { useTrackStore } from "../../stores/trackStore";
import { AnomalyAlert } from "../../types/alert";

const makeAlert = (overrides: Partial<AnomalyAlert> = {}): AnomalyAlert => ({
  alertId: "ALT-001",
  trackId: "TRK-001",
  anomalyType: "SPEED",
  severity: "WATCH",
  confidenceScore: 0.85,
  explanation: "Speed anomaly",
  features: [],
  classification: "UNCLASSIFIED",
  detectedAt: new Date("2026-01-01T12:00:00Z"),
  ...overrides,
});

describe("AlertCard", () => {
  beforeEach(() => {
    useAlertStore.getState().clearAll();
    useTrackStore.getState().clearAll();
  });

  it("renders alert severity and anomaly type", () => {
    render(<AlertCard alert={makeAlert({ severity: "ELEVATED", anomalyType: "ROUTE_DEVIATION" })} />);
    expect(screen.getByText("ELEVATED")).toBeTruthy();
    expect(screen.getByText("ROUTE DEVIATION")).toBeTruthy();
  });

  it("renders track link button", () => {
    render(<AlertCard alert={makeAlert({ trackId: "TRK-X" })} />);
    expect(screen.getByTestId("track-link-TRK-X")).toBeTruthy();
  });

  it("acknowledges alert on click", () => {
    const alert = makeAlert({ alertId: "ALT-001" });
    render(<AlertCard alert={alert} />);
    const card = screen.getByTestId("alert-card-ALT-001");
    fireEvent.click(card);
    expect(useAlertStore.getState().acknowledgedIds.has("ALT-001")).toBe(true);
  });

  it("T09: CRITICAL alert has pulse animation style", () => {
    const alert = makeAlert({ severity: "CRITICAL", alertId: "CRIT-001" });
    render(<AlertCard alert={alert} />);
    const card = screen.getByTestId("alert-card-CRIT-001");
    // Check animation is applied (inline style)
    expect(card.style.animation).toContain("pulse");
  });

  it("non-CRITICAL alert does not have pulse animation", () => {
    const alert = makeAlert({ severity: "WATCH", alertId: "W001" });
    render(<AlertCard alert={alert} />);
    const card = screen.getByTestId("alert-card-W001");
    expect(card.style.animation).toBe("");
  });

  it("renders confidence percentage", () => {
    render(<AlertCard alert={makeAlert({ confidenceScore: 0.85 })} />);
    // The component changed rendering from "85% conf" to a visual bar + "85%"
    expect(screen.getByText("85%")).toBeTruthy();
  });

  it("triggers track selection and detail panel on inspect click", () => {
    const alert = makeAlert({ alertId: "ALT-001", trackId: "TRK-001" });
    render(<AlertCard alert={alert} />);
    const inspectBtn = screen.getByTestId("alert-inspect-ALT-001");
    fireEvent.click(inspectBtn);
    expect(useTrackStore.getState().selectedTrackId).toBe("TRK-001");
  });

  it("handles confirm feedback click", async () => {
    const alert = makeAlert({ alertId: "ALT-001" });
    render(<AlertCard alert={alert} />);
    const confirmBtn = screen.getByTestId("alert-confirm-ALT-001");
    fireEvent.click(confirmBtn);
    // Should show optimistic status
    expect(screen.getByText("Confirmed (⏳)")).toBeTruthy();
  });

  it("handles reject feedback click", async () => {
    const alert = makeAlert({ alertId: "ALT-001" });
    render(<AlertCard alert={alert} />);
    const rejectBtn = screen.getByTestId("alert-reject-ALT-001");
    fireEvent.click(rejectBtn);
    // Should show optimistic status
    expect(screen.getByText("Rejected (⏳)")).toBeTruthy();
  });

  it("handles assign click by dispatching custom event", () => {
    const alert = makeAlert({ alertId: "ALT-001" });
    const spy = vi.fn();
    window.addEventListener("open-assign-popover", spy);
    render(<AlertCard alert={alert} />);
    const assignBtn = screen.getByTestId("alert-assign-ALT-001");
    fireEvent.click(assignBtn);
    expect(spy).toHaveBeenCalled();
    const event = spy.mock.calls[0][0] as CustomEvent;
    expect(event.detail.alertId).toBe("ALT-001");
  });

  it("handles keyboard shortcuts on card focus", () => {
    const alert = makeAlert({ alertId: "ALT-001", trackId: "TRK-001" });
    render(<AlertCard alert={alert} />);
    const card = screen.getByTestId("alert-card-ALT-001");
    card.focus();

    // Enter for inspect
    fireEvent.keyDown(card, { key: "Enter" });
    expect(useTrackStore.getState().selectedTrackId).toBe("TRK-001");

    // Space for acknowledge
    fireEvent.keyDown(card, { key: " " });
    expect(useAlertStore.getState().acknowledgedIds.has("ALT-001")).toBe(true);
  });

  it("handles keyboard feedback shortcuts", () => {
    const alert = makeAlert({ alertId: "ALT-001" });
    render(<AlertCard alert={alert} />);
    const card = screen.getByTestId("alert-card-ALT-001");
    card.focus();

    // 'C' for confirm
    fireEvent.keyDown(card, { key: "c" });
    expect(screen.getByText("Confirmed (⏳)")).toBeTruthy();

    // Reset and try 'R'
    useAlertStore.getState().setFeedbackStatus("ALT-001", "");
    fireEvent.keyDown(card, { key: "r" });
    expect(screen.getByText("Rejected (⏳)")).toBeTruthy();
  });

  it("handles keyboard assign shortcut", () => {
    const alert = makeAlert({ alertId: "ALT-001" });
    const spy = vi.fn();
    window.addEventListener("open-assign-popover", spy);
    render(<AlertCard alert={alert} />);
    const card = screen.getByTestId("alert-card-ALT-001");
    card.focus();

    fireEvent.keyDown(card, { key: "a" });
    expect(spy).toHaveBeenCalled();
  });
});
