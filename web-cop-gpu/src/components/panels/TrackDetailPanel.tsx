// CLASSIFICATION: UNCLASSIFIED
// src/components/panels/TrackDetailPanel.tsx
//
// Displays full track information when the operator clicks a track on the canvas.
// Data flow: pick buffer → Render Worker → Main Thread → selectedTrack signal
//            → fetchTrackDetail (gRPC QueryService) → trackDetail signal → this panel.
// Reference: docs/implementation/v4/phase3_ui_interaction.md §3 U3-3

import { Show } from "solid-js";
import {
  selectedTrack,
  trackDetail,
  trackDetailLoading,
  trackDetailError,
  clearSelectedTrack,
} from "../../signals/track";
import { setFeedbackOpen } from "../../signals/viewport";

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
  return (
    <Show when={selectedTrack() !== null}>
      <div
        style={{
          padding: "0.75rem",
          "border-bottom": "1px solid #1e2a3a",
        }}
        aria-label="Track detail panel"
      >
        {/* Header */}
        <div
          style={{
            display: "flex",
            "justify-content": "space-between",
            "align-items": "center",
            "margin-bottom": "0.6rem",
          }}
        >
          <span style={{ "font-size": "0.75rem", "font-weight": "bold", color: "#f59e0b" }}>
            TRACK DETAIL
          </span>
          <button
            onClick={clearSelectedTrack}
            style={{
              background: "none",
              border: "none",
              color: "#94a3b8",
              cursor: "pointer",
              "font-size": "1rem",
              "line-height": "1",
            }}
            aria-label="Close track detail panel"
          >
            ×
          </button>
        </div>

        {/* Loading state */}
        <Show when={trackDetailLoading()}>
          <div style={{ color: "#94a3b8", "font-size": "0.8rem" }}>Loading…</div>
        </Show>

        {/* Error state */}
        <Show when={trackDetailError() !== null}>
          <div
            style={{ color: "#ef4444", "font-size": "0.8rem" }}
            role="alert"
          >
            {trackDetailError()}
          </div>
        </Show>

        {/* Track data */}
        <Show when={trackDetail() !== null}>
          <Field label="Track ID" value={trackDetail()!.trackId} />
          <Field label="Type" value={trackDetail()!.entityType} />
          <Field label="Classification" value={trackDetail()!.classification} />
          <Field label="IFF" value={trackDetail()!.hostileClass} />
          <Field label="Status" value={trackDetail()!.status} />
          <Field
            label="Confidence"
            value={`${(trackDetail()!.confidenceScore * 100).toFixed(1)}%`}
          />
          <Field label="Sources" value={trackDetail()!.sourceCount} />
          <Field
            label="Position"
            value={`${trackDetail()!.lat.toFixed(4)}°, ${trackDetail()!.lon.toFixed(4)}°`}
          />
          <Field label="Altitude" value={`${trackDetail()!.altitudeMeters.toFixed(0)} m`} />
          <Field label="Speed" value={`${trackDetail()!.speedKnots.toFixed(1)} kts`} />
          <Field label="Heading" value={`${trackDetail()!.headingDeg.toFixed(1)}°`} />
          <Show when={trackDetail()!.label}>
            <Field label="Label" value={trackDetail()!.label} />
          </Show>

          {/* Feedback button */}
          <button
            onClick={() => setFeedbackOpen(true)}
            style={{
              "margin-top": "0.5rem",
              width: "100%",
              background: "#1e3a5f",
              color: "#e2e8f0",
              border: "1px solid #2d5a8f",
              "border-radius": "4px",
              padding: "0.35rem",
              "font-size": "0.75rem",
              cursor: "pointer",
            }}
          >
            Submit Feedback
          </button>
        </Show>

        {/* No detail available yet (pick hash without a full track in DB) */}
        <Show when={!trackDetailLoading() && !trackDetailError() && trackDetail() === null}>
          <Field
            label="Track Hash"
            value={`0x${selectedTrack()!.trackIdHash.toString(16).padStart(8, "0")}`}
          />
          <div style={{ color: "#94a3b8", "font-size": "0.75rem" }}>
            Full detail not available (track not yet indexed).
          </div>
        </Show>
      </div>
    </Show>
  );
}
