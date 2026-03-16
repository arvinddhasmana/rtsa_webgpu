// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/FusionCommanderDashboard.tsx

import type { JSX } from "solid-js";

interface FusionCommanderDashboardProps {
  mapContent: JSX.Element;
  sidePanelContent?: JSX.Element;
}

/**
 * Operations Commander Fusion dashboard skeleton.
 * Provides stable mount points for map, observation layer, fused layer, and side panel.
 */
export function FusionCommanderDashboard(props: FusionCommanderDashboardProps) {
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
        <div
          data-testid="commander-observation-layer-mount"
          aria-label="Observation layer mount"
          style={{
            position: "absolute",
            inset: "0",
            border: "1px dashed rgba(56,189,248,0.35)",
            "pointer-events": "none",
          }}
        />
        <div
          data-testid="commander-fused-layer-mount"
          aria-label="Fused layer mount"
          style={{
            position: "absolute",
            inset: "12px",
            border: "1px dashed rgba(251,191,36,0.35)",
            "pointer-events": "none",
          }}
        />
      </section>

      <aside
        data-testid="commander-fusion-side-panel"
        aria-label="Fusion side panel container"
        style={{
          width: "22rem",
          "flex-shrink": "0",
          background: "rgba(7, 13, 25, 0.92)",
          "border-left": "1px solid #1e2a3a",
          padding: "0.75rem",
          overflow: "hidden auto",
        }}
      >
        {props.sidePanelContent ?? (
          <div style={{ color: "#94a3b8", "font-size": "0.8rem" }}>
            Fusion metrics panel mount
          </div>
        )}
      </aside>
    </div>
  );
}
