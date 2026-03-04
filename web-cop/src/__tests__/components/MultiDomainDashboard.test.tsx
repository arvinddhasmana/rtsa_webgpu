// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/MultiDomainDashboard.test.tsx

import { act, fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { MultiDomainDashboard } from "../../components/layout/MultiDomainDashboard";
import { useUIStore } from "../../stores/uiStore";

// Mock MapView to avoid MapLibre GL dependency issues in JSDOM
vi.mock("../../components/map/MapView", () => ({
  MapView: () => <div data-testid="mock-map-view"></div>,
}));
// Mock DetailPanel and AlertPanel to isolate current setup
vi.mock("../../components/detail/DetailPanel", () => ({
  DetailPanel: () => <div data-testid="mock-detail-panel"></div>,
}));
vi.mock("../../components/alerts/AlertPanel", () => ({
  AlertPanel: () => <div data-testid="mock-alert-panel"></div>,
}));

describe("MultiDomainDashboard", () => {
  beforeEach(() => {
    // toggleLayerVisibility is the actual method on the store
    const store = useUIStore.getState();
    if (store.layerVisibility.sensorCoverage) {
        store.toggleLayerVisibility("sensorCoverage");
    }
    store.closeDetailPanel();
  });

  it("renders main areas of the dashboard", () => {
    render(<MultiDomainDashboard />);

    expect(screen.getByTestId("multi-domain-dashboard")).toBeTruthy();
    expect(screen.getByTestId("mock-map-view")).toBeTruthy();
    expect(screen.getByTestId("domain-metrics-overlay")).toBeTruthy();
    expect(screen.getByText("🚨 ALERTS")).toBeTruthy();
  });

  it("toggles map layers when buttons are clicked", () => {
    render(<MultiDomainDashboard />);

    const coverageBtn = screen.getByText("🛰 Coverage");
    fireEvent.click(coverageBtn);

    expect(useUIStore.getState().layerVisibility.sensorCoverage).toBe(true);
  });

  it("expands and collapses the alert strip", () => {
    render(<MultiDomainDashboard />);

    // Initially closed
    expect(screen.queryByTestId("mock-alert-panel")).toBeNull();

    // Click strip header
    const stripHeader = screen.getByText("🚨 ALERTS").parentElement;
    if (stripHeader) fireEvent.click(stripHeader);

    // Should be open
    expect(screen.getByTestId("mock-alert-panel")).toBeTruthy();
    expect(screen.getByText("▼ Collapse")).toBeTruthy();

    // Click strip header again
    if (stripHeader) fireEvent.click(stripHeader);

    // Should be closed
    expect(screen.queryByTestId("mock-alert-panel")).toBeNull();
    expect(screen.getByText("▲ Expand")).toBeTruthy();
  });

  it("shows detail panel when triggered in UI store", () => {
    render(<MultiDomainDashboard />);

    expect(screen.queryByTestId("mock-detail-panel")).toBeNull();

    act(() => {
        useUIStore.getState().toggleDetailPanel();
    });
    // wait for re-render implicitly via react testing library query
    expect(screen.getByTestId("mock-detail-panel")).toBeTruthy();
  });
});
