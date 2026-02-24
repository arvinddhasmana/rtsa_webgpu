// CLASSIFICATION: UNCLASSIFIED
package redpanda

import (
"context"
"fmt"
"log/slog"

"github.com/twmb/franz-go/pkg/kgo"
)

// Consumer is a Redpanda message consumer backed by franz-go.
type Consumer struct {
client *kgo.Client
logger *slog.Logger
}

// NewConsumer creates a new Consumer connected to the given brokers.
func NewConsumer(brokers []string, groupID string, opts ...kgo.Opt) (*Consumer, error) {
defaultOpts := []kgo.Opt{
kgo.SeedBrokers(brokers...),
kgo.ConsumerGroup(groupID),
kgo.DisableAutoCommit(),
}
defaultOpts = append(defaultOpts, opts...)

client, err := kgo.NewClient(defaultOpts...)
if err != nil {
return nil, fmt.Errorf("[redpanda.NewConsumer]: %w", err)
}
return &Consumer{
client: client,
logger: slog.Default(),
}, nil
}

// Consume starts consuming from topics, calling handler for each message.
func (c *Consumer) Consume(ctx context.Context, topics []string, handler MessageHandler) error {
c.client.AddConsumeTopics(topics...)
for {
fetches := c.client.PollFetches(ctx)
if err := ctx.Err(); err != nil {
return nil
}
if errs := fetches.Errors(); len(errs) > 0 {
for _, fetchErr := range errs {
c.logger.Error("fetch error", "topic", fetchErr.Topic, "partition", fetchErr.Partition, "error", fetchErr.Err)
}
}
fetches.EachRecord(func(rec *kgo.Record) {
msg := &Message{
Topic:     rec.Topic,
Key:       rec.Key,
Value:     rec.Value,
Offset:    rec.Offset,
Partition: rec.Partition,
Timestamp: rec.Timestamp,
Headers:   make(map[string]string, len(rec.Headers)),
}
for _, h := range rec.Headers {
msg.Headers[h.Key] = string(h.Value)
}
if err := handler(ctx, msg); err != nil {
c.logger.Error("handler error", "topic", rec.Topic, "offset", rec.Offset, "error", err)
}
})
if err := c.client.CommitUncommittedOffsets(ctx); err != nil {
c.logger.Error("commit error", "error", err)
}
}
}

// Close stops the consumer.
func (c *Consumer) Close() error {
c.client.Close()
return nil
}
