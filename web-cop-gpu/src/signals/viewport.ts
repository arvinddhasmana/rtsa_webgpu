// CLASSIFICATION: UNCLASSIFIED
// src/signals/viewport.ts — Role, dashboard, and camera viewport signals
//
// Reference: docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md §3.3

import { createSignal } from "solid-js";

/** Operator roles that drive which UI panels are visible. */
export type Role = "sensor_operator" | "operations_commander";

/** Available dashboard views. */
export type Dashboard = "sensor" | "commander" | "analytics" | "health";

/** Active operator role (default: sensor_operator). */
export const [role, setRole] = createSignal<Role>("sensor_operator");

/** Active dashboard view (default: health). */
export const [dashboard, setDashboard] = createSignal<Dashboard>("health");

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
