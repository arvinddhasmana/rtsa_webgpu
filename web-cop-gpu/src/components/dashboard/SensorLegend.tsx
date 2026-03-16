// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/SensorLegend.tsx

import { SensorType } from "@gen/rtsa/common/v1/types_pb.js";
import { Component, For } from "solid-js";

const LEGEND_ITEMS = [
  { type: SensorType.RADAR, label: "RADAR", icon: "📡", color: "#38bdf8" },
  { type: SensorType.AIS_BFT, label: "AIS", icon: "🚢", color: "#fbbf24" },
  { type: SensorType.EW_SIGINT, label: "SIGINT", icon: "📻", color: "#f472b6" },
  { type: SensorType.ELINT_COMINT, label: "ELINT", icon: "📡", color: "#818cf8" },
  { type: SensorType.ISR, label: "EO/IR", icon: "📷", color: "#4ade80" },
];

export const SensorLegend: Component = () => {
  return (
    <div
      data-testid="sensor-legend"
      style={{
        background: "rgba(15, 23, 42, 0.85)",
        border: "1px solid rgba(56, 189, 248, 0.2)",
        "border-radius": "8px",
        padding: "0.75rem",
        color: "#f8fafc",
        display: "flex",
        "flex-direction": "column",
        gap: "0.5rem",
        "backdrop-filter": "blur(4px)",
        width: "140px",
      }}
    >
      <div style={{ "font-size": "0.65rem", "font-weight": "700", color: "#94a3b8", "letter-spacing": "0.05em", "margin-bottom": "0.25rem" }}>
        SENSOR LEGEND
      </div>
      <For each={LEGEND_ITEMS}>
        {(item) => (
          <div style={{ display: "flex", "align-items": "center", gap: "0.5rem" }}>
            <span style={{ "font-size": "0.9rem", color: item.color }}>{item.icon}</span>
            <span style={{ "font-size": "0.7rem", color: "#e2e8f0" }}>{item.label}</span>
          </div>
        )}
      </For>
      <div style={{ height: "1px", background: "rgba(255,255,255,0.05)", "margin-top": "0.25rem" }} />
      <div style={{ display: "flex", "align-items": "center", gap: "0.5rem" }}>
        <div style={{ width: "12px", height: "12px", border: "1px solid #38bdf8", "border-radius": "50%" }} />
        <span style={{ "font-size": "0.65rem", color: "#94a3b8" }}>[OBS]</span>
      </div>
      <div style={{ display: "flex", "align-items": "center", gap: "0.5rem" }}>
        <div style={{ width: "12px", height: "12px", border: "1px solid #fbbf24" }} />
        <span style={{ "font-size": "0.65rem", color: "#94a3b8" }}>[TRACK]</span>
      </div>
    </div>
  );
};
