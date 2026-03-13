// CLASSIFICATION: UNCLASSIFIED
// tests/components/ObsPerSecChart.test.tsx

import { render, screen } from "@solidjs/testing-library";
import { describe, expect, it } from "vitest";
import { ObsPerSecChart } from "../../src/components/dashboard/ObsPerSecChart";

const history = [10, 20, 30, 40, 50, 60, 70, 80, 90, 100];

describe("ObsPerSecChart", () => {
  it("renders chart container", () => {
    render(() => <ObsPerSecChart history={history} />);
    expect(screen.getByTestId("obs-per-sec-chart")).toBeDefined();
  });

  it("renders SVG polyline with correct number of points", () => {
    render(() => <ObsPerSecChart history={history} />);
    const polyline = screen.getByTestId("ops-polyline");
    // 10 data points → 10 coordinate pairs
    const pts = polyline.getAttribute("points")!.trim().split(" ");
    expect(pts.length).toBe(10);
  });

  it("shows avg label", () => {
    render(() => <ObsPerSecChart history={history} />);
    expect(screen.getByText(/Avg:/)).toBeDefined();
  });

  it("shows peak label", () => {
    render(() => <ObsPerSecChart history={history} />);
    expect(screen.getByText(/Peak:/)).toBeDefined();
  });

  it("computes correct avg (55 for 10..100)", () => {
    render(() => <ObsPerSecChart history={history} />);
    // Avg of [10,20,30,40,50,60,70,80,90,100] = 55
    expect(screen.getByText(/55 OPS/)).toBeDefined();
  });

  it("computes correct peak (100)", () => {
    render(() => <ObsPerSecChart history={history} />);
    expect(screen.getByText(/100 OPS/)).toBeDefined();
  });

  it("renders 'No data' when history has fewer than 2 points", () => {
    render(() => <ObsPerSecChart history={[42]} />);
    expect(screen.getByText("No data")).toBeDefined();
  });

  it("renders empty chart without crashing", () => {
    render(() => <ObsPerSecChart history={[]} />);
    expect(screen.getByTestId("obs-per-sec-chart")).toBeDefined();
  });
});
