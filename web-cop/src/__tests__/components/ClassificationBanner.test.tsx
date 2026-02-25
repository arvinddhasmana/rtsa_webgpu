// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/components/ClassificationBanner.test.tsx

import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { ClassificationBanner } from "../../components/layout/ClassificationBanner";

describe("ClassificationBanner", () => {
  it("T08: renders UNCLASSIFIED with green background", () => {
    render(<ClassificationBanner level="UNCLASSIFIED" position="top" />);
    const banner = screen.getByTestId("classification-banner-top");
    expect(banner).toHaveTextContent("CLASSIFICATION: UNCLASSIFIED");
    expect(banner).toHaveStyle({ backgroundColor: "#008000" });
  });

  it("T08: renders PROTECTED_B with blue background", () => {
    render(<ClassificationBanner level="PROTECTED_B" position="top" />);
    const banner = screen.getByTestId("classification-banner-top");
    expect(banner).toHaveTextContent("CLASSIFICATION: PROTECTED B");
    expect(banner).toHaveStyle({ backgroundColor: "#0000FF" });
  });

  it("T08: renders SECRET with red background", () => {
    render(<ClassificationBanner level="SECRET" position="bottom" />);
    const banner = screen.getByTestId("classification-banner-bottom");
    expect(banner).toHaveTextContent("CLASSIFICATION: SECRET");
    expect(banner).toHaveStyle({ backgroundColor: "#FF0000" });
  });

  it("T08: renders PROTECTED_C with red background", () => {
    render(<ClassificationBanner level="PROTECTED_C" position="top" />);
    const banner = screen.getByTestId("classification-banner-top");
    expect(banner).toHaveStyle({ backgroundColor: "#FF0000" });
  });

  it("renders PROTECTED_A with blue background", () => {
    render(<ClassificationBanner level="PROTECTED_A" position="top" />);
    const banner = screen.getByTestId("classification-banner-top");
    expect(banner).toHaveStyle({ backgroundColor: "#0000FF" });
  });

  it("renders at top with fixed position style", () => {
    render(<ClassificationBanner level="UNCLASSIFIED" position="top" />);
    const banner = screen.getByTestId("classification-banner-top");
    expect(banner).toHaveStyle({ position: "fixed", top: "0" });
  });

  it("renders at bottom with fixed position style", () => {
    render(<ClassificationBanner level="UNCLASSIFIED" position="bottom" />);
    const banner = screen.getByTestId("classification-banner-bottom");
    expect(banner).toHaveStyle({ position: "fixed", bottom: "0" });
  });

  it("has role=banner for accessibility", () => {
    render(<ClassificationBanner level="UNCLASSIFIED" position="top" />);
    const banner = screen.getByRole("banner");
    expect(banner).toBeTruthy();
  });
});
