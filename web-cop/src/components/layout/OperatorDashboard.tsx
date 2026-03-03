// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/OperatorDashboard.tsx
// Operations Commander — Operator UI Dashboard.
// Blurred map + glassmorphism panels: Timeline (left), Entity Detail (center), Alerts (right).

import React, { useMemo, useState } from "react";
import { useAlertStore } from "../../stores/alertStore";
import { useUIStore } from "../../stores/uiStore";
import { AlertAssignPopover } from "../alert/AlertAssignPopover";
import { AlertPanel } from "../alerts/AlertPanel";
import { DetailPanel } from "../detail/DetailPanel";
import { MapView } from "../map/MapView";
import { TimelineScrubber } from "../timeline/TimelineScrubber";
import { TimelineView } from "../timeline/TimelineView";
import { CollapsiblePane } from "./CollapsiblePane";

export const OperatorDashboard: React.FC = () => {
  const detailPanelOpen = useUIStore((s) => s.detailPanelOpen);
  const alerts = useAlertStore((s) => s.alerts);

  // Assignment popover
  const [assignAlertId, setAssignAlertId] = useState<string | null>(null);

  // Timeline Scrubber (Operator UI has its own instance)
  const now = Date.now();
  const START_MS = useMemo(() => now - 6 * 60 * 60 * 1000, []);
  const [scrubberOpen, setScrubberOpen] = useState(false);
  const [replayMs, setReplayMs] = useState(now);
  const [isPlaying, setIsPlaying] = useState(false);
  const [playbackSpeed, setPlaybackSpeed] = useState(1);

  const criticalAlerts = useMemo(
    () => Array.from(alerts.values()).filter((a) => a.severity === "CRITICAL").length,
    [alerts]
  );

  const handleAssign = (operatorId: string) => {
    console.log(`[OPERATOR] Alert ${assignAlertId} → assigned to ${operatorId}`);
    setAssignAlertId(null);
  };

  return (
    <div
      data-testid="operator-dashboard"
      style={{
        flex: 1,
        display: "flex",
        overflow: "hidden",
        position: "relative",
      }}
    >
      {/* ── Background: blurred map ──────────────────────── */}
      <div
        style={{
          position: "absolute",
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          zIndex: 0,
        }}
      >
        <MapView />
        <div
          style={{
            position: "absolute",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            backgroundColor: "rgba(15, 23, 42, 0.62)",
            backdropFilter: "blur(4px)",
          }}
        />
      </div>

      {/* ── Floating Interface Layer ──────────────────────── */}
      <div
        style={{
          display: "flex",
          flex: 1,
          zIndex: 1,
          padding: "12px",
          gap: "12px",
          overflow: "hidden",
        }}
      >
        {/* ── Left: Entity Timeline ───────────────────────── */}
        <div
          style={{
            width: "320px",
            flexShrink: 0,
            display: "flex",
            flexDirection: "column",
          }}
        >
          <CollapsiblePane
            title="Entity Timeline"
            width="100%"
            height="100%"
            direction="horizontal"
          >
            <TimelineView />
          </CollapsiblePane>
        </div>

        {/* ── Center: Entity Detail (when track selected) ── */}
        <div
          style={{
            flex: 1,
            display: "flex",
            flexDirection: "column",
            justifyContent: "flex-end",
            pointerEvents: "auto",
            minWidth: 0,
          }}
        >
          {/* Replay scrubber toggle bar */}
          <div
            style={{
              display: "flex",
              justifyContent: "center",
              marginBottom: "8px",
            }}
          >
            <button
              onClick={() => setScrubberOpen((v) => !v)}
              style={{
                padding: "5px 16px",
                backgroundColor: scrubberOpen
                  ? "rgba(59, 130, 246, 0.2)"
                  : "rgba(15, 23, 42, 0.75)",
                color: scrubberOpen ? "#60A5FA" : "#94A3B8",
                border: `1px solid ${scrubberOpen ? "#3B82F6" : "#334155"}`,
                borderRadius: "6px",
                cursor: "pointer",
                fontSize: "0.65rem",
                fontWeight: "bold",
                backdropFilter: "blur(6px)",
              }}
            >
              {scrubberOpen ? "⏹ Back to Live" : "⏮ Replay Mode"}
            </button>
          </div>

          {detailPanelOpen && (
            <CollapsiblePane
              title="Entity Detail"
              width="100%"
              height="320px"
              direction="vertical"
            >
              <DetailPanel />
            </CollapsiblePane>
          )}
        </div>

        {/* ── Right: Alert Triage Queue ────────────────────── */}
        <div
          style={{
            width: "380px",
            flexShrink: 0,
            display: "flex",
            flexDirection: "column",
            position: "relative",
          }}
        >
          {/* Critical badge */}
          {criticalAlerts > 0 && (
            <div
              style={{
                position: "absolute",
                top: "8px",
                right: "8px",
                backgroundColor: "#DC2626",
                color: "#fff",
                borderRadius: "9999px",
                padding: "2px 8px",
                fontSize: "0.65rem",
                fontWeight: "bold",
                zIndex: 5,
                animation: "pulse 1.2s infinite",
              }}
            >
              {criticalAlerts} CRITICAL
            </div>
          )}
          <CollapsiblePane
            title="Alert Triage"
            width="100%"
            height="100%"
            direction="horizontal"
          >
            <AlertPanel />
          </CollapsiblePane>
        </div>
      </div>

      {/* Timeline Scrubber (sits above status bar) */}
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
