// CLASSIFICATION: UNCLASSIFIED

package producer

import (
"context"
"errors"
"testing"

feedbackv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/feedback/v1"
)

func TestMockProducer_Produce_Success(t *testing.T) {
mock := &MockProducer{}
msg := &feedbackv1.OperatorFeedback{FeedbackId: "fb-001"}

if err := mock.Produce(context.Background(), "fb-001", msg); err != nil {
t.Fatalf("expected no error, got %v", err)
}
if len(mock.Messages) != 1 {
t.Fatalf("expected 1 message, got %d", len(mock.Messages))
}
if mock.Messages[0].Key != "fb-001" {
t.Errorf("expected key fb-001, got %s", mock.Messages[0].Key)
}
}

func TestMockProducer_Produce_Error(t *testing.T) {
sentinel := errors.New("broker unavailable")
mock := &MockProducer{Err: sentinel}
msg := &feedbackv1.OperatorFeedback{FeedbackId: "fb-002"}

err := mock.Produce(context.Background(), "fb-002", msg)
if !errors.Is(err, sentinel) {
t.Errorf("expected sentinel error, got %v", err)
}
if len(mock.Messages) != 0 {
t.Error("no messages should be recorded on error")
}
}

func TestMockProducer_Close(t *testing.T) {
mock := &MockProducer{}
if err := mock.Close(); err != nil {
t.Errorf("expected nil from Close, got %v", err)
}
}

func TestMockProducer_MultipleMessages(t *testing.T) {
mock := &MockProducer{}
msgs := []*feedbackv1.OperatorFeedback{
{FeedbackId: "fb-001"},
{FeedbackId: "fb-002"},
{FeedbackId: "fb-003"},
}
for _, m := range msgs {
if err := mock.Produce(context.Background(), m.FeedbackId, m); err != nil {
t.Fatalf("unexpected error: %v", err)
}
}
if len(mock.Messages) != 3 {
t.Errorf("expected 3 messages, got %d", len(mock.Messages))
}
}
