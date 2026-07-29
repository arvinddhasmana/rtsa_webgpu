// CLASSIFICATION: UNCLASSIFIED
package producer_test

import (
	"context"
	"testing"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-cyber-ingestion/internal/producer"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel/trace"
)

type mockProducer struct {
	lastTopic          string
	lastClassification string
	lastTraceID        string
	produceErr         error
}

func (m *mockProducer) Produce(ctx context.Context, topic string, key []byte, value []byte, classification string, traceID string, extraHeaders ...kgo.RecordHeader) error {
	m.lastTopic = topic
	m.lastClassification = classification
	m.lastTraceID = traceID
	return m.produceErr
}

func (m *mockProducer) Close() error                     { return nil }
func (m *mockProducer) Healthy(ctx context.Context) bool { return true }

func TestObservationProducer_Topic(t *testing.T) {
	p := producer.NewObservationProducer(nil, "test-topic")
	if p.Topic() != "test-topic" {
		t.Errorf("expected test-topic, got %s", p.Topic())
	}
}

func TestObservationProducer_NilProducer(t *testing.T) {
	p := producer.NewObservationProducer(nil, "test-topic")
	obs := &ingestionv1.SensorObservation{
		SensorId:       "S-1",
		Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
	}
	err := p.Produce(context.Background(), obs)
	if err == nil {
		t.Error("expected error when internal producer is nil")
	}
}

func TestObservationProducer_Produce_Success(t *testing.T) {
	mock := &mockProducer{}
	p := producer.NewObservationProducer(mock, "cyber-topic")
	obs := &ingestionv1.SensorObservation{
		SensorId:       "S-1",
		Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
	}

	err := p.Produce(context.Background(), obs)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if mock.lastTopic != "cyber-topic" {
		t.Errorf("expected topic cyber-topic, got %s", mock.lastTopic)
	}
	if mock.lastClassification != "SECRET" {
		t.Errorf("expected classification SECRET, got %s", mock.lastClassification)
	}
}

func TestObservationProducer_Produce_WithTracing(t *testing.T) {
	mock := &mockProducer{}
	p := producer.NewObservationProducer(mock, "cyber-topic")
	obs := &ingestionv1.SensorObservation{
		SensorId: "S-1",
	}

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10},
		SpanID:  trace.SpanID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	_ = p.Produce(ctx, obs)
	if mock.lastTraceID == "" {
		t.Error("expected non-empty trace id")
	}
	expectedTraceID := "0102030405060708090a0b0c0d0e0f10"
	if mock.lastTraceID != expectedTraceID {
		t.Errorf("expected trace id %s, got %s", expectedTraceID, mock.lastTraceID)
	}
}

func TestObservationProducer_Produce_Panic(t *testing.T) {
	p := producer.NewObservationProducer(nil, "topic")
	obs := &ingestionv1.SensorObservation{}
	defer func() {
		_ = recover()
	}()
	_ = p.Produce(context.Background(), obs)
}
