// CLASSIFICATION: UNCLASSIFIED

// Package handler implements the gRPC FeedbackService handlers.
package handler

import (
"context"
"fmt"
"time"

"github.com/google/uuid"
auditv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/audit/v1"
commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
feedbackv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/feedback/v1"
"github.com/arvinddhasmana/rtsa_webgpu/svc-feedback/internal/domain"
"github.com/arvinddhasmana/rtsa_webgpu/svc-feedback/internal/producer"
"github.com/arvinddhasmana/rtsa_webgpu/svc-feedback/internal/state"
"go.uber.org/zap"
"google.golang.org/grpc/codes"
"google.golang.org/grpc/status"
"google.golang.org/protobuf/types/known/timestamppb"
)

// AuditEmitter is the interface used by FeedbackHandler to publish audit events.
// Defined locally to avoid coupling to a concrete audit service client.
type AuditEmitter interface {
EmitAudit(ctx context.Context, event *auditv1.AuditEvent) error
}

// FeedbackHandler implements feedbackv1.FeedbackServiceServer.
type FeedbackHandler struct {
feedbackv1.UnimplementedFeedbackServiceServer

trustScorer   *domain.TrustScorer
antiPoison    *domain.AntiPoisonGuard
rateLimiter   *domain.RateLimiter
rawProducer   producer.MessageProducer // feedback.operator.submissions
validProducer producer.MessageProducer // feedback.operator.validated
auditEmitter  AuditEmitter
history       *state.OperatorHistory
serviceName   string
logger        *zap.Logger
}

// NewFeedbackHandler constructs a FeedbackHandler with all required dependencies.
func NewFeedbackHandler(
trustScorer *domain.TrustScorer,
antiPoison *domain.AntiPoisonGuard,
rateLimiter *domain.RateLimiter,
rawProducer producer.MessageProducer,
validProducer producer.MessageProducer,
auditEmitter AuditEmitter,
history *state.OperatorHistory,
serviceName string,
logger *zap.Logger,
) *FeedbackHandler {
return &FeedbackHandler{
trustScorer:   trustScorer,
antiPoison:    antiPoison,
rateLimiter:   rateLimiter,
rawProducer:   rawProducer,
validProducer: validProducer,
auditEmitter:  auditEmitter,
history:       history,
serviceName:   serviceName,
logger:        logger,
}
}

