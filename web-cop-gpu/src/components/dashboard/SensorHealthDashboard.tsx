// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/SensorHealthDashboard.tsx — Main Health Dashboard
//
// Reference: docs/business/usecases/UC017_sensor_health_monitoring.md

import {
    createEffect,
    createResource,
    createSignal,
    onCleanup,
    Show,
} from "solid-js";
import { fetchSensorStatuses } from "../../services/sensor-health";
import {
    cardView,
    selectedSensor,
    setSelectedSensor
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
  const [hoveredSensorId, setHoveredSensorId] = createSignal<
    string | undefined
  >(undefined);
  const [hoveredSensor, setHoveredSensor] =
    createSignal<Parameters<typeof SensorDetailHoverPanel>[0]["sensor"]>(null);

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
        {/* Header removed — now in global AppShell header */}

        <Show
          when={!sensors.error}
          fallback={
            <div style={{ padding: "40px", color: "#f87171" }}>
              Error loading sensor data: {sensors.error?.message}
            </div>
          }
        >
          {/* Level 2: Sensor Diagnostic Detail — replaces dashboard when sensor selected */}
          <Show when={selectedSensor() !== null}>
            <SensorDiagnosticView sensor={selectedSensor()!} />
          </Show>

          {/* Level 1 Dashboard — visible when no sensor selected */}
          <Show when={selectedSensor() === null}>
          <div
            style={{
              display: "flex",
              "flex-direction": "column",
              flex: 1,
              overflow: "hidden",
              "min-height": 0,
            }}
          >
            {/* ── TOP: Sensor Status Cards ── */}
            <div
              style={{
                flex: "0 0 auto",
                "max-height": "42%",
                overflow: "hidden",
                display: "flex",
                "flex-direction": "column",
              }}
            >
              <div
                style={{
                  padding: "6px 24px 4px",
                  "border-bottom": "1px solid rgba(255,255,255,0.04)",
                  display: "flex",
                  "align-items": "center",
                  gap: "8px",
                }}
              >
                <span
                  style={{
                    "font-size": "0.65rem",
                    "font-weight": "600",
                    "text-transform": "uppercase",
                    "letter-spacing": "0.08em",
                    color: "#64748b",
                  }}
                >
                  Sensor Status Cards
                </span>
                <span style={{ "font-size": "0.6rem", color: "#334155" }}>
                  Real-time health overview of active sensor fleet.
                </span>
              </div>
              <SensorGrid
                sensors={sensors() || []}
                cardView={cardView()}
                onSensorSelect={(s) => {
                  setHoveredSensorId(s.sensorId);
                  setHoveredSensor(s);
                  setSelectedSensor(s);
                }}
              />
            </div>

            {/* ── BOTTOM: Sensor Coverage ── */}
            <div
              style={{
                flex: 1,
                "min-height": "480px",
                "flex-shrink": 0,
                "border-top": "1px solid rgba(255,255,255,0.07)",
                background: "rgba(10, 15, 28, 0.45)",
                "backdrop-filter": "blur(20px)",
                display: "flex",
                overflow: "hidden",
              }}
            >
              {/* Section label overlay */}
              <div
                style={{
                  display: "flex",
                  "flex-direction": "column",
                  flex: 1,
                  "min-width": 0,
                  overflow: "hidden",
                }}
              >
                <div
                  style={{
                    padding: "6px 14px",
                    "border-bottom": "1px solid rgba(255,255,255,0.04)",
                    display: "flex",
                    "align-items": "center",
                    "justify-content": "space-between",
                    "flex-shrink": 0,
                  }}
                >
                  <span
                    style={{
                      "font-size": "0.65rem",
                      "font-weight": "600",
                      "text-transform": "uppercase",
                      "letter-spacing": "0.08em",
                      color: "#64748b",
                    }}
                  >
                    Sensor Coverage
                  </span>
                </div>

                <div
                  style={{
                    display: "flex",
                    flex: 1,
                    overflow: "hidden",
                    "min-height": 0,
                  }}
                >
                  {/* LEFT: fleet list + critical alerts (220px) */}
                  <div
                    style={{
                      width: "220px",
                      "flex-shrink": 0,
                      display: "flex",
                      "flex-direction": "column",
                      "border-right": "1px solid rgba(255,255,255,0.05)",
                      overflow: "hidden",
                      background: "rgba(0,0,0,0.2)",
                    }}
                  >
                    <div
                      style={{
                        padding: "8px 12px",
                        "font-size": "0.6rem",
                        "text-transform": "uppercase",
                        "letter-spacing": "0.1em",
                        color: "#475569",
                        "font-family": "monospace",
                        "border-bottom": "1px solid rgba(255,255,255,0.04)",
                        "flex-shrink": 0,
                      }}
                    >
                      Sensor Fleet
                    </div>
                    <div
                      style={{
                        flex: 1,
                        "overflow-y": "auto",
                        padding: "4px 6px",
                      }}
                    >
                      <SensorFleetList
                        sensors={sensors() || []}
                        selectedSensorId={hoveredSensorId()}
                        onSensorSelect={(s) => {
                          setHoveredSensorId(s.sensorId);
                          setHoveredSensor(s);
                        }}
                        onSensorHover={(s) => {
                          // Keep it persistent on click, but update on hover
                          if (s) {
                            setHoveredSensorId(s.sensorId);
                            setHoveredSensor(s);
                          }
                        }}
                        maxHeight="none"
                      />
                    </div>
                    <div
                      style={{
                        "border-top": "1px solid rgba(255,255,255,0.05)",
                        "overflow-y": "auto",
                        "max-height": "180px",
                        "flex-shrink": 0,
                        background: "rgba(255,0,0,0.02)",
                      }}
                    >
                      <CriticalAlertsPanel
                        spatialAlerts={spatialAlerts()}
                        maxHeight="170px"
                      />
                    </div>
                  </div>

                  {/* CENTER: overview map — fills remaining width */}
                  <div
                    style={{
                      flex: 1,
                      "min-width": 0,
                      overflow: "hidden",
                      padding: "12px",
                      position: "relative",
                    }}
                  >
                    <SensorOverviewMap
                      sensors={sensors() || []}
                      spatialAlerts={spatialAlerts()}
                      hoveredSensorId={hoveredSensorId()}
                      onSensorClick={(s) => {
                        setHoveredSensorId(s.sensorId);
                        setHoveredSensor(s);
                      }}
                      height={400}
                    />
                  </div>

                  {/* RIGHT: sensor detail panel (280px) */}
                  <div
                    style={{
                      width: "280px",
                      "flex-shrink": 0,
                      "border-left": "1px solid rgba(255,255,255,0.05)",
                      padding: "12px",
                      overflow: "hidden",
                      display: "flex",
                      "flex-direction": "column",
                      gap: "12px",
                      background: "rgba(0,0,0,0.1)",
                    }}
                  >
                    <Show when={hoveredSensor() !== null}>
                      <div style={{ flex: 1, "overflow-y": "auto" }}>
                        <SensorDetailHoverPanel
                          sensor={hoveredSensor()}
                          width="100%"
                        />
                      </div>
                    </Show>

                    <Show when={hoveredSensor() === null}>
                      <div
                        style={{
                          padding: "16px",
                          background: "rgba(255,255,255,0.02)",
                          "border-radius": "10px",
                          border: "1px solid rgba(255,255,255,0.05)",
                          "margin-top": "4px",
                          "backdrop-filter": "blur(10px)",
                        }}
                      >
                        <div
                          style={{
                            "font-size": "0.6rem",
                            color: "#475569",
                            "text-transform": "uppercase",
                            "letter-spacing": "0.1em",
                            "margin-bottom": "14px",
                            "font-family": "monospace",
                          }}
                        >
                          SENSOR DETAIL
                        </div>
                        <div
                          style={{
                            display: "flex",
                            "flex-direction": "column",
                            gap: "12px",
                          }}
                        >
                          <div style={{ color: "#334155", "font-size": "0.75rem", "text-align": "center", padding: "20px 0" }}>
                            Select a sensor from the fleet or cards to view real-time diagnostics.
                          </div>
                        </div>
                      </div>
                    </Show>
                  </div>
                </div>
              </div>
            </div>
          </div>
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
