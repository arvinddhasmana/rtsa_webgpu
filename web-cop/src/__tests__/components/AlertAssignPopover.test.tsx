// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/AlertAssignPopover.test.tsx

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { AlertAssignPopover } from "../../components/alert/AlertAssignPopover";

describe("AlertAssignPopover", () => {
  const mockOnClose = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders with available operators", () => {
    render(<AlertAssignPopover alertId="ALT-1" onClose={mockOnClose} onAssign={vi.fn()} />);
    expect(screen.getByText("ASSIGN ALERT")).toBeTruthy();

    // Check initial focus element / operators rendered
    expect(screen.getByTestId("assign-op-op-delta-2")).toBeTruthy();
    expect(screen.getByTestId("assign-op-op-echo-3")).toBeTruthy();
  });

  it("handles operator selection and enables confirm button", () => {
    render(<AlertAssignPopover alertId="ALT-1" onClose={mockOnClose} onAssign={vi.fn()} />);

    const confirmButton = screen.getByTestId("assign-confirm-btn");
    expect(confirmButton).toBeDisabled();

    // Select Jane Doe
    const op2 = screen.getByTestId("assign-op-op-delta-2");
    fireEvent.click(op2);

    expect(confirmButton).not.toBeDisabled();
  });

  it("shows assignment confirmation toast and closes", async () => {
    render(<AlertAssignPopover alertId="ALT-1" onClose={mockOnClose} onAssign={vi.fn()} />);

    // Select operator and confirm
    fireEvent.click(screen.getByTestId("assign-op-op-delta-2"));
    fireEvent.click(screen.getByTestId("assign-confirm-btn"));

    // Toast appears
    expect(screen.getByText("Alert Assigned")).toBeTruthy();

    // Check mock is called after delay
    await waitFor(() => {
      expect(mockOnClose).toHaveBeenCalled();
    }, { timeout: 2500 }); // The component has a 2 sec delay
  });

  it("closes when escape key is pressed", () => {
    render(<AlertAssignPopover alertId="ALT-1" onClose={mockOnClose} onAssign={vi.fn()} />);

    fireEvent.keyDown(document.body, { key: "Escape", code: "Escape" });

    expect(mockOnClose).toHaveBeenCalled();
  });
});
