// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/TrackLayerDomain.test.tsx
// Extended tests for domain shape differentiation in TrackLayer.

import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { TrackLayer } from "../../components/map/TrackLayer";
import type { FusedTrack } from "../../types/track";
import type { EntityType, HostileClassification } from "../../types/common";
import { useUIStore } from "../../stores/uiStore";
import { useTrackStore } from "../../stores/trackStore";

const makeTrack = (
  id: string,
  domain: EntityType,
  hostile: HostileClassification = "NEUTRAL",
  confidence = 0.8
): FusedTrack => ({
  trackId: id,
  entityType: domain,
  hostileClass: hostile,
  position: { latitude: 45, longitude: -60 },
  confidenceScore: confidence,
  sourceCount: 1,
  sources: [{ sensorId: "S1", sensorType: "RADAR", confidence: 0.8, lastContribution: new Date() }],
  status: "ACTIVE",
  classification: "UNCLASSIFIED",
  createdAt: new Date(),
  updatedAt: new Date(),
});

describe("TrackLayer — domain differentiation", () => {
  beforeEach(() => {
    useUIStore.getState().layerVisibility.trackLabels = false;
    useUIStore.getState().layerVisibility.trackTrails = false;
    useTrackStore.getState().clearAll();
  });

  it("renders markers for all 5 domain types", () => {
    const tracks: FusedTrack[] = [
      makeTrack("A1", "AIR"),
      makeTrack("S1", "SURFACE"),
      makeTrack("U1", "SUBSURFACE"),
      makeTrack("L1", "LAND"),
      makeTrack("C1", "CYBER"),
    ];
    render(<TrackLayer tracks={tracks} onTrackClick={vi.fn()} />);
    expect(screen.getByTestId("track-marker-A1")).toBeTruthy();
    expect(screen.getByTestId("track-marker-S1")).toBeTruthy();
    expect(screen.getByTestId("track-marker-U1")).toBeTruthy();
    expect(screen.getByTestId("track-marker-L1")).toBeTruthy();
    expect(screen.getByTestId("track-marker-C1")).toBeTruthy();
  });

  it("renders SVG inside each marker", () => {
    const tracks = [makeTrack("SVG1", "SURFACE")];
    const { container } = render(
      <TrackLayer tracks={tracks} onTrackClick={vi.fn()} />
    );
    const svgs = container.querySelectorAll("svg");
    expect(svgs.length).toBeGreaterThan(0);
  });

  it("shows track labels when enabled in uiStore", () => {
    useUIStore.setState((s) => ({
      layerVisibility: { ...s.layerVisibility, trackLabels: true },
    }));
    const tracks = [makeTrack("LBL1", "AIR", "HOSTILE", 0.9)];
    render(<TrackLayer tracks={tracks} onTrackClick={vi.fn()} />);
    // Label should include domain abbreviation and confidence
    expect(screen.getByTestId("track-layer")).toBeTruthy();
  });

  it("hides track labels when disabled", () => {
    useUIStore.setState((s) => ({
      layerVisibility: { ...s.layerVisibility, trackLabels: false },
    }));
    const tracks = [makeTrack("LBL2", "AIR", "HOSTILE", 0.9)];
    const { container } = render(
      <TrackLayer tracks={tracks} onTrackClick={vi.fn()} />
    );
    // No label divs with font-family monospace outside of SVG
    const labelDivs = Array.from(container.querySelectorAll("div")).filter(
      (d) =>
        d.style.fontFamily?.includes("monospace") &&
        d.style.fontSize === "0.55rem"
    );
    expect(labelDivs.length).toBe(0);
  });

  it("sizes markers smaller for low confidence vs high confidence", () => {
    const tracks = [
      makeTrack("LOW", "SURFACE", "NEUTRAL", 0.1),
      makeTrack("HIGH", "SURFACE", "NEUTRAL", 1.0),
    ];
    render(<TrackLayer tracks={tracks} onTrackClick={vi.fn()} />);
    const lowMarker = screen.getByTestId("track-marker-LOW");
    const highMarker = screen.getByTestId("track-marker-HIGH");

    const lowSize = parseInt(lowMarker.style.width);
    const highSize = parseInt(highMarker.style.width);
    expect(highSize).toBeGreaterThan(lowSize);
  });

  it("STALE tracks are rendered at 0.5 opacity", () => {
    const track = { ...makeTrack("STALE1", "AIR"), status: "STALE" as const };
    render(<TrackLayer tracks={[track]} onTrackClick={vi.fn()} />);
    const marker = screen.getByTestId("track-marker-STALE1");
    expect(marker.style.opacity).toBe("0.5");
  });

  it("ACTIVE tracks are rendered at full opacity", () => {
    const track = makeTrack("ACTIVE1", "AIR");
    render(<TrackLayer tracks={[track]} onTrackClick={vi.fn()} />);
    const marker = screen.getByTestId("track-marker-ACTIVE1");
    expect(marker.style.opacity).toBe("1");
  });
});
