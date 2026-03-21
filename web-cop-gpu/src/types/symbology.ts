// CLASSIFICATION: UNCLASSIFIED
// src/types/symbology.ts — MIL-STD-2525 icon-index encoding helpers
//
// Provides the canonical encode/decode contract for the 32-bit `icon_index`
// field written into the 128-byte GPU binary record (SAB offset 0x24).
//
// Encoding formula:
//   icon_index = context * 36 + entity_type * 6 + threat_level
//
// Where:
//   context (0–1): 0 = MILITARY, 1 = CIVILIAN
//     → determines fill style in shader (solid fill vs outline-only)
//   entity_type (0–5): proto EntityType values
//     0=Unspecified, 1=Surface, 2=Air, 3=Subsurface, 4=Land, 5=Cyber
//     → determines inner domain icon shape in shader
//   threat_level (0–5): affiliation from TrackAffiliation enum
//     0=Unknown, 1=Pending, 2=Friendly, 3=Neutral, 4=Suspect, 5=Hostile
//     → determines outer affiliation frame shape and colour in shader
//
// The `threat_level` field in the binary record (offset 0x20) is ALSO written
// separately for backwards compatibility with shaders and pick logic.
//
// Reference: docs/implementation/v5/operations_commander/plan-mil2525SymbologyWestAsiaDemo.md §Phase 1
//            docs/sdlc_guidelines/08_tech_specific/webgpu_guidelines.md §4.1

import { TrackAffiliation, TrackDomain } from "./track-symbol";

// ── Military / Civilian context ──────────────────────────────────────────────

/** Icon rendering context for the GPU shader (packed into icon_index bits). */
export const enum TrackRenderContext {
  /** Military entity — solid affiliation-coloured fill. */
  MILITARY = 0,
  /** Civilian entity — outline-only (no fill) rendering. */
  CIVILIAN = 1,
}

// ── Proto EntityType constants (match gen/go/rtsa/common/v1 EntityType enum) ─

/** Proto EntityType values — used as the entity_type component in icon_index. */
export const ENTITY_TYPE_UNSPECIFIED = 0;
export const ENTITY_TYPE_SURFACE     = 1;
export const ENTITY_TYPE_AIR         = 2;
export const ENTITY_TYPE_SUBSURFACE  = 3;
export const ENTITY_TYPE_LAND        = 4;
export const ENTITY_TYPE_CYBER       = 5;

// ── encode / decode ──────────────────────────────────────────────────────────

/**
 * Encode the three symbology components into a single u32 icon_index.
 *
 * @param context     TrackRenderContext (0=MILITARY, 1=CIVILIAN)
 * @param entityType  Proto EntityType value (0–5)
 * @param threatLevel TrackAffiliation value (0–5)
 * @returns           Packed icon_index u32
 */
export function encodeIconIndex(
  context: number,
  entityType: number,
  threatLevel: number,
): number {
  return (context * 36 + entityType * 6 + threatLevel) >>> 0;
}

/** Decoded components from a packed icon_index value. */
export interface DecodedIconIndex {
  /** 0=MILITARY, 1=CIVILIAN */
  context: number;
  /** Proto EntityType (0–5) */
  entityType: number;
  /** TrackAffiliation (0–5) */
  threatLevel: number;
}

/**
 * Decode a packed icon_index back into its three components.
 *
 * @param iconIndex Packed icon_index u32
 * @returns         Decoded { context, entityType, threatLevel }
 */
export function decodeIconIndex(iconIndex: number): DecodedIconIndex {
  const idx = iconIndex >>> 0; // ensure unsigned
  const context     = Math.floor(idx / 36);
  const entityType  = Math.floor((idx % 36) / 6);
  const threatLevel = idx % 6;
  return { context, entityType, threatLevel };
}

// ── Mapping helpers ───────────────────────────────────────────────────────────

/**
 * Map a proto EntityType value to the frontend TrackDomain enum.
 *
 * EntityType values come from the Go proto generated code
 * (`gen/go/rtsa/common/v1`). Mapping:
 *   0 Unspecified → AIR (default)
 *   1 Surface     → SURFACE
 *   2 Air         → AIR
 *   3 Subsurface  → SUBSURFACE
 *   4 Land        → LAND
 *   5 Cyber       → CYBER
 */
export function entityTypeToTrackDomain(entityType: number): TrackDomain {
  switch (entityType) {
    case ENTITY_TYPE_AIR:        return TrackDomain.AIR;
    case ENTITY_TYPE_SURFACE:    return TrackDomain.SURFACE;
    case ENTITY_TYPE_SUBSURFACE: return TrackDomain.SUBSURFACE;
    case ENTITY_TYPE_LAND:       return TrackDomain.LAND;
    case ENTITY_TYPE_CYBER:      return TrackDomain.CYBER;
    default:                     return TrackDomain.AIR;
  }
}

/**
 * Map a threat level value to the corresponding TrackAffiliation enum member.
 *
 * Threat level values come from the `threat_level` field in the 128-byte binary
 * record (SAB offset 0x20). Mapping:
 *   0 → UNKNOWN
 *   1 → PENDING
 *   2 → FRIENDLY
 *   3 → NEUTRAL
 *   4 → SUSPECT
 *   5 → HOSTILE
 */
export function threatLevelToAffiliation(threatLevel: number): TrackAffiliation {
  switch (threatLevel) {
    case 0:  return TrackAffiliation.UNKNOWN;
    case 1:  return TrackAffiliation.PENDING;
    case 2:  return TrackAffiliation.FRIENDLY;
    case 3:  return TrackAffiliation.NEUTRAL;
    case 4:  return TrackAffiliation.SUSPECT;
    case 5:  return TrackAffiliation.HOSTILE;
    default: return TrackAffiliation.UNKNOWN;
  }
}

/**
 * Map a frontend TrackDomain to the proto EntityType value suitable for
 * inclusion in an encoded icon_index (mock data / test helpers).
 *
 *   AIR        → 2
 *   SURFACE    → 1
 *   SUBSURFACE → 3
 *   LAND       → 4
 *   SPACE      → 0 (no direct proto type; treated as Unspecified)
 *   CYBER      → 5
 */
export function trackDomainToEntityType(domain: TrackDomain): number {
  switch (domain) {
    case TrackDomain.AIR:        return ENTITY_TYPE_AIR;
    case TrackDomain.SURFACE:    return ENTITY_TYPE_SURFACE;
    case TrackDomain.SUBSURFACE: return ENTITY_TYPE_SUBSURFACE;
    case TrackDomain.LAND:       return ENTITY_TYPE_LAND;
    case TrackDomain.CYBER:      return ENTITY_TYPE_CYBER;
    case TrackDomain.SPACE:      return ENTITY_TYPE_UNSPECIFIED;
    default:                     return ENTITY_TYPE_UNSPECIFIED;
  }
}
