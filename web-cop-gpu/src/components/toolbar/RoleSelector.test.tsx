// CLASSIFICATION: UNCLASSIFIED
// src/components/toolbar/RoleSelector.test.tsx — Unit tests for RoleSelector

import { render, screen } from "@solidjs/testing-library";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

const mockRole = vi.hoisted(() => vi.fn(() => "sensor_operator"));
const mockSetRole = vi.hoisted(() => vi.fn());

vi.mock("../../signals/viewport", () => ({
  role: mockRole,
  setRole: mockSetRole,
}));

import { RoleSelector } from "./RoleSelector";

describe("RoleSelector", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockRole.mockReturnValue("sensor_operator");
  });

  it("renders all five role options", () => {
    render(() => <RoleSelector />);
    const select = screen.getByRole("combobox");
    const options = Array.from(select.querySelectorAll("option")).map(
      (o) => o.textContent,
    );

    expect(options).toContain("Ops Commander");
    expect(options).toContain("Intelligence Analyst");
    expect(options).toContain("Security Officer");
    expect(options).toContain("Sensor Operator");
    expect(options).toContain("NATO Liaison");
  });

  it("onChange calls setRole with the selected role", async () => {
    const user = userEvent.setup();
    render(() => <RoleSelector />);

    const select = screen.getByRole("combobox");
    await user.selectOptions(select, "operations_commander");

    expect(mockSetRole).toHaveBeenCalledWith("operations_commander");
  });

  it("exposes test id for integration and e2e selectors", () => {
    render(() => <RoleSelector />);
    expect(screen.getByTestId("role-selector")).toBeTruthy();
  });
});
