// CLASSIFICATION: UNCLASSIFIED
// tests/components/ClassificationBanner.test.tsx

import { describe, it, expect } from "vitest";
import { render, screen } from "@solidjs/testing-library";
import { ClassificationBanner } from "../../src/components/shell/ClassificationBanner";

describe("ClassificationBanner", () => {
  it("renders with default UNCLASSIFIED text", () => {
    render(() => <ClassificationBanner />);
    expect(screen.getByRole("banner").textContent).toBe("UNCLASSIFIED");
  });

  it("has correct aria-label", () => {
    render(() => <ClassificationBanner />);
    const banner = screen.getByRole("banner");
    expect(banner.getAttribute("aria-label")).toContain("UNCLASSIFIED");
  });
});