// SubmitFeedback processes an operator feedback submission.
//
// Flow:
//  1. Validate required fields (track_id, operator_id, feedback_type)
//  2. Enforce per-operator rate limit
//  3. Run anti-poisoning checks (failure logs warning; does NOT block)
//  4. Compute trust score (forced below 0.5 if anti-poison fails)
//  5. Build OperatorFeedback proto
//  6. Produce to feedback.operator.submissions (ALL feedback)
//  7. If trust_score ≥ 0.5 AND anti-poison passed → produce to feedback.operator.validated
//  8. Record in operator history
//  9. Emit audit event
//  10. Return SubmitFeedbackResponse
func (h *FeedbackHandler) SubmitFeedback(
ctx context.Context,
req *feedbackv1.SubmitFeedbackRequest,
) (*feedbackv1.SubmitFeedbackResponse, error) {
// 1. Validate required fields.
if req.GetTrackId() == "" {
return nil, status.Error(codes.InvalidArgument, "track_id is required")
}
if req.GetOperatorId() == "" {
return nil, status.Error(codes.InvalidArgument, "operator_id is required")
}
if req.GetFeedbackType() == commonv1.FeedbackType_FEEDBACK_TYPE_UNSPECIFIED {
return nil, status.Error(codes.InvalidArgument, "feedback_type is required")
}

operatorID := req.GetOperatorId()
trackID := req.GetTrackId()

// 2. Check rate limit.
if !h.rateLimiter.Allow(operatorID) {
h.logger.Warn("rate limit exceeded",
zap.String("operator_id", operatorID),
)
return nil, status.Errorf(codes.ResourceExhausted,
"rate limit exceeded for operator %s", operatorID)
}

// 3. Anti-poisoning check (non-blocking).
poisonResult := h.antiPoison.Check(operatorID)
if !poisonResult.Passed {
h.logger.Warn("anti-poisoning check failed — feedback will not be validated",
zap.String("operator_id", operatorID),
zap.Any("checks", poisonResult.Checks),
)
}

// 4. Compute trust score.
now := time.Now().UTC()
trustParams := domain.TrustParams{
OperatorID:        operatorID,
OperatorClearance: req.GetOperatorClearance(),
TrackID:           trackID,
FeedbackType:      req.GetFeedbackType(),
EventTime:         now, // Use now as event time; callers can extend this
FeedbackTime:      now,
}
trustResult := h.trustScorer.Score(trustParams)

// Force trust below validated threshold if anti-poisoning failed.
effectiveTrustScore := trustResult.TotalScore
validated := trustResult.Validated && poisonResult.Passed
if !poisonResult.Passed && effectiveTrustScore >= 0.5 {
effectiveTrustScore = 0.49
validated = false
}

// 5. Build OperatorFeedback proto.
feedbackID := uuid.New().String()
submittedAt := timestamppb.New(now)

breakdown := &feedbackv1.TrustBreakdown{
ClearanceScore: trustResult.ClearanceScore,
AccuracyScore:  trustResult.AccuracyScore,
TemporalScore:  trustResult.TemporalScore,
DeviationScore: trustResult.DeviationScore,
}

fb := &feedbackv1.OperatorFeedback{
FeedbackId:        feedbackID,
TrackId:           trackID,
OperatorId:        operatorID,
FeedbackType:      req.GetFeedbackType(),
Justification:     req.GetJustification(),
TrustScore:        effectiveTrustScore,
TrustBreakdown:    breakdown,
Classification:    req.GetOperatorClearance(),
SubmittedAt:       submittedAt,
OperatorClearance: req.GetOperatorClearance(),
}

// 6. Produce to submissions topic (ALL feedback).
if err := h.rawProducer.Produce(ctx, feedbackID, fb); err != nil {
h.logger.Error("failed to produce to submissions topic",
zap.String("feedback_id", feedbackID),
zap.Error(err),
)
return nil, status.Errorf(codes.Internal,
"[handler.SubmitFeedback(%s)]: submissions produce: %v", feedbackID, err)
}

// 7. Produce to validated topic (only if trusted AND anti-poison passed).
if validated {
if err := h.validProducer.Produce(ctx, feedbackID, fb); err != nil {
// Log but do not fail — the submission topic write already succeeded.
h.logger.Error("failed to produce to validated topic",
zap.String("feedback_id", feedbackID),
zap.Error(err),
)
}
}

// 8. Record in operator history.
entry := state.FeedbackEntry{
FeedbackID:   feedbackID,
TrackID:      trackID,
FeedbackType: req.GetFeedbackType(),
Timestamp:    now,
TrustScore:   effectiveTrustScore,
Validated:    validated,
}
h.history.RecordFeedback(operatorID, entry)
h.history.RecordProto(fb)

// 9. Emit audit event.
auditEvent := h.buildAuditEvent(feedbackID, operatorID, trackID, req.GetOperatorClearance(), now)
if err := h.auditEmitter.EmitAudit(ctx, auditEvent); err != nil {
// Audit failure is non-fatal; log and continue.
h.logger.Error("failed to emit audit event",
zap.String("feedback_id", feedbackID),
zap.Error(err),
)
}

h.logger.Info("feedback submitted",
zap.String("feedback_id", feedbackID),
zap.String("operator_id", operatorID),
zap.Float64("trust_score", effectiveTrustScore),
zap.Bool("validated", validated),
)

// 10. Return response.
return &feedbackv1.SubmitFeedbackResponse{
FeedbackId:     feedbackID,
TrustScore:     effectiveTrustScore,
TrustBreakdown: breakdown,
Validated:      validated,
SubmittedAt:    submittedAt,
}, nil
}

// GetFeedbackHistory queries feedback history from in-memory state.
// Supports optional filtering by operator_id and track_id.
func (h *FeedbackHandler) GetFeedbackHistory(
ctx context.Context,
req *feedbackv1.GetFeedbackHistoryRequest,
) (*feedbackv1.GetFeedbackHistoryResponse, error) {
operatorID := req.GetOperatorId()
trackID := req.GetTrackId()

results := h.history.QueryHistory(operatorID, trackID)

h.logger.Info("feedback history queried",
zap.String("operator_id", operatorID),
zap.String("track_id", trackID),
zap.Int("result_count", len(results)),
)

return &feedbackv1.GetFeedbackHistoryResponse{
Feedback: results,
Pagination: &commonv1.PaginationResponse{
TotalCount: int32(len(results)),
},
}, nil
}

// buildAuditEvent constructs the audit event for a feedback submission.
// No PII or classified data is embedded in the detail log.
func (h *FeedbackHandler) buildAuditEvent(
feedbackID, operatorID, trackID string,
clearance commonv1.ClassificationLevel,
eventTime time.Time,
) *auditv1.AuditEvent {
return &auditv1.AuditEvent{
AuditId:             fmt.Sprintf("aud-%s", feedbackID),
ServiceId:           h.serviceName,
EventType:           "feedback.submitted",
ActorId:             operatorID,
ActorType:           auditv1.ActorType_ACTOR_TYPE_OPERATOR,
ResourceType:        "track",
ResourceId:          trackID,
Action:              auditv1.AuditAction_AUDIT_ACTION_FEEDBACK,
DetailJson:          fmt.Sprintf(`{"feedback_id":%q}`, feedbackID),
ClassificationLevel: clearance,
EventTime:           timestamppb.New(eventTime),
}
}
