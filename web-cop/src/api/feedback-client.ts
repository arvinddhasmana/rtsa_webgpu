// CLASSIFICATION: UNCLASSIFIED
// src/api/feedback-client.ts

import { FeedbackRequest, FeedbackResponse } from "../types/feedback";

/**
 * FeedbackClient wraps the FeedbackService gRPC-Web unary endpoint.
 * Production: connects to svc-feedback via Envoy at /rtsa.feedback.v1.FeedbackService/SubmitFeedback
 */
export class FeedbackClient {
  async submitFeedback(req: FeedbackRequest): Promise<FeedbackResponse> {
    // Production: call via @protobuf-ts generated client
    // Stub for development/testing
    void req;
    return {
      feedbackId: `fb-${Date.now()}`,
      trustScore: 0.85,
      accepted: true,
      message: "Feedback accepted",
    };
  }
}

export const feedbackClient = new FeedbackClient();
