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
          "font-size": "0.6rem",
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
          "font-size": "0.95rem",
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
    const maxVal = Math.max(...data, band.max);
    if (data.length === 0) return null;
    const widthPct = 100 / data.length;
    return (
      <div
        data-testid="throughput-bars"
        style={{
          position: "relative",
          height: "140px",
          display: "flex",
          gap: "4px",
          "align-items": "flex-end",
          background:
            "linear-gradient(180deg, rgba(255,255,255,0.03), rgba(255,255,255,0.01))",
          padding: "10px",
          "border-radius": "10px",
          border: "1px solid rgba(255,255,255,0.06)",
        }}
      >
        <div
          style={{
            position: "absolute",
            left: 0,
            right: 0,
            bottom: `${(band.min / maxVal) * 100}%`,
            height: `${((band.max - band.min) / maxVal) * 100}%`,
            background: "rgba(96, 165, 250, 0.1)",
            border: "1px dashed rgba(96,165,250,0.35)",
            "border-radius": "6px",
            "pointer-events": "none",
          }}
        />
        {data.map((v, i) => (
          <div
            style={{
              width: `${widthPct}%`,
              height: `${Math.max(0, (v / maxVal) * 100)}%`,
              background: "linear-gradient(180deg, #60a5fa, #2563eb)",
              "border-radius": "6px 6px 2px 2px",
              opacity: 0.9,
            }}
            title={`t-${data.length - i}m: ${v} obs/s`}
          />
        ))}
      </div>
    );
  };

  return (
    <div
      data-testid="sensor-health-diagnostic-card"
      style={{
        width: "100%",
        background: "rgba(8, 12, 24, 0.86)",
        "backdrop-filter": "blur(28px) saturate(160%)",
        "-webkit-backdrop-filter": "blur(28px) saturate(160%)",
        border: `1px solid ${sColor()}35`,
        "border-radius": "14px",
        padding: "16px",
        "box-shadow": `0 22px 70px rgba(0,0,0,0.6), 0 0 30px ${sColor()}15, inset 0 1px 0 rgba(255,255,255,0.05)`,
        display: "flex",
        "flex-direction": "column",
        gap: "12px",
        color: "#f1f5f9",
        "font-family": "monospace",
      }}
    >
      <div
        style={{
          display: "flex",
          "align-items": "center",
          gap: "10px",
          "border-bottom": "1px solid rgba(255,255,255,0.06)",
          padding: "2px 0 10px",
        }}
      >
        <div
          style={{
            display: "flex",
            "align-items": "center",
            gap: "10px",
            flex: 1,
          }}
        >
          <div
            style={{
              width: "36px",
              height: "36px",
              "border-radius": "12px",
              background: `${sColor()}1a`,
              border: `1px solid ${sColor()}50`,
              display: "flex",
              "align-items": "center",
              "justify-content": "center",
              color: sColor(),
            }}
          >
            ⚙️
          </div>
          <div style={{ flex: 1, "min-width": 0 }}>
            <div
              style={{
                "font-size": "0.95rem",
                "font-weight": "700",
                overflow: "hidden",
                "text-overflow": "ellipsis",
                "white-space": "nowrap",
              }}
            >
              {props.sensor.sensorId}
            </div>
            <div
              style={{
                display: "flex",
                gap: "10px",
                "align-items": "center",
                "margin-top": "2px",
              }}
            >
              <span style={{ "font-size": "0.62rem", color: "#94a3b8" }}>
                {props.sensor.sensorType}
              </span>
              <span
                style={{
                  "font-size": "0.62rem",
                  "font-weight": "700",
                  color: sColor(),
                  "text-transform": "uppercase",
                  "letter-spacing": "0.06em",
                }}
              >
                {props.sensor.status}
              </span>
            </div>
          </div>
        </div>

        <button
          data-testid="diagnostic-card-close-btn"
          onClick={() => props.onClose?.()}
          style={{
            background: "rgba(255,255,255,0.06)",
            border: "1px solid rgba(255,255,255,0.15)",
            color: "#f87171",
            cursor: "pointer",
            "border-radius": "8px",
            width: "28px",
            height: "28px",
            display: "flex",
            "align-items": "center",
            "justify-content": "center",
            "font-size": "0.8rem",
            "flex-shrink": 0,
            transition: "all 0.15s",
          }}
        >
          ✕
        </button>
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
                  background: "rgba(255,255,255,0.03)",
                  border: "1px solid rgba(255,255,255,0.06)",
                  "border-radius": "12px",
                  padding: "12px",
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
                        "font-size": "1.4rem",
                        "font-weight": "800",
                        color: healthColor(d().healthScore),
                      }}
                    >
                      {d().healthScore}
                    </div>
                    <div
                      style={{
                        "font-size": "0.6rem",
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
                  gap: "8px",
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
                  background: "rgba(255,255,255,0.02)",
                  border: "1px solid rgba(255,255,255,0.06)",
                  "border-radius": "12px",
                  padding: "10px",
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
                      "font-size": "0.72rem",
                      "text-transform": "uppercase",
                      color: "#94a3b8",
                      "letter-spacing": "0.08em",
                    }}
                  >
                    Diagnostic Event Timeline
                  </span>
                  <span style={{ "font-size": "0.6rem", color: "#475569" }}>
                    Recent
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
                      gap: "8px",
                    }}
                  >
                    <For each={timeline()}>
                      {(evt) => (
                        <div
                          style={{
                            display: "grid",
                            "grid-template-columns": "90px 1fr",
                            gap: "8px",
                            padding: "8px",
                            background: "rgba(255,255,255,0.02)",
                            border: "1px solid rgba(255,255,255,0.04)",
                            "border-radius": "10px",
                          }}
                        >
                          <span
                            style={{ color: "#64748b", "font-size": "0.62rem" }}
                          >
                            {new Date(evt.timeUtc).toLocaleTimeString([], {
                              hour: "2-digit",
                              minute: "2-digit",
                              second: "2-digit",
                            })}
                          </span>
                          <div
                            style={{
                              display: "flex",
                              "flex-direction": "column",
                              gap: "4px",
                            }}
                          >
                            <span
                              style={{
                                color: severityColor(evt.severity),
                                "font-weight": "700",
                                "font-size": "0.72rem",
                              }}
                            >
                              {evt.severity.toUpperCase()}
                            </span>
                            <span
                              style={{
                                color: "#e2e8f0",
                                "font-size": "0.75rem",
                                "line-height": "1.4",
                              }}
                            >
                              {evt.event}
                            </span>
                          </div>
                        </div>
                      )}
                    </For>
                  </div>
                </Show>
              </div>

              <div
                style={{
                  background: "rgba(255,255,255,0.02)",
                  border: "1px solid rgba(255,255,255,0.06)",
                  "border-radius": "12px",
                  padding: "10px",
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
                      "font-size": "0.72rem",
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
                            stroke={
                              seg.reason.includes("timestamp")
                                ? "#22d3ee"
                                : seg.reason.includes("id")
                                  ? "#60a5fa"
                                  : "#f59e0b"
                            }
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
                              "justify-content": "space-between",
                              gap: "10px",
                              "font-size": "0.68rem",
                              color: "#e2e8f0",
                            }}
                          >
                            <span>{seg.reason}</span>
                            <span style={{ color: "#94a3b8" }}>
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

            {/* Throughput + OPS */}
            <div
              style={{
                display: "grid",
                "grid-template-columns": "1fr 1fr",
                gap: "10px",
              }}
            >
              <div
                style={{
                  background: "rgba(255,255,255,0.02)",
                  border: "1px solid rgba(255,255,255,0.06)",
                  "border-radius": "12px",
                  padding: "10px",
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
                      "font-size": "0.72rem",
                      color: "#94a3b8",
                      "letter-spacing": "0.08em",
                      "text-transform": "uppercase",
                    }}
                  >
                    Throughput vs Expected
                  </span>
                  <span style={{ "font-size": "0.6rem", color: "#475569" }}>
                    {expectedRange().min}–{expectedRange().max} obs/s
                  </span>
                </div>
                {throughputBars() ?? (
                  <div style={{ color: "#334155", "font-size": "0.65rem" }}>
                    No throughput history
                  </div>
                )}
              </div>

              <div
                style={{
                  background: "rgba(255,255,255,0.02)",
                  border: "1px solid rgba(255,255,255,0.06)",
                  "border-radius": "12px",
                  padding: "10px",
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
                      "font-size": "0.72rem",
                      color: "#94a3b8",
                      "letter-spacing": "0.08em",
                      "text-transform": "uppercase",
                    }}
                  >
                    Observation Rate — last 15 min
                  </span>
                  <span style={{ "font-size": "0.6rem", color: "#475569" }}>
                    OPS
                  </span>
                </div>
                <ObsPerSecChart history={opsHistory()} height={100} />
              </div>
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
