// CLASSIFICATION: UNCLASSIFIED
// src/types/alert.ts

import { AlertSeverity, AnomalyType, ClassificationLevel } from "./common";

export interface FeatureContribution {
  featureName: string;
  value: number;
  contributionWeight: number;
}

export interface AnomalyAlert {
  alertId: string;
  trackId: string;
  anomalyType: AnomalyType;
  severity: AlertSeverity;
  confidenceScore: number;
  explanation: string;
  features: FeatureContribution[];
  classification: ClassificationLevel;
  detectedAt: Date;
}

export type FeedbackStatus = "idle" | "submitting" | "accepted" | "under_review" | "rejected" | "queued";
