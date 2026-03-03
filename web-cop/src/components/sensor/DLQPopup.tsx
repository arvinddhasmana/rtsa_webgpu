// CLASSIFICATION: UNCLASSIFIED
// src/components/sensor/DLQPopup.tsx

import React, { useEffect, useRef } from "react";
import { getDLQSummary, useSensorHealthStore } from "../../stores/sensorHealthStore";

/**
 * DLQPopup — floating popup showing DLQ pie chart and recent rejection events
 * for a specific sensor. Triggered by clicking the DLQ icon in the sensor table row.
 * Styled like Variant B's floating drill-down panel with glassmorphism.
 */
export const DLQPopup: React.FC = () => {
  const dlqPopupSensorId = useSensorHealthStore((s) => s.dlqPopupSensorId);
  const dlqEvents = useSensorHealthStore((s) => s.dlqEvents);
  const setDLQPopupSensorId = useSensorHealthStore(
    (s) => s.setDLQPopupSensorId
  );
  const popupRef = useRef<HTMLDivElement>(null);

  // Close on click outside
  useEffect(() => {
    if (!dlqPopupSensorId) return;
    const handleClick = (e: MouseEvent) => {
      if (popupRef.current && !popupRef.current.contains(e.target as Node)) {
        setDLQPopupSensorId(null);
      }
    };
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === "Escape") setDLQPopupSensorId(null);
    };
    document.addEventListener("mousedown", handleClick);
    document.addEventListener("keydown", handleEscape);
    return () => {
      document.removeEventListener("mousedown", handleClick);
      document.removeEventListener("keydown", handleEscape);
    };
  }, [dlqPopupSensorId, setDLQPopupSensorId]);

  if (!dlqPopupSensorId) return null;

  const summary = getDLQSummary(dlqEvents, dlqPopupSensorId);
  const recentEvents = dlqEvents
    .filter((e) => e.sensorId === dlqPopupSensorId)
    .slice(0, 8);

  const reasonEntries = Object.entries(summary.byReason);
  const totalReasons = reasonEntries.reduce((acc, [, v]) => acc + v, 0);

  // Pie chart colours
  const pieColors = [
    "#EF4444",
    "#F59E0B",
    "#3B82F6",
    "#A855F7",
    "#10B981",
    "#EC4899",
  ];

  // Generate SVG pie chart
  let currentAngle = 0;
  const pieSlices = reasonEntries.map(([reason, count], i) => {
    const percentage = totalReasons > 0 ? count / totalReasons : 0;
    const angle = percentage * 360;
    const slice = createPieSlice(50, 50, 40, currentAngle, currentAngle + angle);
    currentAngle += angle;
    return { reason, count, percentage, path: slice, color: pieColors[i % pieColors.length] };
  });

  return (
    <div
      ref={popupRef}
      className="ds-popup"
      data-testid="dlq-popup"
      style={{
        position: "fixed",
        top: "50%",
        left: "50%",
        transform: "translate(-50%, -50%)",
        width: "420px",
        maxHeight: "520px",
        zIndex: 100,
      }}
    >
      {/* Header */}
      <div
        style={{
          padding: "12px 16px",
          borderBottom: "1px solid var(--ds-border-default)",
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          background: "var(--ds-bg-secondary)",
        }}
      >
        <div>
          <div style={{ fontWeight: 600, fontSize: "var(--ds-text-base)" }}>
            DLQ Analysis
          </div>
          <div
            style={{
              fontSize: "var(--ds-text-xs)",
              color: "var(--ds-text-secondary)",
            }}
          >
            {dlqPopupSensorId} — {summary.totalCount} rejections
          </div>
        </div>
        <button
          className="ds-btn-icon"
          onClick={() => setDLQPopupSensorId(null)}
          aria-label="Close DLQ popup"
        >
          ✕
        </button>
      </div>

      {/* Body */}
      <div
        style={{
          padding: "16px",
          display: "flex",
          flexDirection: "column",
          gap: "16px",
          maxHeight: "440px",
          overflowY: "auto",
          background: "var(--ds-bg-primary)",
        }}
        className="ds-scroll"
      >
        {/* Pie Chart + Legend */}
        {summary.totalCount > 0 ? (
          <div style={{ display: "flex", gap: "16px", alignItems: "center" }}>
            <svg
              width="100"
              height="100"
              viewBox="0 0 100 100"
              role="img"
              aria-label="DLQ rejection reason breakdown"
            >
              {pieSlices.map((s, i) => (
                <path
                  key={i}
                  d={s.path}
                  fill={s.color}
                  stroke="var(--ds-bg-primary)"
                  strokeWidth="1"
                />
              ))}
              {/* Center hole for donut effect */}
              <circle
                cx="50"
                cy="50"
                r="22"
                fill="var(--ds-bg-primary)"
              />
              <text
                x="50"
                y="48"
                textAnchor="middle"
                fill="var(--ds-text-primary)"
                fontSize="14"
                fontWeight="700"
                fontFamily="var(--ds-font-mono)"
              >
                {summary.totalCount}
              </text>
              <text
                x="50"
                y="60"
                textAnchor="middle"
                fill="var(--ds-text-muted)"
                fontSize="7"
              >
                TOTAL
              </text>
            </svg>

            <div style={{ display: "flex", flexDirection: "column", gap: "6px", flex: 1 }}>
              {pieSlices.map((s, i) => (
                <div
                  key={i}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: "8px",
                    fontSize: "var(--ds-text-xs)",
                  }}
                >
                  <span
                    style={{
                      width: "8px",
                      height: "8px",
                      borderRadius: "2px",
                      backgroundColor: s.color,
                      flexShrink: 0,
                    }}
                  />
                  <span
                    style={{ flex: 1, color: "var(--ds-text-secondary)" }}
                  >
                    {s.reason}
                  </span>
                  <span
                    style={{
                      fontFamily: "var(--ds-font-mono)",
                      color: "var(--ds-text-primary)",
                    }}
                  >
                    {s.count}
                  </span>
                </div>
              ))}
            </div>
          </div>
        ) : (
          <div
            style={{
              textAlign: "center",
              color: "var(--ds-text-muted)",
              padding: "20px",
            }}
          >
            No rejections recorded
          </div>
        )}

        {/* Pattern indicator */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "8px",
            padding: "8px 12px",
            borderRadius: "var(--ds-radius-md)",
            background: "rgba(255, 255, 255, 0.03)",
            border: "1px solid var(--ds-border-subtle)",
          }}
        >
          <span
            style={{
              fontSize: "var(--ds-text-xs)",
              color: "var(--ds-text-secondary)",
            }}
          >
            Pattern:
          </span>
          <span
            className={`ds-badge ds-badge--${
              summary.pattern === "sustained"
                ? "error"
                : summary.pattern === "burst"
                  ? "warn"
                  : "info"
            }`}
          >
            {summary.pattern.toUpperCase()}
          </span>
        </div>

        {/* Recent Events */}
        <div>
          <div className="ds-section-header">Recent Rejections</div>
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              gap: "6px",
            }}
          >
            {recentEvents.length > 0 ? (
              recentEvents.map((evt) => (
                <div
                  key={evt.eventId}
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "flex-start",
                    padding: "6px 8px",
                    borderRadius: "var(--ds-radius-sm)",
                    background: "rgba(255, 255, 255, 0.02)",
                    fontSize: "var(--ds-text-xs)",
                  }}
                >
                  <div>
                    <div style={{ color: "var(--ds-status-error)" }}>
                      {evt.rejectionReason}
                    </div>
                    {evt.details && (
                      <div
                        style={{
                          color: "var(--ds-text-muted)",
                          fontSize: "var(--ds-text-2xs)",
                          marginTop: "2px",
                        }}
                      >
                        {evt.details}
                      </div>
                    )}
                  </div>
                  <span
                    style={{
                      color: "var(--ds-text-muted)",
                      fontFamily: "var(--ds-font-mono)",
                      fontSize: "var(--ds-text-2xs)",
                      flexShrink: 0,
                      marginLeft: "8px",
                    }}
                  >
                    {formatTimeAgo(evt.timestamp)}
                  </span>
                </div>
              ))
            ) : (
              <div
                style={{
                  color: "var(--ds-text-muted)",
                  fontSize: "var(--ds-text-xs)",
                  textAlign: "center",
                  padding: "12px",
                }}
              >
                No recent events
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

// ── Helpers ──────────────────────────────────────────

function createPieSlice(
  cx: number,
  cy: number,
  r: number,
  startAngle: number,
  endAngle: number
): string {
  if (endAngle - startAngle >= 359.99) {
    // Full circle
    return `M${cx - r},${cy} A${r},${r} 0 1,1 ${cx + r},${cy} A${r},${r} 0 1,1 ${cx - r},${cy}`;
  }
  const startRad = ((startAngle - 90) * Math.PI) / 180;
  const endRad = ((endAngle - 90) * Math.PI) / 180;
  const x1 = cx + r * Math.cos(startRad);
  const y1 = cy + r * Math.sin(startRad);
  const x2 = cx + r * Math.cos(endRad);
  const y2 = cy + r * Math.sin(endRad);
  const largeArc = endAngle - startAngle > 180 ? 1 : 0;
  return `M${cx},${cy} L${x1},${y1} A${r},${r} 0 ${largeArc},1 ${x2},${y2} Z`;
}

function formatTimeAgo(date: Date): string {
  const diff = Date.now() - date.getTime();
  if (diff < 60_000) return `${Math.floor(diff / 1000)}s ago`;
  if (diff < 3600_000) return `${Math.floor(diff / 60_000)}m ago`;
  if (diff < 86400_000) return `${Math.floor(diff / 3600_000)}h ago`;
  return `${Math.floor(diff / 86400_000)}d ago`;
}
