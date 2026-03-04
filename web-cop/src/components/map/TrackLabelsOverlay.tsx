// CLASSIFICATION: UNCLASSIFIED
// src/components/map/TrackLabelsOverlay.tsx
//
// B: Track Labels — positioned HTML labels over the MapLibre map.
//
// MapLibre text symbol layers require a glyphs/font PBF server, which is not
// available in offline/dev deployments.  We instead render React <div>s and
// use map.project() to convert geo-coords to pixel offsets each animation
// frame.  This is toggled by layerVisibility.trackLabels in the UI store.
//
// Default:  compact short ID only (e.g. "7c5e8")
// On hover: full tooltip shown by MapTooltip — track labels stay minimal.
//
// Culling:  Labels outside the visible viewport rect are skipped in DOM.
// Density:  At most MAX_LABELS visible labels are rendered.

import React, { useEffect, useRef } from "react";
import { useTrackStore } from "../../stores/trackStore";
import { useUIStore } from "../../stores/uiStore";

const HOSTILE_LABEL_COLORS: Record<string, string> = {
  HOSTILE:   "#F87171",  // red-400
  FRIENDLY:  "#60A5FA",  // blue-400
  NEUTRAL:   "#4ADE80",  // green-400
  UNKNOWN:   "#FBBF24",  // amber-400
};

/** Derive short track ID: last 5 chars of trackId for brevity. */
function shortId(trackId: string): string {
  return trackId.length > 5 ? trackId.slice(-5) : trackId;
}

// Maximum number of labels rendered at once to avoid visual clutter
const MAX_LABELS = 200;

/**
 * TrackLabelsOverlay — minimal always-on short IDs over each track marker
 * when `layerVisibility.trackLabels` is true.
 *
 * Only short IDs are shown by default.  Full details (domain, confidence)
 * are shown in MapTooltip on mouse hover — NOT in these persistent labels.
 *
 * Culling: labels outside the map canvas bounds are skipped. In dense scenes,
 * only the first MAX_LABELS in-viewport labels are rendered.
 *
 * Relies on `window.__RTSA_MAP__` being available (set by MapView on map load).
 */
export const TrackLabelsOverlay: React.FC = () => {
  const isVisible = useUIStore((s) => s.layerVisibility.trackLabels);
  const labelContainerRef = useRef<HTMLDivElement | null>(null);
  const rafRef = useRef<number | null>(null);
  const mountedRef = useRef(true);

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
      if (rafRef.current !== null) {
        cancelAnimationFrame(rafRef.current);
        rafRef.current = null;
      }
    };
  }, []);

  // Re-render labels at 6 Hz — fast enough for tactical labels, low enough
  // to keep the main thread clear when thousands of tracks are streaming in.
  useEffect(() => {
    if (!isVisible) {
      if (labelContainerRef.current) {
        labelContainerRef.current.innerHTML = "";
      }
      return;
    }

    const LABEL_INTERVAL_MS = 167; // ~6 Hz
    let lastLabelRender = 0;

    const render = (now: number) => {
      if (!mountedRef.current) return;

      // Throttle: skip if we rendered too recently
      if (now - lastLabelRender < LABEL_INTERVAL_MS) {
        rafRef.current = requestAnimationFrame(render);
        return;
      }
      lastLabelRender = now;

      const map = (window as any).__RTSA_MAP__;
      const container = labelContainerRef.current;
      if (!map || !container) {
        rafRef.current = requestAnimationFrame(render);
        return;
      }

      const tracks = useTrackStore.getState().tracks;
      const canvas = map.getCanvas();
      const canvasWidth = canvas.width;
      const canvasHeight = canvas.height;

      // Build labels with viewport culling
      let html = "";
      let count = 0;

      for (const t of tracks.values()) {
        if (count >= MAX_LABELS) break;
        try {
          const pt = map.project([t.position.longitude, t.position.latitude]);
          // Skip labels outside the visible viewport (with small margin)
          if (pt.x < -20 || pt.x > canvasWidth + 20 || pt.y < -20 || pt.y > canvasHeight + 20) {
            continue;
          }

          const sid = shortId(t.trackId);
          const color = HOSTILE_LABEL_COLORS[t.hostileClass] ?? "#94A3B8";
          const x = Math.round(pt.x);
          const y = Math.round(pt.y);

          // Compact label: just the short ID, offset 12px right and 14px up
          html += `<div style="
            position:absolute;
            left:${x + 12}px;
            top:${y - 14}px;
            color:${color};
            font-size:9px;
            font-family:monospace;
            font-weight:700;
            white-space:nowrap;
            pointer-events:none;
            text-shadow:0 1px 4px rgba(0,0,0,1),0 0 8px rgba(0,0,0,0.8);
            letter-spacing:0.04em;
            user-select:none;
            opacity:0.85;
          ">${sid}</div>`;
          count++;
        } catch {
          // project() can throw if the map is not fully loaded — skip
        }
      }
      container.innerHTML = html;

      rafRef.current = requestAnimationFrame(render);
    };

    rafRef.current = requestAnimationFrame(render);

    return () => {
      if (rafRef.current !== null) {
        cancelAnimationFrame(rafRef.current);
        rafRef.current = null;
      }
    };
  }, [isVisible]);

  return (
    <div
      ref={labelContainerRef}
      data-testid="track-labels-overlay"
      style={{
        position: "absolute",
        inset: 0,
        pointerEvents: "none",
        zIndex: 5,
        overflow: "hidden",
      }}
    />
  );
};
