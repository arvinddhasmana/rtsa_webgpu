// CLASSIFICATION: UNCLASSIFIED
// tests/components/StatusBar.test.tsx

import { describe, it, expect, afterEach } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { StatusBar } from "../../src/components/status/StatusBar";
import { setFps, setTrackCount, setVisibleCount, setLatencyMs } from "../../src/signals/stats";
import { setWtConnected } from "../../src/signals/connection";

afterEach(() => {
  setFps(0);
  setTrackCount(0);
  setVisibleCount(0);
  setLatencyMs(-1);
  setWtConnected(false);
});

describe("StatusBar", () => {
  it("renders the status bar container", () => {
    render(() => <StatusBar />);
    expect(screen.getByLabelText("Status bar")).toBeDefined();
  });

  it("displays FPS label", () => {
    render(() => <StatusBar />);
    expect(screen.getByText("FPS")).toBeDefined();
  });

  it("displays track count label", () => {
    render(() => <StatusBar />);
    expect(screen.getByText("Tracks")).toBeDefined();
  });

  it("displays visible count label", () => {
    render(() => <StatusBar />);
    expect(screen.getByText("Visible")).toBeDefined();
  });

  it("reflects updated FPS signal", () => {
    setFps(60);
    render(() => <StatusBar />);
    expect(screen.getByText("60")).toBeDefined();
  });

  it("reflects updated track count signal", () => {
    setTrackCount(5000);
    render(() => <StatusBar />);
    expect(screen.getByText("5000")).toBeDefined();
  });

  it("shows WT connected OK", () => {
    setWtConnected(true);
    render(() => <StatusBar />);
    expect(screen.getByText("OK")).toBeDefined();
  });

  it("shows WT disconnected dash", () => {
    setWtConnected(false);
    render(() => <StatusBar />);
    // The "—" character is used for disconnected state
    const wt = screen.getAllByText("—");
    expect(wt.length).toBeGreaterThan(0);
  });

  it("shows latency when connected", () => {
    setLatencyMs(42);
    render(() => <StatusBar />);
    expect(screen.getByText("42 ms")).toBeDefined();
  });

  it("has correct data-testids", () => {
    setFps(60);
    setLatencyMs(42);
    render(() => <StatusBar />);
    expect(screen.getByTestId("status-bar")).toBeInTheDocument();
    expect(screen.getByTestId("fps-display")).toBeInTheDocument();
    expect(screen.getByTestId("latency-display")).toBeInTheDocument();
  });
});
