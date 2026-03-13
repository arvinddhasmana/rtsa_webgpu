// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/ObsPerSecChart.tsx — SVG area chart for observations per second
//
// Reference: docs/implementation/v5/sensordashboard_three_level_plan.md §B9

import { JSX, Show } from "solid-js";

export interface ObsPerSecChartProps {
  history: number[];
  minY?: number;
  maxY?: number;
  width?: number;
  height?: number;
}

/** Pure SVG area/line chart for OPS history — B9 */
export function ObsPerSecChart(props: ObsPerSecChartProps): JSX.Element {
  const svgW = props.width ?? 300;
  const svgH = props.height ?? 80;
  const padL = 4;
  const padR = 4;
  const padT = 6;
  const padB = 4;
  const chartW = svgW - padL - padR;
  const chartH = svgH - padT - padB;

  const history = () => props.history;

  const computedMax = () => {
    const h = history();
    if (h.length === 0) return props.maxY ?? 100;
    return props.maxY ?? Math.max(...h, 1);
  };

  const computedMin = () => props.minY ?? 0;

  const avg = () => {
    const h = history();
    if (h.length === 0) return 0;
    return Math.round(h.reduce((a, b) => a + b, 0) / h.length);
  };

  const peak = () => {
    const h = history();
    if (h.length === 0) return 0;
    return Math.max(...h);
  };

  const points = (): string => {
    const h = history();
    if (h.length < 2) return "";
    const n = h.length;
    const yMin = computedMin();
    const yMax = computedMax();
    const range = yMax - yMin || 1;
    return h
      .map((v, i) => {
        const x = padL + (i / (n - 1)) * chartW;
        const y = padT + chartH - ((v - yMin) / range) * chartH;
        return `${x},${y}`;
      })
      .join(" ");
  };

  const areaPoints = (): string => {
    const h = history();
    if (h.length < 2) return "";
    const n = h.length;
    const yMin = computedMin();
    const yMax = computedMax();
    const range = yMax - yMin || 1;
    const bottom = padT + chartH;
    const firstX = padL;
    const lastX = padL + chartW;
    const linePoints = h
      .map((v, i) => {
        const x = padL + (i / (n - 1)) * chartW;
        const y = padT + chartH - ((v - yMin) / range) * chartH;
        return `${x},${y}`;
      })
      .join(" ");
    return `${firstX},${bottom} ${linePoints} ${lastX},${bottom}`;
  };

  return (
    <div data-testid="obs-per-sec-chart" style={{ display: "flex", "flex-direction": "column", gap: "4px" }}>
      {/* Header with avg/peak */}
      <div style={{
        display: "flex",
        gap: "16px",
        "font-size": "0.65rem",
        "font-family": "monospace",
        color: "#64748b",
      }}>
        <span>
          Avg:{" "}
          <span style={{ color: "#60a5fa", "font-weight": "600" }}>
            {avg()} OPS
          </span>
        </span>
        <span>
          Peak:{" "}
          <span style={{ color: "#4ade80", "font-weight": "600" }}>
            {peak()} OPS
          </span>
        </span>
      </div>

      <Show
        when={history().length >= 2}
        fallback={
          <div style={{ color: "#334155", "font-size": "0.65rem", "font-family": "monospace", height: `${svgH}px`, display: "flex", "align-items": "center", "justify-content": "center" }}>
            No data
          </div>
        }
      >
        <svg
          width="100%"
          height={svgH}
          viewBox={`0 0 ${svgW} ${svgH}`}
          preserveAspectRatio="none"
          style={{ display: "block" }}
        >
          <defs>
            <linearGradient id="ops-gradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="#3b82f6" stop-opacity="0.35" />
              <stop offset="100%" stop-color="#3b82f6" stop-opacity="0.03" />
            </linearGradient>
          </defs>
          {/* Area fill */}
          <polygon
            points={areaPoints()}
            fill="url(#ops-gradient)"
          />
          {/* Line */}
          <polyline
            data-testid="ops-polyline"
            points={points()}
            fill="none"
            stroke="#3b82f6"
            stroke-width="1.5"
            stroke-linejoin="round"
            stroke-linecap="round"
          />
        </svg>
      </Show>
    </div>
  );
}
