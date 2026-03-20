// CLASSIFICATION: UNCLASSIFIED
// src/components/symbols/TrackSymbol.tsx — SVG track symbol renderer
//
// Renders a dynamic vector symbol for a fused track following MIL-STD-2525C /
// NATO APP-6 principles.  Shape is determined by TrackDomain; colour by
// TrackAffiliation; outline style by TrackContext.
//
// The component is intentionally pure SVG (no canvas / WebGPU) so it can be
// embedded anywhere: TrackDetailPanel, dashboards, legends, PDF exports.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md §4.1
//            docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md §3

import { Component, Show, createMemo } from "solid-js";
import {
  TrackAffiliation,
  TrackContext,
  TrackDomain,
  type TrackSymbolProps,
} from "../../types/track-symbol";

// ── Colour palette (MIL-STD-2525C standard) ──────────────────────────────────

const AFFILIATION_FILL: Record<TrackAffiliation, string> = {
  [TrackAffiliation.UNKNOWN]:  "#f59e0b", // Yellow
  [TrackAffiliation.PENDING]:  "#06b6d4", // Cyan
  [TrackAffiliation.FRIENDLY]: "#3b82f6", // Blue
  [TrackAffiliation.NEUTRAL]:  "#22c55e", // Green
  [TrackAffiliation.SUSPECT]:  "#f97316", // Orange
  [TrackAffiliation.HOSTILE]:  "#ef4444", // Red
};

const AFFILIATION_STROKE: Record<TrackAffiliation, string> = {
  [TrackAffiliation.UNKNOWN]:  "#d97706", // Amber border
  [TrackAffiliation.PENDING]:  "#0891b2", // Dark cyan border
  [TrackAffiliation.FRIENDLY]: "#1d4ed8", // Dark blue border
  [TrackAffiliation.NEUTRAL]:  "#15803d", // Dark green border
  [TrackAffiliation.SUSPECT]:  "#c2410c", // Dark orange border
  [TrackAffiliation.HOSTILE]:  "#b91c1c", // Dark red border
};

/** Human-readable label for aria-label / legend. */
const AFFILIATION_LABEL: Record<TrackAffiliation, string> = {
  [TrackAffiliation.UNKNOWN]:  "Unknown",
  [TrackAffiliation.PENDING]:  "Pending",
  [TrackAffiliation.FRIENDLY]: "Friendly",
  [TrackAffiliation.NEUTRAL]:  "Neutral",
  [TrackAffiliation.SUSPECT]:  "Suspect",
  [TrackAffiliation.HOSTILE]:  "Hostile",
};

const DOMAIN_LABEL: Record<TrackDomain, string> = {
  [TrackDomain.AIR]:        "Air",
  [TrackDomain.SURFACE]:    "Surface",
  [TrackDomain.SUBSURFACE]: "Subsurface",
  [TrackDomain.LAND]:       "Land",
  [TrackDomain.SPACE]:      "Space",
  [TrackDomain.CYBER]:      "Cyber",
};

// ── Outline dash-array per context ───────────────────────────────────────────

function contextStrokeDashArray(ctx: TrackContext, strokeW: number): string | undefined {
  switch (ctx) {
    case TrackContext.EXERCISE:   return `${strokeW * 4},${strokeW * 2}`;
    case TrackContext.SIMULATION: return `${strokeW},${strokeW * 2}`;
    case TrackContext.TEST:       return `${strokeW * 2},${strokeW}`;
    default:                      return undefined; // REAL — solid
  }
}

// ── Shape paths (normalised to 0 0 100 100 viewBox) ──────────────────────────

/**
 * Air — swept-back triangle pointing upward (nose = top).
 * Inspired by MIL-STD-2525C Air track silhouette.
 */
const PATH_AIR = "M50,8 L82,72 L50,58 L18,72 Z";

/**
 * Surface — diamond hull (equal four-point diamond).
 * MIL-STD-2525C Surface track uses a square rotated 45°.
 */
const PATH_SURFACE = "M50,10 L90,50 L50,90 L10,50 Z";

/**
 * Subsurface — horizontal ellipse.
 * The ellipse is wider than tall, indicating underwater domain.
 */
const PATH_SUBSURFACE_RX = "40";
const PATH_SUBSURFACE_RY = "22";

/**
 * Land — filled square.
 * MIL-STD-2525C Ground unit uses a rectangle/square with icon overlay.
 */
const PATH_LAND = "M15,15 L85,15 L85,85 L15,85 Z";

/**
 * Space — circle with pointed top (orbiting satellite).
 * Rounded body with a small fin-like protrusion at the top.
 */
const PATH_SPACE_BODY_R = "32";
const PATH_SPACE_FIN = "M50,16 L44,6 L50,0 L56,6 Z";

