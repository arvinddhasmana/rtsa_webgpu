// CLASSIFICATION: UNCLASSIFIED
// src/signals/spatial-alerts.ts — Spatial (coverage-gap) alert signals
//
// Spatial alerts are emitted by the Coverage Analyzer service when sensor
// footprint unions leave an uncovered geographic area. They are distinct from
// point-track alerts and drive Level 3 (Coverage Map) navigation.
//
// Reference: docs/implementation/v5/sensordashboard_three_level_plan.md §A2

import { createSignal } from "solid-js";

/** A geographic coverage-gap alert produced by the Coverage Analyzer. */
export interface SpatialAlertPayload {
  /** Unique alert identifier (UUID or "gap-<timestamp>"). */
  alertId: string;
  /** Operational sector identifier, e.g. "NW-4". */
  sectorId: string;
  /** ID of the sensor that went offline / stale, causing the gap. */
  affectedSensorId: string;
  /** Operational urgency level. */
  severity: "CRITICAL" | "ELEVATED" | "WATCH";
  /** Human-readable summary, e.g. "Data gap in sector NW-4". */
  description: string;
  /** ISO 8601 timestamp of the last known good observation. */
  lastContactUtc: string;
  /** True once an operator has acknowledged this alert. */
  acknowledged: boolean;
  /** Polygon bounding the uncovered area (ordered lat/lon vertices). */
  areaPolygon: Array<{ lat: number; lon: number }>;
}

/** Live list of spatial gap alerts, newest first. */
export const [spatialAlerts, setSpatialAlerts] = createSignal<SpatialAlertPayload[]>([]);

/**
 * The alertId of the spatial alert the operator is currently inspecting,
 * or null when no alert is focused. Setting this and `setDashboard("coverage")`
 * navigates the UI to Level 3 and highlights the affected sector.
 */
export const [activeSpatialAlertId, setActiveSpatialAlertId] = createSignal<string | null>(null);
