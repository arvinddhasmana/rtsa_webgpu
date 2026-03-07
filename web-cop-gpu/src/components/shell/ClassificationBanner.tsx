// CLASSIFICATION: UNCLASSIFIED
// src/components/shell/ClassificationBanner.tsx
//
// Displays a persistent classification banner sourced from the Vite env variable
// VITE_CLASSIFICATION_LEVEL. Always rendered at the top and bottom of the viewport.
// Reference: docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md §7.2

import type { JSX } from "solid-js";

const LEVEL = import.meta.env.VITE_CLASSIFICATION_LEVEL ?? "UNCLASSIFIED";

const levelColors: Record<string, string> = {
  UNCLASSIFIED: "#006400",
  "PROTECTED A": "#005fa3",
  "PROTECTED B": "#005fa3",
  "PROTECTED C": "#7a0080",
  SECRET: "#cc0000",
};

const bannerColor = levelColors[LEVEL] ?? "#333";

const bannerStyle: JSX.CSSProperties = {
  background: bannerColor,
  color: "#fff",
  "text-align": "center",
  "font-size": "0.75rem",
  "font-weight": "bold",
  "letter-spacing": "0.1em",
  padding: "2px 0",
  "z-index": "1000",
  width: "100%",
  "flex-shrink": "0",
};

/** Static classification banner. Rendered at top and bottom of the layout. */
export function ClassificationBanner() {
  return (
    <div
      role="banner"
      data-testid="classification-banner-top"
      aria-label={`Classification level: ${LEVEL}`}
      style={bannerStyle}
    >
      {LEVEL}
    </div>
  );
}
