// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/MultiDomainDashboard.tsx

import React from "react";
import { useUIStore } from "../../stores/uiStore";
import { DomainMetricsOverlay } from "../dashboard/DomainMetricsOverlay";
import { DetailPanel } from "../detail/DetailPanel";
import { MapView } from "../map/MapView";
import { CollapsiblePane } from "./CollapsiblePane";

export const MultiDomainDashboard: React.FC = () => {
  const detailPanelOpen = useUIStore((s) => s.detailPanelOpen);

  return (
    <div style={{ flex: 1, display: "flex", overflow: "hidden", position: "relative" }}>

      {/* Maximize map as the primary Multi-Domain surface */}
      <div style={{ flex: 1, overflow: "hidden", position: "absolute", top: 0, left: 0, right: 0, bottom: 0 }} aria-label="Map View" role="region" tabIndex={0}>
        <MapView />
      </div>

      {/* Floating Domain KPIs Overlay */}
      <DomainMetricsOverlay />

      {/* Detail Panel drops in from bottom if a track is clicked */}
      {detailPanelOpen && (
        <div style={{ position: "absolute", bottom: 0, left: 0, right: 0, zIndex: 10, padding: "16px", pointerEvents: "none" }}>
          <div style={{ pointerEvents: "auto", margin: "0 auto", maxWidth: "1200px" }}>
             <CollapsiblePane title="Target Exploitation / Detail" width="100%" height="320px" direction="vertical">
                <DetailPanel />
             </CollapsiblePane>
          </div>
        </div>
      )}
    </div>
  );
};
