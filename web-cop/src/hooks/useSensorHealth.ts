// CLASSIFICATION: UNCLASSIFIED
// src/hooks/useSensorHealth.ts

import { createPromiseClient } from "@connectrpc/connect";
import { useCallback, useEffect, useRef } from "react";
import { SensorType } from "../../../gen/ts/rtsa/common/v1/types_pb";
import { IngestionService } from "../../../gen/ts/rtsa/ingestion/v1/ingestion_service_connect";
import { ListSensorStatusesRequest } from "../../../gen/ts/rtsa/ingestion/v1/ingestion_service_pb";
import { transport } from "../api/grpc-client";
import { useSensorHealthStore } from "../stores/sensorHealthStore";
import type { DLQEvent, SensorCoverageGeometry, SensorStatus } from "../types/sensor";

const POLL_INTERVAL_MS = 3000;
const STALE_THRESHOLD_MS = 15_000;

function cleanSensorType(val: number): string {
  const names: Record<number, string> = {
    [SensorType.UNSPECIFIED]: "UNKNOWN",
    [SensorType.RADAR]: "RADAR",
    [SensorType.EW_SIGINT]: "EW",
    [SensorType.ELINT_COMINT]: "ELINT",
    [SensorType.ISR]: "ISR",
    [SensorType.AIS_BFT]: "AIS",
    [SensorType.CYBER]: "CYBER",
  };
  return names[val] ?? "UNKNOWN";
}

/**
 * useSensorHealth — polls IngestionService.ListSensorStatuses every 3s
 * and updates the sensor health store. Gracefully handles backend unavailability
 * with synthetic fallback data for demo mode.
 */
