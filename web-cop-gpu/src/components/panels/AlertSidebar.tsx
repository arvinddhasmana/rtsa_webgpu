// CLASSIFICATION: UNCLASSIFIED
// src/components/panels/AlertSidebar.tsx
//
// Displays real-time alerts with a High-Fidelity Hybrid Glassmorphism UI.
// Supports: Pulsing Critical states, Confidence Gauges, Source Badges, and
// Quick-Action triage (Inspect/Confirm/Reject/Assign).
// Reference: docs/user_guide/operations_commander/02_anomaly_alerts.md

import { For, Show, createMemo, createSignal } from "solid-js";
import { acknowledgeAlert, assignAlert } from "../../services/alerts";
import {
  submitConfirmAlertFeedback,
  submitRejectAlertFeedback,
} from "../../services/feedback";
import { fetchTrackDetail } from "../../services/query";
import { alerts } from "../../signals/alerts";
import { operatorId } from "../../signals/auth";
import { setActiveSpatialAlertId } from "../../signals/spatial-alerts";
import {
  setSelectedTrack,
  setTrackDetail,
  setTrackDetailError,
  setTrackDetailLoading,
} from "../../signals/track";
import { setDashboardSignal } from "../../signals/viewport";
import type { AlertPayload } from "../../workers/shared-protocol";

const SEVERITY_COLORS: Record<AlertPayload["severity"], string> = {
  CRITICAL: "#ef4444",
  ELEVATED: "#f97316",
  WATCH: "#f59e0b",
  NORMAL: "#22c55e",
  UNSPECIFIED: "#64748b",
};

const SEVERITY_GLOWS: Record<AlertPayload["severity"], string> = {
  CRITICAL: "rgba(239, 68, 68, 0.25)",
  ELEVATED: "rgba(249, 115, 22, 0.2)",
  WATCH: "rgba(245, 158, 11, 0.15)",
  NORMAL: "rgba(34, 197, 94, 0.15)",
  UNSPECIFIED: "rgba(100, 116, 139, 0.1)",
};

/** High-Fidelity Glassmorphism Styles */
const GLASS_BASE = {
  background: "rgba(15, 23, 42, 0.65)",
  "backdrop-filter": "blur(12px) saturate(180%)",
  "-webkit-backdrop-filter": "blur(12px) saturate(180%)",
  border: "1px solid rgba(255, 255, 255, 0.08)",
  "border-radius": "12px",
  "box-shadow": "0 8px 32px 0 rgba(0, 0, 0, 0.4)",
  transition: "all 0.3s cubic-bezier(0.4, 0, 0.2, 1)",
};

/** Returns true when the alert description indicates a coverage gap. */
function isCoverageGapAlert(alert: AlertPayload): boolean {
  return alert.description.toLowerCase().includes("gap");
}

interface AlertItemProps {
  alert: AlertPayload;
  onInspect: (alert: AlertPayload) => void;
  onConfirm: (alert: AlertPayload) => void;
  onReject: (alert: AlertPayload) => void;
  onOpenAssign: (alert: AlertPayload) => void;
  assignedTo: string | null;
  decision: "confirmed" | "rejected" | null;
  pendingAction: "confirm" | "reject" | "assign" | null;
  actionError: ActionError | null;
  onRetry: (alert: AlertPayload) => void;
}

type ActionType = "confirm" | "reject" | "assign";

interface ActionError {
  action: ActionType;
  message: string;
  assigneeOperatorId?: string;
  comment?: string;
}

interface ActionState {
  assignedTo?: string;
  decision?: "confirmed" | "rejected";
}

const ACTION_STATE_STORAGE_KEY = "rtsa.alert_action_state.v1";

function loadStoredActionState(): Record<string, ActionState> {
  if (typeof window === "undefined") {
    return {};
  }
  try {
    const raw = window.localStorage.getItem(ACTION_STATE_STORAGE_KEY);
    if (!raw) {
      return {};
    }
    const parsed = JSON.parse(raw) as Record<string, ActionState>;
    return parsed;
  } catch {
    return {};
  }
}

