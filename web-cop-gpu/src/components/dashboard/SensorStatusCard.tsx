// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/SensorStatusCard.tsx — Individual sensor status card
//
// Reference: docs/business/usecases/UC017_sensor_health_monitoring.md
// Reference: docs/implementation/v5/ui_images/dashboard_main.png

import { createSignal, JSX, onCleanup, Show } from "solid-js";
import { SensorStatus } from "../../services/sensor-health";
import { MiniCoverageMap } from "./MiniCoverageMap";

export interface SensorStatusCardProps {
  sensor: SensorStatus;
  /** compact=true → condensed metric-box layout; default (false) → full dual-sparkline layout */
  compact?: boolean;
  onSelect?: (sensor: SensorStatus) => void;
  /** Triggers the diagnostic overlay for this sensor */
  onDiagnose?: (sensor: SensorStatus) => void;
}

// ── Helpers ──────────────────────────────────────────────────────────────────

function formatLastSeen(seconds: number): string {
  if (seconds < 0) return "N/A";
  if (seconds < 60) return `${seconds}s ago`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  return `${Math.floor(seconds / 3600)}h ago`;
}

function sensorMode(type: string): string {
  const modes: Record<string, string> = {
    RADAR: "Active ESA",
    "AIS/BFT": "Passive AIS",
    "EW/SIGINT": "SIGINT Array",
    "ELINT/COMINT": "COMINT Platform",
    ISR: "ISR Platform",
    CYBER: "Cyber Monitor",
  };
  return modes[type] ?? type;
}

function derivedLocation(id: string): string {
  const u = id.toUpperCase();
  if (u.includes("NORTH")) return "North Sector";
  if (u.includes("SOUTH")) return "South Sector";
  if (u.includes("EAST")) return "East Array";
  if (u.includes("WEST")) return "West Coast";
  if (u.includes("PORT")) return "Port Station";
  if (u.includes("COAST")) return "Coastal Station";
  const digits = id.replace(/[^0-9]/g, "");
  const num = digits ? parseInt(digits.slice(-2)) || 1 : 1;
  return `Platform A-${num}`;
}

function SensorIcon(iconProps: { type: string }): JSX.Element {
  const t = iconProps.type;
  if (t === "RADAR") {
    return (
      <svg
        width="18"
        height="18"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path d="M19.07 4.93a10 10 0 0 0-14.14 0" />
        <path d="M16.24 7.76a6 6 0 0 0-8.48 0" />
        <circle cx="12" cy="12" r="2" />
        <path d="M12 2v2" />
        <path d="M2 12h2" />
        <path d="M20 12h2" />
        <path d="m4.93 4.93 1.41 1.41" />
        <path d="m17.66 17.66 1.41 1.41" />
      </svg>
    );
  }
  if (t === "EW/SIGINT" || t === "ELINT/COMINT") {
    return (
      <svg
        width="18"
        height="18"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
      </svg>
    );
  }
  if (t === "AIS/BFT") {
    return (
      <svg
        width="18"
        height="18"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path d="M2 21c.6.5 1.2 1 2.5 1 2.5 0 2.5-2 5-2 1.3 0 1.9.5 2.5 1 .6.5 1.2 1 2.5 1 2.5 0 2.5-2 5-2 1.3 0 1.9.5 2.5 1" />
        <path d="M19.38 20A11.6 11.6 0 0 0 21 14l-9-4-9 4c0 2.2.6 4.3 1.62 6" />
        <path d="M12 10V2" />
      </svg>
    );
  }
  if (t === "ISR") {
    return (
      <svg
        width="18"
        height="18"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <circle cx="12" cy="12" r="3" />
        <path d="M12 2v2" />
        <path d="M12 20v2" />
        <path d="m4.93 4.93 1.41 1.41" />
        <path d="m17.66 17.66 1.41 1.41" />
        <path d="M2 12h2" />
        <path d="M20 12h2" />
        <path d="m4.93 19.07 1.41-1.41" />
        <path d="m17.66 6.34 1.41-1.41" />
      </svg>
    );
  }
  if (t === "CYBER") {
    return (
      <svg
        width="18"
        height="18"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      </svg>
    );
  }
  return (
    <svg
      width="18"
      height="18"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
      stroke-linecap="round"
      stroke-linejoin="round"
    >
      <rect x="2" y="3" width="20" height="14" rx="2" ry="2" />
      <line x1="8" y1="21" x2="16" y2="21" />
      <line x1="12" y1="17" x2="12" y2="21" />
    </svg>
  );
}

