// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/SensorOverviewMap.tsx — Global coverage visualization for Level 1
//
// Renders all sensor footprints on a single miniature map with gap hatching,
// coastline overlay, toolbar, sensor labels, and fleet-list integration.
//
// Reference: docs/implementation/v5/sensordashboard_three_level_plan.md §B5

import { createSignal, For, JSX, Show } from "solid-js";
import { SensorStatus } from "../../services/sensor-health";
import { SpatialAlertPayload } from "../../signals/spatial-alerts";
import { statusColor } from "./dashboard-utils";

export interface SensorOverviewMapProps {
  sensors: SensorStatus[];
  spatialAlerts?: SpatialAlertPayload[];
  hoveredSensorId?: string;
  onSensorClick?: (sensor: SensorStatus) => void;
  width?: number;
  height?: number;
}

const DEFAULT_BOUNDS = { minLat: 54, maxLat: 62, minLon: -15, maxLon: -5 };

function polarToCartesian(cx: number, cy: number, r: number, deg: number) {
  const rad = ((deg - 90) * Math.PI) / 180;
  return { x: cx + r * Math.cos(rad), y: cy + r * Math.sin(rad) };
}

function describeArc(cx: number, cy: number, r: number, start: number, end: number): string {
  const s = polarToCartesian(cx, cy, r, end);
  const e = polarToCartesian(cx, cy, r, start);
  const large = Math.abs(end - start) <= 180 ? "0" : "1";
  return `M ${cx} ${cy} L ${s.x} ${s.y} A ${r} ${r} 0 ${large} 0 ${e.x} ${e.y} Z`;
}

function sensorShortLabel(sensor: SensorStatus): string {
  const id = sensor.sensorId;
  if (id.includes("RADAR-NORTH")) return "AEGIS";
  if (id.includes("RADAR-SOUTH")) return "RADAR-S";
  if (id.includes("RADAR-EAST")) return "RADAR-E";
  if (id.includes("RADAR-WEST")) return "RADAR-W";
  if (id.includes("AIS")) return id.replace("AIS-", "AIS-");
  if (id.includes("EW")) return id.replace("EW-", "EW-");
  if (id.includes("ELINT")) return "ELINT";
  if (id.includes("ISR")) return "ISR";
  return id.slice(0, 7);
}

