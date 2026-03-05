// CLASSIFICATION: UNCLASSIFIED
// src/components/detail/EntityTimeline.tsx

import React from "react";
import { useAlertStore } from "../../stores/alertStore";
import { FusedTrack } from "../../types/track";
import { formatZulu } from "../../utils/time";

interface EntityTimelineProps {
  track: FusedTrack;
}

/**
 * EntityTimeline — chronological history of track updates and anomalies.
 */
export const EntityTimeline: React.FC<EntityTimelineProps> = ({ track }) => {
  const getAlertsByTrack = useAlertStore(s => s.getAlertsByTrack);
  const trackAlerts = getAlertsByTrack(track.trackId);

  // Combine track lifecycle events with anomaly alerts
  const events = [
    {
      id: `create-${track.trackId}`,
      time: track.createdAt,
      label: "Track created",
      type: "create",
      color: "#16A34A"
    },
    {
      id: `update-${track.trackId}`,
      time: track.updatedAt,
      label: "Last update",
      type: "update",
      color: "#60A5FA"
    },
    ...trackAlerts.map(alert => ({
      id: alert.alertId,
      time: alert.detectedAt,
      label: `${alert.anomalyType.replace("_", " ")}`,
      type: "anomaly",
      color: alert.severity === "CRITICAL" ? "#DC2626" :
             alert.severity === "ELEVATED" ? "#EA580C" : "#CA8A04",
      score: alert.confidenceScore
    }))
  ].sort((a, b) => b.time.getTime() - a.time.getTime());

  return (
    <div data-testid="entity-timeline" style={{ padding: "16px" }}>
      <div style={{ fontSize: "0.75rem", color: "#94A3B8", marginBottom: "12px", borderBottom: "1px solid rgba(255,255,255,0.05)", paddingBottom: "4px" }}>
        Recent History
      </div>

      <div style={{ display: "flex", flexDirection: "column", gap: "2px" }}>
        {events.map((event, idx) => (
          <div
            key={event.id}
            style={{
              display: "flex",
              gap: "12px",
              padding: "8px 0",
              borderBottom: idx < events.length - 1 ? "1px solid rgba(255,255,255,0.05)" : "none",
            }}
          >
            <div style={{ display: "flex", flexDirection: "column", alignItems: "center", width: "16px" }}>
               <div style={{ width: "8px", height: "8px", borderRadius: "50%", backgroundColor: event.color, marginTop: "4px" }} />
               {idx < events.length - 1 && <div style={{ width: "2px", flex: 1, backgroundColor: "rgba(255,255,255,0.1)", marginTop: "4px" }} />}
            </div>

            <div style={{ display: "flex", flexDirection: "column", flex: 1, minWidth: 0 }}>
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline" }}>
                <span style={{ color: event.color, fontWeight: "bold", fontSize: "0.75rem" }}>
                  {event.label}
                </span>
                <span style={{ color: "#64748B", fontFamily: "monospace", fontSize: "0.65rem" }}>
                  {formatZulu(event.time)}
                </span>
              </div>

              {event.type === "anomaly" && "score" in event && event.score !== undefined && (
                <div style={{ fontSize: "0.65rem", color: "#94A3B8", marginTop: "2px" }}>
                  Confidence: {Math.round(event.score * 100)}%
                </div>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
