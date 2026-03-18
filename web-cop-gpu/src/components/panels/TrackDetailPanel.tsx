// CLASSIFICATION: UNCLASSIFIED
// src/components/panels/TrackDetailPanel.tsx
//
// Displays full track information when the operator clicks a track on the canvas.
// Data flow: pick buffer → Render Worker → Main Thread → selectedTrack signal
//            → fetchTrackDetail (gRPC QueryService) → trackDetail signal → this panel.
// Reference: docs/implementation/v4/phase3_ui_interaction.md §3 U3-3

import { For, Show } from "solid-js";
import {
    clearSelectedTrack,
    selectedTrack,
    trackDetail,
    trackDetailError,
    trackDetailLoading,
} from "../../signals/track";
import { setFeedbackOpen } from "../../signals/viewport";
import { EventTimeline } from "../timeline/EventTimeline";

const labelStyle = {
  "font-size": "0.65rem",
  "text-transform": "uppercase",
  "letter-spacing": "0.06em",
  color: "#94a3b8",
};

const valueStyle = {
  "font-size": "0.8rem",
  color: "#e2e8f0",
  "margin-bottom": "0.4rem",
};

/** High-Fidelity Glassmorphism Styles (Consistent with Sidebar) */
const GLASS_BASE = {
  background: "linear-gradient(135deg, rgba(30, 41, 59, 0.7) 0%, rgba(15, 23, 42, 0.8) 100%)",
  "backdrop-filter": "blur(20px) saturate(180%)",
  "-webkit-backdrop-filter": "blur(20px) saturate(180%)",
  border: "1px solid rgba(255, 255, 255, 0.12)",
  "border-radius": "16px",
  "box-shadow": "0 12px 40px 0 rgba(0, 0, 0, 0.6)",
  transition: "all 0.4s cubic-bezier(0.16, 1, 0.3, 1)",
};

function Field(props: { label: string; value: string | number | undefined }) {
  return (
    <div style={{ "margin-bottom": "0.4rem" }}>
      <div style={labelStyle}>{props.label}</div>
      <div style={valueStyle}>{props.value ?? "—"}</div>
    </div>
  );
}

