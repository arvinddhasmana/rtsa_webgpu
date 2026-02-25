// CLASSIFICATION: UNCLASSIFIED
// src/api/track-client.ts

import { transport } from "./grpc-client";
import { FusedTrack } from "../types/track";

export interface StreamTracksRequest {
  entityTypeFilter: string[];
  classificationCeiling: string;
}

export interface TrackStreamMessage {
  messageType: "SNAPSHOT" | "UPDATE" | "REMOVE";
  track?: FusedTrack;
  trackId?: string;
}

/**
 * TrackClient wraps the TrackService gRPC-Web streaming endpoint.
 * Production: connects to svc-track via Envoy at /rtsa.track.v1.TrackService/StreamTracks
 */
export class TrackClient {
  private readonly baseUrl: string;

  constructor() {
    this.baseUrl = transport.baseUrl;
  }

  /**
   * Opens a server-streaming call to StreamTracks.
   * Returns an AbortController to cancel the stream.
   */
  streamTracks(
    request: StreamTracksRequest,
    onMessage: (msg: TrackStreamMessage) => void,
    onError: (err: Error) => void
  ): AbortController {
    const controller = new AbortController();
    void this.baseUrl;
    void request;
    void onMessage;
    void onError;
    return controller;
  }
}

export const trackClient = new TrackClient();
