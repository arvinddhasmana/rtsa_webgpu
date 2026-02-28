// CLASSIFICATION: UNCLASSIFIED
// src/stores/uiStore.ts

import { create } from "zustand";

export type Theme = "light" | "dark" | "nvg" | "high-contrast";
export type ActiveRole = "commander" | "security" | "analyst" | "sensor_operator" | "nato_liaison";
export type DashboardView = "fusion" | "multi-domain" | "operator" | "forensics" | "audit" | "sensor-health" | "nato-exchange";
export type LayerKey = "trackLabels" | "trackTrails" | "sensorCoverage" | "geofences" | "mgrsGrid";

interface UIState {
  alertPanelOpen: boolean;
  detailPanelOpen: boolean;
  forensicsPanelOpen: boolean;

  mapCenter: [number, number];
  mapZoom: number;

  entityTypeFilter: string[];
  hostileClassFilter: string[];
  showStaleTracksEnabled: boolean;

  theme: Theme;
  activeRole: ActiveRole;
  activeDashboardView: DashboardView;

  layerVisibility: Record<LayerKey, boolean>;

  searchOpen: boolean;
  searchQuery: string;

  toggleAlertPanel: () => void;
  toggleDetailPanel: () => void;
  closeDetailPanel: () => void;
  toggleForensicsPanel: () => void;
  setMapView: (center: [number, number], zoom: number) => void;
  setEntityTypeFilter: (types: string[]) => void;
  setHostileClassFilter: (classes: string[]) => void;
  toggleStaleTracks: () => void;
  setTheme: (theme: Theme) => void;
  setActiveRole: (role: ActiveRole) => void;
  setDashboardView: (view: DashboardView) => void;
  toggleLayerVisibility: (layer: LayerKey) => void;
  openSearch: () => void;
  closeSearch: () => void;
  setSearchQuery: (query: string) => void;
}

export const useUIStore = create<UIState>((set) => ({
  alertPanelOpen: true,
  detailPanelOpen: false,
  forensicsPanelOpen: false,

  // Default: Mid-Atlantic (43-47°N, 55-65°W)
  mapCenter: [-60.0, 45.0],
  mapZoom: 6,

  entityTypeFilter: [],
  hostileClassFilter: [],
  showStaleTracksEnabled: false,

  theme: "dark",
  activeRole: "commander",
  activeDashboardView: "fusion",
  layerVisibility: {
    trackLabels: true,
    trackTrails: false,
    sensorCoverage: true,
    geofences: true,
    mgrsGrid: false,
  },
  searchOpen: false,
  searchQuery: "",

  toggleAlertPanel: () => set((s) => ({ alertPanelOpen: !s.alertPanelOpen })),
  toggleDetailPanel: () =>
    set((s) => ({ detailPanelOpen: !s.detailPanelOpen })),
  closeDetailPanel: () => set({ detailPanelOpen: false }),
  toggleForensicsPanel: () =>
    set((s) => ({ forensicsPanelOpen: !s.forensicsPanelOpen })),
  setMapView: (center, zoom) => set({ mapCenter: center, mapZoom: zoom }),
  setEntityTypeFilter: (types) => set({ entityTypeFilter: types }),
  setHostileClassFilter: (classes) => set({ hostileClassFilter: classes }),
  toggleStaleTracks: () =>
    set((s) => ({ showStaleTracksEnabled: !s.showStaleTracksEnabled })),
  setTheme: (theme) => set({ theme }),
  setActiveRole: (role) => set({ activeRole: role }),
  setDashboardView: (view) => set({ activeDashboardView: view }),
  toggleLayerVisibility: (layer) =>
    set((s) => ({
      layerVisibility: {
        ...s.layerVisibility,
        [layer]: !s.layerVisibility[layer],
      },
    })),
  openSearch: () => set({ searchOpen: true }),
  closeSearch: () => set({ searchOpen: false }),
  setSearchQuery: (query) => set({ searchQuery: query }),
}));
