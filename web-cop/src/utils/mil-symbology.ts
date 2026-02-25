// CLASSIFICATION: UNCLASSIFIED
// src/utils/mil-symbology.ts

import { HostileClassification, EntityType } from "../types/common";

/**
 * Returns the marker color for a hostile classification per MIL-STD-2525.
 */
export function getHostileColor(hostile: HostileClassification): string {
  switch (hostile) {
    case "HOSTILE":
      return "#DC2626";
    case "FRIENDLY":
      return "#2563EB";
    case "NEUTRAL":
      return "#16A34A";
    case "UNKNOWN":
      return "#CA8A04";
    default:
      return "#6B7280";
  }
}

/**
 * Returns a MIL-STD-2525 symbol code character for the entity type.
 */
export function getEntitySymbol(entityType: EntityType): string {
  switch (entityType) {
    case "AIR":
      return "✈";
    case "SURFACE":
      return "⛵";
    case "SUBSURFACE":
      return "🔱";
    case "LAND":
      return "⊕";
    case "CYBER":
      return "⚡";
    default:
      return "●";
  }
}

/**
 * Returns the shape name for the entity type per MIL-STD-2525.
 */
export function getEntityShape(entityType: EntityType): string {
  switch (entityType) {
    case "AIR":
      return "triangle";
    case "SURFACE":
      return "diamond";
    case "SUBSURFACE":
      return "circle";
    case "LAND":
      return "square";
    case "CYBER":
      return "cross";
    default:
      return "circle";
  }
}
