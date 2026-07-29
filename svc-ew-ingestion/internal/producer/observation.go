// CLASSIFICATION: UNCLASSIFIED
package producer

import (
	"context"
	"fmt"

	ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/pkg/classification"
	"github.com/arvinddhasmana/rtsa_webgpu/pkg/redpanda"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/encoding/protojson"
)

// ObservationProducer produces SensorObservation messages to Redpanda.
type ObservationProducer struct {
producer *redpanda.Producer
topic    string
}

// NewObservationProducer creates a producer for a specific topic.
func NewObservationProducer(producer *redpanda.Producer, topic string) *ObservationProducer {
return &ObservationProducer{
producer: producer,
topic:    topic,
}
}

// Produce serializes and sends a SensorObservation.
func (p *ObservationProducer) Produce(ctx context.Context, obs *ingestionv1.SensorObservation) error {
	b, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(obs)
if err != nil {
return fmt.Errorf("producer: marshal observation: %w", err)
}

classificationStr := classification.LevelToString(obs.GetClassification())

traceID := ""
if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
traceID = span.SpanContext().TraceID().String()
}

if err := p.producer.Produce(ctx, p.topic,
[]byte(obs.GetSensorId()), b, classificationStr, traceID); err != nil {
return fmt.Errorf("producer: produce observation to topic %s: %w", p.topic, err)
}
return nil
}

// Topic returns the topic this producer writes to.
func (p *ObservationProducer) Topic() string {
return p.topic
}
