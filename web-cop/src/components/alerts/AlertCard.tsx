// CLASSIFICATION: UNCLASSIFIED
// src/components/alerts/AlertCard.tsx

import React, { useState } from "react";
import { feedbackClient } from "../../api/feedback-client";
import { useAlertStore } from "../../stores/alertStore";
import { useAuthStore } from "../../stores/authStore";
import { useTrackStore } from "../../stores/trackStore";
import { useUIStore } from "../../stores/uiStore";
import { AnomalyAlert } from "../../types/alert";
import { formatZuluTime } from "../../utils/time";

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
export const AlertCard: React.FC<AlertCardProps> = React.memo(({ alert }) => {
  const acknowledgeAlert = useAlertStore((s) => s.acknowledgeAlert);
  const acknowledgedIds = useAlertStore((s) => s.acknowledgedIds);
  const selectTrack = useTrackStore((s) => s.selectTrack);
  const toggleDetailPanel = useUIStore((s) => s.toggleDetailPanel);
  const detailPanelOpen = useUIStore((s) => s.detailPanelOpen);
  const operatorId = useAuthStore((s) => s.operatorId);

  const [feedbackStatus, setFeedbackStatus] = useState<string | null>(null);

  const isAcknowledged = acknowledgedIds.has(alert.alertId);
  const isCritical = alert.severity === "CRITICAL";
  const color = SEVERITY_COLORS[alert.severity] ?? "#6B7280";

  const handleTrackClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    selectTrack(alert.trackId);
    if (!detailPanelOpen) toggleDetailPanel();
  };

  const handleAcknowledge = () => {
    if (!isAcknowledged) {
      acknowledgeAlert(alert.alertId);
    }
  };

  const handleInspect = (e: React.MouseEvent) => {
    e.stopPropagation();
    handleTrackClick(e);
  };

  const handleFeedback = async (e: React.MouseEvent, type: "CONFIRM_ANOMALY" | "REJECT_ANOMALY") => {
    e.stopPropagation();
    const effectiveOperatorId = operatorId || "test-operator-1";
    setFeedbackStatus("Submitting...");
    try {
      await feedbackClient.submitFeedback({
        trackId: alert.trackId,
        alertId: alert.alertId,
        feedbackType: type,
        justification: "Quick action from alert card",
        operatorId: effectiveOperatorId,
      });
      setFeedbackStatus(type === "CONFIRM_ANOMALY" ? "Confirmed" : "Rejected");
    } catch (err) {
      setFeedbackStatus("Error");
    }
  };

  const handleAssign = (e: React.MouseEvent) => {
    e.stopPropagation();
    setFeedbackStatus("Assigned");
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
        animation:
          isCritical && !isAcknowledged ? "pulse 1.5s infinite" : undefined,
        cursor: "pointer",
      }}
      onClick={handleAcknowledge}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
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
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          marginTop: "4px",
        }}
      >
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

      <div
        style={{
          display: "flex",
          gap: "4px",
          marginTop: "8px",
          borderTop: "1px solid #334155",
          paddingTop: "8px",
        }}
      >
        <button
          data-testid={`alert-inspect-${alert.alertId}`}
          onClick={handleInspect}
          style={{ ...actionButtonStyle, backgroundColor: "#1D4ED8" }}
        >
          [Inspect]
        </button>
        <button
          data-testid={`alert-confirm-${alert.alertId}`}
          onClick={(e) => handleFeedback(e, "CONFIRM_ANOMALY")}
          style={{ ...actionButtonStyle, backgroundColor: "#16A34A" }}
        >
          [Confirm]
        </button>
        <button
          data-testid={`alert-reject-${alert.alertId}`}
          onClick={(e) => handleFeedback(e, "REJECT_ANOMALY")}
          style={{ ...actionButtonStyle, backgroundColor: "#DC2626" }}
        >
          [Reject]
        </button>
        <button
          data-testid={`alert-assign-${alert.alertId}`}
          onClick={handleAssign}
          style={{ ...actionButtonStyle, backgroundColor: "#CA8A04" }}
        >
          [Assign]
        </button>
      </div>

      {feedbackStatus && (
        <div style={{ fontSize: "0.65rem", color: "#60A5FA", marginTop: "4px" }}>
          Status: {feedbackStatus}
        </div>
      )}

      {!isAcknowledged && !feedbackStatus && (
        <div
          style={{ fontSize: "0.65rem", color: "#6B7280", marginTop: "4px" }}
        >
          Click to acknowledge
        </div>
      )}
    </div>
  );
});

const actionButtonStyle: React.CSSProperties = {
  flex: 1,
  padding: "2px 0",
  color: "#F1F5F9",
  border: "none",
  borderRadius: "2px",
  fontSize: "0.6rem",
  cursor: "pointer",
  fontWeight: "bold",
};
