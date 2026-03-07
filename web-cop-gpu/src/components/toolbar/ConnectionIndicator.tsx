// CLASSIFICATION: UNCLASSIFIED
// src/components/toolbar/ConnectionIndicator.tsx
//
// Visual indicator for WebTransport (hot path) and gRPC (cold path) connection state.
// Reference: docs/implementation/v4/phase3_ui_interaction.md §3 U3-2

import { wtConnected, grpcConnected, connecting } from "../../signals/connection";
import { latencyMs } from "../../signals/stats";

/** Colour-coded dot + label showing current connection health. */
export function ConnectionIndicator() {
  const dotColor = () => {
    if (connecting()) return "#f59e0b";
    if (wtConnected() && grpcConnected()) return "#22c55e";
    if (!wtConnected() && !grpcConnected()) return "#ef4444";
    return "#f59e0b"; // partial
  };

  const label = () => {
    if (connecting()) return "Connecting…";
    if (wtConnected() && grpcConnected()) return "Connected";
    if (!wtConnected() && !grpcConnected()) return "Disconnected";
    if (!wtConnected()) return "gRPC only";
    return "WT only";
  };

  const latency = () => {
    const ms = latencyMs();
    return ms >= 0 ? `${ms} ms` : "—";
  };

  return (
    <div
      data-testid="connection-indicator"
      style={{
        padding: "0.5rem",
        "border-top": "1px solid #1e2a3a",
        "margin-top": "auto",
      }}
    >
      <div style={{ "font-size": "0.65rem", color: "#94a3b8", "margin-bottom": "0.25rem" }}>
        CONNECTION
      </div>
      <div style={{ display: "flex", "align-items": "center", gap: "0.4rem" }}>
        <div
          style={{
            width: "8px",
            height: "8px",
            "border-radius": "50%",
            background: dotColor(),
            "flex-shrink": "0",
          }}
          aria-label={`Connection: ${label()}`}
        />
        <span style={{ "font-size": "0.75rem" }}>{label()}</span>
      </div>
      <div style={{ "font-size": "0.7rem", color: "#64748b", "margin-top": "0.2rem" }}>
        Latency: {latency()}
      </div>
    </div>
  );
}
