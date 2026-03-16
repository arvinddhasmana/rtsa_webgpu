// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/FusionKPIDashboard.tsx

import { SensorType } from "@gen/rtsa/common/v1/types_pb.js";
import { Component, For } from "solid-js";
import { activeObservationCount, averageObservationConfidence, observationTypeDistribution } from "../../signals/sensor-observations";

const SENSOR_LABELS: Record<number, string> = {
  [SensorType.RADAR]: "RADAR",
  [SensorType.AIS_BFT]: "AIS",
  [SensorType.EW_SIGINT]: "SIGINT",
  [SensorType.ELINT_COMINT]: "ELINT",
  [SensorType.ISR]: "EO/IR",
  [SensorType.CYBER]: "CYBER",
};

export const FusionKPIDashboard: Component = () => {
  return (
    <div
      data-testid="fusion-kpi-dashboard"
      style={{
        display: "flex",
        "flex-direction": "column",
        gap: "1.5rem",
        color: "#e2e8f0",
      }}
    >
      <header>
        <h2 style={{ "font-size": "0.9rem", "font-weight": "600", color: "#38bdf8", "letter-spacing": "0.05em", "margin-bottom": "0.5rem" }}>
          FUSION KPIs
        </h2>
        <div style={{ height: "1px", background: "linear-gradient(90deg, #38bdf8 0%, transparent 100%)", opacity: "0.3" }} />
      </header>

      {/* Primary Metric: Active Observations */}
      <section
        style={{
          background: "rgba(30, 41, 59, 0.5)",
          padding: "1rem",
          "border-radius": "8px",
          border: "1px solid rgba(56, 189, 248, 0.1)",
        }}
      >
        <div style={{ "font-size": "0.75rem", color: "#94a3b8", "margin-bottom": "0.25rem" }}>Active Observations</div>
        <div style={{ "font-size": "2rem", "font-weight": "700", color: "#38bdf8" }}>
          {activeObservationCount().toLocaleString()}
        </div>
        <div style={{ "font-size": "0.7rem", color: "#4ade80", "margin-top": "0.25rem" }}>
          {(averageObservationConfidence() * 100).toFixed(1)}% Avg Confidence
        </div>
      </section>

      {/* Sensor Diversity */}
      <section>
        <h3 style={{ "font-size": "0.75rem", color: "#94a3b8", "margin-bottom": "0.75rem", "text-transform": "uppercase" }}>
          Sensor Diversity
        </h3>
        <div style={{ display: "flex", "flex-direction": "column", gap: "0.5rem" }}>
          <For each={Object.entries(observationTypeDistribution())}>
            {([type, count]) => (
              <div style={{ display: "flex", "align-items": "center", gap: "0.75rem" }}>
                <div style={{ width: "4rem", "font-size": "0.7rem", color: "#cbd5e1" }}>
                  {SENSOR_LABELS[Number(type)] || "OTHER"}
                </div>
                <div style={{ flex: "1", height: "4px", background: "#1e293b", "border-radius": "2px", overflow: "hidden" }}>
                  <div
                    style={{
                      height: "100%",
                      width: `${(count / activeObservationCount()) * 100}%`,
                      background: "#38bdf8",
                      transition: "width 0.3s ease",
                    }}
                  />
                </div>
                <div style={{ width: "2rem", "text-align": "right", "font-size": "0.7rem", color: "#94a3b8" }}>
                  {count}
                </div>
              </div>
            )}
          </For>
        </div>
      </section>

      {/* Fusion Health Gauge Placeholder */}
      <section style={{ "margin-top": "auto" }}>
        <div
          style={{
            height: "120px",
            display: "flex",
            "align-items": "center",
            "justify-content": "center",
            border: "1px dashed rgba(56, 189, 248, 0.2)",
            "border-radius": "8px",
            position: "relative",
            overflow: "hidden"
          }}
        >
          <div style={{ position: "absolute", inset: "0", background: "radial-gradient(circle at center, rgba(56, 189, 248, 0.05) 0%, transparent 70%)" }} />
          <div style={{ "text-align": "center", "z-index": "1" }}>
            <div style={{ "font-size": "1.5rem", "font-weight": "700", color: "#4ade80" }}>88%</div>
            <div style={{ "font-size": "0.65rem", color: "#94a3b8" }}>FUSION CONFIDENCE</div>
          </div>
        </div>
      </section>
    </div>
  );
};
