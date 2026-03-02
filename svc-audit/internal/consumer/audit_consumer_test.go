// CLASSIFICATION: UNCLASSIFIED
package consumer_test

import (
	"context"
	"testing"
	"time"

	auditv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/audit/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-audit/internal/consumer"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockKafka struct {
	fetches kgo.Fetches
	closed  bool
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
func (m *mockKafka) Close() { m.closed = true }

type mockRepo struct {
	lastBatch []*auditv1.AuditEvent
}

func (m *mockRepo) BatchInsert(ctx context.Context, events []*auditv1.AuditEvent) error {
	m.lastBatch = events
	return nil
}

func TestAuditConsumer_ProcessAndFlush(t *testing.T) {
	event := &auditv1.AuditEvent{
		AuditId:            "test-audit-id",
		ServiceId:          "test-svc",
		EventType:          "test-event",
		EventTime:          timestamppb.Now(),
		ClassificationLevel: 1,
	}
	data, _ := protojson.Marshal(event)

	mockK := &mockKafka{
		fetches: kgo.Fetches{
			kgo.Fetch{
				Topics: []kgo.FetchTopic{
					{
						Topic: "audit.events",
						Partitions: []kgo.FetchPartition{
							{
								Records: []*kgo.Record{
									{Value: data},
								},
							},
						},
					},
				},
			},
		},
	}
	mockR := &mockRepo{}

	// Create consumer with small batch size
	c := consumer.NewAuditConsumer(mockK, mockR, 1, 1, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_ = c.Start(ctx)

	if len(mockR.lastBatch) == 0 {
		t.Error("expected batch to be flushed")
	} else if mockR.lastBatch[0].AuditId != "test-audit-id" {
		t.Errorf("expected audit-id test-audit-id, got %s", mockR.lastBatch[0].AuditId)
	}
}

func TestAuditConsumer_MalformedRecord(t *testing.T) {
	mockK := &mockKafka{
		fetches: kgo.Fetches{
			kgo.Fetch{
				Topics: []kgo.FetchTopic{
					{
						Topic: "audit.events",
						Partitions: []kgo.FetchPartition{
							{
								Records: []*kgo.Record{
									{Value: []byte("malformed")},
								},
							},
						},
					},
				},
			},
		},
	}
	mockR := &mockRepo{}

	c := consumer.NewAuditConsumer(mockK, mockR, 1, 1, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_ = c.Start(ctx)

	if len(mockR.lastBatch) != 0 {
		t.Error("expected no batch to be flushed for malformed record")
	}
}

func TestAuditConsumer_Close(t *testing.T) {
	mockK := &mockKafka{}
	c := consumer.NewAuditConsumer(mockK, nil, 1, 1, zap.NewNop())
	c.Close()
	if !mockK.closed {
		t.Error("expected kafka client to be closed")
	}
}
