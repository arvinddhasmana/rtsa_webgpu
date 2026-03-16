// CLASSIFICATION: UNCLASSIFIED
// src/signals/track-selection.ts — Track selection and conflict state

import { createMemo, createSignal } from "solid-js";
import { allObservations } from "./sensor-observations";

export interface TrackConflict {
  trackId: string;
  severity: "LOW" | "MEDIUM" | "HIGH";
  reason: string;
  details: {
    source: string;
    value: string;
  }[];
}

/** Currently selected track ID for detailed audit/monitoring. */
export const [selectedTrackId, setSelectedTrackId] = createSignal<string | null>(null);

/** Observations contributing to the selected track. */
export const contributingObservations = createMemo(() => {
  const trackId = selectedTrackId();
  if (!trackId) return [];
  return allObservations().filter((obs) => obs.correlatedTrackId === trackId);
});

/** Mock conflict state for the selected track (to be driven by backend/fusion engine later). */
export const [trackConflicts, setTrackConflicts] = createSignal<Record<string, TrackConflict>>({});

/** Conflict state for the currently selected track. */
export const selectedTrackConflict = createMemo(() => {
  const id = selectedTrackId();
  return id ? trackConflicts()[id] : null;
});

/** Track Quality Index (TQI) for the selected track.
 * Calculated based on observation age, diversity, and confidence.
 */
export const selectedTrackQualityIndex = createMemo(() => {
  const observations = contributingObservations();
  if (observations.length === 0) return 0;

  const avgConfidence = observations.reduce((acc, o) => acc + o.confidence, 0) / observations.length;
  const diversity = new Set(observations.map((o) => o.type)).size;

  // Simple heuristic: higher diversity and higher confidence = higher quality
  return Math.min(1.0, (avgConfidence * 0.7) + (diversity * 0.1));
});
