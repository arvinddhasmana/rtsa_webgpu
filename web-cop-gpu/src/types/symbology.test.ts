// CLASSIFICATION: UNCLASSIFIED
// src/types/symbology.test.ts — Unit tests for MIL-STD-2525 icon-index helpers
//
// Validates encodeIconIndex / decodeIconIndex roundtrip, mapping functions,
// and encoding of all domain/affiliation/context combinations.
//
// Reference: docs/implementation/v5/operations_commander/plan-mil2525SymbologyWestAsiaDemo.md §Phase 8.1

import { describe, it, expect } from "vitest";
import {
  encodeIconIndex,
  decodeIconIndex,
  entityTypeToTrackDomain,
  threatLevelToAffiliation,
  trackDomainToEntityType,
  TrackRenderContext,
  ENTITY_TYPE_UNSPECIFIED,
  ENTITY_TYPE_SURFACE,
  ENTITY_TYPE_AIR,
  ENTITY_TYPE_SUBSURFACE,
  ENTITY_TYPE_LAND,
  ENTITY_TYPE_CYBER,
} from "./symbology";
import { TrackDomain, TrackAffiliation } from "./track-symbol";

// ── encodeIconIndex / decodeIconIndex roundtrip ───────────────────────────────

describe("encodeIconIndex / decodeIconIndex roundtrip", () => {
  const cases: Array<{ context: number; entityType: number; threatLevel: number }> = [
    { context: 0, entityType: 0, threatLevel: 0 },
    { context: 0, entityType: 2, threatLevel: 5 }, // Military Air Hostile
    { context: 1, entityType: 1, threatLevel: 3 }, // Civilian Surface Neutral
    { context: 0, entityType: 4, threatLevel: 2 }, // Military Land Friendly
    { context: 1, entityType: 5, threatLevel: 0 }, // Civilian Cyber Unknown
    { context: 0, entityType: 3, threatLevel: 4 }, // Military Subsurface Suspect
    { context: 0, entityType: 5, threatLevel: 5 }, // Military Cyber Hostile
    { context: 1, entityType: 2, threatLevel: 1 }, // Civilian Air Pending
  ];

  it.each(cases)("roundtrip: context=$context et=$entityType tl=$threatLevel", ({ context, entityType, threatLevel }) => {
    const encoded = encodeIconIndex(context, entityType, threatLevel);
    const decoded = decodeIconIndex(encoded);
    expect(decoded.context).toBe(context);
    expect(decoded.entityType).toBe(entityType);
    expect(decoded.threatLevel).toBe(threatLevel);
  });

  it("encodes all 72 combinations without collision", () => {
    const seen = new Set<number>();
    for (let ctx = 0; ctx <= 1; ctx++) {
      for (let et = 0; et <= 5; et++) {
        for (let tl = 0; tl <= 5; tl++) {
          const v = encodeIconIndex(ctx, et, tl);
          expect(seen.has(v)).toBe(false);
          seen.add(v);
        }
      }
    }
    expect(seen.size).toBe(72);
  });

  it("returns unsigned (non-negative) values", () => {
    const v = encodeIconIndex(1, 5, 5);
    expect(v).toBeGreaterThanOrEqual(0);
    // max value: 1*36 + 5*6 + 5 = 71
    expect(v).toBe(71);
  });

  it("minimum value is 0 (Military Unspecified Unknown)", () => {
    expect(encodeIconIndex(0, 0, 0)).toBe(0);
    const d = decodeIconIndex(0);
    expect(d).toEqual({ context: 0, entityType: 0, threatLevel: 0 });
  });

  it("military/civilian boundary: context bit separates at 36", () => {
    const mil = encodeIconIndex(TrackRenderContext.MILITARY, ENTITY_TYPE_AIR, 2);
    const civ = encodeIconIndex(TrackRenderContext.CIVILIAN, ENTITY_TYPE_AIR, 2);
    expect(civ - mil).toBe(36);
    expect(decodeIconIndex(mil).context).toBe(0);
    expect(decodeIconIndex(civ).context).toBe(1);
  });
});

// ── entityTypeToTrackDomain ───────────────────────────────────────────────────

