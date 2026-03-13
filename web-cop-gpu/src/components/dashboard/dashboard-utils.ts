// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/dashboard-utils.ts — Shared helper utilities for dashboard components
//
// Reference: docs/implementation/v5/sensordashboard_three_level_plan.md

/** Maps sensor connection status to a colour token shared across all dashboard components. */
export function statusColor(status: string): string {
  switch (status) {
    case "CONNECTED": return "#4ade80";
    case "STALE": return "#fbbf24";
    default: return "#f87171";
  }
}
