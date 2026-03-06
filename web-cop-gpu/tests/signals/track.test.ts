// CLASSIFICATION: UNCLASSIFIED
// tests/signals/track.test.ts

import { describe, it, expect, afterEach } from "vitest";
import {
  selectedTrack,
  setSelectedTrack,
  trackDetail,
  setTrackDetail,
  trackDetailLoading,
  setTrackDetailLoading,
  trackDetailError,
  setTrackDetailError,
  clearSelectedTrack,
} from "../../src/signals/track";
import type { TrackDetail } from "../../src/signals/track";

const mockDetail: TrackDetail = {
  trackId: "t-1",
  entityType: "Air",
  hostileClass: "Hostile",
  status: "Active",
  classification: "UNCLASSIFIED",
  confidenceScore: 0.9,
  sourceCount: 2,
  lat: 45.0,
  lon: -75.0,
  altitudeMeters: 8000,
  speedKnots: 500,
  headingDeg: 180,
  updatedAtMs: Date.now(),
};

afterEach(() => {
  clearSelectedTrack();
});

describe("track signals", () => {
  it("selectedTrack starts null", () => {
    expect(selectedTrack()).toBeNull();
  });

  it("setSelectedTrack updates the signal", () => {
    setSelectedTrack({ trackIdHash: 0xabcd, x: 10, y: 20 });
    expect(selectedTrack()?.trackIdHash).toBe(0xabcd);
  });

  it("trackDetail starts null", () => {
    expect(trackDetail()).toBeNull();
  });

  it("setTrackDetail updates the signal", () => {
    setTrackDetail(mockDetail);
    expect(trackDetail()?.trackId).toBe("t-1");
  });

  it("trackDetailLoading starts false", () => {
    expect(trackDetailLoading()).toBe(false);
  });

  it("setTrackDetailLoading updates the signal", () => {
    setTrackDetailLoading(true);
    expect(trackDetailLoading()).toBe(true);
  });

  it("trackDetailError starts null", () => {
    expect(trackDetailError()).toBeNull();
  });

  it("setTrackDetailError updates the signal", () => {
    setTrackDetailError("Failed to fetch");
    expect(trackDetailError()).toBe("Failed to fetch");
  });

  it("clearSelectedTrack resets all state", () => {
    setSelectedTrack({ trackIdHash: 1, x: 0, y: 0 });
    setTrackDetail(mockDetail);
    setTrackDetailLoading(true);
    setTrackDetailError("err");

    clearSelectedTrack();

    expect(selectedTrack()).toBeNull();
    expect(trackDetail()).toBeNull();
    expect(trackDetailLoading()).toBe(false);
    expect(trackDetailError()).toBeNull();
  });
});
