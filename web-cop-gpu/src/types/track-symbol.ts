// CLASSIFICATION: UNCLASSIFIED
// src/types/track-symbol.ts — MIL-STD-2525 / NATO APP-6 track symbol types
//
// Defines the canonical enumerations and props for the TrackSymbol component.
// Values are intentionally numeric so they can be stored directly in the SAB
// icon_index (domain) and threat_level (affiliation) fields.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md §4.1

// ── Track Domain ─────────────────────────────────────────────────────────────

/**
 * Track domain — the physical environment the entity operates in.
 * Drives the silhouette shape rendered in the WebGPU icon pass
 * (track-icons.wgsl `get_silhouette`) and the SVG TrackSymbol component.
 *
 * Numeric values are stored in the SAB `icon_index` field (u32 @ 0x24).
 */
export enum TrackDomain {
  /** Fixed-wing or rotary-wing aircraft. Shape: swept-wing triangle. */
  AIR        = 0,
  /** Naval / maritime surface vessel. Shape: diamond hull. */
  SURFACE    = 1,
  /** Submarine or underwater asset. Shape: horizontal ellipse. */
  SUBSURFACE = 2,
  /** Ground vehicle or dismounted infantry. Shape: filled square. */
  LAND       = 3,
  /** Satellite or space-based asset. Shape: circle with pointed top. */
  SPACE      = 4,
  /** Cyber entity (logical, non-physical). Shape: hexagon. */
  CYBER      = 5,
}

// ── Track Affiliation ────────────────────────────────────────────────────────

/**
 * Track affiliation — the friend/foe/neutral identity assessment.
 * Drives the fill and stroke colour in both the WebGPU shader and the SVG
 * TrackSymbol component.
 *
 * Numeric values are stored in the SAB `threat_level` field (u32 @ 0x20).
 *
 * Colour mapping (MIL-STD-2525C standard palette):
 *   UNKNOWN   → Yellow  (#f59e0b)
 *   PENDING   → Cyan    (#06b6d4)
 *   FRIENDLY  → Blue    (#3b82f6)
 *   NEUTRAL   → Green   (#22c55e)
 *   SUSPECT   → Orange  (#f97316)
 *   HOSTILE   → Red     (#ef4444)
 */
export enum TrackAffiliation {
  /** Identity undetermined — default until IFF resolution. */
  UNKNOWN  = 0,
  /** Pending IFF resolution — transitional state. */
  PENDING  = 1,
  /** Positively identified as own force / allied. */
  FRIENDLY = 2,
  /** Not a declared threat; benign. */
  NEUTRAL  = 3,
  /** Potentially hostile; subject to further classification. */
  SUSPECT  = 4,
  /** Confirmed enemy force. */
  HOSTILE  = 5,
}

// ── Track Context ─────────────────────────────────────────────────────────────

/**
 * Operational context — modifies outline rendering to indicate whether the
 * track is live operational data, an exercise entity, a simulation artefact,
 * or a test signal.
 *
 * Numeric values are NOT stored in the SAB; they are UI-layer metadata only.
 */
export enum TrackContext {
  /** Live operational data — solid outline. */
  REAL       = 0,
  /** Simulation / rehearsal entity — dashed outline. */
  EXERCISE   = 1,
  /** Computer-generated entity — dotted outline. */
  SIMULATION = 2,
  /** Equipment test signal — thin outline + centre crosshair. */
  TEST       = 3,
}

// ── TrackSymbolProps ─────────────────────────────────────────────────────────

/** Props for the `TrackSymbol` SVG component. */
export interface TrackSymbolProps {
  /** Track domain — determines the silhouette shape. */
  domain: TrackDomain;
  /** Track affiliation — determines fill and stroke colour. */
  affiliation: TrackAffiliation;
  /** Operational context — determines outline style. */
  context: TrackContext;
  /**
   * Pixel size (width and height of the bounding square).
   * Default: 32
   */
  size?: number;
  /** Whether the track is currently selected — renders a cyan selection ring. */
  selected?: boolean;
}
