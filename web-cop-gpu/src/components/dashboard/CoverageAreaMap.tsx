// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/CoverageAreaMap.tsx — Shared map component for L2/L3
//
// Reference: docs/implementation/v5/sensordashboard_three_level_plan.md §B7

import { createEffect, createSignal, For, JSX, onCleanup, Show } from "solid-js";
import { SensorStatus } from "../../services/sensor-health";
import { SpatialAlertPayload } from "../../signals/spatial-alerts";
import { statusColor } from "./dashboard-utils";

export interface CoverageAreaMapBounds {
  minLat: number;
  maxLat: number;
  minLon: number;
  maxLon: number;
}

export interface CoverageAreaMapProps {
  sensors: SensorStatus[];
  spatialAlerts?: SpatialAlertPayload[];
  focusSensorId?: string;
  bounds?: CoverageAreaMapBounds;
  showLabels?: boolean;
  showGapHatching?: boolean;
  showRangeRings?: boolean;
  showSweepAnimation?: boolean;
  showFleetList?: boolean;
  showAlertBanner?: boolean;
  onSensorClick?: (sensor: SensorStatus) => void;
  onGapAlertClick?: (alertId: string) => void;
  width?: string;
  height?: string;
  className?: string;
}

const DEFAULT_BOUNDS: CoverageAreaMapBounds = {
  minLat: 54,
  maxLat: 62,
  minLon: -15,
  maxLon: -5,
};

function sensorLabel(sensor: SensorStatus): string {
  const id = sensor.sensorId;
  if (id.includes("RADAR")) return id.replace("RADAR-", "RDR-");
  if (id.includes("AIS")) return id.replace("AIS-", "AIS-");
  if (id.includes("EW")) return id.replace("EW-", "EW-");
  if (id.includes("ELINT")) return "ELINT";
  if (id.includes("ISR")) return "ISR";
  return id.slice(0, 8);
}

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

function computeFocusBounds(
  sensor: SensorStatus,
  defaultBounds: CoverageAreaMapBounds,
): CoverageAreaMapBounds {
  if (!sensor.coverage) return defaultBounds;
  const c = sensor.coverage;
  const degPerNm = 1 / 60;
  const pad = c.rangeNm * degPerNm * 0.4;
  return {
    minLat: c.centerLat - c.rangeNm * degPerNm - pad,
    maxLat: c.centerLat + c.rangeNm * degPerNm + pad,
    minLon: c.centerLon - c.rangeNm * degPerNm * 1.5 - pad,
    maxLon: c.centerLon + c.rangeNm * degPerNm * 1.5 + pad,
  };
}

