// CLASSIFICATION: UNCLASSIFIED
package consumer

import (
	"context"
	"fmt"

	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
	"go.uber.org/zap"
)

// SensorConsumer wraps the shared Redpanda consumer for multi-topic sensor observation consumption.
type SensorConsumer struct {
	consumer *redpanda.Consumer
	logger   *zap.Logger
}

// NewSensorConsumer creates a SensorConsumer connected to the given brokers and topics.
func NewSensorConsumer(brokers []string, group string, topics []string, logger *zap.Logger) (*SensorConsumer, error) {
	c, err := redpanda.NewConsumer(brokers, group, topics)
	if err != nil {
		return nil, fmt.Errorf("sensor_consumer: NewSensorConsumer: %w", err)
	}
	return &SensorConsumer{consumer: c, logger: logger}, nil
}

// Run starts consuming records and dispatches each to the handler.
// Returns when ctx is cancelled.
func (sc *SensorConsumer) Run(ctx context.Context, handler redpanda.RecordHandler) error {
	sc.logger.Info("sensor consumer started")
	err := sc.consumer.Run(ctx, handler)
	if err != nil && ctx.Err() == nil {
		return fmt.Errorf("sensor_consumer: Run: %w", err)
	}
	return nil
}

// Close shuts down the consumer client.
func (sc *SensorConsumer) Close() {
	sc.consumer.Close()
}
