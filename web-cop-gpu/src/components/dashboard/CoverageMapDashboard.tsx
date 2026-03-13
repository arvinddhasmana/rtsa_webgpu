// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/CoverageMapDashboard.tsx — Level 3: Full Coverage Map
//
// Strategic view of the entire sensor network coverage footprint.
// Highlights geographic gaps and active spatial alerts.
//
// Phase A: Routing stub — wires signals and navigation; full WebGPU rendering
// is added in Phase B.
//
// Reference: docs/implementation/v5/sensordashboard_three_level_plan.md §L3

import { For, Show } from "solid-js";
import {
  activeSpatialAlertId,
  setActiveSpatialAlertId,
  spatialAlerts,
  type SpatialAlertPayload,
} from "../../signals/spatial-alerts";

const SEVERITY_COLOR: Record<SpatialAlertPayload["severity"], string> = {
  CRITICAL: "#ef4444",
  ELEVATED: "#f97316",
  WATCH: "#f59e0b",
};

/** Level 3 — Full Coverage Map Dashboard. */
export function CoverageMapDashboard() {
  return (
    <div
      data-testid="coverage-map-dashboard"
      style={{
        display: "flex",
        height: "100%",
        width: "100%",
        background:
          "radial-gradient(circle at 0% 0%, rgba(30, 58, 138, 0.15) 0%, transparent 50%), " +
          "radial-gradient(circle at 100% 100%, rgba(88, 28, 135, 0.15) 0%, transparent 50%)",
        overflow: "hidden",
      }}
    >
      {/* ── Main map area (Phase B: WebGPU coverage layer) ── */}
      <div
        style={{
          flex: 1,
          display: "flex",
          "flex-direction": "column",
          "align-items": "center",
          "justify-content": "center",
          position: "relative",
          overflow: "hidden",
        }}
      >
        <div
          style={{
            "text-align": "center",
            color: "#4b5563",
          }}
        >
          <div
            style={{
              "font-size": "2rem",
              "margin-bottom": "0.5rem",
            }}
          >
            ◎
          </div>
          <div style={{ "font-size": "0.9rem", "font-weight": "600", color: "#64748b" }}>
            Coverage Map
          </div>
          <div style={{ "font-size": "0.75rem", color: "#374151", "margin-top": "0.4rem" }}>
            WebGPU sensor footprint rendering — Phase B
          </div>
        </div>

        {/* Active alert highlight banner */}
        <Show when={activeSpatialAlertId() !== null}>
          <div
            style={{
              position: "absolute",
              bottom: "0",
              left: "0",
              right: "0",
              background: "rgba(239,68,68,0.12)",
              border: "1px solid rgba(239,68,68,0.3)",
              padding: "10px 16px",
              display: "flex",
              "align-items": "center",
              "justify-content": "space-between",
            }}
          >
            <span style={{ "font-size": "0.8rem", color: "#fca5a5" }}>
              Inspecting alert:{" "}
              <strong style={{ "font-family": "monospace" }}>{activeSpatialAlertId()}</strong>
            </span>
            <button
              onClick={() => setActiveSpatialAlertId(null)}
              style={{
                background: "transparent",
                border: "1px solid rgba(239,68,68,0.4)",
                color: "#fca5a5",
                "border-radius": "4px",
                padding: "2px 8px",
                "font-size": "0.7rem",
                cursor: "pointer",
              }}
            >
              Dismiss
            </button>
          </div>
        </Show>
      </div>

      {/* ── Right panel: spatial alert list ── */}
      <div
        style={{
          width: "320px",
          "border-left": "1px solid rgba(255,255,255,0.05)",
          background: "rgba(13, 20, 36, 0.4)",
          display: "flex",
          "flex-direction": "column",
          overflow: "hidden",
        }}
      >
        {/* Panel header */}
        <div
          style={{
            padding: "12px 16px",
            "border-bottom": "1px solid rgba(255,255,255,0.05)",
            display: "flex",
            "align-items": "center",
            "justify-content": "space-between",
            "flex-shrink": "0",
          }}
        >
          <span style={{ "font-size": "0.75rem", "font-weight": "700", color: "#ef4444" }}>
            COVERAGE GAPS
          </span>
          <span
            style={{
              "font-size": "0.65rem",
              background: "rgba(239,68,68,0.15)",
              color: "#ef4444",
              padding: "0 0.4rem",
              "border-radius": "3px",
            }}
          >
            {spatialAlerts().length}
          </span>
        </div>

        {/* Alert list */}
        <div style={{ flex: "1", "overflow-y": "auto" }}>
          <Show
            when={spatialAlerts().length > 0}
            fallback={
              <div
                style={{
                  padding: "1rem",
                  color: "#64748b",
                  "font-size": "0.75rem",
                  "text-align": "center",
                }}
              >
                No active coverage gaps
              </div>
            }
          >
            <For each={spatialAlerts()}>
              {(alert) => (
                <div
                  style={{
                    padding: "0.6rem 0.75rem",
                    "border-bottom": "1px solid rgba(255,255,255,0.04)",
                    background:
                      activeSpatialAlertId() === alert.alertId
                        ? "rgba(239,68,68,0.08)"
                        : "transparent",
                    opacity: alert.acknowledged ? "0.5" : "1",
                    cursor: "pointer",
                  }}
                  onClick={() =>
                    setActiveSpatialAlertId(
                      activeSpatialAlertId() === alert.alertId ? null : alert.alertId,
                    )
                  }
                  aria-label={`Spatial alert: ${alert.description}`}
                >
                  <div style={{ display: "flex", gap: "0.4rem", "align-items": "flex-start" }}>
                    <div
                      style={{
                        width: "8px",
                        height: "8px",
                        "border-radius": "50%",
                        background: SEVERITY_COLOR[alert.severity],
                        "flex-shrink": "0",
                        "margin-top": "0.3rem",
                      }}
                    />
                    <div style={{ flex: "1", "min-width": "0" }}>
                      <div
                        style={{
                          "font-size": "0.7rem",
                          "font-weight": "bold",
                          color: SEVERITY_COLOR[alert.severity],
                        }}
                      >
                        {alert.severity}
                      </div>
                      <div style={{ "font-size": "0.75rem", "word-break": "break-word" }}>
                        {alert.description}
                      </div>
                      <div
                        style={{
                          "font-size": "0.65rem",
                          color: "#64748b",
                          "margin-top": "0.2rem",
                        }}
                      >
                        Sector: {alert.sectorId} · Sensor: {alert.affectedSensorId}
                      </div>
                    </div>
                  </div>
                </div>
              )}
            </For>
          </Show>
        </div>
      </div>
    </div>
  );
}
