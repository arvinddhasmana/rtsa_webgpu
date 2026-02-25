// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/utils/classification.test.ts

import { describe, it, expect } from "vitest";
import {
  getClassificationStyle,
  getHighestClassification,
  isAccessible,
} from "../../utils/classification";

describe("getClassificationStyle", () => {
  it("UNCLASSIFIED returns green", () => {
    const style = getClassificationStyle("UNCLASSIFIED");
    expect(style.backgroundColor).toBe("#008000");
    expect(style.label).toBe("UNCLASSIFIED");
  });

  it("PROTECTED_A returns blue", () => {
    const style = getClassificationStyle("PROTECTED_A");
    expect(style.backgroundColor).toBe("#0000FF");
    expect(style.label).toBe("PROTECTED A");
  });

  it("PROTECTED_B returns blue", () => {
    const style = getClassificationStyle("PROTECTED_B");
    expect(style.backgroundColor).toBe("#0000FF");
    expect(style.label).toBe("PROTECTED B");
  });

  it("PROTECTED_C returns red", () => {
    const style = getClassificationStyle("PROTECTED_C");
    expect(style.backgroundColor).toBe("#FF0000");
    expect(style.label).toBe("PROTECTED C");
  });

  it("SECRET returns red", () => {
    const style = getClassificationStyle("SECRET");
    expect(style.backgroundColor).toBe("#FF0000");
    expect(style.label).toBe("SECRET");
  });

  it("all styles have white text", () => {
    const levels = ["UNCLASSIFIED", "PROTECTED_A", "PROTECTED_B", "PROTECTED_C", "SECRET"] as const;
    for (const level of levels) {
      expect(getClassificationStyle(level).textColor).toBe("#FFFFFF");
    }
  });
});

describe("getHighestClassification", () => {
  it("returns UNCLASSIFIED for empty array", () => {
    expect(getHighestClassification([])).toBe("UNCLASSIFIED");
  });

  it("returns single item for single-element array", () => {
    expect(getHighestClassification(["PROTECTED_B"])).toBe("PROTECTED_B");
  });

  it("returns the highest level from mixed array", () => {
    expect(getHighestClassification(["UNCLASSIFIED", "SECRET", "PROTECTED_B"])).toBe("SECRET");
  });

  it("returns PROTECTED_C over PROTECTED_B", () => {
    expect(getHighestClassification(["PROTECTED_A", "PROTECTED_C", "PROTECTED_B"])).toBe("PROTECTED_C");
  });
});

describe("isAccessible", () => {
  it("returns true when clearance equals data classification", () => {
    expect(isAccessible("PROTECTED_B", "PROTECTED_B")).toBe(true);
  });

  it("returns true when clearance is higher", () => {
    expect(isAccessible("SECRET", "PROTECTED_C")).toBe(true);
  });

  it("returns false when clearance is lower", () => {
    expect(isAccessible("UNCLASSIFIED", "PROTECTED_A")).toBe(false);
  });
});
