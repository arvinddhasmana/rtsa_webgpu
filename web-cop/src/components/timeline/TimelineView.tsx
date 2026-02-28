// CLASSIFICATION: UNCLASSIFIED
// src/components/timeline/TimelineView.tsx

import React, { useEffect, useState } from "react";
import { queryClient } from "../../api/query-client";
import { formatZuluTime } from "../../utils/time";

export const TimelineView: React.FC = () => {
  const [events, setEvents] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    let mounted = true;

    const fetchTimeline = async () => {
      setLoading(true);
      try {
        const toMs = Date.now();
        const fromMs = toMs - 24 * 60 * 60 * 1000; // Last 24 hours

        const res = await queryClient.getEventTimeline({
          trackId: "TRACK-DEMO", // Temporary hardcoded for demo till selected track is wired
          timeRange: {
            startTime: { seconds: BigInt(Math.floor(fromMs / 1000)), nanos: 0 } as any,
            endTime: { seconds: BigInt(Math.floor(toMs / 1000)), nanos: 0 } as any,
          } as any,
          maxEvents: 50,
        });

        if (mounted) {
          setEvents(res.events || []);
        }
      } catch (err) {
        console.error("Timeline API error:", err);
      } finally {
        if (mounted) setLoading(false);
      }
    };

    fetchTimeline();
    // Poll every 10 seconds for new events
    const interval = setInterval(fetchTimeline, 10000);

    return () => {
      mounted = false;
      clearInterval(interval);
    };
  }, []);

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100%", padding: "8px" }}>
      <h3 style={{ fontSize: "0.85rem", color: "var(--color-accent-amber)", marginBottom: "8px" }}>
        Target Activity Timeline
      </h3>
      {loading && events.length === 0 ? (
        <div style={{ color: "#9CA3AF", fontSize: "0.75rem" }}>Loading timeline...</div>
      ) : (
        <div style={{ flex: 1, overflowY: "auto", display: "flex", flexDirection: "column", gap: "8px" }}>
          {events.length === 0 ? (
            <div style={{ color: "#9CA3AF", fontSize: "0.75rem" }}>No events in the last 24h</div>
          ) : (
            events.map((evt, idx) => (
              <div
                key={evt.eventId || idx}
                style={{
                  padding: "8px",
                  backgroundColor: "rgba(255, 255, 255, 0.05)",
                  borderLeft: `3px solid ${getEventColor(evt.eventType)}`,
                  borderRadius: "0 4px 4px 0",
                  fontSize: "0.75rem",
                }}
              >
                <div style={{ display: "flex", justifyContent: "space-between", marginBottom: "4px" }}>
                  <strong style={{ color: "#E2E8F0" }}>{evt.eventType.replace("_", " ")}</strong>
                  <span style={{ color: "#9CA3AF" }}>
                    {evt.eventTime ? formatZuluTime(new Date(Number(evt.eventTime.seconds) * 1000)) : "Unknown Time"}
                  </span>
                </div>
                <div style={{ color: "#CBD5E1", fontSize: "0.7rem", marginTop: "2px" }}>
                  Source: {evt.sourceSystem} | ID: {evt.trackId || evt.eventId}
                </div>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
};

function getEventColor(type: string): string {
  if (type.includes("ANOMALY")) return "var(--color-accent-amber)";
  if (type.includes("FEEDBACK")) return "var(--color-accent-blue)";
  if (type.includes("TRACK")) return "#10B981"; // Emerald
  return "#64748B"; // Slate
}
