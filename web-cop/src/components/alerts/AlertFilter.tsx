// CLASSIFICATION: UNCLASSIFIED
// src/components/alerts/AlertFilter.tsx

import React from "react";
import { useAlertStore } from "../../stores/alertStore";

const SEVERITY_LEVELS = ["WATCH", "ELEVATED", "CRITICAL"] as const;

const SEVERITY_COLORS: Record<string, string> = {
  WATCH: "#CA8A04",
  ELEVATED: "#EA580C",
  CRITICAL: "#DC2626",
};

/**
 * AlertFilter — severity toggle buttons for filtering the alert list.
 */
export const AlertFilter: React.FC = () => {
  const minSeverityFilter = useAlertStore((s) => s.minSeverityFilter);
  const setMinSeverityFilter = useAlertStore((s) => s.setMinSeverityFilter);

  return (
    <div
      data-testid="alert-filter"
      style={{ display: "flex", gap: "4px", padding: "8px" }}
    >
      {SEVERITY_LEVELS.map((level) => {
        const isActive = minSeverityFilter === level;
        return (
          <button
            key={level}
            data-testid={`filter-${level.toLowerCase()}`}
            onClick={() => setMinSeverityFilter(level)}
            style={{
              flex: 1,
              padding: "4px 8px",
              fontSize: "0.7rem",
              fontWeight: "bold",
              backgroundColor: isActive ? SEVERITY_COLORS[level] : "#374151",
              color: "#F1F5F9",
              border: `1px solid ${SEVERITY_COLORS[level]}`,
              borderRadius: "4px",
              cursor: "pointer",
              letterSpacing: "0.05em",
            }}
          >
            {level}
          </button>
        );
      })}
    </div>
  );
};
