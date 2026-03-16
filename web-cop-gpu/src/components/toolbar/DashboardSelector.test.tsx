// CLASSIFICATION: UNCLASSIFIED
// src/components/toolbar/DashboardSelector.test.tsx — Unit tests for DashboardSelector
//
// Tests:
//   - Dashboard options are role-scoped
//   - Operations Commander gets Fusion/Multi-Domain/Operator UI tabs
//   - onChange calls setDashboard with selected value

import { render, screen } from "@solidjs/testing-library";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

// ── Mock viewport signals ────────────────────────────────────────────────────

const mockRole = vi.hoisted(() => vi.fn(() => "sensor_operator"));
const mockDashboard = vi.hoisted(() => vi.fn(() => "health"));
const mockSetDashboard = vi.hoisted(() => vi.fn());
const mockRoleAllowedDashboards = vi.hoisted(
  () =>
    ({
      operations_commander: ["commander", "coverage", "analytics"],
      intelligence_analyst: ["analytics", "sensor"],
      security_officer: ["commander"],
      sensor_operator: ["health", "coverage"],
      nato_liaison: ["sensor"],
    }) as const,
);

vi.mock("../../signals/viewport", () => ({
  role: mockRole,
  dashboard: mockDashboard,
  setDashboard: mockSetDashboard,
  ROLE_ALLOWED_DASHBOARDS: mockRoleAllowedDashboards,
}));

import { DashboardSelector } from "./DashboardSelector";

describe("DashboardSelector", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockRole.mockReturnValue("sensor_operator");
    mockDashboard.mockReturnValue("health");
  });

  it("renders the Dashboard label", () => {
    render(() => <DashboardSelector />);
    expect(screen.getByText("Dashboard")).toBeTruthy();
  });

  it("renders role-scoped options for sensor operator", () => {
    render(() => <DashboardSelector />);
    const select = screen.getByRole("combobox");
    const options = Array.from(select.querySelectorAll("option")).map(
      (o) => o.textContent,
    );
    expect(options).toContain("Sensor Health");
    expect(options).toContain("Coverage");
    expect(options).not.toContain("Fusion");
  });

  it("renders commander-specific tabs for operations commander", () => {
    mockRole.mockReturnValue("operations_commander");
    mockDashboard.mockReturnValue("commander");

    render(() => <DashboardSelector />);
    const select = screen.getByRole("combobox");
    const options = Array.from(select.querySelectorAll("option")).map(
      (o) => o.textContent,
    );

    expect(options).toContain("Fusion");
    expect(options).toContain("Multi-Domain");
    expect(options).toContain("Operator UI");
    expect(options).not.toContain("Sensor Health");
  });

  it("renders the Coverage option with value='coverage' for sensor operator", () => {
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

  it("onChange calls setDashboard('health') when Sensor Health is selected", async () => {
    mockDashboard.mockReturnValue("coverage");
    const user = userEvent.setup();
    render(() => <DashboardSelector />);
    const select = screen.getByRole("combobox");
    await user.selectOptions(select, "health");
    expect(mockSetDashboard).toHaveBeenCalledWith("health");
  });

  it("supports commander dashboard transitions across Fusion, Multi-Domain, and Operator UI", async () => {
    mockRole.mockReturnValue("operations_commander");
    mockDashboard.mockReturnValue("commander");

    const user = userEvent.setup();
    render(() => <DashboardSelector />);
    const select = screen.getByRole("combobox");

    await user.selectOptions(select, "coverage");
    await user.selectOptions(select, "analytics");

    expect(mockSetDashboard).toHaveBeenCalledWith("coverage");
    expect(mockSetDashboard).toHaveBeenCalledWith("analytics");
  });

  it("has a data-testid for E2E test selection", () => {
    render(() => <DashboardSelector />);
    expect(screen.getByTestId("dashboard-selector")).toBeTruthy();
  });
});