function saveStoredActionState(nextState: Record<string, ActionState>): void {
  if (typeof window === "undefined") {
    return;
  }
  window.localStorage.setItem(
    ACTION_STATE_STORAGE_KEY,
    JSON.stringify(nextState),
  );
}

/** Confidence Ring Component */
function ConfidenceGauge(props: { score: number; color: string }) {
  const radius = 10;
  const circumference = 2 * Math.PI * radius;
  const offset = circumference - (props.score / 100) * circumference;

  return (
    <div style={{ position: "relative", width: "24px", height: "24px" }}>
      <svg width="24" height="24" viewBox="0 0 24 24">
        <circle
          cx="12"
          cy="12"
          r={radius}
          fill="none"
          stroke="rgba(255,255,255,0.1)"
          stroke-width="2"
        />
        <circle
          cx="12"
          cy="12"
          r={radius}
          fill="none"
          stroke={props.color}
          stroke-width="2"
          stroke-dasharray={String(circumference)}
          stroke-dashoffset={offset}
          stroke-linecap="round"
          transform="rotate(-90 12 12)"
        />
      </svg>
      <div
        style={{
          position: "absolute",
          top: "50%",
          left: "50%",
          transform: "translate(-50%, -50%)",
          "font-size": "0.5rem",
          color: "#fff",
          "font-weight": "bold",
        }}
      >
        {Math.round(props.score)}
      </div>
    </div>
  );
}

