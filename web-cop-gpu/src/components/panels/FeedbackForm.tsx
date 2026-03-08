// CLASSIFICATION: UNCLASSIFIED
// src/components/panels/FeedbackForm.tsx
//
// Operator feedback submission form.  Shown as a modal overlay when
// feedbackOpen signal is true and a track is selected.
// Reference: docs/implementation/v4/phase3_ui_interaction.md §3 U3-5

import { Show, createSignal } from "solid-js";
import { feedbackOpen, setFeedbackOpen } from "../../signals/viewport";
import { trackDetail } from "../../signals/track";
import { submitFeedback, type FeedbackTypeOption } from "../../services/feedback";
import { operatorId } from "../../signals/auth";

const FEEDBACK_OPTIONS: { value: FeedbackTypeOption; label: string }[] = [
  { value: "CONFIRM_HOSTILE", label: "Confirm Hostile" },
  { value: "CONFIRM_FRIENDLY", label: "Confirm Friendly" },
  { value: "RECLASSIFY", label: "Reclassify" },
  { value: "REJECT_ANOMALY", label: "Reject Anomaly" },
];

/** Feedback submission modal. Never destructure props. */
export function FeedbackForm() {
  const [feedbackType, setFeedbackType] = createSignal<FeedbackTypeOption>("CONFIRM_HOSTILE");
  const [justification, setJustification] = createSignal("");
  const [submitting, setSubmitting] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);
  const [success, setSuccess] = createSignal(false);

  function close() {
    setFeedbackOpen(false);
    setJustification("");
    setError(null);
    setSuccess(false);
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    const detail = trackDetail();
    if (!detail) return;

    setSubmitting(true);
    setError(null);
    setSuccess(false);

    try {
      await submitFeedback({
        trackId: detail.trackId,
        operatorId: operatorId(),
        feedbackType: feedbackType(),
        justification: justification(),
      });
      setSuccess(true);
      setTimeout(close, 1500);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Submission failed");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Show when={feedbackOpen()}>
      {/* Backdrop */}
      <div
        style={{
          position: "fixed",
          inset: "0",
          background: "rgba(0,0,0,0.6)",
          "z-index": "500",
          display: "flex",
          "align-items": "center",
          "justify-content": "center",
        }}
        onClick={close}
        aria-modal="true"
        role="dialog"
        aria-label="Feedback form"
      >
        {/* Modal */}
        <div
          data-testid="feedback-form"
          style={{
            background: "#0d1424",
            border: "1px solid #2d3f56",
            "border-radius": "6px",
            padding: "1.25rem",
            width: "24rem",
            "max-width": "90vw",
          }}
          onClick={(e) => e.stopPropagation()}
        >
          <h2
            style={{
              margin: "0 0 0.75rem",
              "font-size": "0.85rem",
              color: "#f59e0b",
            }}
          >
            OPERATOR FEEDBACK
          </h2>

          <Show when={trackDetail()}>
            <div style={{ "font-size": "0.7rem", color: "#94a3b8", "margin-bottom": "0.75rem" }}>
              Track: {trackDetail()!.trackId}
            </div>
          </Show>

          <form onSubmit={handleSubmit}>
            {/* Feedback type */}
            <label
              for="feedback-type"
              style={{ display: "block", "font-size": "0.7rem", color: "#94a3b8", "margin-bottom": "0.25rem" }}
            >
              Feedback Type
            </label>
            <select
              id="feedback-type"
              value={feedbackType()}
              onChange={(e) => setFeedbackType(e.currentTarget.value as FeedbackTypeOption)}
              style={{
                width: "100%",
                background: "#1e2a3a",
                color: "#e2e8f0",
                border: "1px solid #2d3f56",
                "border-radius": "4px",
                padding: "0.35rem",
                "font-size": "0.8rem",
                "margin-bottom": "0.75rem",
              }}
            >
              {FEEDBACK_OPTIONS.map((opt) => (
                <option value={opt.value}>{opt.label}</option>
              ))}
            </select>

            {/* Justification */}
            <label
              for="justification"
              style={{ display: "block", "font-size": "0.7rem", color: "#94a3b8", "margin-bottom": "0.25rem" }}
            >
              Justification (required)
            </label>
            <textarea
              id="justification"
              value={justification()}
              onInput={(e) => setJustification(e.currentTarget.value)}
              rows={4}
              maxLength={500}
              placeholder="Provide justification for this feedback…"
              style={{
                width: "100%",
                background: "#1e2a3a",
                color: "#e2e8f0",
                border: "1px solid #2d3f56",
                "border-radius": "4px",
                padding: "0.35rem",
                "font-size": "0.8rem",
                resize: "vertical",
                "box-sizing": "border-box",
                "margin-bottom": "0.75rem",
              }}
            />

            <Show when={error() !== null}>
              <div style={{ color: "#ef4444", "font-size": "0.75rem", "margin-bottom": "0.5rem" }} role="alert">
                {error()}
              </div>
            </Show>

            <Show when={success()}>
              <div style={{ color: "#22c55e", "font-size": "0.75rem", "margin-bottom": "0.5rem" }}>
                Feedback submitted successfully.
              </div>
            </Show>

            <div style={{ display: "flex", gap: "0.5rem", "justify-content": "flex-end" }}>
              <button
                type="button"
                onClick={close}
                style={{
                  background: "none",
                  border: "1px solid #2d3f56",
                  color: "#94a3b8",
                  "border-radius": "4px",
                  padding: "0.35rem 0.75rem",
                  "font-size": "0.8rem",
                  cursor: "pointer",
                }}
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={submitting() || !justification().trim()}
                style={{
                  background: "#1e3a5f",
                  border: "1px solid #2d5a8f",
                  color: "#e2e8f0",
                  "border-radius": "4px",
                  padding: "0.35rem 0.75rem",
                  "font-size": "0.8rem",
                  cursor: submitting() ? "wait" : "pointer",
                  opacity: !justification().trim() ? "0.5" : "1",
                }}
              >
                {submitting() ? "Submitting…" : "Submit"}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Show>
  );
}
