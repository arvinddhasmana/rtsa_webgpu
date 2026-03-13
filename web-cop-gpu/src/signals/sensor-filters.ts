// CLASSIFICATION: UNCLASSIFIED
// src/signals/sensor-filters.ts — Signals for dashboard filtering
//
// Reference: docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md

import { createSignal } from "solid-js";
import type { SensorStatus } from "../services/sensor-health";

/** Active health status filters. */
export const [selectedStatuses, setSelectedStatuses] = createSignal<string[]>([
  "CONNECTED",
  "STALE",
  "OFFLINE",
]);

/** Active sensor type filters. */
export const [selectedTypes, setSelectedTypes] = createSignal<string[]>([
  "RADAR",
  "EW/SIGINT",
  "ELINT/COMINT",
  "ISR",
  "AIS/BFT",
  "CYBER",
]);

/** Sidebar collapse state. */
export const [sidebarCollapsed, setSidebarCollapsed] = createSignal(false);

/** Currently selected sensor for deep diagnostic drill-down. Null = list view. */
export const [selectedSensor, setSelectedSensor] =
  createSignal<SensorStatus | null>(null);

/** Card view mode: full shows rich dual-sparkline detail cards, compact shows condensed metric cards. */
export type CardView = "full" | "compact";
export const [cardView, setCardView] = createSignal<CardView>("full");

/** Toggle a status filter. */
export function toggleStatusFilter(status: string) {
  const current = selectedStatuses();
  if (current.includes(status)) {
    if (current.length > 1) {
      // Don't allow empty filters
      setSelectedStatuses(current.filter((s) => s !== status));
    }
  } else {
    setSelectedStatuses([...current, status]);
  }
}

/** Toggle a type filter. */
export function toggleTypeFilter(type: string) {
  const current = selectedTypes();
  if (current.includes(type)) {
    if (current.length > 1) {
      setSelectedTypes(current.filter((t) => t !== type));
    }
  } else {
    setSelectedTypes([...current, type]);
  }
}
