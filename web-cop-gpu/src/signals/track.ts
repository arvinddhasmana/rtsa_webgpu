// CLASSIFICATION: UNCLASSIFIED
// src/signals/track.ts — Track selection and detail signals
//
// Bridges the pick-buffer result from the Render Worker into SolidJS reactivity.
// Reference: docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md §3.3

import { createSignal } from "solid-js";

/** Lightweight pick result received directly from the Render Worker. */
export interface PickedTrack {
  /** 32-bit hash matching the SAB slot's track_id_hash field. */
  trackIdHash: number;
  /** Canvas-space X coordinate of the click that triggered the pick. */
  x: number;
  /** Canvas-space Y coordinate of the click that triggered the pick. */
  y: number;
}

/** Full track detail fetched via gRPC QueryService.QueryTracks. */
export interface TrackDetail {
  trackId: string;
  entityType: string;
  hostileClass: string;
  status: string;
  classification: string;
  confidenceScore: number;
  sourceCount: number;
  lat: number;
  lon: number;
  speedKnots: number;
  headingDeg: number;
  altitudeMeters: number;
  label?: string;
  updatedAtMs: number;
}

/** The track most recently clicked on the canvas (null = nothing selected). */
export const [selectedTrack, setSelectedTrack] = createSignal<PickedTrack | null>(null);

/** Full detail for the selected track once fetched (null = not loaded). */
export const [trackDetail, setTrackDetail] = createSignal<TrackDetail | null>(null);

/** Whether the detail fetch is in progress. */
export const [trackDetailLoading, setTrackDetailLoading] = createSignal(false);

/** Error message from the last detail fetch (null = no error). */
export const [trackDetailError, setTrackDetailError] = createSignal<string | null>(null);

/** Clear the selected track and all derived state. */
export function clearSelectedTrack(): void {
  setSelectedTrack(null);
  setTrackDetail(null);
  setTrackDetailLoading(false);
  setTrackDetailError(null);
}
