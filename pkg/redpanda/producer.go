// CLASSIFICATION: UNCLASSIFIED
package redpanda

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Producer wraps a franz-go Kafka client for producing records.
type Producer struct {
	client *kgo.Client
}

// NewProducer creates a new Redpanda producer with the given brokers and options.
func NewProducer(brokers []string, opts ...kgo.Opt) (*Producer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("redpanda: NewProducer: at least one broker required")
	}
	baseOpts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
	}
	baseOpts = append(baseOpts, opts...)
	client, err := kgo.NewClient(baseOpts...)
	if err != nil {
		return nil, fmt.Errorf("redpanda: NewProducer: %w", err)
	}
	return &Producer{client: client}, nil
}

// Produce synchronously produces all records and returns the first error encountered.
func (p *Producer) Produce(ctx context.Context, records ...*kgo.Record) error {
	results := p.client.ProduceSync(ctx, records...)
	for _, result := range results {
		if result.Err != nil {
			return fmt.Errorf("redpanda: Produce: %w", result.Err)
		}
	}
	return nil
}

// Close shuts down the producer client.
func (p *Producer) Close() {
	p.client.Close()
}
