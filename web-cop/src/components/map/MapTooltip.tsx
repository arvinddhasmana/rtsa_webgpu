// CLASSIFICATION: UNCLASSIFIED
// src/components/map/MapTooltip.tsx

import React from "react";
import { useTrackStore } from "../../stores/trackStore";

interface MapTooltipProps {
  trackId: string;
  x: number;
  y: number;
}

export const MapTooltip: React.FC<MapTooltipProps> = ({ trackId, x, y }) => {
  const track = useTrackStore((s) => s.getTrackById(trackId));

  if (!track) return null;

  return (
    <div
      data-testid={`map-tooltip-${trackId}`}
      style={{
        position: "absolute",
        left: x + 15,
        top: y + 15,
        backgroundColor: "var(--glass-bg)",
        backdropFilter: "var(--glass-blur)",
        border: "1px solid var(--glass-border)",
        borderRadius: "4px",
        padding: "8px 12px",
        color: "#F1F5F9",
        pointerEvents: "none",
        zIndex: 50,
        boxShadow: "0 4px 6px -1px rgba(0, 0, 0, 0.5)",
        display: "flex",
        flexDirection: "column",
        gap: "4px",
      }}
    >
      <div style={{ display: "flex", justifyContent: "space-between", gap: "16px" }}>
        <span style={{ fontWeight: "bold", color: "#60A5FA" }}>{track.trackId}</span>
        <span style={{ fontSize: "0.8rem", color: "#9CA3AF" }}>
          {Math.round(track.confidenceScore * 100)}% Conf
        </span>
      </div>

      <div style={{ display: "flex", gap: "8px", fontSize: "0.8rem" }}>
        <span style={{ color: "#E2E8F0" }}>{track.entityType}</span>
        <span style={{ color: "#475569" }}>•</span>
        <span style={{
          color: track.hostileClass === "HOSTILE" ? "#DC2626" :
                 track.hostileClass === "FRIENDLY" ? "#2563EB" :
                 track.hostileClass === "UNKNOWN" ? "#CA8A04" : "#16A34A"
        }}>
          {track.hostileClass.replace("_", " ")}
        </span>
      </div>
    </div>
  );
};
