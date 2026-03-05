// CLASSIFICATION: UNCLASSIFIED
// src/hooks/useEventTimeline.ts

import { useEffect, useState } from "react";
import { TimelineEventType } from "../../../gen/ts/rtsa/query/v1/query_service_pb";
import { queryClient } from "../api/query-client";

export interface LocalTimelineEvent {
  id: string;
  eventTypeStr: string;
  summary: string;
  eventTime?: { seconds: bigint };
  typeColor: string;
  icon: string;
}

const EVENT_TYPE_CONFIG: Record<string, { color: string; icon: string }> = {
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

export const useEventTimeline = (
  trackId: string | null,
  filterType: "ALL" | "TRACK" | "ANOMALY" | "FEEDBACK" | "ALERT" = "ALL",
  pollIntervalMs = 10000
) => {
  const [events, setEvents] = useState<LocalTimelineEvent[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;
    let interval: ReturnType<typeof setInterval> | null = null;

    const fetchTimeline = async (isBackground = false) => {
      if (!trackId) {
        setEvents([]);
        return;
      }

      if (!isBackground) {
        setLoading(true);
      } else {
        setRefreshing(true);
      }

      setError(null);

      try {
        const toMs = Date.now();
        const fromMs = toMs - 24 * 60 * 60 * 1000;

        const res = await queryClient.getEventTimeline({
          trackId: trackId,
          timeRange: {
            startTime: { seconds: BigInt(Math.floor(fromMs / 1000)), nanos: 0 } as any,
            endTime: { seconds: BigInt(Math.floor(toMs / 1000)), nanos: 0 } as any,
          } as any,
          maxEvents: 200, // Fetch more since we filter client-side for now
        });

        if (mounted) {
          let mapped = (res.events || []).map((e: any, i: number) => {
            const rawType = TimelineEventType[e.eventType] || "UNKNOWN";
            const eventTypeStr = rawType.replace("TIMELINE_EVENT_TYPE_", "");
            const cfg = getConfig(eventTypeStr);

            return {
              id: (e.detail?.value as any)?.alertId || (e.detail?.value as any)?.auditId || (e.detail?.value as any)?.feedbackId || `evt-${Date.now()}-${i}`,
              eventTypeStr,
              summary: e.summary || "Event recorded",
              eventTime: e.eventTime,
              typeColor: cfg.color,
              icon: cfg.icon,
            };
          });

          // Apply client-side filtering based on generic categories
          if (filterType !== "ALL") {
            mapped = mapped.filter((evt: LocalTimelineEvent) => {
              if (filterType === "TRACK") return evt.eventTypeStr.includes("TRACK");
              if (filterType === "ANOMALY") return evt.eventTypeStr.includes("ANOMALY");
              if (filterType === "FEEDBACK") return evt.eventTypeStr.includes("FEEDBACK");
              if (filterType === "ALERT") return evt.eventTypeStr.includes("ALERT");
              return true;
            });
          }

          setEvents(mapped);
        }
      } catch (err) {
        if (mounted) setError("Timeline unavailable");
        console.error("Timeline API error:", err);
      } finally {
        if (mounted) {
          setLoading(false);
          setRefreshing(false);
        }
      }
    };

    // Initial fetch
    fetchTimeline();

    // Polling setup
    interval = setInterval(() => fetchTimeline(true), pollIntervalMs);

    return () => {
      mounted = false;
      if (interval) clearInterval(interval);
    };
  }, [trackId, filterType, pollIntervalMs]);

  return { events, loading, refreshing, error };
};
