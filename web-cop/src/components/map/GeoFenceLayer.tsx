// CLASSIFICATION: UNCLASSIFIED
// src/components/map/GeoFenceLayer.tsx

import React from "react";

interface GeoFence {
  id: string;
  name: string;
  type: "exclusion" | "inclusion";
}

interface GeoFenceLayerProps {
  fences: GeoFence[];
}

/**
 * GeoFenceLayer — renders geo-fence polygon overlays on the map.
 */
export const GeoFenceLayer: React.FC<GeoFenceLayerProps> = ({ fences }) => {
  return (
    <div data-testid="geofence-layer">
      {fences.map((fence) => (
        <div
          key={fence.id}
          data-testid={`geofence-${fence.id}`}
          style={{
            position: "absolute",
            border: `2px solid ${fence.type === "exclusion" ? "#DC2626" : "#16A34A"}`,
            opacity: 0.5,
            pointerEvents: "none",
          }}
          title={fence.name}
        />
      ))}
    </div>
  );
};