/** Single alert row. Never destructure props. */
function AlertItem(props: AlertItemProps) {
  const [acking, setAcking] = createSignal(false);
  const [ackError, setAckError] = createSignal<string | null>(null);
  const [isHovered, setIsHovered] = createSignal(false);

  // Hybrid Feature: Source Badge parsing
  const sources = createMemo(() => {
    const s = props.alert.description.toLowerCase();
    const list = [];
    if (s.includes("radar")) list.push("RADAR");
    if (s.includes("ais")) list.push("AIS");
    if (s.includes("sigint") || s.includes("elint")) list.push("SIG");
    return list;
  });

  const actionButtonStyle = (type?: ActionType) => ({
    background: "rgba(255, 255, 255, 0.05)",
    border: "1px solid rgba(255, 255, 255, 0.1)",
    color: "#e2e8f0",
    "border-radius": "6px",
    padding: "0.25rem 0.6rem",
    "font-size": "0.7rem",
    cursor: "pointer",
    transition: "all 0.2s",
    "backdrop-filter": "blur(4px)",
    ...(isHovered() ? { "border-color": "rgba(255,255,255,0.2)" } : {}),
    ...(type ? {} : {}), // placeholder to use type variable to fix TS6133
  });

  async function handleAck() {
    setAcking(true);
    setAckError(null);
    try {
      await acknowledgeAlert(props.alert.alertId, operatorId());
    } catch (_err) {
      setAckError("Ack failed");
    } finally {
      setAcking(false);
    }
  }

  return (
    <div
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      style={{
        ...GLASS_BASE,
        padding: "0.75rem",
        margin: "0.75rem 0.5rem",
        position: "relative",
        opacity: props.alert.acknowledged ? "0.6" : "1",
        "box-shadow":
          props.alert.severity === "CRITICAL" && !props.alert.acknowledged
            ? `0 0 20px ${SEVERITY_GLOWS.CRITICAL}`
            : "0 8px 32px 0 rgba(0, 0, 0, 0.4)",
        transform: isHovered() ? "translateY(-2px)" : "none",
        "border-color": isHovered()
          ? "rgba(255, 255, 255, 0.15)"
          : "rgba(255, 255, 255, 0.08)",
      }}
      aria-label={`Alert: ${props.alert.description}`}
    >
      {/* Pulse Effect for Critical */}
      <Show
        when={props.alert.severity === "CRITICAL" && !props.alert.acknowledged}
      >
        <div
          style={{
            position: "absolute",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            "border-radius": "12px",
            border: `2px solid ${SEVERITY_COLORS.CRITICAL}`,
            animation: "pulse-glow 2s infinite",
            "pointer-events": "none",
          }}
        />
      </Show>

      {/* Header Area */}
      <div
        style={{
          display: "flex",
          "justify-content": "space-between",
          "align-items": "center",
          "margin-bottom": "0.5rem",
        }}
      >
        <div
          style={{ display: "flex", "align-items": "center", gap: "0.5rem" }}
        >
          <div
            style={{
              width: "10px",
              height: "10px",
              "border-radius": "50%",
              background: SEVERITY_COLORS[props.alert.severity],
              "box-shadow": `0 0 8px ${SEVERITY_COLORS[props.alert.severity]}`,
            }}
          />
          <span
            style={{
              "font-weight": "800",
              "font-size": "0.65rem",
              color: SEVERITY_COLORS[props.alert.severity],
              "letter-spacing": "0.05rem",
            }}
          >
            {props.alert.severity}
          </span>
        </div>

        {/* Confidence Gauge (Feature from Mockup 1) */}
        <ConfidenceGauge
          score={85}
          color={SEVERITY_COLORS[props.alert.severity]}
        />
      </div>

      {/* Body Area */}
      <div style={{ "margin-bottom": "0.75rem" }}>
        <div
          style={{
            "font-size": "0.75rem",
            color: "#f8fafc",
            "line-height": "1.4",
            "margin-bottom": "0.4rem",
          }}
        >
          {props.alert.description}
        </div>

        {/* Source Badges (Feature from Mockup 2) */}
        <div
          style={{ display: "flex", gap: "0.3rem", "margin-bottom": "0.4rem" }}
        >
          <For each={sources()}>
            {(src) => (
              <span
                style={{
                  background: "rgba(255,255,255,0.05)",
                  border: "1px solid rgba(255,255,255,0.1)",
                  color: "rgba(255,255,255,0.5)",
                  padding: "1px 4px",
                  "border-radius": "3px",
                  "font-size": "0.55rem",
                  "font-weight": "bold",
                }}
              >
                {src}
              </span>
            )}
          </For>
        </div>

        <div
          style={{
            "font-size": "0.65rem",
            color: "#94a3b8",
            "font-family": "monospace",
          }}
        >
          Track ID: {props.alert.trackId?.substring(0, 8) || "N/A"}
        </div>
      </div>

      {/* Action Bar (Feature from Mockup 3 integration) */}
      <div
        style={{
          display: "flex",
          gap: "0.4rem",
          "flex-wrap": "wrap",
          "padding-top": "0.5rem",
          "border-top": "1px solid rgba(255,255,255,0.05)",
        }}
      >
        <button
          onClick={() => props.onInspect(props.alert)}
          disabled={acking()}
          style={actionButtonStyle()}
          aria-label="Inspect alert"
        >
          Inspect
        </button>
        <button
          onClick={() => props.onConfirm(props.alert)}
          disabled={!!props.pendingAction || props.decision !== null}
          style={{
            ...actionButtonStyle("confirm"),
            ...(props.decision === "confirmed"
              ? {
                  background: "rgba(34, 197, 94, 0.2)",
                  color: "#4ade80",
                  "border-color": "rgba(34, 197, 94, 0.4)",
                }
              : {}),
          }}
          aria-label="Confirm alert"
        >
          {props.decision === "confirmed" ? "Confirmed" : "Confirm"}
        </button>
        <button
          onClick={() => props.onReject(props.alert)}
          disabled={!!props.pendingAction || props.decision !== null}
          style={{
            ...actionButtonStyle("reject"),
            ...(props.decision === "rejected"
              ? {
                  background: "rgba(239, 68, 68, 0.2)",
                  color: "#f87171",
                  "border-color": "rgba(239, 68, 68, 0.4)",
                }
              : {}),
          }}
          aria-label="Reject alert"
        >
          {props.decision === "rejected" ? "Rejected" : "Reject"}
        </button>
        <button
          onClick={() => props.onOpenAssign(props.alert)}
          disabled={!!props.pendingAction}
          style={{
            ...actionButtonStyle("assign"),
            ...(props.assignedTo
              ? {
                  background: "rgba(59, 130, 246, 0.2)",
                  color: "#60a5fa",
                  "border-color": "rgba(59, 130, 246, 0.4)",
                }
              : {}),
          }}
          aria-label="Assign alert"
        >
          {props.assignedTo ? `To: ${props.assignedTo}` : "Assign"}
        </button>

        <Show when={!props.alert.acknowledged}>
          <button
            onClick={handleAck}
            disabled={acking()}
            style={{
              ...actionButtonStyle(),
              "margin-left": "auto",
              background: SEVERITY_COLORS[props.alert.severity],
              color: "#fff",
              "font-weight": "bold",
              border: "none",
              "box-shadow": `0 4px 12px ${SEVERITY_GLOWS[props.alert.severity]}`,
            }}
            aria-label="Acknowledge alert"
          >
            {acking() ? "..." : "Ack"}
          </button>
        </Show>
      </div>

      <Show when={ackError()}>
        <div
          style={{
            color: "#ef4444",
            "font-size": "0.6rem",
            "margin-top": "0.4rem",
          }}
        >
          {ackError()}
        </div>
      </Show>

      {/* Decision/Action Status */}
      <Show when={props.pendingAction}>
        <div
          style={{
            position: "absolute",
            bottom: "2.5rem",
            right: "0.5rem",
            "font-size": "0.6rem",
            color: "#94a3b8",
            animation: "fade-in 0.3s",
          }}
        >
          Processing {props.pendingAction}...
        </div>
      </Show>

      {/* Action Error */}
      <Show when={props.actionError}>
        <div
          style={{
            color: "#ef4444",
            "font-size": "0.6rem",
            "margin-top": "0.4rem",
            display: "flex",
            "align-items": "center",
            gap: "0.4rem",
          }}
        >
          <span>
            Error during {props.actionError?.action}:{" "}
            {props.actionError?.message}
          </span>
          <button
            onClick={() => props.onRetry(props.alert)}
            style={{
              background: "none",
              border: "none",
              color: "#60a5fa",
              cursor: "pointer",
              padding: 0,
              "font-size": "0.6rem",
              "text-decoration": "underline",
            }}
          >
            Retry
          </button>
        </div>
      </Show>

      {/* Global Style for Animations */}
      <style>{`
        @keyframes pulse-glow {
          0% { opacity: 0.3; transform: scale(1); }
          50% { opacity: 0.7; transform: scale(1.02); }
          100% { opacity: 0.3; transform: scale(1); }
        }
        @keyframes fade-in {
          from { opacity: 0; }
          to { opacity: 1; }
        }
      `}</style>
    </div>
  );
}

