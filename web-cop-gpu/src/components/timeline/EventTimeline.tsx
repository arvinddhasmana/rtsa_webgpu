// CLASSIFICATION: UNCLASSIFIED
// src/components/timeline/EventTimeline.tsx
//
// Horizontal timeline of historical events for the currently selected track.
// Data fetched via gRPC QueryService.GetEventTimeline (ClickHouse cold path).
// Reference: docs/implementation/v4/phase3_ui_interaction.md §3 U3-7

import { Show, For, createResource } from "solid-js";
import { trackDetail } from "../../signals/track";
import { fetchTimeline } from "../../services/query";

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
        padding: "0.5rem 0.75rem",
        "min-height": "5rem",
      }}
      aria-label="Event timeline"
    >
      <div style={{ "font-size": "0.65rem", color: "#94a3b8", "margin-bottom": "0.4rem" }}>
        EVENT TIMELINE
      </div>

      <Show when={!trackId()}>
        <div style={{ color: "#64748b", "font-size": "0.75rem" }}>
          Select a track to view its event history.
        </div>
      </Show>

      <Show when={timeline.loading}>
        <div style={{ color: "#94a3b8", "font-size": "0.75rem" }}>Loading timeline…</div>
      </Show>

      <Show when={timeline.error}>
        <div style={{ color: "#ef4444", "font-size": "0.75rem" }} role="alert">
          Failed to load timeline
        </div>
      </Show>

      <Show when={!timeline.loading && !timeline.error && trackId()}>
        <div style={{ display: "flex", gap: "0.5rem", "overflow-x": "auto", "padding-bottom": "0.25rem" }}>
          <For
            each={timeline()?.events ?? []}
            fallback={
              <span style={{ color: "#64748b", "font-size": "0.75rem" }}>No events</span>
            }
          >
            {(event) => (
              <div
                style={{
                  "flex-shrink": "0",
                  background: "#1e2a3a",
                  "border-radius": "4px",
                  padding: "0.3rem 0.5rem",
                  "font-size": "0.65rem",
                  "white-space": "nowrap",
                }}
                title={event.summary}
              >
                <div style={{ color: "#f59e0b", "margin-bottom": "0.1rem" }}>
                  {EVENT_TYPE_LABELS[event.eventType] ?? "Event"}
                </div>
                <div style={{ color: "#94a3b8" }}>
                  {event.eventTime
                    ? new Date(Number(event.eventTime.seconds) * 1000).toLocaleTimeString()
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
