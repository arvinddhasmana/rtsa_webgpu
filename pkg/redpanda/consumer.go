// CLASSIFICATION: UNCLASSIFIED
package redpanda

import (
"context"
"fmt"

"github.com/twmb/franz-go/pkg/kgo"
"go.uber.org/zap"
)

// MessageHandler processes a consumed message.
type MessageHandler func(ctx context.Context, record *kgo.Record) error

// Consumer wraps franz-go client for consuming messages from Redpanda.
type Consumer struct {
client  *kgo.Client
handler MessageHandler
topics  []string
group   string
logger  *zap.Logger
}

// ConsumerConfig configures a new Consumer.
type ConsumerConfig struct {
Connection     ConnectionOptions
Topics         []string
ConsumerGroup  string
Handler        MessageHandler
StartOffset    string
MaxPollRecords int
SessionTimeout int
Logger         *zap.Logger
}

// NewConsumer creates a Redpanda consumer.
func NewConsumer(ctx context.Context, cfg ConsumerConfig) (*Consumer, error) {
if cfg.Handler == nil {
return nil, fmt.Errorf("redpanda: consumer handler is required")
}
if len(cfg.Topics) == 0 {
return nil, fmt.Errorf("redpanda: at least one topic is required")
}

kgoOpts, err := cfg.Connection.BuildKgoOpts()
if err != nil {
return nil, fmt.Errorf("redpanda: build connection opts: %w", err)
}

kgoOpts = append(kgoOpts,
kgo.ConsumerGroup(cfg.ConsumerGroup),
kgo.ConsumeTopics(cfg.Topics...),
)

if cfg.StartOffset == "latest" {
kgoOpts = append(kgoOpts, kgo.ConsumeResetOffset(kgo.NewOffset().AtEnd()))
} else {
kgoOpts = append(kgoOpts, kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()))
}

logger := cfg.Logger
if logger == nil {
logger = zap.NewNop()
}

client, err := kgo.NewClient(kgoOpts...)
if err != nil {
return nil, fmt.Errorf("redpanda: create consumer client: %w", err)
}

return &Consumer{
client:  client,
handler: cfg.Handler,
topics:  cfg.Topics,
group:   cfg.ConsumerGroup,
logger:  logger,
}, nil
}

// Start begins consuming in a loop. Blocks until context is cancelled.
func (c *Consumer) Start(ctx context.Context) error {
for {
fetches := c.client.PollFetches(ctx)
if ctx.Err() != nil {
return ctx.Err()
}

if errs := fetches.Errors(); len(errs) > 0 {
for _, ferr := range errs {
c.logger.Error("fetch error",
zap.String("topic", ferr.Topic),
zap.Int32("partition", ferr.Partition),
zap.Error(ferr.Err))
}
}

fetches.EachRecord(func(record *kgo.Record) {
if err := c.handler(ctx, record); err != nil {
c.logger.Error("message handler error",
zap.String("topic", record.Topic),
zap.Int64("offset", record.Offset),
zap.Error(err))
}
})

if err := c.client.CommitUncommittedOffsets(ctx); err != nil {
c.logger.Error("commit offsets error", zap.Error(err))
}
}
}

// Close stops consuming and commits final offsets.
func (c *Consumer) Close() error {
c.client.Close()
return nil
}

// Healthy returns true if the consumer is connected and consuming.
func (c *Consumer) Healthy(ctx context.Context) bool {
return c.client.Ping(ctx) == nil
}
