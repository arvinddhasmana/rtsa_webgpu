// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/IntelSearchDashboard.test.tsx

import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { IntelSearchDashboard } from "../../components/layout/IntelSearchDashboard";

// Mock query client
vi.mock("../../api/query-client", () => ({
  queryClient: {
    queryHistory: vi.fn().mockResolvedValue({ tracks: [], alerts: [], totalCount: 0 }),
    getEventTimeline: vi.fn().mockResolvedValue({ events: [] }),
  },
  HistoricalQueryRequest: class {},
}));

describe("IntelSearchDashboard", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the intel search dashboard", () => {
    render(<IntelSearchDashboard />);
    expect(screen.getByTestId("intel-search-dashboard")).toBeTruthy();
  });

  it("shows INTEL SEARCH heading", () => {
    render(<IntelSearchDashboard />);
    expect(screen.getByText("INTEL SEARCH")).toBeTruthy();
  });

  it("renders the query builder panel", () => {
    render(<IntelSearchDashboard />);
    // QueryBuilder renders a form-like structure with data-testid
    expect(screen.getByTestId("query-builder")).toBeTruthy();
  });

  it("shows empty state prompt when no results", () => {
    render(<IntelSearchDashboard />);
    expect(
      screen.getByText("Build a query to search historical intelligence data")
    ).toBeTruthy();
  });

  it("shows helper text in results header when no query run", () => {
    render(<IntelSearchDashboard />);
    expect(
      screen.getByText("Use the query builder to search historical track data")
    ).toBeTruthy();
  });
});
