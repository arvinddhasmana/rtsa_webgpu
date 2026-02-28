// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/NatoExchangeDashboard.tsx

import React from "react";
import { useTrackStore } from "../../stores/trackStore";
import { MapView } from "../map/MapView";
import { CollapsiblePane } from "./CollapsiblePane";

export const NatoExchangeDashboard: React.FC = () => {
  const currentTracksMap = useTrackStore((s) => s.tracks);

  // Simulated NATOB STANAG-4609 / Link 16 exchange metrics
  const sharedTracks = Array.from(currentTracksMap.values()).filter(t => t.confidenceScore > 0.8).length;

  return (
    <div style={{ flex: 1, display: "flex", overflow: "hidden" }}>
      {/* Left Data Panel */}
      <CollapsiblePane title="NATO BICES Exchange" width="35%" height="100%" direction="horizontal">
        <div style={{ padding: "16px", color: "#F1F5F9", display: "flex", flexDirection: "column", gap: "16px", height: "100%", overflowY: "auto" }}>

          {/* Header Metrics */}
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "12px" }}>
            <MetricBox label="Link 16 Status" value="ACTIVE" color="#10B981" />
            <MetricBox label="STANAG 4609" value="SYNCED" color="#10B981" />
            <MetricBox label="Tracks Exported" value={sharedTracks.toString()} color="#3B82F6" />
            <MetricBox label="Tracks Imported" value="14" color="#A855F7" />
          </div>

          <h4 style={{ fontSize: "0.85rem", color: "#9CA3AF", marginTop: "16px", borderBottom: "1px solid #334155", paddingBottom: "4px" }}>
            Manual Track Nomination
          </h4>

          <div style={{
            backgroundColor: "var(--glass-bg)",
            border: "var(--glass-border)",
            padding: "16px",
            borderRadius: "4px"
          }}>
             <div style={{ fontSize: "0.8rem", color: "#9CA3AF", marginBottom: "12px" }}>
               Select a track on the map to nominate it for export to the NATO Common Operational Picture (N-COP).
               Only UNCLASSIFIED or explicitly cleared tracks may be exported.
             </div>

             <button style={{
               width: "100%",
               padding: "8px",
               backgroundColor: "#2563EB",
               color: "white",
               border: "none",
               borderRadius: "4px",
               cursor: "pointer",
               fontWeight: "bold",
               fontSize: "0.8rem"
             }}>
               Nominate Selected Track
             </button>
          </div>

          <h4 style={{ fontSize: "0.85rem", color: "#9CA3AF", marginTop: "16px", borderBottom: "1px solid #334155", paddingBottom: "4px" }}>
            Exchange Logs
          </h4>
          <div style={{ flex: 1, backgroundColor: "#0F172A", padding: "12px", borderRadius: "4px", fontFamily: "monospace", fontSize: "0.7rem", color: "#64748B", overflowY: "auto" }}>
            <div>[10:45:01Z] TX: J3.2 Air Track (Trk 1045)</div>
            <div>[10:45:03Z] RX: ACK from N-COP Gateway</div>
            <div style={{ color: "#10B981" }}>[10:45:15Z] RX: J3.1 Maritime Track (Trk 992)</div>
            <div>[10:45:16Z] TX: Internal COP Update</div>
            <div style={{ color: "#DC2626" }}>[10:46:01Z] TX DROP: Track 1048 (Clearance mismatch)</div>
          </div>
        </div>
      </CollapsiblePane>

      {/* Main Area: Map */}
      <div style={{ flex: 1, overflow: "hidden", position: "relative" }} aria-label="Map View" role="region" tabIndex={0}>
        <MapView />
      </div>
    </div>
  );
};

const MetricBox: React.FC<{ label: string; value: string; color?: string }> = ({ label, value, color }) => (
  <div style={{
    backgroundColor: "var(--glass-bg)",
    border: "var(--glass-border)",
    padding: "12px",
    borderRadius: "4px",
    display: "flex",
    flexDirection: "column",
    alignItems: "center"
  }}>
    <span style={{ fontSize: "1.1rem", fontWeight: "bold", color: color || "#F1F5F9" }}>{value}</span>
    <span style={{ fontSize: "0.65rem", color: "#9CA3AF", textAlign: "center", marginTop: "4px" }}>{label}</span>
  </div>
);
