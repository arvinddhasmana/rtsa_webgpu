// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/TrackLayer.test.tsx

import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { TrackLayer } from "../../components/map/TrackLayer";
import { FusedTrack } from "../../types/track";

const makeTrack = (overrides: Partial<FusedTrack> = {}): FusedTrack => ({
  trackId: "TRK-001",
  entityType: "SURFACE",
  hostileClass: "HOSTILE",
  position: { latitude: 45, longitude: -60 },
  confidenceScore: 0.9,
  sourceCount: 1,
  sources: [],
  status: "ACTIVE",
  classification: "UNCLASSIFIED",
  createdAt: new Date(),
  updatedAt: new Date(),
  ...overrides,
});

describe("TrackLayer", () => {
  it("renders track layer container", () => {
    render(<TrackLayer tracks={[]} onTrackClick={vi.fn()} />);
    expect(screen.getByTestId("track-layer")).toBeTruthy();
  });

  it("renders a marker for each track", () => {
    const tracks = [
      makeTrack({ trackId: "A" }),
      makeTrack({ trackId: "B" }),
    ];
    render(<TrackLayer tracks={tracks} onTrackClick={vi.fn()} />);
    expect(screen.getByTestId("track-marker-A")).toBeTruthy();
    expect(screen.getByTestId("track-marker-B")).toBeTruthy();
  });

  it("calls onTrackClick with track ID on marker click", () => {
    const onClick = vi.fn();
    render(
      <TrackLayer
        tracks={[makeTrack({ trackId: "TRK-X" })]}
        onTrackClick={onClick}
      />
    );
    fireEvent.click(screen.getByTestId("track-marker-TRK-X"));
    expect(onClick).toHaveBeenCalledWith("TRK-X");
  });

  it("renders stale tracks with reduced opacity", () => {
    render(
      <TrackLayer
        tracks={[makeTrack({ status: "STALE" })]}
        onTrackClick={vi.fn()}
      />
    );
    const marker = screen.getByTestId("track-marker-TRK-001");
    expect(marker.style.opacity).toBe("0.5");
  });

  it("renders active tracks with full opacity", () => {
    render(
      <TrackLayer
        tracks={[makeTrack({ status: "ACTIVE" })]}
        onTrackClick={vi.fn()}
      />
    );
    const marker = screen.getByTestId("track-marker-TRK-001");
    expect(marker.style.opacity).toBe("1");
  });
});
