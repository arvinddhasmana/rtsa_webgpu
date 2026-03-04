// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/MultiDomainDashboard.tsx
// Operations Commander — Multi-Domain fullscreen map with overlay metrics.

import React, { useMemo, useState } from "react";
import { useSensorHealth } from "../../hooks/useSensorHealth";
import { useAlertStore } from "../../stores/alertStore";
import { useUIStore } from "../../stores/uiStore";
import { AlertPanel } from "../alerts/AlertPanel";
import { DomainMetricsOverlay } from "../dashboard/DomainMetricsOverlay";
import { DetailPanel } from "../detail/DetailPanel";
import { MapView } from "../map/MapView";
import { CollapsiblePane } from "./CollapsiblePane";

export const MultiDomainDashboard: React.FC = () => {
  useSensorHealth();
  const detailPanelOpen = useUIStore((s) => s.detailPanelOpen);
  const toggleLayerVisibility = useUIStore((s) => s.toggleLayerVisibility);
  const layerVisibility = useUIStore((s) => s.layerVisibility);
  const alerts = useAlertStore((s) => s.alerts);

  const [alertStripExpanded, setAlertStripExpanded] = useState(false);

  const { criticalCount, unackCount } = useMemo(() => {
    const all = Array.from(alerts.values());
    return {
      criticalCount: all.filter((a) => a.severity === "CRITICAL").length,
      unackCount: all.length,
    };
  }, [alerts]);

  return (
    <div
      data-testid="multi-domain-dashboard"
      style={{
        flex: 1,
        display: "flex",
        overflow: "hidden",
        position: "relative",
      }}
    >
      {/* Full background map */}
      <div
        style={{
          position: "absolute",
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          zIndex: 0,
        }}
        aria-label="Map View"
        role="region"
        tabIndex={0}
      >
        <MapView />
      </div>

      {/* Domain KPIs overlay (top-left) */}
      <DomainMetricsOverlay />

      {/* Layer Toggles (top-right floating) */}
      <div
        style={{
          position: "absolute",
          top: "12px",
          right: "12px",
          display: "flex",
          flexDirection: "column",
          gap: "6px",
          zIndex: 5,
        }}
      >
        <LayerBtn
          label="🛰 Coverage"
          active={layerVisibility.sensorCoverage}
          onClick={() => toggleLayerVisibility("sensorCoverage")}
        />
        <LayerBtn
          label="🏷 Labels"
          active={layerVisibility.trackLabels}
          onClick={() => toggleLayerVisibility("trackLabels")}
        />
        <LayerBtn
          label="〰 Trails"
          active={layerVisibility.trackTrails}
          onClick={() => toggleLayerVisibility("trackTrails")}
        />
        <LayerBtn
          label="🗺 MGRS"
          active={layerVisibility.mgrsGrid}
          onClick={() => toggleLayerVisibility("mgrsGrid")}
        />
      </div>

      {/* Alert strip (bottom, collapsible) */}
      <div
        style={{
          position: "absolute",
          bottom: 0,
          left: 0,
          right: 0,
          zIndex: 10,
          transition: "height 0.25s ease",
        }}
      >
        {/* Strip header — always visible */}
        <div
          onClick={() => setAlertStripExpanded((v) => !v)}
          style={{
            display: "flex",
            alignItems: "center",
            gap: "10px",
            padding: "6px 16px",
            backgroundColor: "rgba(15, 23, 42, 0.9)",
            backdropFilter: "blur(8px)",
            borderTop: `1px solid ${criticalCount > 0 ? "#DC2626" : "#334155"}`,
            cursor: "pointer",
            userSelect: "none",
          }}
        >
          <span
            style={{
              fontSize: "0.7rem",
              fontWeight: "bold",
              color: criticalCount > 0 ? "#EF4444" : "#94A3B8",
              letterSpacing: "0.06em",
            }}
          >
            🚨 ALERTS
          </span>
          {unackCount > 0 && (
            <span
              style={{
                backgroundColor: criticalCount > 0 ? "#DC2626" : "#EA580C",
                color: "#fff",
                borderRadius: "9999px",
                padding: "1px 7px",
                fontSize: "0.65rem",
                fontWeight: "bold",
                animation: criticalCount > 0 ? "pulse 1.2s infinite" : undefined,
              }}
            >
              {unackCount}
            </span>
          )}
          <div style={{ flex: 1 }} />
          <span style={{ fontSize: "0.7rem", color: "#475569" }}>
            {alertStripExpanded ? "▼ Collapse" : "▲ Expand"}
          </span>
        </div>

        {/* Expanded alert panel */}
        {alertStripExpanded && (
          <div
            style={{
              height: "260px",
              backgroundColor: "rgba(15, 23, 42, 0.95)",
              backdropFilter: "blur(10px)",
              borderTop: "1px solid #334155",
            }}
          >
            <AlertPanel />
          </div>
        )}
      </div>

      {/* Detail Panel overlay */}
      {detailPanelOpen && (
        <div
          style={{
            position: "absolute",
            bottom: alertStripExpanded ? "292px" : "40px",
            left: 0,
            right: 0,
            zIndex: 10,
            padding: "0 16px",
            pointerEvents: "none",
            transition: "bottom 0.25s ease",
          }}
        >
          <div
            style={{ pointerEvents: "auto", margin: "0 auto", maxWidth: "1200px" }}
          >
            <CollapsiblePane
              title="Target Exploitation"
              width="100%"
              height="280px"
              direction="vertical"
            >
              <DetailPanel />
            </CollapsiblePane>
          </div>
        </div>
      )}
    </div>
  );
};

const LayerBtn: React.FC<{
  label: string;
  active: boolean;
  onClick: () => void;
}> = ({ label, active, onClick }) => (
  <button
    onClick={onClick}
    style={{
      padding: "5px 10px",
      backgroundColor: active ? "rgba(59, 130, 246, 0.2)" : "rgba(15, 23, 42, 0.75)",
      color: active ? "#60A5FA" : "#94A3B8",
      border: `1px solid ${active ? "#3B82F6" : "#334155"}`,
      borderRadius: "6px",
      cursor: "pointer",
      fontSize: "0.65rem",
      fontWeight: "bold",
      backdropFilter: "blur(6px)",
      whiteSpace: "nowrap",
      transition: "all 0.15s ease",
    }}
  >
    {label}
  </button>
);
