// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/SensorOverviewMap.tsx — Global coverage visualization for Level 1
//
// Renders all sensor footprints on a single miniature map to provide strategic awareness
// of coverage gaps and overlaps without leaving the Health dashboard.

import { For, JSX, Show } from "solid-js";
import { SensorStatus } from "../../services/sensor-health";

interface SensorOverviewMapProps {
  sensors: SensorStatus[];
  width?: number;
  height?: number;
}

export function SensorOverviewMap(props: SensorOverviewMapProps): JSX.Element {
  const w = props.width ?? 400;
  const h = props.height ?? 300;

  // Operational area bound (roughly matching sensor-health-demo.yaml center: 58N 10W)
  const bounds = {
    minLat: 54,
    maxLat: 62,
    minLon: -15,
    maxLon: -5,
  };

  const latToY = (lat: number) => h - ((lat - bounds.minLat) / (bounds.maxLat - bounds.minLat)) * h;
  const lonToX = (lon: number) => ((lon - bounds.minLon) / (bounds.maxLon - bounds.minLon)) * w;

  const nmToPx = (nm: number) => (nm / 60) * (h / (bounds.maxLat - bounds.minLat));

  const describeArc = (cx: number, cy: number, r: number, startAngle: number, endAngle: number) => {
    const start = polarToCartesian(cx, cy, r, endAngle);
    const end = polarToCartesian(cx, cy, r, startAngle);
    const largeArcFlag = Math.abs(endAngle - startAngle) <= 180 ? "0" : "1";

    return [
      "M", cx, cy,
      "L", start.x, start.y,
      "A", r, r, 0, largeArcFlag, 0, end.x, end.y,
      "Z"
    ].join(" ");
  };

  const polarToCartesian = (centerX: number, centerY: number, radius: number, angleInDegrees: number) => {
    const angleInRadians = ((angleInDegrees - 90) * Math.PI) / 180.0;
    return {
      x: centerX + radius * Math.cos(angleInRadians),
      y: centerY + radius * Math.sin(angleInRadians)
    };
  };

  const statusColor = (status: string) => {
    switch (status) {
      case "CONNECTED": return "#4ade80";
      case "STALE": return "#fbbf24";
      default: return "#f87171";
    }
  };

  return (
    <div
      style={{
        background: "rgba(13, 20, 36, 0.4)",
        border: "1px solid rgba(255, 255, 255, 0.08)",
        "border-radius": "12px",
        overflow: "hidden",
        position: "relative",
        width: `${w}px`,
        height: `${h}px`,
        "backdrop-filter": "blur(8px)",
      }}
    >
      <div style={{
          position: "absolute",
          top: "8px",
          left: "8px",
          "font-size": "0.6rem",
          "text-transform": "uppercase",
          "letter-spacing": "0.1em",
          color: "#475569",
          "font-family": "monospace",
          "z-index": 10
      }}>
          Strategic Coverage Overview
      </div>

      <svg width={w} height={h} viewBox={`0 0 ${w} ${h}`} style={{ background: "radial-gradient(circle at center, #0f172a 0%, #020617 100%)" }}>
        {/* Simple grid lines */}
        <For each={[55, 56, 57, 58, 59, 60, 61]}>
          {(lat) => <line x1="0" y1={latToY(lat)} x2={w} y2={latToY(lat)} stroke="rgba(255,255,255,0.03)" stroke-width="0.5" />}
        </For>
        <For each={[-14, -13, -12, -11, -10, -9, -8, -7, -6]}>
          {(lon) => <line x1={lonToX(lon)} y1="0" x2={lonToX(lon)} y2={h} stroke="rgba(255,255,255,0.03)" stroke-width="0.5" />}
        </For>

        {/* Coverage footprints */}
        <For each={props.sensors.filter(s => s.coverage)}>
          {(s) => {
            const cx = lonToX(s.coverage!.centerLon);
            const cy = latToY(s.coverage!.centerLat);
            const r = nmToPx(s.coverage!.rangeNm);
            const color = statusColor(s.status);
            const isFullCircle = Math.abs(s.coverage!.bearingEnd - s.coverage!.bearingStart) >= 360;

            return (
              <g style={{ opacity: s.status === "OFFLINE" ? 0.15 : 0.4 }}>
                <Show
                  when={isFullCircle}
                  fallback={
                    <path
                      d={describeArc(cx, cy, r, s.coverage!.bearingStart, s.coverage!.bearingEnd)}
                      fill={`${color}15`}
                      stroke={color}
                      stroke-width="1"
                    />
                  }
                >
                  <circle cx={cx} cy={cy} r={r} fill={`${color}15`} stroke={color} stroke-width="1" />
                </Show>
                <circle cx={cx} cy={cy} r="1.5" fill={color} />
              </g>
            );
          }}
        </For>
      </svg>

      {/* Legend */}
      <div style={{
          position: "absolute",
          bottom: "8px",
          right: "8px",
          display: "flex",
          gap: "8px",
          "font-size": "0.55rem"
      }}>
          <div style={{ display: "flex", "align-items": "center", gap: "3px" }}>
            <div style={{ width: "6px", height: "6px", "border-radius": "50%", background: "#4ade80" }} />
            <span style={{ color: "#94a3b8" }}>Covered</span>
          </div>
          <div style={{ display: "flex", "align-items": "center", gap: "3px" }}>
            <div style={{ width: "6px", height: "6px", "border-radius": "50%", background: "#f87171" }} />
            <span style={{ color: "#94a3b8" }}>Gap risk</span>
          </div>
      </div>
    </div>
  );
}
