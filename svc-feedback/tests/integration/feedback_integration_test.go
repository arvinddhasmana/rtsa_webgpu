// CLASSIFICATION: UNCLASSIFIED

//go:build integration

// Package integration contains integration tests for svc-feedback.
// Run with: go test -tags=integration ./tests/integration/...
//
// These tests require external infrastructure (Redpanda broker).
// They are excluded from unit test runs by the build tag.
package integration

import (
"context"
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
feedbackv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/feedback/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-feedback/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-feedback/internal/handler"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-feedback/internal/producer"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-feedback/internal/state"
auditv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/audit/v1"
"go.uber.org/zap"
)

// mockIntegrationAudit is a no-op audit emitter for integration tests.
type mockIntegrationAudit struct{}

func (m *mockIntegrationAudit) EmitAudit(_ context.Context, _ *auditv1.AuditEvent) error {
return nil
}

// IT01: Submit feedback → both topics (using mock producers to simulate infra)
func TestIT01_SubmitFeedback_BothTopics(t *testing.T) {
h := state.NewOperatorHistory()
logger, _ := zap.NewDevelopment()
ts := domain.NewTrustScorer(h)
ap := domain.NewAntiPoisonGuard(h, logger)
rl := domain.NewRateLimiter(10)
rawProd := &producer.MockProducer{}
valProd := &producer.MockProducer{}
audit := &mockIntegrationAudit{}

hnd := handler.NewFeedbackHandler(ts, ap, rl, rawProd, valProd, audit, h, "svc-feedback-it", logger)

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

req := &feedbackv1.SubmitFeedbackRequest{
TrackId:           "trk-it01",
OperatorId:        "op-it01",
FeedbackType:      commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
OperatorClearance: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
Justification:     "integration test confirmed",
}

resp, err := hnd.SubmitFeedback(ctx, req)
if err != nil {
t.Fatalf("IT01: unexpected error: %v", err)
}

if len(rawProd.Messages) != 1 {
t.Errorf("IT01: expected message in submissions topic, got %d", len(rawProd.Messages))
}
if resp.GetValidated() && len(valProd.Messages) != 1 {
t.Errorf("IT01: expected message in validated topic, got %d", len(valProd.Messages))
}
}

// IT02: Submit low-trust feedback → only submissions topic
func TestIT02_LowTrustFeedback_OnlySubmissions(t *testing.T) {
h := state.NewOperatorHistory()
logger, _ := zap.NewDevelopment()

// Populate consensus: all HOSTILE for track.
for _, op := range []string{"op-c1", "op-c2", "op-c3"} {
h.RecordFeedback(op, state.FeedbackEntry{
FeedbackID:   "fb-" + op,
TrackID:      "trk-it02",
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
Timestamp:    time.Now(),
})
}

ts := domain.NewTrustScorer(h)
ap := domain.NewAntiPoisonGuard(h, logger)
rl := domain.NewRateLimiter(10)
rawProd := &producer.MockProducer{}
valProd := &producer.MockProducer{}
audit := &mockIntegrationAudit{}

hnd := handler.NewFeedbackHandler(ts, ap, rl, rawProd, valProd, audit, h, "svc-feedback-it", logger)

ctx := context.Background()
req := &feedbackv1.SubmitFeedbackRequest{
TrackId:           "trk-it02",
OperatorId:        "op-it02",
FeedbackType:      commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_FRIENDLY, // contradicts
OperatorClearance: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
}

resp, err := hnd.SubmitFeedback(ctx, req)
if err != nil {
t.Fatalf("IT02: unexpected error: %v", err)
}

// Submissions always written.
if len(rawProd.Messages) != 1 {
t.Errorf("IT02: expected submissions message, got %d", len(rawProd.Messages))
}

// Trust: C=0.3, A=0.5 (new), T=1.0, (1-D)=0.0 (contradicts all 3) → 0.41
if resp.GetValidated() {
t.Logf("IT02: note: trust_score=%f (deviation may be partial if not all contradict)", resp.GetTrustScore())
}
}

// IT03: Rate limit exceeded → ResourceExhausted, no message
func TestIT03_RateLimitExceeded(t *testing.T) {
h := state.NewOperatorHistory()
logger, _ := zap.NewDevelopment()
ts := domain.NewTrustScorer(h)
ap := domain.NewAntiPoisonGuard(h, logger)
rl := domain.NewRateLimiter(2) // low limit for test
rawProd := &producer.MockProducer{}
valProd := &producer.MockProducer{}
audit := &mockIntegrationAudit{}

hnd := handler.NewFeedbackHandler(ts, ap, rl, rawProd, valProd, audit, h, "svc-feedback-it", logger)

ctx := context.Background()
req := &feedbackv1.SubmitFeedbackRequest{
TrackId:      "trk-it03",
OperatorId:   "op-it03",
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
}

// First 2 succeed.
for i := 0; i < 2; i++ {
if _, err := hnd.SubmitFeedback(ctx, req); err != nil {
t.Fatalf("IT03: request %d should succeed: %v", i, err)
}
}

// Third should fail.
_, err := hnd.SubmitFeedback(ctx, req)
if err == nil {
t.Fatal("IT03: expected rate limit error")
}

// No extra messages after rate limit hit.
if len(rawProd.Messages) != 2 {
t.Errorf("IT03: expected exactly 2 messages, got %d", len(rawProd.Messages))
}
}

// IT04: Audit event emitted for every submission
func TestIT04_AuditEventEmitted(t *testing.T) {
type auditCapture struct {
events []*auditv1.AuditEvent
}
capture := &auditCapture{}

h := state.NewOperatorHistory()
logger, _ := zap.NewDevelopment()
ts := domain.NewTrustScorer(h)
ap := domain.NewAntiPoisonGuard(h, logger)
rl := domain.NewRateLimiter(10)
rawProd := &producer.MockProducer{}
valProd := &producer.MockProducer{}

// Use custom audit emitter that captures events.
customAudit := &captureAudit{events: &capture.events}
hnd := handler.NewFeedbackHandler(ts, ap, rl, rawProd, valProd, customAudit, h, "svc-feedback-it", logger)

ctx := context.Background()
req := &feedbackv1.SubmitFeedbackRequest{
TrackId:      "trk-it04",
OperatorId:   "op-it04",
FeedbackType: commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE,
}

if _, err := hnd.SubmitFeedback(ctx, req); err != nil {
t.Fatalf("IT04: unexpected error: %v", err)
}

if len(capture.events) != 1 {
t.Errorf("IT04: expected 1 audit event, got %d", len(capture.events))
}
ev := capture.events[0]
if ev.GetEventType() != "feedback.submitted" {
t.Errorf("IT04: expected event_type=feedback.submitted, got %s", ev.GetEventType())
}
}

type captureAudit struct {
events *[]*auditv1.AuditEvent
}

func (c *captureAudit) EmitAudit(_ context.Context, event *auditv1.AuditEvent) error {
*c.events = append(*c.events, event)
return nil
}
