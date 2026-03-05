// CLASSIFICATION: UNCLASSIFIED
// src/components/detail/FeedbackForm.tsx

import React, { useState } from "react";
import { useFeedback } from "../../hooks/useFeedback";
import { useOfflineMode } from "../../hooks/useOfflineMode";
import { useAuthStore } from "../../stores/authStore";
import { FeedbackType } from "../../types/common";

interface FeedbackFormProps {
  trackId: string;
  alertId?: string;
}

const FEEDBACK_TYPES: FeedbackType[] = [
  "CONFIRM_HOSTILE",
  "CONFIRM_FRIENDLY",
  "RECLASSIFY",
  "REJECT_ANOMALY",
  "CONFIRM_ANOMALY",
];

const MIN_JUSTIFICATION_LENGTH = 10;

/**
 * FeedbackForm — allows operator to submit feedback on a track or alert.
 * Uses useFeedback hook for submission and state.
 */
export const FeedbackForm: React.FC<FeedbackFormProps> = ({ trackId, alertId }) => {
  const operatorId = useAuthStore((s) => s.operatorId);
  const { isOffline } = useOfflineMode();
  const { submit, status, trustScore, error } = useFeedback();

  const [feedbackType, setFeedbackType] = useState<FeedbackType>("CONFIRM_HOSTILE");
  const [justification, setJustification] = useState("");
  const [queued, setQueued] = useState(false);

  const isValid = justification.trim().length >= MIN_JUSTIFICATION_LENGTH;
  const isLoading = status === "submitting";

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!isValid || !operatorId) return;

    if (isOffline) {
      // Queue for later submission (stored in sessionStorage by useOfflineMode)
      setQueued(true);
      return;
    }

    await submit(trackId, alertId || "", feedbackType, justification);
  };

  if (status === "accepted" || status === "under_review" || status === "queued") {
    const isAccepted = status === "accepted";
    return (
      <div data-testid="feedback-success" style={{ padding: "16px", backgroundColor: "rgba(255,255,255,0.02)", borderRadius: "8px", border: "1px solid rgba(255,255,255,0.05)", marginTop: "8px" }}>
        <div style={{ color: isAccepted ? "#10B981" : "#F59E0B", fontSize: "0.85rem", fontWeight: "bold", display: "flex", alignItems: "center", gap: "8px" }}>
          {isAccepted ? "✅ Feedback Accepted" : "⏳ Feedback Under Review"}
        </div>
        {trustScore !== null && (
          <div style={{ fontSize: "0.75rem", color: "#94A3B8", marginTop: "8px", display: "flex", justifyContent: "space-between" }}>
            <span>Impact Score:</span>
            <span style={{ fontWeight: "bold", color: isAccepted ? "#10B981" : "#F59E0B" }}>{Math.round(trustScore * 100)}%</span>
          </div>
        )}
      </div>
    );
  }

  if (queued && isOffline) {
    return (
      <div data-testid="feedback-queued" style={{ padding: "16px", backgroundColor: "rgba(202, 138, 4, 0.1)", borderRadius: "8px", border: "1px solid #CA8A04", marginTop: "8px" }}>
        <div style={{ color: "#FBBF24", fontSize: "0.85rem", fontWeight: "bold" }}>
          ⚡ Queued for offline submission
        </div>
      </div>
    );
  }

  return (
    <form
      data-testid="feedback-form"
      onSubmit={(e) => void handleSubmit(e)}
      style={{ padding: "16px" }}
    >
      <div style={{ fontSize: "0.75rem", color: "#94A3B8", marginBottom: "16px", borderBottom: "1px solid rgba(255,255,255,0.05)", paddingBottom: "4px" }}>
        Submit Operator Feedback
      </div>

      <div style={{ marginBottom: "12px" }}>
        <label
          htmlFor="feedback-type"
          style={{ display: "block", fontSize: "0.7rem", color: "#94A3B8", marginBottom: "6px" }}
        >
          Feedback Type
        </label>
        <select
          id="feedback-type"
          data-testid="feedback-type-select"
          value={feedbackType}
          onChange={(e) => setFeedbackType(e.target.value as FeedbackType)}
          style={{
            width: "100%",
            padding: "8px",
            backgroundColor: "rgba(15, 23, 42, 0.6)",
            color: "#F1F5F9",
            border: "1px solid rgba(255,255,255,0.1)",
            borderRadius: "4px",
            fontSize: "0.75rem",
            outline: "none",
          }}
        >
          {FEEDBACK_TYPES.map((ft) => (
            <option key={ft} value={ft}>
              {ft.replace(/_/g, " ")}
            </option>
          ))}
        </select>
      </div>

      <div style={{ marginBottom: "16px" }}>
        <label
          htmlFor="justification"
          style={{ display: "flex", justifyContent: "space-between", fontSize: "0.7rem", color: "#94A3B8", marginBottom: "6px" }}
        >
          <span>Justification</span>
          <span style={{ color: isValid ? "#10B981" : "#64748B" }}>
            {justification.length}/{MIN_JUSTIFICATION_LENGTH} min chars
          </span>
        </label>
        <textarea
          id="justification"
          data-testid="justification-input"
          value={justification}
          onChange={(e) => setJustification(e.target.value)}
          rows={4}
          style={{
            width: "100%",
            padding: "8px",
            backgroundColor: "rgba(15, 23, 42, 0.6)",
            color: "#F1F5F9",
            border: `1px solid ${!isValid && justification.length > 0 ? "#DC2626" : "rgba(255,255,255,0.1)"}`,
            borderRadius: "4px",
            fontSize: "0.75rem",
            resize: "vertical",
            boxSizing: "border-box",
            outline: "none",
          }}
          placeholder="Enter detailed justification for your evaluation..."
        />
      </div>

      {(error || status === "rejected") && (
        <div
          data-testid="feedback-error"
          style={{
            color: "#FCA5A5",
            fontSize: "0.75rem",
            marginBottom: "12px",
            padding: "8px",
            backgroundColor: "rgba(220, 38, 38, 0.1)",
            borderRadius: "4px",
            border: "1px solid rgba(220, 38, 38, 0.3)"
          }}
        >
          {error?.message || "Submission rejected by trust engine. Please provide more detail."}
        </div>
      )}

      <button
        type="submit"
        data-testid="feedback-submit"
        disabled={!isValid || isLoading || !operatorId}
        style={{
          width: "100%",
          padding: "10px",
          backgroundColor: isValid && !isLoading ? "#2563EB" : "rgba(255,255,255,0.05)",
          color: isValid && !isLoading ? "#F1F5F9" : "#64748B",
          border: "none",
          borderRadius: "4px",
          cursor: isValid && !isLoading ? "pointer" : "not-allowed",
          fontSize: "0.8rem",
          fontWeight: "bold",
          transition: "all 0.2s",
        }}
      >
        {isLoading ? "Analyzing & Submitting..." : isOffline ? "Queue Offline" : "Submit to Trust Engine →"}
      </button>
    </form>
  );
};
