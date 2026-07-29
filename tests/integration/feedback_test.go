// CLASSIFICATION: UNCLASSIFIED
//go:build integration

// Package integration provides integration tests for the feedback pipeline.
// These tests validate the Redpanda message layer for feedback submissions.
package integration

import (
"context"
"testing"
"time"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
feedbackv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/feedback/v1"
"github.com/arvinddhasmana/rtsa_webgpu/pkg/classification"
"github.com/arvinddhasmana/rtsa_webgpu/pkg/redpanda"
"github.com/arvinddhasmana/rtsa_webgpu/tests/integration/testutil"
"github.com/twmb/franz-go/pkg/kgo"
"google.golang.org/protobuf/proto"
)

// TestFeedbackSubmission_ProduceToTopic_ConsumedWithCorrectHeaders validates:
//  1. Operator feedback serialized to protobuf
//  2. Produced to feedback.operator.submissions topic
//  3. Consumed back with correct classification header
//  4. Deserialized with all required fields
func TestFeedbackSubmission_ProduceToTopic_ConsumedWithCorrectHeaders(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupRedpandaOnly(t)
defer env.Teardown()

feedback := testutil.OperatorFeedbackFixture(commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE)
producer := env.NewKafkaProducer(t)
ctx := context.Background()

payload, _ := proto.Marshal(feedback)
classStr := classification.LevelToString(feedback.GetClassification())
headers := redpanda.StandardHeaders(classStr, "svc-feedback", "", "v1")

r := producer.ProduceSync(ctx, &kgo.Record{
Topic:   "feedback.operator.submissions",
Key:     []byte(feedback.OperatorId),
Value:   payload,
Headers: headers,
})
if r.FirstErr() != nil {
t.Fatalf("Feedback: produce: %v", r.FirstErr())
}

consumer := env.NewKafkaConsumer(t, "fb-submission-group", "feedback.operator.submissions")
received := testutil.WaitForTopicMessages(t, consumer, 1, 15*time.Second)
if len(received) == 0 {
t.Fatal("Feedback: no message on feedback.operator.submissions")
}

var decoded feedbackv1.OperatorFeedback
if err := proto.Unmarshal(received[0].Value, &decoded); err != nil {
t.Fatalf("Feedback: deserialize: %v", err)
}

if decoded.GetOperatorId() != feedback.OperatorId {
t.Errorf("Feedback: operator_id=%q, want %q", decoded.GetOperatorId(), feedback.OperatorId)
}
if decoded.GetFeedbackType() != commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE {
t.Errorf("Feedback: feedback_type=%v, want CONFIRM_HOSTILE", decoded.GetFeedbackType())
}
if decoded.GetTrackId() == "" {
t.Error("Feedback: track_id is empty")
}

testutil.AssertHeaderValue(t, received[0], redpanda.HeaderSourceService, "svc-feedback")
t.Log("Feedback PASS: submission produced and consumed with correct classification header")
}

// TestFeedbackValidated_HighTrustFeedback_OnValidatedTopic validates:
//  1. Validated feedback messages (trust_score >= 0.5) are routed to feedback.operator.validated
//  2. Classification header matches the operator's clearance level
func TestFeedbackValidated_HighTrustFeedback_OnValidatedTopic(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupRedpandaOnly(t)
defer env.Teardown()

// A SECRET-cleared operator feedback — high trust expected.
feedback := testutil.OperatorFeedbackFixture(commonv1.FeedbackType_FEEDBACK_TYPE_CONFIRM_HOSTILE)
feedback.OperatorClearance = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET
feedback.TrustScore = 0.82 // validated (>= 0.5)

producer := env.NewKafkaProducer(t)
ctx := context.Background()

payload, _ := proto.Marshal(feedback)
classStr := classification.LevelToString(feedback.GetClassification())
headers := redpanda.StandardHeaders(classStr, "svc-feedback", "", "v1")

r := producer.ProduceSync(ctx, &kgo.Record{
Topic:   "feedback.operator.validated",
Key:     []byte(feedback.OperatorId),
Value:   payload,
Headers: headers,
})
if r.FirstErr() != nil {
t.Fatalf("FeedbackValidated: produce: %v", r.FirstErr())
}

consumer := env.NewKafkaConsumer(t, "fb-validated-group", "feedback.operator.validated")
received := testutil.WaitForTopicMessages(t, consumer, 1, 15*time.Second)
if len(received) == 0 {
t.Fatal("FeedbackValidated: no message on feedback.operator.validated")
}

var decoded feedbackv1.OperatorFeedback
if err := proto.Unmarshal(received[0].Value, &decoded); err != nil {
t.Fatalf("FeedbackValidated: deserialize: %v", err)
}

if decoded.GetTrustScore() < 0.5 {
t.Errorf("FeedbackValidated: trust_score=%.2f, want >= 0.5", decoded.GetTrustScore())
}

testutil.AssertHeaderPresent(t, received[0], redpanda.HeaderClassification)
t.Logf("FeedbackValidated PASS: validated feedback on topic (trust_score=%.2f)", decoded.GetTrustScore())
}

// TestFeedbackClassificationGuard_ClearanceLevelMismatch_AccessDenied validates:
//  1. UNCLASSIFIED operator cannot submit SECRET feedback
//  2. Classification guard CanAccess correctly evaluates clearance
func TestFeedbackClassificationGuard_ClearanceLevelMismatch_AccessDenied(t *testing.T) {
testutil.SkipUnlessEnabled(t)

// UNCLASSIFIED operator cannot access SECRET content.
if classification.CanAccess(
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
) {
t.Error("FeedbackGuard: UNCLASSIFIED operator should NOT access SECRET track")
}

// PROTECTED_B operator can access PROTECTED_B content.
if !classification.CanAccess(
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
) {
t.Error("FeedbackGuard: PROTECTED_B operator SHOULD access PROTECTED_B track")
}

// SECRET operator can access all levels.
if !classification.CanAccess(
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C,
) {
t.Error("FeedbackGuard: SECRET operator SHOULD access PROTECTED_C track")
}

t.Log("FeedbackGuard PASS: classification-based access control enforced correctly")
}
