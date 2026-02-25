// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/stores/uiStore.test.ts

import { describe, it, expect, beforeEach } from "vitest";
import { useUIStore } from "../../stores/uiStore";

describe("UIStore", () => {
  beforeEach(() => {
    // Reset to initial state
    useUIStore.setState({
      alertPanelOpen: true,
      detailPanelOpen: false,
      forensicsPanelOpen: false,
      mapCenter: [-60.0, 45.0],
      mapZoom: 6,
      entityTypeFilter: [],
      hostileClassFilter: [],
      showStaleTracksEnabled: false,
    });
  });

  it("T07: toggleAlertPanel toggles alertPanelOpen", () => {
    expect(useUIStore.getState().alertPanelOpen).toBe(true);
    useUIStore.getState().toggleAlertPanel();
    expect(useUIStore.getState().alertPanelOpen).toBe(false);
    useUIStore.getState().toggleAlertPanel();
    expect(useUIStore.getState().alertPanelOpen).toBe(true);
  });

  it("toggleDetailPanel toggles detailPanelOpen", () => {
    expect(useUIStore.getState().detailPanelOpen).toBe(false);
    useUIStore.getState().toggleDetailPanel();
    expect(useUIStore.getState().detailPanelOpen).toBe(true);
  });

  it("toggleForensicsPanel toggles forensicsPanelOpen", () => {
    expect(useUIStore.getState().forensicsPanelOpen).toBe(false);
    useUIStore.getState().toggleForensicsPanel();
    expect(useUIStore.getState().forensicsPanelOpen).toBe(true);
  });

  it("initial mapCenter is Mid-Atlantic [-60, 45]", () => {
    expect(useUIStore.getState().mapCenter).toEqual([-60.0, 45.0]);
  });

  it("initial mapZoom is 6", () => {
    expect(useUIStore.getState().mapZoom).toBe(6);
  });

  it("setMapView updates center and zoom", () => {
    useUIStore.getState().setMapView([-70.0, 50.0], 8);
    expect(useUIStore.getState().mapCenter).toEqual([-70.0, 50.0]);
    expect(useUIStore.getState().mapZoom).toBe(8);
  });

  it("setEntityTypeFilter updates filter", () => {
    useUIStore.getState().setEntityTypeFilter(["AIR", "SURFACE"]);
    expect(useUIStore.getState().entityTypeFilter).toEqual(["AIR", "SURFACE"]);
  });

  it("setHostileClassFilter updates filter", () => {
    useUIStore.getState().setHostileClassFilter(["HOSTILE"]);
    expect(useUIStore.getState().hostileClassFilter).toEqual(["HOSTILE"]);
  });

  it("toggleStaleTracks toggles showStaleTracksEnabled", () => {
    expect(useUIStore.getState().showStaleTracksEnabled).toBe(false);
    useUIStore.getState().toggleStaleTracks();
    expect(useUIStore.getState().showStaleTracksEnabled).toBe(true);
    useUIStore.getState().toggleStaleTracks();
    expect(useUIStore.getState().showStaleTracksEnabled).toBe(false);
  });
});
