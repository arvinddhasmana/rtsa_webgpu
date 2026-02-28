// CLASSIFICATION: UNCLASSIFIED
import React from "react";
import { ActiveRole, DashboardView, useUIStore } from "../../stores/uiStore";

// Available views per role
const ROLE_VIEWS: Record<ActiveRole, { value: DashboardView; label: string }[]> = {
  commander: [
    { value: "fusion", label: "Fusion Dashboard" },
    { value: "multi-domain", label: "Multi-Domain Dashboard" },
    { value: "operator", label: "Operator UI" },
  ],
  analyst: [
    { value: "forensics", label: "Intelligence Forensics" },
    { value: "operator", label: "Operator UI" },
  ],
  security: [
    { value: "audit", label: "Audit & Feedback" },
  ],
  sensor_operator: [
    { value: "sensor-health", label: "Sensor Health" },
  ],
  nato_liaison: [
    { value: "nato-exchange", label: "NATO Exchange" },
  ],
};

export const DashboardSelector: React.FC = () => {
  const activeRole = useUIStore((s) => s.activeRole);
  const activeDashboardView = useUIStore((s) => s.activeDashboardView);
  const setDashboardView = useUIStore((s) => s.setDashboardView);

  const availableViews = ROLE_VIEWS[activeRole] || [];

  // Auto-switch view if the current view is not available for the new role
  React.useEffect(() => {
    if (availableViews.length > 0 && !availableViews.some(v => v.value === activeDashboardView)) {
      setDashboardView(availableViews[0].value);
    }
  }, [activeRole, activeDashboardView, setDashboardView, availableViews]);

  if (availableViews.length <= 1) {
    return null; // Don't show selector if 0 or 1 choice
  }

  return (
    <select
      data-testid="dashboard-selector"
      aria-label="Dashboard View Selector"
      value={activeDashboardView}
      onChange={(e) => setDashboardView(e.target.value as DashboardView)}
      style={{
        padding: "4px 8px",
        backgroundColor: "var(--glass-bg, #1E293B)",
        color: "var(--color-accent-amber, #f59e0b)",
        border: "1px solid var(--color-accent-amber, #f59e0b)",
        borderRadius: "4px",
        cursor: "pointer",
        fontSize: "0.75rem",
        fontWeight: "bold",
      }}
    >
      {availableViews.map((view) => (
        <option key={view.value} value={view.value} style={{ backgroundColor: "#1E293B", color: "#F1F5F9" }}>
          {view.label}
        </option>
      ))}
    </select>
  );
};
