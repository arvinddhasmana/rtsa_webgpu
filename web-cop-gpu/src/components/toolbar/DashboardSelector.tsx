// CLASSIFICATION: UNCLASSIFIED
// src/components/toolbar/DashboardSelector.tsx
//
// Allows the operator to switch between dashboard views (Sensor / Commander / Analytics).
// Reference: docs/implementation/v4/phase3_ui_interaction.md §3 U3-2, U3-9

import {
  dashboard,
  setDashboard,
  type Dashboard,
} from "../../signals/viewport";

const DASHBOARDS: { value: Dashboard; label: string }[] = [
  { value: "sensor", label: "Map view" },
  { value: "health", label: "Health" },
  { value: "commander", label: "Commander" },
  { value: "analytics", label: "Analytics" },
  { value: "coverage", label: "Coverage" },
];

/** Never destructure props — breaks SolidJS reactivity. */
export function DashboardSelector() {
  function handleChange(event: Event) {
    const target = event.currentTarget as HTMLSelectElement;
    setDashboard(target.value as Dashboard);
  }

  return (
    <div
      data-testid="dashboard-selector"
      style={{
        display: "flex",
        "flex-direction": "row",
        "align-items": "center",
        gap: "6px",
      }}
    >
      <label
        for="dashboard-selector"
        style={{
          "font-size": "0.65rem",
          "text-transform": "uppercase",
          "letter-spacing": "0.08em",
          color: "#94a3b8",
          "white-space": "nowrap",
        }}
      >
        Dashboard
      </label>
      <select
        id="dashboard-selector"
        value={dashboard()}
        onChange={handleChange}
        style={{
          width: "auto",
          background: "#1e2a3a",
          color: "#e2e8f0",
          border: "1px solid #2d3f56",
          "border-radius": "4px",
          padding: "0.25rem 0.5rem",
          "font-size": "0.8rem",
          cursor: "pointer",
        }}
      >
        {DASHBOARDS.map((d) => (
          <option value={d.value}>{d.label}</option>
        ))}
      </select>
    </div>
  );
}