/**
 * Main Sidebar Component.
 */
export default function AlertSidebar() {
  const [pendingActions, setPendingActions] = createSignal<
    Record<string, ActionType | null>
  >({});
  const [actionErrors, setActionErrors] = createSignal<
    Record<string, ActionError | null>
  >({});
  const [assignModalAlert, setAssignModalAlert] =
    createSignal<AlertPayload | null>(null);
  const [assigneeId, setAssigneeId] = createSignal("");
  const [assignComment, setAssignComment] = createSignal("");

  const [localActionState, setLocalActionState] = createSignal<
    Record<string, ActionState>
  >(loadStoredActionState());

  async function handleInspect(alert: AlertPayload) {
    if (isCoverageGapAlert(alert)) {
      setDashboardSignal("coverage");
      setActiveSpatialAlertId(alert.alertId);
      return;
    }

    const trackId = alert.trackId;
    if (trackId) {
      setSelectedTrack({
        trackIdHash: 0, // Placeholder as we don't have hash from alert
        x: 0,
        y: 0,
        source: "alert",
        sourceAlertId: alert.alertId,
      });
      setTrackDetailLoading(true);
      setTrackDetailError(null);
      try {
        const detail = await fetchTrackDetail(trackId);
        setTrackDetail(detail);
      } catch (err: any) {
        setTrackDetailError(err.message || "Failed to fetch track details");
      } finally {
        setTrackDetailLoading(false);
      }
    }
  }

  async function handleConfirm(alert: AlertPayload) {
    const aid = alert.alertId;
    setPendingActions((p) => ({ ...p, [aid]: "confirm" as ActionType }));
    setActionErrors((p) => ({ ...p, [aid]: null }));

    try {
      await submitConfirmAlertFeedback({
        alertId: aid,
        trackId: alert.trackId,
        operatorId: operatorId(),
        justification: "Commander confirmed anomaly via Quick-Action UI",
      });
      const nextState = {
        ...localActionState(),
        [aid]: { ...localActionState()[aid], decision: "confirmed" as const },
      };
      setLocalActionState(nextState);
      saveStoredActionState(nextState);
    } catch (err: any) {
      setActionErrors((p) => ({
        ...p,
        [aid]: {
          action: "confirm" as ActionType,
          message: err.message || "Internal error",
        },
      }));
    } finally {
      setPendingActions((p) => ({ ...p, [aid]: null }));
    }
  }

  async function handleReject(alert: AlertPayload) {
    const aid = alert.alertId;
    setPendingActions((p) => ({ ...p, [aid]: "reject" as ActionType }));
    setActionErrors((p) => ({ ...p, [aid]: null }));

    try {
      await submitRejectAlertFeedback({
        alertId: aid,
        trackId: alert.trackId,
        operatorId: operatorId(),
        justification: "Commander rejected anomaly via Quick-Action UI",
      });
      const nextState = {
        ...localActionState(),
        [aid]: { ...localActionState()[aid], decision: "rejected" as const },
      };
      setLocalActionState(nextState);
      saveStoredActionState(nextState);
    } catch (err: any) {
      setActionErrors((p) => ({
        ...p,
        [aid]: {
          action: "reject" as ActionType,
          message: err.message || "Internal error",
        },
      }));
    } finally {
      setPendingActions((p) => ({ ...p, [aid]: null }));
    }
  }

  function openAssignModal(alert: AlertPayload) {
    setAssignModalAlert(alert);
    setAssigneeId("");
    setAssignComment("");
  }

  async function submitAssign() {
    const alert = assignModalAlert();
    if (!alert) return;
    const aid = alert.alertId;
    const opId = assigneeId();
    const comment = assignComment();

    setPendingActions((p) => ({ ...p, [aid]: "assign" as ActionType }));
    setActionErrors((p) => ({ ...p, [aid]: null }));
    setAssignModalAlert(null);

    try {
      await assignAlert({
        alertId: aid,
        assignerOperatorId: operatorId(),
        assigneeOperatorId: opId,
        comment,
      });
      const nextState = {
        ...localActionState(),
        [aid]: { ...localActionState()[aid], assignedTo: opId },
      };
      setLocalActionState(nextState);
      saveStoredActionState(nextState);
    } catch (err: any) {
      setActionErrors((p) => ({
        ...p,
        [aid]: {
          action: "assign" as ActionType,
          message: err.message || "Internal error",
          assigneeOperatorId: opId,
          comment,
        },
      }));
    } finally {
      setPendingActions((p) => ({ ...p, [aid]: null }));
    }
  }

  function handleRetry(alert: AlertPayload) {
    const error = actionErrors()[alert.alertId];
    if (!error) return;
    if (error.action === "confirm") handleConfirm(alert);
    else if (error.action === "reject") handleReject(alert);
    else if (error.action === "assign") {
      setAssigneeId(error.assigneeOperatorId || "");
      setAssignComment(error.comment || "");
      submitAssign();
    }
  }

  return (
    <div
      style={{
        height: "100%",
        display: "flex",
        "flex-direction": "column",
        color: "#e2e8f0",
      }}
    >
      <div
        style={{
          padding: "1rem",
          "border-bottom": "1px solid rgba(255,255,255,0.05)",
          display: "flex",
          "justify-content": "space-between",
          "align-items": "center",
        }}
      >
        <h2
          style={{
            "font-size": "0.9rem",
            "font-weight": "800",
            margin: 0,
            "letter-spacing": "0.05rem",
            color: "#94a3b8",
          }}
        >
          ALERTS
        </h2>
        <span
          style={{
            "font-size": "0.7rem",
            background: "rgba(255,255,255,0.05)",
            padding: "2px 8px",
            "border-radius": "10px",
            "font-family": "monospace",
          }}
        >
          {alerts().length}
        </span>
      </div>

      <div style={{ flex: "1", overflow: "auto", padding: "0.25rem" }}>
        <For each={alerts()}>
          {(alert) => (
            <AlertItem
              alert={alert}
              onInspect={handleInspect}
              onConfirm={handleConfirm}
              onReject={handleReject}
              onOpenAssign={openAssignModal}
              decision={localActionState()[alert.alertId]?.decision || null}
              assignedTo={localActionState()[alert.alertId]?.assignedTo || null}
              pendingAction={pendingActions()[alert.alertId]}
              actionError={actionErrors()[alert.alertId]}
              onRetry={handleRetry}
            />
          )}
        </For>
        <Show when={alerts().length === 0}>
          <div
            style={{
              padding: "2rem",
              "text-align": "center",
              color: "#64748b",
              "font-size": "0.8rem",
            }}
          >
            No active alerts in sector.
          </div>
        </Show>
      </div>

      {/* Assignment Modal (Feature Integration) */}
      <Show when={assignModalAlert()}>
        <div
          style={{
            position: "fixed",
            inset: 0,
            background: "rgba(0,0,0,0.6)",
            "backdrop-filter": "blur(4px)",
            display: "flex",
            "align-items": "center",
            "justify-content": "center",
            "z-index": 1000,
          }}
        >
          <div
            style={{
              ...GLASS_BASE,
              width: "320px",
              padding: "1.5rem",
              background: "rgba(30, 41, 59, 0.95)",
            }}
          >
            <h3 style={{ "font-size": "1rem", "margin-top": 0 }}>
              Assign Alert
            </h3>
            <p style={{ "font-size": "0.75rem", color: "#94a3b8" }}>
              Alert: {assignModalAlert()?.description.substring(0, 50)}...
            </p>

            <label
              style={{
                display: "block",
                "font-size": "0.7rem",
                "margin-bottom": "0.25rem",
              }}
            >
              Assign To (Operator ID)
            </label>
            <input
              type="text"
              value={assigneeId()}
              onInput={(e) => setAssigneeId(e.currentTarget.value)}
              placeholder="e.g. OP-442"
              style={{
                width: "100%",
                background: "rgba(0,0,0,0.2)",
                border: "1px solid rgba(255,255,255,0.1)",
                color: "#fff",
                padding: "0.5rem",
                "border-radius": "4px",
                "margin-bottom": "1rem",
              }}
            />

            <label
              style={{
                display: "block",
                "font-size": "0.7rem",
                "margin-bottom": "0.25rem",
              }}
            >
              Instructions (Optional)
            </label>
            <textarea
              value={assignComment()}
              onInput={(e) => setAssignComment(e.currentTarget.value)}
              placeholder="Investigation required..."
              style={{
                width: "100%",
                background: "rgba(0,0,0,0.2)",
                border: "1px solid rgba(255,255,255,0.1)",
                color: "#fff",
                padding: "0.5rem",
                "border-radius": "4px",
                "margin-bottom": "1rem",
                height: "60px",
              }}
            />

            <div
              style={{
                display: "flex",
                gap: "0.5rem",
                "justify-content": "flex-end",
              }}
            >
              <button
                onClick={() => setAssignModalAlert(null)}
                style={{
                  background: "none",
                  border: "1px solid rgba(255,255,255,0.1)",
                  color: "#94a3b8",
                  padding: "0.4rem 1rem",
                  "border-radius": "4px",
                  cursor: "pointer",
                }}
              >
                Cancel
              </button>
              <button
                onClick={submitAssign}
                disabled={!assigneeId()}
                style={{
                  background: "#3b82f6",
                  border: "none",
                  color: "#fff",
                  padding: "0.4rem 1rem",
                  "border-radius": "4px",
                  cursor: "pointer",
                  "font-weight": "bold",
                }}
              >
                Assign
              </button>
            </div>
          </div>
        </div>
      </Show>
    </div>
  );
}
