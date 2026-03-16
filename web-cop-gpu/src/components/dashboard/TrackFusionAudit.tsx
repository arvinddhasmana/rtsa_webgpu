// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/TrackFusionAudit.tsx

import { SensorType } from "@gen/rtsa/common/v1/types_pb.js";
import { Component, For, Show } from "solid-js";
import { contributingObservations, selectedTrackId, selectedTrackQualityIndex } from "../../signals/track-selection";

const SENSOR_ICONS: Record<number, string> = {
  [SensorType.RADAR]: "📡",
  [SensorType.AIS_BFT]: "🚢",
  [SensorType.EW_SIGINT]: "📻",
  [SensorType.ELINT_COMINT]: "📡",
  [SensorType.ISR]: "📷",
};

export const TrackFusionAudit: Component = () => {
  return (
    <Show when={selectedTrackId()}>
      <div
        data-testid="track-fusion-audit"
        style={{
          display: "flex",
          "flex-direction": "column",
          gap: "1.25rem",
          color: "#e2e8f0",
          background: "rgba(15, 23, 42, 0.8)",
          padding: "1rem",
          "border-radius": "12px",
          border: "1px solid rgba(56, 189, 248, 0.2)",
          "backdrop-filter": "blur(8px)",
        }}
      >
        <header style={{ display: "flex", "justify-content": "space-between", "align-items": "center" }}>
          <div>
            <div style={{ "font-size": "0.65rem", color: "#94a3b8", "text-transform": "uppercase" }}>Selected Track</div>
            <div style={{ "font-size": "1.1rem", "font-weight": "700", color: "#38bdf8" }}>{selectedTrackId()}</div>
          </div>
          <div style={{ "text-align": "right" }}>
            <div style={{ "font-size": "0.65rem", color: "#94a3b8", "text-transform": "uppercase" }}>Quality Index</div>
            <div style={{ "font-size": "1.1rem", "font-weight": "700", color: selectedTrackQualityIndex() > 0.7 ? "#4ade80" : "#fbbf24" }}>
              {Math.round(selectedTrackQualityIndex() * 100)}%
            </div>
          </div>
        </header>

        {/* Quality Trend Sparkline Placeholder */}
        <section style={{ height: "40px", width: "100%", background: "rgba(30, 41, 59, 0.5)", "border-radius": "4px", position: "relative", overflow: "hidden" }}>
           <svg width="100%" height="100%" viewBox="0 0 100 40" preserveAspectRatio="none">
              <path d="M0,30 Q10,10 20,25 T40,15 T60,35 T80,5 T100,20" fill="none" stroke="#38bdf8" stroke-width="2" opacity="0.6" />
           </svg>
        </section>

        {/* Genealogy List */}
        <section>
          <h3 style={{ "font-size": "0.75rem", color: "#94a3b8", "margin-bottom": "0.75rem", "text-transform": "uppercase" }}>
            Fusion Genealogy
          </h3>
          <div style={{ display: "flex", "flex-direction": "column", gap: "0.5rem" }}>
            <For each={contributingObservations()}>
              {(obs) => (
                <div
                  style={{
                    display: "flex",
                    "align-items": "center",
                    padding: "0.5rem",
                    background: "rgba(30, 41, 59, 0.4)",
                    "border-radius": "6px",
                    gap: "0.75rem",
                    border: "1px solid rgba(255, 255, 255, 0.03)"
                  }}
                >
                  <div style={{ "font-size": "1rem" }}>{SENSOR_ICONS[obs.type] || "❓"}</div>
                  <div style={{ flex: "1" }}>
                    <div style={{ "font-size": "0.7rem", "font-weight": "600" }}>OBS-{obs.id.slice(-4)}</div>
                    <div style={{ "font-size": "0.6rem", color: "#64748b" }}>{new Date(obs.timestampMs).toLocaleTimeString()}</div>
                  </div>
                  <div style={{ "text-align": "right" }}>
                    <div style={{ "font-size": "0.7rem", color: "#38bdf8" }}>{Math.round(obs.confidence * 100)}%</div>
                    <div style={{ "font-size": "0.55rem", color: "#64748b" }}>Conf.</div>
                  </div>
                </div>
              )}
            </For>
            <Show when={contributingObservations().length === 0}>
              <div style={{ "font-size": "0.7rem", color: "#64748b", "text-align": "center", padding: "1rem" }}>
                No active observations found
              </div>
            </Show>
          </div>
        </section>
      </div>
    </Show>
  );
};
