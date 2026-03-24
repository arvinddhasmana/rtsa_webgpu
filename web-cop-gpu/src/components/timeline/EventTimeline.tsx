// CLASSIFICATION: UNCLASSIFIED
// src/components/timeline/EventTimeline.tsx
//
// Horizontal timeline of historical events for the currently selected track.
// Data fetched via gRPC QueryService.GetEventTimeline (ClickHouse cold path).
// Reference: docs/implementation/v4/phase3_ui_interaction.md §3 U3-7

import { For, Show, createResource } from "solid-js";
import { fetchTimeline } from "../../services/query";
import { trackDetail } from "../../signals/track";

const EVENT_TYPE_LABELS: Record<number, string> = {
  0: "Unknown",
  1: "Created",
  2: "Updated",
  3: "Merged",
  4: "Dropped",
  5: "Anomaly",
  6: "Ack",
  7: "Feedback",
  8: "Class Change",
};

/** Horizontal event timeline strip for the active track. */
export function EventTimeline() {
  const trackId = () => trackDetail()?.trackId;

  const [timeline] = createResource(trackId, (id) =>
    fetchTimeline(id, 50),
  );

  return (
    <div
      style={{
        display: "flex",
        "flex-direction": "column",
        padding: "0",
        "min-height": "1.5rem",
      }}
      aria-label="Event timeline"
    >
      <div style={{ "font-size": "0.6rem", color: "#64748b", "margin-bottom": "0.1rem", "font-weight": "800", "text-transform": "uppercase", "letter-spacing": "0.06em" }}>
        Event Intelligence Stream
      </div>

      <Show when={!trackId()}>
        <div style={{ color: "#64748b", "font-size": "0.75rem", opacity: 0.6 }}>
          Select a track to view its event history.
        </div>
      </Show>

      <Show when={timeline.loading}>
        <div style={{ color: "#94a3b8", "font-size": "0.75rem", opacity: 0.6 }}>Synchronizing timeline...</div>
      </Show>

      <Show when={timeline.error}>
        <div style={{ color: "#ef4444", "font-size": "0.75rem", background: "rgba(239, 68, 68, 0.1)", padding: "0.5rem", "border-radius": "4px" }} role="alert">
          Failed to load timeline
        </div>
      </Show>

      <Show when={!timeline.loading && !timeline.error && trackId()}>
        <div style={{
          display: "flex",
          gap: "1rem",
          "overflow-x": "auto",
          "padding": "0.1rem 0.25rem 0.25rem",
          position: "relative",
          "scrollbar-width": "none"
        }}>
          {/* Connecting line */}
          <div style={{
            position: "absolute",
            top: "6px",
            left: "0",
            right: "0",
            height: "1px",
            background: "linear-gradient(90deg, rgba(59, 130, 246, 0.5) 0%, rgba(59, 130, 246, 0) 100%)",
            "z-index": 0
          }}></div>

          <For
            each={timeline() ?? []}
            fallback={
              <span style={{ color: "#64748b", "font-size": "0.75rem", opacity: 0.6 }}>No historical events indexed.</span>
            }
          >
            {(event) => (
              <div
                style={{
                  "flex-shrink": "0",
                  display: "flex",
                  "flex-direction": "column",
                  "align-items": "center",
                  position: "relative",
                  "z-index": 1,
                  "min-width": "60px"
                }}
                title={event.summary}
              >
                <div style={{
                  width: "8px",
                  height: "8px",
                  background: "#0f172a",
                  border: "1px solid #3b82f6",
                  "border-radius": "50%",
                  "margin-bottom": "0.2rem",
                  display: "flex",
                  "align-items": "center",
                  "justify-content": "center",
                  "box-shadow": "0 0 6px rgba(59, 130, 246, 0.4)"
                }}>
                  <div style={{ width: "3px", height: "3px", background: "#3b82f6", "border-radius": "50%" }}></div>
                </div>

                <div style={{
                  "font-size": "0.6rem",
                  "font-weight": "800",
                  color: "#e2e8f0",
                  "text-transform": "uppercase",
                  "margin-bottom": "0.2rem",
                  "white-space": "nowrap"
                }}>
                  {EVENT_TYPE_LABELS[event.eventType] ?? "Event"}
                </div>

                <div style={{
                  "font-size": "0.55rem",
                  color: "#64748b",
                  "font-family": "monospace"
                }}>
                  {event.eventTime
                    ? new Date(Number(event.eventTime.seconds) * 1000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
                    : "—"}
                </div>
              </div>
            )}
          </For>
        </div>
      </Show>
    </div>
  );
}
