// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/FeedbackLogTable.test.tsx

import { render, screen, fireEvent } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { FeedbackLogTable, FeedbackLogEntry } from "../../components/audit/FeedbackLogTable";

const makeEntry = (id: string, overrides: Partial<FeedbackLogEntry> = {}): FeedbackLogEntry => ({
  feedbackId: id,
  operatorId: "op-alpha",
  trackId: "TRK-0001",
  feedbackType: "CONFIRM_HOSTILE",
  trustScore: 0.85,
  accepted: true,
  submittedAt: new Date(),
  ...overrides,
});

describe("FeedbackLogTable", () => {
  it("renders the table", () => {
    render(<FeedbackLogTable entries={[]} />);
    expect(screen.getByTestId("feedback-log-table")).toBeTruthy();
  });

  it("shows empty state message when no entries", () => {
    render(<FeedbackLogTable entries={[]} />);
    expect(screen.getByText("No feedback entries found")).toBeTruthy();
  });

  it("renders a row per entry", () => {
    const entries = [
      makeEntry("fb-001"),
      makeEntry("fb-002", { feedbackType: "RECLASSIFY" }),
    ];
    render(<FeedbackLogTable entries={entries} />);
    expect(screen.getByTestId("feedback-row-fb-001")).toBeTruthy();
    expect(screen.getByTestId("feedback-row-fb-002")).toBeTruthy();
  });

  it("filters entries by type", () => {
    const entries = [
      makeEntry("fb-001", { feedbackType: "CONFIRM_HOSTILE" }),
      makeEntry("fb-002", { feedbackType: "RECLASSIFY" }),
    ];
    render(<FeedbackLogTable entries={entries} />);
    fireEvent.change(screen.getByTestId("feedback-filter-type"), {
      target: { value: "RECLASSIFY" },
    });
    expect(screen.queryByTestId("feedback-row-fb-001")).toBeNull();
    expect(screen.getByTestId("feedback-row-fb-002")).toBeTruthy();
  });

  it("shows accepted entries with check mark", () => {
    render(<FeedbackLogTable entries={[makeEntry("fb-001", { accepted: true })]} />);
    expect(screen.getByText(/✓ ACCEPTED/)).toBeTruthy();
  });

  it("shows rejected entries with cross mark", () => {
    render(
      <FeedbackLogTable entries={[makeEntry("fb-001", { accepted: false })]} />
    );
    expect(screen.getByText(/✗ REJECTED/)).toBeTruthy();
  });
});
