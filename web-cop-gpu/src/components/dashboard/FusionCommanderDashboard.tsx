// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/FusionCommanderDashboard.tsx

import { createEffect, createSignal, For, JSX, onCleanup, Show } from "solid-js";
import { clearSelectedTrack, openTrackDetails, setOpenTrackDetails, setTrackOverlayPositions, trackDetail, trackOverlayPositions } from "../../signals/track";
import { setFeedbackOpen } from "../../signals/viewport";
import { DraggableOverlayCard } from "./DraggableOverlayCard";
import { FusionConflictPanel } from "./FusionConflictPanel";
import { FusionKPIDashboard } from "./FusionKPIDashboard";
import { SensorLegend } from "./SensorLegend";
import { TrackDrillDownOverlay } from "./TrackDrillDownOverlay";
import { TrackFusionAudit } from "./TrackFusionAudit";

interface FusionCommanderDashboardProps {
  mapContent: JSX.Element;
  sidePanelContent?: JSX.Element;
}

/**
 * Operations Commander Fusion Dashboard (Redesigned from Scratch)
 * Focus: Tactical Clarity, Premium Glassmorphism, Zero Clutter.
 */
export function FusionCommanderDashboard(props: FusionCommanderDashboardProps) {
  const [showObs, setShowObs] = createSignal(true);
  const [showTracks, setShowTracks] = createSignal(true);
  const [mapStyle, setMapStyle] = createSignal(0);
  const [legendPos, setLegendPos] = createSignal({ x: 24, y: window.innerHeight - 300 });

  function closeTrackOverlay(trackId: string) {
    setOpenTrackDetails((curr) => curr.filter((t) => t.trackId !== trackId));
    if (trackDetail()?.trackId === trackId) {
       clearSelectedTrack();
    }
  }

  function closeAllTrackOverlays() {
    setOpenTrackDetails([]);
    clearSelectedTrack();
  }

  function updateTrackOverlayPosition(trackId: string, pos: { x: number; y: number }) {
    setTrackOverlayPositions((prev) => ({ ...prev, [trackId]: pos }));
  }

  function autoArrangeTrackOverlays() {
    const root = document.querySelector('[data-testid="fusion-dashboard-root"]');
    if (!root) return;
    const rootR = root.getBoundingClientRect();
    const w = rootR.width;
    const items = openTrackDetails();
    const gap = 20;
    const cardW = 480;
    const cardH = 340;
    const columns = Math.max(1, Math.floor((w - 40) / (cardW + gap)));
    const arranged: Record<string, { x: number; y: number }> = {};
    items.forEach((track, idx) => {
      const col = idx % columns;
      const row = Math.floor(idx / columns);
      arranged[track.trackId] = {
        x: 20 + col * (cardW + gap),
        y: 60 + row * (cardH + gap),
      };
    });
    setTrackOverlayPositions({ ...arranged });
  }

  createEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") {
        setFeedbackOpen(false);
        if (openTrackDetails().length > 0) {
          e.preventDefault();
          closeAllTrackOverlays();
        }
      }
    }
    window.addEventListener("keydown", onKeyDown);
    onCleanup(() => window.removeEventListener("keydown", onKeyDown));
  });

  return (
    <div
      data-testid="fusion-dashboard-root"
      style={{
        display: "flex",
        width: "100%",
        height: "100%",
        background: "#020617", // Deep Navy/Black
        color: "#f8fafc",
        overflow: "hidden",
        "font-family": "Inter, sans-serif",
      }}
    >
      {/* Tactical Map Region */}
      <section
        style={{
          flex: "1",
          position: "relative",
          "min-width": "0",
          overflow: "hidden",
        }}
      >
        <div data-testid="commander-fusion-map-container" style={{ width: "100%", height: "100%", position: "relative" }}>
          {props.mapContent}
          <div data-testid="commander-observation-layer-mount" />
          <div data-testid="commander-fused-layer-mount" />
        </div>

        {/* Floating Top Controls (Glassmorphic) */}
        <div
          style={{
            position: "absolute",
            top: "1.5rem",
            left: "1.5rem",
            right: "1.5rem",
            display: "flex",
            "justify-content": "space-between",
            "pointer-events": "none",
            "z-index": "50",
          }}
        >
          <div style={{ display: "flex", gap: "1rem", "pointer-events": "auto" }}>
            <div
              style={{
                background: "rgba(15, 23, 42, 0.6)",
                "backdrop-filter": "blur(8px)",
                padding: "0.5rem 1rem",
                "border-radius": "8px",
                border: "1px solid rgba(255, 255, 255, 0.1)",
                display: "flex",
                "align-items": "center",
                gap: "1.5rem",
              }}
            >
              <ControlToggle label="RAW OBS" active={showObs()} onClick={() => setShowObs(!showObs())} color="#38bdf8" />
              <ControlToggle label="FUSED TRACKS" active={showTracks()} onClick={() => setShowTracks(!showTracks())} color="#fbbf24" />
            </div>

            <div
              style={{
                background: "rgba(15, 23, 42, 0.6)",
                "backdrop-filter": "blur(8px)",
                padding: "0.5rem",
                "border-radius": "8px",
                border: "1px solid rgba(255, 255, 255, 0.1)",
              }}
            >
              <button
                onClick={() => setMapStyle(mapStyle() === 0 ? 1 : 0)}
                style={{
                  background: mapStyle() === 1 ? "rgba(16, 185, 129, 0.2)" : "transparent",
                  color: mapStyle() === 1 ? "#10b981" : "#94a3b8",
                  border: `1px solid ${mapStyle() === 1 ? "#10b981" : "rgba(255,255,255,0.1)"}`,
                  padding: "4px 12px",
                  "border-radius": "4px",
                  "font-size": "0.7rem",
                  "font-weight": "700",
                  cursor: "pointer",
                  transition: "all 0.2s",
                }}
              >
                {mapStyle() === 1 ? "HD MAP ACTIVE" : "STANDARD MAP"}
              </button>
            </div>
          </div>
        </div>

        {/* Legend Panel (Floating) */}
        <DraggableOverlayCard
          title="PROFESSIONAL GLASS LEGEND"
          position={legendPos()}
          onPositionChange={setLegendPos}
          minWidth="280px"
          constrainToParent
        >
          <SensorLegend />
        </DraggableOverlayCard>

        {/* Track Drill-down (Anchored Overlay) */}
        <Show when={openTrackDetails().length > 0}>
          <div
            data-testid="track-overlay-layer"
            style={{
              position: "absolute",
              inset: "0",
              "pointer-events": "none",
              "z-index": 120,
            }}
          >
            <button
              data-testid="track-auto-arrange"
              onClick={() => autoArrangeTrackOverlays()}
              title="Auto-arrange track cards"
              style={{
                position: "absolute",
                top: "8px",
                right: "8px",
                background: "rgba(255,255,255,0.06)",
                border: "1px solid rgba(255,255,255,0.14)",
                color: "#94a3b8",
                padding: "4px 8px",
                "border-radius": "8px",
                "font-size": "0.6rem",
                "font-weight": "700",
                "text-transform": "uppercase",
                "letter-spacing": "0.1em",
                cursor: "pointer",
                "pointer-events": "auto",
                transition: "all 0.15s ease",
                "box-shadow": "0 2px 8px rgba(0,0,0,0.4)",
              }}
              onMouseEnter={(e) => {
                const t = e.currentTarget;
                t.style.background = "rgba(255,255,255,0.12)";
                t.style.color = "#f8fafc";
              }}
              onMouseLeave={(e) => {
                const t = e.currentTarget;
                t.style.background = "rgba(255,255,255,0.06)";
                t.style.color = "#94a3b8";
              }}
            >
              Auto Arrange
            </button>
            <For each={openTrackDetails()}>
              {(detail) => (
                <DraggableOverlayCard
                  title={`FUSED TRACK DATA | TRACK ID: ${detail.trackId}`}
                  position={trackOverlayPositions()[detail.trackId] ?? { x: window.innerWidth / 2 - 250, y: window.innerHeight / 2 - 275 }}
                  onPositionChange={(pos) => updateTrackOverlayPosition(detail.trackId, pos)}
                  onClose={() => closeTrackOverlay(detail.trackId)}
                  width="450px"
                  minWidth="340px"
                  maxHeight="800px"
                  resizable={true}
                  constrainToParent
                  accentColor="#38bdf8"
                >
                  <TrackDrillDownOverlay track={detail} onClose={() => closeTrackOverlay(detail.trackId)} />
                </DraggableOverlayCard>
              )}
            </For>
          </div>
        </Show>
      </section>

      {/* Mission Analytics Fragment (Floating Draggable Toast) */}
      <MissionAnalyticsToast />
    </div>
  );
}

