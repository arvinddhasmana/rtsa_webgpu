// CLASSIFICATION: UNCLASSIFIED
// src/components/panels/AlertSidebar.tsx
//
// Displays real-time alerts from the Data Worker / gRPC stream.
// Operators can acknowledge alerts via the gRPC AlertService.
// Reference: docs/implementation/v4/phase3_ui_interaction.md §3 U3-4

import { For, Show, createSignal } from "solid-js";
import { alerts } from "../../signals/alerts";
import { acknowledgeAlert } from "../../services/alerts";
import type { AlertPayload } from "../../workers/shared-protocol";

const SEVERITY_COLORS: Record<AlertPayload["severity"], string> = {
  CRITICAL: "#ef4444",
  ELEVATED: "#f97316",
  WATCH: "#f59e0b",
  NORMAL: "#22c55e",
  UNSPECIFIED: "#64748b",
};

interface AlertItemProps {
  alert: AlertPayload;
}

/** Single alert row. Never destructure props. */
function AlertItem(props: AlertItemProps) {
  const [acking, setAcking] = createSignal(false);
  const [ackError, setAckError] = createSignal<string | null>(null);

  async function handleAck() {
    setAcking(true);
    setAckError(null);
    try {
      await acknowledgeAlert(props.alert.alertId, "operator");
    } catch (_err) {
      setAckError("Ack failed");
    } finally {
      setAcking(false);
    }
  }

  return (
    <div
      style={{
        padding: "0.5rem",
        "border-bottom": "1px solid #1e2a3a",
        opacity: props.alert.acknowledged ? "0.5" : "1",
      }}
      aria-label={`Alert: ${props.alert.description}`}
    >
      <div style={{ display: "flex", "align-items": "flex-start", gap: "0.4rem" }}>
        {/* Severity dot */}
        <div
          style={{
            width: "8px",
            height: "8px",
            "border-radius": "50%",
            background: SEVERITY_COLORS[props.alert.severity],
            "flex-shrink": "0",
            "margin-top": "0.3rem",
          }}
          aria-label={props.alert.severity}
        />
        <div style={{ flex: "1", "min-width": "0" }}>
          <div
            style={{
              "font-size": "0.7rem",
              "font-weight": "bold",
              color: SEVERITY_COLORS[props.alert.severity],
            }}
          >
            {props.alert.severity}
          </div>
          <div style={{ "font-size": "0.75rem", "word-break": "break-word" }}>
            {props.alert.description}
          </div>
          <div style={{ "font-size": "0.65rem", color: "#64748b", "margin-top": "0.2rem" }}>
            Track: {props.alert.trackId.slice(0, 8)}…
          </div>
        </div>

        {/* Acknowledge button */}
        <Show when={!props.alert.acknowledged}>
          <button
            onClick={handleAck}
            disabled={acking()}
            style={{
              "flex-shrink": "0",
              background: "#1e2a3a",
              border: "1px solid #2d3f56",
              color: "#e2e8f0",
              "border-radius": "3px",
              padding: "0.2rem 0.4rem",
              "font-size": "0.65rem",
              cursor: acking() ? "wait" : "pointer",
            }}
            aria-label="Acknowledge alert"
          >
            Ack
          </button>
        </Show>
      </div>

      <Show when={ackError() !== null}>
        <div style={{ "font-size": "0.65rem", color: "#ef4444", "margin-top": "0.2rem" }} role="alert">
          {ackError()}
        </div>
      </Show>
    </div>
  );
}

/** Alert sidebar panel showing the live alert list. */
export function AlertSidebar() {
  const visibleAlerts = () => alerts();

  return (
    <div style={{ display: "flex", "flex-direction": "column", height: "100%" }}>
      {/* Header */}
      <div
        style={{
          padding: "0.5rem 0.75rem",
          "border-bottom": "1px solid #1e2a3a",
          display: "flex",
          "align-items": "center",
          "justify-content": "space-between",
          "flex-shrink": "0",
        }}
      >
        <span style={{ "font-size": "0.75rem", "font-weight": "bold", color: "#f59e0b" }}>
          ALERTS
        </span>
        <span
          style={{
            "font-size": "0.65rem",
            background: "#1e2a3a",
            padding: "0 0.4rem",
            "border-radius": "3px",
          }}
        >
          {visibleAlerts().length}
        </span>
      </div>

      {/* Alert list */}
      <div style={{ flex: "1", "overflow-y": "auto" }}>
        <Show
          when={visibleAlerts().length > 0}
          fallback={
            <div
              style={{
                padding: "1rem",
                color: "#64748b",
                "font-size": "0.75rem",
                "text-align": "center",
              }}
            >
              No active alerts
            </div>
          }
        >
          <For each={visibleAlerts()}>
            {(alert) => <AlertItem alert={alert} />}
          </For>
        </Show>
      </div>
    </div>
  );
}
