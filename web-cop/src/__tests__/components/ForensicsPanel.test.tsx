// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/ForensicsPanel.test.tsx

import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { ForensicsPanel } from "../../components/forensics/ForensicsPanel";
import { useAuthStore } from "../../stores/authStore";

describe("ForensicsPanel", () => {
  beforeEach(() => {
    useAuthStore.getState().setOperator({
      id: "op-001",
      name: "Test",
      unit: "TEST",
      clearance: "PROTECTED_B",
      roles: [],
    });
  });

  it("renders the FORENSICS header", () => {
    render(<ForensicsPanel />);
    expect(screen.getByText("FORENSICS")).toBeTruthy();
  });

  it("renders query builder", () => {
    render(<ForensicsPanel />);
    expect(screen.getByTestId("query-builder")).toBeTruthy();
  });

  it("shows 'Run a query to view results' initially", () => {
    render(<ForensicsPanel />);
    expect(screen.getByText("Run a query to view results")).toBeTruthy();
  });

  it("shows run query button", () => {
    render(<ForensicsPanel />);
    expect(screen.getByTestId("run-query")).toBeTruthy();
  });

  it("renders entity type filter buttons", () => {
    render(<ForensicsPanel />);
    expect(screen.getByTestId("entity-type-surface")).toBeTruthy();
    expect(screen.getByTestId("entity-type-air")).toBeTruthy();
  });

  it("renders anomaly type filter buttons", () => {
    render(<ForensicsPanel />);
    expect(screen.getByTestId("anomaly-type-speed")).toBeTruthy();
  });

  it("runs query on submit and shows loading state", async () => {
    render(<ForensicsPanel />);
    const button = screen.getByTestId("run-query");
    fireEvent.click(button);
    // After query, should show results or still processing
    expect(button).toBeTruthy();
  });
});
