// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/FusionCommanderDashboard.tsx

import type { JSX } from "solid-js";

interface FusionCommanderDashboardProps {
  mapContent: JSX.Element;
  sidePanelContent?: JSX.Element;
}

import { createSignal } from "solid-js";
import { FusionConflictPanel } from "./FusionConflictPanel";
import { FusionKPIDashboard } from "./FusionKPIDashboard";
import { SensorLegend } from "./SensorLegend";
import { TrackFusionAudit } from "./TrackFusionAudit";

interface FusionCommanderDashboardProps {
  mapContent: JSX.Element;
  sidePanelContent?: JSX.Element;
}

/**
 * Operations Commander Fusion dashboard.
 * Primary view for mission overview, track audit, and conflict resolution.
 */
export function FusionCommanderDashboard(props: FusionCommanderDashboardProps) {
  const [showObs, setShowObs] = createSignal(true);
  const [showTracks, setShowTracks] = createSignal(true);

  return (
    <div
      data-testid="commander-fusion-dashboard"
      style={{
        display: "flex",
        width: "100%",
        height: "100%",
        background: "#070d19",
        overflow: "hidden",
      }}
    >
      <section
        data-testid="commander-fusion-map-container"
        aria-label="Fusion map container"
        style={{
          flex: "1",
          position: "relative",
          "min-width": "0",
          overflow: "hidden",
        }}
      >
        {props.mapContent}

        {/* Observation Layer Toggle */}
        <div
          data-testid="commander-observation-layer-mount"
          aria-label="Observation layer mount"
          style={{
            position: "absolute",
            inset: "0",
            border: showObs() ? "1px dashed rgba(56,189,248,0.35)" : "none",
            "pointer-events": "none",
            opacity: showObs() ? "1" : "0",
            transition: "opacity 0.2s ease",
          }}
        />

        {/* Fused Layer Toggle */}
        <div
          data-testid="commander-fused-layer-mount"
          aria-label="Fused layer mount"
          style={{
            position: "absolute",
            inset: "12px",
            border: showTracks() ? "1px dashed rgba(251,191,36,0.35)" : "none",
            "pointer-events": "none",
            opacity: showTracks() ? "1" : "0",
            transition: "opacity 0.2s ease",
          }}
        />

        {/* Map Legend */}
        <div style={{ position: "absolute", bottom: "1.5rem", left: "1.5rem", "z-index": "10" }}>
          <SensorLegend />
        </div>

        {/* Layer Controls */}
        <div
          style={{
            position: "absolute",
            top: "1.5rem",
            right: "1.5rem",
            display: "flex",
            gap: "0.5rem",
            background: "rgba(15, 23, 42, 0.7)",
            padding: "0.5rem",
            "border-radius": "8px",
            border: "1px solid rgba(255, 255, 255, 0.1)",
            "backdrop-filter": "blur(4px)"
          }}
        >
          <button
            style={{
              background: showObs() ? "#38bdf8" : "transparent",
              color: showObs() ? "#0f172a" : "#94a3b8",
              border: "1px solid rgba(56, 189, 248, 0.5)",
              "font-size": "0.65rem",
              "font-weight": "700",
              padding: "0.25rem 0.5rem",
              "border-radius": "4px",
              cursor: "pointer"
            }}
            onClick={() => setShowObs(!showObs())}
          >
            RAW OBS
          </button>
          <button
            style={{
              background: showTracks() ? "#fbbf24" : "transparent",
              color: showTracks() ? "#0f172a" : "#94a3b8",
              border: "1px solid rgba(251, 191, 36, 0.5)",
              "font-size": "0.65rem",
              "font-weight": "700",
              padding: "0.25rem 0.5rem",
              "border-radius": "4px",
              cursor: "pointer"
            }}
            onClick={() => setShowTracks(!showTracks())}
          >
            FUSED TRACKS
          </button>
        </div>
      </section>

      <aside
        data-testid="commander-fusion-side-panel"
        aria-label="Fusion side panel container"
        style={{
          width: "22rem",
          "flex-shrink": "0",
          background: "rgba(7, 13, 25, 0.95)",
          "border-left": "1px solid #1e2a3a",
          padding: "1rem",
          overflow: "hidden auto",
          display: "flex",
          "flex-direction": "column",
          gap: "1.5rem",
        }}
      >
        {props.sidePanelContent ?? (
          <>
            <FusionKPIDashboard />
            <div style={{ height: "1px", background: "rgba(255,255,255,0.05)" }} />
            <FusionConflictPanel />
            <TrackFusionAudit />
          </>
        )}
      </aside>
    </div>
  );
}
