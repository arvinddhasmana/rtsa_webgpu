// CLASSIFICATION: UNCLASSIFIED
// Package consumer provides the noop training pipeline consumer.
//
// Feature: FEAT-12 Training Pipeline
// UC: UC014, UC015
// Requirements: CR-FB-003, CR-FB-004
package consumer

import (
"context"
"encoding/json"
"fmt"
"time"

"github.com/twmb/franz-go/pkg/kgo"
"go.uber.org/zap"
)

// NoopModelCandidate is the JSON payload produced to models.anomaly.candidates.
type NoopModelCandidate struct {
ModelID   string    `json:"model_id"`
Status    string    `json:"status"`
Timestamp time.Time `json:"timestamp"`
}

// TrainingConsumer reads validated feedback and produces noop model candidates.
type TrainingConsumer struct {
consumer *kgo.Client
producer *kgo.Client
outTopic string
logger   *zap.Logger
}

// New creates a new TrainingConsumer.
func New(consumer, producer *kgo.Client, outTopic string, logger *zap.Logger) *TrainingConsumer {
return &TrainingConsumer{
consumer: consumer,
producer: producer,
outTopic: outTopic,
logger:   logger,
}
}

// Run starts the consume→produce loop. It blocks until ctx is cancelled.
func (tc *TrainingConsumer) Run(ctx context.Context) error {
for {
fetches := tc.consumer.PollFetches(ctx)
if errs := fetches.Errors(); len(errs) > 0 {
for _, e := range errs {
if e.Err == context.Canceled {
return nil
}
tc.logger.Error("fetch error", zap.Error(e.Err))
}
continue
}

iter := fetches.RecordIter()
for !iter.Done() {
record := iter.Next()
tc.logger.Info("noop training — received feedback",
zap.String("topic", record.Topic),
zap.Int64("offset", record.Offset),
zap.Int("payload_bytes", len(record.Value)),
)

candidate := NoopModelCandidate{
ModelID:   "noop-v0",
Status:    "stub",
Timestamp: time.Now().UTC(),
}
payload, err := json.Marshal(candidate)
if err != nil {
return fmt.Errorf("marshal noop candidate: %w", err)
}

tc.producer.Produce(ctx, &kgo.Record{
Topic: tc.outTopic,
Value: payload,
}, func(_ *kgo.Record, err error) {
if err != nil {
tc.logger.Error("produce noop candidate failed", zap.Error(err))
}
})
}
}
}
