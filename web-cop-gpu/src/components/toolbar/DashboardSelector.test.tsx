// CLASSIFICATION: UNCLASSIFIED
// src/components/toolbar/DashboardSelector.test.tsx — Unit tests for DashboardSelector
//
// Tests:
//   - All 5 dashboard options are rendered (sensor/health/commander/analytics/coverage)
//   - "Coverage" option is present
//   - onChange calls setDashboard with the selected value

import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import userEvent from "@testing-library/user-event";

// ── Mock viewport signals ────────────────────────────────────────────────────

const mockDashboard = vi.hoisted(() => vi.fn(() => "health"));
const mockSetDashboard = vi.hoisted(() => vi.fn());

vi.mock("../../signals/viewport", () => ({
  dashboard: mockDashboard,
  setDashboard: mockSetDashboard,
}));

import { DashboardSelector } from "./DashboardSelector";

describe("DashboardSelector", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockDashboard.mockReturnValue("health");
  });

  it("renders the Dashboard label", () => {
    render(() => <DashboardSelector />);
    expect(screen.getByText("Dashboard")).toBeTruthy();
  });

  it("renders all 5 dashboard options", () => {
    render(() => <DashboardSelector />);
    const select = screen.getByRole("combobox");
    const options = Array.from(select.querySelectorAll("option")).map((o) => o.textContent);
    expect(options).toContain("Map view");
    expect(options).toContain("Health");
    expect(options).toContain("Commander");
    expect(options).toContain("Analytics");
    expect(options).toContain("Coverage");
  });

  it("renders the 'Coverage' option with value='coverage'", () => {
    render(() => <DashboardSelector />);
    const select = screen.getByRole("combobox");
    const coverageOption = Array.from(select.querySelectorAll("option")).find(
      (o) => o.getAttribute("value") === "coverage",
    );
    expect(coverageOption).toBeDefined();
    expect(coverageOption?.textContent).toBe("Coverage");
  });

  it("onChange calls setDashboard with the selected value", async () => {
    const user = userEvent.setup();
    render(() => <DashboardSelector />);
    const select = screen.getByRole("combobox");
    await user.selectOptions(select, "coverage");
    expect(mockSetDashboard).toHaveBeenCalledWith("coverage");
  });

  it("onChange calls setDashboard('health') when Health is selected", async () => {
    mockDashboard.mockReturnValue("sensor");
    const user = userEvent.setup();
    render(() => <DashboardSelector />);
    const select = screen.getByRole("combobox");
    await user.selectOptions(select, "health");
    expect(mockSetDashboard).toHaveBeenCalledWith("health");
  });

  it("has a data-testid for E2E test selection", () => {
    render(() => <DashboardSelector />);
    expect(screen.getByTestId("dashboard-selector")).toBeTruthy();
  });
});
