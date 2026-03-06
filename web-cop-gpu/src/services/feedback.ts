// CLASSIFICATION: UNCLASSIFIED
// src/services/feedback.ts — FeedbackService gRPC calls
//
// Submits operator feedback on tracks via the gRPC cold path.
// Reference: docs/implementation/v4/phase3_ui_interaction.md §5
// Reference: docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md §6

import { createPromiseClient } from "@connectrpc/connect";
import { FeedbackService } from "@gen/rtsa/feedback/v1/feedback_service_connect.js";
import { transport } from "./grpc-client";
import {
  ClassificationLevel,
  FeedbackType,
} from "@gen/rtsa/common/v1/types_pb.js";

const client = createPromiseClient(FeedbackService, transport);

/** Feedback type options available in the UI. */
export type FeedbackTypeOption =
  | "CONFIRM_HOSTILE"
  | "CONFIRM_FRIENDLY"
  | "RECLASSIFY"
  | "REJECT_ANOMALY";

const feedbackTypeMap: Record<FeedbackTypeOption, FeedbackType> = {
  CONFIRM_HOSTILE: FeedbackType.CONFIRM_HOSTILE,
  CONFIRM_FRIENDLY: FeedbackType.CONFIRM_FRIENDLY,
  RECLASSIFY: FeedbackType.RECLASSIFY,
  REJECT_ANOMALY: FeedbackType.REJECT_ANOMALY,
};

export interface SubmitFeedbackParams {
  trackId: string;
  operatorId: string;
  feedbackType: FeedbackTypeOption;
  justification: string;
  alertId?: string;
}

export interface SubmitFeedbackResult {
  feedbackId: string;
  trustScore: number;
  validated: boolean;
}

/**
 * Submit operator feedback on a track via gRPC FeedbackService.
 * All inputs are validated before sending.
 */
export async function submitFeedback(
  params: SubmitFeedbackParams,
): Promise<SubmitFeedbackResult> {
  // Input validation
  const trackId = params.trackId.trim();
  const operatorId = params.operatorId.trim();
  const justification = params.justification.trim();

  if (!trackId) throw new Error("trackId is required");
  if (!operatorId) throw new Error("operatorId is required");
  if (!justification) throw new Error("justification is required");
  if (justification.length > 500) throw new Error("justification too long (max 500 chars)");

  const response = await client.submitFeedback({
    trackId,
    operatorId,
    feedbackType: feedbackTypeMap[params.feedbackType],
    justification,
    alertId: params.alertId,
    operatorClearance: ClassificationLevel.UNCLASSIFIED,
  });

  return {
    feedbackId: response.feedbackId,
    trustScore: response.trustScore,
    validated: response.validated,
  };
}
