// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/AuditDashboard.test.tsx

import { render, screen, fireEvent } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { AuditDashboard } from "../../components/layout/AuditDashboard";

// Mock audit client
vi.mock("../../api/audit-client", () => ({
  auditClient: {
    queryAuditLog: vi.fn().mockResolvedValue({ events: [] }),
  },
}));

describe("AuditDashboard", () => {
  it("renders the audit dashboard", () => {
    render(<AuditDashboard />);
    expect(screen.getByTestId("audit-dashboard")).toBeTruthy();
  });

  it("shows SECURITY AUDIT heading", () => {
    render(<AuditDashboard />);
    expect(screen.getByText("🔐 SECURITY AUDIT")).toBeTruthy();
  });

  it("renders KPI pills (Submissions, Avg Trust, Accepted, Rejected)", () => {
    render(<AuditDashboard />);
    expect(screen.getByText("Submissions")).toBeTruthy();
    expect(screen.getByText("Avg Trust")).toBeTruthy();
    expect(screen.getByText("Accepted")).toBeTruthy();
    expect(screen.getByText("Rejected")).toBeTruthy();
  });

  it("renders Feedback Log tab by default", () => {
    render(<AuditDashboard />);
    expect(screen.getByTestId("feedback-log-table")).toBeTruthy();
  });

  it("switches to Audit Log tab", () => {
    render(<AuditDashboard />);
    const auditBtn = screen.getByText("Audit Log");
    fireEvent.click(auditBtn);
    expect(screen.getByTestId("audit-log")).toBeTruthy();
  });

  it("renders TrustScoreHistogram in feedback tab", () => {
    render(<AuditDashboard />);
    expect(screen.getByTestId("trust-score-histogram")).toBeTruthy();
  });
});