/** Core shared SVG map component used by L2 diagnostics and L3 coverage dashboard — B7 */
export function CoverageAreaMap(props: CoverageAreaMapProps): JSX.Element {
  const SVG_W = 500;
  const SVG_H = 380;

  // Effective bounds: focus sensor if specified, else use prop bounds or default
  const effectiveBounds = (): CoverageAreaMapBounds => {
    if (props.focusSensorId) {
      const focused = props.sensors.find((s) => s.sensorId === props.focusSensorId);
      if (focused) return computeFocusBounds(focused, DEFAULT_BOUNDS);
    }
    return props.bounds ?? DEFAULT_BOUNDS;
  };

  // Zoom state (local)
  const [zoomLevel, setZoomLevel] = createSignal(0); // offset in degrees steps
  const zoomedBounds = (): CoverageAreaMapBounds => {
    const b = effectiveBounds();
    const z = zoomLevel();
    const latPad = ((b.maxLat - b.minLat) * 0.1) * z;
    const lonPad = ((b.maxLon - b.minLon) * 0.1) * z;
    return {
      minLat: b.minLat + latPad,
      maxLat: b.maxLat - latPad,
      minLon: b.minLon + lonPad,
      maxLon: b.maxLon - lonPad,
    };
  };

  const latToY = (lat: number) => {
    const b = zoomedBounds();
    return SVG_H - ((lat - b.minLat) / (b.maxLat - b.minLat)) * SVG_H;
  };
  const lonToX = (lon: number) => {
    const b = zoomedBounds();
    return ((lon - b.minLon) / (b.maxLon - b.minLon)) * SVG_W;
  };
  const nmToPx = (nm: number) => {
    const b = zoomedBounds();
    return (nm / 60) * (SVG_H / (b.maxLat - b.minLat));
  };

  // Sweep animation for RADAR
  const [sweepAngle, setSweepAngle] = createSignal(0);
  let animFrame: number | null = null;

  createEffect(() => {
    if (props.showSweepAnimation) {
      let last = 0;
      const animate = (ts: number) => {
        if (last === 0) last = ts;
        const delta = ts - last;
        last = ts;
        setSweepAngle((a) => (a + delta * 0.06) % 360);
        animFrame = requestAnimationFrame(animate);
      };
      animFrame = requestAnimationFrame(animate);
      // Cancel the frame when showSweepAnimation toggles off or component unmounts
      onCleanup(() => {
        if (animFrame !== null) {
          cancelAnimationFrame(animFrame);
          animFrame = null;
        }
      });
    } else {
      if (animFrame !== null) {
        cancelAnimationFrame(animFrame);
        animFrame = null;
      }
    }
  });

  const offlineSensors = () =>
    props.sensors.filter((s) => s.status === "OFFLINE" && s.coverage);

  const focusedSensor = () =>
    props.focusSensorId
      ? props.sensors.find((s) => s.sensorId === props.focusSensorId)
      : undefined;

  return (
    <div
      data-testid="coverage-area-map"
      style={{
        position: "relative",
        width: props.width ?? "100%",
        height: props.height ?? "100%",
        background: "rgba(13, 20, 36, 0.5)",
        border: "1px solid rgba(255,255,255,0.08)",
        "border-radius": "10px",
        overflow: "hidden",
        display: "flex",
        "flex-direction": "column",
      }}
      class={props.className}
    >
      {/* ── Layer toolbar ── */}
      <div style={{
        display: "flex",
        gap: "4px",
        padding: "6px 10px",
        "border-bottom": "1px solid rgba(255,255,255,0.05)",
        background: "rgba(13,20,36,0.6)",
        "align-items": "center",
      }}>
        <button
          data-testid="map-zoom-in"
          onClick={() => setZoomLevel((z) => Math.min(z + 1, 4))}
          style={toolbarBtnStyle}
          title="Zoom In"
        >
          +
        </button>
        <button
          data-testid="map-zoom-out"
          onClick={() => setZoomLevel((z) => Math.max(z - 1, 0))}
          style={toolbarBtnStyle}
          title="Zoom Out"
        >
          −
        </button>
        <div style={{ width: "1px", height: "14px", background: "rgba(255,255,255,0.1)", margin: "0 2px" }} />
        <button style={toolbarBtnStyle} title="Layers">⊞ Layers</button>
        <button style={toolbarBtnStyle} title="Alerts">⚠ Alerts</button>
        <button style={toolbarBtnStyle} title="Map Style">◫ Style</button>

        {/* Gap badge */}
        <Show when={offlineSensors().length > 0}>
          <div
            data-testid="gap-badge"
            style={{
              "margin-left": "auto",
              background: "rgba(248, 113, 113, 0.15)",
              color: "#f87171",
              border: "1px solid rgba(248,113,113,0.3)",
              padding: "2px 8px",
              "border-radius": "10px",
              "font-size": "0.62rem",
              "font-family": "monospace",
              "font-weight": "600",
            }}
          >
            ⚠ {offlineSensors().length} GAP{offlineSensors().length > 1 ? "S" : ""}
          </div>
        </Show>
      </div>

      {/* ── SVG Map ── */}
      <div style={{ flex: 1, position: "relative", overflow: "hidden" }}>
        <svg
          width="100%"
          height="100%"
          viewBox={`0 0 ${SVG_W} ${SVG_H}`}
          preserveAspectRatio="xMidYMid meet"
          style={{ display: "block" }}
        >
          <defs>
            {/* Hatching pattern for offline/gap areas */}
            <pattern id="gap-hatch" patternUnits="userSpaceOnUse" width="8" height="8" patternTransform="rotate(45)">
              <line x1="0" y1="0" x2="0" y2="8" stroke="#f87171" stroke-width="1.5" stroke-opacity="0.5" />
            </pattern>
          </defs>

          {/* Background gradient */}
          <rect width={SVG_W} height={SVG_H} fill="url(#map-bg)" />
          <defs>
            <radialGradient id="map-bg" cx="50%" cy="50%" r="70%">
              <stop offset="0%" stop-color="#0f172a" />
              <stop offset="100%" stop-color="#020617" />
            </radialGradient>
          </defs>

          {/* Grid lines */}
          <For each={[54, 55, 56, 57, 58, 59, 60, 61, 62]}>
            {(lat) => (
              <line
                x1="0" y1={latToY(lat)} x2={SVG_W} y2={latToY(lat)}
                stroke="rgba(255,255,255,0.04)" stroke-width="0.5"
              />
            )}
          </For>
          <For each={[-15, -14, -13, -12, -11, -10, -9, -8, -7, -6, -5]}>
            {(lon) => (
              <line
                x1={lonToX(lon)} y1="0" x2={lonToX(lon)} y2={SVG_H}
                stroke="rgba(255,255,255,0.04)" stroke-width="0.5"
              />
            )}
          </For>

          {/* Simplified UK / North Atlantic coastline approximation */}
          <polyline
            points={[
              [-5.5, 58.5], [-5.0, 58.8], [-4.5, 59.0], [-3.8, 58.6],
              [-3.5, 58.0], [-4.0, 57.5], [-5.0, 57.0], [-5.5, 56.5],
              [-5.8, 55.9], [-6.5, 55.2], [-7.0, 55.0], [-7.5, 55.5],
              [-7.8, 56.0], [-8.0, 57.5], [-8.5, 58.0], [-9.0, 58.5],
              [-10.0, 59.0], [-10.5, 59.5], [-11.0, 60.0], [-12.0, 60.5],
            ]
              .map(([lon, lat]) => `${lonToX(lon)},${latToY(lat)}`)
              .join(" ")}
            fill="none"
            stroke="rgba(148, 163, 184, 0.18)"
            stroke-width="1.5"
            stroke-linejoin="round"
          />

          {/* ── Coverage footprints ── */}
          <For each={props.sensors.filter((s) => s.coverage)}>
            {(s) => {
              const cx = () => lonToX(s.coverage!.centerLon);
              const cy = () => latToY(s.coverage!.centerLat);
              const r = () => nmToPx(s.coverage!.rangeNm);
              const color = statusColor(s.status);
              const isOffline = s.status === "OFFLINE";
              const isFullCircle =
                Math.abs(s.coverage!.bearingEnd - s.coverage!.bearingStart) >= 360;

              return (
                <g
                  style={{ cursor: props.onSensorClick ? "pointer" : "default" }}
                  onClick={() => props.onSensorClick?.(s)}
                >
                  {/* Gap hatching for offline sensors */}
                  <Show when={isOffline && props.showGapHatching !== false}>
                    <Show
                      when={isFullCircle}
                      fallback={
                        <path
                          d={describeArc(cx(), cy(), r(), s.coverage!.bearingStart, s.coverage!.bearingEnd)}
                          fill="url(#gap-hatch)"
                          stroke="#f87171"
                          stroke-width="1"
                          stroke-opacity="0.5"
                        />
                      }
                    >
                      <circle cx={cx()} cy={cy()} r={r()} fill="url(#gap-hatch)" stroke="#f87171" stroke-width="1" stroke-opacity="0.5" />
                    </Show>
                  </Show>

                  {/* Normal footprint */}
                  <Show when={!isOffline}>
                    <Show
                      when={isFullCircle}
                      fallback={
                        <path
                          d={describeArc(cx(), cy(), r(), s.coverage!.bearingStart, s.coverage!.bearingEnd)}
                          fill={`${color}12`}
                          stroke={color}
                          stroke-width="1"
                          stroke-opacity="0.6"
                        />
                      }
                    >
                      <circle cx={cx()} cy={cy()} r={r()} fill={`${color}12`} stroke={color} stroke-width="1" stroke-opacity="0.6" />
                    </Show>
                  </Show>

                  {/* Sensor dot */}
                  <circle cx={cx()} cy={cy()} r="3" fill={color} />

                  {/* Sensor label */}
                  <Show when={props.showLabels}>
                    <text
                      x={cx() + 5}
                      y={cy() - 5}
                      fill={isOffline ? "#f87171" : "#94a3b8"}
                      font-size="9"
                      font-family="monospace"
                    >
                      {sensorLabel(s)}
                    </text>
                  </Show>

                  {/* GAP DETECTED annotation for offline sensors */}
                  <Show when={isOffline}>
                    <text
                      x={cx()}
                      y={cy()}
                      text-anchor="middle"
                      fill="#f87171"
                      font-size="8"
                      font-family="monospace"
                      font-weight="bold"
                      style={{ opacity: 0.8 }}
                    >
                      GAP DETECTED
                    </text>
                  </Show>
                </g>
              );
            }}
          </For>

          {/* ── Range rings for focused sensor (L2) ── */}
          <Show when={props.showRangeRings && focusedSensor() && focusedSensor()!.coverage}>
            <g style={{ "pointer-events": "none" }}>
              <For each={[0.25, 0.5, 0.75, 1.0]}>
                {(pct) => {
                  const cx = () => lonToX(focusedSensor()!.coverage!.centerLon);
                  const cy = () => latToY(focusedSensor()!.coverage!.centerLat);
                  const ring_r = () => nmToPx(focusedSensor()!.coverage!.rangeNm) * pct;
                  const nm = () => Math.round(focusedSensor()!.coverage!.rangeNm * pct);
                  return (
                    <g>
                      <circle cx={cx()} cy={cy()} r={ring_r()} fill="none" stroke="rgba(96,165,250,0.2)" stroke-width="0.75" stroke-dasharray="3,3" />
                      <text x={cx() + ring_r() + 2} y={cy()} fill="rgba(96,165,250,0.5)" font-size="7" font-family="monospace">{nm()}nm</text>
                    </g>
                  );
                }}
              </For>
            </g>
          </Show>

          {/* ── Radar sweep animation (L2 RADAR) ── */}
          <Show when={props.showSweepAnimation && focusedSensor()?.sensorType === "RADAR" && focusedSensor()!.coverage}>
            {(() => {
              const cx = () => lonToX(focusedSensor()!.coverage!.centerLon);
              const cy = () => latToY(focusedSensor()!.coverage!.centerLat);
              const r = () => nmToPx(focusedSensor()!.coverage!.rangeNm);
              const tip = () => polarToCartesian(cx(), cy(), r(), sweepAngle());
              return (
                <line
                  data-testid="radar-sweep-line"
                  x1={cx()} y1={cy()}
                  x2={tip().x} y2={tip().y}
                  stroke="rgba(96,165,250,0.6)"
                  stroke-width="1.5"
                />
              );
            })()}
          </Show>
        </svg>

        {/* ── Sensor position label (L2) ── */}
        <Show when={focusedSensor()?.coverage}>
          <div style={{
            position: "absolute",
            bottom: "6px",
            left: "10px",
            "font-size": "0.6rem",
            "font-family": "monospace",
            color: "#475569",
          }}>
            {focusedSensor()!.coverage!.centerLat.toFixed(2)}°N {Math.abs(focusedSensor()!.coverage!.centerLon).toFixed(2)}°W
          </div>
        </Show>
      </div>
    </div>
  );
}

const toolbarBtnStyle: Record<string, string> = {
  background: "rgba(255,255,255,0.04)",
  border: "1px solid rgba(255,255,255,0.08)",
  color: "#64748b",
  padding: "2px 7px",
  "border-radius": "4px",
  cursor: "pointer",
  "font-size": "0.62rem",
  "font-family": "monospace",
};
