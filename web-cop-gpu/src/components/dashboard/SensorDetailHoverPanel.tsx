// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/SensorDetailHoverPanel.tsx — Sensor detail hover panel
//
// Reference: docs/implementation/v5/sensordashboard_three_level_plan.md §B3

import { createResource, JSX, Show } from "solid-js";
import { fetchSensorDiagnostic, SensorStatus } from "../../services/sensor-health";
import { statusColor } from "./dashboard-utils";
import { MiniCoverageMap } from "./MiniCoverageMap";
import { ObsPerSecChart } from "./ObsPerSecChart";

export interface SensorDetailHoverPanelProps {
  sensor: SensorStatus | null;
  width?: string;
}

function sensorTypeBadgeColor(t: string): string {
  const map: Record<string, string> = {
    RADAR: "#3b82f6",
    "EW/SIGINT": "#8b5cf6",
    "ELINT/COMINT": "#ec4899",
    ISR: "#14b8a6",
    "AIS/BFT": "#22c55e",
    CYBER: "#f97316",
  };
  return map[t] ?? "#64748b";
}

/** Uptime segmented bar (12h / 30d / 90d) */
function UptimeBar(barProps: { pct: number }): JSX.Element {
  const green = barProps.pct;
  const amber = Math.max(0, Math.min(10, 100 - green));
  const red = Math.max(0, 100 - green - amber);
  return (
    <div style={{ display: "flex", height: "6px", "border-radius": "3px", overflow: "hidden", gap: "1px" }}>
      <div style={{ flex: green, background: "#4ade80", "min-width": green > 0 ? "2px" : "0" }} />
      <div style={{ flex: amber, background: "#fbbf24", "min-width": amber > 0 ? "2px" : "0" }} />
      <div style={{ flex: red, background: "#f87171", "min-width": red > 0 ? "2px" : "0" }} />
    </div>
  );
}

