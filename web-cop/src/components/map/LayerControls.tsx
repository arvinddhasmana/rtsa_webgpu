// CLASSIFICATION: UNCLASSIFIED
// src/components/map/LayerControls.tsx

import React from "react";
import { LayerKey, useUIStore } from "../../stores/uiStore";

const LAYERS: { key: LayerKey; label: string }[] = [
  { key: "trackLabels", label: "Track Labels" },
  { key: "trackTrails", label: "Track Trails" },
  { key: "sensorCoverage", label: "Sensor Coverage" },
  { key: "geofences", label: "Geo-fences" },
  { key: "mgrsGrid", label: "MGRS Grid" },
];

/**
 * LayerControls — map overlay to toggle visibility of different MapLibre layers.
 */
export const LayerControls: React.FC = () => {
  const layerVisibility = useUIStore((s) => s.layerVisibility);
  const toggleLayerVisibility = useUIStore((s) => s.toggleLayerVisibility);

  return (
    <div
      data-testid="layer-controls"
      style={{
        position: "absolute",
        bottom: "16px",
        right: "16px",
        backgroundColor: "#1E293B",
        border: "1px solid #334155",
        borderRadius: "4px",
        padding: "8px",
        display: "flex",
        flexDirection: "column",
        gap: "6px",
        zIndex: 10,
        boxShadow: "0 4px 6px -1px rgba(0, 0, 0, 0.1)",
      }}
    >
      <div style={{ fontSize: "0.75rem", fontWeight: "bold", color: "#F1F5F9", marginBottom: "4px" }}>
        LAYERS
      </div>
      {LAYERS.map(({ key, label }) => (
        <label
          key={key}
          style={{
            display: "flex",
            alignItems: "center",
            gap: "8px",
            fontSize: "0.75rem",
            color: "#9CA3AF",
            cursor: "pointer",
          }}
        >
          <input
            type="checkbox"
            data-testid={`layer-toggle-${key}`}
            checked={layerVisibility[key]}
            onChange={() => toggleLayerVisibility(key)}
            style={{ cursor: "pointer" }}
          />
          {label}
        </label>
      ))}
    </div>
  );
};
