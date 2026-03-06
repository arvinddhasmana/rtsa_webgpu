// CLASSIFICATION: UNCLASSIFIED
// src/components/map/TrackLayer.tsx
//
// Enhanced TrackLayer — renders track markers with:
//   - Domain-specific SVG shapes (SURFACE=diamond, AIR=triangle,
//     SUBSURFACE=submarine-circle, LAND=square, CYBER=cross, UNKNOWN=circle)
//   - Color by hostile classification (HOSTILE=red, FRIENDLY=blue,
//     NEUTRAL=green, UNKNOWN=amber)
//   - Size scaled by confidence (8–14 px)
//   - Track labels toggle (reads uiStore.layerVisibility.trackLabels)
//   - Track trails toggle (reads uiStore.layerVisibility.trackTrails +
//     trackStore.trackHistory)

import React from "react";
import { useTrackStore } from "../../stores/trackStore";
import { useUIStore } from "../../stores/uiStore";
import type { FusedTrack } from "../../types/track";
import { getHostileColor } from "../../utils/mil-symbology";

interface TrackLayerProps {
  tracks: FusedTrack[];
  onTrackClick: (trackId: string) => void;
}

/** Scale marker size 8–14 px by confidence 0–1. */
function getMarkerSize(confidence: number): number {
  return Math.round(8 + Math.max(0, Math.min(1, confidence)) * 6);
}

/** Domain-specific SVG shape rendered at the given size in the given fill colour. */
function DomainShape({
  entityType,
  size,
  color,
}: {
  entityType: string;
  size: number;
  color: string;
}): React.ReactElement {
  const half = size / 2;
  switch (entityType) {
    case "AIR":
      // Upward-pointing triangle (fixed-wing silhouette)
      return (
        <svg
          width={size}
          height={size}
          viewBox={`0 0 ${size} ${size}`}
          style={{ display: "block" }}
        >
          <polygon
            points={`${half},2 ${size - 2},${size - 2} 2,${size - 2}`}
            fill={color}
            stroke="white"
            strokeWidth="1.2"
          />
        </svg>
      );

    case "SURFACE":
      // Diamond (surface vessel)
      return (
        <svg
          width={size}
          height={size}
          viewBox={`0 0 ${size} ${size}`}
          style={{ display: "block" }}
        >
          <polygon
            points={`${half},2 ${size - 2},${half} ${half},${size - 2} 2,${half}`}
            fill={color}
            stroke="white"
            strokeWidth="1.2"
          />
        </svg>
      );

    case "SUBSURFACE":
      // Filled circle with horizontal line (submarine)
      return (
        <svg
          width={size}
          height={size}
          viewBox={`0 0 ${size} ${size}`}
          style={{ display: "block" }}
        >
          <circle
            cx={half}
            cy={half}
            r={half - 1}
            fill={color}
            stroke="white"
            strokeWidth="1.2"
          />
          <line
            x1="2"
            y1={half}
            x2={size - 2}
            y2={half}
            stroke="white"
            strokeWidth="1.2"
          />
        </svg>
      );

    case "LAND":
      // Square (ground vehicle / land unit)
      return (
        <svg
          width={size}
          height={size}
          viewBox={`0 0 ${size} ${size}`}
          style={{ display: "block" }}
        >
          <rect
            x="2"
            y="2"
            width={size - 4}
            height={size - 4}
            fill={color}
            stroke="white"
            strokeWidth="1.2"
          />
        </svg>
      );

    case "CYBER":
      // Bold cross / lightning (cyber entity)
      return (
        <svg
          width={size}
          height={size}
          viewBox={`0 0 ${size} ${size}`}
          style={{ display: "block" }}
        >
          <line
            x1={half}
            y1="1"
            x2={half}
            y2={size - 1}
            stroke={color}
            strokeWidth="3"
            strokeLinecap="round"
          />
          <line
            x1="1"
            y1={half}
            x2={size - 1}
            y2={half}
            stroke={color}
            strokeWidth="3"
            strokeLinecap="round"
          />
        </svg>
      );

    default:
      // Circle (UNKNOWN / fallback)
      return (
        <svg
          width={size}
          height={size}
          viewBox={`0 0 ${size} ${size}`}
          style={{ display: "block" }}
        >
          <circle
            cx={half}
            cy={half}
            r={half - 1}
            fill={color}
            stroke="white"
            strokeWidth="1.2"
          />
        </svg>
      );
  }
}

