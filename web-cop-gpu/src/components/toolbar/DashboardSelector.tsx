// CLASSIFICATION: UNCLASSIFIED
// src/components/toolbar/DashboardSelector.tsx
//
// Allows the operator to switch among role-scoped dashboard views.
// Reference: docs/implementation/v4/phase3_ui_interaction.md §3 U3-2, U3-9

import {
  dashboard,
  role,
  ROLE_ALLOWED_DASHBOARDS,
  setDashboard,
  type Dashboard,
  type Role,
} from "../../signals/viewport";

const DASHBOARD_LABELS_BY_ROLE: Readonly<
  Record<Role, Readonly<Record<Dashboard, string>>>
> = {
  operations_commander: {
    commander: "Fusion",
    coverage: "Multi-Domain",
    analytics: "Operator UI",
    sensor: "Map View",
    health: "Sensor Health",
  },
  intelligence_analyst: {
    analytics: "Forensics",
    sensor: "Intel Search",
    commander: "Commander",
    health: "Health",
    coverage: "Coverage",
  },
  security_officer: {
    commander: "Audit & Feedback",
    sensor: "Map View",
    analytics: "Analytics",
    health: "Health",
    coverage: "Coverage",
  },
  sensor_operator: {
    health: "Sensor Health",
    coverage: "Coverage",
    sensor: "Map View",
    commander: "Commander",
    analytics: "Analytics",
  },
  nato_liaison: {
    sensor: "NATO Exchange",
    commander: "Commander",
    analytics: "Analytics",
    health: "Health",
    coverage: "Coverage",
  },
};

function dashboardsForRole(
  currentRole: Role,
): { value: Dashboard; label: string }[] {
  return ROLE_ALLOWED_DASHBOARDS[currentRole].map((value) => ({
    value,
    label: DASHBOARD_LABELS_BY_ROLE[currentRole][value],
  }));
}

/** Never destructure props — breaks SolidJS reactivity. */
export function DashboardSelector() {
  const visibleDashboards = () => dashboardsForRole(role());

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
        {visibleDashboards().map((d) => (
          <option value={d.value}>{d.label}</option>
        ))}
      </select>
    </div>
  );
}
