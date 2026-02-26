// CLASSIFICATION: UNCLASSIFIED
// src/stores/trackStore.ts

import { create } from "zustand";
import { FusedTrack } from "../types/track";

interface TrackState {
  tracks: Map<string, FusedTrack>;
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

export const useTrackStore = create<TrackState>((set, get) => ({
  tracks: new Map(),
  selectedTrackId: null,
  lastUpdateTime: null,

  upsertTrack: (track) =>
    set((state) => {
      const newTracks = new Map(state.tracks);
      newTracks.set(track.trackId, track);
      return { tracks: newTracks, lastUpdateTime: new Date() };
    }),

  // Batch variant: applies all updates in a single Map copy → one Zustand
  // notification instead of one per track. Use for stream bursts.
  batchUpsertTracks: (incoming) =>
    set((state) => {
      const newTracks = new Map(state.tracks);
      for (const track of incoming) {
        newTracks.set(track.trackId, track);
      }
      return { tracks: newTracks, lastUpdateTime: new Date() };
    }),

  removeTrack: (trackId) =>
    set((state) => {
      const newTracks = new Map(state.tracks);
      newTracks.delete(trackId);
      return { tracks: newTracks };
    }),

  selectTrack: (trackId) => set({ selectedTrackId: trackId }),
  clearAll: () => set({ tracks: new Map(), selectedTrackId: null }),

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
