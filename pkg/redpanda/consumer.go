// CLASSIFICATION: UNCLASSIFIED
package redpanda

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

// RecordHandler is a function that processes a single Kafka record.
type RecordHandler func(ctx context.Context, record *kgo.Record) error

// Consumer wraps a franz-go Kafka client for consuming records.
type Consumer struct {
	client *kgo.Client
}

// NewConsumer creates a new Redpanda consumer for the given group and topics.
func NewConsumer(brokers []string, group string, topics []string, opts ...kgo.Opt) (*Consumer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("redpanda: NewConsumer: at least one broker required")
	}
	baseOpts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topics...),
	}
	baseOpts = append(baseOpts, opts...)
	client, err := kgo.NewClient(baseOpts...)
	if err != nil {
		return nil, fmt.Errorf("redpanda: NewConsumer: %w", err)
	}
	return &Consumer{client: client}, nil
}

// Run polls for records and calls handler for each one, committing offsets after each batch.
// Returns when ctx is cancelled.
func (c *Consumer) Run(ctx context.Context, handler RecordHandler) error {
	for {
		fetches := c.client.PollFetches(ctx)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if errs := fetches.Errors(); len(errs) > 0 {
			return fmt.Errorf("redpanda: poll: %v", errs[0].Err)
		}
		fetches.EachRecord(func(r *kgo.Record) {
			_ = handler(ctx, r)
		})
		if err := c.client.CommitUncommittedOffsets(ctx); err != nil && ctx.Err() == nil {
			return fmt.Errorf("redpanda: commit: %w", err)
		}
	}
}

// Close shuts down the consumer client.
func (c *Consumer) Close() {
	c.client.Close()
}
