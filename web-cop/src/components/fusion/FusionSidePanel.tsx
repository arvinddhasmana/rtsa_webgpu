// CLASSIFICATION: UNCLASSIFIED
// src/components/fusion/FusionSidePanel.tsx

import React from "react";
import { useSensorStore } from "../../stores/sensorStore";
import { useTrackStore } from "../../stores/trackStore";

export const FusionSidePanel: React.FC = () => {
  const currentTracksMap = useTrackStore((s) => s.tracks);
  const currentObservations = useSensorStore((s) => s.rawObservations);

  const totalTracks = currentTracksMap.size;
  const hostileTracks = Array.from(currentTracksMap.values()).filter(t => t.hostileClass === "HOSTILE").length;

  const correlatedSensors = Array.from(currentObservations.values()).filter(o => o.correlatedTrackId).length;
  const rawSensors = currentObservations.size;

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100%", padding: "16px", color: "#F1F5F9" }}>
      <h3 style={{ fontSize: "0.85rem", color: "var(--color-accent-amber)", marginBottom: "16px" }}>
        Fusion Engine Telemetry
      </h3>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "12px", marginBottom: "24px" }}>
        <MetricBox label="Active Tracks" value={totalTracks.toString()} />
        <MetricBox label="Hostile Targets" value={hostileTracks.toString()} color="#DC2626" />
        <MetricBox label="Raw Observations (/10s)" value={rawSensors.toString()} />
        <MetricBox label="Correlated Obs." value={correlatedSensors.toString()} color="#3B82F6" />
      </div>

      <h4 style={{ fontSize: "0.75rem", color: "#9CA3AF", marginBottom: "8px", textTransform: "uppercase" }}>
        Sensor Contributions
      </h4>
      <div style={{ flex: 1, backgroundColor: "rgba(0,0,0,0.2)", borderRadius: "4px", padding: "8px" }}>
        {/* Placeholder for a breakdown of which sensors contributed most recently */}
        <div style={{ fontSize: "0.7rem", color: "#64748B", textAlign: "center", marginTop: "24px" }}>
          Streaming live sensor integration metrics...
        </div>
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
    <span style={{ fontSize: "1.5rem", fontWeight: "bold", color: color || "#F1F5F9" }}>{value}</span>
    <span style={{ fontSize: "0.65rem", color: "#9CA3AF", textAlign: "center", marginTop: "4px" }}>{label}</span>
  </div>
);
