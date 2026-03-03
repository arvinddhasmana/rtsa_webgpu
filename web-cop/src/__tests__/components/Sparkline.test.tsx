// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/Sparkline.test.tsx

import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Sparkline } from "../../components/sensor/Sparkline";

describe("Sparkline", () => {
  it("renders nothing when data has fewer than 2 points", () => {
    const { container } = render(<Sparkline data={[5]} />);
    expect(container.firstChild).toBeNull();
  });

  it("renders an SVG when data has 2+ points", () => {
    render(<Sparkline data={[1, 2, 3, 4, 5]} />);
    expect(document.querySelector("svg")).toBeTruthy();
  });

  it("renders a polyline inside the SVG", () => {
    render(<Sparkline data={[1, 2, 3]} />);
    const polyline = document.querySelector("polyline");
    expect(polyline).toBeTruthy();
  });

  it("sets the accessible aria-label with last value", () => {
    render(<Sparkline data={[1.0, 2.0, 3.5]} />);
    expect(screen.getByRole("img")).toHaveAttribute(
      "aria-label",
      "Rate trend: 3.5 events/sec"
    );
  });

  it("respects custom width and height", () => {
    render(<Sparkline data={[1, 2, 3]} width={80} height={30} />);
    const svg = document.querySelector("svg");
    expect(svg?.getAttribute("width")).toBe("80");
    expect(svg?.getAttribute("height")).toBe("30");
  });

  it("handles a flat data series without error", () => {
    expect(() => render(<Sparkline data={[5, 5, 5, 5]} />)).not.toThrow();
  });

  it("handles data with all zeros", () => {
    expect(() => render(<Sparkline data={[0, 0, 0]} />)).not.toThrow();
  });
});
