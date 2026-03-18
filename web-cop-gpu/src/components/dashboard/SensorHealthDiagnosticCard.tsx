// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/SensorHealthDiagnosticCard.tsx — Glassmorphism diagnostic overlay card
//
// Appears as a floating card overlay when a sensor card is hovered or focused,
// providing a compact health summary before navigating to the full L2 diagnostic view.
//
// Reference: docs/implementation/v5/sensordashboard_three_level_plan.md §B3
// Reference: docs/business/usecases/UC017_sensor_health_monitoring.md

import {
    createMemo,
    createResource,
    createSignal,
    For,
    JSX,
    Show,
} from "solid-js";
import {
    fetchSensorDiagnostic,
    SensorStatus,
} from "../../services/sensor-health";
import { setSelectedSensor } from "../../signals/sensor-filters";
import { statusColor } from "./dashboard-utils";
import { ObsPerSecChart } from "./ObsPerSecChart";

export interface SensorHealthDiagnosticCardProps {
  sensor: SensorStatus;
  onClose?: () => void;
}

function healthLabel(score: number): string {
  if (score >= 85) return "NOMINAL";
  if (score >= 60) return "DEGRADED";
  return "CRITICAL";
}

function healthColor(score: number): string {
  if (score >= 85) return "#4ade80";
  if (score >= 60) return "#fbbf24";
  return "#f87171";
}

function dlqSegmentColor(reason: string): string {
  const r = reason.toLowerCase();
  if (r.includes("timestamp") || r.includes("schema") || r.includes("format"))
    return "#22d3ee";
  if (r.includes("id") || r.includes("missing") || r.includes("crc"))
    return "#60a5fa";
  if (
    r.includes("rate") ||
    r.includes("limit") ||
    r.includes("late") ||
    r.includes("packet")
  )
    return "#f59e0b";
  return "#64748b";
}

/**
 * Compact glassmorphism overlay card showing a sensor health summary.
 * Appears on the Sensor Health dashboard as a draggable overlay.
 * Clicking "View Details" navigates to the full L2 SensorDiagnosticView.
 *
 * Never destructure props — breaks SolidJS reactivity.
 */
