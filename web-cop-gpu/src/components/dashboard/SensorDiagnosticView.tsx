// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/SensorDiagnosticView.tsx — Deep diagnostic view for a single sensor
//
// Reference: docs/business/usecases/UC017_sensor_health_monitoring.md
// Reference: docs/implementation/v5/ui_images/diagnostic_deep_dive.png

import { createResource, For, Show } from "solid-js";
import {
    fetchSensorDiagnostic,
    type SensorStatus,
} from "../../services/sensor-health";
import { setSelectedSensor } from "../../signals/sensor-filters";
import { MiniCoverageMap } from "./MiniCoverageMap";

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
    case "warn":
      return "⚠";
    case "error":
      return "✖";
    default:
      return "ℹ";
  }
}

function severityColor(sev: string): string {
  switch (sev) {
    case "warn":
      return "#fbbf24";
    case "error":
      return "#f87171";
    default:
      return "#94a3b8";
  }
}

function subStatusColor(s: string): string {
  switch (s) {
    case "ACTIVE":
      return "#4ade80";
    case "DEGRADED":
      return "#fbbf24";
    default:
      return "#f87171";
  }
}

function healthScoreColor(score: number): string {
  if (score >= 85) return "#4ade80";
  if (score >= 60) return "#fbbf24";
  return "#f87171";
}

function healthLabel(score: number): string {
  if (score >= 85) return "NOMINAL";
  if (score >= 60) return "DEGRADED";
  return "CRITICAL";
}

function pad3(n: number | null): string {
  return n === null ? "---" : String(n).padStart(3, "0");
}

