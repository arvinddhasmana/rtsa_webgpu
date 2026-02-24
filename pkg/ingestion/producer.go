// CLASSIFICATION: UNCLASSIFIED
package ingestion

import (
"context"
"fmt"

ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"google.golang.org/protobuf/proto"
)

// LogProducer is a simple producer that logs messages (dev/test use).
// In production, replace with a real Redpanda producer.
type LogProducer struct {
topic string
}

// NewLogProducer creates a new LogProducer for the given topic.
func NewLogProducer(topic string) *LogProducer {
return &LogProducer{topic: topic}
}

// Produce marshals and "produces" the observation (dev: discards after marshal check).
func (p *LogProducer) Produce(_ context.Context, obs *ingestionv1.SensorObservation) error {
b, err := proto.Marshal(obs)
if err != nil {
return fmt.Errorf("producer: marshal failed: %w", err)
}
_ = b // In production: send to Redpanda
return nil
}

// Close is a no-op for LogProducer.
func (p *LogProducer) Close() error { return nil }

// NoopProducer discards all messages (for DLQ in tests).
type NoopProducer struct{}

// Produce discards the observation.
func (p *NoopProducer) Produce(_ context.Context, _ *ingestionv1.SensorObservation) error {
return nil
}

// Close is a no-op.
func (p *NoopProducer) Close() error { return nil }
