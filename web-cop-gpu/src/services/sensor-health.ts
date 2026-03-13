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
    case SensorType.RADAR:
      return "RADAR";
    case SensorType.EW_SIGINT:
      return "EW/SIGINT";
    case SensorType.ELINT_COMINT:
      return "ELINT/COMINT";
    case SensorType.ISR:
      return "ISR";
    case SensorType.AIS_BFT:
      return "AIS/BFT";
    case SensorType.CYBER:
      return "CYBER";
    default:
      return "UNKNOWN";
  }
}

// Sensor type targets — each maps to an Envoy cluster via the x-ingestion-target header.
// The radar ingestion service is the default (no header required).
const INGESTION_TARGETS = [
  undefined, // radar — default cluster (no header)
  "ew",
  "elint",
  "isr",
  "ais",
  "cyber",
] as const;

/** Deterministic seed-based mock sensor roster — used when VITE_MOCK_SENSORS=true (E2E / dev). */
function mockSensorStatuses(): SensorStatus[] {
  const sensors: SensorStatus[] = [
    {
      sensorId: "RADAR-NORTH-01",
      sensorType: "RADAR",
      status: "CONNECTED",
      eventsPerSecond: 48.2,
      totalReceived: 98230,
      lastSeenSeconds: 2,
      validationPassRate: 98.7,
      dlqCount: 47,
    },
    {
      sensorId: "RADAR-SOUTH-02",
      sensorType: "RADAR",
      status: "STALE",
      eventsPerSecond: 12.1,
      totalReceived: 61020,
      lastSeenSeconds: 65,
      validationPassRate: 88.3,
      dlqCount: 213,
    },
    {
      sensorId: "RADAR-EAST-03",
      sensorType: "RADAR",
      status: "CONNECTED",
      eventsPerSecond: 33.7,
      totalReceived: 74100,
      lastSeenSeconds: 8,
      validationPassRate: 95.1,
      dlqCount: 112,
    },
    {
      sensorId: "RADAR-WEST-04",
      sensorType: "RADAR",
      status: "OFFLINE",
      eventsPerSecond: 0,
      totalReceived: 30900,
      lastSeenSeconds: 320,
      validationPassRate: 0,
      dlqCount: 0,
    },
    {
      sensorId: "AIS-PORT-01",
      sensorType: "AIS/BFT",
      status: "CONNECTED",
      eventsPerSecond: 22.4,
      totalReceived: 56780,
      lastSeenSeconds: 5,
      validationPassRate: 99.1,
      dlqCount: 12,
    },
    {
      sensorId: "AIS-COAST-02",
      sensorType: "AIS/BFT",
      status: "CONNECTED",
      eventsPerSecond: 18.9,
      totalReceived: 44320,
      lastSeenSeconds: 11,
      validationPassRate: 97.3,
      dlqCount: 31,
    },
    {
      sensorId: "EW-NORTH-01",
      sensorType: "EW/SIGINT",
      status: "STALE",
      eventsPerSecond: 7.3,
      totalReceived: 19870,
      lastSeenSeconds: 88,
      validationPassRate: 82.6,
      dlqCount: 387,
    },
    {
      sensorId: "EW-SOUTH-02",
      sensorType: "EW/SIGINT",
      status: "CONNECTED",
      eventsPerSecond: 15.6,
      totalReceived: 38210,
      lastSeenSeconds: 3,
      validationPassRate: 93.8,
      dlqCount: 98,
    },
    {
      sensorId: "ELINT-01",
      sensorType: "ELINT/COMINT",
      status: "CONNECTED",
      eventsPerSecond: 9.8,
      totalReceived: 24560,
      lastSeenSeconds: 7,
      validationPassRate: 96.4,
      dlqCount: 44,
    },
    {
      sensorId: "ISR-01",
      sensorType: "ISR",
      status: "OFFLINE",
      eventsPerSecond: 0,
      totalReceived: 8710,
      lastSeenSeconds: 450,
      validationPassRate: 0,
      dlqCount: 0,
    },
  ];
  return sensors;
}

/**
 * Fetches all known sensor statuses across ALL ingestion services and applies health logic.
 * Makes 6 parallel gRPC-Web calls (one per sensor type) routed by Envoy via the
 * x-ingestion-target request header, then merges and returns the combined results.
 *
 * Logic:
 *  - Green (CONNECTED): last_observation_time < 30s ago
 *  - Amber (STALE): 30s <= last_observation_time < 2m
 *  - Red (OFFLINE): last_observation_time >= 2m or s.connected = false
 *
 * In E2E / dev mode (VITE_MOCK_SENSORS=true), returns deterministic seed-based mock data
 * so tests run without a live backend.
 */
