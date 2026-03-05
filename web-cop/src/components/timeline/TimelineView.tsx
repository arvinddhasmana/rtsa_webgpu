// CLASSIFICATION: UNCLASSIFIED
// src/components/timeline/TimelineView.tsx
// Entity event timeline — wired to the currently selected track.

import React, { useState } from "react";
import { useEventTimeline } from "../../hooks/useEventTimeline";
import { useTrackStore } from "../../stores/trackStore";
import { formatZuluTime } from "../../utils/time";

type TimelineFilter = "ALL" | "TRACK" | "ANOMALY" | "FEEDBACK" | "ALERT";

export const TimelineView: React.FC = () => {
  const selectedTrackId = useTrackStore((s) => s.selectedTrackId);
  const [filter, setFilter] = useState<TimelineFilter>("ALL");

  const { events, loading, refreshing, error } = useEventTimeline(selectedTrackId, filter);

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
          color: "#94A3B8",
          fontSize: "0.85rem",
          textAlign: "center",
          gap: "12px",
        }}
      >
        <span style={{ fontSize: "2rem", opacity: 0.5 }}>⏱</span>
        <p>Select a track from the Map or Alert Queue to view its historical event timeline.</p>
      </div>
    );
  }

  return (
    <div
      data-testid="timeline-view"
      style={{ display: "flex", flexDirection: "column", height: "100%" }}
    >
      {/* Header & Filter Chips */}
      <div
        style={{
          padding: "10px 14px 8px",
          borderBottom: "1px solid rgba(255,255,255,0.05)",
          display: "flex",
          flexDirection: "column",
          gap: "8px",
        }}
      >
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
          <div>
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
              {selectedTrackId.slice(0, 18)}...
            </div>
          </div>
          {refreshing && (
             <div style={{ fontSize: "0.6rem", color: "#94A3B8", animation: "pulse 1.5s infinite" }}>
               Refreshing...
             </div>
          )}
        </div>

        {/* Filter Chips */}
        <div style={{ display: "flex", gap: "4px", flexWrap: "wrap", marginTop: "4px" }}>
          {(["ALL", "TRACK", "ANOMALY", "FEEDBACK", "ALERT"] as TimelineFilter[]).map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              style={{
                padding: "2px 8px",
                fontSize: "0.6rem",
                fontWeight: "bold",
                borderRadius: "4px",
                backgroundColor: filter === f ? "rgba(96, 165, 250, 0.2)" : "rgba(255,255,255,0.05)",
                color: filter === f ? "#60A5FA" : "#94A3B8",
                border: `1px solid ${filter === f ? "rgba(96, 165, 250, 0.5)" : "transparent"}`,
                cursor: "pointer",
                transition: "all 0.2s",
              }}
            >
              {f}
            </button>
          ))}
        </div>

        {error && (
          <div style={{ fontSize: "0.6rem", color: "#F43F5E", marginTop: "4px" }}>
            ⚠ {error}. Retrying...
          </div>
        )}
      </div>

      {/* Event list */}
      <div style={{ flex: 1, overflowY: "auto", padding: "8px 0" }}>
        {loading ? (
          <div
            style={{
              padding: "24px",
              textAlign: "center",
              color: "#64748B",
              fontSize: "0.75rem",
            }}
          >
            Loading timeline events...
          </div>
        ) : events.length === 0 ? (
          <div
            style={{
              padding: "24px",
              textAlign: "center",
              color: "#64748B",
              fontSize: "0.75rem",
            }}
          >
            No {filter !== "ALL" ? filter.toLowerCase() : ""} events found in the last 24h.
          </div>
        ) : (
          events.map((evt, idx) => {
            const ts = evt.eventTime
              ? formatZuluTime(new Date(Number(evt.eventTime.seconds) * 1000))
              : "Unknown";

            return (
              <div
                key={evt.id || idx}
                style={{
                  display: "flex",
                  gap: "12px",
                  padding: "8px 14px",
                  borderLeft: `3px solid ${evt.typeColor}`,
                  backgroundColor: idx % 2 === 0 ? "rgba(255,255,255,0.02)" : "transparent",
                  borderBottom: "1px solid rgba(255,255,255,0.03)",
                }}
              >
                {/* Icon Column */}
                <div
                  style={{
                    display: "flex",
                    flexDirection: "column",
                    alignItems: "center",
                    gap: "2px",
                    flexShrink: 0,
                  }}
                >
                  <span style={{ fontSize: "1rem" }}>{evt.icon}</span>
                </div>

                {/* Content Column */}
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
                        color: evt.typeColor,
                        letterSpacing: "0.02em",
                      }}
                    >
                      {evt.eventTypeStr.replace(/_/g, " ")}
                    </span>
                    <span
                      style={{
                        fontSize: "0.6rem",
                        color: "#64748B",
                        fontFamily: "monospace",
                        flexShrink: 0,
                        marginLeft: "8px",
                      }}
                    >
                      {ts}
                    </span>
                  </div>
                  <div
                    style={{
                      fontSize: "0.7rem",
                      color: "#CBD5E1",
                      marginTop: "4px",
                      lineHeight: "1.4"
                    }}
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
