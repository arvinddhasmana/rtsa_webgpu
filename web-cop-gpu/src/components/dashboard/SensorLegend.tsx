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
          <div style={{ "font-size": "0.65rem", color: "#38bdf8", "font-weight": "700" }}>FRIENDLY</div>
          <div style={{ display: "grid", "grid-template-columns": "24px 1fr", "align-items": "center", gap: "8px" }}>
            <AirIcon color="#38bdf8" />
            <span style={{ "font-size": "0.7rem", color: "#e2e8f0" }}>Air Track</span>
          </div>
          <div style={{ display: "grid", "grid-template-columns": "24px 1fr", "align-items": "center", gap: "8px" }}>
            <SurfaceIcon color="#38bdf8" />
            <span style={{ "font-size": "0.7rem", color: "#e2e8f0" }}>Surface</span>
          </div>
        </div>

        {/* Hostile Section */}
        <div style={{ display: "flex", "flex-direction": "column", gap: "0.5rem" }}>
          <div style={{ "font-size": "0.65rem", color: "#f87171", "font-weight": "700" }}>HOSTILE</div>
          <div style={{ display: "grid", "grid-template-columns": "24px 1fr", "align-items": "center", gap: "8px" }}>
            <SurfaceIcon color="#f87171" />
            <span style={{ "font-size": "0.7rem", color: "#e2e8f0" }}>Hostile Vessel</span>
          </div>
          <div style={{ display: "grid", "grid-template-columns": "24px 1fr", "align-items": "center", gap: "8px" }}>
            <SubsurfaceIcon color="#f87171" />
            <span style={{ "font-size": "0.7rem", color: "#e2e8f0" }}>Subsurface</span>
          </div>
        </div>
      </div>

      <div style={{ height: "1px", background: "rgba(255,255,255,0.05)" }} />

      {/* Neutral/Unknown Section */}
      <div style={{ display: "flex", "flex-direction": "column", gap: "0.5rem" }}>
        <div style={{ "font-size": "0.65rem", color: "#fbbf24", "font-weight": "700" }}>NEUTRAL / UNKNOWN</div>
        <div style={{ display: "grid", "grid-template-columns": "24px 1fr", "align-items": "center", gap: "8px" }}>
          <AirIcon color="#fbbf24" />
          <span style={{ "font-size": "0.7rem", color: "#94a3b8" }}>Unidentified Air</span>
        </div>
      </div>
    </div>
  );
};

const AirIcon = (props: { color: string }) => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke={props.color} stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
    <path d="M12 2L4 22L12 18L20 22L12 2Z" fill={props.color} fill-opacity="0.3" />
  </svg>
);

const SurfaceIcon = (props: { color: string }) => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke={props.color} stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
    <path d="M12 2L22 12L12 22L2 12L12 2Z" fill={props.color} fill-opacity="0.3" />
  </svg>
);

const SubsurfaceIcon = (props: { color: string }) => (
  <svg width="20" height="12" viewBox="0 0 40 24" fill="none" stroke={props.color} stroke-width="4" stroke-linecap="round" stroke-linejoin="round">
    <ellipse cx="20" cy="12" rx="18" ry="10" fill={props.color} fill-opacity="0.3" />
  </svg>
);
