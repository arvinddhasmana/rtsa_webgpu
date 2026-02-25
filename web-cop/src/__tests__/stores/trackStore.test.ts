// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/stores/trackStore.test.ts

import { describe, it, expect, beforeEach } from "vitest";
import { useTrackStore } from "../../stores/trackStore";
import { FusedTrack } from "../../types/track";

function makeTrack(overrides: Partial<FusedTrack> = {}): FusedTrack {
  return {
    trackId: "TRK-001",
    entityType: "SURFACE",
    hostileClass: "UNKNOWN",
    position: { latitude: 45.0, longitude: -60.0 },
    confidenceScore: 0.9,
    sourceCount: 1,
    sources: [],
    status: "ACTIVE",
    classification: "UNCLASSIFIED",
    createdAt: new Date(),
    updatedAt: new Date(),
    ...overrides,
  };
}

describe("TrackStore", () => {
  beforeEach(() => {
    useTrackStore.getState().clearAll();
  });

  it("T01: upsertTrack adds a track to the store", () => {
    const track = makeTrack({ trackId: "TRK-001" });
    useTrackStore.getState().upsertTrack(track);

    expect(useTrackStore.getState().tracks.size).toBe(1);
    expect(useTrackStore.getState().tracks.get("TRK-001")).toEqual(track);
  });

  it("upsertTrack updates lastUpdateTime", () => {
    const track = makeTrack();
    useTrackStore.getState().upsertTrack(track);
    expect(useTrackStore.getState().lastUpdateTime).toBeInstanceOf(Date);
  });

  it("upsertTrack replaces existing track with same ID", () => {
    const track1 = makeTrack({ confidenceScore: 0.5 });
    const track2 = makeTrack({ confidenceScore: 0.9 });
    useTrackStore.getState().upsertTrack(track1);
    useTrackStore.getState().upsertTrack(track2);

    expect(useTrackStore.getState().tracks.size).toBe(1);
    expect(useTrackStore.getState().tracks.get("TRK-001")?.confidenceScore).toBe(0.9);
  });

  it("removeTrack removes a track by ID", () => {
    useTrackStore.getState().upsertTrack(makeTrack());
    useTrackStore.getState().removeTrack("TRK-001");
    expect(useTrackStore.getState().tracks.size).toBe(0);
  });

  it("removeTrack is a no-op for unknown track ID", () => {
    useTrackStore.getState().upsertTrack(makeTrack());
    useTrackStore.getState().removeTrack("NONEXISTENT");
    expect(useTrackStore.getState().tracks.size).toBe(1);
  });

  it("selectTrack sets selectedTrackId", () => {
    useTrackStore.getState().selectTrack("TRK-001");
    expect(useTrackStore.getState().selectedTrackId).toBe("TRK-001");
  });

  it("selectTrack with null clears selection", () => {
    useTrackStore.getState().selectTrack("TRK-001");
    useTrackStore.getState().selectTrack(null);
    expect(useTrackStore.getState().selectedTrackId).toBeNull();
  });

  it("clearAll removes all tracks and selection", () => {
    useTrackStore.getState().upsertTrack(makeTrack({ trackId: "A" }));
    useTrackStore.getState().upsertTrack(makeTrack({ trackId: "B" }));
    useTrackStore.getState().selectTrack("A");
    useTrackStore.getState().clearAll();

    expect(useTrackStore.getState().tracks.size).toBe(0);
    expect(useTrackStore.getState().selectedTrackId).toBeNull();
  });

  it("getTrackById returns correct track", () => {
    const track = makeTrack({ trackId: "TRK-X" });
    useTrackStore.getState().upsertTrack(track);
    expect(useTrackStore.getState().getTrackById("TRK-X")).toEqual(track);
  });

  it("getTrackById returns undefined for missing ID", () => {
    expect(useTrackStore.getState().getTrackById("MISSING")).toBeUndefined();
  });

  it("T02: getTracksByType filters by entity type", () => {
    useTrackStore.getState().upsertTrack(makeTrack({ trackId: "A", entityType: "AIR" }));
    useTrackStore.getState().upsertTrack(makeTrack({ trackId: "B", entityType: "SURFACE" }));
    useTrackStore.getState().upsertTrack(makeTrack({ trackId: "C", entityType: "AIR" }));

    const airTracks = useTrackStore.getState().getTracksByType("AIR");
    expect(airTracks).toHaveLength(2);
    expect(airTracks.every((t) => t.entityType === "AIR")).toBe(true);
  });

  it("T02: getHostileTracks returns only HOSTILE tracks", () => {
    useTrackStore.getState().upsertTrack(makeTrack({ trackId: "H", hostileClass: "HOSTILE" }));
    useTrackStore.getState().upsertTrack(makeTrack({ trackId: "F", hostileClass: "FRIENDLY" }));
    useTrackStore.getState().upsertTrack(makeTrack({ trackId: "N", hostileClass: "NEUTRAL" }));

    const hostile = useTrackStore.getState().getHostileTracks();
    expect(hostile).toHaveLength(1);
    expect(hostile[0].trackId).toBe("H");
  });

  it("getActiveTrackCount counts only ACTIVE status tracks", () => {
    useTrackStore.getState().upsertTrack(makeTrack({ trackId: "A1", status: "ACTIVE" }));
    useTrackStore.getState().upsertTrack(makeTrack({ trackId: "A2", status: "ACTIVE" }));
    useTrackStore.getState().upsertTrack(makeTrack({ trackId: "S1", status: "STALE" }));

    expect(useTrackStore.getState().getActiveTrackCount()).toBe(2);
  });
});
