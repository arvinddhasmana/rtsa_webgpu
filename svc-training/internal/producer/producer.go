// CLASSIFICATION: UNCLASSIFIED
// Package producer provides the noop model candidate producer for svc-training.
//
// Feature: FEAT-12 Training Pipeline
// Requirements: CR-FB-003, CR-FB-004
package producer

import (
"context"
"encoding/json"
"fmt"
"time"

"github.com/twmb/franz-go/pkg/kgo"
"go.uber.org/zap"
)

// ModelCandidate is the JSON payload produced to the output topic.
type ModelCandidate struct {
ModelID   string    `json:"model_id"`
Status    string    `json:"status"`
Timestamp time.Time `json:"timestamp"`
}

// NoopProducer publishes noop model candidates to the output topic.
type NoopProducer struct {
client   *kgo.Client
outTopic string
logger   *zap.Logger
}

// New creates a new NoopProducer.
func New(client *kgo.Client, outTopic string, logger *zap.Logger) *NoopProducer {
return &NoopProducer{
client:   client,
outTopic: outTopic,
logger:   logger,
}
}

// PublishCandidate produces a noop model candidate to the output topic.
func (p *NoopProducer) PublishCandidate(ctx context.Context) error {
candidate := ModelCandidate{
ModelID:   "noop-v0",
Status:    "stub",
Timestamp: time.Now().UTC(),
}
payload, err := json.Marshal(candidate)
if err != nil {
return fmt.Errorf("producer.PublishCandidate: marshal: %w", err)
}

var produceErr error
p.client.Produce(ctx, &kgo.Record{
Topic: p.outTopic,
Value: payload,
}, func(_ *kgo.Record, err error) {
if err != nil {
produceErr = fmt.Errorf("producer.PublishCandidate: produce: %w", err)
p.logger.Error("noop candidate produce failed", zap.Error(err))
}
})
return produceErr
}
