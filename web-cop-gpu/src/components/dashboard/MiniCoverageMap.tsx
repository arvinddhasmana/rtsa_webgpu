// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/MiniCoverageMap.tsx — miniature coverage overview
//
// Shows a small SVG-based representation of the sensor footprint and any gaps.
// Used in Level 1 (Sensor Status Cards) and Level 2 (Diagnostic Mini-Map).

import { JSX, Show } from "solid-js";

interface MiniCoverageMapProps {
  rangeNm: number;
  bearingStart: number;
  bearingEnd: number;
  alertLevel: 0 | 1 | 2; // 0=None, 1=Gap Warning, 2=Critical Gap
  width?: number;
  height?: number;
}

export function MiniCoverageMap(props: MiniCoverageMapProps): JSX.Element {
  const w = props.width ?? 60;
  const h = props.height ?? 60;
  const cx = w / 2;
  const cy = h / 2;
  const maxRadius = Math.min(cx, cy) - 5;

  // Convert rangeNm to a relative scale (cap at 150Nm for visualization)
  const radius = (Math.min(props.rangeNm, 150) / 150) * maxRadius;

  const statusColor = () => {
    if (props.alertLevel === 2) return "#f87171"; // Red
    if (props.alertLevel === 1) return "#fbbf24"; // Amber
    return "#3b82f6"; // Blue (Normal)
  };

  const describeArc = (x: number, y: number, r: number, startAngle: number, endAngle: number) => {
    const start = polarToCartesian(x, y, r, endAngle);
    const end = polarToCartesian(x, y, r, startAngle);

    const largeArcFlag = endAngle - startAngle <= 180 ? "0" : "1";

    return [
      "M", x, y,
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

  return (
    <div style={{ position: "relative", width: `${w}px`, height: `${h}px` }}>
      <svg width={w} height={h} viewBox={`0 0 ${w} ${h}`}>
        {/* Footprint Sector / Circle */}
        <Show
          when={Math.abs(props.bearingEnd - props.bearingStart) >= 360}
          fallback={
            <path
              d={describeArc(cx, cy, radius, props.bearingStart, props.bearingEnd)}
              fill={`${statusColor()}20`}
              stroke={statusColor()}
              stroke-width="1.5"
              style={{
                transition: "all 0.4s ease",
                filter: `drop-shadow(0 0 2px ${statusColor()})`,
              }}
            />
          }
        >
          <circle
            cx={cx}
            cy={cy}
            r={radius}
            fill={`${statusColor()}20`}
            stroke={statusColor()}
            stroke-width="1.5"
            style={{
              transition: "all 0.4s ease",
              filter: `drop-shadow(0 0 2px ${statusColor()})`,
            }}
          />
        </Show>

        {/* Center Point */}
        <circle cx={cx} cy={cy} r="2" fill="#fff" />

        {/* Pulsing Gap Indicator (if alert) */}
        <Show when={props.alertLevel > 0}>
          <circle cx={cx + radius * 0.7} cy={cy - radius * 0.3} r="4" fill="#f87171">
            <animate
              attributeName="opacity"
              values="1;0.2;1"
              dur="1.5s"
              repeatCount="indefinite"
            />
            <animate
              attributeName="r"
              values="3;6;3"
              dur="1.5s"
              repeatCount="indefinite"
            />
          </circle>
        </Show>
      </svg>
    </div>
  );
}
