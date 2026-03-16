// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/FusionConflictPanel.tsx

import { Component, Show } from "solid-js";
import { selectedTrackConflict } from "../../signals/track-selection";

export const FusionConflictPanel: Component = () => {
  return (
    <Show when={selectedTrackConflict()}>
      {(conflict) => (
        <div
          data-testid="fusion-conflict-panel"
          style={{
            background: "rgba(127, 29, 29, 0.15)",
            border: "1px solid #7f1d1d",
            "border-radius": "8px",
            padding: "1rem",
            color: "#fecaca",
            display: "flex",
            "flex-direction": "column",
            gap: "0.75rem",
            "backdrop-filter": "blur(4px)",
          }}
        >
          <header style={{ display: "flex", "align-items": "center", gap: "0.5rem" }}>
            <span style={{ color: "#ef4444" }}>⚠️</span>
            <span style={{ "font-size": "0.8rem", "font-weight": "700" }}>IDENTITY MISMATCH</span>
          </header>

          <p style={{ "font-size": "0.75rem", opacity: "0.9" }}>
            Track {conflict().trackId} has conflicting sensor evidence.
          </p>

          <div style={{ display: "flex", "flex-direction": "column", gap: "0.5rem" }}>
            {conflict().details.map((detail) => (
              <div
                style={{
                  display: "flex",
                  "justify-content": "space-between",
                  padding: "0.4rem",
                  background: "rgba(0, 0, 0, 0.2)",
                  "border-radius": "4px",
                  "font-size": "0.7rem",
                }}
              >
                <span style={{ color: "#f87171" }}>{detail.source}</span>
                <span style={{ "font-weight": "600" }}>{detail.value}</span>
              </div>
            ))}
          </div>

          <div style={{ display: "flex", gap: "0.5rem", "margin-top": "0.5rem" }}>
            <button
              style={{
                flex: "1",
                padding: "0.5rem",
                background: "rgba(127, 29, 29, 0.4)",
                border: "1px solid #991b1b",
                color: "white",
                "font-size": "0.7rem",
                "font-weight": "600",
                cursor: "pointer",
                "border-radius": "4px",
              }}
              onClick={() => {}}
            >
              MANUAL OVERRIDE
            </button>
            <button
              style={{
                flex: "1",
                padding: "0.5rem",
                background: "transparent",
                border: "1px solid rgba(254, 202, 202, 0.3)",
                color: "#fecaca",
                "font-size": "0.7rem",
                cursor: "pointer",
                "border-radius": "4px",
              }}
              onClick={() => {}}
            >
              RE-CORRELATE
            </button>
          </div>
        </div>
      )}
    </Show>
  );
};
