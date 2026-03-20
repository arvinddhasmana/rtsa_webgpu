// CLASSIFICATION: UNCLASSIFIED
// tests/components/TrackSymbol.test.tsx — Unit tests for the TrackSymbol SVG component
//
// Verifies correct rendering of shapes, colours, and modifiers for all
// TrackDomain, TrackAffiliation, and TrackContext combinations.
//
// Reference: docs/sdlc_guidelines/05_testing/testing_strategy.md

import { render, screen } from "@solidjs/testing-library";
import { afterEach, describe, expect, it } from "vitest";
import { cleanup } from "@solidjs/testing-library";
import { TrackSymbol } from "../../src/components/symbols/TrackSymbol";
import {
  TrackAffiliation,
  TrackContext,
  TrackDomain,
} from "../../src/types/track-symbol";

afterEach(() => {
  cleanup();
});

// ── Rendering sanity ──────────────────────────────────────────────────────────

describe("TrackSymbol — renders an SVG element", () => {
  it("renders an svg with role=img for a hostile air track", () => {
    render(() => (
      <TrackSymbol
        domain={TrackDomain.AIR}
        affiliation={TrackAffiliation.HOSTILE}
        context={TrackContext.REAL}
      />
    ));
    const svg = screen.getByRole("img");
    expect(svg).toBeDefined();
    expect(svg.tagName.toLowerCase()).toBe("svg");
  });

  it("uses the default size (32) when size prop is omitted", () => {
    render(() => (
      <TrackSymbol
        domain={TrackDomain.SURFACE}
        affiliation={TrackAffiliation.FRIENDLY}
        context={TrackContext.REAL}
      />
    ));
    const svg = screen.getByRole("img");
    expect(svg.getAttribute("width")).toBe("32");
    expect(svg.getAttribute("height")).toBe("32");
  });

  it("respects a custom size prop", () => {
    render(() => (
      <TrackSymbol
        domain={TrackDomain.LAND}
        affiliation={TrackAffiliation.NEUTRAL}
        context={TrackContext.REAL}
        size={48}
      />
    ));
    const svg = screen.getByRole("img");
    expect(svg.getAttribute("width")).toBe("48");
    expect(svg.getAttribute("height")).toBe("48");
  });
});

// ── aria-label reflects domain + affiliation ──────────────────────────────────

describe("TrackSymbol — aria-label", () => {
  const cases: [TrackDomain, TrackAffiliation, string][] = [
    [TrackDomain.AIR,        TrackAffiliation.HOSTILE,  "Hostile Air track"],
    [TrackDomain.SURFACE,    TrackAffiliation.FRIENDLY, "Friendly Surface track"],
    [TrackDomain.SUBSURFACE, TrackAffiliation.SUSPECT,  "Suspect Subsurface track"],
    [TrackDomain.LAND,       TrackAffiliation.NEUTRAL,  "Neutral Land track"],
    [TrackDomain.SPACE,      TrackAffiliation.UNKNOWN,  "Unknown Space track"],
    [TrackDomain.CYBER,      TrackAffiliation.PENDING,  "Pending Cyber track"],
  ];

  for (const [domain, affiliation, expected] of cases) {
    it(`aria-label is "${expected}"`, () => {
      render(() => (
        <TrackSymbol domain={domain} affiliation={affiliation} context={TrackContext.REAL} />
      ));
      const svg = screen.getByRole("img");
      expect(svg.getAttribute("aria-label")).toBe(expected);
    });
  }
});

// ── Affiliation colours ───────────────────────────────────────────────────────

describe("TrackSymbol — affiliation fill colours", () => {
  const affiliationFills: [TrackAffiliation, string][] = [
    [TrackAffiliation.UNKNOWN,  "#f59e0b"],
    [TrackAffiliation.PENDING,  "#06b6d4"],
    [TrackAffiliation.FRIENDLY, "#3b82f6"],
    [TrackAffiliation.NEUTRAL,  "#22c55e"],
    [TrackAffiliation.SUSPECT,  "#f97316"],
    [TrackAffiliation.HOSTILE,  "#ef4444"],
  ];

  for (const [affiliation, expectedFill] of affiliationFills) {
    it(`affiliation ${TrackAffiliation[affiliation]} uses fill ${expectedFill}`, () => {
      const { container } = render(() => (
        <TrackSymbol
          domain={TrackDomain.AIR}
          affiliation={affiliation}
          context={TrackContext.REAL}
        />
      ));
      // Find the first shape element (path or ellipse or circle) inside the svg
      const shape =
        container.querySelector("path") ??
        container.querySelector("ellipse") ??
        container.querySelector("circle");
      expect(shape).not.toBeNull();
      expect(shape!.getAttribute("fill")).toBe(expectedFill);
    });
  }
});

