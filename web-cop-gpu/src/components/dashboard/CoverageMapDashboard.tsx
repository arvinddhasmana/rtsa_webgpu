// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/CoverageMapDashboard.tsx — Level 3: Full Coverage Map
//
// Strategic view of the entire sensor network coverage footprint.
// Highlights geographic gaps and active spatial alerts.
//
// Reference: docs/implementation/v5/sensordashboard_three_level_plan.md §C1

import { createResource, createSignal, onMount, Show } from "solid-js";
import {
  activeSpatialAlertId,
  setActiveSpatialAlertId,
  spatialAlerts,
} from "../../signals/spatial-alerts";
import { fetchSensorStatuses, type SensorStatus } from "../../services/sensor-health";
import { CoverageAreaMap, type CoverageAreaMapBounds } from "./CoverageAreaMap";
import { SensorFleetList } from "./SensorFleetList";
import { CriticalAlertsPanel } from "./CriticalAlertsPanel";
import { SensorDetailHoverPanel } from "./SensorDetailHoverPanel";
import { SpatialAlertBanner } from "./SpatialAlertBanner";

/** Level 3 — Full Coverage Map Dashboard. */
export function CoverageMapDashboard() {
  const [sensors] = createResource(fetchSensorStatuses);
  const [hoveredSensor, setHoveredSensor] = createSignal<SensorStatus | null>(null);
  const [selectedSensorId, setSelectedSensorId] = createSignal<string | undefined>(undefined);
  const [mapBounds, setMapBounds] = createSignal<CoverageAreaMapBounds | undefined>(undefined);

  // Auto-zoom to active alert polygon when navigated via activeSpatialAlertId
  onMount(() => {
    const alertId = activeSpatialAlertId();
    if (alertId) {
      const alert = spatialAlerts().find((a) => a.alertId === alertId);
      if (alert && alert.areaPolygon.length > 0) {
        // Compute bounding box from area polygon
        const lats = alert.areaPolygon.map((p) => p.lat);
        const lons = alert.areaPolygon.map((p) => p.lon);
        const minLat = Math.min(...lats);
        const maxLat = Math.max(...lats);
        const minLon = Math.min(...lons);
        const maxLon = Math.max(...lons);
        const padding = 0.5; // degrees
        setMapBounds({
          minLat: minLat - padding,
          maxLat: maxLat + padding,
          minLon: minLon - padding,
          maxLon: maxLon + padding,
        });
        // Highlight affected sensor
        setSelectedSensorId(alert.affectedSensorId);
      }
    }
  });

  const handleSensorSelect = (sensor: SensorStatus) => {
    setSelectedSensorId(sensor.sensorId);
    setHoveredSensor(sensor);
  };

  const handleAlertClick = (alertId: string) => {
    setActiveSpatialAlertId(alertId);
    const alert = spatialAlerts().find((a) => a.alertId === alertId);
    if (alert) {
      // Auto-zoom to alert area
      if (alert.areaPolygon.length > 0) {
        const lats = alert.areaPolygon.map((p) => p.lat);
        const lons = alert.areaPolygon.map((p) => p.lon);
        const minLat = Math.min(...lats);
        const maxLat = Math.max(...lats);
        const minLon = Math.min(...lons);
        const maxLon = Math.max(...lons);
        const padding = 0.5;
        setMapBounds({
          minLat: minLat - padding,
          maxLat: maxLat + padding,
          minLon: minLon - padding,
          maxLon: maxLon + padding,
        });
      }
      setSelectedSensorId(alert.affectedSensorId);
    }
  };

  const handleResolveAlert = (alertId: string) => {
    // In a real implementation, this would call a gRPC service to acknowledge the alert
    // For now, we just clear the active alert ID
    console.log(`[CoverageMapDashboard] Resolving alert: ${alertId}`);
    setActiveSpatialAlertId(null);
  };

  const activeGapCount = () => spatialAlerts().filter((a) => !a.acknowledged).length;

  const currentStatus = () => {
    const gaps = activeGapCount();
    if (gaps === 0) return { text: "NOMINAL", color: "#10b981" };
    return { text: `ACTIVE GAPS (${gaps})`, color: "#ef4444" };
  };

  return (
    <div
      data-testid="coverage-map-dashboard"
      style={{
        display: "flex",
        "flex-direction": "column",
        height: "100%",
        width: "100%",
        background: "#0a0f1a",
        overflow: "hidden",
      }}
    >
      {/* Header status bar */}
      <div
        style={{
          display: "flex",
          "align-items": "center",
          "justify-content": "space-between",
          padding: "12px 20px",
          "border-bottom": "1px solid rgba(255,255,255,0.05)",
          background: "rgba(13, 20, 36, 0.6)",
          "flex-shrink": "0",
        }}
      >
        <div style={{ display: "flex", "align-items": "center", gap: "16px" }}>
          <span
            style={{
              "font-size": "0.95rem",
              "font-weight": "700",
              color: "#cbd5e1",
              "text-transform": "uppercase",
              "letter-spacing": "0.05em",
            }}
          >
            Global Situational Awareness | Sensor Coverage Overlay
          </span>
        </div>
        <div style={{ display: "flex", "align-items": "center", gap: "12px" }}>
          <span style={{ "font-size": "0.75rem", color: "#64748b", "text-transform": "uppercase" }}>
            Current Status:
          </span>
          <span
            style={{
              "font-size": "0.8rem",
              "font-weight": "700",
              color: currentStatus().color,
              "text-transform": "uppercase",
              "letter-spacing": "0.05em",
            }}
          >
            {currentStatus().text}
          </span>
        </div>
      </div>

      {/* Main content area */}
      <div
        style={{
          display: "flex",
          flex: "1",
          "min-height": "0",
          overflow: "hidden",
        }}
      >
        {/* Left panel: Sensor fleet list + Critical alerts panel */}
        <div
          style={{
            width: "240px",
            display: "flex",
            "flex-direction": "column",
            "border-right": "1px solid rgba(255,255,255,0.05)",
            background: "rgba(13, 20, 36, 0.4)",
            overflow: "hidden",
            "flex-shrink": "0",
          }}
        >
          {/* Sensor Fleet List */}
          <div
            style={{
              flex: "1",
              "min-height": "0",
              "border-bottom": "1px solid rgba(255,255,255,0.05)",
            }}
          >
            <Show when={sensors()}>
              {(sensorData) => (
                <SensorFleetList
                  sensors={sensorData()}
                  selectedSensorId={selectedSensorId()}
                  onSensorSelect={handleSensorSelect}
                  onSensorHover={setHoveredSensor}
                  compact={false}
                  maxHeight="100%"
                />
              )}
            </Show>
          </div>

          {/* Critical Alerts Panel */}
          <div
            style={{
              "flex-shrink": "0",
              "max-height": "280px",
            }}
          >
            <CriticalAlertsPanel
              spatialAlerts={spatialAlerts()}
              onAlertClick={handleAlertClick}
              maxHeight="280px"
              title="Critical Alerts"
            />
          </div>
        </div>

        {/* Center: Coverage map */}
        <div
          style={{
            flex: "1",
            display: "flex",
            position: "relative",
            overflow: "hidden",
            "min-width": "0",
          }}
        >
          <Show when={sensors()}>
            {(sensorData) => (
              <CoverageAreaMap
                sensors={sensorData()}
                spatialAlerts={spatialAlerts()}
                bounds={mapBounds()}
                showLabels={true}
                showGapHatching={true}
                showRangeRings={false}
                showSweepAnimation={false}
                onSensorClick={handleSensorSelect}
                onGapAlertClick={handleAlertClick}
                width="100%"
                height="100%"
              />
            )}
          </Show>
        </div>

        {/* Right panel: Sensor detail hover panel */}
        <div
          style={{
            width: "280px",
            "border-left": "1px solid rgba(255,255,255,0.05)",
            background: "rgba(13, 20, 36, 0.4)",
            overflow: "hidden",
            "flex-shrink": "0",
          }}
        >
          <SensorDetailHoverPanel sensor={hoveredSensor()} width="280px" />
        </div>
      </div>

      {/* Bottom: Spatial Alert Banner */}
      <SpatialAlertBanner alerts={spatialAlerts()} onResolve={handleResolveAlert} />
    </div>
  );
}