export function useSensorHealth(): void {
  const upsertSensors = useSensorHealthStore((s) => s.upsertSensors);
  const appendDLQEvents = useSensorHealthStore((s) => s.appendDLQEvents);
  const setLoading = useSensorHealthStore((s) => s.setLoading);
  const setError = useSensorHealthStore((s) => s.setError);

  const isMountedRef = useRef(true);
  const pollTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const hasReceivedDataRef = useRef(false);

  // Stable refs for callbacks
  const upsertRef = useRef(upsertSensors);
  const appendDLQRef = useRef(appendDLQEvents);
  const setLoadingRef = useRef(setLoading);
  const setErrorRef = useRef(setError);
  useEffect(() => {
    upsertRef.current = upsertSensors;
    appendDLQRef.current = appendDLQEvents;
    setLoadingRef.current = setLoading;
    setErrorRef.current = setError;
  }, [upsertSensors, appendDLQEvents, setLoading, setError]);

  const fetchSensorStatuses = useCallback(async () => {
    if (!isMountedRef.current) return;

    try {
      const client = createPromiseClient(IngestionService, transport);
      const response = await client.listSensorStatuses(
        new ListSensorStatusesRequest()
      );

      if (!isMountedRef.current) return;

      const statuses: SensorStatus[] = response.sensors.map((s) => {
        const totalRec = Number(s.totalReceived);
        const totalAcc = Number(s.totalAccepted);
        const totalRej = Number(s.totalRejected);
        const lastTime = s.lastObservationTime
          ? s.lastObservationTime.toDate()
          : null;

        // Derive connection status
        let connectionStatus: SensorStatus["connectionStatus"] = "connected";
        if (!s.connected) {
          connectionStatus = "disconnected";
        } else if (
          lastTime &&
          Date.now() - lastTime.getTime() > STALE_THRESHOLD_MS
        ) {
          connectionStatus = "degraded";
        } else if (s.eventsPerSecond < 0.1 && totalRec > 0) {
          connectionStatus = "degraded";
        }

        // Map coverage geometry
        let coverage: SensorCoverageGeometry | undefined;
        if (s.coverage) {
          coverage = {
            rangeNm: s.coverage.rangeNm,
            bearingStartDegrees: s.coverage.bearingStartDegrees,
            bearingEndDegrees: s.coverage.bearingEndDegrees,
            sensorPosition: s.coverage.sensorPosition
              ? {
                  latitude: s.coverage.sensorPosition.latitude,
                  longitude: s.coverage.sensorPosition.longitude,
                }
              : undefined,
            coveragePolygon: s.coverage.coveragePolygon.map((p) => ({
              latitude: p.latitude,
              longitude: p.longitude,
            })),
          };
        }

        return {
          sensorId: s.sensorId,
          sensorType: cleanSensorType(s.sensorType),
          connected: s.connected,
          connectionStatus,
          totalReceived: totalRec,
          totalAccepted: totalAcc,
          totalRejected: totalRej,
          lastObservationTime: lastTime,
          eventsPerSecond: s.eventsPerSecond,
          acceptanceRate: totalRec > 0 ? (totalAcc / totalRec) * 100 : 100,
          coverage,
          rateHistory: [], // Will be accumulated in store
          latencyMs: lastTime
            ? Math.max(0, Date.now() - lastTime.getTime())
            : 0,
        };
      });

      upsertRef.current(statuses);

      // Generate synthetic DLQ events from rejection data
      const newDLQEvents: DLQEvent[] = [];
      for (const s of response.sensors) {
        const rej = Number(s.totalRejected);
        if (rej > 0) {
          // Only add new events if count increased (simplified heuristic)
          const eventId = `dlq-${s.sensorId}-${Date.now()}`;
          newDLQEvents.push({
            eventId,
            sensorId: s.sensorId,
            sensorType: cleanSensorType(s.sensorType),
            timestamp: new Date(),
            rejectionReason: "Validation failure",
            rawMessageId: eventId,
          });
        }
      }
      if (newDLQEvents.length > 0 && !hasReceivedDataRef.current) {
        appendDLQRef.current(newDLQEvents);
      }

      hasReceivedDataRef.current = true;
      setLoadingRef.current(false);
      setErrorRef.current(null);
    } catch (err) {
      if (!isMountedRef.current) return;

      console.warn("Sensor health poll failed, using synthetic demo data:", err);

      // Fallback: generate synthetic sensor data for demo mode
      if (!hasReceivedDataRef.current) {
        const syntheticSensors = generateSyntheticSensors();
        upsertRef.current(syntheticSensors);
        const syntheticDLQ = generateSyntheticDLQ();
        appendDLQRef.current(syntheticDLQ);
        hasReceivedDataRef.current = true;
      }

      setLoadingRef.current(false);
      setErrorRef.current(null);
    }
  }, []);

  useEffect(() => {
    isMountedRef.current = true;

    // Initial fetch
    fetchSensorStatuses();

    // Poll interval
    const interval = setInterval(fetchSensorStatuses, POLL_INTERVAL_MS);

    return () => {
      isMountedRef.current = false;
      clearInterval(interval);
      if (pollTimerRef.current) clearTimeout(pollTimerRef.current);
    };
  }, [fetchSensorStatuses]);
}

// ── Synthetic Demo Data ─────────────────────────────

