// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/MainLayout.tsx

import React from "react";
import { useAlertStream } from "../../hooks/useAlertStream";
import { useClassification } from "../../hooks/useClassification";
import { useTrackStream } from "../../hooks/useTrackStream";
import { useAuthStore } from "../../stores/authStore";
import { useUIStore, type Theme } from "../../stores/uiStore";
import { AlertPanel } from "../alerts/AlertPanel";
import { DetailPanel } from "../detail/DetailPanel";
import { ForensicsPanel } from "../forensics/ForensicsPanel";
import { MapView } from "../map/MapView";
import { ClassificationBanner } from "./ClassificationBanner";
import { CollapsiblePane } from "./CollapsiblePane";
import { ConnectionIndicator } from "./ConnectionIndicator";
import { DashboardSelector } from "./DashboardSelector";
import { RoleSelector } from "./RoleSelector";
import { SearchOverlay } from "./SearchOverlay";
import { SensorHealthPanel } from "./SensorHealthPanel";

const toolbarButtonStyle: React.CSSProperties = {
  padding: "4px 12px",
  backgroundColor: "#374151",
  color: "#F1F5F9",
  border: "none",
  borderRadius: "4px",
  cursor: "pointer",
  fontSize: "0.75rem",
};

/**
 * MainLayout — root grid layout with classification banners, map, alert panel,
 * detail panel, and forensics panel.
 *
 * Layout:
 *   [ClassificationBanner top]
 *   [MapView 70% | AlertPanel 30%]
 *   [DetailPanel (collapsible bottom)]
 *   [ForensicsPanel (collapsible bottom)]
 *   [ClassificationBanner bottom]
 */
