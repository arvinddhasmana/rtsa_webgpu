// CLASSIFICATION: UNCLASSIFIED
package producer

import (
	"context"
	"fmt"

	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
	"github.com/twmb/franz-go/pkg/kgo"
)

// RedpandaMessageProducer adapts the shared Redpanda producer to MessageProducer.
type RedpandaMessageProducer struct {
	producer *redpanda.Producer
}

// NewRedpandaMessageProducer creates an adapter-backed Redpanda producer.
func NewRedpandaMessageProducer(ctx context.Context, brokers []string, serviceName string) (*RedpandaMessageProducer, error) {
	p, err := redpanda.NewProducer(ctx, redpanda.ProducerConfig{
		Connection:  redpanda.ConnectionOptions{Brokers: brokers},
		ServiceName: serviceName,
	})
	if err != nil {
		return nil, fmt.Errorf("[producer.NewRedpandaMessageProducer]: %w", err)
	}
	return &RedpandaMessageProducer{producer: p}, nil
}

// Produce publishes a message to Redpanda with converted headers.
func (p *RedpandaMessageProducer) Produce(ctx context.Context, topic string, key, value []byte, headers map[string]string) error {
	extraHeaders := make([]kgo.RecordHeader, 0, len(headers))
	for k, v := range headers {
		extraHeaders = append(extraHeaders, kgo.RecordHeader{Key: k, Value: []byte(v)})
	}
	return p.producer.Produce(ctx, topic, key, value, "UNCLASSIFIED", "", extraHeaders...)
}

// Close shuts down the underlying Redpanda producer.
func (p *RedpandaMessageProducer) Close() error {
	return p.producer.Close()
}
