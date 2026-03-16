// CLASSIFICATION: UNCLASSIFIED
// src/signals/sensor-observations.ts — State store for sensor observations

import { SensorType } from "@gen/rtsa/common/v1/types_pb.js";
import type { SensorObservationUpdate } from "@gen/rtsa/entity/v1/track_service_pb.js";
import { createMemo, createSignal } from "solid-js";

export interface ObservationState {
  id: string;
  type: SensorType;
  lat: number;
  lon: number;
  timestampMs: number;
  correlatedTrackId?: string;
  confidence: number;
}

const [observations, setObservations] = createSignal<Record<string, ObservationState>>({});

/** Update the observation state with a new update from the stream. */
export function updateObservations(update: SensorObservationUpdate): void {
  const obs = update.observation;
  if (!obs) return;

  const id = obs.observationId;

  // Extract confidence from sensor-specific data
  let confidence = 0;
  if (obs.sensorData.case === "radar") confidence = obs.sensorData.value.trackQuality;
  else if (obs.sensorData.case === "ewSigint") confidence = obs.sensorData.value.confidence;
  else if (obs.sensorData.case === "elintComint") confidence = obs.sensorData.value.confidence;
  else if (obs.sensorData.case === "cyber") confidence = obs.sensorData.value.confidence;
  else if (obs.sensorData.case === "isr") {
    const detections = obs.sensorData.value.detections;
    if (detections.length > 0) {
      confidence = detections.reduce((acc, d) => acc + d.confidence, 0) / detections.length;
    }
  }

  const state: ObservationState = {
    id,
    type: obs.sensorType,
    lat: obs.position?.latitude ?? 0,
    lon: obs.position?.longitude ?? 0,
    timestampMs: obs.observationTime ? Number(obs.observationTime.seconds) * 1000 : Date.now(),
    correlatedTrackId: update.correlatedTrackId,
    confidence,
  };

  setObservations((prev) => ({
    ...prev,
    [id]: state,
  }));
}

/** Get all current observations as an array. */
export const allObservations = createMemo(() => Object.values(observations()));

/** Derived KPI: Total observation count. */
export const activeObservationCount = createMemo(() => allObservations().length);

/** Derived KPI: Observation density by sensor type. */
export const observationTypeDistribution = createMemo(() => {
  const dist: Record<number, number> = {};
  for (const obs of allObservations()) {
    dist[obs.type] = (dist[obs.type] ?? 0) + 1;
  }
  return dist;
});

/** Derived KPI: Average confidence level. */
export const averageObservationConfidence = createMemo(() => {
  const obs = allObservations();
  if (obs.length === 0) return 0;
  const sum = obs.reduce((acc, o) => acc + o.confidence, 0);
  return sum / obs.length;
});
