// CLASSIFICATION: UNCLASSIFIED
// src/hooks/useTrackStream.ts

import { createPromiseClient } from "@connectrpc/connect";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  ClassificationLevel as GenClassificationLevel,
  EntityType as GenEntityType,
  HostileClassification as GenHostileClassification,
  SensorType as GenSensorType,
  TrackStatus as GenTrackStatus,
} from "../../../gen/ts/rtsa/common/v1/types_pb";
import { FusedTrack as ProtoFusedTrack } from "../../../gen/ts/rtsa/entity/v1/fused_track_pb";
import { TrackService } from "../../../gen/ts/rtsa/entity/v1/track_service_connect";
import {
  StreamTracksRequest,
  TrackUpdate_UpdateType,
} from "../../../gen/ts/rtsa/entity/v1/track_service_pb";
import { transport } from "../api/grpc-client";
import { useTrackStore } from "../stores/trackStore";
import type {
  ClassificationLevel,
  EntityType,
  HostileClassification,
  TrackStatus,
} from "../types/common";
import type { FusedTrack } from "../types/track";

const BASE_DELAY_MS = 1000;
const MAX_DELAY_MS = 30000;

function cleanEnum(val: string, prefix: string): string {
  if (!val) return val;
  return val.replace(prefix, "");
}

function protoTrackToLocal(t: ProtoFusedTrack): FusedTrack {
  const pos = t.estimatedPosition;

  return {
    trackId: t.trackId,
    entityType: (cleanEnum(GenEntityType[t.entityType], "ENTITY_TYPE_") ||
      "UNKNOWN") as EntityType,
    hostileClass: (cleanEnum(
      GenHostileClassification[t.hostileClass],
      "HOSTILE_CLASSIFICATION_",
    ) || "UNKNOWN") as HostileClassification,
    position: {
      latitude: pos?.latitude || 0,
      longitude: pos?.longitude || 0,
      altitudeMeters: pos?.altitudeMeters,
      speedKnots: pos?.speedKnots,
      headingDegrees: pos?.headingDegrees,
    },
    status: (cleanEnum(GenTrackStatus[t.status], "TRACK_STATUS_") ||
      "ACTIVE") as TrackStatus,
    confidenceScore: t.confidenceScore,
    classification: (cleanEnum(
      GenClassificationLevel[t.classification],
      "CLASSIFICATION_LEVEL_",
    ) || "UNCLASSIFIED") as ClassificationLevel,
    sourceCount: t.sources.length,
    sources: t.sources.map((s) => ({
      sensorId: s.sensorId,
      sensorType: cleanEnum(GenSensorType[s.sensorType], "SENSOR_TYPE_"),
      confidence: s.confidence,
      lastContribution: s.lastContribution
        ? s.lastContribution.toDate()
        : new Date(),
    })),
    createdAt: t.createdAt ? t.createdAt.toDate() : new Date(),
    updatedAt: t.updatedAt ? t.updatedAt.toDate() : new Date(),
  };
}

interface StreamState {
  isConnected: boolean;
  error: Error | null;
  reconnectAttempts: number;
}

export function useTrackStream(): StreamState {
  const batchUpsertTracks = useTrackStore((s) => s.batchUpsertTracks);
  const removeTrack = useTrackStore((s) => s.removeTrack);

  const [state, setState] = useState<StreamState>({
    isConnected: false,
    error: null,
    reconnectAttempts: 0,
  });

  // State refs to access latest values in callbacks without re-triggering effects
  const isMountedRef = useRef(true);
  const reconnectAttemptsRef = useRef(0);

  // Connection management
  const abortControllerRef = useRef<AbortController | null>(null);
  const retryTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Batching refs — collect updates during a frame, flush once per RAF tick
  const pendingUpsertsRef = useRef<FusedTrack[]>([]);
  const pendingRemovesRef = useRef<string[]>([]);
  const rafRef = useRef<number | null>(null);

  // Keep stable refs to the latest store actions so the RAF flush callback
  // never captures stale closures.
  const batchUpsertRef = useRef(batchUpsertTracks);
  const removeTrackRef = useRef(removeTrack);
  useEffect(() => {
    batchUpsertRef.current = batchUpsertTracks;
  }, [batchUpsertTracks]);
  useEffect(() => {
    removeTrackRef.current = removeTrack;
  }, [removeTrack]);

  // Flush accumulated updates: one Map clone for all upserts, one call per remove.
  const flushPending = useCallback(() => {
    rafRef.current = null;
    const upserts = pendingUpsertsRef.current.splice(0);
    const removes = pendingRemovesRef.current.splice(0);
    if (upserts.length > 0) batchUpsertRef.current(upserts);
    for (const id of removes) removeTrackRef.current(id);
  }, []);

  const connect = useCallback(() => {
    if (!isMountedRef.current) return;

    // Clear any pending retry
    if (retryTimeoutRef.current) {
      clearTimeout(retryTimeoutRef.current);
      retryTimeoutRef.current = null;
    }

    // Abort previous request
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
    abortControllerRef.current = new AbortController();

    // Create client
    const client = createPromiseClient(TrackService, transport);

    // Start stream
    (async () => {
      try {
        const stream = client.streamTracks(new StreamTracksRequest(), {
          signal: abortControllerRef.current?.signal,
        });

        let firstMessageReceived = false;

        for await (const update of stream) {
          if (!isMountedRef.current) break;

          if (!firstMessageReceived) {
            firstMessageReceived = true;
            reconnectAttemptsRef.current = 0;
            setState({ isConnected: true, error: null, reconnectAttempts: 0 });
          }

          if (
            update.updateType === TrackUpdate_UpdateType.DROPPED &&
            update.track
          ) {
            pendingRemovesRef.current.push(update.track.trackId);
          } else if (update.track) {
            pendingUpsertsRef.current.push(protoTrackToLocal(update.track));
          }
          // Schedule a single flush for this animation frame; multiple
          // messages arriving in the same frame are coalesced.
          if (rafRef.current === null) {
            rafRef.current = requestAnimationFrame(flushPending);
          }
        }
      } catch (err: any) {
        if (!isMountedRef.current) return;
        if (err.name === "AbortError") return;

        console.error("Track stream error:", err);

        setState({
          isConnected: false,
          error: err,
          reconnectAttempts: reconnectAttemptsRef.current + 1,
        });

        // Exponential backoff
        const delay = Math.min(
          BASE_DELAY_MS * Math.pow(1.5, reconnectAttemptsRef.current),
          MAX_DELAY_MS,
        );
        reconnectAttemptsRef.current++;

        retryTimeoutRef.current = setTimeout(connect, delay);
      }
    })();
  }, [flushPending]);

  useEffect(() => {
    isMountedRef.current = true;
    connect();

    return () => {
      isMountedRef.current = false;
      if (rafRef.current !== null) {
        cancelAnimationFrame(rafRef.current);
        rafRef.current = null;
      }
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
      if (retryTimeoutRef.current) {
        clearTimeout(retryTimeoutRef.current);
      }
    };
  }, []); // Run once on mount

  return state;
}
