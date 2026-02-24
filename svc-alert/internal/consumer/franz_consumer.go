// CLASSIFICATION: UNCLASSIFIED

//go:build integration

package consumer

import (
"context"
"fmt"
"log/slog"

"github.com/twmb/franz-go/pkg/kgo"
)

// FranzConsumerClient implements ConsumerClient using the franz-go Kafka library.
// It is used in production to consume from Redpanda.
type FranzConsumerClient struct {
client *kgo.Client
logger *slog.Logger
}

// NewFranzConsumerClient creates a new FranzConsumerClient connected to the given brokers.
func NewFranzConsumerClient(brokers []string, group string, logger *slog.Logger) (*FranzConsumerClient, error) {
opts := []kgo.Opt{
kgo.SeedBrokers(brokers...),
kgo.ConsumerGroup(group),
kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
}

client, err := kgo.NewClient(opts...)
if err != nil {
return nil, fmt.Errorf("[consumer].[NewFranzConsumerClient]: %w", err)
}

return &FranzConsumerClient{
client: client,
logger: logger,
}, nil
}

// Consume polls Redpanda for messages and invokes handler for each record.
// Blocks until ctx is cancelled or the client is closed.
func (f *FranzConsumerClient) Consume(ctx context.Context, topics []string, handler MessageHandler) error {
f.client.AddConsumeTopics(topics...)

for {
fetches := f.client.PollFetches(ctx)

if ctx.Err() != nil {
return nil // context cancelled — normal shutdown
}
if fetches.IsClientClosed() {
return fmt.Errorf("[consumer].[FranzConsumerClient.Consume]: client closed unexpectedly")
}

fetches.EachError(func(t string, p int32, err error) {
f.logger.ErrorContext(ctx, "fetch error",
"topic", t,
"partition", p,
"error", err.Error(),
)
})

fetches.EachRecord(func(rec *kgo.Record) {
if err := handler(ctx, rec.Topic, rec.Key, rec.Value); err != nil {
f.logger.ErrorContext(ctx, "message handler error",
"topic", rec.Topic,
"offset", rec.Offset,
"error", err.Error(),
)
}
})
}
}

// Close shuts down the franz-go client.
func (f *FranzConsumerClient) Close() error {
f.client.Close()
return nil
}
