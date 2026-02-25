// CLASSIFICATION: UNCLASSIFIED
// src/components/alerts/AlertCard.tsx

import React from "react";
import { AnomalyAlert } from "../../types/alert";
import { formatZuluTime } from "../../utils/time";
import { useAlertStore } from "../../stores/alertStore";
import { useTrackStore } from "../../stores/trackStore";
import { useUIStore } from "../../stores/uiStore";

interface AlertCardProps {
  alert: AnomalyAlert;
}

const SEVERITY_COLORS: Record<string, string> = {
  CRITICAL: "#DC2626",
  ELEVATED: "#EA580C",
  WATCH: "#CA8A04",
  NORMAL: "#6B7280",
};

/**
 * AlertCard — individual alert card in the alert list.
 *
 * CRITICAL alerts pulse with red border animation.
 * Click to acknowledge, click track ID to open DetailPanel.
 */
export const AlertCard: React.FC<AlertCardProps> = ({ alert }) => {
  const acknowledgeAlert = useAlertStore((s) => s.acknowledgeAlert);
  const acknowledgedIds = useAlertStore((s) => s.acknowledgedIds);
  const selectTrack = useTrackStore((s) => s.selectTrack);
  const toggleDetailPanel = useUIStore((s) => s.toggleDetailPanel);
  const detailPanelOpen = useUIStore((s) => s.detailPanelOpen);

  const isAcknowledged = acknowledgedIds.has(alert.alertId);
  const isCritical = alert.severity === "CRITICAL";
  const color = SEVERITY_COLORS[alert.severity] ?? "#6B7280";

  const handleTrackClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    selectTrack(alert.trackId);
    if (!detailPanelOpen) toggleDetailPanel();
  };

  const handleAcknowledge = () => {
    acknowledgeAlert(alert.alertId);
  };

  return (
    <div
      data-testid={`alert-card-${alert.alertId}`}
      style={{
        margin: "4px 8px",
        padding: "8px",
        backgroundColor: "#0F172A",
        border: `1px solid ${color}`,
        borderRadius: "4px",
        opacity: isAcknowledged ? 0.5 : 1,
        animation: isCritical && !isAcknowledged ? "pulse 1.5s infinite" : undefined,
        cursor: "pointer",
      }}
      onClick={handleAcknowledge}
    >
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
        <span
          style={{
            fontSize: "0.7rem",
            fontWeight: "bold",
            color,
            letterSpacing: "0.05em",
          }}
        >
          {alert.severity}
        </span>
        <span style={{ fontSize: "0.65rem", color: "#9CA3AF" }}>
          {formatZuluTime(alert.detectedAt)}
        </span>
      </div>
      <div style={{ fontSize: "0.75rem", marginTop: "4px", color: "#E2E8F0" }}>
        {alert.anomalyType.replace("_", " ")}
      </div>
      <div style={{ display: "flex", justifyContent: "space-between", marginTop: "4px" }}>
        <button
          data-testid={`track-link-${alert.trackId}`}
          onClick={handleTrackClick}
          style={{
            background: "none",
            border: "none",
            color: "#60A5FA",
            fontSize: "0.7rem",
            cursor: "pointer",
            padding: 0,
            textDecoration: "underline",
          }}
        >
          {alert.trackId}
        </button>
        <span style={{ fontSize: "0.65rem", color: "#9CA3AF" }}>
          {Math.round(alert.confidenceScore * 100)}% conf
        </span>
      </div>
      {!isAcknowledged && (
        <div style={{ fontSize: "0.65rem", color: "#6B7280", marginTop: "4px" }}>
          Click to acknowledge
        </div>
      )}
    </div>
  );
};