/** Right-side sensor detail panel that appears on hover/select — B3 */
export function SensorDetailHoverPanel(props: SensorDetailHoverPanelProps): JSX.Element {
  const sensor = () => props.sensor;

  const [diag] = createResource(sensor, (s) =>
    s ? fetchSensorDiagnostic(s) : Promise.resolve(null)
  );

  return (
    <Show when={sensor() !== null}>
      <div
        data-testid="sensor-detail-hover-panel"
        style={{
          width: props.width ?? "280px",
          "flex-shrink": 0,
          background: "linear-gradient(180deg, rgba(13, 20, 36, 0.4) 0%, rgba(13, 20, 36, 0.7) 100%)",
          "backdrop-filter": "blur(25px)",
          "-webkit-backdrop-filter": "blur(25px)",
          border: "1px solid rgba(255,255,255,0.08)",
          "border-radius": "12px",
          padding: "16px",
          display: "flex",
          "flex-direction": "column",
          gap: "18px",
          overflow: "auto",
          "box-shadow": "0 10px 40px rgba(0,0,0,0.4)",
        }}
      >
        {/* ── Title ── */}
        <div
          style={{
            "font-size": "0.65rem",
            color: "#64748b",
            "text-transform": "uppercase",
            "letter-spacing": "0.12em",
            "font-family": "monospace",
            "border-bottom": "1px solid rgba(255,255,255,0.05)",
            "padding-bottom": "8px",
          }}
        >
          SENSOR DETAIL:
        </div>

        {/* ── Header ── */}
        <div style={{ display: "flex", "flex-direction": "column", gap: "4px" }}>
          <div style={{
            color: "#f1f5f9",
            "font-size": "1rem",
            "font-weight": "700",
            overflow: "hidden",
            "text-overflow": "ellipsis",
            "white-space": "nowrap",
          }}>
            {sensor()!.sensorId}
          </div>
          <div style={{ display: "flex", gap: "8px", "align-items": "center" }}>
            <span style={{
              color: sensorTypeBadgeColor(sensor()!.sensorType),
              "font-size": "0.7rem",
              "font-weight": "600",
            }}>
              {sensor()!.sensorType}
            </span>
            <div style={{ width: "4px", height: "4px", background: "#334155", "border-radius": "50%" }} />
            <span style={{
              color: statusColor(sensor()!.status),
              "font-size": "0.7rem",
              "font-weight": "600",
              "text-transform": "uppercase",
            }}>
              {sensor()!.status}
            </span>
          </div>
        </div>

        {/* ── Health Sparkline ── */}
        <div>
          <div style={{
            "font-size": "0.6rem",
            "text-transform": "uppercase",
            "letter-spacing": "0.1em",
            color: "#64748b",
            "font-family": "monospace",
            "margin-bottom": "8px",
          }}>
            HEALTH GRAPH
          </div>
          <div style={{ background: "rgba(0,0,0,0.2)", "border-radius": "8px", padding: "8px", border: "1px solid rgba(255,255,255,0.03)" }}>
            <Show
              when={diag() && (diag()! as NonNullable<Awaited<ReturnType<typeof fetchSensorDiagnostic>>>).throughputHistory.length > 0}
              fallback={
                <ObsPerSecChart history={Array(15).fill(0)} height={60} />
              }
            >
              <ObsPerSecChart
                history={(diag()! as NonNullable<Awaited<ReturnType<typeof fetchSensorDiagnostic>>>).throughputHistory}
                height={60}
              />
            </Show>
          </div>
        </div>

        {/* ── Connection Uptime ── */}
        <Show when={diag()}>
          <div style={{ display: "flex", "flex-direction": "column", gap: "8px" }}>
            <div style={{
              "font-size": "0.6rem",
              "text-transform": "uppercase",
              "letter-spacing": "0.1em",
              color: "#64748b",
              "font-family": "monospace",
            }}>
              UPTIME
            </div>
            <UptimeBar pct={(diag()! as NonNullable<Awaited<ReturnType<typeof fetchSensorDiagnostic>>>).connectionUptimePct} />
            <div style={{
              display: "flex",
              "justify-content": "space-between",
              "align-items": "baseline",
              "margin-top": "4px",
            }}>
              <div style={{
                "font-size": "0.6rem",
                "text-transform": "uppercase",
                "letter-spacing": "0.1em",
                color: "#475569",
                "font-family": "monospace",
              }}>
                CONNECTION UPTIME
              </div>
              <div style={{
                color: "#4ade80",
                "font-size": "1.2rem",
                "font-weight": "800",
                "font-family": "monospace",
              }}>
                {(diag()! as NonNullable<Awaited<ReturnType<typeof fetchSensorDiagnostic>>>).connectionUptimePct.toFixed(1)}%
              </div>
            </div>
          </div>
        </Show>

        {/* ── Coverage Heatmap ── */}
        <Show when={sensor()!.coverage}>
          <div>
            <div style={{
              "font-size": "0.6rem",
              "text-transform": "uppercase",
              "letter-spacing": "0.1em",
              color: "#64748b",
              "font-family": "monospace",
              "margin-bottom": "8px",
            }}>
              COVERAGE HEATMAP
            </div>
            <div style={{
              display: "flex",
              "justify-content": "center",
              background: "rgba(0,0,0,0.3)",
              "border-radius": "10px",
              padding: "12px",
              border: "1px solid rgba(255,255,255,0.03)",
            }}>
              <MiniCoverageMap
                rangeNm={sensor()!.coverage!.rangeNm}
                bearingStart={sensor()!.coverage!.bearingStart}
                bearingEnd={sensor()!.coverage!.bearingEnd}
                alertLevel={sensor()!.dlqCount > 100 ? 2 : sensor()!.dlqCount > 50 ? 1 : 0}
                width={140}
                height={140}
              />
            </div>
          </div>
        </Show>

        {/* ── Loading state ── */}
        <Show when={diag.loading}>
          <div style={{ color: "#475569", "font-size": "0.68rem", "font-family": "monospace", "text-align": "center" }}>
            Updating Diagnostics…
          </div>
        </Show>
      </div>
    </Show>
  );
}