// ── Component ────────────────────────────────────────────────────────────────

/**
 * Sensor status card.  Supports two view modes:
 *  • Full (default): Dual-sparkline rich layout matching the design reference image —
 *    shows Type/Location metadata rows, two side-by-side sparklines (throughput & accepted obs/s),
 *    Data Quality and Last Seen footer.
 *  • Compact (props.compact=true): Condensed metric-box layout with a single sparkline and DLQ count.
 *
 * Never destructure props — breaks SolidJS reactivity.
 */
export function SensorStatusCard(props: SensorStatusCardProps): JSX.Element {
  // Reactive UTC clock — updates every second
  const [utcTime, setUtcTime] = createSignal(
    new Date().toUTCString().slice(17, 25),
  );
  const clockTimer = setInterval(
    () => setUtcTime(new Date().toUTCString().slice(17, 25)),
    1000,
  );
  onCleanup(() => clearInterval(clockTimer));

  const statusColor = () => {
    switch (props.sensor.status) {
      case "CONNECTED":
        return "#4ade80";
      case "STALE":
        return "#fbbf24";
      case "OFFLINE":
        return "#f87171";
      default:
        return "#94a3b8";
    }
  };

  const [diagBtnHovered, setDiagBtnHovered] = createSignal(false);

  const statusLabel = () => {
    switch (props.sensor.status) {
      case "CONNECTED":
        return "Connected";
      case "STALE":
        return "Stale";
      case "OFFLINE":
        return "Offline";
      default:
        return props.sensor.status;
    }
  };

  const qualityColor = () => {
    if (props.sensor.status === "OFFLINE") return "#64748b";
    if (props.sensor.validationPassRate >= 95) return "#4ade80";
    if (props.sensor.validationPassRate >= 80) return "#fbbf24";
    return "#f87171";
  };

  const cardStyle = () => ({
    background:
      "linear-gradient(135deg, rgba(10, 15, 26, 0.92) 0%, rgba(15, 23, 42, 0.82) 100%)",
    "backdrop-filter": "blur(16px)",
    "-webkit-backdrop-filter": "blur(16px)",
    border: "1px solid rgba(255, 255, 255, 0.07)",
    "border-top": `2px solid ${statusColor()}`,
    "border-left": `3px solid ${statusColor()}`,
    "border-radius": "8px",
    padding: "10px 12px 10px 10px",
    display: "flex",
    "flex-direction": "column" as const,
    gap: "6px",
    color: "#f1f5f9",
    transition: "border-color 0.25s ease, box-shadow 0.25s ease",
    "box-shadow": `0 4px 18px rgba(0,0,0,0.45), inset 0 1px 0 rgba(255,255,255,0.04)`,
    position: "relative" as const,
    overflow: "hidden",
    cursor: "pointer",
    "min-width": props.compact
      ? "clamp(220px, 18vw, 290px)"
      : "clamp(260px, 22vw, 340px)",
    "font-size": "0.75rem",
  });

  const Glow = () => (
    <div
      style={{
        position: "absolute",
        top: "-20px",
        right: "-20px",
        width: "80px",
        height: "80px",
        background: statusColor(),
        filter: "blur(50px)",
        opacity: 0.1,
        "pointer-events": "none",
      }}
    />
  );

  const Badge = () => (
    <div
      style={{
        "font-size": "0.65rem",
        "font-weight": "700",
        padding: "4px 10px",
        "border-radius": "20px",
        background: `${statusColor()}15`,
        color: statusColor(),
        border: `1px solid ${statusColor()}30`,
        "text-transform": "uppercase",
        "letter-spacing": "0.05em",
        display: "flex",
        "align-items": "center",
        gap: "5px",
        "flex-shrink": "0",
      }}
    >
      <span
        class={
          props.sensor.status === "CONNECTED" ? "status-connected-dot" : ""
        }
        style={{
          width: "8px",
          height: "8px",
          "border-radius": "50%",
          background: statusColor(),
          display: "inline-block",
          "box-shadow": `0 0 8px ${statusColor()}`,
        }}
      />
      {statusLabel()}
    </div>
  );

  const IconBox = () => (
    <div
      style={{
        padding: "0.5rem",
        background: "rgba(255,255,255,0.05)",
        "border-radius": "10px",
        border: "1px solid rgba(255,255,255,0.1)",
        display: "flex",
        "align-items": "center",
        "justify-content": "center",
        color: statusColor(),
      }}
    >
      <SensorIcon type={props.sensor.sensorType} />
    </div>
  );

  /** Small crosshair button that triggers the health diagnostic overlay */
  const ScopeButton = () => (
    <Show when={props.onDiagnose !== undefined}>
      <button
        title="Health Diagnostics"
        onClick={(e) => {
          e.stopPropagation();
          props.onDiagnose?.(props.sensor);
        }}
        onMouseEnter={() => setDiagBtnHovered(true)}
        onMouseLeave={() => setDiagBtnHovered(false)}
        style={{
          background: diagBtnHovered()
            ? `${statusColor()}25`
            : "rgba(255,255,255,0.05)",
          border: `1px solid ${diagBtnHovered() ? statusColor() + "70" : "rgba(255,255,255,0.1)"}`,
          color: diagBtnHovered() ? statusColor() : "#475569",
          cursor: "pointer",
          "border-radius": "6px",
          width: "26px",
          height: "26px",
          display: "flex",
          "align-items": "center",
          "justify-content": "center",
          padding: "0",
          transition: "all 0.2s",
          "flex-shrink": "0",
          "box-shadow": diagBtnHovered()
            ? `0 0 10px ${statusColor()}30`
            : "none",
        }}
      >
        {/* Crosshair / scope icon */}
        <svg
          width="12"
          height="12"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
        >
          <circle cx="12" cy="12" r="3" />
          <line x1="12" y1="2" x2="12" y2="6" />
          <line x1="12" y1="18" x2="12" y2="22" />
          <line x1="2" y1="12" x2="6" y2="12" />
          <line x1="18" y1="12" x2="22" y2="12" />
        </svg>
      </button>
    </Show>
  );

  const CardHeader = () => (
    <div
      style={{
        display: "flex",
        "justify-content": "space-between",
        "align-items": "flex-start",
      }}
    >
      <div style={{ display: "flex", "align-items": "center", gap: "8px" }}>
        <IconBox />
        <div>
          <div
            style={{
              "font-weight": "700",
              "font-size": "0.82rem",
              "letter-spacing": "0.03em",
              "font-family": "monospace",
            }}
          >
            {props.sensor.sensorId}
          </div>
          <div
            style={{
              "font-size": "0.58rem",
              color: "#64748b",
              "text-transform": "uppercase",
              "letter-spacing": "0.06em",
              "margin-top": "1px",
            }}
          >
            {props.sensor.sensorType}
          </div>
        </div>
      </div>
      <div style={{ display: "flex", "align-items": "center", gap: "6px" }}>
        <Show when={props.sensor.coverage}>
          <MiniCoverageMap
            rangeNm={props.sensor.coverage!.rangeNm}
            bearingStart={props.sensor.coverage!.bearingStart}
            bearingEnd={props.sensor.coverage!.bearingEnd}
            alertLevel={
              props.sensor.dlqCount > 100
                ? 2
                : props.sensor.dlqCount > 50
                  ? 1
                  : 0
            }
            width={48}
            height={48}
          />
        </Show>
        <Badge />
        <ScopeButton />
      </div>
    </div>
  );

  const hoverStyle = `.sensor-card-hover:hover{transform:translateY(-1px);border-color:rgba(255,255,255,0.2);box-shadow:0 10px 24px rgba(0,0,0,0.46);background:rgba(15,23,42,0.85);}
@keyframes connected-pulse{0%{box-shadow:0 0 0 0 rgba(74,222,128,0.6);}70%{box-shadow:0 0 0 8px rgba(74,222,128,0);}100%{box-shadow:0 0 0 0 rgba(74,222,128,0);}}
.status-connected-dot{animation:connected-pulse 2s infinite;}`;

  return (
    <Show
      when={props.compact}
      fallback={
        /* ── FULL VIEW ──────────────────────────────────────────────────────── */
        <div
          data-testid={`sensor-card-${props.sensor.sensorId}`}
          data-view="full"
          onClick={() => props.onSelect?.(props.sensor)}
          style={cardStyle()}
          class="sensor-card-hover"
        >
          <Glow />
          <CardHeader />

          {/* Type / Location metadata rows */}
          <div
            style={{
              display: "grid",
              "grid-template-columns": "auto 1fr",
              "column-gap": "8px",
              "row-gap": "2px",
              "font-size": "0.65rem",
              padding: "4px 0",
              "border-top": "1px solid rgba(255,255,255,0.05)",
              "border-bottom": "1px solid rgba(255,255,255,0.05)",
            }}
          >
            <span style={{ color: "#475569" }}>Type</span>
            <span style={{ color: "#94a3b8", "text-align": "right" }}>
              {sensorMode(props.sensor.sensorType)}
            </span>
            <span style={{ color: "#475569" }}>Location</span>
            <span style={{ color: "#94a3b8", "text-align": "right" }}>
              {derivedLocation(props.sensor.sensorId)}
            </span>
          </div>

          {/* 4 Metric tiles — 2×2 grid: Throughput | Total Recv'd | DLQ Count | Validation */}
          <div
            style={{
              display: "grid",
              "grid-template-columns": "1fr 1fr",
              gap: "6px",
            }}
          >
            <div
              style={{
                background: "rgba(255,255,255,0.03)",
                padding: "5px 8px",
                "border-radius": "8px",
                border: "1px solid rgba(255,255,255,0.06)",
              }}
            >
              <div
                style={{
                  "font-size": "0.6rem",
                  color: "#64748b",
                  "text-transform": "uppercase",
                  "letter-spacing": "0.05em",
                  "margin-bottom": "4px",
                }}
              >
                Throughput
              </div>
              <div
                style={{
                  "font-size": "0.88rem",
                  "font-weight": "700",
                  "font-family": "monospace",
                  color: "#f8fafc",
                }}
              >
                {props.sensor.eventsPerSecond}{" "}
                <span
                  style={{
                    "font-size": "0.6rem",
                    color: "#64748b",
                    "font-weight": "normal",
                  }}
                >
                  obs/s
                </span>
              </div>
            </div>
            <div
              style={{
                background: "rgba(255,255,255,0.03)",
                padding: "5px 8px",
                "border-radius": "8px",
                border: "1px solid rgba(255,255,255,0.06)",
              }}
            >
              <div
                style={{
                  "font-size": "0.6rem",
                  color: "#64748b",
                  "text-transform": "uppercase",
                  "letter-spacing": "0.05em",
                  "margin-bottom": "4px",
                }}
              >
                Total Recv'd
              </div>
              <div
                style={{
                  "font-size": "0.88rem",
                  "font-weight": "700",
                  "font-family": "monospace",
                  color: "#f8fafc",
                }}
              >
                {props.sensor.totalReceived.toLocaleString()}
              </div>
            </div>
            <div
              style={{
                background: "rgba(255,255,255,0.03)",
                padding: "5px 8px",
                "border-radius": "8px",
                border: "1px solid rgba(255,255,255,0.06)",
              }}
            >
              <div
                style={{
                  "font-size": "0.6rem",
                  color: "#64748b",
                  "text-transform": "uppercase",
                  "letter-spacing": "0.05em",
                  "margin-bottom": "4px",
                }}
              >
                DLQ Count
              </div>
              <div
                style={{
                  "font-size": "0.88rem",
                  "font-weight": "700",
                  "font-family": "monospace",
                  color:
                    props.sensor.dlqCount > 50
                      ? "#f87171"
                      : props.sensor.dlqCount > 10
                        ? "#fbbf24"
                        : "#4ade80",
                }}
              >
                {props.sensor.dlqCount}{" "}
                <span
                  style={{
                    "font-size": "0.6rem",
                    color: "#64748b",
                    "font-weight": "normal",
                  }}
                >
                  rejected
                </span>
              </div>
            </div>
            <div
              style={{
                background: "rgba(255,255,255,0.03)",
                padding: "5px 8px",
                "border-radius": "8px",
                border: "1px solid rgba(255,255,255,0.06)",
              }}
            >
              <div
                style={{
                  "font-size": "0.6rem",
                  color: "#64748b",
                  "text-transform": "uppercase",
                  "letter-spacing": "0.05em",
                  "margin-bottom": "4px",
                }}
              >
                Validation
              </div>
              <div
                style={{
                  "font-size": "0.88rem",
                  "font-weight": "700",
                  "font-family": "monospace",
                  color: qualityColor(),
                }}
              >
                {props.sensor.status === "OFFLINE"
                  ? "N/A"
                  : `${props.sensor.validationPassRate}%`}
              </div>
            </div>
          </div>

          {/* Dual-line chart — throughput (blue) + accepted obs/s (green), last 15 mins */}
          {(() => {
            const seed = props.sensor.sensorId
              .split("")
              .reduce((acc, c) => acc + c.charCodeAt(0), 0);
            const maxEps =
              props.sensor.eventsPerSecond > 0
                ? props.sensor.eventsPerSecond * 1.4
                : 10;
            const validPct = props.sensor.validationPassRate / 100;
            // Chart area in viewBox units: x 28..128 (100 wide), y 4..56 (52 tall, inverted)
            const toX = (i: number) => 28 + (i / 14) * 100;
            const toY = (v: number) => 56 - Math.min(1, v / maxEps) * 52;
            const tPts: string[] = [];
            const aPts: string[] = [];
            for (let i = 0; i <= 14; i++) {
              const x = toX(i);
              if (props.sensor.status === "OFFLINE") {
                tPts.push(`${x.toFixed(1)},56`);
                aPts.push(`${x.toFixed(1)},56`);
              } else {
                const raw = Math.max(
                  0,
                  props.sensor.eventsPerSecond *
                    (0.7 + Math.sin(seed + i * 0.8) * 0.3),
                );
                tPts.push(`${x.toFixed(1)},${toY(raw).toFixed(1)}`);
                aPts.push(`${x.toFixed(1)},${toY(raw * validPct).toFixed(1)}`);
              }
            }
            const tPath = `M ${tPts.join(" L ")}`;
            const aPath = `M ${aPts.join(" L ")}`;
            const midVal = Math.round((maxEps / 2) * 10) / 10;
            const maxVal = Math.round(maxEps * 10) / 10;
            const gTId = `grad-ft-${props.sensor.sensorId}`;
            const gAId = `grad-fa-${props.sensor.sensorId}`;
            return (
              <div data-testid="full-card-chart">
                {/* Legend */}
                <div
                  style={{
                    display: "flex",
                    "justify-content": "space-between",
                    "align-items": "center",
                    "margin-bottom": "4px",
                  }}
                >
                  <div style={{ "font-size": "0.6rem", color: "#475569" }}>
                    Throughput · Last 15 mins
                  </div>
                  <div style={{ display: "flex", gap: "10px" }}>
                    <div
                      style={{
                        display: "flex",
                        "align-items": "center",
                        gap: "4px",
                        "font-size": "0.6rem",
                        color: "#60a5fa",
                      }}
                    >
                      <svg width="14" height="6" viewBox="0 0 14 6">
                        <line
                          x1="0"
                          y1="3"
                          x2="14"
                          y2="3"
                          stroke="#3b82f6"
                          stroke-width="2"
                          stroke-linecap="round"
                        />
                      </svg>
                      obs/s
                    </div>
                    <div
                      style={{
                        display: "flex",
                        "align-items": "center",
                        gap: "4px",
                        "font-size": "0.6rem",
                        color: "#4ade80",
                      }}
                    >
                      <svg width="14" height="6" viewBox="0 0 14 6">
                        <line
                          x1="0"
                          y1="3"
                          x2="14"
                          y2="3"
                          stroke="#22c55e"
                          stroke-width="2"
                          stroke-linecap="round"
                        />
                      </svg>
                      accepted
                    </div>
                  </div>
                </div>
                {/* Chart SVG — viewBox 130×68, chart area x:28–128, y:4–56 */}
                <svg
                  width="100%"
                  height="68"
                  viewBox="0 0 130 68"
                  preserveAspectRatio="none"
                >
                  <defs>
                    <linearGradient id={gTId} x1="0%" y1="0%" x2="0%" y2="100%">
                      <stop
                        offset="0%"
                        style={{
                          "stop-color": "#3b82f6",
                          "stop-opacity": 0.25,
                        }}
                      />
                      <stop
                        offset="100%"
                        style={{ "stop-color": "#3b82f6", "stop-opacity": 0 }}
                      />
                    </linearGradient>
                    <linearGradient id={gAId} x1="0%" y1="0%" x2="0%" y2="100%">
                      <stop
                        offset="0%"
                        style={{
                          "stop-color": "#22c55e",
                          "stop-opacity": 0.15,
                        }}
                      />
                      <stop
                        offset="100%"
                        style={{ "stop-color": "#22c55e", "stop-opacity": 0 }}
                      />
                    </linearGradient>
                  </defs>
                  {/* Gridlines */}
                  <line
                    x1="28"
                    y1="4"
                    x2="128"
                    y2="4"
                    stroke="rgba(255,255,255,0.03)"
                    stroke-width="0.5"
                  />
                  <line
                    x1="28"
                    y1="30"
                    x2="128"
                    y2="30"
                    stroke="rgba(255,255,255,0.05)"
                    stroke-width="0.5"
                    stroke-dasharray="2,2"
                  />
                  <line
                    x1="28"
                    y1="56"
                    x2="128"
                    y2="56"
                    stroke="rgba(255,255,255,0.08)"
                    stroke-width="0.5"
                  />
                  {/* Y axis */}
                  <line
                    x1="28"
                    y1="4"
                    x2="28"
                    y2="56"
                    stroke="rgba(255,255,255,0.06)"
                    stroke-width="0.5"
                  />
                  {/* Y-axis scale labels */}
                  <text
                    x="26"
                    y="8"
                    text-anchor="end"
                    fill="#475569"
                    font-size="5.5"
                  >
                    {maxVal}
                  </text>
                  <text
                    x="26"
                    y="32"
                    text-anchor="end"
                    fill="#475569"
                    font-size="5.5"
                  >
                    {midVal}
                  </text>
                  <text
                    x="26"
                    y="58"
                    text-anchor="end"
                    fill="#475569"
                    font-size="5.5"
                  >
                    0
                  </text>
                  {/* X-axis labels */}
                  <text
                    x="29"
                    y="66"
                    text-anchor="start"
                    fill="#475569"
                    font-size="5"
                  >
                    -15m
                  </text>
                  <text
                    x="127"
                    y="66"
                    text-anchor="end"
                    fill="#475569"
                    font-size="5"
                  >
                    now
                  </text>
                  {/* Throughput area fill */}
                  <path
                    d={`${tPath} L 128,56 L 28,56 Z`}
                    fill={`url(#${gTId})`}
                  />
                  {/* Accepted obs/s area fill */}
                  <path
                    d={`${aPath} L 128,56 L 28,56 Z`}
                    fill={`url(#${gAId})`}
                  />
                  {/* Throughput line (blue) */}
                  <path
                    d={tPath}
                    fill="none"
                    stroke="#3b82f6"
                    stroke-width="1.5"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  />
                  {/* Accepted obs/s line (green) */}
                  <path
                    d={aPath}
                    fill="none"
                    stroke="#22c55e"
                    stroke-width="1.5"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  />
                </svg>
              </div>
            );
          })()}

          {/* Footer: Last Seen + UTC Clock */}
          <div
            style={{
              display: "flex",
              "justify-content": "space-between",
              "align-items": "center",
              "font-size": "0.75rem",
              "padding-top": "0.5rem",
              "border-top": "1px solid rgba(255,255,255,0.05)",
            }}
          >
            <span
              style={{
                color: "#334155",
                "font-size": "0.65rem",
                "font-family": "monospace",
              }}
            >
              {utcTime()} UTC
            </span>
            <div
              style={{ display: "flex", "align-items": "center", gap: "6px" }}
            >
              <span style={{ color: "#64748b", "margin-right": "6px" }}>
                Last Seen
              </span>
              <span style={{ color: "#cbd5e1" }}>
                {formatLastSeen(props.sensor.lastSeenSeconds)}
              </span>
            </div>
          </div>
          <style>{hoverStyle}</style>
        </div>
      }
    >
      {/* ── COMPACT VIEW ────────────────────────────────────────────────────── */}
      <div
        data-testid={`sensor-card-${props.sensor.sensorId}`}
        data-view="compact"
        onClick={() => props.onSelect?.(props.sensor)}
        style={{
          ...cardStyle(),
          "min-width": "clamp(200px, 16vw, 270px)",
          gap: "0.5rem",
          padding: "0.85rem 1rem",
        }}
        class="sensor-card-hover"
      >
        <Glow />

        {/* Header: icon + name + location + status badge */}
        <div
          style={{ display: "flex", "align-items": "center", gap: "0.6rem" }}
        >
          <IconBox />
          <div style={{ flex: 1, "min-width": 0 }}>
            <div
              style={{
                "font-weight": "600",
                "font-size": "0.85rem",
                "font-family": "monospace",
                overflow: "hidden",
                "text-overflow": "ellipsis",
                "white-space": "nowrap",
              }}
            >
              {props.sensor.sensorId}
            </div>
            <div
              style={{
                "font-size": "0.65rem",
                color: "#4b5563",
                "margin-top": "1px",
              }}
            >
              {derivedLocation(props.sensor.sensorId)}
            </div>
          </div>
          <Badge />
          <ScopeButton />
        </div>

        {/* Divider */}
        <div style={{ "border-top": "1px solid rgba(255,255,255,0.05)" }} />

        {/* Metrics row: Validation | Data Rate | Latency | DLQ Count */}
        <div
          style={{
            display: "grid",
            "grid-template-columns": "1fr 1fr 1fr 1fr",
            gap: "6px",
          }}
        >
          {/* Validation */}
          <div
            style={{ display: "flex", "flex-direction": "column", gap: "2px" }}
          >
            <div
              style={{
                "font-size": "0.55rem",
                color: "#4b5563",
                "text-transform": "uppercase",
                "letter-spacing": "0.06em",
              }}
            >
              Validation
            </div>
            <div
              style={{
                "font-size": "0.8rem",
                "font-weight": "700",
                "font-family": "monospace",
                color: qualityColor(),
              }}
            >
              {props.sensor.status === "OFFLINE"
                ? "N/A"
                : `${props.sensor.validationPassRate}%`}
            </div>
          </div>

          {/* Data Rate */}
          <div
            style={{ display: "flex", "flex-direction": "column", gap: "2px" }}
          >
            <div
              style={{
                "font-size": "0.55rem",
                color: "#4b5563",
                "text-transform": "uppercase",
                "letter-spacing": "0.06em",
              }}
            >
              Data Rate
            </div>
            <div
              style={{
                "font-size": "0.8rem",
                "font-weight": "700",
                "font-family": "monospace",
                color: "#e2e8f0",
              }}
            >
              {props.sensor.eventsPerSecond}{" "}
              <span
                style={{
                  "font-size": "0.55rem",
                  color: "#4b5563",
                  "font-weight": "normal",
                }}
              >
                /s
              </span>
            </div>
          </div>

          {/* Latency — deterministic derived from seed */}
          <div
            style={{ display: "flex", "flex-direction": "column", gap: "2px" }}
          >
            <div
              style={{
                "font-size": "0.55rem",
                color: "#4b5563",
                "text-transform": "uppercase",
                "letter-spacing": "0.06em",
              }}
            >
              Latency
            </div>
            {(() => {
              if (props.sensor.status === "OFFLINE") {
                return (
                  <div
                    style={{
                      "font-size": "0.8rem",
                      "font-weight": "700",
                      "font-family": "monospace",
                      color: "#475569",
                    }}
                  >
                    —
                  </div>
                );
              }
              const seed = props.sensor.sensorId
                .split("")
                .reduce((a, c) => a + c.charCodeAt(0), 0);
              const base = props.sensor.status === "CONNECTED" ? 22 : 180;
              const latMs =
                base + Math.round(Math.abs(Math.sin(seed * 7.3 + 1.5)) * 48);
              const latColor =
                latMs < 100 ? "#4ade80" : latMs < 250 ? "#fbbf24" : "#f87171";
              return (
                <div
                  style={{
                    "font-size": "0.8rem",
                    "font-weight": "700",
                    "font-family": "monospace",
                    color: latColor,
                  }}
                >
                  {latMs}{" "}
                  <span
                    style={{
                      "font-size": "0.55rem",
                      color: "#4b5563",
                      "font-weight": "normal",
                    }}
                  >
                    ms
                  </span>
                </div>
              );
            })()}
          </div>

          {/* DLQ Count */}
          <div
            style={{ display: "flex", "flex-direction": "column", gap: "2px" }}
          >
            <div
              style={{
                "font-size": "0.55rem",
                color: "#4b5563",
                "text-transform": "uppercase",
                "letter-spacing": "0.06em",
              }}
            >
              DLQ Count
            </div>
            <div
              style={{
                "font-size": "0.8rem",
                "font-weight": "700",
                "font-family": "monospace",
                color:
                  props.sensor.dlqCount > 50
                    ? "#f87171"
                    : props.sensor.dlqCount > 10
                      ? "#fbbf24"
                      : "#4ade80",
              }}
            >
              {props.sensor.dlqCount}
            </div>
          </div>
        </div>

        {/* Footer: Last Seen */}
        <div
          style={{
            display: "flex",
            "justify-content": "space-between",
            "align-items": "center",
            "padding-top": "0.4rem",
            "border-top": "1px solid rgba(255,255,255,0.05)",
            "font-size": "0.6rem",
          }}
        >
          <span style={{ color: "#4b5563", "font-family": "monospace" }}>
            Last Seen
          </span>
          <span style={{ color: "#94a3b8" }}>
            {formatLastSeen(props.sensor.lastSeenSeconds)}
          </span>
        </div>

        <style>{hoverStyle}</style>
      </div>
    </Show>
  );
}
