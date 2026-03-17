// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/FusionCommanderDashboard.tsx

import { createSignal, JSX } from "solid-js";
import { clearSelectedTrack, trackDetail } from "../../signals/track";
import { mapStyle, setMapStyle } from "../../signals/viewport";
import { FusionConflictPanel } from "./FusionConflictPanel";
import { FusionKPIDashboard } from "./FusionKPIDashboard";
import { SensorLegend } from "./SensorLegend";
import { TrackDrillDownOverlay } from "./TrackDrillDownOverlay";
import { TrackFusionAudit } from "./TrackFusionAudit";

interface FusionCommanderDashboardProps {
  mapContent: JSX.Element;
  sidePanelContent?: JSX.Element;
}

/**
 * Operations Commander Fusion Dashboard (Redesigned from Scratch)
 * Focus: Tactical Clarity, Premium Glassmorphism, Zero Clutter.
 */
export function FusionCommanderDashboard(props: FusionCommanderDashboardProps) {
  const [showObs, setShowObs] = createSignal(false);
  const [showTracks, setShowTracks] = createSignal(true);

  return (
    <div
      data-testid="fusion-dashboard-root"
      style={{
        display: "flex",
        width: "100%",
        height: "100%",
        background: "#020617", // Deep Navy/Black
        color: "#f8fafc",
        overflow: "hidden",
        "font-family": "Inter, sans-serif",
      }}
    >
      {/* Tactical Map Region */}
      <section
        style={{
          flex: "1",
          position: "relative",
          "min-width": "0",
          overflow: "hidden",
        }}
      >
        {props.mapContent}

        {/* Floating Top Controls (Glassmorphic) */}
        <div
          style={{
            position: "absolute",
            top: "1.5rem",
            left: "1.5rem",
            right: "1.5rem",
            display: "flex",
            "justify-content": "space-between",
            "pointer-events": "none",
            "z-index": "50",
          }}
        >
          <div style={{ display: "flex", gap: "1rem", "pointer-events": "auto" }}>
            <div
              style={{
                background: "rgba(15, 23, 42, 0.6)",
                "backdrop-filter": "blur(8px)",
                padding: "0.5rem 1rem",
                "border-radius": "8px",
                border: "1px solid rgba(255, 255, 255, 0.1)",
                display: "flex",
                "align-items": "center",
                gap: "1.5rem",
              }}
            >
              <ControlToggle label="RAW OBS" active={showObs()} onClick={() => setShowObs(!showObs())} color="#38bdf8" />
              <ControlToggle label="FUSED TRACKS" active={showTracks()} onClick={() => setShowTracks(!showTracks())} color="#fbbf24" />
            </div>

            <div
              style={{
                background: "rgba(15, 23, 42, 0.6)",
                "backdrop-filter": "blur(8px)",
                padding: "0.5rem",
                "border-radius": "8px",
                border: "1px solid rgba(255, 255, 255, 0.1)",
              }}
            >
              <button
                onClick={() => setMapStyle(mapStyle() === 0 ? 1 : 0)}
                style={{
                  background: mapStyle() === 1 ? "rgba(16, 185, 129, 0.2)" : "transparent",
                  color: mapStyle() === 1 ? "#10b981" : "#94a3b8",
                  border: `1px solid ${mapStyle() === 1 ? "#10b981" : "rgba(255,255,255,0.1)"}`,
                  padding: "4px 12px",
                  "border-radius": "4px",
                  "font-size": "0.7rem",
                  "font-weight": "700",
                  cursor: "pointer",
                  transition: "all 0.2s",
                }}
              >
                {mapStyle() === 1 ? "HD MAP ACTIVE" : "STANDARD MAP"}
              </button>
            </div>
          </div>
        </div>

        {/* Legend Panel (Bottom Left) */}
        <div
          style={{
            position: "absolute",
            bottom: "1.5rem",
            left: "1.5rem",
            "z-index": "20",
            background: "rgba(15, 23, 42, 0.4)",
            "backdrop-filter": "blur(4px)",
            padding: "0.5rem",
            "border-radius": "8px",
            border: "1px solid rgba(255,255,255,0.05)",
          }}
        >
          <SensorLegend />
        </div>

        {/* Track Drill-down (Anchored Overlay) */}
        <TrackDrillDownOverlay
          track={trackDetail()}
          onClose={() => clearSelectedTrack()}
        />
      </section>

      {/* Right Side Panel: Intelligence & Analytics */}
      <aside
        style={{
          width: "24rem",
          background: "linear-gradient(180deg, #0f172a 0%, #020617 100%)",
          "border-left": "1px solid rgba(255, 255, 255, 0.05)",
          display: "flex",
          "flex-direction": "column",
          padding: "1.5rem",
          gap: "2rem",
          overflow: "hidden auto",
          "box-shadow": "-10px 0 30px rgba(0,0,0,0.5)",
        }}
      >
        <FusionKPIDashboard />

        <div style={{ height: "1px", background: "linear-gradient(90deg, transparent, rgba(255,255,255,0.1), transparent)" }} />

        <FusionConflictPanel />

        <div style={{ height: "1px", background: "linear-gradient(90deg, transparent, rgba(255,255,255,0.1), transparent)" }} />

        <TrackFusionAudit />
      </aside>
    </div>
  );
}

const ControlToggle = (props: { label: string; active: boolean; onClick: () => void; color: string }) => (
  <button
    onClick={props.onClick}
    style={{
      background: "transparent",
      border: "none",
      color: props.active ? props.color : "#64748b",
      "font-size": "0.75rem",
      "font-weight": "700",
      "letter-spacing": "0.05em",
      cursor: "pointer",
      display: "flex",
      "align-items": "center",
      gap: "8px",
      transition: "color 0.2s",
    }}
  >
    <div style={{
      width: "8px",
      height: "8px",
      "border-radius": "50%",
      background: props.active ? props.color : "transparent",
      border: `1px solid ${props.active ? props.color : "#475569"}`,
      "box-shadow": props.active ? `0 0 8px ${props.color}` : "none",
    }} />
    {props.label}
  </button>
);
