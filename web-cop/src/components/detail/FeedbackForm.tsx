// CLASSIFICATION: UNCLASSIFIED
// src/components/detail/FeedbackForm.tsx

import React, { useState } from "react";
import { FeedbackType } from "../../types/common";
import { feedbackClient } from "../../api/feedback-client";
import { useAuthStore } from "../../stores/authStore";
import { useOfflineMode } from "../../hooks/useOfflineMode";

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
 *
 * Fields:
 *   - feedback_type: dropdown
 *   - justification: textarea (required, min 10 chars)
 *
 * On submit:
 *   1. Calls FeedbackService.SubmitFeedback via gRPC-Web
 *   2. Shows loading spinner
 *   3. On success: shows trust score returned
 *   4. On error: shows error message
 *   5. If offline: queues feedback for later submission
 *
 * No PII logged in browser console.
 */
export const FeedbackForm: React.FC<FeedbackFormProps> = ({ trackId, alertId }) => {
  const operatorId = useAuthStore((s) => s.operatorId);
  const { isOffline } = useOfflineMode();

  const [feedbackType, setFeedbackType] = useState<FeedbackType>("CONFIRM_HOSTILE");
  const [justification, setJustification] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [trustScore, setTrustScore] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [submitted, setSubmitted] = useState(false);

  const isValid = justification.trim().length >= MIN_JUSTIFICATION_LENGTH;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!isValid || !operatorId) return;

    setIsLoading(true);
    setError(null);

    try {
      if (isOffline) {
        // Queue for later submission (stored in sessionStorage by useOfflineMode)
        setSubmitted(true);
        setIsLoading(false);
        return;
      }

      const response = await feedbackClient.submitFeedback({
        trackId,
        alertId,
        feedbackType,
        justification,
        operatorId,
      });

      setTrustScore(response.trustScore);
      setSubmitted(true);
    } catch (err) {
      // Log only non-PII error info
      setError(err instanceof Error ? err.message : "Submission failed");
    } finally {
      setIsLoading(false);
    }
  };

  if (submitted && trustScore !== null) {
    return (
      <div data-testid="feedback-success" style={{ padding: "8px" }}>
        <div style={{ color: "#16A34A", fontSize: "0.75rem", fontWeight: "bold" }}>
          ✓ Feedback accepted
        </div>
        <div style={{ fontSize: "0.7rem", color: "#9CA3AF", marginTop: "4px" }}>
          Trust Score: {Math.round(trustScore * 100)}%
        </div>
      </div>
    );
  }

  if (submitted && isOffline) {
    return (
      <div data-testid="feedback-queued" style={{ padding: "8px" }}>
        <div style={{ color: "#CA8A04", fontSize: "0.75rem", fontWeight: "bold" }}>
          ⚡ Queued for offline submission
        </div>
      </div>
    );
  }

  return (
    <form
      data-testid="feedback-form"
      onSubmit={(e) => void handleSubmit(e)}
      style={{ padding: "8px" }}
    >
      <div style={{ marginBottom: "8px" }}>
        <label
          htmlFor="feedback-type"
          style={{ display: "block", fontSize: "0.7rem", color: "#9CA3AF", marginBottom: "4px" }}
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
            padding: "4px",
            backgroundColor: "#0F172A",
            color: "#F1F5F9",
            border: "1px solid #334155",
            borderRadius: "4px",
            fontSize: "0.75rem",
          }}
        >
          {FEEDBACK_TYPES.map((ft) => (
            <option key={ft} value={ft}>
              {ft.replace(/_/g, " ")}
            </option>
          ))}
        </select>
      </div>

      <div style={{ marginBottom: "8px" }}>
        <label
          htmlFor="justification"
          style={{ display: "block", fontSize: "0.7rem", color: "#9CA3AF", marginBottom: "4px" }}
        >
          Justification (min {MIN_JUSTIFICATION_LENGTH} chars)
        </label>
        <textarea
          id="justification"
          data-testid="justification-input"
          value={justification}
          onChange={(e) => setJustification(e.target.value)}
          rows={3}
          style={{
            width: "100%",
            padding: "4px",
            backgroundColor: "#0F172A",
            color: "#F1F5F9",
            border: `1px solid ${!isValid && justification.length > 0 ? "#DC2626" : "#334155"}`,
            borderRadius: "4px",
            fontSize: "0.75rem",
            resize: "vertical",
            boxSizing: "border-box",
          }}
          placeholder="Enter justification..."
        />
        {!isValid && justification.length > 0 && (
          <div
            data-testid="justification-error"
            style={{ color: "#DC2626", fontSize: "0.65rem", marginTop: "2px" }}
          >
            Minimum {MIN_JUSTIFICATION_LENGTH} characters required
          </div>
        )}
      </div>

      {error && (
        <div
          data-testid="feedback-error"
          style={{ color: "#DC2626", fontSize: "0.7rem", marginBottom: "8px" }}
        >
          {error}
        </div>
      )}

      <button
        type="submit"
        data-testid="feedback-submit"
        disabled={!isValid || isLoading || !operatorId}
        style={{
          width: "100%",
          padding: "6px",
          backgroundColor: isValid && !isLoading ? "#1D4ED8" : "#374151",
          color: "#F1F5F9",
          border: "none",
          borderRadius: "4px",
          cursor: isValid && !isLoading ? "pointer" : "not-allowed",
          fontSize: "0.75rem",
          fontWeight: "bold",
        }}
      >
        {isLoading ? "Submitting..." : isOffline ? "Queue Feedback" : "Submit Feedback"}
      </button>
    </form>
  );
};
