// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/FusionDashboard.tsx

import React from "react";
import { useSensorStream } from "../../hooks/useSensorStream";
import { useUIStore } from "../../stores/uiStore";
import { DetailPanel } from "../detail/DetailPanel";
import { FusionSidePanel } from "../fusion/FusionSidePanel";
import { MapView } from "../map/MapView";
import { CollapsiblePane } from "./CollapsiblePane";

export const FusionDashboard: React.FC = () => {
  const detailPanelOpen = useUIStore((s) => s.detailPanelOpen);

  // Hook to connect the raw sensor stream (since it's only active in Fusion Dashboard)
  useSensorStream();

  return (
    <div style={{ flex: 1, display: "flex", overflow: "hidden" }}>
      {/* Left Data Panel */}
      <CollapsiblePane title="Multi-Sensor Fusion" width="30%" height="100%" direction="horizontal">
        <FusionSidePanel />
      </CollapsiblePane>

      {/* Main Area: Map + Bottom Detail Panel */}
      <div style={{ flex: 1, display: "flex", flexDirection: "column", overflow: "hidden", position: "relative" }}>
        <div style={{ flex: 1, overflow: "hidden" }} aria-label="Map View" role="region" tabIndex={0}>
          <MapView />
        </div>

        {/* Bottom Detail Panel Overlay (Absolute positioned or flex column) */}
        {detailPanelOpen && (
          <CollapsiblePane title="Track Details" width="100%" height="280px" direction="vertical">
            <DetailPanel />
          </CollapsiblePane>
        )}
      </div>
    </div>
  );
};
