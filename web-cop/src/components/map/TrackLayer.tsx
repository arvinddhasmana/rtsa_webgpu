// CLASSIFICATION: UNCLASSIFIED
// src/components/map/TrackLayer.tsx

import React from "react";
import { FusedTrack } from "../../types/track";
import { getHostileColor } from "../../utils/mil-symbology";

interface TrackLayerProps {
  tracks: FusedTrack[];
  onTrackClick: (trackId: string) => void;
}

/**
 * TrackLayer — renders track markers with MIL-STD-2525 symbology.
 * Used as an overlay layer within MapView.
 */
export const TrackLayer: React.FC<TrackLayerProps> = ({ tracks, onTrackClick }) => {
  return (
    <div data-testid="track-layer">
      {tracks.map((track) => (
        <div
          key={track.trackId}
          data-testid={`track-marker-${track.trackId}`}
          onClick={() => onTrackClick(track.trackId)}
          style={{
            position: "absolute",
            width: "12px",
            height: "12px",
            borderRadius: "50%",
            backgroundColor: getHostileColor(track.hostileClass),
            border: "2px solid white",
            cursor: "pointer",
            opacity: track.status === "STALE" ? 0.5 : 1,
          }}
          title={`${track.trackId} (${track.hostileClass})`}
        />
      ))}
    </div>
  );
};