/** Panel shown on the right sidebar when a track is selected. */
export function TrackDetailPanel() {
  const containerStyle = {
    ...GLASS_BASE,
    background: "rgba(15, 23, 42, 0.4)",
    margin: "0.75rem",
    padding: "1.25rem",
    border: "1px solid rgba(255, 255, 255, 0.05)",
  };

  return (
    <Show when={selectedTrack() !== null}>
      <div
        style={{
          height: "100%",
          overflow: "auto",
        }}
        aria-label="Track detail panel"
      >
        <div style={containerStyle}>
          {/* Header */}
          <div
            style={{
              display: "flex",
              "justify-content": "space-between",
              "align-items": "center",
              "margin-bottom": "1.25rem",
            }}
          >
            <span
              style={{
                "font-size": "0.85rem",
                "font-weight": "900",
                color: "#f59e0b",
                "letter-spacing": "0.08em",
              }}
            >
              SIGNAL INTELLIGENCE
            </span>
            <button
              onClick={clearSelectedTrack}
              style={{
                background: "none",
                border: "none",
                color: "#94a3b8",
                cursor: "pointer",
                "font-size": "1.5rem",
                "line-height": "1",
              }}
              aria-label="Close track detail panel"
            >
              ×
            </button>
          </div>

          <Show
            when={
              selectedTrack()?.source === "alert" &&
              selectedTrack()?.sourceAlertId
            }
          >
            <div style={{
              background: "rgba(245, 158, 11, 0.08)",
              border: "1px solid rgba(245, 158, 11, 0.15)",
              padding: "0.6rem 0.8rem",
              "border-radius": "10px",
              "margin-bottom": "1.25rem",
              display: "flex",
              "align-items": "center",
              gap: "0.5rem"
            }}>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#f59e0b" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>
              <div style={{ flex: 1 }}>
                <div style={{ "font-size": "0.6rem", color: "#f59e0b", "font-weight": "800", "text-transform": "uppercase" }}>Target Identified from Alert</div>
                <div style={{ "font-size": "0.75rem", color: "#e2e8f0" }}>{selectedTrack()!.sourceAlertId}</div>
              </div>
            </div>
          </Show>

          <Show when={trackDetailLoading()}>
            <div style={{ color: "#94a3b8", "font-size": "0.8rem", padding: "2rem", "text-align": "center", opacity: 0.6 }}>
              <div style={{ border: "2px solid #3b82f6", "border-top": "2px solid transparent", width: "20px", height: "20px", "border-radius": "50%", animation: "spin 1s linear infinite", margin: "0 auto 1rem" }}></div>
              Synchronizing Intelligence...
            </div>
          </Show>

          {/* Error state */}
          <Show when={trackDetailError() !== null}>
            <div style={{ color: "#ef4444", "font-size": "0.85rem", background: "rgba(239, 68, 68, 0.1)", padding: "1rem", "border-radius": "10px", border: "1px solid rgba(239, 68, 68, 0.2)" }} role="alert">
              {trackDetailError()}
            </div>
          </Show>

          {/* Track data */}
          <Show when={trackDetail() !== null}>
            <div style={{ display: "grid", "grid-template-columns": "1fr 1fr", gap: "1rem" }}>
              <Field label="Track ID" value={trackDetail()!.trackId.substring(0, 8)} />
              <Field label="Type" value={trackDetail()!.entityType} />
              <Field label="Classification" value={trackDetail()!.classification} />
              <Field label="IFF State" value={trackDetail()!.hostileClass} />
              <Field label="Status" value={trackDetail()!.status} />
              <div style={{
                background: "rgba(255,255,255,0.03)",
                padding: "0.5rem",
                "border-radius": "8px",
                border: "1px solid rgba(255,255,255,0.05)",
                "text-align": "center"
              }}>
                <div style={labelStyle}>Confidence</div>
                <div style={{ ...valueStyle, "font-size": "1rem", "font-weight": "900", color: "#60a5fa", margin: 0 }}>
                  {(trackDetail()!.confidenceScore * 100).toFixed(0)}%
                </div>
              </div>
            </div>

            <div style={{ "margin-top": "1.5rem", "padding-top": "1.25rem", "border-top": "1px solid rgba(255,255,255,0.1)" }}>
              <div style={{ display: "flex", "justify-content": "space-between", "align-items": "baseline", "margin-bottom": "1rem" }}>
                <h4 style={{ ...labelStyle, color: "#64748b", "font-weight": "800" }}>Live Kinematics</h4>
                <div style={{ "font-size": "0.6rem", color: "#3b82f6", background: "rgba(59, 130, 246, 0.1)", padding: "0.1rem 0.4rem", "border-radius": "4px" }}>REAL-TIME UPDATING</div>
              </div>
              <div style={{ display: "grid", "grid-template-columns": "1fr 1fr", gap: "1rem" }}>
                <Field
                  label="Position"
                  value={`${trackDetail()!.lat.toFixed(4)}, ${trackDetail()!.lon.toFixed(4)}`}
                />
                <Field
                  label="Altitude"
                  value={`${trackDetail()!.altitudeMeters.toFixed(0)}m`}
                />
                <Field
                  label="Speed (SOG)"
                  value={`${trackDetail()!.speedKnots.toFixed(0)}kts`}
                />
                <Field
                  label="Course (COG)"
                  value={`${trackDetail()!.headingDeg.toFixed(0)}°`}
                />
              </div>
            </div>

            {/* Source Pedigree Section */}
            <div style={{ "margin-top": "1.5rem", "padding-top": "1.25rem", "border-top": "1px solid rgba(255,255,255,0.1)" }}>
              <h4 style={{ ...labelStyle, "margin-bottom": "1rem", color: "#64748b", "font-weight": "800" }}>Source Pedigree</h4>
              <div style={{ display: "flex", "flex-direction": "column", gap: "0.75rem" }}>
                <For each={trackDetail()!.sourceContributions}>
                  {(source) => (
                    <div style={{
                      background: "rgba(255,255,255,0.02)",
                      border: "1px solid rgba(255,255,255,0.05)",
                      "border-radius": "8px",
                      padding: "0.6rem",
                      display: "flex",
                      "align-items": "center",
                      gap: "0.75rem"
                    }}>
                      <div style={{
                        width: "32px",
                        height: "32px",
                        background: "rgba(59, 130, 246, 0.1)",
                        "border-radius": "6px",
                        display: "flex",
                        "align-items": "center",
                        "justify-content": "center"
                      }}>
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#3b82f6" stroke-width="2"><circle cx="12" cy="12" r="3"></circle><path d="M12 2v2m0 16v2M2 12h2m16 0h2m-3.17-6.83l-1.42 1.42m-10.83 10.83l-1.42 1.42m0-13.67l1.42 1.42m10.83 10.83l1.42 1.42"></path></svg>
                      </div>
                      <div style={{ flex: 1 }}>
                        <div style={{ display: "flex", "justify-content": "space-between", "margin-bottom": "0.15rem" }}>
                          <span style={{ "font-size": "0.7rem", "font-weight": "700", color: "#e2e8f0" }}>{source.sourceName}</span>
                          <span style={{ "font-size": "0.6rem", color: "#94a3b8" }}>{new Date(source.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
                        </div>
                        <div style={{ "font-size": "0.65rem", color: "#64748b" }}>{source.data}</div>
                        <div style={{
                          height: "3px",
                          width: "100%",
                          background: "rgba(255,255,255,0.05)",
                          "border-radius": "2px",
                          "margin-top": "0.4rem",
                          overflow: "hidden"
                        }}>
                          <div style={{
                            height: "100%",
                            width: `${source.signalStrength * 100}%`,
                            background: source.signalStrength > 0.8 ? "#10b981" : "#3b82f6",
                            "box-shadow": "0 0 8px rgba(59, 130, 246, 0.5)"
                          }}></div>
                        </div>
                      </div>
                    </div>
                  )}
                </For>
              </div>
            </div>

            {/* Integrated Timeline */}
            <div style={{ "margin-top": "1.5rem", "padding-top": "1.25rem", "border-top": "1px solid rgba(255,255,255,0.1)" }}>
              <EventTimeline />
            </div>

            {/* Feedback button */}
            <button
              onClick={() => setFeedbackOpen(true)}
              style={{
                "margin-top": "2rem",
                width: "100%",
                background: "linear-gradient(135deg, rgba(59, 130, 246, 0.2) 0%, rgba(37, 99, 235, 0.1) 100%)",
                color: "#60a5fa",
                border: "1px solid rgba(59, 130, 246, 0.4)",
                "border-radius": "12px",
                padding: "0.8rem",
                "font-size": "0.8rem",
                "font-weight": "800",
                cursor: "pointer",
                transition: "all 0.3s cubic-bezier(0.16, 1, 0.3, 1)",
                "backdrop-filter": "blur(8px)",
                "text-transform": "uppercase",
                "letter-spacing": "0.04em",
                "box-shadow": "0 4px 20px rgba(0,0,0,0.3)"
              }}
            >
              Submit Feedback
            </button>
          </Show>

          {/* No detail fallback */}
          <Show
            when={
              !trackDetailLoading() &&
              !trackDetailError() &&
              trackDetail() === null
            }
          >
            <div style={{ padding: "1rem 0", "text-align": "center", opacity: 0.6 }}>
              <Field
                label="Track Hash Signature"
                value={`0x${selectedTrack()!.trackIdHash.toString(16).padStart(8, "0")}`}
              />
              <div style={{ color: "#94a3b8", "font-size": "0.75rem", "margin-top": "1rem", "font-style": "italic" }}>
                Intelligence signature detected. Awaiting metadata indexing...
              </div>
            </div>
          </Show>
        </div>
      </div>
      <style>{`
        @keyframes spin {
          to { transform: rotate(360deg); }
        }
      `}</style>
    </Show>
  );
}
