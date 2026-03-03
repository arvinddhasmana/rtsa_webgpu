// CLASSIFICATION: UNCLASSIFIED
// src/components/timeline/TimelineView.tsx
// Entity event timeline — wired to the currently selected track.

import React, { useEffect, useState } from "react";
import { queryClient } from "../../api/query-client";
import { useTrackStore } from "../../stores/trackStore";
import { formatZuluTime } from "../../utils/time";

import { TimelineEventType } from "../../../../gen/ts/rtsa/query/v1/query_service_pb";

interface LocalTimelineEvent {
  id: string;
  eventTypeStr: string;
  summary: string;
  eventTime?: { seconds: bigint };
}

const EVENT_TYPE_CONFIG: Record<
  string,
  { color: string; icon: string }
> = {
  TRACK_UPDATE: { color: "#10B981", icon: "📡" },
  ANOMALY_DETECTED: { color: "#F59E0B", icon: "⚠️" },
  ANOMALY_RESOLVED: { color: "#10B981", icon: "✅" },
  FEEDBACK_SUBMITTED: { color: "#3B82F6", icon: "💬" },
  TRACK_CREATED: { color: "#6366F1", icon: "🆕" },
  TRACK_CLOSED: { color: "#6B7280", icon: "🔒" },
  ALERT_TRIGGERED: { color: "#EF4444", icon: "🚨" },
};

function getConfig(type: string) {
  return EVENT_TYPE_CONFIG[type] ?? { color: "#64748B", icon: "ℹ️" };
}

export const TimelineView: React.FC = () => {
  const selectedTrackId = useTrackStore((s) => s.selectedTrackId);
  const [events, setEvents] = useState<LocalTimelineEvent[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;
    let interval: ReturnType<typeof setInterval> | null = null;

    const fetchTimeline = async () => {
      if (!selectedTrackId) {
        setEvents([]);
        return;
      }
      setLoading(true);
      setError(null);
      try {
        const toMs = Date.now();
        const fromMs = toMs - 24 * 60 * 60 * 1000;

        const res = await queryClient.getEventTimeline({
          trackId: selectedTrackId,
          timeRange: {
            startTime: { seconds: BigInt(Math.floor(fromMs / 1000)), nanos: 0 } as any,
            endTime: { seconds: BigInt(Math.floor(toMs / 1000)), nanos: 0 } as any,
          } as any,
          maxEvents: 80,
        });

        if (mounted) {
          const mapped = (res.events || []).map((e: any, i: number) => {
            const typeStr = TimelineEventType[e.eventType] || "UNKNOWN";
            return {
              id: (e.detail?.value as any)?.alertId || (e.detail?.value as any)?.auditId || (e.detail?.value as any)?.feedbackId || `evt-${Date.now()}-${i}`,
              eventTypeStr: typeStr.replace("TIMELINE_EVENT_TYPE_", ""),
              summary: e.summary || "Event recorded",
              eventTime: e.eventTime,
            };
          });
          setEvents(mapped);
        }
      } catch (err) {
        if (mounted) setError("Timeline unavailable — showing cached data");
        console.error("Timeline API error:", err);
      } finally {
        if (mounted) setLoading(false);
      }
    };

    fetchTimeline();
    interval = setInterval(fetchTimeline, 10_000);

    return () => {
      mounted = false;
      if (interval) clearInterval(interval);
    };
  }, [selectedTrackId]);

  if (!selectedTrackId) {
    return (
      <div
        data-testid="timeline-empty"
        style={{
          display: "flex",
          flexDirection: "column",
          height: "100%",
          padding: "24px 16px",
          alignItems: "center",
          justifyContent: "center",
          color: "#475569",
          fontSize: "0.8rem",
          textAlign: "center",
          gap: "8px",
        }}
      >
        <span style={{ fontSize: "1.5rem" }}>⏱</span>
        <p>Select a track to view its event timeline</p>
      </div>
    );
  }

  return (
    <div
      data-testid="timeline-view"
      style={{ display: "flex", flexDirection: "column", height: "100%" }}
    >
      {/* Header */}
      <div
        style={{
          padding: "10px 14px 8px",
          borderBottom: "1px solid #1E293B",
        }}
      >
        <div style={{ fontSize: "0.65rem", color: "#64748B", marginBottom: "2px" }}>
          ENTITY TIMELINE
        </div>
        <div
          style={{
            fontFamily: "monospace",
            fontSize: "0.75rem",
            color: "#60A5FA",
            fontWeight: "bold",
          }}
        >
          {selectedTrackId}
        </div>
        {error && (
          <div style={{ fontSize: "0.6rem", color: "#F59E0B", marginTop: "4px" }}>
            ⚠ {error}
          </div>
        )}
      </div>

      {/* Event list */}
      <div style={{ flex: 1, overflowY: "auto", padding: "8px 0" }}>
        {loading && events.length === 0 ? (
          <div
            style={{
              padding: "24px",
              textAlign: "center",
              color: "#475569",
              fontSize: "0.75rem",
            }}
          >
            Loading…
          </div>
        ) : events.length === 0 ? (
          <div
            style={{
              padding: "24px",
              textAlign: "center",
              color: "#475569",
              fontSize: "0.75rem",
            }}
          >
            No events in the last 24h for this track
          </div>
        ) : (
          events.map((evt, idx) => {
            const cfg = getConfig(evt.eventTypeStr);
            const ts = evt.eventTime
              ? formatZuluTime(new Date(Number(evt.eventTime.seconds) * 1000))
              : "Unknown";
            return (
              <div
                key={evt.id || idx}
                style={{
                  display: "flex",
                  gap: "10px",
                  padding: "6px 14px",
                  borderBottom: "1px solid rgba(255,255,255,0.03)",
                }}
              >
                {/* Timeline line */}
                <div
                  style={{
                    display: "flex",
                    flexDirection: "column",
                    alignItems: "center",
                    gap: "2px",
                    flexShrink: 0,
                  }}
                >
                  <span style={{ fontSize: "0.8rem" }}>{cfg.icon}</span>
                  {idx < events.length - 1 && (
                    <div
                      style={{
                        width: "1px",
                        height: "20px",
                        backgroundColor: "#1E293B",
                        flexShrink: 0,
                      }}
                    />
                  )}
                </div>
                {/* Content */}
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div
                    style={{
                      display: "flex",
                      justifyContent: "space-between",
                      alignItems: "baseline",
                    }}
                  >
                    <span
                      style={{
                        fontSize: "0.7rem",
                        fontWeight: "bold",
                        color: cfg.color,
                      }}
                    >
                      {evt.eventTypeStr.replace(/_/g, " ")}
                    </span>
                    <span
                      style={{
                        fontSize: "0.6rem",
                        color: "#475569",
                        fontFamily: "monospace",
                        flexShrink: 0,
                        marginLeft: "8px",
                      }}
                    >
                      {ts}
                    </span>
                  </div>
                  <div
                    style={{ fontSize: "0.65rem", color: "#64748B", marginTop: "2px" }}
                  >
                    {evt.summary}
                  </div>
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
};
