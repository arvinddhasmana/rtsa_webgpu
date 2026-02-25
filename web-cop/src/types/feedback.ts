// CLASSIFICATION: UNCLASSIFIED
// src/types/feedback.ts

import { FeedbackType } from "./common";

export interface FeedbackRequest {
  trackId: string;
  alertId?: string;
  feedbackType: FeedbackType;
  justification: string;
  operatorId: string;
}

export interface FeedbackResponse {
  feedbackId: string;
  trustScore: number;
  accepted: boolean;
  message: string;
}