export const MainLayout: React.FC = () => {
  const { effectiveClassification } = useClassification();
  const forensicsPanelOpen = useUIStore((s) => s.forensicsPanelOpen);
  const detailPanelOpen = useUIStore((s) => s.detailPanelOpen);
  const toggleForensicsPanel = useUIStore((s) => s.toggleForensicsPanel);
  const closeDetailPanel = useUIStore((s) => s.closeDetailPanel);
  const openSearch = useUIStore((s) => s.openSearch);
  const theme = useUIStore((s) => s.theme);
  const setTheme = useUIStore((s) => s.setTheme);
  const activeDashboardView = useUIStore((s) => s.activeDashboardView);
  const toggleAlertPanel = useUIStore((s) => s.toggleAlertPanel);

  const operatorName = useAuthStore((s) => s.operatorName) || "Operator";
  const operatorClearance = useAuthStore((s) => s.clearanceLevel);

  // Keyboard Shortcuts (Component 10)
  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Don't intercept if typing in an input
      if (["INPUT", "TEXTAREA", "SELECT"].includes((e.target as HTMLElement).tagName)) return;

      switch (e.key.toLowerCase()) {
        case "m":
          break; // Map focus placeholder
        case "a":
          toggleAlertPanel();
          break;
        case "h":
          toggleForensicsPanel();
          break;
        case "f":
          break; // Fullscreen map placeholder
        case "escape":
          closeDetailPanel();
          break;
      }

      if (e.ctrlKey && e.key.toLowerCase() === "f") {
        e.preventDefault();
        openSearch();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [toggleAlertPanel, toggleForensicsPanel, closeDetailPanel, openSearch]);

  // Start streaming
  useTrackStream();
  useAlertStream();

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        height: "100vh",
        backgroundColor: "#0F172A",
        color: "#F1F5F9",
        paddingTop: "28px",
        paddingBottom: "28px",
        boxSizing: "border-box",
      }}
    >
      <ClassificationBanner level={effectiveClassification} position="top" />
      <SearchOverlay />

      {/* Toolbar */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          padding: "4px 16px",
          backgroundColor: "#1E293B",
          borderBottom: "1px solid #334155",
          gap: "12px",
        }}
      >
        <span style={{ fontWeight: "bold", fontSize: "1rem", marginRight: "8px" }}>RTSA COP</span>
        <ConnectionIndicator />

        <RoleSelector />
        <DashboardSelector />

        <div style={{ flex: 1 }} />

        <div style={{ display: "flex", gap: "8px" }}>
          <button data-testid="toolbar-map" style={toolbarButtonStyle} onClick={closeDetailPanel}>🧭 Map</button>
          <button data-testid="toolbar-alerts" style={toolbarButtonStyle} onClick={toggleAlertPanel}>🚨 Alerts</button>
          <button data-testid="toolbar-history" style={toolbarButtonStyle} onClick={toggleForensicsPanel}>🔍 History</button>
          <button data-testid="toolbar-sensors" style={toolbarButtonStyle}>📡 Sensors</button>
          <button data-testid="toolbar-nato" style={toolbarButtonStyle}>🌐 NATO</button>
          <button data-testid="toolbar-audit" style={toolbarButtonStyle}>🔒 Audit</button>
        </div>

        <div style={{ width: "1px", height: "20px", backgroundColor: "#475569", margin: "0 8px" }} />

        <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
          <select
            data-testid="toolbar-settings"
            aria-label="Theme selector"
            value={theme}
            onChange={(e) => setTheme(e.target.value as Theme)}
            style={{
              padding: "4px 8px",
              backgroundColor: "#374151",
              color: "#F1F5F9",
              border: "1px solid #4B5563",
              borderRadius: "4px",
              cursor: "pointer",
              fontSize: "0.75rem",
            }}
          >
            <option value="dark">DARK</option>
            <option value="light">LIGHT</option>
            <option value="nvg">NVG</option>
          </select>

          <div data-testid="toolbar-profile" style={{ fontSize: "0.75rem", color: "#9CA3AF", display: "flex", flexDirection: "column", alignItems: "flex-end" }}>
            <span style={{ fontWeight: "bold", color: "#F1F5F9" }}>👤 {operatorName}</span>
            <span style={{ fontSize: "0.65rem" }}>{operatorClearance}</span>
          </div>
        </div>
      </div>

      {/* Main content switched by dashboard view */}
      <div style={{ display: "flex", flex: 1, overflow: "hidden", position: "relative" }}>

        {/* Placeholder routing logic for the Level 2 Dashboards */}
        {activeDashboardView === "operator" ? (
          <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center" }}>[Operator UI]</div>
        ) : activeDashboardView === "fusion" ? (
          <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center" }}>[Fusion Dashboard]</div>
        ) : activeDashboardView === "multi-domain" ? (
          <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center" }}>[Multi-Domain Dashboard]</div>
        ) : activeDashboardView === "audit" ? (
          <div data-testid="audit-view" style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center" }}>[Security] Audit & Feedback Queue View</div>
        ) : activeDashboardView === "forensics" ? (
          <div style={{ flex: 1, display: "flex" }}>
            <div style={{ flex: 1 }}><MapView /></div>
            <CollapsiblePane title="Intelligence Forensics" width="40%" height="100%" direction="horizontal">
              <ForensicsPanel />
            </CollapsiblePane>
          </div>
        ) : activeDashboardView === "sensor-health" ? (
          <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center" }}>[Sensor Health Dashboard]</div>
        ) : activeDashboardView === "nato-exchange" ? (
          <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center" }}>[NATO Exchange Dashboard]</div>
        ) : (
          /* Fallback generic layout (legacy-like) */
          <>
            <div style={{ flex: 1, display: "flex", flexDirection: "column", overflow: "hidden" }}>
              <div style={{ flex: 1, overflow: "hidden" }} aria-label="Map View" role="region" tabIndex={0}>
                <MapView />
              </div>
              {detailPanelOpen && (
                <CollapsiblePane title="Track Details" width="100%" height="280px" direction="vertical">
                  <DetailPanel />
                </CollapsiblePane>
              )}
              {forensicsPanelOpen && (
                <CollapsiblePane title="Forensics" width="100%" height="360px" direction="vertical">
                  <ForensicsPanel />
                </CollapsiblePane>
              )}
            </div>

            <CollapsiblePane title="Alerts & Notifications" width="30%" height="100%" direction="horizontal">
              <AlertPanel />
            </CollapsiblePane>
          </>
        )}
      </div>

      <SensorHealthPanel />
      <ClassificationBanner level={effectiveClassification} position="bottom" />
    </div>
  );
};
