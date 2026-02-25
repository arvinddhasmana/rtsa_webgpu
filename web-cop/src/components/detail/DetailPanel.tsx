// CLASSIFICATION: UNCLASSIFIED
// src/components/detail/DetailPanel.tsx

import React, { useState } from "react";
import { useTrackStore } from "../../stores/trackStore";
import { useAuthStore } from "../../stores/authStore";
import { IdentitySection } from "./IdentitySection";
import { PositionSection } from "./PositionSection";
import { SourceAttributionSection } from "./SourceAttribution";
import { EntityTimeline } from "./EntityTimeline";
import { FeedbackForm } from "./FeedbackForm";

type ActiveTab = "identity" | "position" | "sources" | "timeline" | "feedback";

/**
 * DetailPanel — shows full details for the selected track entity.
 *
 * Sections (tabs):
 *   1. Identity: track_id, entity_type, hostile_class, confidence, status
 *   2. Position: lat/lon (DMS), altitude, speed, heading, last update
 *   3. Sources: contributing sensors with confidence per source
 *   4. Timeline: chronological history of track updates
 *   5. Feedback: operator feedback submission form
 *
 * Classification: greys out if above operator clearance.
 */
export const DetailPanel: React.FC = () => {
  const selectedTrackId = useTrackStore((s) => s.selectedTrackId);
  const getTrackById = useTrackStore((s) => s.getTrackById);
  const canAccess = useAuthStore((s) => s.canAccess);
  const [activeTab, setActiveTab] = useState<ActiveTab>("identity");

  if (!selectedTrackId) {
    return (
      <div
        data-testid="detail-panel-empty"
        style={{
          padding: "16px",
          color: "#9CA3AF",
          fontSize: "0.8rem",
          textAlign: "center",
        }}
      >
        Select a track to view details
      </div>
    );
  }

  const track = getTrackById(selectedTrackId);

  if (!track) {
    return (
      <div data-testid="detail-panel-not-found" style={{ padding: "16px", color: "#DC2626", fontSize: "0.8rem" }}>
        Track not found
      </div>
    );
  }

  if (!canAccess(track.classification)) {
    return (
      <div data-testid="detail-panel-restricted" style={{ padding: "16px", color: "#6B7280", fontSize: "0.8rem" }}>
        RESTRICTED — insufficient clearance to view this track
      </div>
    );
  }

  const TABS: { key: ActiveTab; label: string }[] = [
    { key: "identity", label: "Identity" },
    { key: "position", label: "Position" },
    { key: "sources", label: "Sources" },
    { key: "timeline", label: "Timeline" },
    { key: "feedback", label: "Feedback" },
  ];

  return (
    <div data-testid="detail-panel" style={{ height: "100%", display: "flex", flexDirection: "column" }}>
      {/* Tab bar */}
      <div
        style={{
          display: "flex",
          borderBottom: "1px solid #334155",
          backgroundColor: "#0F172A",
        }}
      >
        {TABS.map((tab) => (
          <button
            key={tab.key}
            data-testid={`tab-${tab.key}`}
            onClick={() => setActiveTab(tab.key)}
            style={{
              padding: "6px 12px",
              fontSize: "0.7rem",
              fontWeight: activeTab === tab.key ? "bold" : "normal",
              backgroundColor:
                activeTab === tab.key ? "#1E293B" : "transparent",
              color: activeTab === tab.key ? "#F1F5F9" : "#9CA3AF",
              border: "none",
              borderBottom:
                activeTab === tab.key ? "2px solid #3B82F6" : "2px solid transparent",
              cursor: "pointer",
              letterSpacing: "0.05em",
            }}
          >
            {tab.label.toUpperCase()}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div style={{ flex: 1, overflowY: "auto" }}>
        {activeTab === "identity" && <IdentitySection track={track} />}
        {activeTab === "position" && <PositionSection track={track} />}
        {activeTab === "sources" && <SourceAttributionSection track={track} />}
        {activeTab === "timeline" && <EntityTimeline track={track} />}
        {activeTab === "feedback" && (
          <FeedbackForm trackId={track.trackId} />
        )}
      </div>
    </div>
  );
};
