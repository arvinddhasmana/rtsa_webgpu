// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/MapReplay.test.tsx

import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MapReplay } from "../../components/forensics/MapReplay";
import { FusedTrack } from "../../types/track";

const makeTracks = (count: number): FusedTrack[] =>
  Array.from({ length: count }, (_, i) => ({
    trackId: `TRK-${i}`,
    entityType: "SURFACE" as const,
    hostileClass: "UNKNOWN" as const,
    position: { latitude: 45 + i * 0.1, longitude: -60 + i * 0.1 },
    confidenceScore: 0.9,
    sourceCount: 1,
    sources: [],
    status: "ACTIVE" as const,
    classification: "UNCLASSIFIED" as const,
    createdAt: new Date(),
    updatedAt: new Date(),
  }));

describe("MapReplay", () => {
  it("renders play button", () => {
    render(<MapReplay tracks={makeTracks(3)} />);
    expect(screen.getByTestId("replay-play-pause")).toBeTruthy();
  });

  it("renders scrubber", () => {
    render(<MapReplay tracks={makeTracks(3)} />);
    expect(screen.getByTestId("replay-scrubber")).toBeTruthy();
  });

  it("renders speed selector", () => {
    render(<MapReplay tracks={makeTracks(3)} />);
    expect(screen.getByTestId("replay-speed")).toBeTruthy();
  });

  it("toggles play/pause on click", () => {
    render(<MapReplay tracks={makeTracks(3)} />);
    const btn = screen.getByTestId("replay-play-pause");
    expect(btn.textContent).toContain("PLAY");
    fireEvent.click(btn);
    expect(btn.textContent).toContain("PAUSE");
    fireEvent.click(btn);
    expect(btn.textContent).toContain("PLAY");
  });

  it("updates position when scrubber changes", () => {
    render(<MapReplay tracks={makeTracks(5)} />);
    const scrubber = screen.getByTestId("replay-scrubber");
    fireEvent.change(scrubber, { target: { value: "2" } });
    expect(screen.getByText(/3\s*\/\s*5/)).toBeTruthy();
  });

  it("shows 1/1 for empty tracks array", () => {
    render(<MapReplay tracks={[]} />);
    expect(screen.getByText(/1\s*\/\s*1/)).toBeTruthy();
  });
});
