// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/ConnectivityTimeline.tsx — Connectivity event timeline
//
// Reference: docs/implementation/v5/sensordashboard_three_level_plan.md §B11

import { For, JSX, Show } from "solid-js";

export interface ConnectivityEventItem {
  timestamp: string;
  description: string;
  eventType: "NB" | "EY" | "R" | string;
}

export interface ConnectivityTimelineProps {
  events: ConnectivityEventItem[];
}

function eventColor(eventType: string): string {
  switch (eventType) {
    case "NB": return "#4ade80";
    case "EY": return "#f87171";
    case "R": return "#fbbf24";
    default: return "#64748b";
  }
}

function eventLabel(eventType: string): string {
  switch (eventType) {
    case "NB": return "Connected";
    case "EY": return "Interrupted";
    case "R": return "Reconnecting";
    default: return eventType;
  }
}

/** Vertical connectivity event list — B11 */
export function ConnectivityTimeline(props: ConnectivityTimelineProps): JSX.Element {
  return (
    <div data-testid="connectivity-timeline" style={{ display: "flex", "flex-direction": "column", gap: "0" }}>
      <Show
        when={props.events.length > 0}
        fallback={
          <div style={{ color: "#334155", "font-size": "0.68rem", "font-family": "monospace" }}>
            No events recorded
          </div>
        }
      >
        <For each={props.events}>
          {(evt, i) => {
            const color = eventColor(evt.eventType);
            return (
              <div
                data-testid={`timeline-event-${i()}`}
                style={{
                  display: "flex",
                  gap: "10px",
                  "align-items": "flex-start",
                  "padding-bottom": "10px",
                  position: "relative",
                }}
              >
                {/* Vertical connector line */}
                <Show when={i() < props.events.length - 1}>
                  <div style={{
                    position: "absolute",
                    left: "5px",
                    top: "14px",
                    width: "2px",
                    height: "calc(100% - 4px)",
                    background: "rgba(255,255,255,0.06)",
                  }} />
                </Show>

                {/* Dot */}
                <div style={{
                  width: "12px",
                  height: "12px",
                  "border-radius": "50%",
                  background: color,
                  "flex-shrink": 0,
                  "margin-top": "2px",
                  "box-shadow": `0 0 6px ${color}60`,
                  "z-index": 1,
                }} />

                {/* Text */}
                <div style={{ flex: 1, "min-width": 0 }}>
                  <div style={{
                    color: "#e2e8f0",
                    "font-size": "0.7rem",
                    "font-family": "monospace",
                    "font-weight": "500",
                  }}>
                    {evt.description}
                  </div>
                  <div style={{
                    display: "flex",
                    gap: "6px",
                    "margin-top": "2px",
                  }}>
                    <span style={{
                      color,
                      "font-size": "0.6rem",
                      "font-family": "monospace",
                      "text-transform": "uppercase",
                    }}>
                      {eventLabel(evt.eventType)}
                    </span>
                    <span style={{ color: "#334155", "font-size": "0.6rem", "font-family": "monospace" }}>
                      {evt.timestamp}
                    </span>
                  </div>
                </div>
              </div>
            );
          }}
        </For>
      </Show>
    </div>
  );
}
