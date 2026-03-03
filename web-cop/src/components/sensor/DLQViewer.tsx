// CLASSIFICATION: UNCLASSIFIED
// src/components/sensor/DLQViewer.tsx

import React from "react";
import {
    getFilteredDLQEvents,
    useSensorHealthStore,
} from "../../stores/sensorHealthStore";

/**
 * DLQViewer — Dead Letter Queue browser with filtering.
 * Shows rejected events as cards with sensor type, ID, timestamp, reason, details.
 */
export const DLQViewer: React.FC = () => {
  const dlqEvents = useSensorHealthStore((s) => s.dlqEvents);
  const dlqFilterSensorType = useSensorHealthStore(
    (s) => s.dlqFilterSensorType
  );
  const dlqFilterReason = useSensorHealthStore((s) => s.dlqFilterReason);
  const dlqFilterTimeRange = useSensorHealthStore((s) => s.dlqFilterTimeRange);
  const setDLQFilter = useSensorHealthStore((s) => s.setDLQFilter);

  const filteredEvents = getFilteredDLQEvents(dlqEvents, {
    sensorType: dlqFilterSensorType,
    reason: dlqFilterReason,
    timeRange: dlqFilterTimeRange,
  });

  // Get unique values for filter dropdowns
  const sensorTypes = [...new Set(dlqEvents.map((e) => e.sensorType))];
  const reasons = [...new Set(dlqEvents.map((e) => e.rejectionReason))];

  return (
    <div
      data-testid="dlq-viewer"
      style={{
        display: "flex",
        flexDirection: "column",
        height: "100%",
        gap: "var(--ds-space-md)",
      }}
    >
      {/* Filter Bar */}
      <div
        style={{
          display: "flex",
          gap: "var(--ds-space-sm)",
          flexWrap: "wrap",
          padding: "0 var(--ds-space-lg)",
          paddingTop: "var(--ds-space-md)",
        }}
      >
        <select
          aria-label="Filter by sensor type"
          value={dlqFilterSensorType}
          onChange={(e) => setDLQFilter({ sensorType: e.target.value })}
          style={filterSelectStyle}
        >
          <option value="">All Sensor Types</option>
          {sensorTypes.map((t) => (
            <option key={t} value={t}>
              {t}
            </option>
          ))}
        </select>

        <select
          aria-label="Filter by rejection reason"
          value={dlqFilterReason}
          onChange={(e) => setDLQFilter({ reason: e.target.value })}
          style={filterSelectStyle}
        >
          <option value="">All Reasons</option>
          {reasons.map((r) => (
            <option key={r} value={r}>
              {r}
            </option>
          ))}
        </select>

        <select
          aria-label="Filter by time range"
          value={dlqFilterTimeRange}
          onChange={(e) =>
            setDLQFilter({
              timeRange: e.target.value as "1h" | "6h" | "24h" | "all",
            })
          }
          style={filterSelectStyle}
        >
          <option value="1h">Last 1 Hour</option>
          <option value="6h">Last 6 Hours</option>
          <option value="24h">Last 24 Hours</option>
          <option value="all">All Time</option>
        </select>

        <div style={{ flex: 1 }} />

        <span
          style={{
            fontSize: "var(--ds-text-xs)",
            color: "var(--ds-text-muted)",
            alignSelf: "center",
          }}
        >
          {filteredEvents.length} event{filteredEvents.length !== 1 ? "s" : ""}
        </span>
      </div>

      {/* Event Cards */}
      <div
        className="ds-scroll"
        style={{
          flex: 1,
          padding: "0 var(--ds-space-lg)",
          paddingBottom: "var(--ds-space-lg)",
          display: "flex",
          flexDirection: "column",
          gap: "var(--ds-space-sm)",
        }}
      >
        {filteredEvents.length > 0 ? (
          filteredEvents.map((evt) => (
            <div
              key={evt.eventId}
              data-testid="dlq-event-card"
              style={{
                padding: "var(--ds-space-md)",
                borderRadius: "var(--ds-radius-md)",
                background: "rgba(255, 255, 255, 0.02)",
                border: "1px solid var(--ds-border-subtle)",
                display: "flex",
                gap: "var(--ds-space-md)",
                alignItems: "flex-start",
              }}
            >
              {/* Type icon */}
              <div
                style={{
                  width: "32px",
                  height: "32px",
                  borderRadius: "var(--ds-radius-sm)",
                  background: "rgba(239, 68, 68, 0.1)",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  fontSize: "14px",
                  flexShrink: 0,
                }}
              >
                {getSensorIcon(evt.sensorType)}
              </div>

              {/* Content */}
              <div style={{ flex: 1, minWidth: 0 }}>
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "flex-start",
                    gap: "var(--ds-space-sm)",
                  }}
                >
                  <div>
                    <span
                      style={{
                        fontWeight: 600,
                        fontSize: "var(--ds-text-sm)",
                        color: "var(--ds-text-primary)",
                      }}
                    >
                      {evt.sensorId}
                    </span>
                    <span
                      style={{
                        marginLeft: "8px",
                        fontSize: "var(--ds-text-2xs)",
                        color: "var(--ds-text-muted)",
                      }}
                    >
                      {evt.sensorType}
                    </span>
                  </div>
                  <span
                    style={{
                      fontSize: "var(--ds-text-2xs)",
                      color: "var(--ds-text-muted)",
                      fontFamily: "var(--ds-font-mono)",
                      flexShrink: 0,
                    }}
                  >
                    {evt.timestamp.toISOString().substring(11, 19)}Z
                  </span>
                </div>

                <div
                  style={{
                    marginTop: "4px",
                    fontSize: "var(--ds-text-xs)",
                    color: "var(--ds-status-error)",
                    fontWeight: 500,
                  }}
                >
                  ⚠ {evt.rejectionReason}
                </div>

                {evt.details && (
                  <div
                    style={{
                      marginTop: "4px",
                      fontSize: "var(--ds-text-2xs)",
                      color: "var(--ds-text-muted)",
                    }}
                  >
                    {evt.details}
                  </div>
                )}

                {evt.rawMessageId && (
                  <div
                    style={{
                      marginTop: "4px",
                      fontSize: "var(--ds-text-2xs)",
                      color: "var(--ds-text-muted)",
                      fontFamily: "var(--ds-font-mono)",
                    }}
                  >
                    MSG: {evt.rawMessageId}
                  </div>
                )}
              </div>
            </div>
          ))
        ) : (
          <div
            style={{
              textAlign: "center",
              color: "var(--ds-text-muted)",
              fontSize: "var(--ds-text-sm)",
              padding: "40px 0",
            }}
          >
            No DLQ events match current filters
          </div>
        )}
      </div>
    </div>
  );
};

const filterSelectStyle: React.CSSProperties = {
  padding: "4px 8px",
  backgroundColor: "var(--ds-bg-tertiary)",
  color: "var(--ds-text-primary)",
  border: "1px solid var(--ds-border-default)",
  borderRadius: "var(--ds-radius-sm)",
  fontSize: "var(--ds-text-xs)",
  cursor: "pointer",
  fontFamily: "var(--ds-font-sans)",
};

function getSensorIcon(type: string): string {
  switch (type.toUpperCase()) {
    case "RADAR":
      return "📡";
    case "EW":
      return "📻";
    case "ELINT":
      return "🔊";
    case "ISR":
      return "🛰️";
    case "AIS":
      return "🚢";
    case "CYBER":
      return "💻";
    default:
      return "📡";
  }
}