// ── Context outline modifiers ─────────────────────────────────────────────────

describe("TrackSymbol — context stroke-dasharray", () => {
  it("REAL context renders solid outline (no stroke-dasharray)", () => {
    const { container } = render(() => (
      <TrackSymbol
        domain={TrackDomain.AIR}
        affiliation={TrackAffiliation.HOSTILE}
        context={TrackContext.REAL}
      />
    ));
    const shape = container.querySelector("path");
    // No dasharray attribute set for REAL context
    expect(shape?.getAttribute("stroke-dasharray")).toBeNull();
  });

  it("EXERCISE context renders dashed outline", () => {
    const { container } = render(() => (
      <TrackSymbol
        domain={TrackDomain.AIR}
        affiliation={TrackAffiliation.HOSTILE}
        context={TrackContext.EXERCISE}
      />
    ));
    const shape = container.querySelector("path");
    const dashArray = shape?.getAttribute("stroke-dasharray");
    expect(dashArray).not.toBeNull();
    expect(dashArray).toMatch(/\d/); // contains at least one digit
  });

  it("SIMULATION context sets fill-opacity to 0.35 (semi-transparent)", () => {
    const { container } = render(() => (
      <TrackSymbol
        domain={TrackDomain.AIR}
        affiliation={TrackAffiliation.NEUTRAL}
        context={TrackContext.SIMULATION}
      />
    ));
    const g = container.querySelector("g");
    expect(g?.getAttribute("fill-opacity")).toBe("0.35");
  });

  it("TEST context renders a crosshair overlay", () => {
    const { container } = render(() => (
      <TrackSymbol
        domain={TrackDomain.LAND}
        affiliation={TrackAffiliation.NEUTRAL}
        context={TrackContext.TEST}
      />
    ));
    // Two <line> elements are added for the crosshair
    const lines = container.querySelectorAll("line");
    expect(lines.length).toBeGreaterThanOrEqual(2);
  });
});

// ── Selection ring ────────────────────────────────────────────────────────────

describe("TrackSymbol — selection ring", () => {
  it("does NOT render a selection ring when selected=false", () => {
    const { container } = render(() => (
      <TrackSymbol
        domain={TrackDomain.AIR}
        affiliation={TrackAffiliation.HOSTILE}
        context={TrackContext.REAL}
        selected={false}
      />
    ));
    // Selection ring is a circle with stroke="#00ffff"
    const rings = Array.from(container.querySelectorAll("circle")).filter(
      (el) => el.getAttribute("stroke") === "#00ffff",
    );
    expect(rings.length).toBe(0);
  });

  it("renders a cyan selection ring when selected=true", () => {
    const { container } = render(() => (
      <TrackSymbol
        domain={TrackDomain.AIR}
        affiliation={TrackAffiliation.HOSTILE}
        context={TrackContext.REAL}
        selected={true}
      />
    ));
    const rings = Array.from(container.querySelectorAll("circle")).filter(
      (el) => el.getAttribute("stroke") === "#00ffff",
    );
    expect(rings.length).toBe(1);
  });
});

// ── All six domains render without errors ─────────────────────────────────────

describe("TrackSymbol — all domains render", () => {
  const domains = [
    TrackDomain.AIR,
    TrackDomain.SURFACE,
    TrackDomain.SUBSURFACE,
    TrackDomain.LAND,
    TrackDomain.SPACE,
    TrackDomain.CYBER,
  ];

  for (const domain of domains) {
    it(`domain ${TrackDomain[domain]} renders without throwing`, () => {
      expect(() =>
        render(() => (
          <TrackSymbol
            domain={domain}
            affiliation={TrackAffiliation.UNKNOWN}
            context={TrackContext.REAL}
          />
        )),
      ).not.toThrow();

      const svg = screen.getByRole("img");
      expect(svg).toBeDefined();
    });
  }
});
