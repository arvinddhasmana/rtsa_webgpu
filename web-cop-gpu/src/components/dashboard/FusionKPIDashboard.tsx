// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/FusionKPIDashboard.tsx

import { Component, For, createMemo } from "solid-js";
import { trackCount } from "../../signals/stats";

export const FusionKPIDashboard: Component = () => {
  // Mock buckets for high-fidelity look
  const confidenceBuckets = createMemo(() => [
    { label: "High", value: 74, color: "#4ade80" },
    { label: "Mid", value: 19, color: "#fbbf24" },
    { label: "Low", value: 7, color: "#f87171" },
  ]);

  const sensorContribution = createMemo(() => [
    { label: "RADAR", value: 45, color: "#38bdf8" },
    { label: "SIGINT", value: 30, color: "#818cf8" },
    { label: "SATELLITE", value: 20, color: "#94a3b8" },
    { label: "EW", value: 5, color: "#6366f1" },
  ]);

  return (
    <div
      data-testid="fusion-kpi-dashboard"
      style={{
        display: "flex",
        "flex-direction": "column",
        gap: "1.25rem",
        color: "#e2e8f0",
        "font-family": "Inter, sans-serif",
      }}
    >
      {/* Active Observations HUD */}
      <section style={{
        background: "rgba(15, 23, 42, 0.4)",
        border: "1px solid rgba(255, 255, 255, 0.1)",
        "border-radius": "8px",
        padding: "1rem",
        position: "relative",
        overflow: "hidden"
      }}>
        <div style={{ "font-size": "0.65rem", color: "#94a3b8", "margin-bottom": "0.5rem" }}>ACTIVE OBS</div>
        <div style={{ display: "flex", "align-items": "baseline", gap: "0.5rem" }}>
          <span style={{ "font-size": "2.5rem", "font-weight": "700", color: "#f8fafc" }}>
            {trackCount().toLocaleString()}
          </span>
          <span style={{ "font-size": "0.8rem", color: "#64748b" }}>Tracks</span>
        </div>
        <div style={{ position: "absolute", top: "1rem", right: "1rem" }} title="Total number of active fused tracks currently being monitored by the Operations Commander.">
            <div style={{ width: "12px", height: "12px", border: "1px solid #4ade80", "border-radius": "50%", display: "flex", "align-items": "center", "justify-content": "center", "font-size": "8px", color: "#4ade80", cursor: "help" }}>i</div>
        </div>
      </section>

      {/* Confidence Buckets */}
      <section>
        <h3 style={{ "font-size": "0.65rem", color: "#64748b", "margin-bottom": "0.75rem", "font-weight": "700" }}>CONFIDENCE BUCKETS</h3>
        <div style={{ display: "flex", "flex-direction": "column", gap: "0.6rem" }}>
          <For each={confidenceBuckets()}>
            {(bucket) => (
              <div style={{ display: "flex", "align-items": "center", gap: "1rem" }}>
                <div style={{ width: "2.5rem", "font-size": "0.65rem", color: "#94a3b8" }}>{bucket.label}</div>
                <div style={{ flex: "1", height: "6px", background: "rgba(255,255,255,0.05)", "border-radius": "3px", overflow: "hidden" }}>
                  <div style={{ width: `${bucket.value}%`, height: "100%", background: bucket.color, "border-radius": "3px" }} />
                </div>
                <div style={{ width: "2rem", "font-size": "0.65rem", "text-align": "right", color: "#e2e8f0" }}>{bucket.value}%</div>
              </div>
            )}
          </For>
        </div>
      </section>

      {/* Sensor Contribution */}
      <section>
        <h3 style={{ "font-size": "0.65rem", color: "#64748b", "margin-bottom": "0.75rem", "font-weight": "700" }}>SENSOR CONTRIBUTION</h3>
        <div style={{ display: "flex", "flex-direction": "column", gap: "0.6rem" }}>
          <For each={sensorContribution()}>
            {(item) => (
              <div style={{ display: "flex", "align-items": "center", gap: "1rem" }}>
                <div style={{ width: "4rem", "font-size": "0.65rem", color: "#94a3b8" }}>{item.label}:</div>
                <div style={{ flex: "1", height: "6px", background: "rgba(255,255,255,0.05)", "border-radius": "3px", overflow: "hidden" }}>
                  <div style={{ width: `${item.value}%`, height: "100%", background: item.color, "border-radius": "3px" }} />
                </div>
                <div style={{ width: "2rem", "font-size": "0.65rem", "text-align": "right", color: "#e2e8f0" }}>{item.value}%</div>
              </div>
            )}
          </For>
        </div>
      </section>

      {/* Latency Stats */}
      <section style={{ "margin-top": "0.5rem" }}>
        <h3 style={{ "font-size": "0.65rem", color: "#64748b", "margin-bottom": "0.75rem", "font-weight": "700", display: "flex", "justify-content": "space-between" }}>
          LATENCY
          <span style={{ display: "flex", gap: "4px" }}>
            <div style={{ width: "4px", height: "4px", "border-radius": "50%", background: "#4ade80" }} />
            <div style={{ width: "4px", height: "4px", "border-radius": "50%", background: "#4ade80" }} />
            <div style={{ width: "4px", height: "4px", "border-radius": "50%", background: "#94a3b8" }} />
          </span>
        </h3>
        <div style={{ display: "flex", "flex-direction": "column", gap: "0.4rem" }}>
            <div style={{ display: "flex", "justify-content": "space-between", "font-size": "0.7rem" }}>
                <span style={{ color: "#94a3b8" }}>Avg: 11ms</span>
                <div style={{ width: "40px", height: "2px", background: "#4ade80", "align-self": "center" }} />
            </div>
            <div style={{ display: "flex", "justify-content": "space-between", "font-size": "0.7rem" }}>
                <span style={{ color: "#94a3b8" }}>Max: 32ms</span>
                <div style={{ width: "40px", height: "2px", background: "#f87171", "align-self": "center" }} />
            </div>
        </div>

        {/* Sparkline Placeholder */}
        <div style={{ height: "40px", background: "rgba(56, 189, 248, 0.05)", "margin-top": "1rem", "border-radius": "4px", position: "relative" }}>
             <svg width="100%" height="100%" viewBox="0 0 100 40" preserveAspectRatio="none">
                <path d="M0 35 L10 30 L20 37 L30 25 L40 28 L50 15 L60 22 L70 10 L80 18 L90 5 L100 12 L100 40 L0 40 Z" fill="rgba(56, 189, 248, 0.1)" />
                <path d="M0 35 L10 30 L20 37 L30 25 L40 28 L50 15 L60 22 L70 10 L80 18 L90 5 L100 12" fill="none" stroke="#38bdf8" stroke-width="1" />
             </svg>
        </div>

        <div style={{ display: "flex", "justify-content": "space-between", "margin-top": "0.5rem" }}>
            <span style={{ "font-size": "0.65rem", color: "#64748b" }}>STATUS</span>
            <div style={{ display: "flex", gap: "4px" }}>
                <div style={{ width: "12px", height: "4px", background: "#4ade80", "border-radius": "1px" }} />
                <div style={{ width: "12px", height: "4px", background: "#4ade80", "border-radius": "1px" }} />
                <div style={{ width: "12px", height: "4px", background: "#f59e0b", "border-radius": "1px" }} />
                <div style={{ width: "12px", height: "4px", background: "#4b5563", "border-radius": "1px" }} />
            </div>
        </div>
      </section>
    </div>
  );
};
