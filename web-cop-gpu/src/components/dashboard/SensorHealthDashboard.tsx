// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/SensorHealthDashboard.tsx — Main Health Dashboard
//
// Reference: docs/business/usecases/UC017_sensor_health_monitoring.md

import { createEffect, createResource, createSignal, onCleanup, Show } from "solid-js";
import { fetchSensorStatuses } from "../../services/sensor-health";
import {
    cardView,
    selectedSensor,
    setCardView,
    setSelectedSensor,
} from "../../signals/sensor-filters";
import { spatialAlerts } from "../../signals/spatial-alerts";
import { dashboard } from "../../signals/viewport";
import { CriticalAlertsPanel } from "./CriticalAlertsPanel";
import { DashboardSidebar } from "./DashboardSidebar";
import { SensorDetailHoverPanel } from "./SensorDetailHoverPanel";
import { SensorDiagnosticView } from "./SensorDiagnosticView";
import { SensorFleetList } from "./SensorFleetList";
import { SensorGrid } from "./SensorGrid";
import { SensorOverviewMap } from "./SensorOverviewMap";

/**
 * Sensor Health Monitoring Dashboard.
 * Orchestrates data fetching, top-level layout, and filtering.
 */
export function SensorHealthDashboard() {
  const [sensors, { refetch }] = createResource(fetchSensorStatuses);
  const [hoveredSensorId, setHoveredSensorId] = createSignal<string | undefined>(undefined);
  const [hoveredSensor, setHoveredSensor] = createSignal<Parameters<typeof SensorDetailHoverPanel>[0]["sensor"]>(null);

  // Auto-refresh every 10 seconds as per requirements
  const timer = setInterval(refetch, 10000);
  onCleanup(() => clearInterval(timer));

  // Clear the selected sensor when dashboard changes
  createEffect(() => {
    void dashboard();
    setSelectedSensor(null);
  });

  return (
    <div
      style={{
        display: "flex",
        height: "100%",
        width: "100%",
        background:
          "radial-gradient(circle at 0% 0%, rgba(30, 58, 138, 0.15) 0%, transparent 50%), radial-gradient(circle at 100% 100%, rgba(88, 28, 135, 0.15) 0%, transparent 50%)",
        overflow: "hidden",
      }}
    >
      <DashboardSidebar sensors={sensors() || []} />

      <div
        style={{
          flex: 1,
          display: "flex",
          "flex-direction": "column",
          overflow: "hidden",
        }}
      >
        {/* Header / Search bar local to health dashboard could go here */}
        <div
          style={{
            padding: "16px 24px",
            "border-bottom": "1px solid rgba(255,255,255,0.05)",
            background: "rgba(13, 20, 36, 0.4)",
            display: "flex",
            "justify-content": "space-between",
            "align-items": "center",
          }}
        >
          <h2
            style={{
              "font-size": "1.25rem",
              "font-weight": "600",
              color: "#f8fafc",
              margin: 0,
            }}
          >
            Sensor Health Monitor
          </h2>
          <div
            style={{ display: "flex", gap: "12px", "align-items": "center" }}
          >
            {/* View mode toggle */}
            <div
              data-testid="view-toggle"
              style={{
                display: "flex",
                gap: "2px",
                background: "rgba(255,255,255,0.05)",
                padding: "3px",
                "border-radius": "8px",
              }}
            >
              <button
                data-testid="view-toggle-full"
                title="Full card view — dual sparklines with Type &amp; Location details"
                onClick={() => setCardView("full")}
                style={{
                  background:
                    cardView() === "full"
                      ? "rgba(59,130,246,0.25)"
                      : "transparent",
                  border:
                    cardView() === "full"
                      ? "1px solid rgba(59,130,246,0.5)"
                      : "1px solid transparent",
                  color: cardView() === "full" ? "#60a5fa" : "#64748b",
                  padding: "4px 10px",
                  "border-radius": "6px",
                  cursor: "pointer",
                  "font-size": "0.7rem",
                  "font-weight": "600",
                  display: "flex",
                  "align-items": "center",
                  gap: "5px",
                  transition: "all 0.15s ease",
                }}
              >
                {/* 2×2 grid icon */}
                <svg
                  width="12"
                  height="12"
                  viewBox="0 0 12 12"
                  fill="currentColor"
                >
                  <rect x="0" y="0" width="5" height="5" rx="1" />
                  <rect x="7" y="0" width="5" height="5" rx="1" />
                  <rect x="0" y="7" width="5" height="5" rx="1" />
                  <rect x="7" y="7" width="5" height="5" rx="1" />
                </svg>
                Full
              </button>
              <button
                data-testid="view-toggle-compact"
                title="Compact card view — condensed metrics"
                onClick={() => setCardView("compact")}
                style={{
                  background:
                    cardView() === "compact"
                      ? "rgba(59,130,246,0.25)"
                      : "transparent",
                  border:
                    cardView() === "compact"
                      ? "1px solid rgba(59,130,246,0.5)"
                      : "1px solid transparent",
                  color: cardView() === "compact" ? "#60a5fa" : "#64748b",
                  padding: "4px 10px",
                  "border-radius": "6px",
                  cursor: "pointer",
                  "font-size": "0.7rem",
                  "font-weight": "600",
                  display: "flex",
                  "align-items": "center",
                  gap: "5px",
                  transition: "all 0.15s ease",
                }}
              >
                {/* horizontal lines icon */}
                <svg
                  width="12"
                  height="12"
                  viewBox="0 0 12 12"
                  fill="currentColor"
                >
                  <rect x="0" y="1" width="12" height="2" rx="1" />
                  <rect x="0" y="5" width="12" height="2" rx="1" />
                  <rect x="0" y="9" width="12" height="2" rx="1" />
                </svg>
                Compact
              </button>
            </div>
            <div style={{ "font-size": "0.75rem", color: "#94a3b8" }}>
              {sensors.loading
                ? "Updating..."
                : `Last updated: ${new Date().toLocaleTimeString()}`}
            </div>
            <button
              onClick={() => refetch()}
              style={{
                background: "rgba(59, 130, 246, 0.1)",
                border: "1px solid rgba(59, 130, 246, 0.2)",
                color: "#60a5fa",
                padding: "4px 12px",
                "border-radius": "4px",
                cursor: "pointer",
                "font-size": "0.75rem",
              }}
              class="refresh-btn"
            >
              Refresh
            </button>
          </div>
        </div>

        <Show
          when={!sensors.error}
          fallback={
            <div style={{ padding: "40px", color: "#f87171" }}>
              Error loading sensor data: {sensors.error?.message}
            </div>
          }
        >
          <Show
            when={selectedSensor() !== null}
            fallback={
              <div style={{ display: "flex", flex: 1, overflow: "hidden" }}>
                <SensorGrid
                  sensors={sensors() || []}
                  cardView={cardView()}
                  onSensorSelect={setSelectedSensor}
                />

                {/* Level 1: Upgraded right sidebar — fleet list + map + detail panel */}
                <div
                  style={{
                    width: "480px",
                    "flex-shrink": 0,
                    "border-left": "1px solid rgba(255,255,255,0.05)",
                    background: "rgba(13, 20, 36, 0.2)",
                    display: "flex",
                    "flex-direction": "column",
                    overflow: "hidden",
                  }}
                >
                  {/* Top: fleet list + map side by side */}
                  <div style={{ display: "flex", flex: 1, overflow: "hidden", "min-height": 0 }}>
                    {/* Left: fleet list + critical alerts */}
                    <div style={{
                      width: "200px",
                      "flex-shrink": 0,
                      display: "flex",
                      "flex-direction": "column",
                      "border-right": "1px solid rgba(255,255,255,0.05)",
                      overflow: "hidden",
                    }}>
                      {/* Fleet list header */}
                      <div style={{
                        padding: "10px 12px 6px",
                        "font-size": "0.6rem",
                        "text-transform": "uppercase",
                        "letter-spacing": "0.1em",
                        color: "#475569",
                        "font-family": "monospace",
                        "border-bottom": "1px solid rgba(255,255,255,0.04)",
                      }}>
                        Sensor Fleet
                      </div>
                      {/* Fleet list */}
                      <div style={{ flex: 1, "overflow-y": "auto", padding: "6px 8px" }}>
                        <SensorFleetList
                          sensors={sensors() || []}
                          selectedSensorId={hoveredSensorId()}
                          onSensorSelect={(s) => {
                            setHoveredSensorId(s.sensorId);
                            setHoveredSensor(s);
                          }}
                          onSensorHover={(s) => {
                            setHoveredSensorId(s?.sensorId);
                            setHoveredSensor(s ?? null);
                          }}
                          maxHeight="none"
                        />
                      </div>

                      {/* Critical alerts */}
                      <div style={{
                        "border-top": "1px solid rgba(255,255,255,0.05)",
                        padding: "10px 10px 12px",
                        "overflow-y": "auto",
                        "max-height": "180px",
                      }}>
                        <CriticalAlertsPanel
                          spatialAlerts={spatialAlerts()}
                          maxHeight="160px"
                        />
                      </div>
                    </div>

                    {/* Right: overview map + detail panel */}
                    <div style={{ flex: 1, display: "flex", "flex-direction": "column", overflow: "hidden", padding: "10px", gap: "10px" }}>
                      {/* Overview map */}
                      <SensorOverviewMap
                        sensors={sensors() || []}
                        spatialAlerts={spatialAlerts()}
                        hoveredSensorId={hoveredSensorId()}
                        onSensorClick={(s) => {
                          setHoveredSensorId(s.sensorId);
                          setHoveredSensor(s);
                          setSelectedSensor(s);
                        }}
                        width={260}
                        height={220}
                      />

                      {/* Sensor detail hover panel */}
                      <Show when={hoveredSensor() !== null}>
                        <div style={{ flex: 1, "overflow-y": "auto" }}>
                          <SensorDetailHoverPanel
                            sensor={hoveredSensor()}
                            width="100%"
                          />
                        </div>
                      </Show>

                      {/* Coverage statistics (when no sensor hovered) */}
                      <Show when={hoveredSensor() === null}>
                        <div style={{
                          padding: "12px 14px",
                          background: "rgba(255,255,255,0.02)",
                          "border-radius": "8px",
                          border: "1px solid rgba(255,255,255,0.05)",
                        }}>
                          <div style={{ "font-size": "0.65rem", color: "#475569", "text-transform": "uppercase", "letter-spacing": "0.08em", "margin-bottom": "10px", "font-family": "monospace" }}>
                            Fleet Coverage Health
                          </div>
                          <div style={{ display: "flex", "flex-direction": "column", gap: "8px" }}>
                            {(() => {
                              const all = sensors() || [];
                              const connected = all.filter((s) => s.status === "CONNECTED").length;
                              const offline = all.filter((s) => s.status === "OFFLINE").length;
                              const totalOps = all.reduce((acc, s) => acc + s.eventsPerSecond, 0);
                              const avgValidation = all.length > 0
                                ? (all.reduce((acc, s) => acc + s.validationPassRate, 0) / all.length).toFixed(1)
                                : "0.0";
                              return (
                                <>
                                  <div style={{ display: "flex", "justify-content": "space-between" }}>
                                    <span style={{ color: "#4b5563", "font-size": "0.72rem" }}>Active Sensors</span>
                                    <span style={{ color: "#4ade80", "font-weight": "600", "font-size": "0.72rem" }}>{connected} / {all.length}</span>
                                  </div>
                                  <div style={{ display: "flex", "justify-content": "space-between" }}>
                                    <span style={{ color: "#4b5563", "font-size": "0.72rem" }}>Total OPS</span>
                                    <span style={{ color: "#f8fafc", "font-weight": "600", "font-size": "0.72rem" }}>{totalOps.toFixed(1)}</span>
                                  </div>
                                  <div style={{ display: "flex", "justify-content": "space-between" }}>
                                    <span style={{ color: "#4b5563", "font-size": "0.72rem" }}>Avg Validation</span>
                                    <span style={{ color: "#f8fafc", "font-weight": "600", "font-size": "0.72rem" }}>{avgValidation}%</span>
                                  </div>
                                  <div style={{ display: "flex", "justify-content": "space-between" }}>
                                    <span style={{ color: "#4b5563", "font-size": "0.72rem" }}>Coverage Gaps</span>
                                    <span style={{ color: offline > 0 ? "#f87171" : "#4ade80", "font-weight": "600", "font-size": "0.72rem" }}>{offline}</span>
                                  </div>
                                </>
                              );
                            })()}
                          </div>
                        </div>
                      </Show>
                    </div>
                  </div>
                </div>
              </div>
            }
          >
            <SensorDiagnosticView sensor={selectedSensor()!} />
          </Show>
        </Show>
      </div>

      <style>{`
        .refresh-btn:hover {
            background: rgba(59, 130, 246, 0.2);
            color: white;
        }
      `}</style>
    </div>
  );
}
