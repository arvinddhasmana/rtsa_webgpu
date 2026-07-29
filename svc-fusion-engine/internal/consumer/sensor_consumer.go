// CLASSIFICATION: UNCLASSIFIED
package consumer

import (
	"context"
	"fmt"

	"github.com/arvinddhasmana/rtsa_webgpu/pkg/redpanda"
	"go.uber.org/zap"
)

// SensorConsumer wraps the shared Redpanda consumer for multi-topic sensor observation consumption.
type SensorConsumer struct {
	brokers    []string
	group      string
	topics     []string
	underlying *redpanda.Consumer
	logger     *zap.Logger
}

// NewSensorConsumer creates a SensorConsumer that will connect to the given brokers and topics.
// The actual Redpanda connection is established when Run is called.
func NewSensorConsumer(brokers []string, group string, topics []string, logger *zap.Logger) (*SensorConsumer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("sensor_consumer: at least one broker required")
	}
	if len(topics) == 0 {
		return nil, fmt.Errorf("sensor_consumer: at least one topic required")
	}
	return &SensorConsumer{
		brokers: brokers,
		group:   group,
		topics:  topics,
		logger:  logger,
	}, nil
}

// Run starts consuming records and dispatches each to the handler.
// Returns when ctx is cancelled.
func (sc *SensorConsumer) Run(ctx context.Context, handler redpanda.MessageHandler) error {
	c, err := redpanda.NewConsumer(ctx, redpanda.ConsumerConfig{
		Connection: redpanda.ConnectionOptions{
			Brokers: sc.brokers,
		},
		Topics:        sc.topics,
		ConsumerGroup: sc.group,
		Handler:       handler,
		Logger:        sc.logger,
	})
	if err != nil {
		return fmt.Errorf("sensor_consumer: NewConsumer: %w", err)
	}
	sc.underlying = c
	sc.logger.Info("sensor consumer started")
	err = c.Start(ctx)
	if err != nil && ctx.Err() == nil {
		return fmt.Errorf("sensor_consumer: Run: %w", err)
	}
	return nil
}

// Close shuts down the consumer client if it has been started.
func (sc *SensorConsumer) Close() {
	if sc.underlying != nil {
		sc.underlying.Close() //nolint:errcheck
	}
}
