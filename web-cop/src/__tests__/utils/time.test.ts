// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/utils/time.test.ts

import { describe, it, expect } from "vitest";
import { formatZulu, formatZuluTime, relativeTime } from "../../utils/time";

describe("formatZulu", () => {
  it("formats a date as ISO Zulu string", () => {
    const date = new Date("2026-02-25T04:30:00.000Z");
    const result = formatZulu(date);
    expect(result).toContain("2026-02-25");
    expect(result).toContain("Z");
  });
});

describe("formatZuluTime", () => {
  it("formats only the time portion", () => {
    const date = new Date("2026-02-25T14:30:45.000Z");
    const result = formatZuluTime(date);
    expect(result).toBe("14:30:45Z");
  });

  it("pads single-digit hours with zero", () => {
    const date = new Date("2026-02-25T04:05:09.000Z");
    const result = formatZuluTime(date);
    expect(result).toBe("04:05:09Z");
  });
});

describe("relativeTime", () => {
  it("shows seconds for recent dates", () => {
    const date = new Date(Date.now() - 30_000);
    expect(relativeTime(date)).toContain("s ago");
  });

  it("shows minutes for dates older than a minute", () => {
    const date = new Date(Date.now() - 3 * 60 * 1000);
    expect(relativeTime(date)).toContain("min ago");
  });

  it("shows hours for dates older than an hour", () => {
    const date = new Date(Date.now() - 3 * 60 * 60 * 1000);
    expect(relativeTime(date)).toContain("hr ago");
  });
});