function MissionAnalyticsToast() {
  const [pos, setPos] = createSignal({ x: window.innerWidth - 420, y: 80 });
  const [minimized, setMinimized] = createSignal(false);
  const [isDragging, setIsDragging] = createSignal(false);
  let dragOffset = { x: 0, y: 0 };

  const onMouseDown = (e: MouseEvent) => {
    setIsDragging(true);
    dragOffset = {
      x: e.clientX - pos().x,
      y: e.clientY - pos().y
    };
    window.addEventListener("mousemove", onMouseMove);
    window.addEventListener("mouseup", onMouseUp);
  };

  const onMouseMove = (e: MouseEvent) => {
    if (!isDragging()) return;
    setPos({
      x: e.clientX - dragOffset.x,
      y: e.clientY - dragOffset.y
    });
  };

  const onMouseUp = () => {
    setIsDragging(false);
    window.removeEventListener("mousemove", onMouseMove);
    window.removeEventListener("mouseup", onMouseUp);
  };

  return (
    <div
      data-testid="commander-fusion-side-panel"
      style={{
        position: "absolute",
        left: `${pos().x}px`,
        top: `${pos().y}px`,
        width: "410px",
        "max-height": minimized() ? "48px" : "calc(100% - 100px)",
        background: "rgba(7, 12, 24, 0.8)",
        "backdrop-filter": "blur(30px) saturate(200%)",
        border: "1px solid rgba(255, 255, 255, 0.1)",
        "box-shadow": "0 25px 50px -12px rgba(0, 0, 0, 0.8), inset 0 1px 1px rgba(255,255,255,0.05)",
        "border-radius": "16px",
        display: "flex",
        "flex-direction": "column",
        "z-index": "100",
        overflow: "hidden",
        transition: "max-height 0.4s cubic-bezier(0.16, 1, 0.3, 1)",
      }}
    >
      {/* Header / Drag Handle - Mockup Matched */}
      <div
        onMouseDown={onMouseDown}
        style={{
          padding: "1rem 1.5rem",
          display: "flex",
          "justify-content": "space-between",
          "align-items": "center",
          cursor: "grab",
          background: "linear-gradient(90deg, rgba(30, 41, 59, 0.6) 0%, transparent 100%)",
          "border-bottom": minimized() ? "none" : "1px solid rgba(255,255,255,0.05)",
        }}
      >
        <span style={{
          "font-size": "0.75rem",
          "font-weight": "900",
          color: "#38bdf8",
          "letter-spacing": "0.15em",
          "text-transform": "uppercase",
          "text-shadow": "0 0 10px rgba(56, 189, 248, 0.3)"
        }}>
          MISSION ANALYTICS
        </span>
        <button
          onClick={() => setMinimized(!minimized())}
          style={{
            background: "rgba(255,255,255,0.08)",
            border: "1px solid rgba(255,255,255,0.1)",
            color: "#e2e8f0",
            width: "28px",
            height: "28px",
            "border-radius": "6px",
            cursor: "pointer",
            display: "flex",
            "align-items": "center",
            "justify-content": "center",
            transition: "background 0.2s"
          }}
        >
          {minimized() ? "＋" : "－"}
        </button>
      </div>

      {/* Content Area */}
      {!minimized() && (
        <div style={{
          padding: "1.5rem",
          display: "flex",
          "flex-direction": "column",
          gap: "2rem",
          overflow: "hidden auto",
          "scrollbar-width": "none"
        }}>
          <FusionKPIDashboard />
          <div style={{ height: "1px", background: "linear-gradient(90deg, transparent, rgba(56, 189, 248, 0.2), transparent)" }} />
          <FusionConflictPanel />
          <div style={{ height: "1px", background: "linear-gradient(90deg, transparent, rgba(56, 189, 248, 0.2), transparent)" }} />
          <TrackFusionAudit />
        </div>
      )}
    </div>
  );
}

const ControlToggle = (props: { label: string; active: boolean; onClick: () => void; color: string }) => (
  <button
    onClick={props.onClick}
    style={{
      background: "transparent",
      border: "none",
      color: props.active ? props.color : "#64748b",
      "font-size": "0.75rem",
      "font-weight": "700",
      "letter-spacing": "0.05em",
      cursor: "pointer",
      display: "flex",
      "align-items": "center",
      gap: "8px",
      transition: "color 0.2s",
    }}
  >
    <div style={{
      width: "8px",
      height: "8px",
      "border-radius": "50%",
      background: props.active ? props.color : "transparent",
      border: `1px solid ${props.active ? props.color : "#475569"}`,
      "box-shadow": props.active ? `0 0 8px ${props.color}` : "none",
    }} />
    {props.label}
  </button>
);
