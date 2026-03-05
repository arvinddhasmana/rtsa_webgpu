// CLASSIFICATION: UNCLASSIFIED
// src/components/detail/SourceAttribution.tsx

import React from "react";
import { FusedTrack } from "../../types/track";
import { relativeTime } from "../../utils/time";

interface SourceAttributionProps {
  track: FusedTrack;
}

const getConfidenceColor = (conf: number) => {
  if (conf >= 0.8) return "#10B981"; // Green
  if (conf >= 0.5) return "#F59E0B"; // Amber
  return "#EF4444"; // Red
};

/**
 * SourceAttribution — lists contributing sensors with confidence bars per source.
 */
export const SourceAttributionSection: React.FC<SourceAttributionProps> = ({ track }) => {
  return (
    <div data-testid="source-attribution" style={{ padding: "16px" }}>
      <div style={{ fontSize: "0.75rem", color: "#94A3B8", marginBottom: "12px", borderBottom: "1px solid rgba(255,255,255,0.05)", paddingBottom: "4px" }}>
        Contributing Sensors ({track.sourceCount})
      </div>
      {track.sources.length === 0 ? (
        <div style={{ fontSize: "0.75rem", color: "#64748B", textAlign: "center", padding: "16px" }}>
          No source data available
        </div>
      ) : (
        <div style={{ display: "flex", flexDirection: "column", gap: "12px" }}>
          {track.sources.map((src) => {
             const color = getConfidenceColor(src.confidence);
             return (
              <div
                key={src.sensorId}
                style={{
                  display: "flex",
                  flexDirection: "column",
                  gap: "4px",
                  padding: "8px",
                  backgroundColor: "rgba(255,255,255,0.02)",
                  borderRadius: "4px",
                  border: "1px solid rgba(255,255,255,0.05)"
                }}
              >
                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                  <span style={{ fontFamily: "monospace", color: "#60A5FA", fontSize: "0.75rem", fontWeight: "bold" }}>
                    {src.sensorId}
                  </span>
                  <span style={{ color: "#94A3B8", fontSize: "0.65rem", backgroundColor: "rgba(15, 23, 42, 0.5)", padding: "2px 6px", borderRadius: "12px" }}>
                    {src.sensorType}
                  </span>
                </div>

                <div style={{ display: "flex", alignItems: "center", gap: "8px", marginTop: "2px" }}>
                  <span style={{ fontSize: "0.65rem", color: "#CBD5E1", minWidth: "24px" }}>
                    {Math.round(src.confidence * 100)}%
                  </span>
                  <div style={{ flex: 1, height: "4px", backgroundColor: "rgba(255,255,255,0.1)", borderRadius: "2px", overflow: "hidden" }}>
                    <div style={{
                      height: "100%",
                      width: `${src.confidence * 100}%`,
                      backgroundColor: color
                    }} />
                  </div>
                  <span style={{ color: "#64748B", fontSize: "0.65rem", minWidth: "40px", textAlign: "right" }}>
                    {relativeTime(src.lastContribution)}
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
