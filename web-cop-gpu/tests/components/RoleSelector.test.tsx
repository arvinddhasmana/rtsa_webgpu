// CLASSIFICATION: UNCLASSIFIED
// tests/components/RoleSelector.test.tsx

import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, fireEvent } from "@solidjs/testing-library";
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

  it("calls onChange callback when role changes", () => {
    const onChange = vi.fn();
    render(() => <RoleSelector onChange={onChange} />);
    const select = screen.getByRole("combobox") as HTMLSelectElement;
    fireEvent.change(select, { target: { value: "operations_commander" } });
    expect(onChange).toHaveBeenCalledWith("operations_commander");
  });
});
