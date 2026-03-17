// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/SensorLegend.tsx

import { Component } from "solid-js";

export const SensorLegend: Component = () => {
  return (
    <div
      data-testid="sensor-legend"
      style={{
        background: "rgba(15, 23, 42, 0.6)",
        "backdrop-filter": "blur(12px)",
        border: "1px solid rgba(255, 255, 255, 0.1)",
        "border-radius": "12px",
        padding: "1rem",
        color: "#f8fafc",
        width: "280px",
        display: "flex",
        "flex-direction": "column",
        gap: "1rem",
        "box-shadow": "0 10px 40px rgba(0,0,0,0.4)",
      }}
    >
      <header style={{ "font-size": "0.7rem", "font-weight": "800", color: "#94a3b8", "letter-spacing": "0.1em", "text-transform": "uppercase" }}>
        PROFESSIONAL GLASS LEGEND
      </header>

      <div style={{ display: "grid", "grid-template-columns": "1fr 1fr", gap: "1rem" }}>
        {/* Friendly Section */}
        <div style={{ display: "flex", "flex-direction": "column", gap: "0.5rem" }}>
          <div style={{ "font-size": "0.65rem", color: "#64748b", "font-weight": "700" }}>FRIENDLY</div>
          <div style={{ display: "grid", "grid-template-columns": "20px 1fr", "align-items": "center", gap: "8px" }}>
            <span style={{ color: "#38bdf8", "font-size": "0.9rem" }}>✈</span>
            <span style={{ "font-size": "0.7rem", color: "#e2e8f0" }}>Friendly</span>
          </div>
          <div style={{ display: "grid", "grid-template-columns": "20px 1fr", "align-items": "center", gap: "8px" }}>
            <span style={{ color: "#38bdf8", "font-size": "0.9rem" }}>△</span>
            <span style={{ "font-size": "0.7rem", color: "#e2e8f0" }}>Drone</span>
          </div>
        </div>

        {/* Hostile Section */}
        <div style={{ display: "flex", "flex-direction": "column", gap: "0.5rem" }}>
          <div style={{ "font-size": "0.65rem", color: "#f87171", "font-weight": "700" }}>HOSTILE</div>
          <div style={{ display: "grid", "grid-template-columns": "20px 1fr", "align-items": "center", gap: "8px" }}>
            <span style={{ color: "#f87171", "font-size": "0.9rem" }}>⬥</span>
            <span style={{ "font-size": "0.7rem", color: "#e2e8f0" }}>Hostile Icons</span>
          </div>
          <div style={{ display: "grid", "grid-template-columns": "20px 1fr", "align-items": "center", gap: "8px" }}>
            <span style={{ color: "#f87171", "font-size": "0.9rem" }}>◇</span>
            <span style={{ "font-size": "0.7rem", color: "#e2e8f0" }}>Statuses</span>
          </div>
        </div>
      </div>

      <div style={{ height: "1px", background: "rgba(255,255,255,0.05)" }} />

      {/* Statuses Section */}
      <div style={{ display: "flex", "flex-direction": "column", gap: "0.65rem" }}>
        <div style={{ "font-size": "0.65rem", color: "#64748b", "font-weight": "700" }}>STATUSES</div>
        <div style={{ display: "grid", "grid-template-columns": "40px 1fr", "align-items": "center", gap: "12px" }}>
          <div style={{ height: "2px", width: "30px", background: "#38bdf8" }} />
          <span style={{ "font-size": "0.7rem", color: "#94a3b8" }}>Frigate</span>
        </div>
        <div style={{ display: "grid", "grid-template-columns": "40px 1fr", "align-items": "center", gap: "12px" }}>
            <div style={{ height: "4px", width: "30px", background: "#f87171" }} />
          <span style={{ "font-size": "0.7rem", color: "#94a3b8" }}>Destroyer</span>
        </div>
        <div style={{ display: "grid", "grid-template-columns": "40px 1fr", "align-items": "center", gap: "12px" }}>
          <div style={{ height: "4px", width: "20px", background: "#38bdf8", "border-radius": "10px 10px 0 0" }} />
          <span style={{ "font-size": "0.7rem", color: "#94a3b8" }}>Submarine silhouette</span>
        </div>
      </div>
    </div>
  );
};
