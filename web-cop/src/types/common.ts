// CLASSIFICATION: UNCLASSIFIED
// src/types/common.ts

export type ClassificationLevel =
  | "UNCLASSIFIED"
  | "PROTECTED_A"
  | "PROTECTED_B"
  | "PROTECTED_C"
  | "SECRET";

export type EntityType = "SURFACE" | "AIR" | "SUBSURFACE" | "LAND" | "CYBER";

export type HostileClassification =
  | "HOSTILE"
  | "FRIENDLY"
  | "NEUTRAL"
  | "UNKNOWN";

export type TrackStatus = "ACTIVE" | "STALE" | "DROPPED" | "MERGED";

export type AlertSeverity = "NORMAL" | "WATCH" | "ELEVATED" | "CRITICAL";

export type AnomalyType =
  | "SPEED"
  | "ROUTE_DEVIATION"
  | "AIS_MANIPULATION"
  | "BEHAVIORAL"
  | "TEMPORAL"
  | "PROXIMITY";

export type FeedbackType =
  | "CONFIRM_HOSTILE"
  | "CONFIRM_FRIENDLY"
  | "RECLASSIFY"
  | "REJECT_ANOMALY"
  | "CONFIRM_ANOMALY";
