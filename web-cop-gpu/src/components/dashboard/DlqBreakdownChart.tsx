// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/DlqBreakdownChart.tsx — SVG horizontal bar chart for DLQ reasons
//
// Reference: docs/implementation/v5/sensordashboard_three_level_plan.md §B10

import { For, JSX, Show } from "solid-js";

export interface DlqReasonBar {
  reason: string;
  percentage: number;
  count: number;
}

export interface DlqBreakdownChartProps {
  reasons: DlqReasonBar[];
}

function barColor(reason: string): string {
  const r = reason.toLowerCase();
  if (r.includes("schema") || r.includes("mismatch")) return "#f59e0b";
  if (r.includes("crc") || r.includes("error")) return "#f97316";
  if (r.includes("rate") || r.includes("limit")) return "#6366f1";
  return "#64748b";
}

/** Horizontal bar chart for DLQ rejection reasons — B10 */
export function DlqBreakdownChart(props: DlqBreakdownChartProps): JSX.Element {
  return (
    <div data-testid="dlq-breakdown-chart" style={{ display: "flex", "flex-direction": "column", gap: "8px" }}>
      <Show
        when={props.reasons.length > 0}
        fallback={
          <div style={{ color: "#334155", "font-size": "0.68rem", "font-family": "monospace" }}>
            No DLQ rejections
          </div>
        }
      >
        <For each={props.reasons}>
          {(row) => {
            const color = barColor(row.reason);
            const pct = Math.max(0, Math.min(100, row.percentage));
            return (
              <div
                data-testid={`dlq-bar-${row.reason.replace(/\s+/g, "-").toLowerCase()}`}
                style={{ display: "flex", "flex-direction": "column", gap: "3px" }}
              >
                <div style={{ display: "flex", "justify-content": "space-between", "align-items": "center" }}>
                  <span style={{ color: "#94a3b8", "font-size": "0.68rem", "font-family": "monospace" }}>
                    {row.reason}
                  </span>
                  <span style={{ color, "font-size": "0.68rem", "font-family": "monospace", "font-weight": "600" }}>
                    {pct.toFixed(1)}%
                  </span>
                </div>
                <div style={{
                  height: "6px",
                  background: "rgba(255,255,255,0.06)",
                  "border-radius": "3px",
                  overflow: "hidden",
                }}>
                  <div style={{
                    width: `${pct}%`,
                    height: "100%",
                    background: color,
                    "border-radius": "3px",
                    transition: "width 0.3s ease",
                  }} />
                </div>
              </div>
            );
          }}
        </For>
      </Show>
    </div>
  );
}
