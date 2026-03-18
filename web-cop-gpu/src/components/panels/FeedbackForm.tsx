// CLASSIFICATION: UNCLASSIFIED
// src/components/panels/FeedbackForm.tsx
//
// Operator feedback submission form.  Shown as a modal overlay when
// feedbackOpen signal is true and a track is selected.
// Reference: docs/implementation/v4/phase3_ui_interaction.md §3 U3-5

import { For, Show, createSignal } from "solid-js";
import { submitFeedback, type FeedbackTypeOption } from "../../services/feedback";
import { operatorId } from "../../signals/auth";
import { trackDetail } from "../../signals/track";
import { feedbackOpen, setFeedbackOpen } from "../../signals/viewport";

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
    if (e) e.preventDefault();
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

  const handleQuickAction = (type: FeedbackTypeOption) => {
    setFeedbackType(type);
    if (justification().trim() === "") {
        setJustification(`Operator ${type} action triggered.`);
    }
  };

  return (
    <Show when={feedbackOpen()}>
      {/* Backdrop */}
      <div
        style={{
          position: "fixed",
          inset: "0",
          background: "rgba(15, 23, 42, 0.8)",
          "backdrop-filter": "blur(12px)",
          "-webkit-backdrop-filter": "blur(12px)",
          "z-index": "1000",
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
            background: "linear-gradient(135deg, rgba(30, 41, 59, 0.95) 0%, rgba(15, 23, 42, 0.98) 100%)",
            border: "1px solid rgba(59, 130, 246, 0.3)",
            "border-radius": "16px",
            padding: "1.75rem",
            width: "30rem",
            "max-width": "90vw",
            "box-shadow": "0 25px 50px -12px rgba(0, 0, 0, 0.7), 0 0 40px rgba(59, 130, 246, 0.15)",
            position: "relative",
            overflow: "hidden"
          }}
          onClick={(e) => e.stopPropagation()}
        >
          {/* Decorative Corner */}
          <div style={{ position: "absolute", top: 0, right: 0, width: "100px", height: "100px", background: "linear-gradient(225deg, rgba(59, 130, 246, 0.1) 0%, transparent 70%)", "pointer-events": "none" }}></div>

          <div style={{ display: "flex", "justify-content": "space-between", "align-items": "center", "margin-bottom": "1.5rem" }}>
            <h2
                style={{
                margin: 0,
                "font-size": "1rem",
                "font-weight": "900",
                color: "#60a5fa",
                "text-transform": "uppercase",
                "letter-spacing": "0.1em",
                }}
            >
                Tactical Feedback Loop
            </h2>
            <div style={{ "font-size": "0.6rem", color: "#94a3b8", background: "rgba(255,255,255,0.05)", padding: "2px 8px", "border-radius": "4px", "font-family": "monospace" }}>
                SID: {trackDetail()?.trackId.substring(0, 8) ?? "NONE"}
            </div>
          </div>

          <div style={{ display: "grid", "grid-template-columns": "1fr 1fr", gap: "0.75rem", "margin-bottom": "1.5rem" }}>
            <For each={FEEDBACK_OPTIONS}>
                {(opt: { value: FeedbackTypeOption; label: string }) => (
                    <button
                        onClick={() => handleQuickAction(opt.value)}
                        style={{
                            background: feedbackType() === opt.value ? "rgba(59, 130, 246, 0.2)" : "rgba(255,255,255,0.03)",
                            border: `1px solid ${feedbackType() === opt.value ? "#3b82f6" : "rgba(255,255,255,0.1)"}`,
                            color: feedbackType() === opt.value ? "#fff" : "#94a3b8",
                            padding: "0.75rem",
                            "border-radius": "8px",
                            "font-size": "0.75rem",
                            "font-weight": "700",
                            cursor: "pointer",
                            transition: "all 0.2s ease",
                            "text-align": "center",
                            "box-shadow": feedbackType() === opt.value ? "0 0 15px rgba(59, 130, 246, 0.3)" : "none"
                        }}
                    >
                        {opt.label}
                    </button>
                )}
            </For>
          </div>

          <form onSubmit={handleSubmit}>
            <label
              for="justification"
              style={{ display: "block", "font-size": "0.65rem", color: "#64748b", "margin-bottom": "0.5rem", "font-weight": "800", "text-transform": "uppercase" }}
            >
              Operational Justification
            </label>
            <textarea
              id="justification"
              value={justification()}
              onInput={(e) => setJustification(e.currentTarget.value)}
              rows={4}
              maxLength={500}
              placeholder="Enter context for this action..."
              style={{
                width: "100%",
                background: "rgba(15, 23, 42, 0.6)",
                color: "#e2e8f0",
                border: "1px solid rgba(59, 130, 246, 0.2)",
                "border-radius": "8px",
                padding: "0.75rem",
                "font-size": "0.85rem",
                resize: "none",
                "box-sizing": "border-box",
                "margin-bottom": "1.25rem",
                transition: "border-color 0.2s ease",
                outline: "none"
              }}
              onFocus={(e) => e.currentTarget.style.borderColor = "rgba(59, 130, 246, 0.5)"}
              onBlur={(e) => e.currentTarget.style.borderColor = "rgba(59, 130, 246, 0.2)"}
            />

            <Show when={error() !== null}>
              <div style={{ color: "#ef4444", "font-size": "0.75rem", "margin-bottom": "1rem", background: "rgba(239, 68, 68, 0.1)", padding: "0.75rem", "border-radius": "8px", border: "1px solid rgba(239, 68, 68, 0.2)" }} role="alert">
                {error()}
              </div>
            </Show>

            <Show when={success()}>
              <div style={{ color: "#10b981", "font-size": "0.75rem", "margin-bottom": "1rem", background: "rgba(16, 185, 129, 0.1)", padding: "0.75rem", "border-radius": "8px", border: "1px solid rgba(16, 185, 129, 0.2)" }}>
                Identity and action synchronized with Command.
              </div>
            </Show>

            <div style={{ display: "flex", gap: "1rem", "justify-content": "flex-end" }}>
              <button
                type="button"
                onClick={close}
                style={{
                  background: "transparent",
                  border: "1px solid rgba(255,255,255,0.1)",
                  color: "#94a3b8",
                  "border-radius": "8px",
                  padding: "0.75rem 1.5rem",
                  "font-size": "0.8rem",
                  "font-weight": "700",
                  cursor: "pointer",
                  transition: "all 0.2s ease"
                }}
              >
                DISMISS
              </button>
              <button
                type="submit"
                disabled={submitting() || !justification().trim()}
                style={{
                  background: "linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)",
                  border: "none",
                  color: "#fff",
                  "border-radius": "8px",
                  padding: "0.75rem 2rem",
                  "font-size": "0.8rem",
                  "font-weight": "900",
                  cursor: submitting() ? "wait" : "pointer",
                  opacity: !justification().trim() ? "0.5" : "1",
                  "box-shadow": "0 4px 15px rgba(59, 130, 246, 0.4)",
                  "text-transform": "uppercase",
                  "letter-spacing": "0.05em"
                }}
              >
                {submitting() ? "TRANSMITTING..." : "COMMIT ACTION"}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Show>
  );
}