/**
 * Cyber — regular hexagon (logical / non-physical domain).
 */
const PATH_CYBER = "M50,5 L90,27 L90,73 L50,95 L10,73 L10,27 Z";

// ── Shape renderer ────────────────────────────────────────────────────────────

interface ShapeProps {
  domain:      TrackDomain;
  fill:        string;
  stroke:      string;
  strokeW:     number;
  dashArray:   string | undefined;
  context:     TrackContext;
}

function TrackShape(props: ShapeProps) {
  const baseShapeProps = () => ({
    fill:             props.fill,
    stroke:           props.stroke,
    "stroke-width":   props.strokeW,
    "stroke-dasharray": props.dashArray,
    "stroke-linejoin": "round" as const,
    "stroke-linecap":  "round" as const,
  });

  switch (props.domain) {
    case TrackDomain.AIR:
      return <path d={PATH_AIR} {...baseShapeProps()} />;

    case TrackDomain.SURFACE:
      return <path d={PATH_SURFACE} {...baseShapeProps()} />;

    case TrackDomain.SUBSURFACE:
      return (
        <ellipse
          cx="50" cy="50"
          rx={PATH_SUBSURFACE_RX} ry={PATH_SUBSURFACE_RY}
          {...baseShapeProps()}
        />
      );

    case TrackDomain.LAND:
      return <path d={PATH_LAND} {...baseShapeProps()} />;

    case TrackDomain.SPACE:
      return (
        <>
          <circle
            cx="50" cy="52"
            r={PATH_SPACE_BODY_R}
            {...baseShapeProps()}
          />
          <path d={PATH_SPACE_FIN} {...baseShapeProps()} />
        </>
      );

    case TrackDomain.CYBER:
      return <path d={PATH_CYBER} {...baseShapeProps()} />;

    default:
      return <path d={PATH_AIR} {...baseShapeProps()} />;
  }
}

// ── TEST context crosshair overlay ───────────────────────────────────────────

function TestCrosshair(props: { stroke: string; strokeW: number }) {
  return (
    <>
      <line x1="50" y1="20" x2="50" y2="80" stroke={props.stroke} stroke-width={props.strokeW * 0.5} />
      <line x1="20" y1="50" x2="80" y2="50" stroke={props.stroke} stroke-width={props.strokeW * 0.5} />
    </>
  );
}

// ── Selection ring ────────────────────────────────────────────────────────────

function SelectionRing() {
  return (
    <circle
      cx="50" cy="50" r="46"
      fill="none"
      stroke="#00ffff"
      stroke-width="3"
      stroke-dasharray="6,3"
      opacity="0.9"
    />
  );
}

// ── Main component ────────────────────────────────────────────────────────────

/**
 * `TrackSymbol` renders a scalable vector symbol for a fused track.
 *
 * @example
 * ```tsx
 * <TrackSymbol
 *   domain={TrackDomain.AIR}
 *   affiliation={TrackAffiliation.HOSTILE}
 *   context={TrackContext.REAL}
 *   size={40}
 *   selected={false}
 * />
 * ```
 */
export const TrackSymbol: Component<TrackSymbolProps> = (props) => {
  const size        = () => props.size ?? 32;
  const fill        = () => AFFILIATION_FILL[props.affiliation];
  const stroke      = () => AFFILIATION_STROKE[props.affiliation];
  const strokeW     = () => Math.max(3, size() * 0.09); // scale stroke with icon size
  const dashArray   = () => contextStrokeDashArray(props.context, strokeW());
  const fillOpacity = createMemo(() =>
    props.context === TrackContext.SIMULATION ? 0.35 : 0.82,
  );

  const ariaLabel = () =>
    `${AFFILIATION_LABEL[props.affiliation]} ${DOMAIN_LABEL[props.domain]} track`;

  return (
    <svg
      width={size()}
      height={size()}
      viewBox="0 0 100 100"
      aria-label={ariaLabel()}
      role="img"
      style={{ display: "inline-block", "vertical-align": "middle", "flex-shrink": "0" }}
    >
      <g fill-opacity={fillOpacity()}>
        <TrackShape
          domain={props.domain}
          fill={fill()}
          stroke={stroke()}
          strokeW={strokeW()}
          dashArray={dashArray()}
          context={props.context}
        />
      </g>

      {/* TEST context: additional crosshair overlay */}
      <Show when={props.context === TrackContext.TEST}>
        <TestCrosshair stroke={stroke()} strokeW={strokeW()} />
      </Show>

      {/* Selection ring */}
      <Show when={props.selected}>
        <SelectionRing />
      </Show>
    </svg>
  );
};
