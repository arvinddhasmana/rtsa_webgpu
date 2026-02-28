// CLASSIFICATION: UNCLASSIFIED
// src/stores/sensorStore.ts

import { create } from "zustand";
import { RawSensorObservation } from "../types/track";

interface SensorState {
  rawObservations: Map<string, RawSensorObservation>; // keyed by observation_id or synthetic ID

  batchUpsertObservations: (obs: RawSensorObservation[]) => void;
  removeStaleObservations: (olderThanMs: number) => void;
  clearAll: () => void;
}

export const useSensorStore = create<SensorState>((set) => ({
  rawObservations: new Map(),

  batchUpsertObservations: (obs) => {
    set((state) => {
      const nextMap = new Map(state.rawObservations);
      obs.forEach((o) => nextMap.set(o.observationId, o));
      return { rawObservations: nextMap };
    });
  },

  removeStaleObservations: (olderThanMs) => {
    const cutoff = Date.now() - olderThanMs;
    set((state) => {
      let changed = false;
      const nextMap = new Map(state.rawObservations);
      for (const [id, obs] of nextMap.entries()) {
        if (obs.timestamp.getTime() < cutoff) {
          nextMap.delete(id);
          changed = true;
        }
      }
      return changed ? { rawObservations: nextMap } : state;
    });
  },

  clearAll: () => set({ rawObservations: new Map() }),
}));
