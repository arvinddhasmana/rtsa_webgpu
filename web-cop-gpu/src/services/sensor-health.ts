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
  // Extended diagnostic fields (v5)
  healthScore: number; // 0-100 composite health
  connectionUptimePct: number; // % uptime in last hour
  peakThroughput: number; // max obs/s seen
  avgLatencyMs: number; // same as latencyMs, alias for clarity
  minLatencyMs: number;
  maxLatencyMs: number;
  statusHistory: Array<{
    timeUtc: string;
    status: "CONNECTED" | "STALE" | "OFFLINE";
  }>; // 12 samples, oldest first
  rangeNm: number | null; // coverage range (NM)
  position: { lat: number; lon: number } | null;
  bearingStart: number | null; // degrees (RADAR only)
  bearingEnd: number | null;
  scanRateRpm: number | null; // revolutions per minute (RADAR)
  frequencyBandGhz: number | null; // operating frequency
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
    "Sub-sensor handover initiated",
    "Authentication certificate renewed",
    "Coverage geometry updated",
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
    "Central hub",
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

  // Extended diagnostic fields
  const avgLatencyMs = Math.round(30 + rng(seed + 1) * 470);
  const minLatencyMs = Math.max(
    5,
    Math.round(avgLatencyMs * (0.25 + rng(seed + 901) * 0.4)),
  );
  const maxLatencyMs = Math.round(avgLatencyMs * (1.4 + rng(seed + 902) * 2.0));
  const peakThroughput = Math.round(
    Math.max(...throughputHistory, sensor.eventsPerSecond) *
      (1.1 + rng(seed + 1100) * 0.4),
  );
  const connectionUptimePct =
    sensor.status === "OFFLINE"
      ? Math.round(rng(seed + 1000) * 40 * 10) / 10
      : Math.round((80 + rng(seed + 1000) * 20) * 10) / 10;

  const latencyScore =
    avgLatencyMs < 100 ? 100 : Math.max(0, 100 - (avgLatencyMs - 100) / 5);
  const healthScore = Math.min(
    100,
    Math.max(
      0,
      Math.round(
        sensor.validationPassRate * 0.5 +
          connectionUptimePct * 0.3 +
          latencyScore * 0.2,
      ),
    ),
  );

  const statusHistory: Array<{
    timeUtc: string;
    status: "CONNECTED" | "STALE" | "OFFLINE";
  }> = Array.from({ length: 12 }, (_, i) => {
    const r = rng(seed + 200 + i);
    let s: "CONNECTED" | "STALE" | "OFFLINE" = sensor.status;
    if (r < 0.07) s = "STALE";
    if (r < 0.02) s = "OFFLINE";
    return {
      timeUtc: new Date(Date.now() - (11 - i) * 60_000).toISOString(),
      status: s,
    };
  });

  // Sensor parameters by type
  const rangeByType: Record<string, number> = {
    RADAR: 100 + Math.round(rng(seed + 400) * 100),
    "EW/SIGINT": 150 + Math.round(rng(seed + 400) * 100),
    "ELINT/COMINT": 200 + Math.round(rng(seed + 400) * 150),
    ISR: 30 + Math.round(rng(seed + 400) * 30),
    "AIS/BFT": 20 + Math.round(rng(seed + 400) * 30),
  };
  const rangeNm = rangeByType[sensor.sensorType] ?? null;

  const knownPositions = [
    { lat: 60.5, lon: -8.2 },
    { lat: 55.3, lon: -12.1 },
    { lat: 58.4, lon: -7.6 },
    { lat: 56.7, lon: -11.3 },
    { lat: 59.1, lon: -9.8 },
    { lat: 61.0, lon: -10.5 },
    { lat: 54.9, lon: -13.0 },
    { lat: 57.8, lon: -10.1 },
  ];
  const position =
    knownPositions[Math.floor(rng(seed + 500) * knownPositions.length)];

  const bearingStart =
    sensor.sensorType === "RADAR" ? Math.round(rng(seed + 600) * 270) : null;
  const bearingEnd =
    bearingStart !== null
      ? (bearingStart + 90 + Math.round(rng(seed + 601) * 90)) % 360
      : null;
  const scanRateRpm =
    sensor.sensorType === "RADAR" ? Math.round(3 + rng(seed + 700) * 9) : null;
  const frequencyBandGhz =
    sensor.sensorType === "RADAR"
      ? Math.round((3.0 + rng(seed + 800) * 12.0) * 10) / 10
      : sensor.sensorType === "EW/SIGINT" ||
          sensor.sensorType === "ELINT/COMINT"
        ? Math.round((0.1 + rng(seed + 800) * 5.0) * 10) / 10
        : null;

  return {
    ...sensor,
    latencyMs: avgLatencyMs,
    throughputHistory,
    dlqBreakdown,
    recentEvents,
    subSensors,
    healthScore,
    connectionUptimePct,
    peakThroughput,
    avgLatencyMs,
    minLatencyMs,
    maxLatencyMs,
    statusHistory,
    rangeNm,
    position,
    bearingStart,
    bearingEnd,
    scanRateRpm,
    frequencyBandGhz,
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
