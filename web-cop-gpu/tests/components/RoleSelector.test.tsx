// CLASSIFICATION: UNCLASSIFIED
// tests/components/RoleSelector.test.tsx

import { fireEvent, render, screen } from "@solidjs/testing-library";
import { afterEach, describe, expect, it } from "vitest";
import { RoleSelector } from "../../src/components/toolbar/RoleSelector";
import { role, setRole } from "../../src/signals/viewport";

afterEach(() => {
  setRole("sensor_operator");
});

describe("RoleSelector", () => {
  it("renders role label", () => {
    render(() => <RoleSelector />);
    expect(screen.getByText("Role")).toBeDefined();
  });

  it("shows current role as selected", () => {
    render(() => <RoleSelector />);
    const select = screen.getByRole("combobox") as HTMLSelectElement;
    expect(select.value).toBe("sensor_operator");
  });

  it("changes role when a new option is selected", () => {
    render(() => <RoleSelector />);
    const select = screen.getByRole("combobox") as HTMLSelectElement;
    fireEvent.change(select, { target: { value: "operations_commander" } });
    expect(role()).toBe("operations_commander");
  });

  it("has data-testid role-selector on root element", () => {
    render(() => <RoleSelector />);
    expect(screen.getByTestId("role-selector")).toBeDefined();
  });
});
