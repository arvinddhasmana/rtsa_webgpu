// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/SourceAttribution.test.tsx

import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { SourceAttributionSection } from "../../components/detail/SourceAttribution";
import { FusedTrack } from "../../types/track";

const makeTrack = (overrides: Partial<FusedTrack> = {}): FusedTrack => ({
  trackId: "TRK-001",
  entityType: "SURFACE",
  hostileClass: "UNKNOWN",
  position: { latitude: 45, longitude: -60 },
  confidenceScore: 0.9,
  sourceCount: 2,
  sources: [
    {
      sensorId: "RADAR-01",
      sensorType: "RADAR",
      confidence: 0.9,
      lastContribution: new Date("2026-01-01T12:00:00Z"),
    },
    {
      sensorId: "AIS-02",
      sensorType: "AIS",
      confidence: 0.7,
      lastContribution: new Date("2026-01-01T11:00:00Z"),
    },
  ],
  status: "ACTIVE",
  classification: "UNCLASSIFIED",
  createdAt: new Date(),
  updatedAt: new Date(),
  ...overrides,
});

describe("SourceAttributionSection", () => {
  it("renders source attribution container", () => {
    render(<SourceAttributionSection track={makeTrack()} />);
    expect(screen.getByTestId("source-attribution")).toBeTruthy();
  });

  it("shows contributing sensor count", () => {
    render(<SourceAttributionSection track={makeTrack()} />);
    expect(screen.getByText("Contributing Sensors (2)")).toBeTruthy();
  });

  it("renders each sensor ID", () => {
    render(<SourceAttributionSection track={makeTrack()} />);
    expect(screen.getByText("RADAR-01")).toBeTruthy();
    expect(screen.getByText("AIS-02")).toBeTruthy();
  });

  it("renders sensor type", () => {
    render(<SourceAttributionSection track={makeTrack()} />);
    expect(screen.getByText("RADAR")).toBeTruthy();
    expect(screen.getByText("AIS")).toBeTruthy();
  });

  it("shows confidence percentages", () => {
    render(<SourceAttributionSection track={makeTrack()} />);
    expect(screen.getByText("90%")).toBeTruthy();
    expect(screen.getByText("70%")).toBeTruthy();
  });

  it("shows 'No source data available' when sources array is empty", () => {
    render(
      <SourceAttributionSection
        track={makeTrack({ sources: [], sourceCount: 0 })}
      />
    );
    expect(screen.getByText("No source data available")).toBeTruthy();
  });
});
