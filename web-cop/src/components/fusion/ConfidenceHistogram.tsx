// CLASSIFICATION: UNCLASSIFIED
// src/components/fusion/ConfidenceHistogram.tsx
//
// SVG histogram showing confidence distribution across 4 bands:
//   HIGH (≥0.80), MEDIUM (0.60-0.79), LOW (0.40-0.59), TENTATIVE (<0.40)
//
// Used by FusionSidePanel and IntelSearchPanel.

import React, { useMemo } from "react";
import type { FusedTrack } from "../../types/track";

interface ConfidenceHistogramProps {
  tracks: FusedTrack[];
  /** Optional height of the SVG canvas (default 80) */
  height?: number;
}

const BANDS: Array<{
  label: string;
  color: string;
  min: number;
  max: number;
}> = [
  { label: "HIGH",      color: "#10B981", min: 0.80, max: 1.01 },
  { label: "MED",       color: "#3B82F6", min: 0.60, max: 0.80 },
  { label: "LOW",       color: "#F59E0B", min: 0.40, max: 0.60 },
  { label: "TENT",      color: "#EF4444", min: 0.00, max: 0.40 },
];

/**
 * ConfidenceHistogram — compact SVG bar chart of track confidence bands.
 *
 * Renders a 4-bar histogram with counts for HIGH / MEDIUM / LOW / TENTATIVE
 * confidence levels, colour-coded by band severity.
 */
export const ConfidenceHistogram: React.FC<ConfidenceHistogramProps> = ({
  tracks,
  height = 80,
}) => {
  const counts = useMemo(() => {
    return BANDS.map((b) => ({
      ...b,
      count: tracks.filter(
        (t) => t.confidenceScore >= b.min && t.confidenceScore < b.max,
      ).length,
    }));
  }, [tracks]);

  const maxCount = Math.max(...counts.map((c) => c.count), 1);

  const barWidth = 32;
  const gap = 8;
  const paddingX = 8;
  const paddingY = 8;
  const labelH = 14;
  const chartH = height - paddingY * 2 - labelH;
  const totalW = BANDS.length * (barWidth + gap) - gap + paddingX * 2;

  return (
    <div data-testid="confidence-histogram" style={{ width: "100%" }}>
      <svg
        width="100%"
        viewBox={`0 0 ${totalW} ${height}`}
        aria-label="Confidence distribution histogram"
        role="img"
      >
        {counts.map((band, idx) => {
          const barH = Math.max(2, (band.count / maxCount) * chartH);
          const x = paddingX + idx * (barWidth + gap);
          const y = paddingY + chartH - barH;

          return (
            <g key={band.label} data-testid={`histogram-bar-${band.label}`}>
              {/* Bar */}
              <rect
                x={x}
                y={y}
                width={barWidth}
                height={barH}
                fill={band.color}
                opacity={0.85}
                rx={2}
              />

              {/* Count label inside / above bar */}
              {band.count > 0 && (
                <text
                  x={x + barWidth / 2}
                  y={barH > 16 ? y + 12 : y - 3}
                  textAnchor="middle"
                  fontSize="9"
                  fontFamily="monospace"
                  fill="#F1F5F9"
                >
                  {band.count}
                </text>
              )}

              {/* Band label below bar */}
              <text
                x={x + barWidth / 2}
                y={paddingY + chartH + labelH}
                textAnchor="middle"
                fontSize="8"
                fontFamily="monospace"
                fill={band.color}
                fontWeight="bold"
              >
                {band.label}
              </text>
            </g>
          );
        })}
      </svg>
    </div>
  );
};
