// CLASSIFICATION: UNCLASSIFIED
// src/components/detail/SourceAttribution.tsx

import React from "react";
import { FusedTrack } from "../../types/track";
import { relativeTime } from "../../utils/time";

interface SourceAttributionProps {
  track: FusedTrack;
}

/**
 * SourceAttribution — lists contributing sensors with confidence per source.
 */
export const SourceAttributionSection: React.FC<SourceAttributionProps> = ({ track }) => {
  return (
    <div data-testid="source-attribution" style={{ padding: "8px" }}>
      <div style={{ fontSize: "0.7rem", color: "#9CA3AF", marginBottom: "4px" }}>
        Contributing Sensors ({track.sourceCount})
      </div>
      {track.sources.length === 0 ? (
        <div style={{ fontSize: "0.75rem", color: "#6B7280" }}>No sources</div>
      ) : (
        track.sources.map((src) => (
          <div
            key={src.sensorId}
            style={{
              display: "flex",
              justifyContent: "space-between",
              fontSize: "0.7rem",
              padding: "2px 0",
              borderBottom: "1px solid #1E293B",
            }}
          >
            <span style={{ fontFamily: "monospace", color: "#60A5FA" }}>
              {src.sensorId}
            </span>
            <span style={{ color: "#9CA3AF" }}>{src.sensorType}</span>
            <span>{Math.round(src.confidence * 100)}%</span>
            <span style={{ color: "#6B7280" }}>{relativeTime(src.lastContribution)}</span>
          </div>
        ))
      )}
    </div>
  );
};
