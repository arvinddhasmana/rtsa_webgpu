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
          width: props.width ?? "260px",
          "flex-shrink": 0,
          background: "rgba(13, 20, 40, 0.65)",
          "backdrop-filter": "blur(12px)",
          border: "1px solid rgba(255,255,255,0.08)",
          "border-radius": "10px",
          padding: "14px 16px",
          display: "flex",
          "flex-direction": "column",
          gap: "14px",
          overflow: "auto",
        }}
      >
        {/* ── Header ── */}
        <div style={{ display: "flex", "align-items": "flex-start", gap: "8px" }}>
          <div style={{ flex: 1, "min-width": 0 }}>
            <div style={{
              color: "#e2e8f0",
              "font-size": "0.78rem",
              "font-weight": "600",
              "font-family": "monospace",
              overflow: "hidden",
              "text-overflow": "ellipsis",
              "white-space": "nowrap",
            }}>
              {sensor()!.sensorId}
            </div>
            <div style={{ display: "flex", gap: "6px", "margin-top": "4px" }}>
              <span style={{
                background: `${sensorTypeBadgeColor(sensor()!.sensorType)}18`,
                color: sensorTypeBadgeColor(sensor()!.sensorType),
                border: `1px solid ${sensorTypeBadgeColor(sensor()!.sensorType)}35`,
                padding: "1px 7px",
                "border-radius": "10px",
                "font-size": "0.58rem",
                "text-transform": "uppercase",
                "letter-spacing": "0.06em",
              }}>
                {sensor()!.sensorType}
              </span>
              <span style={{
                background: `${statusColor(sensor()!.status)}12`,
                color: statusColor(sensor()!.status),
                border: `1px solid ${statusColor(sensor()!.status)}28`,
                padding: "1px 7px",
                "border-radius": "10px",
                "font-size": "0.58rem",
                "text-transform": "uppercase",
                "letter-spacing": "0.06em",
              }}>
                {sensor()!.status}
              </span>
            </div>
          </div>
        </div>

        {/* ── Health Sparkline ── */}
        <div>
          <div style={{
            "font-size": "0.6rem",
            "text-transform": "uppercase",
            "letter-spacing": "0.1em",
            color: "#475569",
            "font-family": "monospace",
            "margin-bottom": "6px",
          }}>
            Obs/s Trend
          </div>
          <Show
            when={diag() && (diag()! as NonNullable<Awaited<ReturnType<typeof fetchSensorDiagnostic>>>).throughputHistory.length > 0}
            fallback={
              <ObsPerSecChart history={Array(10).fill(0)} height={50} />
            }
          >
            <ObsPerSecChart
              history={(diag()! as NonNullable<Awaited<ReturnType<typeof fetchSensorDiagnostic>>>).throughputHistory}
              height={50}
            />
          </Show>
        </div>

        {/* ── Connection Uptime ── */}
        <Show when={diag()}>
          <div>
            <div style={{
              display: "flex",
              "justify-content": "space-between",
              "align-items": "center",
              "margin-bottom": "6px",
            }}>
              <div style={{
                "font-size": "0.6rem",
                "text-transform": "uppercase",
                "letter-spacing": "0.1em",
                color: "#475569",
                "font-family": "monospace",
              }}>
                Connection Uptime
              </div>
              <div style={{
                color: "#e2e8f0",
                "font-size": "0.95rem",
                "font-weight": "700",
                "font-family": "monospace",
              }}>
                {(diag()! as NonNullable<Awaited<ReturnType<typeof fetchSensorDiagnostic>>>).connectionUptimePct.toFixed(1)}%
              </div>
            </div>
            <UptimeBar pct={(diag()! as NonNullable<Awaited<ReturnType<typeof fetchSensorDiagnostic>>>).connectionUptimePct} />
          </div>
        </Show>

        {/* ── Coverage Heatmap ── */}
        <Show when={sensor()!.coverage}>
          <div>
            <div style={{
              "font-size": "0.6rem",
              "text-transform": "uppercase",
              "letter-spacing": "0.1em",
              color: "#475569",
              "font-family": "monospace",
              "margin-bottom": "6px",
            }}>
              Coverage Zone
            </div>
            <div style={{ display: "flex", "justify-content": "center" }}>
              <MiniCoverageMap
                rangeNm={sensor()!.coverage!.rangeNm}
                bearingStart={sensor()!.coverage!.bearingStart}
                bearingEnd={sensor()!.coverage!.bearingEnd}
                alertLevel={sensor()!.dlqCount > 100 ? 2 : sensor()!.dlqCount > 50 ? 1 : 0}
                width={100}
                height={100}
              />
            </div>
          </div>
        </Show>

        {/* ── Loading state ── */}
        <Show when={diag.loading}>
          <div style={{ color: "#475569", "font-size": "0.68rem", "font-family": "monospace", "text-align": "center" }}>
            Loading…
          </div>
        </Show>
      </div>
    </Show>
  );
}
