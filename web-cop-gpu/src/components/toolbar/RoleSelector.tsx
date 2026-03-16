// CLASSIFICATION: UNCLASSIFIED
// src/components/toolbar/RoleSelector.tsx
//
// Switches the active operator role across all supported RTSA operator roles.
// Role changes filter UI panel visibility without interrupting data flows.
// Reference: docs/implementation/v4/phase3_ui_interaction.md §3 U3-2, U3-9

import { role, setRole, type Role } from "../../signals/viewport";

interface RoleSelectorProps {
  /** Optional callback fired after the role changes. */
  onChange?: (newRole: Role) => void;
}

const ROLES: { value: Role; label: string }[] = [
  { value: "operations_commander", label: "Ops Commander" },
  { value: "intelligence_analyst", label: "Intelligence Analyst" },
  { value: "security_officer", label: "Security Officer" },
  { value: "sensor_operator", label: "Sensor Operator" },
  { value: "nato_liaison", label: "NATO Liaison" },
];

/** Never destructure props — breaks SolidJS reactivity. */
export function RoleSelector(props: RoleSelectorProps) {
  function handleChange(event: Event) {
    const target = event.currentTarget as HTMLSelectElement;
    const newRole = target.value as Role;
    setRole(newRole);
    props.onChange?.(newRole);
  }

  return (
    <div
      data-testid="role-selector"
      style={{
        display: "flex",
        "flex-direction": "row",
        "align-items": "center",
        gap: "6px",
      }}
    >
      <label
        for="role-selector"
        style={{
          "font-size": "0.65rem",
          "text-transform": "uppercase",
          "letter-spacing": "0.08em",
          color: "#94a3b8",
          "white-space": "nowrap",
        }}
      >
        Role
      </label>
      <select
        id="role-selector"
        value={role()}
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
        {ROLES.map((r) => (
          <option value={r.value}>{r.label}</option>
        ))}
      </select>
    </div>
  );
}
