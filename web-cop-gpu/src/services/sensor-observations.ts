// CLASSIFICATION: UNCLASSIFIED
// src/services/sensor-observations.ts — Observation stream adapter

import { createPromiseClient } from "@connectrpc/connect";
import { ClassificationLevel } from "@gen/rtsa/common/v1/types_pb.js";
import { TrackService } from "@gen/rtsa/entity/v1/track_service_connect.js";
import { createEffect } from "solid-js";
import { updateObservations } from "../signals/sensor-observations.js";
import { boundingBox, role } from "../signals/viewport.js";
import { transport } from "./grpc-client";

const client = createPromiseClient(TrackService, transport);
let activeController: AbortController | null = null;

/**
 * Internal function to start the actual gRPC stream.
 */
async function runStream(signal: AbortSignal) {
  try {
    const bb = boundingBox();
    const stream = client.streamSensorObservations(
      {
        boundingBox: {
          minLatitude: bb.minLat,
          maxLatitude: bb.maxLat,
          minLongitude: bb.minLon,
          maxLongitude: bb.maxLon,
        },
        clearanceLevel: ClassificationLevel.UNCLASSIFIED,
        sensorTypes: [],
      },
      { signal }
    );

    for await (const update of stream) {
      if (update.observation) {
        updateObservations(update);
      }
    }
    console.log("[SensorObservationService] Stream finished normally");
  } catch (err) {
    console.error("[SensorObservationService] Stream error:", err);
  }
}

/**
 * Starts the observation stream management.
 * Automatically starts/stops/restarts based on role and boundingBox signals.
 */
export function startObservationStream(): void {
  console.log("[SensorObservationService] Management started");
  createEffect(() => {
    // Only stream if in Operations Commander role
    if (role() !== "operations_commander") {
      if (activeController) {
        activeController.abort();
        activeController = null;
      }
      return;
    }

    // Restart stream whenever boundingBox changes
    const bb = boundingBox();
    void bb.minLat; // Satisfy lint

    if (activeController) {
      activeController.abort();
    }

    activeController = new AbortController();
    runStream(activeController.signal);
  });
}

/**
 * Manually stops the observation stream.
 */
export function stopObservationStream(): void {
  if (activeController) {
    activeController.abort();
    activeController = null;
  }
}
