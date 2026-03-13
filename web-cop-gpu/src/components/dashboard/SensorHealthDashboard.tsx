// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/SensorHealthDashboard.tsx — Main Health Dashboard
//
// Reference: docs/business/usecases/UC017_sensor_health_monitoring.md

import { createEffect, createResource, onCleanup, Show } from "solid-js";
import { fetchSensorStatuses } from "../../services/sensor-health";
import {
  cardView,
  selectedSensor,
  setCardView,
  setSelectedSensor,
} from "../../signals/sensor-filters";
import { dashboard } from "../../signals/viewport";
import { DashboardSidebar } from "./DashboardSidebar";
import { SensorDiagnosticView } from "./SensorDiagnosticView";
import { SensorGrid } from "./SensorGrid";

/**
 * Sensor Health Monitoring Dashboard.
 * Orchestrates data fetching, top-level layout, and filtering.
 */
export function SensorHealthDashboard() {
  const [sensors, { refetch }] = createResource(fetchSensorStatuses);

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
              <SensorGrid
                sensors={sensors() || []}
                cardView={cardView()}
                onSensorSelect={setSelectedSensor}
              />
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
