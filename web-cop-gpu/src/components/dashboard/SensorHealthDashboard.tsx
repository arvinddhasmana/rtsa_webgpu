// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/SensorHealthDashboard.tsx — Main Health Dashboard
//
// Reference: docs/business/usecases/UC017_sensor_health_monitoring.md

import { createResource, onCleanup, Show } from "solid-js";
import { fetchSensorStatuses } from "../../services/sensor-health";
import { DashboardSidebar } from "./DashboardSidebar";
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

  return (
    <div style={{
      display: "flex",
      height: "100%",
      width: "100%",
      background: "radial-gradient(circle at 0% 0%, rgba(30, 58, 138, 0.15) 0%, transparent 50%), radial-gradient(circle at 100% 100%, rgba(88, 28, 135, 0.15) 0%, transparent 50%)",
      overflow: "hidden"
    }}>
      <DashboardSidebar sensors={sensors() || []} />

      <div style={{ flex: 1, display: "flex", "flex-direction": "column", overflow: "hidden" }}>
        {/* Header / Search bar local to health dashboard could go here */}
        <div style={{
            padding: "16px 24px",
            "border-bottom": "1px solid rgba(255,255,255,0.05)",
            background: "rgba(13, 20, 36, 0.4)",
            display: "flex",
            "justify-content": "space-between",
            "align-items": "center"
        }}>
            <h2 style={{ "font-size": "1.25rem", "font-weight": "600", color: "#f8fafc", margin: 0 }}>Sensor Health Monitor</h2>
            <div style={{ display: "flex", gap: "12px", "align-items": "center" }}>
                <div style={{ "font-size": "0.75rem", color: "#94a3b8" }}>
                    {sensors.loading ? "Updating..." : `Last updated: ${new Date().toLocaleTimeString()}`}
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
                        "font-size": "0.75rem"
                    }}
                    class="refresh-btn"
                >
                    Refresh
                </button>
            </div>
        </div>

        <Show
          when={!sensors.error}
          fallback={<div style={{ padding: "40px", color: "#f87171" }}>Error loading sensor data: {sensors.error?.message}</div>}
        >
          <SensorGrid sensors={sensors() || []} />
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
