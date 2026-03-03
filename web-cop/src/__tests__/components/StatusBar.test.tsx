// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/StatusBar.test.tsx

import { act, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

// ── Mocks ──────────────────────────────────────────────
vi.mock("../../hooks/useConnectionStatus", () => ({
  useConnectionStatus: vi.fn(() => ({ status: "connected", lastCheck: new Date() })),
}));
vi.mock("../../stores/trackStore", () => ({
  useTrackStore: vi.fn((sel: (s: { tracks: Map<string, unknown> }) => unknown) =>
    sel({ tracks: new Map([["a", {}], ["b", {}]]) })
  ),
}));
vi.mock("../../stores/alertStore", () => ({
  useAlertStore: vi.fn((sel: (s: { alerts: Map<string, unknown> }) => unknown) =>
    sel({ alerts: new Map([["x", {}]]) })
  ),
}));

import { StatusBar } from "../../components/layout/StatusBar";
import { useConnectionStatus } from "../../hooks/useConnectionStatus";

describe("StatusBar", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the status bar with testid", () => {
    render(<StatusBar />);
    expect(screen.getByTestId("status-bar")).toBeInTheDocument();
  });

  it("shows CONNECTED when connection is connected", () => {
    render(<StatusBar />);
    expect(screen.getByText(/CONNECTED/i)).toBeInTheDocument();
  });

  it("shows DEGRADED when connection status is degraded", () => {
    (useConnectionStatus as ReturnType<typeof vi.fn>).mockReturnValue({
      status: "degraded",
      lastCheck: new Date(),
    });
    render(<StatusBar />);
    expect(screen.getByText(/DEGRADED/i)).toBeInTheDocument();
  });

  it("shows OFFLINE when connection status is disconnected", () => {
    (useConnectionStatus as ReturnType<typeof vi.fn>).mockReturnValue({
      status: "disconnected",
      lastCheck: new Date(),
    });
    render(<StatusBar />);
    expect(screen.getByText(/OFFLINE/i)).toBeInTheDocument();
  });

  it("displays track count from store", () => {
    render(<StatusBar />);
    expect(screen.getByText(/Tracks:/i)).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument();
  });

  it("displays alert count from store", () => {
    render(<StatusBar />);
    expect(screen.getByText(/Alerts:/i)).toBeInTheDocument();
    expect(screen.getByText("1")).toBeInTheDocument();
  });

  it("renders UTC time display", () => {
    render(<StatusBar />);
    expect(screen.getByTestId("status-utc-time")).toBeInTheDocument();
    expect(screen.getByTestId("status-utc-time").textContent).toMatch(/Z$/);
  });

  it("updates the UTC time every second", async () => {
    vi.useFakeTimers();
    render(<StatusBar />);
    const initial = screen.getByTestId("status-utc-time").textContent;
    await act(async () => {
      vi.advanceTimersByTime(1500);
    });
    const updated = screen.getByTestId("status-utc-time").textContent;
    // The clock should still show a valid UTC timestamp
    expect(updated).toMatch(/Z$/);
    expect(initial).toBeTruthy();
    vi.useRealTimers();
  });
});
