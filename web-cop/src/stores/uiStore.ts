// CLASSIFICATION: UNCLASSIFIED
// src/stores/uiStore.ts

import { create } from "zustand";

export type Theme = "light" | "dark" | "nvg";

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

  toggleAlertPanel: () => void;
  toggleDetailPanel: () => void;
  toggleForensicsPanel: () => void;
  setMapView: (center: [number, number], zoom: number) => void;
  setEntityTypeFilter: (types: string[]) => void;
  setHostileClassFilter: (classes: string[]) => void;
  toggleStaleTracks: () => void;
  setTheme: (theme: Theme) => void;
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

  toggleAlertPanel: () => set((s) => ({ alertPanelOpen: !s.alertPanelOpen })),
  toggleDetailPanel: () =>
    set((s) => ({ detailPanelOpen: !s.detailPanelOpen })),
  toggleForensicsPanel: () =>
    set((s) => ({ forensicsPanelOpen: !s.forensicsPanelOpen })),
  setMapView: (center, zoom) => set({ mapCenter: center, mapZoom: zoom }),
  setEntityTypeFilter: (types) => set({ entityTypeFilter: types }),
  setHostileClassFilter: (classes) => set({ hostileClassFilter: classes }),
  toggleStaleTracks: () =>
    set((s) => ({ showStaleTracksEnabled: !s.showStaleTracksEnabled })),
  setTheme: (theme) => set({ theme }),
}));
