// CLASSIFICATION: UNCLASSIFIED
// src/services/feedback.ts — FeedbackService gRPC calls
//
// Submits operator feedback on tracks via the gRPC cold path.
// Reference: docs/implementation/v4/phase3_ui_interaction.md §5
// Reference: docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md §6

import { createPromiseClient } from "@connectrpc/connect";
import {
  ClassificationLevel,
  FeedbackType,
} from "@gen/rtsa/common/v1/types_pb.js";
import { FeedbackService } from "@gen/rtsa/feedback/v1/feedback_service_connect.js";
import { transport } from "./grpc-client";

const client = createPromiseClient(FeedbackService, transport);

type E2EFeedbackMocks = {
  submitFeedback?: (request: {
    trackId: string;
    operatorId: string;
    feedbackType: FeedbackType;
    justification: string;
    alertId?: string;
    operatorClearance: ClassificationLevel;
  }) => Promise<{
    feedbackId: string;
    trustScore: number;
    validated: boolean;
  }>;
};

function getE2EFeedbackMocks(): E2EFeedbackMocks | undefined {
  const maybeGlobal = globalThis as typeof globalThis & {
    __RTSA_E2E_MOCKS__?: E2EFeedbackMocks;
  };
  return maybeGlobal.__RTSA_E2E_MOCKS__;
}

/** Feedback type options available in the UI. */
export type FeedbackTypeOption =
  | "CONFIRM_HOSTILE"
  | "CONFIRM_FRIENDLY"
  | "RECLASSIFY"
  | "REJECT_ANOMALY"
  | "CONFIRM_ANOMALY";

const feedbackTypeMap: Record<FeedbackTypeOption, FeedbackType> = {
  CONFIRM_HOSTILE: FeedbackType.CONFIRM_HOSTILE,
  CONFIRM_FRIENDLY: FeedbackType.CONFIRM_FRIENDLY,
  RECLASSIFY: FeedbackType.RECLASSIFY,
  REJECT_ANOMALY: FeedbackType.REJECT_ANOMALY,
  CONFIRM_ANOMALY: FeedbackType.CONFIRM_ANOMALY,
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

export interface AlertFeedbackParams {
  alertId: string;
  trackId: string;
  operatorId: string;
  justification?: string;
}

function normalizeAlertFeedbackParams(params: AlertFeedbackParams): {
  alertId: string;
  trackId: string;
  operatorId: string;
  justification: string;
} {
  const alertId = params.alertId.trim();
  const trackId = params.trackId.trim();
  const operatorId = params.operatorId.trim();
  const justification = (params.justification ?? "").trim();

  if (!alertId) throw new Error("alertId is required");
  if (!trackId) throw new Error("trackId is required");
  if (!operatorId) throw new Error("operatorId is required");

  return {
    alertId,
    trackId,
    operatorId,
    justification,
  };
}

/**
 * Build a strongly typed SubmitFeedback payload for "Confirm" alert action.
 */
export function buildConfirmAlertFeedbackRequest(
  params: AlertFeedbackParams,
): SubmitFeedbackParams {
  const normalized = normalizeAlertFeedbackParams(params);
  return {
    trackId: normalized.trackId,
    operatorId: normalized.operatorId,
    feedbackType: "CONFIRM_ANOMALY",
    justification:
      normalized.justification || "Alert confirmed by operator quick-action.",
    alertId: normalized.alertId,
  };
}

/**
 * Build a strongly typed SubmitFeedback payload for "Reject" alert action.
 */
export function buildRejectAlertFeedbackRequest(
  params: AlertFeedbackParams,
): SubmitFeedbackParams {
  const normalized = normalizeAlertFeedbackParams(params);
  return {
    trackId: normalized.trackId,
    operatorId: normalized.operatorId,
    feedbackType: "REJECT_ANOMALY",
    justification:
      normalized.justification || "Alert rejected by operator quick-action.",
    alertId: normalized.alertId,
  };
}

/**
 * Submit a confirm-anomaly feedback entry for an alert quick-action.
 */
export async function submitConfirmAlertFeedback(
  params: AlertFeedbackParams,
): Promise<SubmitFeedbackResult> {
  return submitFeedback(buildConfirmAlertFeedbackRequest(params));
}

/**
 * Submit a reject-anomaly feedback entry for an alert quick-action.
 */
export async function submitRejectAlertFeedback(
  params: AlertFeedbackParams,
): Promise<SubmitFeedbackResult> {
  return submitFeedback(buildRejectAlertFeedbackRequest(params));
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
  if (justification.length > 500)
    throw new Error("justification too long (max 500 chars)");

  const request = {
    trackId,
    operatorId,
    feedbackType: feedbackTypeMap[params.feedbackType],
    justification,
    alertId: params.alertId,
    operatorClearance: ClassificationLevel.UNCLASSIFIED,
  };

  const response = getE2EFeedbackMocks()?.submitFeedback
    ? await getE2EFeedbackMocks()!.submitFeedback!(request)
    : await client.submitFeedback(request);

  return {
    feedbackId: response.feedbackId,
    trustScore: response.trustScore,
    validated: response.validated,
  };
}
