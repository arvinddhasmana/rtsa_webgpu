// CLASSIFICATION: UNCLASSIFIED
package producer_test

import (
	"context"
	"testing"

	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-training/internal/producer"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type mockKafka struct {
	lastRecord *kgo.Record
}

func (m *mockKafka) Produce(ctx context.Context, r *kgo.Record, promise func(*kgo.Record, error)) {
	m.lastRecord = r
	if promise != nil {
		promise(r, nil)
	}
}

func TestNew(t *testing.T) {
	p := producer.New(nil, "topic", zap.NewNop())
	if p == nil {
		t.Fatal("expected non-nil NoopProducer")
	}
}

func TestNoopProducer_PublishCandidate(t *testing.T) {
	mock := &mockKafka{}
	p := producer.New(mock, "training-out", zap.NewNop())

	err := p.PublishCandidate(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if mock.lastRecord == nil {
		t.Fatal("expected record to be produced")
	}
	if mock.lastRecord.Topic != "training-out" {
		t.Errorf("expected topic training-out, got %s", mock.lastRecord.Topic)
	}
}

func TestNoopProducer_PublishCandidate_ProduceError(t *testing.T) {
	mock := &mockKafkaErr{}
	p := producer.New(mock, "training-out", zap.NewNop())

	err := p.PublishCandidate(context.Background())
	if err == nil {
		t.Error("expected error from produce callback")
	}
}

type mockKafkaErr struct{}

func (m *mockKafkaErr) Produce(ctx context.Context, r *kgo.Record, promise func(*kgo.Record, error)) {
	if promise != nil {
		promise(r, context.DeadlineExceeded)
	}
}
