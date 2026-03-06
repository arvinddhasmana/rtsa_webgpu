// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/GeoFenceLayer.test.tsx

import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import type { GeoFence } from "../../components/map/GeoFenceLayer";
import {
  DEFAULT_DEMO_GEOFENCE,
  GeoFenceLayer,
} from "../../components/map/GeoFenceLayer";

describe("GeoFenceLayer", () => {
  it("renders the geofence-layer container", () => {
    render(<GeoFenceLayer fences={[]} />);
    expect(screen.getByTestId("geofence-layer")).toBeTruthy();
  });

  it("renders a div per fence", () => {
    const fences: GeoFence[] = [
      { id: "fence-1", name: "Test Zone", type: "inclusion" },
      { id: "fence-2", name: "Exclusion Zone", type: "exclusion" },
    ];
    render(<GeoFenceLayer fences={fences} />);
    expect(screen.getByTestId("geofence-fence-1")).toBeTruthy();
    expect(screen.getByTestId("geofence-fence-2")).toBeTruthy();
  });

  it("uses green fill for inclusion fences", () => {
    const fences: GeoFence[] = [
      { id: "inc", name: "Inclusion", type: "inclusion" },
    ];
    render(<GeoFenceLayer fences={fences} />);
    const el = screen.getByTestId("geofence-inc");
    // Should have green background-color (rgba(34, 197, 94, 0.15))
    expect(el.style.backgroundColor).toContain("34, 197, 94");
  });

  it("uses red fill for exclusion fences", () => {
    const fences: GeoFence[] = [
      { id: "exc", name: "Exclusion", type: "exclusion" },
    ];
    render(<GeoFenceLayer fences={fences} />);
    const el = screen.getByTestId("geofence-exc");
    expect(el.style.backgroundColor).toContain("239, 68, 68");
  });

  it("has dashed border style", () => {
    const fences: GeoFence[] = [
      { id: "dash", name: "Dashed", type: "inclusion" },
    ];
    render(<GeoFenceLayer fences={fences} />);
    const el = screen.getByTestId("geofence-dash");
    expect(el.style.border).toContain("dashed");
  });

  it("exports a default demo geofence constant", () => {
    expect(DEFAULT_DEMO_GEOFENCE.id).toBe("default-oparea");
    expect(DEFAULT_DEMO_GEOFENCE.type).toBe("inclusion");
    expect(DEFAULT_DEMO_GEOFENCE.bounds).toBeDefined();
  });
});
