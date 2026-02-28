// CLASSIFICATION: UNCLASSIFIED
// src/hooks/useSensorStream.ts

import { createPromiseClient } from "@connectrpc/connect";
import { useCallback, useEffect, useRef, useState } from "react";
import { SensorType } from "../../../gen/ts/rtsa/common/v1/types_pb";
import { TrackService } from "../../../gen/ts/rtsa/entity/v1/track_service_connect";
import { StreamSensorObservationsRequest } from "../../../gen/ts/rtsa/entity/v1/track_service_pb";
import { SensorObservation } from "../../../gen/ts/rtsa/ingestion/v1/sensor_observation_pb";
import { transport } from "../api/grpc-client";
import { useSensorStore } from "../stores/sensorStore";
import { RawSensorObservation } from "../types/track";

const BASE_DELAY_MS = 1000;
const MAX_DELAY_MS = 30000;

function cleanEnum(val: string, prefix: string): string {
  if (!val) return val;
  return val.replace(prefix, "");
}

function protoObservationToLocal(
  protoObj: SensorObservation,
  correlatedTrackId?: string
): RawSensorObservation | null {
  const detail = protoObj.sensorData;
  if (!detail || !detail.case) return null;

  // Attempt to extract position using known structures
  let lat = 0, lon = 0, alt: number | undefined, spd: number | undefined, hdg: number | undefined;

  const val = detail.value as any;
  if (val.position) {
    lat = val.position.latitude;
    lon = val.position.longitude;
    alt = val.position.altitudeMeters;
    spd = val.position.speedKnots;
    hdg = val.position.headingDegrees;
  } else if (val.estimatedPosition) {
    lat = val.estimatedPosition.latitude;
    lon = val.estimatedPosition.longitude;
    alt = val.estimatedPosition.altitudeMeters;
    spd = val.estimatedPosition.speedKnots;
    hdg = val.estimatedPosition.headingDegrees;
  } else if (detail.case === "cyber") {
    // Cyber IOCs typically don't have lat/lon in this context
    return null;
  }

  return {
    observationId: protoObj.observationId,
    sensorId: protoObj.sensorId,
    sensorType: cleanEnum(SensorType[protoObj.sensorType], "SENSOR_TYPE_"),
    timestamp: protoObj.observationTime ? protoObj.observationTime.toDate() : new Date(),
    latitude: lat,
    longitude: lon,
    altitudeMeters: alt,
    speedKnots: spd,
    headingDegrees: hdg,
    correlatedTrackId: correlatedTrackId,
  };
}

interface StreamState {
  isConnected: boolean;
  error: Error | null;
  reconnectAttempts: number;
}

export function useSensorStream(): StreamState {
  const batchUpsertObservations = useSensorStore((s) => s.batchUpsertObservations);
  const removeStaleObservations = useSensorStore((s) => s.removeStaleObservations);

  const [state, setState] = useState<StreamState>({
    isConnected: false,
    error: null,
    reconnectAttempts: 0,
  });

  const isMountedRef = useRef(true);
  const reconnectAttemptsRef = useRef(0);
  const abortControllerRef = useRef<AbortController | null>(null);
  const retryTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const pendingUpsertsRef = useRef<RawSensorObservation[]>([]);
  const rafRef = useRef<number | null>(null);

  const batchUpsertRef = useRef(batchUpsertObservations);
  useEffect(() => {
    batchUpsertRef.current = batchUpsertObservations;
  }, [batchUpsertObservations]);

  // Clean up stale observations every 5 seconds (decay after 10s)
  useEffect(() => {
    const interval = setInterval(() => {
      removeStaleObservations(10000);
    }, 5000);
    return () => clearInterval(interval);
  }, [removeStaleObservations]);

  const flushPending = useCallback(() => {
    rafRef.current = null;
    const upserts = pendingUpsertsRef.current.splice(0);
    if (upserts.length > 0) batchUpsertRef.current(upserts);
  }, []);

  const connect = useCallback(() => {
    if (!isMountedRef.current) return;

    if (retryTimeoutRef.current) {
      clearTimeout(retryTimeoutRef.current);
      retryTimeoutRef.current = null;
    }
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
    abortControllerRef.current = new AbortController();

    const client = createPromiseClient(TrackService, transport);

    (async () => {
      try {
        // Stream all sensors currently available.
        const stream = client.streamSensorObservations(new StreamSensorObservationsRequest(), {
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

          if (update.observation) {
            const obs = protoObservationToLocal(update.observation, update.correlatedTrackId);
            if (obs) {
              pendingUpsertsRef.current.push(obs);
              if (rafRef.current === null) {
                rafRef.current = requestAnimationFrame(flushPending);
              }
            }
          }
        }
      } catch (err: any) {
        if (!isMountedRef.current) return;
        if (err.name === "AbortError") return;

        console.error("Sensor stream error:", err);

        setState({
          isConnected: false,
          error: err,
          reconnectAttempts: reconnectAttemptsRef.current + 1,
        });

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
