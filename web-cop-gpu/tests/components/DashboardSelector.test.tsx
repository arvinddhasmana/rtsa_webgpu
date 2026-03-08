// CLASSIFICATION: UNCLASSIFIED
// tests/components/DashboardSelector.test.tsx

import { describe, it, expect, afterEach } from "vitest";
import { render, screen, fireEvent } from "@solidjs/testing-library";
import { DashboardSelector } from "../../src/components/toolbar/DashboardSelector";
import { dashboard, setDashboard } from "../../src/signals/viewport";

afterEach(() => {
  setDashboard("sensor");
});

describe("DashboardSelector", () => {
  it("renders dashboard label", () => {
    render(() => <DashboardSelector />);
    expect(screen.getByText("Dashboard")).toBeDefined();
  });

  it("shows current dashboard as selected", () => {
    render(() => <DashboardSelector />);
    const select = screen.getByRole("combobox") as HTMLSelectElement;
    expect(select.value).toBe("sensor");
  });

  it("changes dashboard when a new option is selected", () => {
    render(() => <DashboardSelector />);
    const select = screen.getByRole("combobox") as HTMLSelectElement;
    fireEvent.change(select, { target: { value: "commander" } });
    expect(dashboard()).toBe("commander");
  });

  it("has data-testid dashboard-selector on root element", () => {
    render(() => <DashboardSelector />);
    expect(screen.getByTestId("dashboard-selector")).toBeDefined();
  });
});
