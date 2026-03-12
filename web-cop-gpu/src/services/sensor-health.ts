// CLASSIFICATION: UNCLASSIFIED
// src/services/sensor-health.ts — Sensor health monitoring service
//
// Reference: docs/business/usecases/UC017_sensor_health_monitoring.md

import { createPromiseClient } from "@connectrpc/connect";
import { SensorType } from "@gen/rtsa/common/v1/types_pb.js";
import { IngestionService } from "@gen/rtsa/ingestion/v1/ingestion_service_connect.js";
import { transport } from "./grpc-client";

const client = createPromiseClient(IngestionService, transport);

export interface SensorStatus {
  sensorId: string;
  sensorType: string;
  status: "CONNECTED" | "STALE" | "OFFLINE";
  eventsPerSecond: number;
  totalReceived: number;
  lastSeenSeconds: number;
  validationPassRate: number;
  dlqCount: number;
}

/** Maps SensorType enum to human-readable labels. */
export function sensorTypeLabel(t: SensorType): string {
  switch (t) {
    case SensorType.RADAR: return "RADAR";
    case SensorType.EW_SIGINT: return "EW/SIGINT";
    case SensorType.ELINT_COMINT: return "ELINT/COMINT";
    case SensorType.ISR: return "ISR";
    case SensorType.AIS_BFT: return "AIS/BFT";
    case SensorType.CYBER: return "CYBER";
    default: return "UNKNOWN";
  }
}

/**
 * Fetches all known sensor statuses and applies health logic.
 * Logic:
 *  - Green (CONNECTED): last_observation_time < 30s ago
 *  - Amber (STALE): 30s <= last_observation_time < 2m
 *  - Red (OFFLINE): last_observation_time >= 2m or s.connected = false
 */
export async function fetchSensorStatuses(): Promise<SensorStatus[]> {
  const response = await client.listSensorStatuses({
    activeWithinSeconds: 0,
    sensorTypes: [],
  });

  const now = Date.now();

  return response.sensors.map((s) => {
    const lastSeenMs = s.lastObservationTime
      ? Number(s.lastObservationTime.seconds) * 1000 + (s.lastObservationTime.nanos / 1000000)
      : 0;
    const diffSeconds = lastSeenMs > 0 ? (now - lastSeenMs) / 1000 : Infinity;

    let status: "CONNECTED" | "STALE" | "OFFLINE" = "OFFLINE";
    if (s.connected) {
      if (diffSeconds < 30) {
        status = "CONNECTED";
      } else if (diffSeconds < 120) {
        status = "STALE";
      }
    }

    const totalReceived = Number(s.totalReceived);
    const totalAccepted = Number(s.totalAccepted);
    const passRate = totalReceived > 0 ? (totalAccepted / totalReceived) * 100 : 0;

    return {
      sensorId: s.sensorId,
      sensorType: sensorTypeLabel(s.sensorType),
      status,
      eventsPerSecond: Math.round(s.eventsPerSecond * 10) / 10,
      totalReceived,
      lastSeenSeconds: diffSeconds === Infinity ? -1 : Math.floor(diffSeconds),
      validationPassRate: Math.round(passRate * 10) / 10,
      dlqCount: Number(s.totalRejected),
    };
  });
}