export function SensorHealthDiagnosticCard(
  props: SensorHealthDiagnosticCardProps,
): JSX.Element {
  const [diag] = createResource(() => props.sensor, fetchSensorDiagnostic);
  const [btnHovered, setBtnHovered] = createSignal(false);

  const sColor = () => statusColor(props.sensor.status);
  const dlqSegments = createMemo(() => {
    const data = diag()?.dlqReasons ?? [];
    const total = data.reduce((a, b) => a + b.count, 0) || 1;
    let acc = 0;
    return data.map((item) => {
      const pct = Math.round((item.count / total) * 1000) / 10;
      const start = acc;
      acc += pct;
      return { ...item, pct, start };
    });
  });

  const throughputHistory = () => diag()?.throughputHistory ?? [];
  const opsHistory = () => (diag()?.obsPerSecHistory ?? []).slice(-15);
  const timeline = () => (diag()?.recentEvents ?? []).slice(0, 5);

  const expectedRange = () => {
    const base = props.sensor.eventsPerSecond || 0;
    const min = Math.max(5, Math.round(base * 0.65));
    const max = Math.max(min + 10, Math.round(base * 1.4) || 80);
    return { min, max };
  };

  const headlineAlert = () => {
    if (props.sensor.status === "OFFLINE") return "ALERT: SENSOR OFFLINE";
    if (props.sensor.dlqCount >= 20) return "ALERT: DLQ SPIKE DETECTED";
    if (props.sensor.status === "STALE")
      return "ALERT: OBSERVATION DELAY DETECTED";
    return "STATUS: TELEMETRY NOMINAL";
  };

  function validationColor(rate: number): string {
    if (rate >= 95) return "#4ade80";
    if (rate >= 80) return "#fbbf24";
    return "#f87171";
  }

  function dlqColor(count: number): string {
    if (count > 50) return "#f87171";
    if (count > 10) return "#fbbf24";
    return "#4ade80";
  }

  const MetricTile = (metricProps: {
    label: string;
    value: JSX.Element;
    accent?: string;
  }): JSX.Element => (
    <div
      style={{
        background: "rgba(255,255,255,0.02)",
        border: "1px solid rgba(255,255,255,0.06)",
        "border-radius": "10px",
        padding: "10px",
        display: "flex",
        "flex-direction": "column",
        gap: "6px",
        "min-height": "78px",
      }}
    >
      <span
        style={{
          "font-size": "clamp(0.6rem, 0.55rem + 0.12vw, 0.7rem)",
          color: "#64748b",
          "text-transform": "uppercase",
          "letter-spacing": "0.08em",
          "font-family": "monospace",
        }}
      >
        {metricProps.label}
      </span>
      <div
        style={{
          "font-size": "clamp(0.95rem, 0.84rem + 0.22vw, 1.08rem)",
          "font-weight": "700",
          color: metricProps.accent ?? "#e2e8f0",
        }}
      >
        {metricProps.value}
      </div>
    </div>
  );

  const severityColor = (sev: string) => {
    const s = sev.toLowerCase();
    if (s.includes("error")) return "#f87171";
    if (s.includes("warn") || s.includes("alert")) return "#fbbf24";
    return "#4ade80";
  };

  const throughputBars = () => {
    const data = throughputHistory();
    const band = expectedRange();
    const maxVal = Math.max(...data, band.max, 1);
    if (data.length === 0) return null;
    const widthPct = 100 / data.length;
    const lastIdx = data.length - 1;
    const now = new Date();
    const startTime = new Date(now.getTime() - data.length * 60_000);
    const fmtTime = (d: Date) =>
      d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    // Chart inner height (px) used for band/label calculation
    const chartH = 104;
    return (
      <div
        data-testid="throughput-bars"
        style={{ display: "flex", "flex-direction": "column", gap: "4px" }}
      >
        {/* Legend */}
        <div
          style={{
            display: "flex",
            gap: "14px",
            "align-items": "center",
            "flex-wrap": "wrap",
          }}
        >
          <div
            style={{
              display: "flex",
              "align-items": "center",
              gap: "5px",
              "font-size": "0.6rem",
              color: "#94a3b8",
              "font-family": "monospace",
            }}
          >
            <div
              style={{
                width: "12px",
                height: "10px",
                background: "linear-gradient(180deg,#60a5fa,#2563eb)",
                "border-radius": "2px",
              }}
            />
            Current Throughput
          </div>
          <div
            style={{
              display: "flex",
              "align-items": "center",
              gap: "5px",
              "font-size": "0.6rem",
              color: "#94a3b8",
              "font-family": "monospace",
            }}
          >
            <div
              style={{
                width: "12px",
                height: "10px",
                background: "rgba(96,165,250,0.15)",
                border: "1px dashed rgba(96,165,250,0.42)",
                "border-radius": "2px",
              }}
            />
            Expected Range: {band.min}–{band.max} obs/s
          </div>
        </div>

        {/* Chart area */}
        <div
          style={{
            position: "relative",
            height: "130px",
            display: "flex",
            gap: "3px",
            "align-items": "flex-end",
            background:
              "linear-gradient(180deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))",
            padding: `8px 44px 8px 10px`,
            "border-radius": "10px",
            border: "1px solid rgba(255,255,255,0.06)",
            overflow: "visible",
          }}
        >
          {/* Y-axis max label */}
          <div
            style={{
              position: "absolute",
              right: "4px",
              top: "6px",
              "font-size": "0.52rem",
              color: "#475569",
              "font-family": "monospace",
              "white-space": "nowrap",
            }}
          >
            {Math.round(maxVal)}
          </div>
          {/* Y-axis band label */}
          <div
            style={{
              position: "absolute",
              right: "4px",
              bottom: `${8 + (band.min / maxVal) * chartH}px`,
              "font-size": "0.52rem",
              color: "rgba(96,165,250,0.8)",
              "font-family": "monospace",
              "white-space": "nowrap",
            }}
          >
            {band.min} pkts/s
          </div>

          {/* Expected range band */}
          <div
            style={{
              position: "absolute",
              left: "10px",
              right: "44px",
              bottom: `${8 + (band.min / maxVal) * chartH}px`,
              height: `${((band.max - band.min) / maxVal) * chartH}px`,
              background: "rgba(96,165,250,0.1)",
              border: "1px dashed rgba(96,165,250,0.35)",
              "border-radius": "4px",
              "pointer-events": "none",
            }}
          />

          {/* Bars */}
          {data.map((v, i) => (
            <div
              style={{
                width: `${widthPct}%`,
                height: `${Math.max(2, (v / maxVal) * 100)}%`,
                background:
                  i === lastIdx
                    ? "linear-gradient(180deg, #67e8f9, #22d3ee)"
                    : "linear-gradient(180deg, #60a5fa, #2563eb)",
                "border-radius": "3px 3px 1px 1px",
                opacity: i === lastIdx ? 1 : 0.82,
                "flex-shrink": 0,
                position: "relative",
              }}
              title={`t-${data.length - i}m: ${v} obs/s`}
            >
              {i === lastIdx && (
                <div
                  style={{
                    position: "absolute",
                    top: "-18px",
                    left: "50%",
                    transform: "translateX(-50%)",
                    "font-size": "0.58rem",
                    color: "#22d3ee",
                    "font-weight": "700",
                    "font-family": "monospace",
                    "white-space": "nowrap",
                    "pointer-events": "none",
                  }}
                >
                  {v}
                </div>
              )}
            </div>
          ))}
        </div>

        {/* X-axis time labels */}
        <div
          style={{
            display: "flex",
            "justify-content": "space-between",
            "font-size": "0.52rem",
            color: "#475569",
            "font-family": "monospace",
            padding: "0 4px",
          }}
        >
          <span>[{fmtTime(startTime)}]</span>
          <span>Last 15 Min</span>
          <span>[{fmtTime(now)}]</span>
        </div>
      </div>
    );
  };

  return (
    <div
      data-testid="sensor-health-diagnostic-card"
      style={{
        width: "100%",
        background:
          "linear-gradient(180deg, rgba(9,18,37,0.92) 0%, rgba(6,14,30,0.88) 100%)",
        "backdrop-filter": "blur(24px) saturate(150%)",
        "-webkit-backdrop-filter": "blur(24px) saturate(150%)",
        border: `1px solid ${sColor()}45`,
        "border-radius": "14px",
        padding: "9px",
        "box-shadow": `0 22px 70px rgba(0,0,0,0.62), 0 0 24px ${sColor()}20, inset 0 1px 0 rgba(255,255,255,0.06)`,
        display: "flex",
        "flex-direction": "column",
        gap: "9px",
        color: "#f1f5f9",
        "font-family": "monospace",
      }}
    >
      <div
        style={{
          height: "2px",
          background:
            "linear-gradient(90deg, rgba(34,211,238,0.8), rgba(59,130,246,0.05))",
          "border-radius": "999px",
        }}
      />

      <div style={{ display: "flex", "justify-content": "space-between", "align-items": "flex-start", padding: "0 4px", "margin-bottom": "2px" }}>
        <div style={{ display: "flex", "flex-direction": "column" }}>
            <span style={{ "font-size": "0.55rem", color: "#64748b", "text-transform": "uppercase", "letter-spacing": "0.15em", "font-weight": "800" }}>
                {props.sensor.sensorType}
            </span>
            <span style={{ "font-size": "0.85rem", "font-weight": "900", color: "#e2e8f0", "letter-spacing": "0.02em" }}>
                {props.sensor.sensorId}
            </span>
        </div>
        <button
          data-testid="diagnostic-card-close-btn"
          onClick={() => props.onClose?.()}
          style={{
            background: "rgba(255,255,255,0.03)",
            border: "1px solid rgba(255,255,255,0.08)",
            color: "#94a3b8",
            "border-radius": "6px",
            width: "22px",
            height: "22px",
            display: "flex",
            "align-items": "center",
            "justify-content": "center",
            cursor: "pointer",
            transition: "all 0.2s",
            "font-size": "1.1rem",
            "line-height": "1"
          }}
          onMouseEnter={(e) => {
            const t = e.currentTarget;
            t.style.background = "rgba(255,255,255,0.08)";
            t.style.color = "#f1f5f9";
          }}
          onMouseLeave={(e) => {
            const t = e.currentTarget;
            t.style.background = "rgba(255,255,255,0.03)";
            t.style.color = "#94a3b8";
          }}
        >
          ×
        </button>
      </div>

      <div
        style={{
          display: "flex",
          "align-items": "center",
          "justify-content": "space-between",
          gap: "8px",
          padding: "6px 8px",
          "border-radius": "8px",
          background:
            props.sensor.status === "CONNECTED"
              ? "linear-gradient(90deg, rgba(74,222,128,0.14), rgba(74,222,128,0.04))"
              : "linear-gradient(90deg, rgba(251,191,36,0.18), rgba(248,113,113,0.08))",
          border:
            props.sensor.status === "CONNECTED"
              ? "1px solid rgba(74,222,128,0.3)"
              : "1px solid rgba(251,191,36,0.35)",
        }}
      >
        <div style={{ display: "flex", "flex-direction": "column", gap: "1px" }}>
            <span
              style={{
                "font-size": "0.52rem",
                "font-weight": "800",
                color: props.sensor.status === "CONNECTED" ? "#86efac" : "#fbbf24",
                opacity: 0.7,
                "letter-spacing": "0.1em"
              }}
            >
              {props.sensor.status}
            </span>
            <span
              style={{
                "font-size": "0.72rem",
                "font-weight": "800",
                color: props.sensor.status === "CONNECTED" ? "#86efac" : "#fbbf24",
                "letter-spacing": "0.02em",
                "text-transform": "uppercase",
              }}
            >
              {headlineAlert()}
            </span>
        </div>
        <span style={{ "font-size": "0.6rem", color: "#94a3b8" }}>
          {new Date().toLocaleTimeString([], {
            hour: "2-digit",
            minute: "2-digit",
            second: "2-digit",
          })}
        </span>
      </div>



      <Show
        when={diag()}
        fallback={
          <div
            style={{
              color: "#334155",
              "font-size": "0.65rem",
              "text-align": "center",
              padding: "6px 0",
            }}
          >
            Loading diagnostics…
          </div>
        }
      >
        {(d) => (
          <>
            {/* Health + summary row */}
            <div
              style={{
                display: "grid",
                "grid-template-columns": "1fr 1fr",
                gap: "10px",
              }}
            >
              <div
                style={{
                  background:
                    "linear-gradient(180deg, rgba(251,191,36,0.12) 0%, rgba(255,255,255,0.02) 100%)",
                  border: "1px solid rgba(251,191,36,0.25)",
                  "border-radius": "12px",
                  padding: "11px",
                  display: "flex",
                  gap: "12px",
                  "align-items": "center",
                }}
              >
                <div
                  style={{
                    position: "relative",
                    width: "88px",
                    height: "88px",
                  }}
                >
                  {(() => {
                    const circumference = 2 * Math.PI * 38;
                    const pct = Math.max(0, Math.min(100, d().healthScore));
                    const offset = circumference * (1 - pct / 100);
                    return (
                      <svg
                        width="88"
                        height="88"
                        viewBox="0 0 88 88"
                        style={{ transform: "rotate(-90deg)" }}
                      >
                        <circle
                          cx="44"
                          cy="44"
                          r="38"
                          stroke="rgba(255,255,255,0.08)"
                          stroke-width="8"
                          fill="none"
                        />
                        <circle
                          cx="44"
                          cy="44"
                          r="38"
                          stroke={healthColor(pct)}
                          stroke-width="8"
                          stroke-linecap="round"
                          fill="none"
                          stroke-dasharray={`${circumference} ${circumference}`}
                          stroke-dashoffset={offset}
                        />
                      </svg>
                    );
                  })()}
                  <div
                    style={{
                      position: "absolute",
                      inset: 0,
                      display: "flex",
                      "align-items": "center",
                      "justify-content": "center",
                      "flex-direction": "column",
                      "pointer-events": "none",
                    }}
                  >
                    <div
                      style={{
                        "font-size": "clamp(1.35rem, 1.2rem + 0.25vw, 1.56rem)",
                        "font-weight": "800",
                        color: healthColor(d().healthScore),
                      }}
                    >
                      {d().healthScore}
                    </div>
                    <div
                      style={{
                        "font-size": "clamp(0.6rem, 0.56rem + 0.08vw, 0.66rem)",
                        color: "#64748b",
                        "letter-spacing": "0.06em",
                      }}
                    >
                      {healthLabel(d().healthScore)}
                    </div>
                  </div>
                </div>
                <div
                  style={{
                    display: "flex",
                    "flex-direction": "column",
                    gap: "8px",
                    flex: 1,
                  }}
                >
                  <MetricTile
                    label="Validation"
                    value={
                      <span
                        style={{
                          color: validationColor(
                            props.sensor.validationPassRate,
                          ),
                        }}
                      >
                        {props.sensor.status === "OFFLINE"
                          ? "N/A"
                          : `${props.sensor.validationPassRate}%`}
                      </span>
                    }
                  />
                  <MetricTile
                    label="Avg Latency"
                    value={
                      <span>
                        {d().avgLatencyMs}{" "}
                        <span
                          style={{ color: "#475569", "font-size": "0.65rem" }}
                        >
                          ms
                        </span>
                      </span>
                    }
                    accent={
                      d().avgLatencyMs < 150
                        ? "#4ade80"
                        : d().avgLatencyMs < 250
                          ? "#fbbf24"
                          : "#f87171"
                    }
                  />
                </div>
              </div>

              <div
                style={{
                  display: "grid",
                  "grid-template-columns": "1fr 1fr",
                  gap: "7px",
                }}
              >
                <MetricTile
                  label="Throughput"
                  value={
                    <span>
                      {props.sensor.eventsPerSecond.toFixed(1)}{" "}
                      <span style={{ color: "#475569", "font-size": "0.6rem" }}>
                        obs/s
                      </span>
                    </span>
                  }
                  accent="#60a5fa"
                />
                <MetricTile
                  label="Uptime"
                  value={<span>{d().connectionUptimePct.toFixed(1)}%</span>}
                  accent={d().connectionUptimePct >= 95 ? "#4ade80" : "#fbbf24"}
                />
                <MetricTile
                  label="DLQ Count"
                  value={
                    <span style={{ color: dlqColor(props.sensor.dlqCount) }}>
                      {props.sensor.dlqCount}
                    </span>
                  }
                  accent={dlqColor(props.sensor.dlqCount)}
                />
                <MetricTile
                  label="Peak Tput"
                  value={
                    <span>
                      {d().peakThroughput}{" "}
                      <span style={{ color: "#475569", "font-size": "0.6rem" }}>
                        obs/s
                      </span>
                    </span>
                  }
                  accent="#c084fc"
                />
              </div>
            </div>

            {/* Timeline + DLQ donut */}
            <div
              style={{
                display: "grid",
                "grid-template-columns": "1.2fr 1fr",
                gap: "10px",
              }}
            >
              <div
                style={{
                  background: "rgba(11,22,41,0.78)",
                  border: "1px solid rgba(56,189,248,0.18)",
                  "border-radius": "12px",
                  padding: "9px 10px",
                  display: "flex",
                  "flex-direction": "column",
                  gap: "8px",
                }}
              >
                <div
                  style={{
                    display: "flex",
                    "justify-content": "space-between",
                    "align-items": "center",
                  }}
                >
                  <span
                    style={{
                      "font-size": "clamp(0.72rem, 0.65rem + 0.12vw, 0.8rem)",
                      "text-transform": "uppercase",
                      color: "#94a3b8",
                      "letter-spacing": "0.08em",
                    }}
                  >
                    Diagnostic Event Timeline
                  </span>
                  <span
                    style={{
                      "font-size": "0.8rem",
                      color: "#475569",
                      "line-height": 1,
                    }}
                  >
                    ∨
                  </span>
                </div>
                <Show
                  when={timeline().length > 0}
                  fallback={
                    <div style={{ color: "#334155", "font-size": "0.65rem" }}>
                      No recent events
                    </div>
                  }
                >
                  <div
                    style={{
                      display: "flex",
                      "flex-direction": "column",
                      gap: "6px",
                      "overflow-y": "auto",
                      "max-height": "260px",
                    }}
                  >
                    <For each={timeline()}>
                      {(evt) => {
                        const sColor = severityColor(evt.severity);
                        const isStatusEvt =
                          evt.event.toLowerCase().includes("connect") ||
                          evt.event.toLowerCase().includes("initialized") ||
                          evt.event.toLowerCase().includes("status:");
                        const category = isStatusEvt ? "STATUS" : "EVENTS";
                        const colonIdx = evt.event.indexOf(":");
                        const title =
                          colonIdx > -1
                            ? evt.event.slice(0, colonIdx).trim()
                            : evt.event;
                        const sub =
                          colonIdx > -1
                            ? evt.event.slice(colonIdx + 1).trim()
                            : "";
                        const icon =
                          isStatusEvt &&
                          evt.event.toLowerCase().includes("connect") &&
                          evt.severity === "info"
                            ? "✓"
                            : evt.severity === "error"
                              ? "⚠"
                              : evt.severity === "warn"
                                ? "⚠"
                                : "●";
                        return (
                          <div
                            style={{
                              padding: "7px 8px 7px 10px",
                              background: "rgba(255,255,255,0.02)",
                              border: `1px solid ${sColor}22`,
                              "border-left": `3px solid ${sColor}`,
                              "border-radius": "8px",
                              display: "flex",
                              "flex-direction": "column",
                              gap: "3px",
                            }}
                          >
                            {/* Time + icon + category row */}
                            <div
                              style={{
                                display: "flex",
                                "align-items": "center",
                                gap: "5px",
                              }}
                            >
                              <span
                                style={{
                                  color: "#64748b",
                                  "font-size": "0.58rem",
                                  "font-family": "monospace",
                                  "flex-shrink": 0,
                                }}
                              >
                                {new Date(evt.timeUtc).toLocaleTimeString([], {
                                  hour: "2-digit",
                                  minute: "2-digit",
                                  second: "2-digit",
                                })}
                              </span>
                              <span
                                style={{
                                  color: sColor,
                                  "font-size": "0.72rem",
                                  "line-height": 1,
                                  "flex-shrink": 0,
                                }}
                              >
                                {icon}
                              </span>
                              <span
                                style={{
                                  background: `${sColor}1a`,
                                  color: sColor,
                                  border: `1px solid ${sColor}33`,
                                  "border-radius": "3px",
                                  padding: "1px 5px",
                                  "font-size": "0.5rem",
                                  "font-weight": "700",
                                  "text-transform": "uppercase",
                                  "letter-spacing": "0.08em",
                                  "font-family": "monospace",
                                  "flex-shrink": 0,
                                }}
                              >
                                {category}
                              </span>
                              <div style={{ flex: 1 }} />
                              <span
                                style={{
                                  color: "#334155",
                                  "font-size": "0.55rem",
                                }}
                              >
                                ···
                              </span>
                            </div>
                            {/* Event title */}
                            <span
                              style={{
                                color: "#e2e8f0",
                                "font-weight": "700",
                                "font-size":
                                  "clamp(0.7rem, 0.64rem + 0.1vw, 0.78rem)",
                                "line-height": "1.3",
                                "padding-left": "12px",
                              }}
                            >
                              {title}
                            </span>
                            {/* Sub-description */}
                            <Show when={sub.length > 0}>
                              <span
                                style={{
                                  color: "#64748b",
                                  "font-size":
                                    "clamp(0.6rem, 0.56rem + 0.08vw, 0.66rem)",
                                  "line-height": "1.4",
                                  "padding-left": "12px",
                                }}
                              >
                                {sub}
                              </span>
                            </Show>
                          </div>
                        );
                      }}
                    </For>
                  </div>
                </Show>
              </div>

              <div
                style={{
                  background: "rgba(11,22,41,0.78)",
                  border: "1px solid rgba(56,189,248,0.18)",
                  "border-radius": "12px",
                  padding: "9px 10px",
                  display: "flex",
                  "flex-direction": "column",
                  gap: "8px",
                }}
              >
                <div
                  style={{
                    display: "flex",
                    "justify-content": "space-between",
                    "align-items": "center",
                  }}
                >
                  <span
                    style={{
                      "font-size": "clamp(0.72rem, 0.65rem + 0.12vw, 0.8rem)",
                      "text-transform": "uppercase",
                      color: "#94a3b8",
                      "letter-spacing": "0.08em",
                    }}
                  >
                    DLQ Rejection Reasons
                  </span>
                  <span style={{ "font-size": "0.6rem", color: "#475569" }}>
                    Total: {dlqSegments().reduce((a, b) => a + b.count, 0)}
                  </span>
                </div>

                <Show
                  when={dlqSegments().length > 0}
                  fallback={
                    <div style={{ color: "#334155", "font-size": "0.65rem" }}>
                      No DLQ entries
                    </div>
                  }
                >
                  <div
                    style={{
                      display: "flex",
                      gap: "12px",
                      "align-items": "center",
                    }}
                  >
                    <svg width="160" height="160" viewBox="0 0 160 160">
                      {(() => {
                        const radius = 60;
                        const circumference = 2 * Math.PI * radius;
                        return dlqSegments().map((seg) => (
                          <circle
                            cx="80"
                            cy="80"
                            r={radius}
                            fill="none"
                            stroke={dlqSegmentColor(seg.reason)}
                            stroke-width="18"
                            stroke-dasharray={`${(seg.pct / 100) * circumference} ${circumference}`}
                            stroke-dashoffset={
                              -(seg.start / 100) * circumference
                            }
                            stroke-linecap="butt"
                          />
                        ));
                      })()}
                      <circle
                        cx="80"
                        cy="80"
                        r="40"
                        fill="rgba(8,12,24,0.96)"
                      />
                      <text
                        x="80"
                        y="78"
                        text-anchor="middle"
                        fill="#e2e8f0"
                        font-size="14"
                        font-weight="700"
                      >
                        Total
                      </text>
                      <text
                        x="80"
                        y="98"
                        text-anchor="middle"
                        fill="#94a3b8"
                        font-size="12"
                      >
                        {dlqSegments().reduce((a, b) => a + b.count, 0)} rejects
                      </text>
                    </svg>
                    <div
                      style={{
                        display: "flex",
                        "flex-direction": "column",
                        gap: "6px",
                      }}
                    >
                      <For each={dlqSegments()}>
                        {(seg) => (
                          <div
                            style={{
                              display: "flex",
                              "align-items": "center",
                              gap: "6px",
                              "font-size": "0.63rem",
                              color: "#e2e8f0",
                            }}
                          >
                            <div
                              style={{
                                width: "8px",
                                height: "8px",
                                "border-radius": "2px",
                                background: dlqSegmentColor(seg.reason),
                                "flex-shrink": 0,
                              }}
                            />
                            <span
                              style={{
                                flex: 1,
                                color: "#94a3b8",
                                "white-space": "nowrap",
                                overflow: "hidden",
                                "text-overflow": "ellipsis",
                              }}
                            >
                              {seg.reason}
                            </span>
                            <span
                              style={{
                                color: dlqSegmentColor(seg.reason),
                                "font-weight": "600",
                                "flex-shrink": 0,
                              }}
                            >
                              {seg.count}
                            </span>
                            <span
                              style={{ color: "#475569", "flex-shrink": 0 }}
                            >
                              {seg.pct.toFixed(1)}%
                            </span>
                          </div>
                        )}
                      </For>
                    </div>
                  </div>
                </Show>
              </div>
            </div>

            {/* Throughput — full width */}
            <div
              style={{
                background: "rgba(11,22,41,0.78)",
                border: "1px solid rgba(96,165,250,0.2)",
                "border-radius": "12px",
                padding: "9px 10px",
                display: "flex",
                "flex-direction": "column",
                gap: "8px",
              }}
            >
              <span
                style={{
                  "font-size": "clamp(0.72rem, 0.65rem + 0.12vw, 0.8rem)",
                  color: "#94a3b8",
                  "letter-spacing": "0.08em",
                  "text-transform": "uppercase",
                }}
              >
                Throughput vs. Expected Range
              </span>
              {throughputBars() ?? (
                <div style={{ color: "#334155", "font-size": "0.65rem" }}>
                  No throughput history
                </div>
              )}
            </div>

            {/* Observation Rate — full width */}
            <div
              style={{
                background: "rgba(11,22,41,0.78)",
                border: "1px solid rgba(96,165,250,0.2)",
                "border-radius": "12px",
                padding: "9px 10px",
                display: "flex",
                "flex-direction": "column",
                gap: "8px",
              }}
            >
              <div
                style={{
                  display: "flex",
                  "justify-content": "space-between",
                  "align-items": "center",
                }}
              >
                <div style={{ display: "flex", "align-items": "center", gap: "6px" }}>
                  <span
                    style={{
                      "font-size": "clamp(0.72rem, 0.65rem + 0.12vw, 0.8rem)",
                      color: "#94a3b8",
                      "letter-spacing": "0.08em",
                      "text-transform": "uppercase",
                    }}
                  >
                    Observation Rate — Last 15 Min
                  </span>
                  <div style={{
                    width: "8px", height: "8px", background: "#3b82f6", "border-radius": "1px"
                  }} />
                </div>
                <span style={{ "font-size": "0.6rem", color: "#475569" }}>
                  OBS / SEC
                </span>
              </div>
              <ObsPerSecChart history={opsHistory()} height={90} />
            </div>

            {/* CTA */}
            <button
              data-testid="diagnostic-card-view-full-btn"
              onClick={() => setSelectedSensor(props.sensor)}
              onMouseEnter={() => setBtnHovered(true)}
              onMouseLeave={() => setBtnHovered(false)}
              style={{
                background: btnHovered()
                  ? `${sColor()}25`
                  : `linear-gradient(135deg, ${sColor()}20, ${sColor()}10)`,
                border: `1px solid ${btnHovered() ? sColor() + "70" : sColor() + "40"}`,
                "box-shadow": btnHovered() ? `0 0 12px ${sColor()}20` : "none",
                color: sColor(),
                cursor: "pointer",
                "border-radius": "10px",
                padding: "10px 12px",
                "font-size": "0.68rem",
                "font-weight": "700",
                "text-transform": "uppercase",
                "letter-spacing": "0.08em",
                "font-family": "monospace",
                transition: "all 0.2s",
                width: "100%",
                display: "flex",
                "align-items": "center",
                "justify-content": "center",
                gap: "6px",
              }}
            >
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2.5"
              >
                <polyline points="9 18 15 12 9 6" />
              </svg>
              View Full Diagnostics
            </button>
          </>
        )}
      </Show>
    </div>
  );
}
