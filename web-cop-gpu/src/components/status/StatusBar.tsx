// CLASSIFICATION: UNCLASSIFIED
// src/components/status/StatusBar.tsx
//
// Bottom status bar showing live FPS, track counts, WebTransport latency,
// and connection state.  All values are driven by SolidJS signals that are
// updated from Worker postMessage events — zero main-thread computation.
// Reference: docs/implementation/v4/phase3_ui_interaction.md §3 U3-8

import type { JSX } from "solid-js";
import { fps, trackCount, visibleCount, latencyMs } from "../../signals/stats";
import { wtConnected } from "../../signals/connection";

const labelStyle: JSX.CSSProperties = {
  "font-size": "0.55rem",
  "text-transform": "uppercase",
  "letter-spacing": "0.08em",
  color: "#64748b",
};

const valueStyle: JSX.CSSProperties = {
  "font-size": "0.8rem",
  color: "#e2e8f0",
  "font-variant-numeric": "tabular-nums",
};

const itemStyle: JSX.CSSProperties = {
  display: "flex",
  "flex-direction": "column",
  "align-items": "center",
  padding: "0 0.75rem",
  "border-right": "1px solid #1e2a3a",
};

/** Bottom status bar — driven entirely by signals (no local state). */
export function StatusBar() {
  const fpsColor = () => (fps() >= 55 ? "#22c55e" : fps() >= 30 ? "#f59e0b" : "#ef4444");

  const latencyDisplay = () => {
    const ms = latencyMs();
    return ms >= 0 ? `${ms} ms` : "—";
  };

  return (
    <div
      style={{
        display: "flex",
        "align-items": "stretch",
        height: "2rem",
        "font-family": "monospace",
      }}
      aria-label="Status bar"
    >
      {/* FPS */}
      <div style={itemStyle}>
        <span style={labelStyle}>FPS</span>
        <span style={{ ...valueStyle, color: fpsColor() }}>{fps().toFixed(0)}</span>
      </div>

      {/* Track count */}
      <div style={itemStyle}>
        <span style={labelStyle}>Tracks</span>
        <span style={valueStyle}>{trackCount()}</span>
      </div>

      {/* Visible count */}
      <div style={itemStyle}>
        <span style={labelStyle}>Visible</span>
        <span style={valueStyle}>{visibleCount()}</span>
      </div>

      {/* Latency */}
      <div style={itemStyle}>
        <span style={labelStyle}>Latency</span>
        <span style={valueStyle}>{latencyDisplay()}</span>
      </div>

      {/* Connection */}
      <div style={{ ...itemStyle, "border-right": "none" }}>
        <span style={labelStyle}>WT</span>
        <span
          style={{
            ...valueStyle,
            color: wtConnected() ? "#22c55e" : "#ef4444",
          }}
        >
          {wtConnected() ? "OK" : "—"}
        </span>
      </div>
    </div>
  );
}
