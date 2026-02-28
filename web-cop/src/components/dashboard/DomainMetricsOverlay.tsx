// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/DomainMetricsOverlay.tsx

import React, { useMemo } from "react";
import { useTrackStore } from "../../stores/trackStore";

export const DomainMetricsOverlay: React.FC = () => {
  const currentTracksMap = useTrackStore((s) => s.tracks);

  // Group by Entity Type prefixes implicitly or handle as categories
  const metrics = useMemo(() => {
    let air = 0, surface = 0, sub = 0, land = 0, space = 0, cyber = 0;

    currentTracksMap.forEach(t => {
      const typeStr = t.entityType.toLowerCase();
      if (typeStr.includes("air")) air++;
      else if (typeStr.includes("surface") || typeStr.includes("ship") || typeStr.includes("boat")) surface++;
      else if (typeStr.includes("sub")) sub++;
      else if (typeStr.includes("land") || typeStr.includes("vehicle") || typeStr.includes("person")) land++;
      else if (typeStr.includes("space") || typeStr.includes("satellite")) space++;
      else if (typeStr.includes("cyber") || typeStr.includes("node")) cyber++;
    });

    return [
      { domain: "AIR", count: air, color: "#38BDF8" }, // Light Blue
      { domain: "SURFACE", count: surface, color: "#1E3A8A" }, // Deep Blue
      { domain: "SUB-SURFACE", count: sub, color: "#0F766E" }, // Teal
      { domain: "LAND", count: land, color: "#854D0E" }, // Brown
      { domain: "SPACE", count: space, color: "#A855F7" }, // Purple
      { domain: "CYBER", count: cyber, color: "#10B981" }, // Emerald
    ];
  }, [currentTracksMap]);

  return (
    <div style={{
      position: "absolute",
      top: "16px",
      right: "16px",
      display: "flex",
      flexDirection: "column",
      gap: "8px",
      zIndex: 5,
      pointerEvents: "none" // Let clicks pass through to map
    }}>
      {metrics.map(m => (
        <div key={m.domain} style={{
          backgroundColor: "var(--glass-bg)",
          backdropFilter: "var(--glass-blur)",
          border: "var(--glass-border)",
          borderLeft: `3px solid ${m.color}`,
          borderRadius: "4px",
          padding: "8px 12px",
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          width: "160px",
          pointerEvents: "auto"
        }}>
          <span style={{ fontSize: "0.75rem", fontWeight: "bold", color: "#F1F5F9" }}>{m.domain}</span>
          <span style={{ fontSize: "1.1rem", fontFamily: "monospace", color: m.color }}>{m.count}</span>
        </div>
      ))}
    </div>
  );
};
