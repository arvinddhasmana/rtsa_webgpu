// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/utils/coordinates.test.ts

import { describe, it, expect } from "vitest";
import { toDMS, formatPosition, knotsToKmh } from "../../utils/coordinates";

describe("toDMS", () => {
  it("converts positive latitude to DMS N", () => {
    const result = toDMS(45.5, true);
    expect(result).toContain("N");
    expect(result).toContain("45°");
  });

  it("converts negative latitude to DMS S", () => {
    const result = toDMS(-33.8, true);
    expect(result).toContain("S");
  });

  it("converts positive longitude to DMS E", () => {
    const result = toDMS(10.0, false);
    expect(result).toContain("E");
  });

  it("converts negative longitude to DMS W", () => {
    const result = toDMS(-60.0, false);
    expect(result).toContain("W");
    expect(result).toContain("60°");
  });

  it("handles zero", () => {
    const result = toDMS(0, true);
    expect(result).toContain("N");
    expect(result).toContain("0°");
  });
});

describe("formatPosition", () => {
  it("formats position as DMS pair", () => {
    const result = formatPosition(45.0, -60.0);
    expect(result).toContain("N");
    expect(result).toContain("W");
  });
});

describe("knotsToKmh", () => {
  it("converts 1 knot to 1.852 km/h", () => {
    expect(knotsToKmh(1)).toBeCloseTo(1.852);
  });

  it("converts 0 knots", () => {
    expect(knotsToKmh(0)).toBe(0);
  });

  it("converts 10 knots", () => {
    expect(knotsToKmh(10)).toBeCloseTo(18.52);
  });
});
