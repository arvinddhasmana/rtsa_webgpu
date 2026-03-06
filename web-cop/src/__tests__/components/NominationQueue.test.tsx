// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/NominationQueue.test.tsx

import { render, screen, fireEvent } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { NominationQueue } from "../../components/nato/NominationQueue";
import type { FusedTrack } from "../../types/track";

const makeTrack = (id: string, confidence = 0.85, cls = "UNCLASSIFIED"): FusedTrack => ({
  trackId: id,
  entityType: "SURFACE",
  hostileClass: "NEUTRAL",
  position: { latitude: 45, longitude: -60 },
  confidenceScore: confidence,
  sourceCount: 1,
  sources: [],
  status: "ACTIVE",
  classification: cls as any,
  createdAt: new Date(),
  updatedAt: new Date(),
});

describe("NominationQueue", () => {
  it("renders the nomination queue container", () => {
    render(<NominationQueue tracks={[]} />);
    expect(screen.getByTestId("nomination-queue")).toBeTruthy();
  });

  it("shows empty state when no qualifying tracks", () => {
    // Track with low confidence won't qualify
    render(<NominationQueue tracks={[makeTrack("T1", 0.5)]} />);
    expect(screen.getByText(/No tracks qualify/i)).toBeTruthy();
  });

  it("shows nomination entry for qualifying track", () => {
    const tracks = [makeTrack("TRK-HIGH", 0.85)];
    render(<NominationQueue tracks={tracks} />);
    expect(screen.getByTestId("nomination-entry-TRK-HIGH")).toBeTruthy();
  });

  it("nominate button transitions status", () => {
    const tracks = [makeTrack("TRK-001", 0.9, "UNCLASSIFIED")];
    render(<NominationQueue tracks={tracks} />);

    const nominateBtn = screen.getByTestId("nominate-btn-TRK-001");
    fireEvent.click(nominateBtn);

    // After nomination, Revoke button should appear
    expect(screen.getByTestId("revoke-btn-TRK-001")).toBeTruthy();
  });

  it("revoke button transitions nominated track back", () => {
    const tracks = [makeTrack("TRK-002", 0.9, "UNCLASSIFIED")];
    render(<NominationQueue tracks={tracks} />);

    fireEvent.click(screen.getByTestId("nominate-btn-TRK-002"));
    fireEvent.click(screen.getByTestId("revoke-btn-TRK-002"));

    // After revoke, track should show REVOKED status
    expect(screen.getByText("REVOKED")).toBeTruthy();
  });

  it("blocks track with high classification", () => {
    const blockedTrack = makeTrack("TRK-SEC", 0.9, "SECRET");
    render(
      <NominationQueue
        tracks={[blockedTrack]}
        maxShareableClassification="PROTECTED_A"
      />
    );
    expect(screen.getByText("BLOCKED")).toBeTruthy();
  });
});
