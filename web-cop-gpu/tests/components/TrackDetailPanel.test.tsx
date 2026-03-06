// CLASSIFICATION: UNCLASSIFIED
// tests/components/TrackDetailPanel.test.tsx

import { describe, it, expect, afterEach } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { TrackDetailPanel } from "../../src/components/panels/TrackDetailPanel";
import {
  setSelectedTrack,
  setTrackDetail,
  setTrackDetailLoading,
  setTrackDetailError,
  clearSelectedTrack,
} from "../../src/signals/track";
import type { TrackDetail } from "../../src/signals/track";

const mockTrack: TrackDetail = {
  trackId: "track-abc-123",
  entityType: "Air",
  hostileClass: "Hostile",
  status: "Active",
  classification: "UNCLASSIFIED",
  confidenceScore: 0.92,
  sourceCount: 3,
  lat: 45.1234,
  lon: -75.5678,
  altitudeMeters: 8000,
  speedKnots: 480,
  headingDeg: 270,
  updatedAtMs: Date.now(),
};

afterEach(() => {
  clearSelectedTrack();
});

describe("TrackDetailPanel", () => {
  it("renders nothing when no track is selected", () => {
    const { container } = render(() => <TrackDetailPanel />);
    expect(container.children.length).toBe(0);
  });

  it("renders panel when a track is selected", () => {
    setSelectedTrack({ trackIdHash: 0x1234, x: 100, y: 200 });
    render(() => <TrackDetailPanel />);
    expect(screen.getByLabelText("Track detail panel")).toBeDefined();
  });

  it("shows loading state while fetching detail", () => {
    setSelectedTrack({ trackIdHash: 0x1234, x: 0, y: 0 });
    setTrackDetailLoading(true);
    render(() => <TrackDetailPanel />);
    expect(screen.getByText("Loading…")).toBeDefined();
  });

  it("shows error state on fetch failure", () => {
    setSelectedTrack({ trackIdHash: 0x1234, x: 0, y: 0 });
    setTrackDetailError("Network error");
    render(() => <TrackDetailPanel />);
    expect(screen.getByText("Network error")).toBeDefined();
  });

  it("renders track info when trackDetail is set", () => {
    setSelectedTrack({ trackIdHash: 0x1234, x: 0, y: 0 });
    setTrackDetail(mockTrack);
    render(() => <TrackDetailPanel />);
    expect(screen.getByText("track-abc-123")).toBeDefined();
    expect(screen.getByText("Air")).toBeDefined();
    expect(screen.getByText("Hostile")).toBeDefined();
    expect(screen.getByText("Active")).toBeDefined();
  });

  it("shows confidence percentage correctly", () => {
    setSelectedTrack({ trackIdHash: 0x1234, x: 0, y: 0 });
    setTrackDetail(mockTrack);
    render(() => <TrackDetailPanel />);
    expect(screen.getByText("92.0%")).toBeDefined();
  });

  it("clears panel when close button is clicked", async () => {
    setSelectedTrack({ trackIdHash: 0x1234, x: 0, y: 0 });
    setTrackDetail(mockTrack);
    render(() => <TrackDetailPanel />);

    const closeBtn = screen.getByLabelText("Close track detail panel");
    closeBtn.click();

    // After clicking close, the panel should disappear
    expect(screen.queryByLabelText("Track detail panel")).toBeNull();
  });

  it("shows track hash when detail is not available", () => {
    setSelectedTrack({ trackIdHash: 0x1234, x: 0, y: 0 });
    render(() => <TrackDetailPanel />);
    // No detail set → shows hash
    expect(screen.getByText("0x00001234")).toBeDefined();
  });
});
