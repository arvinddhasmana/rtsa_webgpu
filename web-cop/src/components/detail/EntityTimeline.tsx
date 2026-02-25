// CLASSIFICATION: UNCLASSIFIED
// src/components/detail/EntityTimeline.tsx

import React from "react";
import { FusedTrack } from "../../types/track";
import { formatZulu } from "../../utils/time";

interface EntityTimelineProps {
  track: FusedTrack;
}

/**
 * EntityTimeline — chronological history of track updates and alerts.
 */
export const EntityTimeline: React.FC<EntityTimelineProps> = ({ track }) => {
  const events = [
    { time: track.createdAt, label: "Track created", type: "create" },
    { time: track.updatedAt, label: "Last update", type: "update" },
  ].sort((a, b) => b.time.getTime() - a.time.getTime());

  return (
    <div data-testid="entity-timeline" style={{ padding: "8px" }}>
      <div style={{ fontSize: "0.7rem", color: "#9CA3AF", marginBottom: "4px" }}>
        Timeline
      </div>
      {events.map((event, idx) => (
        <div
          key={idx}
          style={{
            display: "flex",
            gap: "8px",
            fontSize: "0.7rem",
            padding: "3px 0",
            borderLeft: "2px solid #334155",
            paddingLeft: "8px",
            marginLeft: "4px",
          }}
        >
          <span style={{ color: "#6B7280", fontFamily: "monospace", minWidth: "160px" }}>
            {formatZulu(event.time)}
          </span>
          <span style={{ color: event.type === "create" ? "#16A34A" : "#60A5FA" }}>
            {event.label}
          </span>
        </div>
      ))}
    </div>
  );
};
