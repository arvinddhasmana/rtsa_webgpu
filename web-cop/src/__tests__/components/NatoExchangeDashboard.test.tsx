// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/NatoExchangeDashboard.test.tsx

import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { NatoExchangeDashboard } from "../../components/layout/NatoExchangeDashboard";

// Mock MapView to avoid MapLibre dependencies
vi.mock("../../components/map/MapView", () => ({
  MapView: () => <div data-testid="mock-map-view" />,
}));

describe("NatoExchangeDashboard", () => {
  it("renders the dashboard container", () => {
    render(<NatoExchangeDashboard />);
    expect(screen.getByTestId("nato-exchange-dashboard")).toBeTruthy();
  });

  it("shows link status header", () => {
    render(<NatoExchangeDashboard />);
    expect(screen.getByTestId("link-status-header")).toBeTruthy();
    expect(screen.getByText("NATO LINKS")).toBeTruthy();
  });

  it("renders nomination queue", () => {
    render(<NatoExchangeDashboard />);
    expect(screen.getByTestId("nomination-queue")).toBeTruthy();
  });

  it("renders inbound tracks panel", () => {
    render(<NatoExchangeDashboard />);
    expect(screen.getByTestId("inbound-tracks-panel")).toBeTruthy();
  });

  it("shows Link 16 status indicator", () => {
    render(<NatoExchangeDashboard />);
    expect(screen.getByTestId("link-status-Link-16")).toBeTruthy();
  });
});
