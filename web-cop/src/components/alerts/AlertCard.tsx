// CLASSIFICATION: UNCLASSIFIED
// src/components/alerts/AlertCard.tsx
// Operations Commander — Enhanced Alert Card

import React, { useEffect, useRef } from "react";
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

// Play audio cue for CRITICAL alerts
const playAudioCue = () => {
  try {
    const ctx = new (window.AudioContext || (window as any).webkitAudioContext)();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();

    osc.type = "square";
    osc.frequency.setValueAtTime(800, ctx.currentTime);
    osc.frequency.setValueAtTime(1000, ctx.currentTime + 0.1);

    gain.gain.setValueAtTime(0, ctx.currentTime);
    gain.gain.linearRampToValueAtTime(0.1, ctx.currentTime + 0.05); // low volume
    gain.gain.linearRampToValueAtTime(0, ctx.currentTime + 0.3);

    osc.connect(gain);
    gain.connect(ctx.destination);

    osc.start();
    osc.stop(ctx.currentTime + 0.3);
  } catch (e) {
    // Ignore audio context errors (e.g. before user interaction)
    console.debug("Audio cue prevented:", e);
  }
};

export const AlertCard: React.FC<AlertCardProps> = React.memo(({ alert }) => {
  const acknowledgeAlert = useAlertStore((s) => s.acknowledgeAlert);
  const acknowledgedIds = useAlertStore((s) => s.acknowledgedIds);
  const selectTrack = useTrackStore((s) => s.selectTrack);
  const toggleDetailPanel = useUIStore((s) => s.toggleDetailPanel);
  const openDetailPanel = useUIStore((s) => s.openDetailPanel);
  const detailPanelOpen = useUIStore((s) => s.detailPanelOpen);
  const operatorId = useAuthStore((s) => s.operatorId);

  const feedbackStatuses = useAlertStore((s) => s.feedbackStatuses);
  const setStoreFeedbackStatus = useAlertStore((s) => s.setFeedbackStatus);

  const feedbackStatus = feedbackStatuses.get(alert.alertId) || null;
  const setFeedbackStatus = (status: string) => setStoreFeedbackStatus(alert.alertId, status);
  const hasPlayedAudioRef = useRef(false);
  const cardRef = useRef<HTMLDivElement>(null);

  const isAcknowledged = acknowledgedIds.has(alert.alertId);
  const isCritical = alert.severity === "CRITICAL";
  const color = SEVERITY_COLORS[alert.severity] ?? "#6B7280";

  // Audio cue for new unacknowledged CRITICAL alerts
  useEffect(() => {
    if (isCritical && !isAcknowledged && !hasPlayedAudioRef.current) {
      playAudioCue();
      hasPlayedAudioRef.current = true;
    }
  }, [isCritical, isAcknowledged]);

  const handleTrackClick = (e: React.MouseEvent | React.KeyboardEvent) => {
    e.stopPropagation();
    selectTrack(alert.trackId);
    if (!detailPanelOpen && openDetailPanel) {
      openDetailPanel();
    } else if (!detailPanelOpen) {
      toggleDetailPanel();
    }
  };

  const handleAcknowledge = () => {
    if (!isAcknowledged) {
      acknowledgeAlert(alert.alertId);
    }
  };

  const handleInspect = (e: React.MouseEvent | React.KeyboardEvent) => {
    e.stopPropagation();
    handleTrackClick(e);
  };

  const handleFeedback = async (
    e: React.MouseEvent | React.KeyboardEvent,
    type: "CONFIRM_ANOMALY" | "REJECT_ANOMALY"
  ) => {
    e.stopPropagation();

    // Optimistic UI update
    setFeedbackStatus(type === "CONFIRM_ANOMALY" ? "Confirmed (⏳)" : "Rejected (⏳)");

    const effectiveOperatorId = operatorId || "test-operator-1";
    try {
      const response = await feedbackClient.submitFeedback({
        trackId: alert.trackId,
        alertId: alert.alertId,
        feedbackType: type,
        justification: "Quick action from alert card",
        operatorId: effectiveOperatorId,
      });

      // Full trust score response handling
      if (response && response.trustScore !== undefined) {
         setFeedbackStatus(type === "CONFIRM_ANOMALY"
           ? `✅ Accepted [${Math.round(response.trustScore * 100)}%]`
           : "❌ Rejected");
      } else {
         setFeedbackStatus(type === "CONFIRM_ANOMALY" ? "✅ Confirmed" : "❌ Rejected");
      }

      handleAcknowledge();
    } catch (err) {
      setFeedbackStatus("Error submitting feedback");
    }
  };

  const handleAssign = (e: React.MouseEvent | React.KeyboardEvent) => {
    e.stopPropagation();
    // Dispatch custom event to OperatorDashboard since this is a global action
    const event = new CustomEvent("open-assign-popover", {
      detail: { alertId: alert.alertId }
    });
    window.dispatchEvent(event);
  };

  // Keyboard Shortcuts
  const handleKeyDown = (e: React.KeyboardEvent) => {
    // Only trigger if focus is on the card itself, not children inputs
    if (e.target !== cardRef.current) return;

    switch(e.key) {
      case "Enter":
        handleInspect(e);
        break;
      case "c":
      case "C":
        handleFeedback(e, "CONFIRM_ANOMALY");
        break;
      case "r":
      case "R":
        handleFeedback(e, "REJECT_ANOMALY");
        break;
      case "a":
      case "A":
        handleAssign(e);
        break;
      case " ":
        // Space to acknowledge
        e.preventDefault();
        handleAcknowledge();
        break;
    }
  };

  // Status visual mapping
  let bgColor = "rgba(15, 23, 42, 0.45)"; // default dark glass
  let borderColor = isAcknowledged ? "rgba(255,255,255,0.1)" : color;

  if (feedbackStatus?.includes("✅") || feedbackStatus?.includes("Confirmed")) {
    bgColor = "rgba(22, 163, 74, 0.15)";
    borderColor = "#16A34A";
  } else if (feedbackStatus?.includes("❌") || feedbackStatus?.includes("Rejected")) {
    bgColor = "rgba(220, 38, 38, 0.15)";
    borderColor = "#DC2626";
  } else if (feedbackStatus?.includes("Error")) {
    borderColor = "#F43F5E";
  }

  // Explanation mock (since AnomalyAlert type might not have full explanation text yet)
  const explanation = `Detected deviation in expected kinematic profile for ${alert.anomalyType.replace("_", " ")}. Confidence is ${Math.round(alert.confidenceScore * 100)}% based on multi-sensor correlation.`;

  return (
    <div
      ref={cardRef}
      tabIndex={0} // Make focusable for shortcuts
      onKeyDown={handleKeyDown}
      data-testid={`alert-card-${alert.alertId}`}
      style={{
        margin: "8px",
        padding: "12px",
        backgroundColor: bgColor,
        border: `1px solid ${borderColor}`,
        borderLeft: `4px solid ${borderColor}`,
        borderRadius: "6px",
        opacity: isAcknowledged ? 0.6 : 1,
        animation: isCritical && !isAcknowledged ? "pulse 2s infinite" : undefined,
        cursor: "pointer",
        transition: "all 0.2s ease-out",
        outline: "none", // customize focus state below via CSS class if needed, or rely on browser default for now
      }}
      onClick={handleAcknowledge}
      className="alert-card-focus-target"
    >
      {/* Header Row */}
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "flex-start",
          marginBottom: "6px",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
          <span
            style={{
              padding: "2px 6px",
              borderRadius: "4px",
              backgroundColor: isCritical && !isAcknowledged ? "#DC2626" : "rgba(255,255,255,0.1)",
              color: isCritical && !isAcknowledged ? "#FFF" : color,
              fontSize: "0.65rem",
              fontWeight: "bold",
              letterSpacing: "0.05em",
            }}
          >
            {alert.severity}
          </span>
          <span style={{ fontSize: "0.75rem", fontWeight: "bold", color: "#F8FAFC" }}>
            {alert.anomalyType.replace("_", " ")}
          </span>
        </div>
        <span style={{ fontSize: "0.65rem", color: "#94A3B8" }}>
          {formatZuluTime(alert.detectedAt)}
        </span>
      </div>

      {/* Explanation Text */}
      <div style={{ fontSize: "0.75rem", color: "#CBD5E1", lineHeight: "1.4", margin: "8px 0", display: "-webkit-box", WebkitLineClamp: 2, WebkitBoxOrient: "vertical", overflow: "hidden" }}>
        {explanation}
      </div>

      {/* Track Link & Score Bar */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          margin: "8px 0",
          gap: "12px"
        }}
      >
        <button
          data-testid={`track-link-${alert.trackId}`}
          onClick={handleTrackClick}
          style={{
            background: "none",
            border: "none",
            color: "#60A5FA",
            fontSize: "0.75rem",
            cursor: "pointer",
            padding: 0,
            fontFamily: "monospace",
          }}
        >
          Track: {alert.trackId.slice(0,8)}...
        </button>

        {/* Score Bar */}
        <div style={{ flex: 1, display: "flex", alignItems: "center", gap: "6px" }}>
          <div style={{ flex: 1, height: "4px", backgroundColor: "rgba(255,255,255,0.1)", borderRadius: "2px", overflow: "hidden" }}>
             <div style={{
               height: "100%",
               width: `${alert.confidenceScore * 100}%`,
               backgroundColor: color
             }} />
          </div>
          <span style={{ fontSize: "0.65rem", color: "#94A3B8", minWidth: "30px", textAlign: "right" }}>
            {Math.round(alert.confidenceScore * 100)}%
          </span>
        </div>
      </div>

      {/* Status or Actions */}
      {feedbackStatus ? (
        <div style={{
          fontSize: "0.75rem",
          fontWeight: "bold",
          marginTop: "12px",
          padding: "6px 0",
          borderTop: "1px solid rgba(255,255,255,0.05)",
          color: feedbackStatus.includes("Error") ? "#F43F5E" : "#E2E8F0"
        }}>
          {feedbackStatus}
        </div>
      ) : (
        <div
          style={{
            display: "flex",
            gap: "6px",
            marginTop: "12px",
          }}
        >
          <button
            data-testid={`alert-inspect-${alert.alertId}`}
            onClick={handleInspect}
            style={{ ...actionButtonStyle, backgroundColor: "rgba(37, 99, 235, 0.2)", color: "#60A5FA", border: "1px solid rgba(59, 130, 246, 0.3)" }}
            title="Inspect Details (Enter)"
          >
            Inspect
          </button>

          <button
            data-testid={`alert-confirm-${alert.alertId}`}
            onClick={(e) => handleFeedback(e, "CONFIRM_ANOMALY")}
            style={{ ...actionButtonStyle, backgroundColor: "rgba(22, 163, 74, 0.2)", color: "#4ADE80", border: "1px solid rgba(34, 197, 94, 0.3)" }}
            title="Confirm Anomaly (C)"
          >
            Confirm
          </button>

          <button
            data-testid={`alert-reject-${alert.alertId}`}
            onClick={(e) => handleFeedback(e, "REJECT_ANOMALY")}
            style={{ ...actionButtonStyle, backgroundColor: "rgba(220, 38, 38, 0.2)", color: "#F87171", border: "1px solid rgba(239, 68, 68, 0.3)" }}
            title="Reject Anomaly (R)"
          >
            Reject
          </button>

          <button
            data-testid={`alert-assign-${alert.alertId}`}
            onClick={handleAssign}
            style={{ ...actionButtonStyle, backgroundColor: "rgba(202, 138, 4, 0.2)", color: "#FACC15", border: "1px solid rgba(234, 179, 8, 0.3)" }}
            title="Assign Alert"
          >
            Assign
          </button>
        </div>
      )}
    </div>
  );
});

const actionButtonStyle: React.CSSProperties = {
  flex: 1,
  padding: "6px 0",
  borderRadius: "4px",
  fontSize: "0.65rem",
  cursor: "pointer",
  fontWeight: "bold",
  transition: "all 0.15s ease",
};
