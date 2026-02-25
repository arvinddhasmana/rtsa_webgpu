// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/DetailPanel.test.tsx

import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { DetailPanel } from "../../components/detail/DetailPanel";
import { useTrackStore } from "../../stores/trackStore";
import { useAuthStore } from "../../stores/authStore";
import { FusedTrack } from "../../types/track";

const makeTrack = (overrides: Partial<FusedTrack> = {}): FusedTrack => ({
  trackId: "TRK-001",
  entityType: "SURFACE",
  hostileClass: "UNKNOWN",
  position: { latitude: 45.1, longitude: -60.2, speedKnots: 12.5, headingDegrees: 270 },
  confidenceScore: 0.9,
  sourceCount: 2,
  sources: [
    { sensorId: "RADAR-01", sensorType: "RADAR", confidence: 0.9, lastContribution: new Date() },
  ],
  status: "ACTIVE",
  classification: "UNCLASSIFIED",
  createdAt: new Date("2026-01-01T10:00:00Z"),
  updatedAt: new Date("2026-01-01T12:00:00Z"),
  ...overrides,
});

describe("DetailPanel", () => {
  beforeEach(() => {
    useTrackStore.getState().clearAll();
    useAuthStore.getState().setOperator({
      id: "op-001",
      name: "Test Operator",
      unit: "TEST",
      clearance: "PROTECTED_B",
      roles: ["OPERATOR"],
    });
  });

  it("shows empty state when no track selected", () => {
    render(<DetailPanel />);
    expect(screen.getByTestId("detail-panel-empty")).toBeTruthy();
  });

  it("shows not found when track ID selected but not in store", () => {
    useTrackStore.getState().selectTrack("MISSING");
    render(<DetailPanel />);
    expect(screen.getByTestId("detail-panel-not-found")).toBeTruthy();
  });

  it("T12: shows detail panel with tabs when track is selected", () => {
    const track = makeTrack();
    useTrackStore.getState().upsertTrack(track);
    useTrackStore.getState().selectTrack("TRK-001");
    render(<DetailPanel />);
    expect(screen.getByTestId("detail-panel")).toBeTruthy();
    expect(screen.getByTestId("tab-identity")).toBeTruthy();
    expect(screen.getByTestId("tab-position")).toBeTruthy();
    expect(screen.getByTestId("tab-sources")).toBeTruthy();
    expect(screen.getByTestId("tab-timeline")).toBeTruthy();
    expect(screen.getByTestId("tab-feedback")).toBeTruthy();
  });

  it("T12: identity section shows track details", () => {
    const track = makeTrack({ trackId: "TRK-001", hostileClass: "FRIENDLY" });
    useTrackStore.getState().upsertTrack(track);
    useTrackStore.getState().selectTrack("TRK-001");
    render(<DetailPanel />);
    expect(screen.getByTestId("identity-section")).toBeTruthy();
    expect(screen.getByText("TRK-001")).toBeTruthy();
  });

  it("shows restricted message when operator lacks clearance", () => {
    const track = makeTrack({ classification: "SECRET" });
    useTrackStore.getState().upsertTrack(track);
    useTrackStore.getState().selectTrack("TRK-001");
    render(<DetailPanel />);
    expect(screen.getByTestId("detail-panel-restricted")).toBeTruthy();
  });

  it("switches to position tab on click", () => {
    const track = makeTrack();
    useTrackStore.getState().upsertTrack(track);
    useTrackStore.getState().selectTrack("TRK-001");
    render(<DetailPanel />);
    fireEvent.click(screen.getByTestId("tab-position"));
    expect(screen.getByTestId("position-section")).toBeTruthy();
  });

  it("switches to feedback tab on click", () => {
    const track = makeTrack();
    useTrackStore.getState().upsertTrack(track);
    useTrackStore.getState().selectTrack("TRK-001");
    render(<DetailPanel />);
    fireEvent.click(screen.getByTestId("tab-feedback"));
    expect(screen.getByTestId("feedback-form")).toBeTruthy();
  });
});