/** Format abbreviated label: {sensorPrefix}-{last4} | {domain} | {confidence%} */
function formatLabel(track: FusedTrack): string {
  const sensorPrefix =
    track.sources.length > 0
      ? track.sources[0].sensorType.slice(0, 3).toUpperCase()
      : "UNK";
  const last4 = track.trackId.slice(-4);
  const pct = Math.round(track.confidenceScore * 100);
  return `${sensorPrefix}-${last4} | ${track.entityType} | ${pct}%`;
}

/**
 * TrackLayer — renders fused track markers as an absolutely-positioned
 * HTML overlay on top of a map.  The map parent is responsible for
 * computing pixel-space positions and passing final (x,y) coordinates if
 * embedded in a full MapLibre context; in the current DOM-overlay mode the
 * markers are stacked at the top-left of the container and differentiated by
 * shape/colour.
 *
 * All MIL-STD-2525-inspired symbology rules:
 *   Domain  → shape
 *   Threat  → colour
 *   Score   → size
 *   Status  → opacity
 */
export const TrackLayer: React.FC<TrackLayerProps> = ({
  tracks,
  onTrackClick,
}) => {
  const showLabels = useUIStore((s) => s.layerVisibility.trackLabels);
  const showTrails = useUIStore((s) => s.layerVisibility.trackTrails);
  const trackHistory = useTrackStore((s) => s.trackHistory);

  return (
    <div data-testid="track-layer">
      {tracks.map((track) => {
        const color = getHostileColor(track.hostileClass);
        const size = getMarkerSize(track.confidenceScore);
        const opacity = track.status === "STALE" ? 0.5 : 1;
        const trail = showTrails ? (trackHistory.get(track.trackId) ?? []) : [];

        return (
          <React.Fragment key={track.trackId}>
            {/* ── Trail breadcrumbs ───────────────────── */}
            {trail.length > 1 && (
              <svg
                style={{
                  position: "absolute",
                  top: 0,
                  left: 0,
                  width: "100%",
                  height: "100%",
                  pointerEvents: "none",
                  overflow: "visible",
                }}
                aria-hidden="true"
              >
                {trail.slice(0, -1).map((pt, idx) => {
                  const next = trail[idx + 1];
                  const trailOpacity = ((idx + 1) / trail.length) * 0.5;
                  return (
                    <line
                      key={idx}
                      x1={pt[0]}
                      y1={pt[1]}
                      x2={next[0]}
                      y2={next[1]}
                      stroke={color}
                      strokeWidth="1.5"
                      strokeOpacity={trailOpacity}
                      strokeLinecap="round"
                    />
                  );
                })}
              </svg>
            )}

            {/* ── Track marker ────────────────────────── */}
            <div
              data-testid={`track-marker-${track.trackId}`}
              onClick={() => onTrackClick(track.trackId)}
              title={`${track.trackId} (${track.entityType} / ${track.hostileClass}) — ${Math.round(track.confidenceScore * 100)}% conf`}
              style={{
                position: "absolute",
                width: `${size}px`,
                height: `${size}px`,
                cursor: "pointer",
                opacity,
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                filter:
                  track.hostileClass === "HOSTILE"
                    ? `drop-shadow(0 0 3px ${color})`
                    : undefined,
              }}
            >
              <DomainShape
                entityType={track.entityType}
                size={size}
                color={color}
              />
            </div>

            {/* ── Track label ─────────────────────────── */}
            {showLabels && (
              <div
                style={{
                  position: "absolute",
                  fontSize: "0.55rem",
                  fontFamily: "monospace",
                  color: "#E2E8F0",
                  backgroundColor: "rgba(10, 15, 26, 0.7)",
                  padding: "1px 4px",
                  borderRadius: "2px",
                  whiteSpace: "nowrap",
                  pointerEvents: "none",
                  border: `1px solid ${color}44`,
                  userSelect: "none",
                }}
                aria-hidden="true"
              >
                {formatLabel(track)}
              </div>
            )}
          </React.Fragment>
        );
      })}
    </div>
  );
};
