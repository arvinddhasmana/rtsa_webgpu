// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/FusionDashboard.tsx
// Operations Commander — Fusion Dashboard (default Level-2 view).

import React, { useMemo, useState } from "react";
import { useSensorStream } from "../../hooks/useSensorStream";
import { useUIStore } from "../../stores/uiStore";
import { AlertAssignPopover } from "../alert/AlertAssignPopover";
import { DetailPanel } from "../detail/DetailPanel";
import { FusionSidePanel } from "../fusion/FusionSidePanel";
import { MapView } from "../map/MapView";
import { TimelineScrubber } from "../timeline/TimelineScrubber";
import { CollapsiblePane } from "./CollapsiblePane";

export const FusionDashboard: React.FC = () => {
  const detailPanelOpen = useUIStore((s) => s.detailPanelOpen);
  const toggleLayerVisibility = useUIStore((s) => s.toggleLayerVisibility);
  const layerVisibility = useUIStore((s) => s.layerVisibility);

  // Raw sensor stream active only on this dashboard
  useSensorStream();

  // Timeline scrubber state (local — commander-only)
  const now = Date.now();
  const [scrubberOpen, setScrubberOpen] = useState(false);
  const [replayMs, setReplayMs] = useState(now);
  const [isPlaying, setIsPlaying] = useState(false);
  const [playbackSpeed, setPlaybackSpeed] = useState(1);
  const START_MS = useMemo(() => now - 6 * 60 * 60 * 1000, []);  // 6h window

  // Alert assignment popover state
  const [assignAlertId, setAssignAlertId] = useState<string | null>(null);

  const handleAssign = (operatorId: string) => {
    console.log(`[FUSION] Alert ${assignAlertId} → assigned to ${operatorId}`);
    setAssignAlertId(null);
  };

  return (
    <div
      data-testid="fusion-dashboard"
      style={{ flex: 1, display: "flex", overflow: "hidden", position: "relative" }}
    >
      {/* ── Left: Collapsible Fusion Side Panel ─────────── */}
      <CollapsiblePane
        title="Multi-Sensor Fusion"
        width="320px"
        height="100%"
        direction="horizontal"
      >
        <FusionSidePanel />
      </CollapsiblePane>

      {/* ── Main area: Map + controls + bottom detail panel */}
      <div
        style={{
          flex: 1,
          display: "flex",
          flexDirection: "column",
          overflow: "hidden",
          position: "relative",
        }}
      >
        {/* Map */}
        <div
          style={{ flex: 1, overflow: "hidden" }}
          aria-label="Map View"
          role="region"
          tabIndex={0}
        >
          <MapView />
        </div>

        {/* Floating top-right controls */}
        <div
          style={{
            position: "absolute",
            top: "12px",
            right: "12px",
            display: "flex",
            flexDirection: "column",
            gap: "6px",
            zIndex: 10,
          }}
        >
          {/* Layer toggles */}
          <LayerButton
            label="🛰 Sensors"
            active={layerVisibility.sensorCoverage}
            onClick={() => toggleLayerVisibility("sensorCoverage")}
          />
          <LayerButton
            label="🏷 Labels"
            active={layerVisibility.trackLabels}
            onClick={() => toggleLayerVisibility("trackLabels")}
          />
          <LayerButton
            label="〰 Trails"
            active={layerVisibility.trackTrails}
            onClick={() => toggleLayerVisibility("trackTrails")}
          />
          <LayerButton
            label={scrubberOpen ? "⏹ Live" : "⏮ Replay"}
            active={scrubberOpen}
            onClick={() => {
              setScrubberOpen((v) => !v);
              setIsPlaying(false);
              setReplayMs(Date.now());
            }}
            accent={scrubberOpen ? "#3B82F6" : undefined}
          />
        </div>

        {/* Timeline Scrubber */}
        {scrubberOpen && (
          <TimelineScrubber
            startMs={START_MS}
            endMs={Date.now()}
            currentMs={replayMs}
            isPlaying={isPlaying}
            speed={playbackSpeed}
            onSeek={setReplayMs}
            onPlay={() => setIsPlaying(true)}
            onPause={() => setIsPlaying(false)}
            onSpeedChange={setPlaybackSpeed}
            onClose={() => {
              setScrubberOpen(false);
              setIsPlaying(false);
            }}
          />
        )}

        {/* Bottom Detail Panel */}
        {detailPanelOpen && (
          <CollapsiblePane
            title="Track Details"
            width="100%"
            height="280px"
            direction="vertical"
          >
            <DetailPanel />
          </CollapsiblePane>
        )}
      </div>

      {/* Alert Assign Popover */}
      {assignAlertId && (
        <AlertAssignPopover
          alertId={assignAlertId}
          onAssign={handleAssign}
          onClose={() => setAssignAlertId(null)}
        />
      )}
    </div>
  );
};

const LayerButton: React.FC<{
  label: string;
  active: boolean;
  onClick: () => void;
  accent?: string;
}> = ({ label, active, onClick, accent }) => (
  <button
    onClick={onClick}
    style={{
      padding: "5px 10px",
      backgroundColor: active
        ? (accent ? `${accent}33` : "rgba(59, 130, 246, 0.2)")
        : "rgba(15, 23, 42, 0.75)",
      color: active ? (accent ?? "#60A5FA") : "#94A3B8",
      border: `1px solid ${active ? (accent ?? "#3B82F6") : "#334155"}`,
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
