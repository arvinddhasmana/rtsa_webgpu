// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/OperatorDashboard.tsx
// Operations Commander — Operator UI Dashboard (Option A).
// 3-column glassmorphism layout over a blurred map background with resizable panes.

import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useAlertStore } from "../../stores/alertStore";
import { useTrackStore } from "../../stores/trackStore";
import { useUIStore } from "../../stores/uiStore";
import { AlertAssignPopover } from "../alert/AlertAssignPopover";
import { AlertPanel } from "../alerts/AlertPanel";
import { DetailPanel } from "../detail/DetailPanel";
import { MapView } from "../map/MapView";
import { TimelineScrubber } from "../timeline/TimelineScrubber";
import { TimelineView } from "../timeline/TimelineView";

const DEFAULT_LEFT_WIDTH = 320;
const DEFAULT_RIGHT_WIDTH = 380;
const MIN_PANEL_WIDTH = 250;
const MAX_PANEL_WIDTH = 600;

export const OperatorDashboard: React.FC = () => {
  const selectedTrackId = useTrackStore((s) => s.selectedTrackId);
  const detailPanelOpen = useUIStore((s) => s.detailPanelOpen);
  const openDetailPanel = useUIStore((s) => s.openDetailPanel);
  const alerts = useAlertStore((s) => s.alerts);
  const acknowledgedIds = useAlertStore((s) => s.acknowledgedIds);

  // Resizable pane states
  const [leftWidth, setLeftWidth] = useState(DEFAULT_LEFT_WIDTH);
  const [rightWidth, setRightWidth] = useState(DEFAULT_RIGHT_WIDTH);
  const containerRef = useRef<HTMLDivElement>(null);
  const isDraggingLeftRef = useRef(false);
  const isDraggingRightRef = useRef(false);

  // Auto-expand detail panel when a track is selected
  useEffect(() => {
    if (selectedTrackId && !detailPanelOpen) {
      openDetailPanel();
    }
  }, [selectedTrackId, detailPanelOpen, openDetailPanel]);

  // Assignment popover
  const [assignAlertId, setAssignAlertId] = useState<string | null>(null);

  // Expose setAssignAlertId globally for AlertCard to use if needed,
  // or pass down via context/props. For now we use the store pattern or global event.
  // We'll add a window listener for a custom event as a quick bridge if needed.
  useEffect(() => {
    const handleAssignEvent = (e: CustomEvent) => {
      setAssignAlertId(e.detail.alertId);
    };
    window.addEventListener("open-assign-popover" as any, handleAssignEvent);
    return () => window.removeEventListener("open-assign-popover" as any, handleAssignEvent);
  }, []);

  // Timeline Scrubber
  const now = Date.now();
  const START_MS = useMemo(() => now - 6 * 60 * 60 * 1000, []);
  const [scrubberOpen, setScrubberOpen] = useState(false);
  const [replayMs, setReplayMs] = useState(now);
  const [isPlaying, setIsPlaying] = useState(false);
  const [playbackSpeed, setPlaybackSpeed] = useState(1);

  const unacknowledgedCount = useMemo(() => {
    return Array.from(alerts.values()).filter(a => !acknowledgedIds.has(a.alertId)).length;
  }, [alerts, acknowledgedIds]);

  const handleAssign = (operatorId: string) => {
    console.log(`[OPERATOR] Alert ${assignAlertId} → assigned to ${operatorId}`);
    setAssignAlertId(null);
  };

  // Drag resize handlers
  const onMouseDownLeft = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    isDraggingLeftRef.current = true;
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";

    const onMouseMove = (ev: MouseEvent) => {
      if (!isDraggingLeftRef.current || !containerRef.current) return;
      const rect = containerRef.current.getBoundingClientRect();
      const newWidth = ev.clientX - rect.left;
      setLeftWidth(Math.max(MIN_PANEL_WIDTH, Math.min(MAX_PANEL_WIDTH, newWidth)));
    };

    const onMouseUp = () => {
      isDraggingLeftRef.current = false;
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      document.removeEventListener("mousemove", onMouseMove);
      document.removeEventListener("mouseup", onMouseUp);
    };

    document.addEventListener("mousemove", onMouseMove);
    document.addEventListener("mouseup", onMouseUp);
  }, []);

  const onMouseDownRight = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    isDraggingRightRef.current = true;
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";

    const onMouseMove = (ev: MouseEvent) => {
      if (!isDraggingRightRef.current || !containerRef.current) return;
      const rect = containerRef.current.getBoundingClientRect();
      const newWidth = rect.right - ev.clientX;
      setRightWidth(Math.max(MIN_PANEL_WIDTH, Math.min(MAX_PANEL_WIDTH, newWidth)));
    };

    const onMouseUp = () => {
      isDraggingRightRef.current = false;
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      document.removeEventListener("mousemove", onMouseMove);
      document.removeEventListener("mouseup", onMouseUp);
    };

    document.addEventListener("mousemove", onMouseMove);
    document.addEventListener("mouseup", onMouseUp);
  }, []);

  return (
    <div
      ref={containerRef}
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
            backgroundColor: "rgba(10, 15, 30, 0.65)", // Option A dark tint
            backdropFilter: "blur(6px)",
          }}
        />
      </div>

      {/* ── Top Header Strip (Option A specific) ─────────── */}
      <div
        style={{
          position: "absolute",
          top: 12,
          left: 0,
          right: 0,
          zIndex: 2,
          display: "flex",
          justifyContent: "center",
          pointerEvents: "none",
        }}
      >
        {unacknowledgedCount > 0 && (
          <div
            style={{
              backgroundColor: "rgba(15, 23, 42, 0.8)",
              border: "1px solid rgba(245, 158, 11, 0.5)",
              color: "#F59E0B",
              padding: "6px 16px",
              borderRadius: "20px",
              fontSize: "0.8rem",
              fontWeight: "bold",
              backdropFilter: "blur(8px)",
              boxShadow: "0 0 15px rgba(245, 158, 11, 0.2)",
              pointerEvents: "auto",
              letterSpacing: "0.05em",
            }}
          >
            {unacknowledgedCount} UNACKNOWLEDGED ALERTS
          </div>
        )}
      </div>

      {/* ── Floating Interface Layer ──────────────────────── */}
      <div
        style={{
          display: "flex",
          flex: 1,
          zIndex: 1,
          padding: "16px",
          gap: "8px",
          overflow: "hidden",
        }}
      >
        {/* ── Left: Entity Timeline ───────────────────────── */}
        <div
          style={{
            width: `${leftWidth}px`,
            flexShrink: 0,
            display: "flex",
            flexDirection: "column",
            ...glassPanelStyle,
          }}
        >
          <div style={panelHeaderStyle}>ENTITY TIMELINE</div>
          <div style={{ flex: 1, overflow: "hidden" }}>
            <TimelineView />
          </div>
        </div>

        {/* ── Left Resize Handle ──────────────────────────── */}
        <div
          className="ds-resize-handle"
          onMouseDown={onMouseDownLeft}
          style={resizeHandleStyle}
        />

        {/* ── Center: Map Area & Entity Detail ────────────── */}
        <div
          style={{
            flex: 1,
            display: "flex",
            flexDirection: "column",
            justifyContent: "flex-end",
            pointerEvents: "auto", // allow clicking through
            minWidth: 0,
            position: "relative",
          }}
        >
          {/* Replay scrubber toggle (floating bottom-right of center area) */}
          <div
            style={{
              position: "absolute",
              bottom: detailPanelOpen ? "340px" : "16px",
              right: "0px",
              transition: "bottom 0.3s ease",
            }}
          >
            <button
              onClick={() => setScrubberOpen((v) => !v)}
              style={{
                padding: "8px 16px",
                backgroundColor: scrubberOpen
                  ? "rgba(59, 130, 246, 0.2)"
                  : "rgba(15, 23, 42, 0.75)",
                color: scrubberOpen ? "#60A5FA" : "#94A3B8",
                border: `1px solid ${scrubberOpen ? "#3B82F6" : "rgba(255,255,255,0.1)"}`,
                borderRadius: "6px",
                cursor: "pointer",
                fontSize: "0.75rem",
                fontWeight: "bold",
                backdropFilter: "blur(6px)",
                display: "flex",
                alignItems: "center",
                gap: "6px",
              }}
            >
              {scrubberOpen ? (
                <><span>⏹</span> Back to Live</>
              ) : (
                <><span>▶</span> Replay Mode</>
              )}
            </button>
          </div>

          {/* Expandable Entity Detail */}
          <div
            style={{
              height: detailPanelOpen ? "320px" : "0px",
              opacity: detailPanelOpen ? 1 : 0,
              overflow: "hidden",
              transition: "all 0.3s cubic-bezier(0.4, 0, 0.2, 1)",
              ...glassPanelStyle,
              marginBottom: detailPanelOpen ? "0" : "-16px",
              display: "flex",
              flexDirection: "column",
            }}
          >
            <div style={{ flex: 1, overflow: "hidden" }}>
              {detailPanelOpen && <DetailPanel />}
            </div>
          </div>
        </div>

        {/* ── Right Resize Handle ─────────────────────────── */}
        <div
          className="ds-resize-handle"
          onMouseDown={onMouseDownRight}
          style={resizeHandleStyle}
        />

        {/* ── Right: Alert Triage Queue ────────────────────── */}
        <div
          style={{
            width: `${rightWidth}px`,
            flexShrink: 0,
            display: "flex",
            flexDirection: "column",
            position: "relative",
            ...glassPanelStyle,
          }}
        >
          <div style={{ flex: 1, overflow: "hidden" }}>
            <AlertPanel />
          </div>
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

// -- Reusable Styles for Option A --
const glassPanelStyle: React.CSSProperties = {
  backgroundColor: "rgba(15, 23, 42, 0.7)",
  backdropFilter: "blur(12px)",
  border: "1px solid rgba(255, 255, 255, 0.08)",
  borderRadius: "8px",
  boxShadow: "0 8px 32px rgba(0, 0, 0, 0.5)",
};

const panelHeaderStyle: React.CSSProperties = {
  padding: "12px 16px",
  borderBottom: "1px solid rgba(255, 255, 255, 0.08)",
  fontSize: "0.85rem",
  fontWeight: "bold",
  color: "#F1F5F9",
  letterSpacing: "0.05em",
};

const resizeHandleStyle: React.CSSProperties = {
  width: "8px",
  cursor: "col-resize",
  zIndex: 10,
  margin: "0 -4px", // overlap slightly for easier grabbing
};
