// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/SensorDiagnosticView.tsx — Deep diagnostic view for a single sensor
//
// Reference: docs/business/usecases/UC017_sensor_health_monitoring.md
// Reference: docs/implementation/v5/ui_images/diagnostic_deep_dive.png

import { createResource, For, Show } from "solid-js";
import { fetchSensorDiagnostic, type SensorStatus } from "../../services/sensor-health";
import { setSelectedSensor } from "../../signals/sensor-filters";

interface SensorDiagnosticViewProps {
  sensor: SensorStatus;
}

function statusColor(s: string): string {
  switch (s) {
    case "CONNECTED":
    case "ACTIVE":
      return "#4ade80";
    case "STALE":
    case "DEGRADED":
      return "#fbbf24";
    default:
      return "#f87171";
  }
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

function severityIcon(sev: string): string {
  switch (sev) {
    case "warn": return "⚠";
    case "error": return "✖";
    default: return "ℹ";
  }
}

function severityColor(sev: string): string {
  switch (sev) {
    case "warn": return "#fbbf24";
    case "error": return "#f87171";
    default: return "#94a3b8";
  }
}

function subStatusColor(s: string): string {
  switch (s) {
    case "ACTIVE": return "#4ade80";
    case "DEGRADED": return "#fbbf24";
    default: return "#f87171";
  }
}

/** Full-pane deep diagnostic view for a single sensor. Never destructure props. */
export function SensorDiagnosticView(props: SensorDiagnosticViewProps) {
  const [data] = createResource(() => props.sensor, fetchSensorDiagnostic);

  const cardStyle = {
    background: "rgba(15, 23, 42, 0.6)",
    "backdrop-filter": "blur(12px)",
    border: "1px solid rgba(255,255,255,0.08)",
    "border-radius": "12px",
    padding: "1.25rem",
  };

  const sectionHeadingStyle = {
    "font-size": "0.75rem",
    "text-transform": "uppercase" as const,
    "letter-spacing": "0.08em",
    color: "#64748b",
    "margin-bottom": "0.75rem",
  };

  return (
    <div
      data-testid="sensor-diagnostic-view"
      style={{
        display: "flex",
        "flex-direction": "column",
        height: "100%",
        overflow: "auto",
        padding: "24px",
        gap: "20px",
        background: "radial-gradient(circle at 0% 0%, rgba(30,58,138,0.15) 0%, transparent 50%)",
        "font-family": "monospace",
      }}
    >
      {/* ── Header row ────────────────────────────────────────────────── */}
      <div style={{ display: "flex", "align-items": "center", gap: "12px", "flex-wrap": "wrap" }}>
        <button
          data-testid="diagnostic-back-btn"
          onClick={() => setSelectedSensor(null)}
          style={{
            background: "rgba(255,255,255,0.05)",
            border: "1px solid rgba(255,255,255,0.1)",
            color: "#e2e8f0",
            padding: "6px 12px",
            "border-radius": "6px",
            cursor: "pointer",
            "font-size": "0.9rem",
            "font-family": "monospace",
          }}
        >
          «
        </button>

        <div style={{ color: "#64748b", "font-size": "0.8rem", "font-family": "monospace" }}>
          Sensor Health /{" "}
          <span style={{ color: "#e2e8f0", "font-weight": "600" }}>{props.sensor.sensorId}</span>
        </div>

        {/* Sensor type badge */}
        <div style={{
          background: `${sensorTypeBadgeColor(props.sensor.sensorType)}20`,
          color: sensorTypeBadgeColor(props.sensor.sensorType),
          border: `1px solid ${sensorTypeBadgeColor(props.sensor.sensorType)}40`,
          padding: "3px 10px",
          "border-radius": "20px",
          "font-size": "0.7rem",
          "text-transform": "uppercase",
          "letter-spacing": "0.05em",
        }}>
          {props.sensor.sensorType}
        </div>

        {/* Status badge */}
        <div style={{
          background: `${statusColor(props.sensor.status)}15`,
          color: statusColor(props.sensor.status),
          border: `1px solid ${statusColor(props.sensor.status)}30`,
          padding: "3px 10px",
          "border-radius": "20px",
          "font-size": "0.7rem",
          "text-transform": "uppercase",
          "letter-spacing": "0.05em",
          display: "flex",
          "align-items": "center",
          gap: "5px",
        }}>
          <span style={{
            width: "6px",
            height: "6px",
            "border-radius": "50%",
            background: statusColor(props.sensor.status),
            display: "inline-block",
          }} />
          {props.sensor.status}
        </div>

        <div style={{ "margin-left": "auto", color: "#64748b", "font-size": "0.75rem" }}>
          Last updated: {new Date().toLocaleTimeString()}
        </div>
      </div>

      {/* ── Loading / error states ─────────────────────────────────────── */}
      <Show when={data.loading}>
        <div style={{ color: "#64748b", padding: "40px", "text-align": "center" }}>
          Loading diagnostic data…
        </div>
      </Show>

      <Show when={data.error}>
        <div style={{ color: "#f87171", padding: "20px" }}>
          Error loading diagnostic: {(data.error as Error)?.message ?? "Unknown error"}
        </div>
      </Show>

      <Show when={data()}>
        {/* ── Metrics row ───────────────────────────────────────────────── */}
        <div style={{
          display: "grid",
          "grid-template-columns": "repeat(auto-fill, minmax(180px, 1fr))",
          gap: "16px",
        }}>
          {/* Throughput */}
          <div style={cardStyle}>
            <div style={sectionHeadingStyle}>Throughput</div>
            <div style={{ "font-size": "1.6rem", "font-weight": "700", color: "#f8fafc" }}>
              {data()!.eventsPerSecond}
            </div>
            <div style={{ "font-size": "0.7rem", color: "#64748b" }}>obs/s</div>
          </div>

          {/* DLQ Count */}
          <div style={cardStyle}>
            <div style={sectionHeadingStyle}>DLQ Count</div>
            <div style={{
              "font-size": "1.6rem",
              "font-weight": "700",
              color: data()!.dlqCount > 50 ? "#f87171" : "#f8fafc",
            }}>
              {data()!.dlqCount}
            </div>
            <div style={{ "font-size": "0.7rem", color: "#64748b" }}>rejected observations</div>
          </div>

          {/* Validation % */}
          <div style={cardStyle}>
            <div style={sectionHeadingStyle}>Validation</div>
            <div style={{
              "font-size": "1.6rem",
              "font-weight": "700",
              color: data()!.validationPassRate > 95
                ? "#4ade80"
                : data()!.validationPassRate > 80
                  ? "#fbbf24"
                  : "#f87171",
            }}>
              {data()!.validationPassRate}%
            </div>
            <div style={{ "font-size": "0.7rem", color: "#64748b" }}>pass rate</div>
          </div>

          {/* Latency */}
          <div style={cardStyle}>
            <div style={sectionHeadingStyle}>Latency</div>
            <div style={{
              "font-size": "1.6rem",
              "font-weight": "700",
              color: data()!.latencyMs > 500 ? "#f87171" : "#f8fafc",
            }}>
              {data()!.latencyMs}
            </div>
            <div style={{ "font-size": "0.7rem", color: "#64748b" }}>ms</div>
          </div>
        </div>

        {/* ── Throughput History chart ───────────────────────────────────── */}
        <div style={cardStyle}>
          <div style={sectionHeadingStyle}>Throughput History (last 20 samples)</div>
          <div style={{ height: "80px", width: "100%" }}>
            {(() => {
              const pts = data()!.throughputHistory;
              const max = Math.max(...pts, 1);
              const w = 100;
              const h = 70;
              const points = pts.map((v, i) => {
                const x = (i / (pts.length - 1)) * w;
                const y = h - (v / max) * h;
                return `${x},${y}`;
              });
              const polyline = points.join(" ");
              const fill = `${polyline} ${w},${h} 0,${h}`;
              const sc = statusColor(props.sensor.status);
              return (
                <svg width="100%" height="100%" viewBox={`0 0 ${w} ${h}`} preserveAspectRatio="none">
                  <defs>
                    <linearGradient id="diag-grad" x1="0%" y1="0%" x2="0%" y2="100%">
                      <stop offset="0%" style={{ "stop-color": sc, "stop-opacity": 0.35 }} />
                      <stop offset="100%" style={{ "stop-color": sc, "stop-opacity": 0 }} />
                    </linearGradient>
                  </defs>
                  {/* Expected range band */}
                  <rect x="0" y={h * 0.1} width={w} height={h * 0.5}
                    fill="rgba(255,255,255,0.04)" rx="2" />
                  {/* Area fill */}
                  <polygon points={fill} fill="url(#diag-grad)" />
                  {/* Line */}
                  <polyline points={polyline} fill="none" stroke={sc} stroke-width="1.5"
                    stroke-linecap="round" stroke-linejoin="round" />
                </svg>
              );
            })()}
          </div>
        </div>

        {/* ── Sub-Sensors table ──────────────────────────────────────────── */}
        <Show when={data()!.subSensors.length > 0}>
          <div style={cardStyle}>
            <div style={sectionHeadingStyle}>Sub-Sensors</div>
            <table style={{ width: "100%", "border-collapse": "collapse", "font-size": "0.8rem" }}>
              <thead>
                <tr style={{ color: "#64748b", "text-align": "left" }}>
                  <th style={{ padding: "6px 8px", "border-bottom": "1px solid rgba(255,255,255,0.08)" }}>Sub-Sensor ID</th>
                  <th style={{ padding: "6px 8px", "border-bottom": "1px solid rgba(255,255,255,0.08)" }}>Status</th>
                  <th style={{ padding: "6px 8px", "border-bottom": "1px solid rgba(255,255,255,0.08)" }}>Location</th>
                  <th style={{ padding: "6px 8px", "border-bottom": "1px solid rgba(255,255,255,0.08)" }}>Last Seen</th>
                </tr>
              </thead>
              <tbody>
                <For each={data()!.subSensors}>
                  {(sub) => (
                    <tr>
                      <td style={{ padding: "6px 8px", color: "#e2e8f0" }}>{sub.id}</td>
                      <td style={{ padding: "6px 8px" }}>
                        <span style={{
                          background: `${subStatusColor(sub.status)}15`,
                          color: subStatusColor(sub.status),
                          border: `1px solid ${subStatusColor(sub.status)}30`,
                          padding: "2px 8px",
                          "border-radius": "10px",
                          "font-size": "0.65rem",
                        }}>{sub.status}</span>
                      </td>
                      <td style={{ padding: "6px 8px", color: "#94a3b8" }}>{sub.location}</td>
                      <td style={{ padding: "6px 8px", color: "#94a3b8" }}>
                        {sub.lastSeenSeconds < 0 ? "N/A" : `${sub.lastSeenSeconds}s ago`}
                      </td>
                    </tr>
                  )}
                </For>
              </tbody>
            </table>
          </div>
        </Show>

        {/* ── DLQ Breakdown ─────────────────────────────────────────────── */}
        <Show when={data()!.dlqBreakdown.length > 0}>
          <div style={cardStyle}>
            <div style={sectionHeadingStyle}>DLQ Breakdown</div>
            {(() => {
              const maxCount = Math.max(...data()!.dlqBreakdown.map(d => d.count), 1);
              return (
                <table style={{ width: "100%", "border-collapse": "collapse", "font-size": "0.8rem" }}>
                  <thead>
                    <tr style={{ color: "#64748b", "text-align": "left" }}>
                      <th style={{ padding: "6px 8px", "border-bottom": "1px solid rgba(255,255,255,0.08)" }}>Rejection Reason</th>
                      <th style={{ padding: "6px 8px", "border-bottom": "1px solid rgba(255,255,255,0.08)", width: "60px" }}>Count</th>
                      <th style={{ padding: "6px 8px", "border-bottom": "1px solid rgba(255,255,255,0.08)" }}>Bar</th>
                    </tr>
                  </thead>
                  <tbody>
                    <For each={data()!.dlqBreakdown}>
                      {(d) => {
                        const pct = (d.count / maxCount) * 100;
                        const barColor = pct > 60 ? "#f87171" : pct > 30 ? "#fbbf24" : "#60a5fa";
                        return (
                          <tr>
                            <td style={{ padding: "6px 8px", color: "#e2e8f0", "font-family": "monospace" }}>{d.reason}</td>
                            <td style={{ padding: "6px 8px", color: "#f87171", "text-align": "right" }}>{d.count}</td>
                            <td style={{ padding: "6px 8px" }}>
                              <div style={{
                                width: `${pct}%`,
                                "min-width": "4px",
                                height: "10px",
                                background: barColor,
                                "border-radius": "3px",
                                opacity: 0.8,
                              }} />
                            </td>
                          </tr>
                        );
                      }}
                    </For>
                  </tbody>
                </table>
              );
            })()}
          </div>
        </Show>

        {/* ── Recent Events timeline ─────────────────────────────────────── */}
        <div style={cardStyle}>
          <div style={sectionHeadingStyle}>Recent Events</div>
          <div style={{ display: "flex", "flex-direction": "column", gap: "8px" }}>
            <For each={data()!.recentEvents}>
              {(ev) => (
                <div style={{
                  display: "flex",
                  gap: "12px",
                  "align-items": "flex-start",
                  "border-bottom": "1px solid rgba(255,255,255,0.04)",
                  "padding-bottom": "8px",
                }}>
                  <span style={{ color: "#64748b", "font-size": "0.7rem", "white-space": "nowrap", "margin-top": "1px" }}>
                    {new Date(ev.timeUtc).toLocaleTimeString()}
                  </span>
                  <span style={{ color: severityColor(ev.severity), "flex-shrink": "0", "font-size": "0.9rem" }}>
                    {severityIcon(ev.severity)}
                  </span>
                  <span style={{ color: "#e2e8f0", "font-size": "0.8rem" }}>{ev.event}</span>
                </div>
              )}
            </For>
          </div>
        </div>
      </Show>
    </div>
  );
}