/** Full-pane deep diagnostic view for a single sensor. Never destructure props. */
export function SensorDiagnosticView(props: SensorDiagnosticViewProps) {
  const [data] = createResource(() => props.sensor, fetchSensorDiagnostic);

  const cardStyle = {
    background: "rgba(13, 20, 40, 0.75)",
    "backdrop-filter": "blur(16px)",
    border: "1px solid rgba(255,255,255,0.07)",
    "border-radius": "12px",
    padding: "1.1rem 1.25rem",
  };

  const secHead = {
    "font-size": "0.68rem",
    "text-transform": "uppercase" as const,
    "letter-spacing": "0.1em",
    color: "#475569",
    "margin-bottom": "0.8rem",
    "font-family": "monospace",
  };

  return (
    <div
      data-testid="sensor-diagnostic-view"
      style={{
        display: "flex",
        "flex-direction": "column",
        height: "100%",
        overflow: "auto",
        padding: "18px 22px",
        gap: "14px",
        background:
          "radial-gradient(ellipse at 8% 4%, rgba(30,58,138,0.22) 0%, transparent 50%), radial-gradient(ellipse at 92% 96%, rgba(6,182,212,0.08) 0%, transparent 40%)",
        "font-family": "monospace",
      }}
    >
      {/* ── Header ───────────────────────────────────────────────────────── */}
      <div
        style={{
          display: "flex",
          "align-items": "center",
          gap: "10px",
          "flex-wrap": "wrap",
          background: "rgba(13,20,40,0.6)",
          border: "1px solid rgba(255,255,255,0.06)",
          "border-radius": "10px",
          padding: "10px 14px",
        }}
      >
        <button
          data-testid="diagnostic-back-btn"
          onClick={() => setSelectedSensor(null)}
          style={{
            background: "rgba(255,255,255,0.05)",
            border: "1px solid rgba(255,255,255,0.1)",
            color: "#94a3b8",
            padding: "4px 10px",
            "border-radius": "6px",
            cursor: "pointer",
            "font-size": "0.85rem",
            "font-family": "monospace",
          }}
        >
          «
        </button>

        <div style={{ color: "#475569", "font-size": "0.72rem" }}>
          Sensor Health /{" "}
          <span style={{ color: "#e2e8f0", "font-weight": "600" }}>
            {props.sensor.sensorId}
          </span>
        </div>

        {/* Type badge */}
        <div
          style={{
            background: `${sensorTypeBadgeColor(props.sensor.sensorType)}18`,
            color: sensorTypeBadgeColor(props.sensor.sensorType),
            border: `1px solid ${sensorTypeBadgeColor(props.sensor.sensorType)}35`,
            padding: "2px 10px",
            "border-radius": "20px",
            "font-size": "0.62rem",
            "text-transform": "uppercase",
            "letter-spacing": "0.07em",
          }}
        >
          {props.sensor.sensorType}
        </div>

        {/* Status badge */}
        <div
          style={{
            background: `${statusColor(props.sensor.status)}12`,
            color: statusColor(props.sensor.status),
            border: `1px solid ${statusColor(props.sensor.status)}28`,
            padding: "2px 10px",
            "border-radius": "20px",
            "font-size": "0.62rem",
            "text-transform": "uppercase",
            "letter-spacing": "0.07em",
            display: "flex",
            "align-items": "center",
            gap: "5px",
          }}
        >
          <span
            style={{
              width: "5px",
              height: "5px",
              "border-radius": "50%",
              background: statusColor(props.sensor.status),
              display: "inline-block",
            }}
          />
          {props.sensor.status}
        </div>

        {/* Quick inline stats */}
        <div
          style={{
            display: "flex",
            gap: "18px",
            "margin-left": "6px",
            "flex-wrap": "wrap",
            "font-size": "0.7rem",
          }}
        >
          <span style={{ color: "#64748b" }}>
            <span style={{ color: "#cbd5e1" }}>
              {props.sensor.eventsPerSecond}
            </span>{" "}
            obs/s
          </span>
          <span style={{ color: "#64748b" }}>
            DLQ:{" "}
            <span
              style={{
                color: props.sensor.dlqCount > 50 ? "#f87171" : "#cbd5e1",
              }}
            >
              {props.sensor.dlqCount}
            </span>
          </span>
          <span style={{ color: "#64748b" }}>
            Qual:{" "}
            <span
              style={{
                color:
                  props.sensor.validationPassRate > 95
                    ? "#4ade80"
                    : props.sensor.validationPassRate > 80
                      ? "#fbbf24"
                      : "#f87171",
              }}
            >
              {props.sensor.validationPassRate}%
            </span>
          </span>
          <span style={{ color: "#64748b" }}>
            Rcvd:{" "}
            <span style={{ color: "#cbd5e1" }}>
              {props.sensor.totalReceived.toLocaleString()}
            </span>
          </span>
        </div>

        <div
          style={{
            "margin-left": "auto",
            display: "flex",
            "align-items": "center",
            gap: "12px",
          }}
        >
          <Show when={props.sensor.coverage}>
            <MiniCoverageMap
              rangeNm={props.sensor.coverage!.rangeNm}
              bearingStart={props.sensor.coverage!.bearingStart}
              bearingEnd={props.sensor.coverage!.bearingEnd}
              alertLevel={props.sensor.dlqCount > 100 ? 2 : props.sensor.dlqCount > 50 ? 1 : 0}
              width={64}
              height={64}
            />
          </Show>
          <div
            style={{
              color: "#334155",
              "font-size": "0.67rem",
            }}
          >
            {new Date().toLocaleTimeString([], {
              hour: "2-digit",
              minute: "2-digit",
              second: "2-digit",
            })}
          </div>
        </div>
      </div>

      {/* ── Loading / error ───────────────────────────────────────────────── */}
      <Show when={data.loading}>
        <div
          style={{ color: "#64748b", padding: "40px", "text-align": "center" }}
        >
          Loading diagnostic data…
        </div>
      </Show>

      <Show when={data.error}>
        <div style={{ color: "#f87171", padding: "20px" }}>
          Error loading diagnostic:{" "}
          {(data.error as Error)?.message ?? "Unknown error"}
        </div>
      </Show>

      <Show when={data()}>
        {/* ── Two-column layout ─────────────────────────────────────────── */}
        <div
          style={{
            display: "grid",
            "grid-template-columns": "minmax(0, 56fr) minmax(0, 44fr)",
            gap: "14px",
            "align-items": "start",
          }}
        >
          {/* ═══ LEFT COLUMN ═════════════════════════════════════════════ */}
          <div
            style={{
              display: "flex",
              "flex-direction": "column",
              gap: "12px",
            }}
          >
            {/* KPI cards — 4-up row */}
            <div
              style={{
                display: "grid",
                "grid-template-columns": "repeat(4, 1fr)",
                gap: "10px",
              }}
            >
              {/* Throughput */}
              <div style={cardStyle}>
                <div style={secHead}>Throughput</div>
                <div
                  style={{
                    "font-size": "1.45rem",
                    "font-weight": "700",
                    color: "#f8fafc",
                    "line-height": "1",
                  }}
                >
                  {data()!.eventsPerSecond}
                </div>
                <div
                  style={{
                    "font-size": "0.62rem",
                    color: "#475569",
                    "margin-top": "3px",
                  }}
                >
                  obs/s
                </div>
                <div
                  style={{
                    "font-size": "0.6rem",
                    color: "#334155",
                    "margin-top": "5px",
                  }}
                >
                  Peak: {data()!.peakThroughput}
                </div>
              </div>

              {/* DLQ */}
              <div style={cardStyle}>
                <div style={secHead}>DLQ Count</div>
                <div
                  style={{
                    "font-size": "1.45rem",
                    "font-weight": "700",
                    "line-height": "1",
                    color: data()!.dlqCount > 50 ? "#f87171" : "#f8fafc",
                  }}
                >
                  {data()!.dlqCount}
                </div>
                <div
                  style={{
                    "font-size": "0.62rem",
                    color: "#475569",
                    "margin-top": "3px",
                  }}
                >
                  rejected
                </div>
                <div
                  style={{
                    "font-size": "0.6rem",
                    color: "#334155",
                    "margin-top": "5px",
                  }}
                >
                  {data()!.dlqBreakdown.length} reason(s)
                </div>
              </div>

              {/* Validation */}
              <div style={cardStyle}>
                <div style={secHead}>Validation</div>
                <div
                  style={{
                    "font-size": "1.45rem",
                    "font-weight": "700",
                    "line-height": "1",
                    color:
                      data()!.validationPassRate > 95
                        ? "#4ade80"
                        : data()!.validationPassRate > 80
                          ? "#fbbf24"
                          : "#f87171",
                  }}
                >
                  {data()!.validationPassRate}%
                </div>
                <div
                  style={{
                    "font-size": "0.62rem",
                    color: "#475569",
                    "margin-top": "3px",
                  }}
                >
                  pass rate
                </div>
                <div
                  style={{
                    "font-size": "0.6rem",
                    color: "#334155",
                    "margin-top": "5px",
                  }}
                >
                  {data()!.totalReceived.toLocaleString()} total
                </div>
              </div>

              {/* Latency */}
              <div style={cardStyle}>
                <div style={secHead}>Avg Latency</div>
                <div
                  style={{
                    "font-size": "1.45rem",
                    "font-weight": "700",
                    "line-height": "1",
                    color: data()!.avgLatencyMs > 500 ? "#f87171" : "#f8fafc",
                  }}
                >
                  {data()!.avgLatencyMs}
                </div>
                <div
                  style={{
                    "font-size": "0.62rem",
                    color: "#475569",
                    "margin-top": "3px",
                  }}
                >
                  ms
                </div>
                <div
                  style={{
                    "font-size": "0.6rem",
                    color: "#334155",
                    "margin-top": "5px",
                  }}
                >
                  {data()!.minLatencyMs}–{data()!.maxLatencyMs} ms
                </div>
              </div>
            </div>

            {/* Sensor Parameters */}
            <div style={cardStyle}>
              <div style={secHead}>Sensor Parameters</div>
              <div
                style={{
                  display: "grid",
                  "grid-template-columns": "1fr 1fr",
                  gap: "10px 20px",
                  "font-size": "0.78rem",
                }}
              >
                <Show when={data()!.rangeNm !== null}>
                  <div>
                    <div style={{ color: "#475569", "font-size": "0.62rem" }}>
                      Coverage Range
                    </div>
                    <div style={{ color: "#e2e8f0", "font-weight": "600" }}>
                      {data()!.rangeNm} NM
                    </div>
                  </div>
                </Show>

                <Show when={data()!.position !== null}>
                  <div>
                    <div style={{ color: "#475569", "font-size": "0.62rem" }}>
                      Position
                    </div>
                    <div style={{ color: "#e2e8f0", "font-size": "0.72rem" }}>
                      {data()!.position!.lat.toFixed(1)}°N{" "}
                      {Math.abs(data()!.position!.lon).toFixed(1)}°
                      {data()!.position!.lon < 0 ? "W" : "E"}
                    </div>
                  </div>
                </Show>

                <Show
                  when={
                    data()!.bearingStart !== null && data()!.bearingEnd !== null
                  }
                >
                  <div style={{ "margin-top": "16px", "margin-bottom": "8px", display: "flex", "justify-content": "center" }}>
                    <MiniCoverageMap
                      rangeNm={data()!.rangeNm || 120}
                      bearingStart={data()!.bearingStart!}
                      bearingEnd={data()!.bearingEnd!}
                      alertLevel={data()!.dlqCount > 100 ? 2 : data()!.dlqCount > 50 ? 1 : 0}
                      width={180}
                      height={180}
                    />
                  </div>
                  <div>
                    <div style={{ color: "#475569", "font-size": "0.62rem" }}>
                      Bearing Sector
                    </div>
                    <div style={{ color: "#e2e8f0", "font-weight": "600" }}>
                      {pad3(data()!.bearingStart)}°–{pad3(data()!.bearingEnd)}°
                    </div>
                  </div>
                </Show>

                <Show when={data()!.scanRateRpm !== null}>
                  <div>
                    <div style={{ color: "#475569", "font-size": "0.62rem" }}>
                      Scan Rate
                    </div>
                    <div style={{ color: "#e2e8f0", "font-weight": "600" }}>
                      {data()!.scanRateRpm} rpm
                    </div>
                  </div>
                </Show>

                <Show when={data()!.frequencyBandGhz !== null}>
                  <div>
                    <div style={{ color: "#475569", "font-size": "0.62rem" }}>
                      Frequency
                    </div>
                    <div style={{ color: "#e2e8f0", "font-weight": "600" }}>
                      {data()!.frequencyBandGhz} GHz
                    </div>
                  </div>
                </Show>

                <div>
                  <div style={{ color: "#475569", "font-size": "0.62rem" }}>
                    Last Seen
                  </div>
                  <div
                    style={{
                      color:
                        data()!.lastSeenSeconds < 0
                          ? "#f87171"
                          : data()!.lastSeenSeconds > 120
                            ? "#f87171"
                            : data()!.lastSeenSeconds > 30
                              ? "#fbbf24"
                              : "#e2e8f0",
                      "font-weight": "600",
                    }}
                  >
                    {data()!.lastSeenSeconds < 0
                      ? "N/A"
                      : `${data()!.lastSeenSeconds}s ago`}
                  </div>
                </div>
              </div>
            </div>

            {/* Throughput History sparkline */}
            <div style={cardStyle}>
              <div
                style={{
                  display: "flex",
                  "justify-content": "space-between",
                  "align-items": "center",
                  "margin-bottom": "0.75rem",
                }}
              >
                <div style={secHead}>Throughput History (20 samples)</div>
                <span style={{ "font-size": "0.62rem", color: "#334155" }}>
                  Peak {data()!.peakThroughput} obs/s
                </span>
              </div>
              <div style={{ height: "68px" }}>
                {(() => {
                  const pts = data()!.throughputHistory;
                  const max = Math.max(...pts, 1);
                  const W = 200;
                  const H = 58;
                  const polyline = pts
                    .map(
                      (v, i) =>
                        `${(i / (pts.length - 1)) * W},${H - (v / max) * H}`,
                    )
                    .join(" ");
                  const fill = `${polyline} ${W},${H} 0,${H}`;
                  const sc = statusColor(props.sensor.status);
                  const lastY = H - (pts[pts.length - 1] / max) * H;
                  return (
                    <svg
                      width="100%"
                      height="100%"
                      viewBox={`0 0 ${W} ${H}`}
                      preserveAspectRatio="none"
                    >
                      <defs>
                        <linearGradient
                          id="diag-grad"
                          x1="0%"
                          y1="0%"
                          x2="0%"
                          y2="100%"
                        >
                          <stop
                            offset="0%"
                            style={{
                              "stop-color": sc,
                              "stop-opacity": 0.38,
                            }}
                          />
                          <stop
                            offset="100%"
                            style={{
                              "stop-color": sc,
                              "stop-opacity": 0,
                            }}
                          />
                        </linearGradient>
                      </defs>
                      {/* Expected range band */}
                      <rect
                        x="0"
                        y={H * 0.1}
                        width={W}
                        height={H * 0.5}
                        fill="rgba(255,255,255,0.025)"
                        rx="2"
                      />
                      {/* 50% mid-line */}
                      <line
                        x1="0"
                        y1={H * 0.5}
                        x2={W}
                        y2={H * 0.5}
                        stroke="rgba(255,255,255,0.05)"
                        stroke-width="1"
                      />
                      {/* Area fill */}
                      <polygon points={fill} fill="url(#diag-grad)" />
                      {/* Line */}
                      <polyline
                        points={polyline}
                        fill="none"
                        stroke={sc}
                        stroke-width="1.5"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                      />
                      {/* Latest value dot */}
                      <circle cx={W} cy={lastY} r="3" fill={sc} />
                    </svg>
                  );
                })()}
              </div>
            </div>

            {/* Sub-Sensors table */}
            <Show when={data()!.subSensors.length > 0}>
              <div style={cardStyle}>
                <div style={secHead}>
                  Sub-Sensors ({data()!.subSensors.length})
                </div>
                <table
                  style={{
                    width: "100%",
                    "border-collapse": "collapse",
                    "font-size": "0.76rem",
                  }}
                >
                  <thead>
                    <tr style={{ color: "#475569", "text-align": "left" }}>
                      <th
                        style={{
                          padding: "5px 8px",
                          "border-bottom": "1px solid rgba(255,255,255,0.06)",
                        }}
                      >
                        Sub-Sensor ID
                      </th>
                      <th
                        style={{
                          padding: "5px 8px",
                          "border-bottom": "1px solid rgba(255,255,255,0.06)",
                        }}
                      >
                        Status
                      </th>
                      <th
                        style={{
                          padding: "5px 8px",
                          "border-bottom": "1px solid rgba(255,255,255,0.06)",
                        }}
                      >
                        Location
                      </th>
                      <th
                        style={{
                          padding: "5px 8px",
                          "border-bottom": "1px solid rgba(255,255,255,0.06)",
                        }}
                      >
                        Last Seen
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    <For each={data()!.subSensors}>
                      {(sub) => (
                        <tr
                          style={{
                            "border-bottom": "1px solid rgba(255,255,255,0.03)",
                          }}
                        >
                          <td style={{ padding: "5px 8px", color: "#e2e8f0" }}>
                            {sub.id}
                          </td>
                          <td style={{ padding: "5px 8px" }}>
                            <span
                              style={{
                                background: `${subStatusColor(sub.status)}14`,
                                color: subStatusColor(sub.status),
                                border: `1px solid ${subStatusColor(sub.status)}28`,
                                padding: "1px 7px",
                                "border-radius": "10px",
                                "font-size": "0.6rem",
                              }}
                            >
                              {sub.status}
                            </span>
                          </td>
                          <td style={{ padding: "5px 8px", color: "#94a3b8" }}>
                            {sub.location}
                          </td>
                          <td style={{ padding: "5px 8px", color: "#94a3b8" }}>
                            {sub.lastSeenSeconds < 0
                              ? "N/A"
                              : `${sub.lastSeenSeconds}s`}
                          </td>
                        </tr>
                      )}
                    </For>
                  </tbody>
                </table>
              </div>
            </Show>
          </div>

          {/* ═══ RIGHT COLUMN ════════════════════════════════════════════ */}
          <div
            style={{
              display: "flex",
              "flex-direction": "column",
              gap: "12px",
            }}
          >
            {/* Health Score Ring — dominant visual element */}
            <div
              style={{
                ...cardStyle,
                "text-align": "center",
                border: `1px solid ${healthScoreColor(data()!.healthScore)}20`,
              }}
            >
              <div style={secHead}>Overall Health Score</div>
              <div
                style={{
                  display: "flex",
                  "justify-content": "center",
                  "align-items": "center",
                  gap: "20px",
                  "flex-wrap": "wrap",
                }}
              >
                {/* SVG Gauge */}
                {(() => {
                  const score = data()!.healthScore;
                  const hColor = healthScoreColor(score);
                  const R = 75;
                  const CIRC = 2 * Math.PI * R;
                  const offset = CIRC * (1 - score / 100);
                  return (
                    <svg
                      width="180"
                      height="180"
                      viewBox="0 0 180 180"
                      style={{ "flex-shrink": "0" }}
                    >
                      <defs>
                        <filter
                          id="hglow"
                          x="-20%"
                          y="-20%"
                          width="140%"
                          height="140%"
                        >
                          <feGaussianBlur stdDeviation="3" result="blur" />
                          <feMerge>
                            <feMergeNode in="blur" />
                            <feMergeNode in="SourceGraphic" />
                          </feMerge>
                        </filter>
                        <filter
                          id="cglow"
                          x="-50%"
                          y="-50%"
                          width="200%"
                          height="200%"
                        >
                          <feGaussianBlur stdDeviation="4" result="blur" />
                          <feMerge>
                            <feMergeNode in="blur" />
                            <feMergeNode in="SourceGraphic" />
                          </feMerge>
                        </filter>
                      </defs>
                      {/* Outer track */}
                      <circle
                        cx="90"
                        cy="90"
                        r={R}
                        fill="none"
                        stroke="rgba(255,255,255,0.06)"
                        stroke-width="14"
                      />
                      {/* Health arc */}
                      <circle
                        cx="90"
                        cy="90"
                        r={R}
                        fill="none"
                        stroke={hColor}
                        stroke-width="14"
                        stroke-linecap="round"
                        stroke-dasharray={`${CIRC}`}
                        stroke-dashoffset={`${offset}`}
                        transform="rotate(-90 90 90)"
                        filter="url(#hglow)"
                      />
                      {/* Cyan dashed inner ring (matching reference image aesthetic) */}
                      <circle
                        cx="90"
                        cy="90"
                        r="59"
                        fill="none"
                        stroke="#22d3ee"
                        stroke-width="1.5"
                        stroke-dasharray="6 3"
                        opacity="0.35"
                        filter="url(#cglow)"
                      />
                      {/* Cyan inner glow disc */}
                      <circle
                        cx="90"
                        cy="90"
                        r="53"
                        fill="rgba(34,211,238,0.04)"
                      />
                      {/* Score value */}
                      <text
                        x="90"
                        y="84"
                        text-anchor="middle"
                        font-size="38"
                        font-weight="700"
                        fill={hColor}
                        font-family="monospace"
                      >
                        {score}
                      </text>
                      {/* Label */}
                      <text
                        x="90"
                        y="102"
                        text-anchor="middle"
                        font-size="8"
                        fill="#475569"
                        font-family="monospace"
                        letter-spacing="1.5"
                      >
                        HEALTH SCORE
                      </text>
                      {/* State label */}
                      <text
                        x="90"
                        y="117"
                        text-anchor="middle"
                        font-size="10"
                        fill={hColor}
                        font-family="monospace"
                        font-weight="600"
                      >
                        {healthLabel(score)}
                      </text>
                    </svg>
                  );
                })()}

                {/* Side KPIs */}
                <div
                  style={{
                    display: "flex",
                    "flex-direction": "column",
                    gap: "14px",
                    "text-align": "left",
                  }}
                >
                  <div>
                    <div style={{ "font-size": "0.6rem", color: "#475569" }}>
                      Uptime (1 hr)
                    </div>
                    <div
                      style={{
                        "font-size": "1.15rem",
                        "font-weight": "700",
                        color: "#22d3ee",
                      }}
                    >
                      {data()!.connectionUptimePct}%
                    </div>
                  </div>
                  <div>
                    <div style={{ "font-size": "0.6rem", color: "#475569" }}>
                      Total Received
                    </div>
                    <div
                      style={{
                        "font-size": "1.1rem",
                        "font-weight": "700",
                        color: "#e2e8f0",
                      }}
                    >
                      {data()!.totalReceived.toLocaleString()}
                    </div>
                  </div>
                  <div>
                    <div style={{ "font-size": "0.6rem", color: "#475569" }}>
                      Peak obs/s
                    </div>
                    <div
                      style={{
                        "font-size": "1.1rem",
                        "font-weight": "700",
                        color: "#e2e8f0",
                      }}
                    >
                      {data()!.peakThroughput}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Latency Distribution */}
            <div style={cardStyle}>
              <div style={secHead}>Latency Distribution</div>
              {(() => {
                const minL = data()!.minLatencyMs;
                const avgL = data()!.avgLatencyMs;
                const maxL = data()!.maxLatencyMs;
                const safe = maxL > 0 ? maxL : 1;
                const minPct = (minL / safe) * 100;
                const avgPct = (avgL / safe) * 100;
                const latColor =
                  avgL > 500 ? "#f87171" : avgL > 200 ? "#fbbf24" : "#22d3ee";
                return (
                  <div>
                    <div
                      style={{
                        position: "relative",
                        height: "18px",
                        "border-radius": "5px",
                        background: "rgba(255,255,255,0.05)",
                        "margin-bottom": "7px",
                      }}
                    >
                      {/* Min→avg range */}
                      <div
                        style={{
                          position: "absolute",
                          left: `${minPct}%`,
                          width: `${Math.max(avgPct - minPct, 2)}%`,
                          height: "100%",
                          background: `${latColor}25`,
                          "border-radius": "4px",
                        }}
                      />
                      {/* Avg→max range */}
                      <div
                        style={{
                          position: "absolute",
                          left: `${avgPct}%`,
                          width: `${100 - avgPct}%`,
                          height: "100%",
                          background: "rgba(248,113,113,0.07)",
                          "border-radius": "4px",
                        }}
                      />
                      {/* Avg marker */}
                      <div
                        style={{
                          position: "absolute",
                          left: `${avgPct}%`,
                          width: "2px",
                          height: "100%",
                          background: latColor,
                          "border-radius": "1px",
                        }}
                      />
                    </div>
                    <div
                      style={{
                        display: "flex",
                        "justify-content": "space-between",
                        "font-size": "0.68rem",
                      }}
                    >
                      <span>
                        <span style={{ color: "#475569" }}>Min </span>
                        <span style={{ color: "#cbd5e1" }}>{minL} ms</span>
                      </span>
                      <span>
                        <span style={{ color: "#475569" }}>Avg </span>
                        <span style={{ color: latColor }}>{avgL} ms</span>
                      </span>
                      <span>
                        <span style={{ color: "#475569" }}>Max </span>
                        <span style={{ color: "#f87171" }}>{maxL} ms</span>
                      </span>
                    </div>
                  </div>
                );
              })()}
            </div>

            {/* Status History Strip */}
            <div style={cardStyle}>
              <div style={secHead}>Status History (12 min)</div>
              <div
                style={{
                  display: "flex",
                  gap: "3px",
                  height: "26px",
                }}
              >
                <For each={data()!.statusHistory}>
                  {(h) => (
                    <div
                      title={`${new Date(h.timeUtc).toLocaleTimeString()}: ${h.status}`}
                      style={{
                        flex: "1",
                        "border-radius": "3px",
                        background: statusColor(h.status),
                        opacity:
                          h.status === "OFFLINE"
                            ? "0.9"
                            : h.status === "STALE"
                              ? "0.75"
                              : "0.55",
                      }}
                    />
                  )}
                </For>
              </div>
              <div
                style={{
                  display: "flex",
                  "justify-content": "space-between",
                  "margin-top": "4px",
                  "font-size": "0.6rem",
                  color: "#334155",
                }}
              >
                <span>12 min ago</span>
                <span>Now</span>
              </div>
            </div>

            {/* Connection Uptime bar */}
            <div style={cardStyle}>
              <div
                style={{
                  display: "flex",
                  "justify-content": "space-between",
                  "align-items": "baseline",
                  "margin-bottom": "8px",
                }}
              >
                <div style={secHead}>Connection Uptime</div>
                <span
                  style={{
                    "font-size": "0.85rem",
                    "font-weight": "700",
                    color: "#22d3ee",
                  }}
                >
                  {data()!.connectionUptimePct}%
                </span>
              </div>
              <div
                style={{
                  height: "7px",
                  "border-radius": "4px",
                  background: "rgba(255,255,255,0.05)",
                }}
              >
                <div
                  style={{
                    height: "100%",
                    width: `${data()!.connectionUptimePct}%`,
                    "border-radius": "4px",
                    background: "linear-gradient(90deg, #22d3ee, #3b82f6)",
                  }}
                />
              </div>
              <div
                style={{
                  display: "flex",
                  "justify-content": "space-between",
                  "margin-top": "4px",
                  "font-size": "0.6rem",
                  color: "#334155",
                }}
              >
                <span>Last hour</span>
                <span>
                  {Math.round((data()!.connectionUptimePct / 100) * 60)} / 60
                  min connected
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* ── DLQ Breakdown — full width ──────────────────────────────────── */}
        <Show when={data()!.dlqBreakdown.length > 0}>
          <div style={cardStyle}>
            <div style={secHead}>DLQ Breakdown</div>
            {(() => {
              const maxCount = Math.max(
                ...data()!.dlqBreakdown.map((d) => d.count),
                1,
              );
              return (
                <div
                  style={{
                    display: "flex",
                    "flex-direction": "column",
                    gap: "7px",
                  }}
                >
                  <For each={data()!.dlqBreakdown}>
                    {(d) => {
                      const pct = (d.count / maxCount) * 100;
                      const barColor =
                        pct > 60 ? "#f87171" : pct > 30 ? "#fbbf24" : "#60a5fa";
                      return (
                        <div
                          style={{
                            display: "grid",
                            "grid-template-columns": "220px 48px 1fr",
                            "align-items": "center",
                            gap: "10px",
                            "font-size": "0.76rem",
                          }}
                        >
                          <span
                            style={{
                              color: "#e2e8f0",
                              "font-family": "monospace",
                            }}
                          >
                            {d.reason}
                          </span>
                          <span
                            style={{
                              color: "#f87171",
                              "text-align": "right",
                            }}
                          >
                            {d.count}
                          </span>
                          <div
                            style={{
                              position: "relative",
                              height: "9px",
                              background: "rgba(255,255,255,0.04)",
                              "border-radius": "4px",
                            }}
                          >
                            <div
                              style={{
                                width: `${pct}%`,
                                "min-width": "4px",
                                height: "100%",
                                background: barColor,
                                "border-radius": "4px",
                                opacity: "0.85",
                              }}
                            />
                          </div>
                        </div>
                      );
                    }}
                  </For>
                </div>
              );
            })()}
          </div>
        </Show>

        {/* ── Recent Events — full width ──────────────────────────────────── */}
        <div style={cardStyle}>
          <div style={secHead}>Recent Events</div>
          <div
            style={{ display: "flex", "flex-direction": "column", gap: "5px" }}
          >
            <For each={data()!.recentEvents}>
              {(ev) => (
                <div
                  style={{
                    display: "grid",
                    "grid-template-columns": "72px 18px 1fr",
                    "align-items": "flex-start",
                    gap: "10px",
                    "border-bottom": "1px solid rgba(255,255,255,0.03)",
                    "padding-bottom": "5px",
                  }}
                >
                  <span
                    style={{
                      color: "#334155",
                      "font-size": "0.66rem",
                      "white-space": "nowrap",
                      "margin-top": "1px",
                    }}
                  >
                    {new Date(ev.timeUtc).toLocaleTimeString([], {
                      hour: "2-digit",
                      minute: "2-digit",
                      second: "2-digit",
                    })}
                  </span>
                  <span
                    style={{
                      color: severityColor(ev.severity),
                      "font-size": "0.82rem",
                    }}
                  >
                    {severityIcon(ev.severity)}
                  </span>
                  <span style={{ color: "#cbd5e1", "font-size": "0.76rem" }}>
                    {ev.event}
                  </span>
                </div>
              )}
            </For>
          </div>
        </div>
      </Show>
    </div>
  );
}
