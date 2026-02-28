// CLASSIFICATION: UNCLASSIFIED
// src/components/map/MapLayerToggle.tsx

import React, { useState } from "react";
import { useUIStore } from "../../stores/uiStore";

export const MapLayerToggle: React.FC = () => {
  const [isOpen, setIsOpen] = useState(false);
  const layerVisibility = useUIStore((s) => s.layerVisibility);
  const toggleLayerVisibility = useUIStore((s) => s.toggleLayerVisibility);

  return (
    <div style={{ position: "absolute", top: "16px", left: "16px", zIndex: 10 }}>
      {/* Floating Toggle Button */}
      <button
        data-testid="layer-toggle-button"
        onClick={() => setIsOpen(!isOpen)}
        style={{
          width: "40px",
          height: "40px",
          borderRadius: "8px",
          backgroundColor: isOpen ? "#3B82F6" : "var(--glass-bg)",
          backdropFilter: "var(--glass-blur)",
          border: "var(--glass-border)",
          color: "#F1F5F9",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          cursor: "pointer",
          boxShadow: "0 4px 6px -1px rgba(0, 0, 0, 0.5)",
          transition: "all 0.2s"
        }}
        title="Map Layers"
      >
        <span style={{ fontSize: "1.2rem" }}>☰</span>
      </button>

      {/* Layer Checklist Menu */}
      {isOpen && (
        <div style={{
          position: "absolute",
          top: "48px",
          left: "0",
          backgroundColor: "#1E293B",
          border: "1px solid #334155",
          borderRadius: "8px",
          padding: "12px",
          minWidth: "200px",
          boxShadow: "0 10px 15px -3px rgba(0, 0, 0, 0.5)",
          display: "flex",
          flexDirection: "column",
          gap: "8px"
        }}>
          <h4 style={{ margin: "0 0 8px 0", fontSize: "0.8rem", color: "#9CA3AF", textTransform: "uppercase", letterSpacing: "0.05em" }}>Layer Controls</h4>

          <LayerCheckbox
            label="Geofences"
            checked={layerVisibility.geofences}
            onChange={() => toggleLayerVisibility("geofences")}
          />
          <LayerCheckbox
            label="Track Trails"
            checked={layerVisibility.trackTrails}
            onChange={() => toggleLayerVisibility("trackTrails")}
          />
          <LayerCheckbox
            label="Track Labels"
            checked={layerVisibility.trackLabels}
            onChange={() => toggleLayerVisibility("trackLabels")}
          />
          <LayerCheckbox
            label="Sensor Coverage"
            checked={layerVisibility.sensorCoverage}
            onChange={() => toggleLayerVisibility("sensorCoverage")}
          />
          <LayerCheckbox
            label="MGRS Grid"
            checked={layerVisibility.mgrsGrid}
            onChange={() => toggleLayerVisibility("mgrsGrid")}
          />
        </div>
      )}
    </div>
  );
};

const LayerCheckbox: React.FC<{ label: string; checked: boolean; onChange: () => void }> = ({ label, checked, onChange }) => (
  <label style={{ display: "flex", alignItems: "center", gap: "8px", cursor: "pointer", fontSize: "0.85rem", color: "#F1F5F9" }}>
    <input
      type="checkbox"
      checked={checked}
      onChange={onChange}
      style={{ accentColor: "#3B82F6", cursor: "pointer" }}
    />
    {label}
  </label>
);
