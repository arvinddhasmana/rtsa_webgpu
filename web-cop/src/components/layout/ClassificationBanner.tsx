// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/ClassificationBanner.tsx

import React from "react";
import { ClassificationLevel } from "../../types/common";
import { getClassificationStyle } from "../../utils/classification";

interface ClassificationBannerProps {
  level: ClassificationLevel;
  position: "top" | "bottom";
}

/**
 * ClassificationBanner — displayed at TOP and BOTTOM of viewport at all times.
 * Shows the highest classification level of data currently displayed.
 *
 * Colors follow Government of Canada classification marking:
 *   UNCLASSIFIED     → green (#008000) with white text
 *   PROTECTED A/B    → blue (#0000FF) with white text
 *   PROTECTED C      → red (#FF0000) with white text
 *   SECRET           → red (#FF0000) with white text
 *
 * Banner MUST be visible at all times — never hidden by scroll or overlay.
 */
export const ClassificationBanner: React.FC<ClassificationBannerProps> = ({
  level,
  position,
}) => {
  const style = getClassificationStyle(level);

  const positionStyle: React.CSSProperties =
    position === "top"
      ? { position: "fixed", top: 0, left: 0, right: 0, zIndex: 9999 }
      : { position: "fixed", bottom: 0, left: 0, right: 0, zIndex: 9999 };

  return (
    <div
      data-testid={`classification-banner-${position}`}
      role="banner"
      aria-label={`Classification: ${style.label}`}
      style={{
        ...positionStyle,
        backgroundColor: style.backgroundColor,
        color: style.textColor,
        textAlign: "center",
        fontWeight: "bold",
        fontSize: "0.875rem",
        letterSpacing: "0.1em",
        padding: "4px 0",
        userSelect: "none",
      }}
    >
      {`CLASSIFICATION: ${style.label}`}
    </div>
  );
};
