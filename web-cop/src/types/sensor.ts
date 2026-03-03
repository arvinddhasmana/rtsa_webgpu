// CLASSIFICATION: UNCLASSIFIED
// src/types/sensor.ts

/**
 * Sensor Health types for the Sensor Operator dashboard.
 * Maps to proto: rtsa.ingestion.v1.SensorStatusResponse
 */

export type SensorConnectionStatus = 'connected' | 'degraded' | 'disconnected';

export interface SensorCoverageGeometry {
  /** Coverage polygon vertices (for ISR, geo-fence type sensors) */
  coveragePolygon?: { latitude: number; longitude: number }[];
  /** Maximum sensor range in nautical miles (for radar, EW) */
  rangeNm?: number;
  /** Bearing sector start in degrees true (for directional sensors) */
  bearingStartDegrees?: number;
  /** Bearing sector end in degrees true (for directional sensors) */
  bearingEndDegrees?: number;
  /** Sensor geographic position */
  sensorPosition?: { latitude: number; longitude: number };
}

export interface SensorStatus {
  sensorId: string;
  sensorType: string;
  connected: boolean;
  connectionStatus: SensorConnectionStatus;
  totalReceived: number;
  totalAccepted: number;
  totalRejected: number;
  lastObservationTime: Date | null;
  eventsPerSecond: number;
  /** Acceptance rate as 0–100 percentage */
  acceptanceRate: number;
  /** Sensor coverage geometry for map rendering */
  coverage?: SensorCoverageGeometry;
  /** Rolling history of events/sec for sparkline rendering (last 30 points) */
  rateHistory: number[];
  /** Estimated latency in ms */
  latencyMs: number;
}

export interface DLQEvent {
  eventId: string;
  sensorId: string;
  sensorType: string;
  timestamp: Date;
  rejectionReason: string;
  rawMessageId?: string;
  details?: string;
}

export type DLQPattern = 'isolated' | 'burst' | 'sustained';

export interface DLQSummary {
  totalCount: number;
  bySensor: Record<string, number>;
  byReason: Record<string, number>;
  pattern: DLQPattern;
}

export type SortField = 'sensorId' | 'sensorType' | 'eventsPerSecond' | 'totalRejected' | 'latencyMs' | 'lastObservationTime' | 'connectionStatus';
export type SortDirection = 'asc' | 'desc';