export function SensorOverviewMap(props: SensorOverviewMapProps): JSX.Element {
  const w = props.width ?? 400;
  const h = props.height ?? 300;

  const [zoomOffset, setZoomOffset] = createSignal(0);

  const bounds = () => {
    const z = zoomOffset();
    const pad = z * 0.5;
    return {
      minLat: DEFAULT_BOUNDS.minLat + pad,
      maxLat: DEFAULT_BOUNDS.maxLat - pad,
      minLon: DEFAULT_BOUNDS.minLon + pad * 1.5,
      maxLon: DEFAULT_BOUNDS.maxLon - pad * 1.5,
    };
  };

  const latToY = (lat: number) => {
    const b = bounds();
    return h - ((lat - b.minLat) / (b.maxLat - b.minLat)) * h;
  };
  const lonToX = (lon: number) => {
    const b = bounds();
    return ((lon - b.minLon) / (b.maxLon - b.minLon)) * w;
  };
  const nmToPx = (nm: number) => {
    const b = bounds();
    return (nm / 60) * (h / (b.maxLat - b.minLat));
  };

  const offlineSensors = () =>
    props.sensors.filter((s) => s.status === "OFFLINE" && s.coverage);

  const gapCount = () => offlineSensors().length;

  const toolbarBtn: Record<string, string> = {
    background: "rgba(255,255,255,0.04)",
    border: "1px solid rgba(255,255,255,0.08)",
    color: "#64748b",
    padding: "2px 6px",
    "border-radius": "4px",
    cursor: "pointer",
    "font-size": "0.58rem",
    "font-family": "monospace",
    display: "inline-flex",
    "align-items": "center",
  };

  return (
    <div
      data-testid="sensor-overview-map"
      style={{
        background: "rgba(13, 20, 36, 0.4)",
        border: "1px solid rgba(255, 255, 255, 0.08)",
        "border-radius": "12px",
        overflow: "hidden",
        position: "relative",
        width: `${w}px`,
        "flex-shrink": 0,
        "backdrop-filter": "blur(8px)",
      }}
    >
      {/* ── Title row ── */}
      <div style={{
        display: "flex",
        "align-items": "center",
        "justify-content": "space-between",
        padding: "6px 10px",
        "border-bottom": "1px solid rgba(255,255,255,0.05)",
      }}>
        <div style={{
          "font-size": "0.6rem",
          "text-transform": "uppercase",
          "letter-spacing": "0.1em",
          color: "#475569",
          "font-family": "monospace",
        }}>
          Strategic Coverage Overview
        </div>

        {/* Gap badge */}
        <Show when={gapCount() > 0}>
          <div
            data-testid="gap-count-badge"
            style={{
              background: "rgba(248,113,113,0.12)",
              color: "#f87171",
              border: "1px solid rgba(248,113,113,0.25)",
              padding: "1px 7px",
              "border-radius": "10px",
              "font-size": "0.6rem",
              "font-family": "monospace",
              "font-weight": "700",
            }}
          >
            ⚠ {gapCount()} GAP{gapCount() > 1 ? "S" : ""}
          </div>
        </Show>
      </div>

      {/* ── Layer toolbar ── */}
      <div style={{
        display: "flex",
        gap: "4px",
        padding: "4px 8px",
        "border-bottom": "1px solid rgba(255,255,255,0.04)",
        "align-items": "center",
        background: "rgba(0,0,0,0.1)",
      }}>
        <button
          data-testid="overview-zoom-in"
          style={toolbarBtn}
          onClick={() => setZoomOffset((z) => Math.min(z + 1, 3))}
          title="Zoom In"
        >
          +
        </button>
        <button
          data-testid="overview-zoom-out"
          style={toolbarBtn}
          onClick={() => setZoomOffset((z) => Math.max(z - 1, 0))}
          title="Zoom Out"
        >
          −
        </button>
        <div style={{ width: "1px", height: "12px", background: "rgba(255,255,255,0.08)", margin: "0 2px" }} />
        <button style={toolbarBtn}>⊞ Layers</button>
        <button style={toolbarBtn}>⚠ Alerts</button>
        <button style={toolbarBtn}>◫ Style</button>
      </div>

      {/* ── SVG map ── */}
      <svg
        width={w}
        height={h}
        viewBox={`0 0 ${w} ${h}`}
        style={{ background: "radial-gradient(circle at center, #0f172a 0%, #020617 100%)", display: "block" }}
      >
        <defs>
          {/* Gap hatching */}
          <pattern id="ovmap-hatch" patternUnits="userSpaceOnUse" width="6" height="6" patternTransform="rotate(45)">
            <line x1="0" y1="0" x2="0" y2="6" stroke="#f87171" stroke-width="1.2" stroke-opacity="0.45" />
          </pattern>
        </defs>

        {/* Grid lines */}
        <For each={[55, 56, 57, 58, 59, 60, 61]}>
          {(lat) => <line x1="0" y1={latToY(lat)} x2={w} y2={latToY(lat)} stroke="rgba(255,255,255,0.03)" stroke-width="0.5" />}
        </For>
        <For each={[-14, -13, -12, -11, -10, -9, -8, -7, -6]}>
          {(lon) => <line x1={lonToX(lon)} y1="0" x2={lonToX(lon)} y2={h} stroke="rgba(255,255,255,0.03)" stroke-width="0.5" />}
        </For>

        {/* Simplified UK / North Atlantic coastline */}
        <polyline
          points={[
            [-5.5, 58.5], [-5.0, 58.8], [-4.5, 59.0], [-3.8, 58.6],
            [-3.5, 58.0], [-4.0, 57.5], [-5.0, 57.0], [-5.5, 56.5],
            [-5.8, 55.9], [-6.5, 55.2], [-7.0, 55.0], [-7.5, 55.5],
            [-7.8, 56.0], [-8.0, 57.5], [-8.5, 58.0], [-9.0, 58.5],
            [-10.0, 59.0], [-10.5, 59.5],
          ]
            .map(([lon, lat]) => `${lonToX(lon)},${latToY(lat)}`)
            .join(" ")}
          fill="none"
          stroke="rgba(148,163,184,0.15)"
          stroke-width="1.2"
          stroke-linejoin="round"
        />

        {/* Coverage footprints */}
        <For each={props.sensors.filter((s) => s.coverage)}>
          {(s) => {
            const cx = () => lonToX(s.coverage!.centerLon);
            const cy = () => latToY(s.coverage!.centerLat);
            const r = () => nmToPx(s.coverage!.rangeNm);
            const color = statusColor(s.status);
            const isOffline = s.status === "OFFLINE";
            const isHovered = () => props.hoveredSensorId === s.sensorId;
            const isFullCircle =
              Math.abs(s.coverage!.bearingEnd - s.coverage!.bearingStart) >= 360;

            return (
              <g
                style={{
                  opacity: isOffline ? 0.8 : isHovered() ? 1 : 0.45,
                  cursor: props.onSensorClick ? "pointer" : "default",
                }}
                onClick={() => props.onSensorClick?.(s)}
              >
                {/* Hatch for offline */}
                <Show when={isOffline}>
                  <Show
                    when={isFullCircle}
                    fallback={
                      <path
                        d={describeArc(cx(), cy(), r(), s.coverage!.bearingStart, s.coverage!.bearingEnd)}
                        fill="url(#ovmap-hatch)"
                        stroke="#f87171"
                        stroke-width="0.8"
                      />
                    }
                  >
                    <circle cx={cx()} cy={cy()} r={r()} fill="url(#ovmap-hatch)" stroke="#f87171" stroke-width="0.8" />
                  </Show>
                </Show>

                {/* Normal footprint */}
                <Show when={!isOffline}>
                  <Show
                    when={isFullCircle}
                    fallback={
                      <path
                        d={describeArc(cx(), cy(), r(), s.coverage!.bearingStart, s.coverage!.bearingEnd)}
                        fill={`${color}15`}
                        stroke={color}
                        stroke-width="0.8"
                      />
                    }
                  >
                    <circle cx={cx()} cy={cy()} r={r()} fill={`${color}12`} stroke={color} stroke-width="0.8" />
                  </Show>
                </Show>

                {/* Hover highlight ring */}
                <Show when={isHovered()}>
                  <circle cx={cx()} cy={cy()} r={r() + 3} fill="none" stroke={color} stroke-width="1.5" stroke-dasharray="4,3" stroke-opacity="0.7" />
                </Show>

                {/* Sensor dot */}
                <circle cx={cx()} cy={cy()} r="2" fill={color} />

                {/* Sensor label callout */}
                <text
                  x={cx() + 4}
                  y={cy() - 4}
                  fill={isOffline ? "#f87171" : "#94a3b8"}
                  font-size="7"
                  font-family="monospace"
                  style={{ "pointer-events": "none" }}
                >
                  {sensorShortLabel(s)}
                </text>

                {/* GAP DETECTED annotation */}
                <Show when={isOffline}>
                  <text
                    x={cx()}
                    y={cy() + 4}
                    text-anchor="middle"
                    fill="#f87171"
                    font-size="7"
                    font-family="monospace"
                    font-weight="bold"
                    style={{ opacity: 0.85 }}
                  >
                    GAP DETECTED
                  </text>
                </Show>
              </g>
            );
          }}
        </For>
      </svg>

      {/* ── Legend ── */}
      <div style={{
        display: "flex",
        gap: "8px",
        "font-size": "0.55rem",
        padding: "4px 8px",
        "border-top": "1px solid rgba(255,255,255,0.04)",
      }}>
        <div style={{ display: "flex", "align-items": "center", gap: "3px" }}>
          <div style={{ width: "6px", height: "6px", "border-radius": "50%", background: "#4ade80" }} />
          <span style={{ color: "#94a3b8" }}>Connected</span>
        </div>
        <div style={{ display: "flex", "align-items": "center", gap: "3px" }}>
          <div style={{ width: "6px", height: "6px", "border-radius": "50%", background: "#fbbf24" }} />
          <span style={{ color: "#94a3b8" }}>Stale</span>
        </div>
        <div style={{ display: "flex", "align-items": "center", gap: "3px" }}>
          <div style={{ width: "6px", height: "6px", "border-radius": "50%", background: "#f87171" }} />
          <span style={{ color: "#94a3b8" }}>Gap</span>
        </div>
      </div>
    </div>
  );
}
