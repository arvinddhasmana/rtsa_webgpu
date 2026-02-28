// CLASSIFICATION: UNCLASSIFIED
// src/components/layout/RoleSelector.tsx

import React from "react";
import { ActiveRole, useUIStore } from "../../stores/uiStore";

const ROLES: { value: ActiveRole; label: string }[] = [
  { value: "commander", label: "Operations Commander" },
  { value: "security", label: "Security Officer" },
  { value: "analyst", label: "Intelligence Analyst" },
  { value: "sensor_operator", label: "Sensor Operator" },
  { value: "nato_liaison", label: "NATO Liaison" },
];

/**
 * RoleSelector — dropdown to select the active role for the dashboard layout.
 */
export const RoleSelector: React.FC = () => {
  const activeRole = useUIStore((s) => s.activeRole);
  const setActiveRole = useUIStore((s) => s.setActiveRole);

  return (
    <select
      data-testid="role-selector"
      aria-label="Role selector"
      value={activeRole}
      onChange={(e) => setActiveRole(e.target.value as ActiveRole)}
      style={{
        padding: "4px 8px",
        backgroundColor: "#374151",
        color: "#F1F5F9",
        border: "1px solid #4B5563",
        borderRadius: "4px",
        cursor: "pointer",
        fontSize: "0.75rem",
      }}
    >
      {ROLES.map((role) => (
        <option key={role.value} value={role.value}>
          {role.label}
        </option>
      ))}
    </select>
  );
};
