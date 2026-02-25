// CLASSIFICATION: UNCLASSIFIED
package consumer

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

// RedpandaMessageConsumer adapts the shared Redpanda client to MessageConsumer.
type RedpandaMessageConsumer struct {
	brokers []string
	group   string
	logger  *slog.Logger
	client  *redpanda.Consumer
}

// NewRedpandaMessageConsumer creates a Redpanda adapter-backed message consumer.
func NewRedpandaMessageConsumer(brokers []string, group string, logger *slog.Logger) *RedpandaMessageConsumer {
	return &RedpandaMessageConsumer{brokers: brokers, group: group, logger: logger}
}

// Consume starts Redpanda consumption and forwards records to the provided handler.
func (c *RedpandaMessageConsumer) Consume(ctx context.Context, topics []string, handler MessageHandler) error {
	client, err := redpanda.NewConsumer(ctx, redpanda.ConsumerConfig{
		Connection:    redpanda.ConnectionOptions{Brokers: c.brokers},
		Topics:        topics,
		ConsumerGroup: c.group,
		Logger:        zap.NewNop(),
		Handler: func(ctx context.Context, record *kgo.Record) error {
			return handler(ctx, &Message{Topic: record.Topic, Value: record.Value, Offset: record.Offset})
		},
	})
	if err != nil {
		return fmt.Errorf("[consumer.RedpandaMessageConsumer.Consume]: %w", err)
	}
	c.client = client

	if err := c.client.Start(ctx); err != nil && ctx.Err() == nil {
		return fmt.Errorf("[consumer.RedpandaMessageConsumer.Consume]: %w", err)
	}
	c.logger.Debug("redpanda consumer stopped")
	return nil
}

// Close closes the underlying Redpanda consumer.
func (c *RedpandaMessageConsumer) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}
