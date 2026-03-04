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
// Label format: "{shortId} | {DOMAIN} | {confidence}%"
// e.g.  "7c5 | SURFACE | 90%"

import React, { useEffect, useRef } from "react";
import { useTrackStore } from "../../stores/trackStore";
import { useUIStore } from "../../stores/uiStore";

interface LabelPos {
  trackId: string;
  shortId: string;
  label: string;
  hostileClass: string;
  x: number;
  y: number;
}

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

/** Domain from entityType string (mirrors MapView.getEntityDomain). */
function domain(entityType: string): string {
  const t = entityType.toUpperCase();
  if (t.includes("AIR") || t.includes("AIRCRAFT")) return "AIR";
  if (t.includes("SUB")) return "SUB";
  if (t.includes("LAND") || t.includes("VEHICLE")) return "LAND";
  if (t.includes("SURFACE") || t.includes("SHIP") || t.includes("VESSEL")) return "SFC";
  return "UNK";
}

/**
 * TrackLabelsOverlay — headless React component that renders small HTML labels
 * over each track marker when `layerVisibility.trackLabels` is true.
 *
 * Relies on `window.__RTSA_MAP__` being available (set by MapView on map load).
 */
export const TrackLabelsOverlay: React.FC = () => {
  const isVisible = useUIStore((s) => s.layerVisibility.trackLabels);
  const labelContainerRef = useRef<HTMLDivElement | null>(null);
  // Keep a stable div for labels so we can do direct DOM updates in RAF
  // without causing React re-renders.
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

  // Re-render labels every frame when visible
  useEffect(() => {
    if (!isVisible) {
      // Clear any existing labels
      if (labelContainerRef.current) {
        labelContainerRef.current.innerHTML = "";
      }
      return;
    }

    const render = () => {
      if (!mountedRef.current) return;

      const map = (window as any).__RTSA_MAP__;
      const container = labelContainerRef.current;
      if (!map || !container) {
        // Map not ready yet — retry next frame
        rafRef.current = requestAnimationFrame(render);
        return;
      }

      const tracks = useTrackStore.getState().tracks;
      const labels: LabelPos[] = [];

      for (const t of tracks.values()) {
        try {
          const pt = map.project([t.position.longitude, t.position.latitude]);
          const dom = domain(t.entityType);
          const conf = Math.round(t.confidenceScore * 100);
          const sid = shortId(t.trackId);
          labels.push({
            trackId: t.trackId,
            shortId: sid,
            label: `${sid} | ${dom} | ${conf}%`,
            hostileClass: t.hostileClass,
            x: Math.round(pt.x),
            y: Math.round(pt.y),
          });
        } catch {
          // project() can throw if the map is not fully loaded — skip
        }
      }

      // Fast DOM update — recreate inner HTML rather than reconciling React children.
      // This avoids per-label React renders while keeping label divs lightweight.
      let html = "";
      for (const lbl of labels) {
        const color = HOSTILE_LABEL_COLORS[lbl.hostileClass] ?? "#94A3B8";
        // Offset the label 10px right and 16px up from the icon center
        html += `<div style="
          position:absolute;
          left:${lbl.x + 10}px;
          top:${lbl.y - 16}px;
          color:${color};
          font-size:10px;
          font-family:monospace;
          font-weight:600;
          white-space:nowrap;
          pointer-events:none;
          text-shadow:0 1px 3px rgba(0,0,0,0.9),0 0 6px rgba(0,0,0,0.7);
          letter-spacing:0.03em;
          user-select:none;
        ">${lbl.label}</div>`;
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

  // This div is absolutely positioned over the map. Its parent is the map container.
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
