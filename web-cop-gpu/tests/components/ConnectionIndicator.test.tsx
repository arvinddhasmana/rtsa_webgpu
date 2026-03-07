// CLASSIFICATION: UNCLASSIFIED
// tests/components/ConnectionIndicator.test.tsx

import { describe, it, expect, afterEach } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { ConnectionIndicator } from "../../src/components/toolbar/ConnectionIndicator";
import { setWtConnected, setGrpcConnected, setConnecting } from "../../src/signals/connection";
import { setLatencyMs } from "../../src/signals/stats";

afterEach(() => {
  setWtConnected(false);
  setGrpcConnected(true);
  setConnecting(true);
  setLatencyMs(-1);
});

describe("ConnectionIndicator", () => {
  it("renders connection label area", () => {
    render(() => <ConnectionIndicator />);
    expect(screen.getByText("CONNECTION")).toBeDefined();
  });

  it("shows Connecting when connecting is true", () => {
    setConnecting(true);
    render(() => <ConnectionIndicator />);
    expect(screen.getByText("Connecting…")).toBeDefined();
  });

  it("shows Connected when both connections are up", () => {
    setConnecting(false);
    setWtConnected(true);
    setGrpcConnected(true);
    render(() => <ConnectionIndicator />);
    expect(screen.getByText("Connected")).toBeDefined();
  });

  it("shows Disconnected when both connections are down", () => {
    setConnecting(false);
    setWtConnected(false);
    setGrpcConnected(false);
    render(() => <ConnectionIndicator />);
    expect(screen.getByText("Disconnected")).toBeDefined();
  });

  it("shows latency when connected", () => {
    setConnecting(false);
    setWtConnected(true);
    setLatencyMs(25);
    render(() => <ConnectionIndicator />);
    expect(screen.getByText("Latency: 25 ms")).toBeDefined();
  });

  it("shows dash for latency when disconnected", () => {
    setConnecting(false);
    setWtConnected(false);
    setLatencyMs(-1);
    render(() => <ConnectionIndicator />);
    expect(screen.getByText("Latency: —")).toBeDefined();
  });

  it("has correct data-testid", () => {
    render(() => <ConnectionIndicator />);
    expect(screen.getByTestId("connection-indicator")).toBeInTheDocument();
  });
});
