// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/utils/mil-symbology.test.ts

import { describe, it, expect } from "vitest";
import {
  getHostileColor,
  getEntitySymbol,
  getEntityShape,
} from "../../utils/mil-symbology";

describe("getHostileColor", () => {
  it("HOSTILE returns red", () => {
    expect(getHostileColor("HOSTILE")).toBe("#DC2626");
  });

  it("FRIENDLY returns blue", () => {
    expect(getHostileColor("FRIENDLY")).toBe("#2563EB");
  });

  it("NEUTRAL returns green", () => {
    expect(getHostileColor("NEUTRAL")).toBe("#16A34A");
  });

  it("UNKNOWN returns yellow", () => {
    expect(getHostileColor("UNKNOWN")).toBe("#CA8A04");
  });

  it("unknown values return grey", () => {
    expect(getHostileColor("OTHER" as "HOSTILE")).toBe("#6B7280");
  });
});

describe("getEntitySymbol", () => {
  it("returns a symbol for each entity type", () => {
    expect(getEntitySymbol("AIR")).toBeTruthy();
    expect(getEntitySymbol("SURFACE")).toBeTruthy();
    expect(getEntitySymbol("SUBSURFACE")).toBeTruthy();
    expect(getEntitySymbol("LAND")).toBeTruthy();
    expect(getEntitySymbol("CYBER")).toBeTruthy();
  });

  it("returns fallback for unknown type", () => {
    expect(getEntitySymbol("UNKNOWN" as "AIR")).toBe("●");
  });
});

describe("getEntityShape", () => {
  it("AIR returns triangle", () => {
    expect(getEntityShape("AIR")).toBe("triangle");
  });

  it("SURFACE returns diamond", () => {
    expect(getEntityShape("SURFACE")).toBe("diamond");
  });

  it("SUBSURFACE returns circle", () => {
    expect(getEntityShape("SUBSURFACE")).toBe("circle");
  });

  it("LAND returns square", () => {
    expect(getEntityShape("LAND")).toBe("square");
  });

  it("CYBER returns cross", () => {
    expect(getEntityShape("CYBER")).toBe("cross");
  });
});
