// CLASSIFICATION: UNCLASSIFIED
// src/components/map/ThreatHaloLayer.tsx

import React from "react";
import { FusedTrack } from "../../types/track";

interface ThreatHaloLayerProps {
  hostileTracks: FusedTrack[];
  radiusKm?: number;
}

/**
 * ThreatHaloLayer — renders proximity circles around hostile tracks.
 */
export const ThreatHaloLayer: React.FC<ThreatHaloLayerProps> = ({
  hostileTracks,
  radiusKm = 50,
}) => {
  return (
    <div data-testid="threat-halo-layer">
      {hostileTracks.map((track) => (
        <div
          key={`halo-${track.trackId}`}
          data-testid={`threat-halo-${track.trackId}`}
          style={{
            position: "absolute",
            borderRadius: "50%",
            border: "2px dashed #DC2626",
            opacity: 0.4,
            pointerEvents: "none",
          }}
          title={`Threat halo: ${radiusKm}km radius`}
        />
      ))}
    </div>
  );
};
