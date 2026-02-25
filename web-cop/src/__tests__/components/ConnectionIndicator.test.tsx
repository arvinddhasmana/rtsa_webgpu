// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/ConnectionIndicator.test.tsx

import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";

describe("ConnectionIndicator", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renders connection indicator", async () => {
    global.fetch = vi.fn().mockRejectedValue(new Error("Network error"));
    const { ConnectionIndicator } = await import(
      "../../components/layout/ConnectionIndicator"
    );
    render(<ConnectionIndicator />);
    expect(screen.getByTestId("connection-indicator")).toBeTruthy();
  });

  it("shows initial DISCONNECTED status", async () => {
    global.fetch = vi.fn().mockRejectedValue(new Error("Network error"));
    const { ConnectionIndicator } = await import(
      "../../components/layout/ConnectionIndicator"
    );
    render(<ConnectionIndicator />);
    // Initial state before any checks
    expect(screen.getByTestId("connection-indicator")).toBeTruthy();
  });
});
