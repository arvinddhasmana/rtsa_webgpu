// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/MainLayout.tsx

import React from "react";
import { ClassificationBanner } from "./ClassificationBanner";
import { ConnectionIndicator } from "./ConnectionIndicator";
import { SensorHealthPanel } from "./SensorHealthPanel";
import { MapView } from "../map/MapView";
import { AlertPanel } from "../alerts/AlertPanel";
import { DetailPanel } from "../detail/DetailPanel";
import { ForensicsPanel } from "../forensics/ForensicsPanel";
import { useUIStore, type Theme } from "../../stores/uiStore";
import { useClassification } from "../../hooks/useClassification";
import { useTrackStream } from "../../hooks/useTrackStream";
import { useAlertStream } from "../../hooks/useAlertStream";

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
  const theme = useUIStore((s) => s.theme);
  const setTheme = useUIStore((s) => s.setTheme);

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

      {/* Toolbar */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          padding: "4px 16px",
          backgroundColor: "#1E293B",
          borderBottom: "1px solid #334155",
          gap: "16px",
        }}
      >
        <span style={{ fontWeight: "bold", fontSize: "1rem" }}>RTSA COP</span>
        <ConnectionIndicator />
        <div style={{ flex: 1 }} />
        <button
          onClick={toggleForensicsPanel}
          style={{
            padding: "4px 12px",
            backgroundColor: forensicsPanelOpen ? "#1D4ED8" : "#374151",
            color: "#F1F5F9",
            border: "none",
            borderRadius: "4px",
            cursor: "pointer",
            fontSize: "0.75rem",
          }}
        >
          FORENSICS
        </button>
        <select
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
      </div>

      {/* Main content */}
      <div style={{ display: "flex", flex: 1, overflow: "hidden" }}>
        {/* Map + detail area */}
        <div
          style={{
            flex: 1,
            display: "flex",
            flexDirection: "column",
            overflow: "hidden",
          }}
        >
          <div style={{ flex: 1, overflow: "hidden" }}>
            <MapView />
          </div>
          {detailPanelOpen && (
            <div
              style={{
                height: "280px",
                backgroundColor: "#1E293B",
                borderTop: "1px solid #334155",
                overflow: "auto",
              }}
            >
              <DetailPanel />
            </div>
          )}
          {forensicsPanelOpen && (
            <div
              style={{
                height: "360px",
                backgroundColor: "#1E293B",
                borderTop: "1px solid #334155",
                overflow: "auto",
              }}
            >
              <ForensicsPanel />
            </div>
          )}
        </div>

        {/* Alert panel (30% width) */}
        <div
          style={{
            width: "30%",
            minWidth: "280px",
            maxWidth: "420px",
            backgroundColor: "#1E293B",
            borderLeft: "1px solid #334155",
            overflow: "hidden",
          }}
        >
          <AlertPanel />
        </div>
      </div>

      <SensorHealthPanel />
      <ClassificationBanner level={effectiveClassification} position="bottom" />
    </div>
  );
};
