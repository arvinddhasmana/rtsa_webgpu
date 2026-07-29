// CLASSIFICATION: UNCLASSIFIED
package consumer_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/arvinddhasmana/rtsa_webgpu/svc-training/internal/consumer"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type mockKafka struct {
	fetches    kgo.Fetches
	lastRecord *kgo.Record
	cancelFunc context.CancelFunc
}

func (m *mockKafka) PollFetches(ctx context.Context) kgo.Fetches {
	if !m.fetches.Empty() {
		f := m.fetches
		m.fetches = kgo.Fetches{}
		return f
	}
	select {
	case <-ctx.Done():
		return kgo.NewErrFetch(ctx.Err())
	case <-time.After(10 * time.Millisecond):
		return kgo.Fetches{}
	}
}

func (m *mockKafka) Produce(ctx context.Context, r *kgo.Record, promise func(*kgo.Record, error)) {
	m.lastRecord = r
	if promise != nil {
		promise(r, nil)
	}
	if m.cancelFunc != nil {
		m.cancelFunc()
	}
}

func TestNew(t *testing.T) {
	tc := consumer.New(nil, nil, "out", zap.NewNop())
	if tc == nil {
		t.Fatal("expected non-nil TrainingConsumer")
	}
}

func TestTrainingConsumer_Run_Cancelled(t *testing.T) {
	mock := &mockKafka{}
	tc := consumer.New(mock, mock, "out", zap.NewNop())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := tc.Run(ctx)
	if err != nil {
		t.Errorf("expected no error on cancel, got %v", err)
	}
}

func TestTrainingConsumer_Run_ProcessRecord(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	mock := &mockKafka{
		fetches: kgo.Fetches{
			kgo.Fetch{
				Topics: []kgo.FetchTopic{
					{
						Topic: "in",
						Partitions: []kgo.FetchPartition{
							{
								Records: []*kgo.Record{
									{Value: []byte("feedback")},
								},
							},
						},
					},
				},
			},
		},
		cancelFunc: cancel, // cancel the context once the record is produced
	}
	tc := consumer.New(mock, mock, "out", zap.NewNop())

	_ = tc.Run(ctx)

	if mock.lastRecord == nil {
		t.Error("expected record to be produced")
	}
}

func TestNoopModelCandidateJSONMarshal(t *testing.T) {
	candidate := consumer.NoopModelCandidate{
		ModelID: "noop",
	}
	data, err := json.Marshal(candidate)
	if err != nil || len(data) == 0 {
		t.Fatal("marshal failed")
	}
}
