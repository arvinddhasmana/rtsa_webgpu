// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/StatusBar.tsx

import React, { useEffect, useState } from "react";
import { useConnectionStatus } from "../../hooks/useConnectionStatus";
import { useAlertStore } from "../../stores/alertStore";
import { useTrackStore } from "../../stores/trackStore";

/**
 * StatusBar — persistent bottom bar showing system-wide indicators.
 * Shared across all roles.
 */
export const StatusBar: React.FC = () => {
  const { status: connectionState } = useConnectionStatus();
  const trackCount = useTrackStore((s) => s.tracks.size);
  const alertCount = useAlertStore((s) => s.alerts.size);
  const [utcTime, setUtcTime] = useState(formatUTC());

  useEffect(() => {
    const timer = setInterval(() => setUtcTime(formatUTC()), 1000);
    return () => clearInterval(timer);
  }, []);

  return (
    <div className="ds-status-bar" data-testid="status-bar" role="status">
      <div className="ds-status-bar__item">
        <span
          className={`ds-status-dot ds-status-dot--${
            connectionState === "connected"
              ? "ok"
              : connectionState === "degraded"
                ? "warn"
                : "error"
          }`}
        />
        <span>
          {connectionState === "connected"
            ? "CONNECTED"
            : connectionState === "degraded"
              ? "DEGRADED"
              : "OFFLINE"}
        </span>
      </div>
      <div className="ds-status-bar__item">
        <span style={{ color: "var(--ds-accent-cyan)" }}>⬡</span>
        <span>
          Tracks: <strong>{trackCount}</strong>
        </span>
      </div>
      <div className="ds-status-bar__item">
        <span style={{ color: "var(--ds-accent-amber)" }}>⚠</span>
        <span>
          Alerts: <strong>{alertCount}</strong>
        </span>
      </div>
      <div style={{ flex: 1 }} />
      <div className="ds-status-bar__item" data-testid="status-utc-time">
        <span style={{ fontFamily: "var(--ds-font-mono)", letterSpacing: "0.05em" }}>
          {utcTime}
        </span>
      </div>
    </div>
  );
};

function formatUTC(): string {
  const now = new Date();
  return now.toISOString().replace("T", " ").substring(0, 19) + "Z";
}
