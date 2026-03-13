// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/CriticalAlertsPanel.tsx — Critical alerts panel
//
// Reference: docs/implementation/v5/sensordashboard_three_level_plan.md §B4

import { For, JSX, Show } from "solid-js";
import { SpatialAlertPayload } from "../../signals/spatial-alerts";

export interface CriticalAlertsPanelProps {
  spatialAlerts: SpatialAlertPayload[];
  onAlertClick?: (alertId: string) => void;
  maxHeight?: string;
  title?: string;
}

function severityColor(severity: string): string {
  switch (severity) {
    case "CRITICAL": return "#f87171";
    case "ELEVATED": return "#fbbf24";
    default: return "#60a5fa";
  }
}

function severityIcon(severity: string): string {
  switch (severity) {
    case "CRITICAL": return "🔴";
    case "ELEVATED": return "🟡";
    default: return "🔵";
  }
}

function formatTimestamp(isoString: string): string {
  try {
    const d = new Date(isoString);
    return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" });
  } catch {
    return "—";
  }
}

/** Scrollable critical alerts list — B4 */
export function CriticalAlertsPanel(props: CriticalAlertsPanelProps): JSX.Element {
  return (
    <div
      data-testid="critical-alerts-panel"
      style={{
        display: "flex",
        "flex-direction": "column",
        gap: "2px",
      }}
    >
      <div style={{
        "font-size": "0.65rem",
        "text-transform": "uppercase",
        "letter-spacing": "0.1em",
        color: "#475569",
        "font-family": "monospace",
        "margin-bottom": "6px",
      }}>
        {props.title ?? "Critical Alerts"}
      </div>

      <div style={{
        display: "flex",
        "flex-direction": "column",
        gap: "4px",
        "overflow-y": "auto",
        "max-height": props.maxHeight ?? "200px",
      }}>
        <Show
          when={props.spatialAlerts.length > 0}
          fallback={
            <div style={{
              color: "#334155",
              "font-size": "0.7rem",
              "font-family": "monospace",
              padding: "12px 8px",
              "text-align": "center",
              border: "1px dashed rgba(255,255,255,0.05)",
              "border-radius": "6px",
            }}>
              No active alerts
            </div>
          }
        >
          <For each={props.spatialAlerts}>
            {(alert) => {
              const color = severityColor(alert.severity);
              return (
                <div
                  data-testid={`alert-item-${alert.alertId}`}
                  onClick={() => props.onAlertClick?.(alert.alertId)}
                  style={{
                    display: "flex",
                    gap: "8px",
                    padding: "8px 10px",
                    "border-radius": "6px",
                    background: `${color}08`,
                    border: `1px solid ${color}25`,
                    cursor: props.onAlertClick ? "pointer" : "default",
                    "border-left": `3px solid ${color}`,
                    transition: "background 0.15s ease",
                  }}
                >
                  {/* Icon */}
                  <div style={{ "flex-shrink": 0, "font-size": "0.75rem", "line-height": "1.4" }}>
                    {severityIcon(alert.severity)}
                  </div>

                  {/* Content */}
                  <div style={{ flex: 1, "min-width": 0 }}>
                    <div style={{
                      color,
                      "font-size": "0.65rem",
                      "font-weight": "700",
                      "text-transform": "uppercase",
                      "letter-spacing": "0.05em",
                      "font-family": "monospace",
                    }}>
                      [{alert.severity}] SENSOR OFFLINE
                    </div>
                    <div style={{
                      color: "#e2e8f0",
                      "font-size": "0.7rem",
                      "font-family": "monospace",
                      "margin-top": "2px",
                      overflow: "hidden",
                      "text-overflow": "ellipsis",
                      "white-space": "nowrap",
                    }}>
                      {alert.affectedSensorId}
                    </div>
                    <div style={{
                      color: "#475569",
                      "font-size": "0.62rem",
                      "font-family": "monospace",
                      "margin-top": "2px",
                      display: "flex",
                      gap: "6px",
                    }}>
                      <span>{formatTimestamp(alert.lastContactUtc)}</span>
                      <span>·</span>
                      <span>{alert.sectorId}</span>
                    </div>
                  </div>
                </div>
              );
            }}
          </For>
        </Show>
      </div>
    </div>
  );
}
