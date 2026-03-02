// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/OperatorDashboard.tsx

import React from "react";
import { useUIStore } from "../../stores/uiStore";
import { AlertPanel } from "../alerts/AlertPanel";
import { DetailPanel } from "../detail/DetailPanel";
import { MapView } from "../map/MapView";
import { TimelineView } from "../timeline/TimelineView";
import { CollapsiblePane } from "./CollapsiblePane";

export const OperatorDashboard: React.FC = () => {
  const detailPanelOpen = useUIStore((s) => s.detailPanelOpen);

  return (
    <div style={{ flex: 1, display: "flex", overflow: "hidden", position: "relative" }}>
      {/* Background Map with Blur */}
      <div style={{ position: "absolute", top: 0, left: 0, right: 0, bottom: 0, zIndex: 0 }}>
        <MapView />
        <div style={{ position: "absolute", top: 0, left: 0, right: 0, bottom: 0, backgroundColor: "rgba(15, 23, 42, 0.6)", backdropFilter: "blur(4px)" }} />
      </div>

      {/* Floating Interface */}
      <div style={{ display: "flex", flex: 1, zIndex: 1, padding: "16px", gap: "16px", pointerEvents: "none" }}>

        {/* Left Side: Timeline (Vertical Panel) */}
        <div style={{ flex: 1, maxWidth: "400px", display: "flex", flexDirection: "column", pointerEvents: "auto" }}>
          <CollapsiblePane title="Entity Timeline" width="100%" height="100%" direction="horizontal">
            <TimelineView />
          </CollapsiblePane>
        </div>

        {/* Center: Open Space for clicking map through the transparent gap. Could host DetailPanel later */}
        <div style={{ flex: 2, pointerEvents: "auto", display: "flex", flexDirection: "column", justifyContent: "flex-end" }}>
           {detailPanelOpen && (
              <CollapsiblePane title="Entity Detail" width="100%" height="320px" direction="vertical">
                <DetailPanel />
              </CollapsiblePane>
           )}
        </div>

        {/* Right Side: Alerts */}
        <div style={{ flex: 1, maxWidth: "420px", display: "flex", flexDirection: "column", pointerEvents: "auto" }}>
          <CollapsiblePane title="Alert Triage Queue" width="100%" height="100%" direction="horizontal">
            <AlertPanel />
          </CollapsiblePane>
        </div>

      </div>
    </div>
  );
};
