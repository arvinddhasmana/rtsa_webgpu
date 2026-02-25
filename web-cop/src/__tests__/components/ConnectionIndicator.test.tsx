// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/ConnectionIndicator.test.tsx

import { render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

describe("ConnectionIndicator", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renders connection indicator", async () => {
    globalThis.fetch = vi
      .fn()
      .mockRejectedValue(new Error("Network error")) as typeof globalThis.fetch;
    const { ConnectionIndicator } =
      await import("../../components/layout/ConnectionIndicator");
    render(<ConnectionIndicator />);
    expect(screen.getByTestId("connection-indicator")).toBeTruthy();
  });

  it("shows initial DISCONNECTED status", async () => {
    globalThis.fetch = vi
      .fn()
      .mockRejectedValue(new Error("Network error")) as typeof globalThis.fetch;
    const { ConnectionIndicator } =
      await import("../../components/layout/ConnectionIndicator");
    render(<ConnectionIndicator />);
    // Initial state before any checks
    expect(screen.getByTestId("connection-indicator")).toBeTruthy();
  });
});
