// CLASSIFICATION: UNCLASSIFIED
// src/gpu/lod.ts — Level-of-Detail system for track rendering
//
// At high zoom-out levels (many km per pixel), rendering all 50k tracks
// at full detail wastes GPU time. The LOD system reduces instance count
// and disables expensive effects (halos, labels) at low zoom levels.
//
// Reference: docs/implementation/v4/phase4_hardening_cutover.md H4-2

/**
 * LOD level thresholds based on camera scale (world-units per pixel).
 *
 * Scale semantics: higher scale value = more zoomed in.
 */
export const LOD_LEVELS = {
  /** Full detail: halos, labels, trails, icons at full quality. */
  FULL: 0.5,
  /** Medium: trails disabled, labels reduced. */
  MEDIUM: 0.1,
  /** Minimal: only icons rendered, no trails, no halos, no labels. */
  MINIMAL: 0.01,
} as const;

export type LodLevel = "full" | "medium" | "minimal";

/** Rendering flags controlled by the current LOD level. */
export interface LodFlags {
  level: LodLevel;
  /** Render trail lines. */
  renderTrails: boolean;
  /** Render alert halos. */
  renderHalos: boolean;
  /** Render SDF labels. */
  renderLabels: boolean;
  /** Maximum number of instances to render. Always an explicit count ≥ 0. */
  maxInstances: number;
}

/**
 * Compute the LOD flags for the current camera scale.
 *
 * @param scale - Camera scale (world-units per pixel); from camera.scale.
 * @param trackCount - Current active track count.
 */
export function computeLod(scale: number, trackCount: number): LodFlags {
  if (scale >= LOD_LEVELS.FULL) {
    return {
      level: "full",
      renderTrails: true,
      renderHalos: true,
      renderLabels: true,
      maxInstances: trackCount,
    };
  }

  if (scale >= LOD_LEVELS.MEDIUM) {
    return {
      level: "medium",
      renderTrails: false,
      renderHalos: true,
      renderLabels: trackCount <= 10_000,
      // At medium LOD with many tracks, cap rendered instances to 20k
      maxInstances: Math.min(trackCount, 20_000),
    };
  }

  return {
    level: "minimal",
    renderTrails: false,
    renderHalos: false,
    renderLabels: false,
    // At minimal LOD, render at most 10k icons (culling handles the rest)
    maxInstances: Math.min(trackCount, 10_000),
  };
}
