// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/FeedbackForm.test.tsx

import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { FeedbackForm } from "../../components/detail/FeedbackForm";
import { useAuthStore } from "../../stores/authStore";

describe("FeedbackForm", () => {
  beforeEach(() => {
    useAuthStore.getState().setOperator({
      id: "op-001",
      name: "Test Operator",
      unit: "TEST",
      clearance: "PROTECTED_B",
      roles: ["OPERATOR"],
    });
  });

  it("renders feedback type selector and justification textarea", () => {
    render(<FeedbackForm trackId="TRK-001" />);
    expect(screen.getByTestId("feedback-type-select")).toBeTruthy();
    expect(screen.getByTestId("justification-input")).toBeTruthy();
    expect(screen.getByTestId("feedback-submit")).toBeTruthy();
  });

  it("T10: shows error when justification < 10 chars", () => {
    render(<FeedbackForm trackId="TRK-001" />);
    const textarea = screen.getByTestId("justification-input");
    fireEvent.change(textarea, { target: { value: "short" } });
    expect(screen.getByTestId("justification-error")).toBeTruthy();
  });

  it("T10: submit button disabled when justification too short", () => {
    render(<FeedbackForm trackId="TRK-001" />);
    const textarea = screen.getByTestId("justification-input");
    fireEvent.change(textarea, { target: { value: "short" } });
    const button = screen.getByTestId("feedback-submit");
    expect(button).toBeDisabled();
  });

  it("submit button enabled when justification >= 10 chars", () => {
    render(<FeedbackForm trackId="TRK-001" />);
    const textarea = screen.getByTestId("justification-input");
    fireEvent.change(textarea, { target: { value: "This is a valid justification" } });
    const button = screen.getByTestId("feedback-submit");
    expect(button).not.toBeDisabled();
  });

  it("T11: shows trust score on successful submission", async () => {
    render(<FeedbackForm trackId="TRK-001" />);
    const textarea = screen.getByTestId("justification-input");
    fireEvent.change(textarea, { target: { value: "This is a valid justification text" } });
    const button = screen.getByTestId("feedback-submit");
    fireEvent.click(button);
    await waitFor(() => {
      expect(screen.getByTestId("feedback-success")).toBeTruthy();
    });
  });

  it("no validation error shown when justification is empty (not yet touched)", () => {
    render(<FeedbackForm trackId="TRK-001" />);
    expect(screen.queryByTestId("justification-error")).toBeNull();
  });

  it("shows all feedback type options", () => {
    render(<FeedbackForm trackId="TRK-001" />);
    const select = screen.getByTestId("feedback-type-select");
    const options = select.querySelectorAll("option");
    expect(options.length).toBe(5);
  });
});
