// CLASSIFICATION: UNCLASSIFIED
// src/components/audit/TrustScoreHistogram.tsx
//
// SVG histogram of operator trust score distribution.
// Bins: 0-20%, 20-40%, 40-60%, 60-80%, 80-100%.

import React, { useMemo } from "react";
import type { FeedbackLogEntry } from "./FeedbackLogTable";

interface TrustScoreHistogramProps {
  entries: FeedbackLogEntry[];
  height?: number;
}

const BINS = [
  { label: "0-20%",  min: 0.0, max: 0.2, color: "#EF4444" },
  { label: "20-40%", min: 0.2, max: 0.4, color: "#F97316" },
  { label: "40-60%", min: 0.4, max: 0.6, color: "#F59E0B" },
  { label: "60-80%", min: 0.6, max: 0.8, color: "#10B981" },
  { label: "80-100%",min: 0.8, max: 1.01,color: "#22C55E" },
];

/**
 * TrustScoreHistogram — SVG bar chart of trust score distribution
 * across operators' historical feedback submissions.
 */
export const TrustScoreHistogram: React.FC<TrustScoreHistogramProps> = ({
  entries,
  height = 100,
}) => {
  const counts = useMemo(() => {
    return BINS.map((b) => ({
      ...b,
      count: entries.filter(
        (e) => e.trustScore >= b.min && e.trustScore < b.max
      ).length,
    }));
  }, [entries]);

  const maxCount = Math.max(...counts.map((c) => c.count), 1);

  const barWidth = 40;
  const gap = 8;
  const paddingX = 8;
  const paddingY = 8;
  const labelH = 20;
  const chartH = height - paddingY * 2 - labelH;
  const totalW = BINS.length * (barWidth + gap) - gap + paddingX * 2;

  return (
    <div data-testid="trust-score-histogram" style={{ width: "100%" }}>
      <svg
        width="100%"
        viewBox={`0 0 ${totalW} ${height}`}
        aria-label="Trust score distribution"
        role="img"
      >
        {counts.map((bin, idx) => {
          const barH = Math.max(2, (bin.count / maxCount) * chartH);
          const x = paddingX + idx * (barWidth + gap);
          const y = paddingY + chartH - barH;

          return (
            <g key={bin.label} data-testid={`trust-bin-${idx}`}>
              <rect
                x={x}
                y={y}
                width={barWidth}
                height={barH}
                fill={bin.color}
                opacity={0.85}
                rx={2}
              />
              {bin.count > 0 && (
                <text
                  x={x + barWidth / 2}
                  y={barH > 14 ? y + 12 : y - 3}
                  textAnchor="middle"
                  fontSize="9"
                  fontFamily="monospace"
                  fill="#F1F5F9"
                >
                  {bin.count}
                </text>
              )}
              <text
                x={x + barWidth / 2}
                y={paddingY + chartH + labelH}
                textAnchor="middle"
                fontSize="8"
                fontFamily="monospace"
                fill={bin.color}
                fontWeight="bold"
              >
                {bin.label}
              </text>
            </g>
          );
        })}
      </svg>
    </div>
  );
};
