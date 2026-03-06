// CLASSIFICATION: UNCLASSIFIED
// src/components/nato/InboundTracksPanel.tsx
//
// Allied track list received from NATO N-COP.
// Displays inbound tracks with NATO icon and REL TO classification labels.

import React, { useState } from "react";

interface InboundTrack {
  trackId: string;
  releasableTo: string[];
  entityType: string;
  hostileClass: string;
  position?: { latitude: number; longitude: number };
  confidence: number;
  sourceNation: string;
  lastUpdated: Date;
}

interface InboundTracksPanelProps {
  /** Live inbound tracks from NATO partners (prop-driven or demo data) */
  tracks?: InboundTrack[];
}

const DEMO_INBOUND: InboundTrack[] = [
  {
    trackId: "NATO-TRK-0922",
    releasableTo: ["CAN", "GBR", "USA"],
    entityType: "AIR",
    hostileClass: "UNKNOWN",
    position: { latitude: 44.5, longitude: -59.2 },
    confidence: 0.82,
    sourceNation: "GBR",
    lastUpdated: new Date(Date.now() - 25_000),
  },
  {
    trackId: "NATO-TRK-0714",
    releasableTo: ["CAN", "FRA", "DEU"],
    entityType: "SURFACE",
    hostileClass: "NEUTRAL",
    position: { latitude: 46.1, longitude: -62.0 },
    confidence: 0.94,
    sourceNation: "FRA",
    lastUpdated: new Date(Date.now() - 90_000),
  },
  {
    trackId: "NATO-TRK-0501",
    releasableTo: ["CAN", "USA"],
    entityType: "SUBSURFACE",
    hostileClass: "HOSTILE",
    position: { latitude: 43.8, longitude: -57.6 },
    confidence: 0.67,
    sourceNation: "USA",
    lastUpdated: new Date(Date.now() - 180_000),
  },
];

const HOSTILE_COLOR: Record<string, string> = {
  HOSTILE:  "#EF4444",
  NEUTRAL:  "#10B981",
  FRIENDLY: "#3B82F6",
  UNKNOWN:  "#F59E0B",
};

/**
 * InboundTracksPanel — Allied track list from NATO N-COP.
 *
 * Shows inbound tracks with:
 *   - NATO ⊕ icon, source nation
 *   - REL TO releasability labels
 *   - Hostile classification colour
 *   - Confidence and time-since-update
 */
export const InboundTracksPanel: React.FC<InboundTracksPanelProps> = ({
  tracks = DEMO_INBOUND,
}) => {
  const [search, setSearch] = useState("");

  const filtered = tracks.filter(
    (t) =>
      t.trackId.toLowerCase().includes(search.toLowerCase()) ||
      t.sourceNation.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div
      data-testid="inbound-tracks-panel"
      style={{
        display: "flex",
        flexDirection: "column",
        height: "100%",
        overflow: "hidden",
      }}
    >
      {/* Header */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: "10px",
          padding: "8px 12px",
          borderBottom: "1px solid #334155",
        }}
      >
        <span
          style={{
            fontSize: "0.7rem",
            fontWeight: "bold",
            color: "#A855F7",
          }}
        >
          ⊕ INBOUND ALLIED TRACKS ({tracks.length})
        </span>
        <input
          placeholder="Search…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          data-testid="inbound-search"
          style={{
            marginLeft: "auto",
            backgroundColor: "#1E293B",
            border: "1px solid #334155",
            borderRadius: "4px",
            color: "#F1F5F9",
            padding: "3px 8px",
            fontSize: "0.65rem",
            width: "120px",
          }}
        />
      </div>

      {/* List */}
      <div style={{ flex: 1, overflowY: "auto" }}>
        {filtered.map((track) => {
          const age = Math.round((Date.now() - track.lastUpdated.getTime()) / 1000);
          const ageStr = age < 60 ? `${age}s ago` : `${Math.round(age / 60)}m ago`;
          const hColor = HOSTILE_COLOR[track.hostileClass] ?? "#64748B";

          return (
            <div
              key={track.trackId}
              data-testid={`inbound-track-${track.trackId}`}
              style={{
                display: "flex",
                flexDirection: "column",
                padding: "8px 12px",
                borderBottom: "1px solid rgba(255,255,255,0.04)",
                gap: "4px",
              }}
            >
              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "baseline",
                }}
              >
                <span
                  style={{
                    fontFamily: "monospace",
                    fontSize: "0.7rem",
                    color: "#60A5FA",
                    fontWeight: "bold",
                  }}
                >
                  ⊕ {track.trackId}
                </span>
                <span style={{ fontSize: "0.6rem", color: "#475569" }}>
                  {ageStr}
                </span>
              </div>

              <div
                style={{
                  display: "flex",
                  gap: "8px",
                  alignItems: "center",
                  fontSize: "0.65rem",
                  color: "#94A3B8",
                }}
              >
                <span
                  style={{ color: hColor, fontWeight: "bold", minWidth: "60px" }}
                >
                  {track.hostileClass}
                </span>
                <span>{track.entityType}</span>
                <span style={{ color: "#475569" }}>|</span>
                <span style={{ color: "#10B981" }}>
                  {(track.confidence * 100).toFixed(0)}% conf
                </span>
                <span style={{ color: "#475569" }}>|</span>
                <span style={{ color: "#A855F7" }}>
                  {track.sourceNation}
                </span>
              </div>

              {/* REL TO labels */}
              <div style={{ display: "flex", gap: "4px", flexWrap: "wrap" }}>
                {track.releasableTo.map((nation) => (
                  <span
                    key={nation}
                    style={{
                      fontSize: "0.55rem",
                      fontWeight: "bold",
                      padding: "1px 5px",
                      borderRadius: "3px",
                      backgroundColor: "rgba(168, 85, 247, 0.15)",
                      color: "#A855F7",
                      border: "1px solid rgba(168,85,247,0.3)",
                    }}
                  >
                    REL {nation}
                  </span>
                ))}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