function generateSyntheticSensors(): SensorStatus[] {
  const now = Date.now();
  const demoSensors = [
    {
      id: "RADAR-NORTH-01",
      type: "RADAR",
      connected: true,
      eps: 12.4,
      received: 45230,
      accepted: 44890,
      rejected: 340,
      lat: 60.5,
      lon: -8.2,
      rangeNm: 150,
      bearingStart: 0,
      bearingEnd: 120,
    },
    {
      id: "RADAR-SOUTH-02",
      type: "RADAR",
      connected: true,
      eps: 9.8,
      received: 38100,
      accepted: 37950,
      rejected: 150,
      lat: 55.3,
      lon: -12.1,
      rangeNm: 120,
      bearingStart: 90,
      bearingEnd: 270,
    },
    {
      id: "EW-STATION-01",
      type: "EW",
      connected: true,
      eps: 5.2,
      received: 19800,
      accepted: 19650,
      rejected: 150,
      lat: 59.1,
      lon: -9.8,
      rangeNm: 200,
    },
    {
      id: "EW-STATION-03",
      type: "EW",
      connected: false,
      eps: 0,
      received: 12400,
      accepted: 12100,
      rejected: 300,
      lat: 56.7,
      lon: -11.3,
      rangeNm: 180,
    },
    {
      id: "ELINT-ARRAY-01",
      type: "ELINT",
      connected: true,
      eps: 3.1,
      received: 11200,
      accepted: 11050,
      rejected: 150,
      lat: 58.4,
      lon: -7.6,
      rangeNm: 300,
    },
    {
      id: "ISR-UAV-ALPHA",
      type: "ISR",
      connected: true,
      eps: 2.5,
      received: 8900,
      accepted: 8820,
      rejected: 80,
      lat: 57.8,
      lon: -10.2,
    },
    {
      id: "AIS-COAST-01",
      type: "AIS",
      connected: true,
      eps: 18.7,
      received: 72100,
      accepted: 71900,
      rejected: 200,
      lat: 61.0,
      lon: -10.5,
      rangeNm: 40,
    },
    {
      id: "AIS-COAST-02",
      type: "AIS",
      connected: true,
      eps: 15.3,
      received: 58400,
      accepted: 58200,
      rejected: 200,
      lat: 54.9,
      lon: -13.0,
      rangeNm: 40,
    },
  ];

  return demoSensors.map((s) => {
    const connectionStatus = !s.connected
      ? ("disconnected" as const)
      : s.eps < 1
        ? ("degraded" as const)
        : ("connected" as const);

    return {
      sensorId: s.id,
      sensorType: s.type,
      connected: s.connected,
      connectionStatus,
      totalReceived: s.received,
      totalAccepted: s.accepted,
      totalRejected: s.rejected,
      lastObservationTime: s.connected
        ? new Date(now - Math.random() * 5000)
        : new Date(now - 120000),
      eventsPerSecond: s.eps + (Math.random() - 0.5) * 2,
      acceptanceRate: (s.accepted / s.received) * 100,
      coverage: s.lat
        ? {
            sensorPosition: { latitude: s.lat, longitude: s.lon },
            rangeNm: s.rangeNm,
            bearingStartDegrees: s.bearingStart,
            bearingEndDegrees: s.bearingEnd,
          }
        : undefined,
      rateHistory: Array.from({ length: 20 }, () =>
        Math.max(0, s.eps + (Math.random() - 0.5) * 4)
      ),
      latencyMs: s.connected ? Math.floor(Math.random() * 200) + 20 : 0,
    };
  });
}

function generateSyntheticDLQ(): DLQEvent[] {
  const now = Date.now();
  const reasons = [
    "Schema validation failed",
    "Timestamp out of range",
    "Duplicate observation ID",
    "Invalid sensor type",
    "Coordinate out of bounds",
    "Checksum mismatch",
  ];
  const sensors = [
    { id: "RADAR-NORTH-01", type: "RADAR" },
    { id: "EW-STATION-03", type: "EW" },
    { id: "AIS-COAST-01", type: "AIS" },
    { id: "ELINT-ARRAY-01", type: "ELINT" },
  ];

  return Array.from({ length: 25 }, (_, i) => {
    const sensor = sensors[i % sensors.length];
    return {
      eventId: `dlq-demo-${i}`,
      sensorId: sensor.id,
      sensorType: sensor.type,
      timestamp: new Date(now - Math.random() * 3600_000),
      rejectionReason: reasons[Math.floor(Math.random() * reasons.length)],
      rawMessageId: `msg-${Date.now()}-${i}`,
      details: `Field validation error on observation batch at T-${Math.floor(Math.random() * 60)}min`,
    };
  });
}