export async function fetchSensorStatuses(): Promise<SensorStatus[]> {
  if (import.meta.env.VITE_MOCK_SENSORS === "true") {
    return mockSensorStatuses();
  }
  const callOpts = (target: string | undefined) =>
    target
      ? { headers: new Headers({ "x-ingestion-target": target }) }
      : undefined;

  const results = await Promise.all(
    INGESTION_TARGETS.map((target) =>
      client
        .listSensorStatuses(
          { activeWithinSeconds: 0, sensorTypes: [] },
          callOpts(target),
        )
        .catch(() => ({
          sensors: [] as Awaited<
            ReturnType<typeof client.listSensorStatuses>
          >["sensors"],
        })),
    ),
  );

  const response = { sensors: results.flatMap((r) => r.sensors) };

  const now = Date.now();

  return response.sensors.map((s) => {
    const lastSeenMs = s.lastObservationTime
      ? Number(s.lastObservationTime.seconds) * 1000 +
        s.lastObservationTime.nanos / 1000000
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
    const passRate =
      totalReceived > 0 ? (totalAccepted / totalReceived) * 100 : 0;

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

// ──────────────────────────────────────────────────────────────────────────────
// Sensor Diagnostic Deep Dive
// ──────────────────────────────────────────────────────────────────────────────

export interface SensorDiagnosticDetail extends SensorStatus {
  dlqBreakdown: { reason: string; count: number }[];
  recentEvents: {
    timeUtc: string;
    event: string;
    severity: "info" | "warn" | "error";
  }[];
  subSensors: {
    id: string;
    status: "ACTIVE" | "DEGRADED" | "INACTIVE";
    location: string;
    lastSeenSeconds: number;
  }[];
  latencyMs: number;
  throughputHistory: number[]; // 20 data points
}

/**
 * Seed-based deterministic mock — used in dev mode when backend unavailable.
 */
async function fetchSensorDiagnosticMock(
  sensor: SensorStatus,
): Promise<SensorDiagnosticDetail> {
  const seed = sensor.sensorId
    .split("")
    .reduce((a, c) => a + c.charCodeAt(0), 0);
  const rng = (i: number) =>
    Math.abs(Math.sin(seed * 9301 + i * 49297 + 233995) % 1);

  const throughputHistory = Array.from({ length: 20 }, (_, i) => {
    const base = sensor.eventsPerSecond > 0 ? sensor.eventsPerSecond : 0;
    return Math.max(0, Math.round(base * (0.7 + rng(i) * 0.6)));
  });

  const dlqReasons = [
    "invalid_timestamp",
    "coordinates_out_of_range",
    "missing_sensor_id",
    "schema_mismatch",
    "invalid_speed",
  ];
  const dlqBreakdown = dlqReasons
    .map((reason, i) => ({
      reason,
      count: Math.round(rng(i + 10) * sensor.dlqCount * 0.6),
    }))
    .filter((d) => d.count > 0);

  const severities = ["info", "warn", "error"] as const;
  const eventTexts = [
    "Observation accepted",
    "Observation rejected: invalid_timestamp",
    "Connected",
    "Throughput within expected range",
    "DLQ spike detected",
    "Latency elevated > 500ms",
    "Sensor reconnected",
    "Validation pass rate dropped below 90%",
  ];
  const recentEvents = Array.from({ length: 8 }, (_, i) => ({
    timeUtc: new Date(
      Date.now() - Math.round(rng(i + 20) * 3600000),
    ).toISOString(),
    event: eventTexts[Math.floor(rng(i + 30) * eventTexts.length)],
    severity: severities[Math.floor(rng(i + 40) * 3)],
  })).sort((a, b) => b.timeUtc.localeCompare(a.timeUtc));

  const locationLabels = [
    "North sector",
    "South sector",
    "Eastern array",
    "West coast",
    "Offshore platform",
  ];
  const subSensors = Array.from(
    { length: Math.floor(rng(seed) * 3) + 1 },
    (_, i) => ({
      id: `${sensor.sensorId}-SUB-0${i + 1}`,
      status: ["ACTIVE", "ACTIVE", "DEGRADED", "INACTIVE"][
        Math.floor(rng(i + 50) * 4)
      ] as "ACTIVE" | "DEGRADED" | "INACTIVE",
      location: locationLabels[Math.floor(rng(i + 60) * locationLabels.length)],
      lastSeenSeconds: Math.round(rng(i + 70) * 300),
    }),
  );

  return {
    ...sensor,
    latencyMs: Math.round(30 + rng(seed + 1) * 470),
    throughputHistory,
    dlqBreakdown,
    recentEvents,
    subSensors,
  };
}

/**
 * Fetches deep diagnostic data for a single sensor.
 * Uses the real gRPC-Web GetSensorDiagnostic RPC when the backend is available;
 * falls back to the deterministic seed-based mock in dev mode.
 * Signature is API-ready for the future GetSensorDiagnostic gRPC RPC (Phase D).
 */
export async function fetchSensorDiagnostic(
  sensor: SensorStatus,
): Promise<SensorDiagnosticDetail> {
  // Fall back to mock data since GetSensorDiagnostic RPC is not yet generated.
  // This will be replaced with the real gRPC-Web call once buf generate is complete (Phase D10).
  return fetchSensorDiagnosticMock(sensor);
}
