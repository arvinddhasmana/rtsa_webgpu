// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/FusionDashboard.tsx
// Operations Commander — Fusion Dashboard (Variant C split-pane layout).

import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useSensorStream } from "../../hooks/useSensorStream";
import { AlertAssignPopover } from "../alert/AlertAssignPopover";
import { FusionSidePanel } from "../fusion/FusionSidePanel";
import { MapView } from "../map/MapView";
import { TimelineScrubber } from "../timeline/TimelineScrubber";

const DEFAULT_LEFT_PERCENT = 35;
const MIN_LEFT_PERCENT = 20;
const MAX_LEFT_PERCENT = 65;

export const FusionDashboard: React.FC = () => {
  // Raw sensor stream active only on this dashboard
  useSensorStream();

  // Resizable pane state
  const [leftPercent, setLeftPercent] = useState(DEFAULT_LEFT_PERCENT);
  const containerRef = useRef<HTMLDivElement>(null);
  const isDraggingRef = useRef(false);

  // Timeline scrubber state (local — commander-only)
  const now = Date.now();
  const [scrubberOpen, setScrubberOpen] = useState(false);
  const [replayMs, setReplayMs] = useState(now);
  const [isPlaying, setIsPlaying] = useState(false);
  const [playbackSpeed, setPlaybackSpeed] = useState(1);
  const START_MS = useMemo(() => now - 6 * 60 * 60 * 1000, []);

  // Alert assignment popover state
  const [assignAlertId, setAssignAlertId] = useState<string | null>(null);

  const handleAssign = (operatorId: string) => {
    console.log(`[FUSION] Alert ${assignAlertId} → assigned to ${operatorId}`);
    setAssignAlertId(null);
  };

  // Escape to reset layout
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        setLeftPercent(DEFAULT_LEFT_PERCENT);
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  // Drag resize handler
  const onMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    isDraggingRef.current = true;
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";

    const onMouseMove = (ev: MouseEvent) => {
      if (!isDraggingRef.current || !containerRef.current) return;
      const rect = containerRef.current.getBoundingClientRect();
      const pct = ((ev.clientX - rect.left) / rect.width) * 100;
      setLeftPercent(Math.max(MIN_LEFT_PERCENT, Math.min(MAX_LEFT_PERCENT, pct)));
    };

    const onMouseUp = () => {
      isDraggingRef.current = false;
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
      data-testid="fusion-dashboard"
      style={{
        flex: 1,
        display: "flex",
        overflow: "hidden",
        position: "relative",
      }}
    >
      {/* ── Left Panel (Fusion Side Panel) ────────────── */}
      <div
        style={{
          width: `${leftPercent}%`,
          minWidth: "280px",
          display: "flex",
          flexDirection: "column",
          background: "var(--ds-bg-primary, #0F172A)",
          borderRight: "1px solid var(--ds-border-default, #334155)",
          overflow: "hidden",
        }}
      >
        <FusionSidePanel
          onOpenScrubber={() => setScrubberOpen(true)}
          scrubberOpen={scrubberOpen}
        />
      </div>

      {/* ── Resize Handle ─────────────────────────────── */}
      <div
        className="ds-resize-handle"
        onMouseDown={onMouseDown}
        role="separator"
        aria-label="Resize panels"
        tabIndex={0}
        data-testid="fusion-resize-handle"
      />

      {/* ── Right Panel (Map) ─────────────────────────── */}
      <div
        style={{
          flex: 1,
          overflow: "hidden",
          position: "relative",
          display: "flex",
          flexDirection: "column",
        }}
        aria-label="Map View"
        role="region"
        tabIndex={0}
      >
        <MapView />

        {/* Timeline Scrubber at bottom of map */}
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
