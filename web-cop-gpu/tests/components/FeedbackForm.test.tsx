// CLASSIFICATION: UNCLASSIFIED
// tests/components/FeedbackForm.test.tsx

import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, fireEvent } from "@solidjs/testing-library";
import { FeedbackForm } from "../../src/components/panels/FeedbackForm";
import { setFeedbackOpen } from "../../src/signals/viewport";
import { setTrackDetail } from "../../src/signals/track";
import { setOperatorId } from "../../src/signals/auth";
import type { TrackDetail } from "../../src/signals/track";

// Mock the feedback service
const submitFeedbackMock = vi.fn().mockResolvedValue({
  feedbackId: "fb-1",
  trustScore: 0.8,
  validated: true,
});

vi.mock("../../src/services/feedback", () => ({
  submitFeedback: (...args: unknown[]) => submitFeedbackMock(...args),
}));

const mockTrack: TrackDetail = {
  trackId: "track-xyz-999",
  entityType: "Air",
  hostileClass: "Suspect",
  status: "Active",
  classification: "UNCLASSIFIED",
  confidenceScore: 0.75,
  sourceCount: 2,
  lat: 45.0,
  lon: -75.0,
  altitudeMeters: 5000,
  speedKnots: 300,
  headingDeg: 90,
  updatedAtMs: Date.now(),
};

afterEach(() => {
  setFeedbackOpen(false);
  setTrackDetail(null);
  setOperatorId("anonymous");
  submitFeedbackMock.mockClear();
});

describe("FeedbackForm", () => {
  it("renders nothing when feedbackOpen is false", () => {
    const { container } = render(() => <FeedbackForm />);
    expect(container.children.length).toBe(0);
  });

  it("renders modal when feedbackOpen is true", () => {
    setFeedbackOpen(true);
    render(() => <FeedbackForm />);
    expect(screen.getByRole("dialog")).toBeDefined();
    expect(screen.getByText("OPERATOR FEEDBACK")).toBeDefined();
  });

  it("shows track ID when trackDetail is set", () => {
    setFeedbackOpen(true);
    setTrackDetail(mockTrack);
    render(() => <FeedbackForm />);
    expect(screen.getByText(/track-xyz-999/)).toBeDefined();
  });

  it("shows feedback type selector", () => {
    setFeedbackOpen(true);
    render(() => <FeedbackForm />);
    expect(screen.getByLabelText("Feedback Type")).toBeDefined();
  });

  it("shows justification textarea", () => {
    setFeedbackOpen(true);
    render(() => <FeedbackForm />);
    expect(screen.getByLabelText("Justification (required)")).toBeDefined();
  });

  it("Submit button is disabled when justification is empty", () => {
    setFeedbackOpen(true);
    render(() => <FeedbackForm />);
    const submitBtn = screen.getByRole("button", { name: "Submit" });
    expect((submitBtn as HTMLButtonElement).disabled).toBe(true);
  });

  it("Submit button is enabled when justification is filled", () => {
    setFeedbackOpen(true);
    render(() => <FeedbackForm />);
    const textarea = screen.getByLabelText("Justification (required)") as HTMLTextAreaElement;
    fireEvent.input(textarea, { target: { value: "Confirmed by radar analysis" } });
    const submitBtn = screen.getByRole("button", { name: "Submit" });
    expect((submitBtn as HTMLButtonElement).disabled).toBe(false);
  });

  it("has data-testid feedback-form on modal container", () => {
    setFeedbackOpen(true);
    render(() => <FeedbackForm />);
    expect(screen.getByTestId("feedback-form")).toBeDefined();
  });

  it("submits with operatorId from auth signal, not hardcoded string", async () => {
    setOperatorId("op-sentinel-1");
    setFeedbackOpen(true);
    setTrackDetail(mockTrack);
    render(() => <FeedbackForm />);

    const textarea = screen.getByLabelText("Justification (required)") as HTMLTextAreaElement;
    fireEvent.input(textarea, { target: { value: "Radar confirmed hostile intent" } });
    const submitBtn = screen.getByRole("button", { name: "Submit" });
    fireEvent.click(submitBtn);

    // Wait for async submission
    await new Promise((r) => setTimeout(r, 50));

    expect(submitFeedbackMock).toHaveBeenCalledTimes(1);
    const callArgs = submitFeedbackMock.mock.calls[0][0] as { operatorId: string };
    expect(callArgs.operatorId).toBe("op-sentinel-1");
    expect(callArgs.operatorId).not.toBe("operator");
  });

  it("falls back to anonymous operatorId when no token has been acquired", async () => {
    // Default signal value is "anonymous"
    setFeedbackOpen(true);
    setTrackDetail(mockTrack);
    render(() => <FeedbackForm />);

    const textarea = screen.getByLabelText("Justification (required)") as HTMLTextAreaElement;
    fireEvent.input(textarea, { target: { value: "Fallback operator test" } });
    fireEvent.click(screen.getByRole("button", { name: "Submit" }));

    await new Promise((r) => setTimeout(r, 50));

    const callArgs = submitFeedbackMock.mock.calls[0][0] as { operatorId: string };
    expect(callArgs.operatorId).toBe("anonymous");
    expect(callArgs.operatorId).not.toBe("operator");
  });
});