describe("entityTypeToTrackDomain", () => {
  it("maps ENTITY_TYPE_AIR (2) to TrackDomain.AIR", () => {
    expect(entityTypeToTrackDomain(ENTITY_TYPE_AIR)).toBe(TrackDomain.AIR);
  });

  it("maps ENTITY_TYPE_SURFACE (1) to TrackDomain.SURFACE", () => {
    expect(entityTypeToTrackDomain(ENTITY_TYPE_SURFACE)).toBe(TrackDomain.SURFACE);
  });

  it("maps ENTITY_TYPE_SUBSURFACE (3) to TrackDomain.SUBSURFACE", () => {
    expect(entityTypeToTrackDomain(ENTITY_TYPE_SUBSURFACE)).toBe(TrackDomain.SUBSURFACE);
  });

  it("maps ENTITY_TYPE_LAND (4) to TrackDomain.LAND", () => {
    expect(entityTypeToTrackDomain(ENTITY_TYPE_LAND)).toBe(TrackDomain.LAND);
  });

  it("maps ENTITY_TYPE_CYBER (5) to TrackDomain.CYBER", () => {
    expect(entityTypeToTrackDomain(ENTITY_TYPE_CYBER)).toBe(TrackDomain.CYBER);
  });

  it("defaults ENTITY_TYPE_UNSPECIFIED (0) to TrackDomain.AIR", () => {
    expect(entityTypeToTrackDomain(ENTITY_TYPE_UNSPECIFIED)).toBe(TrackDomain.AIR);
  });

  it("defaults unknown values to TrackDomain.AIR", () => {
    expect(entityTypeToTrackDomain(99)).toBe(TrackDomain.AIR);
  });
});

// ── trackDomainToEntityType ───────────────────────────────────────────────────

describe("trackDomainToEntityType", () => {
  const mapping: Array<[TrackDomain, number]> = [
    [TrackDomain.AIR,        ENTITY_TYPE_AIR],
    [TrackDomain.SURFACE,    ENTITY_TYPE_SURFACE],
    [TrackDomain.SUBSURFACE, ENTITY_TYPE_SUBSURFACE],
    [TrackDomain.LAND,       ENTITY_TYPE_LAND],
    [TrackDomain.CYBER,      ENTITY_TYPE_CYBER],
    [TrackDomain.SPACE,      ENTITY_TYPE_UNSPECIFIED],
  ];

  it.each(mapping)("domain %i → entityType %i", (domain, expected) => {
    expect(trackDomainToEntityType(domain)).toBe(expected);
  });

  it("is the inverse of entityTypeToTrackDomain for the main domains", () => {
    const domains = [TrackDomain.AIR, TrackDomain.SURFACE, TrackDomain.SUBSURFACE, TrackDomain.LAND, TrackDomain.CYBER];
    for (const d of domains) {
      const et = trackDomainToEntityType(d);
      expect(entityTypeToTrackDomain(et)).toBe(d);
    }
  });
});

// ── threatLevelToAffiliation ──────────────────────────────────────────────────

describe("threatLevelToAffiliation", () => {
  const mapping: Array<[number, TrackAffiliation]> = [
    [0, TrackAffiliation.UNKNOWN],
    [1, TrackAffiliation.PENDING],
    [2, TrackAffiliation.FRIENDLY],
    [3, TrackAffiliation.NEUTRAL],
    [4, TrackAffiliation.SUSPECT],
    [5, TrackAffiliation.HOSTILE],
  ];

  it.each(mapping)("threatLevel %i → affiliation %i", (tl, expected) => {
    expect(threatLevelToAffiliation(tl)).toBe(expected);
  });

  it("defaults unknown values to UNKNOWN", () => {
    expect(threatLevelToAffiliation(99)).toBe(TrackAffiliation.UNKNOWN);
  });
});

// ── Context constants ─────────────────────────────────────────────────────────

describe("TrackRenderContext", () => {
  it("MILITARY is 0", () => expect(TrackRenderContext.MILITARY).toBe(0));
  it("CIVILIAN is 1", () => expect(TrackRenderContext.CIVILIAN).toBe(1));
});

// ── Mock-data encoding compatibility ─────────────────────────────────────────

describe("icon_index encoding for mock-data", () => {
  it("Military Air Hostile encodes to correct value", () => {
    // context=0, entityType=2(AIR), threatLevel=5(HOSTILE) → 0*36 + 2*6 + 5 = 17
    expect(encodeIconIndex(TrackRenderContext.MILITARY, ENTITY_TYPE_AIR, 5)).toBe(17);
  });

  it("Civilian Surface Neutral encodes to correct value", () => {
    // context=1, entityType=1(SURFACE), threatLevel=3(NEUTRAL) → 1*36 + 1*6 + 3 = 45
    expect(encodeIconIndex(TrackRenderContext.CIVILIAN, ENTITY_TYPE_SURFACE, 3)).toBe(45);
  });

  it("shader entity_type extraction: (icon_index % 36) / 6 gives correct entity_type", () => {
    for (let et = 0; et <= 5; et++) {
      const encoded = encodeIconIndex(0, et, 3);
      const extracted = Math.floor((encoded % 36) / 6);
      expect(extracted).toBe(et);
    }
  });

  it("shader context extraction: icon_index / 36 gives correct context", () => {
    expect(Math.floor(encodeIconIndex(0, 2, 5) / 36)).toBe(0);
    expect(Math.floor(encodeIconIndex(1, 2, 5) / 36)).toBe(1);
  });

  it("shader threat_level extraction: icon_index % 6 gives correct threat_level", () => {
    for (let tl = 0; tl <= 5; tl++) {
      const encoded = encodeIconIndex(0, 3, tl);
      expect(encoded % 6).toBe(tl);
    }
  });
});
