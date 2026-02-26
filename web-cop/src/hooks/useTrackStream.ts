// CLASSIFICATION: UNCLASSIFIED
// src/hooks/useTrackStream.ts

import { toJson } from "@bufbuild/protobuf";
import { Code, ConnectError, createClient } from "@connectrpc/connect";
import { FusedTrackSchema } from "@gen/rtsa/entity/v1/fused_track_pb";
import {
  TrackService,
  TrackUpdate_UpdateType,
} from "@gen/rtsa/entity/v1/track_service_pb";
import { useCallback, useEffect, useRef, useState } from "react";
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

const trackClient = createClient(TrackService, transport);

function stripPrefix(val: string | undefined, prefix: string): string {
  if (!val) return "";
  return val.startsWith(prefix) ? val.slice(prefix.length) : val;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function protoTrackToLocal(protoTrack: any): FusedTrack {
  // toJson normalises enums to proto names and Timestamps to RFC 3339 strings.
  const j = toJson(FusedTrackSchema, protoTrack, {
    alwaysEmitImplicit: true,
  }) as Record<string, unknown>;
  const pos = (j["estimatedPosition"] ?? {}) as Record<string, unknown>;
  return {
    trackId: (j["trackId"] ?? "") as string,
    entityType: (stripPrefix(j["entityType"] as string, "ENTITY_TYPE_") ||
      "SURFACE") as EntityType,
    hostileClass: (stripPrefix(
      j["hostileClass"] as string,
      "HOSTILE_CLASSIFICATION_",
    ) || "UNKNOWN") as HostileClassification,
    position: {
      latitude: (pos["latitude"] ?? 0) as number,
      longitude: (pos["longitude"] ?? 0) as number,
      altitudeMeters: pos["altitudeMeters"] as number | undefined,
      speedKnots: pos["speedKnots"] as number | undefined,
      headingDegrees: pos["headingDegrees"] as number | undefined,
    },
    confidenceScore: (j["confidenceScore"] ?? 0) as number,
    sourceCount: (j["sourceCount"] ?? 0) as number,
    sources: ((j["sources"] ?? []) as Array<Record<string, unknown>>).map(
      (s) => ({
        sensorId: (s["sensorId"] ?? "") as string,
        sensorType: (s["sensorType"] ?? "") as string,
        confidence: (s["confidence"] ?? 0) as number,
        lastContribution: s["lastContribution"]
          ? new Date(s["lastContribution"] as string)
          : new Date(),
      }),
    ),
    status: (stripPrefix(j["status"] as string, "TRACK_STATUS_") ||
      "ACTIVE") as TrackStatus,
    classification: (stripPrefix(
      j["classification"] as string,
      "CLASSIFICATION_LEVEL_",
    ) || "UNCLASSIFIED") as ClassificationLevel,
    createdAt: j["createdAt"] ? new Date(j["createdAt"] as string) : new Date(),
    updatedAt: j["updatedAt"] ? new Date(j["updatedAt"] as string) : new Date(),
  };
}

interface StreamState {
  isConnected: boolean;
  error: Error | null;
  reconnectAttempts: number;
}

/**
 * useTrackStream — subscribes to real-time track updates via gRPC-Web server-streaming.
 *
 * Uses the generated TrackService client with @connectrpc/connect-web transport
 * for proper binary protobuf framing and gRPC-Web envelope encoding.
 *
 * Flow:
 *   1. Opens StreamTracks gRPC-Web stream to svc-track via Envoy
 *   2. Receives SNAPSHOT of all current tracks then incremental updates
 *   3. Maps proto FusedTrack → local FusedTrack type and updates TrackStore
 *   4. On connect error: exponential backoff reconnect (1s, 2s … max 30s)
 *   5. Aborts cleanly on unmount
 */
export function useTrackStream(): StreamState {
  const upsertTrack = useTrackStore((s) => s.upsertTrack);
  const removeTrack = useTrackStore((s) => s.removeTrack);

  const [state, setState] = useState<StreamState>({
    isConnected: false,
    error: null,
    reconnectAttempts: 0,
  });

  const abortRef = useRef<AbortController | null>(null);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const attemptsRef = useRef(0);
  const mountedRef = useRef(true);
  const connectedRef = useRef(false);

  const connect = useCallback(() => {
    if (!mountedRef.current) return;

    abortRef.current = new AbortController();
    const signal = abortRef.current.signal;

    void (async () => {
      try {
        for await (const update of trackClient.streamTracks({}, { signal })) {
          if (!mountedRef.current) return;

          if (!connectedRef.current) {
            connectedRef.current = true;
            attemptsRef.current = 0;
            setState({ isConnected: true, error: null, reconnectAttempts: 0 });
          }

          const { updateType, track } = update;
          if (!track) continue;

          if (
            updateType === TrackUpdate_UpdateType.DROPPED ||
            updateType === TrackUpdate_UpdateType.MERGED
          ) {
            removeTrack(track.trackId);
          } else {
            upsertTrack(protoTrackToLocal(track));
          }
        }
        // Stream ended cleanly — reconnect
        if (mountedRef.current) throw new Error("Stream ended");
      } catch (err) {
        if (!mountedRef.current) return;
        if (err instanceof ConnectError && err.code === Code.Canceled) return;

        connectedRef.current = false;

        setState((s) => ({
          isConnected: false,
          error: err instanceof Error ? err : new Error(String(err)),
          reconnectAttempts: s.reconnectAttempts + 1,
        }));
        attemptsRef.current += 1;

        const delay = Math.min(
          BASE_DELAY_MS * Math.pow(2, attemptsRef.current - 1),
          MAX_DELAY_MS,
        );
        timeoutRef.current = setTimeout(connect, delay);
      }
    })();
  }, [upsertTrack, removeTrack]);

  useEffect(() => {
    mountedRef.current = true;
    connect();

    return () => {
      mountedRef.current = false;
      connectedRef.current = false;
      abortRef.current?.abort();
      if (timeoutRef.current !== null) clearTimeout(timeoutRef.current);
    };
  }, [connect]);

  return state;
}
