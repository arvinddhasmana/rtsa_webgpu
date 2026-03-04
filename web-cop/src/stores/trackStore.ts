// CLASSIFICATION: UNCLASSIFIED
// src/stores/trackStore.ts

import { create } from "zustand";
import { FusedTrack } from "../types/track";

const TRACK_HISTORY_MAX = 20;

interface TrackState {
  tracks: Map<string, FusedTrack>;
  trackHistory: Map<string, [number, number][]>; // [lng, lat] tuples
  selectedTrackId: string | null;
  lastUpdateTime: Date | null;

  upsertTrack: (track: FusedTrack) => void;
  batchUpsertTracks: (tracks: FusedTrack[]) => void;
  removeTrack: (trackId: string) => void;
  selectTrack: (trackId: string | null) => void;
  clearAll: () => void;

  getTrackById: (trackId: string) => FusedTrack | undefined;
  getTracksByType: (entityType: string) => FusedTrack[];
  getHostileTracks: () => FusedTrack[];
  getActiveTrackCount: () => number;
}

/** Append new position to history for one track, trimmed to TRACK_HISTORY_MAX. */
function appendHistory(
  history: Map<string, [number, number][]>,
  trackId: string,
  lng: number,
  lat: number,
): Map<string, [number, number][]> {
  const prev = history.get(trackId) ?? [];
  const next = [...prev, [lng, lat] as [number, number]];
  if (next.length > TRACK_HISTORY_MAX) next.splice(0, next.length - TRACK_HISTORY_MAX);
  const newHistory = new Map(history);
  newHistory.set(trackId, next);
  return newHistory;
}

export const useTrackStore = create<TrackState>((set, get) => ({
  tracks: new Map(),
  trackHistory: new Map(),
  selectedTrackId: null,
  lastUpdateTime: null,

  upsertTrack: (track) =>
    set((state) => {
      const newTracks = new Map(state.tracks);
      newTracks.set(track.trackId, track);
      const newHistory = appendHistory(
        state.trackHistory,
        track.trackId,
        track.position.longitude,
        track.position.latitude,
      );
      return { tracks: newTracks, trackHistory: newHistory, lastUpdateTime: new Date() };
    }),

  // Batch variant: applies all updates in a single Map copy → one Zustand
  // notification instead of one per track. Use for stream bursts.
  batchUpsertTracks: (incoming) =>
    set((state) => {
      const newTracks = new Map(state.tracks);
      let newHistory = state.trackHistory;
      for (const track of incoming) {
        newTracks.set(track.trackId, track);
        newHistory = appendHistory(
          newHistory,
          track.trackId,
          track.position.longitude,
          track.position.latitude,
        );
      }
      return { tracks: newTracks, trackHistory: newHistory, lastUpdateTime: new Date() };
    }),

  removeTrack: (trackId) =>
    set((state) => {
      const newTracks = new Map(state.tracks);
      newTracks.delete(trackId);
      const newHistory = new Map(state.trackHistory);
      newHistory.delete(trackId);
      return { tracks: newTracks, trackHistory: newHistory };
    }),

  selectTrack: (trackId) => set({ selectedTrackId: trackId }),
  clearAll: () => set({ tracks: new Map(), trackHistory: new Map(), selectedTrackId: null }),

  getTrackById: (trackId) => get().tracks.get(trackId),
  getTracksByType: (entityType) =>
    Array.from(get().tracks.values()).filter(
      (t) => t.entityType === entityType,
    ),
  getHostileTracks: () =>
    Array.from(get().tracks.values()).filter(
      (t) => t.hostileClass === "HOSTILE",
    ),
  getActiveTrackCount: () =>
    Array.from(get().tracks.values()).filter((t) => t.status === "ACTIVE")
      .length,
}));
