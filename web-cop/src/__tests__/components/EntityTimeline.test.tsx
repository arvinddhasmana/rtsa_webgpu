// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/EntityTimeline.test.tsx

import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { EntityTimeline } from "../../components/detail/EntityTimeline";
import { FusedTrack } from "../../types/track";

const makeTrack = (overrides: Partial<FusedTrack> = {}): FusedTrack => ({
  trackId: "TRK-001",
  entityType: "SURFACE",
  hostileClass: "UNKNOWN",
  position: { latitude: 45, longitude: -60 },
  confidenceScore: 0.9,
  sourceCount: 1,
  sources: [],
  status: "ACTIVE",
  classification: "UNCLASSIFIED",
  createdAt: new Date("2026-01-01T10:00:00.000Z"),
  updatedAt: new Date("2026-01-01T12:00:00.000Z"),
  ...overrides,
});

describe("EntityTimeline", () => {
  it("renders timeline container", () => {
    render(<EntityTimeline track={makeTrack()} />);
    expect(screen.getByTestId("entity-timeline")).toBeTruthy();
  });

  it("renders 'Timeline' label", () => {
    render(<EntityTimeline track={makeTrack()} />);
    expect(screen.getByText("Timeline")).toBeTruthy();
  });

  it("renders Track created event", () => {
    render(<EntityTimeline track={makeTrack()} />);
    expect(screen.getByText("Track created")).toBeTruthy();
  });

  it("renders Last update event", () => {
    render(<EntityTimeline track={makeTrack()} />);
    expect(screen.getByText("Last update")).toBeTruthy();
  });

  it("shows creation date in Zulu ISO format", () => {
    render(<EntityTimeline track={makeTrack()} />);
    // formatZulu returns ISO string with Z suffix
    expect(screen.getByText(/2026-01-01T10:00:00/)).toBeTruthy();
  });

  it("shows update time in Zulu ISO format", () => {
    render(<EntityTimeline track={makeTrack()} />);
    expect(screen.getByText(/2026-01-01T12:00:00/)).toBeTruthy();
  });
});
