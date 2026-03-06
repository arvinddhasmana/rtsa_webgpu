// CLASSIFICATION: UNCLASSIFIED
// src/components/map/GeoFenceLayer.tsx
//
// GeoFenceLayer — renders geo-fence polygon overlays.
//
// Fixes from initial stub:
//   - Semi-transparent green (#22c55e at 15% opacity) fill for inclusion zones
//   - Semi-transparent red (#ef4444 at 12% opacity) fill for exclusion zones
//   - Dashed 2 px border with appropriate colour
//   - Default demo fence centred on Gulf of Oman area (matching simulator)

import React from "react";

export interface GeoFence {
  id: string;
  name: string;
  type: "exclusion" | "inclusion";
  /** Optional WGS-84 bounding box — used for display / tooltip context */
  bounds?: {
    north: number;
    south: number;
    east: number;
    west: number;
  };
}

interface GeoFenceLayerProps {
  fences: GeoFence[];
}

/** Default demo geo-fence — Gulf of Oman operational area */
export const DEFAULT_DEMO_GEOFENCE: GeoFence = {
  id: "default-oparea",
  name: "Gulf of Oman Operational Area",
  type: "inclusion",
  bounds: { north: 26.0, south: 22.0, east: 60.0, west: 56.0 },
};

/**
 * GeoFenceLayer — renders geo-fence polygon overlays on the map.
 *
 * Visual specification:
 *   - inclusion: semi-transparent green fill, dashed green border
 *   - exclusion: semi-transparent red fill, dashed red border
 */
export const GeoFenceLayer: React.FC<GeoFenceLayerProps> = ({ fences }) => {
  return (
    <div data-testid="geofence-layer">
      {fences.map((fence) => {
        const isExclusion = fence.type === "exclusion";
        const fillColor = isExclusion
          ? "rgba(239, 68, 68, 0.12)" // red at 12%
          : "rgba(34, 197, 94, 0.15)"; // green at 15%
        const borderColor = isExclusion ? "#ef4444" : "#22c55e";

        return (
          <div
            key={fence.id}
            data-testid={`geofence-${fence.id}`}
            style={{
              position: "absolute",
              backgroundColor: fillColor,
              border: `2px dashed ${borderColor}`,
              borderRadius: "2px",
              pointerEvents: "none",
              boxSizing: "border-box",
            }}
            title={`${fence.name} (${fence.type})`}
            aria-label={`Geo-fence: ${fence.name}`}
          />
        );
      })}
    </div>
  );
};
