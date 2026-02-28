// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/SensorHealthDashboard.tsx

import React from "react";
import { useSensorCoverage } from "../../hooks/useSensorCoverage";
import { MapView } from "../map/MapView";
import { CollapsiblePane } from "./CollapsiblePane";

export const SensorHealthDashboard: React.FC = () => {
  const sensors = useSensorCoverage();

  const activeSensors = sensors.filter(s => s.connected).length;
  const inactiveSensors = sensors.length - activeSensors;

  const totalEvents = sensors.reduce((acc, s) => acc + BigInt(s.totalReceived), 0n);

  return (
    <div style={{ flex: 1, display: "flex", overflow: "hidden" }}>
      {/* Left Data Panel */}
      <CollapsiblePane title="Sensor Grid Health" width="35%" height="100%" direction="horizontal">
        <div style={{ padding: "16px", color: "#F1F5F9", display: "flex", flexDirection: "column", gap: "16px", height: "100%", overflowY: "auto" }}>

          {/* Header Metrics */}
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "12px" }}>
            <MetricBox label="Active Nodes" value={activeSensors.toString()} color="#10B981" />
            <MetricBox label="Degraded/Offline" value={inactiveSensors.toString()} color={inactiveSensors > 0 ? "#DC2626" : "#64748B"} />
            <MetricBox label="Total Processed" value={(Number(totalEvents) / 1000).toFixed(1) + "k"} />
            <MetricBox label="Grid Throughput" value="~120 EPS" color="#3B82F6" />
          </div>

          <h4 style={{ fontSize: "0.85rem", color: "#9CA3AF", marginTop: "16px", borderBottom: "1px solid #334155", paddingBottom: "4px" }}>
            Node Status
          </h4>

          {/* Node List */}
          <div style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
            {sensors.map(sensor => (
              <div key={sensor.sensorId} style={{
                backgroundColor: "var(--glass-bg)",
                border: "var(--glass-border)",
                padding: "12px",
                borderRadius: "4px",
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center"
              }}>
                <div>
                  <div style={{ fontWeight: "bold", fontSize: "0.85rem" }}>{sensor.sensorId}</div>
                  <div style={{ fontSize: "0.7rem", color: "#9CA3AF" }}>Type: {sensor.sensorType}</div>
                  <div style={{ fontSize: "0.7rem", color: "#64748B" }}>Throughput: {sensor.eventsPerSecond.toFixed(1)} EPS</div>
                </div>
                <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
                   <div style={{
                     width: "12px", height: "12px", borderRadius: "50%",
                     backgroundColor: sensor.connected ? "#10B981" : "#DC2626"
                   }} />
                   <span style={{ fontSize: "0.75rem", color: sensor.connected ? "#10B981" : "#DC2626" }}>
                     {sensor.connected ? "ONLINE" : "OFFLINE"}
                   </span>
                </div>
              </div>
            ))}

            {sensors.length === 0 && (
              <div style={{ color: "#64748B", fontSize: "0.8rem", textAlign: "center", marginTop: "20px" }}>
                Awaiting sensor telemetry...
              </div>
            )}
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
    <span style={{ fontSize: "1.25rem", fontWeight: "bold", color: color || "#F1F5F9" }}>{value}</span>
    <span style={{ fontSize: "0.65rem", color: "#9CA3AF", textAlign: "center", marginTop: "4px" }}>{label}</span>
  </div>
);
