// CLASSIFICATION: UNCLASSIFIED
// tests/components/DashboardSelector.test.tsx

import { fireEvent, render, screen } from "@solidjs/testing-library";
import { afterEach, describe, expect, it } from "vitest";
import { DashboardSelector } from "../../src/components/toolbar/DashboardSelector";
import {
  dashboard,
  role,
  setDashboard,
  setRole,
} from "../../src/signals/viewport";

afterEach(() => {
  setRole("sensor_operator");
  setDashboard("health");
});

describe("DashboardSelector", () => {
  it("renders dashboard label", () => {
    render(() => <DashboardSelector />);
    expect(screen.getByText("Dashboard")).toBeDefined();
  });

  it("shows current dashboard as selected", () => {
    render(() => <DashboardSelector />);
    const select = screen.getByRole("combobox") as HTMLSelectElement;
    expect(select.value).toBe("health");
  });

  it("changes dashboard when selecting an allowed dashboard", () => {
    render(() => <DashboardSelector />);
    const select = screen.getByRole("combobox") as HTMLSelectElement;
    fireEvent.change(select, { target: { value: "coverage" } });
    expect(dashboard()).toBe("coverage");
  });

  it("resets to role default when selecting a disallowed dashboard", () => {
    render(() => <DashboardSelector />);
    const select = screen.getByRole("combobox") as HTMLSelectElement;
    fireEvent.change(select, { target: { value: "commander" } });
    expect(role()).toBe("sensor_operator");
    expect(dashboard()).toBe("health");
  });

  it("has data-testid dashboard-selector on root element", () => {
    render(() => <DashboardSelector />);
    expect(screen.getByTestId("dashboard-selector")).toBeDefined();
  });

  it("operations commander can switch across commander, coverage, and analytics", () => {
    setRole("operations_commander");
    render(() => <DashboardSelector />);

    const select = screen.getByRole("combobox") as HTMLSelectElement;
    expect(select.value).toBe("commander");

    fireEvent.change(select, { target: { value: "coverage" } });
    expect(dashboard()).toBe("coverage");

    fireEvent.change(select, { target: { value: "analytics" } });
    expect(dashboard()).toBe("analytics");
  });

  it("sensor operator dashboard options remain health and coverage only", () => {
    setRole("sensor_operator");
    render(() => <DashboardSelector />);

    const select = screen.getByRole("combobox") as HTMLSelectElement;
    const optionValues = Array.from(select.options).map((o) => o.value);
    expect(optionValues).toEqual(["health", "coverage"]);
  });
});
