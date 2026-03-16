// CLASSIFICATION: UNCLASSIFIED
// src/signals/viewport.ts — Role, dashboard, and camera viewport signals
//
// Reference: docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md §3.3

import { createSignal } from "solid-js";

/** Operator roles that drive which UI panels are visible. */
export type Role =
  | "operations_commander"
  | "intelligence_analyst"
  | "security_officer"
  | "sensor_operator"
  | "nato_liaison";

/** Available dashboard views. */
export type Dashboard =
  | "sensor"
  | "commander"
  | "analytics"
  | "health"
  | "coverage";

/** Default dashboard each role lands on. */
export const ROLE_DEFAULT_DASHBOARD: Readonly<Record<Role, Dashboard>> = {
  operations_commander: "commander",
  intelligence_analyst: "analytics",
  security_officer: "commander",
  sensor_operator: "health",
  nato_liaison: "sensor",
};

/** Allowed dashboards per role. */
export const ROLE_ALLOWED_DASHBOARDS: Readonly<
  Record<Role, readonly Dashboard[]>
> = {
  operations_commander: ["commander", "coverage", "analytics"],
  intelligence_analyst: ["analytics", "sensor"],
  security_officer: ["commander"],
  sensor_operator: ["health", "coverage"],
  nato_liaison: ["sensor"],
};

/** Active operator role (default: sensor_operator). */
export const [role, setRoleSignal] = createSignal<Role>("sensor_operator");

/** Active dashboard view (default: health). */
export const [dashboard, setDashboardSignal] = createSignal<Dashboard>(
  ROLE_DEFAULT_DASHBOARD.sensor_operator,
);

/** Change role and deterministically land on the role's default dashboard. */
export function setRole(nextRole: Role): void {
  setRoleSignal(nextRole);
  setDashboardSignal(ROLE_DEFAULT_DASHBOARD[nextRole]);
}

/** Change dashboard with role policy enforcement. */
export function setDashboard(nextDashboard: Dashboard): void {
  const currentRole = role();
  const allowedDashboards = ROLE_ALLOWED_DASHBOARDS[currentRole];
  if (allowedDashboards.includes(nextDashboard)) {
    setDashboardSignal(nextDashboard);
    return;
  }
  setDashboardSignal(ROLE_DEFAULT_DASHBOARD[currentRole]);
}

/** Ensure current role/dashboard pair is valid; reset when invalid. */
export function enforceRoleDashboardGuard(): void {
  const currentRole = role();
  const currentDashboard = dashboard();
  if (!ROLE_ALLOWED_DASHBOARDS[currentRole].includes(currentDashboard)) {
    setDashboardSignal(ROLE_DEFAULT_DASHBOARD[currentRole]);
  }
}

/** Whether the search overlay is open. */
export const [searchOpen, setSearchOpen] = createSignal(false);

/** Whether the feedback form is open. */
export const [feedbackOpen, setFeedbackOpen] = createSignal(false);

/** Current map viewport. */
export interface Viewport {
  centerLat: number;
  centerLon: number;
  zoom: number;
}

export const [viewport, setViewport] = createSignal<Viewport>({
  centerLat: 0,
  centerLon: 0,
  zoom: 2,
});
