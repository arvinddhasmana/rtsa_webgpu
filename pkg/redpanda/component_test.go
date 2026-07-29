// CLASSIFICATION: UNCLASSIFIED
package redpanda_test

import (
	"context"
	"testing"
	"time"

	"github.com/arvinddhasmana/rtsa_webgpu/pkg/redpanda"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestNewConsumer_Validation(t *testing.T) {
	ctx := context.Background()

	t.Run("missing handler", func(t *testing.T) {
		_, err := redpanda.NewConsumer(ctx, redpanda.ConsumerConfig{
			Topics: []string{"test"},
		})
		if err == nil {
			t.Error("expected error for missing handler")
		}
	})

	t.Run("missing topics", func(t *testing.T) {
		_, err := redpanda.NewConsumer(ctx, redpanda.ConsumerConfig{
			Handler: func(ctx context.Context, record *kgo.Record) error { return nil },
		})
		if err == nil {
			t.Error("expected error for missing topics")
		}
	})

	t.Run("invalid connection opts", func(t *testing.T) {
		_, err := redpanda.NewConsumer(ctx, redpanda.ConsumerConfig{
			Handler: func(ctx context.Context, record *kgo.Record) error { return nil },
			Topics:  []string{"test"},
			Connection: redpanda.ConnectionOptions{
				SASL: &redpanda.SASLConfig{Mechanism: "INVALID"},
			},
		})
		if err == nil {
			t.Error("expected error for invalid connection opts")
		}
	})
}

func TestNewProducer_ConnectError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// NewProducer calls Ping, which should fail quickly with no brokers
	_, err := redpanda.NewProducer(ctx, redpanda.ProducerConfig{
		Connection: redpanda.ConnectionOptions{
			Brokers: []string{"localhost:1"}, // Invalid port
		},
	})
	if err == nil {
		t.Error("expected error for unreachable brokers in NewProducer")
	}
}

func TestProducer_CompressionAndAcks(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	configs := []redpanda.ProducerConfig{
		{Compression: "snappy", RequiredAcks: "leader"},
		{Compression: "lz4", RequiredAcks: "none"},
		{Compression: "zstd", RequiredAcks: "all"},
	}

	for _, cfg := range configs {
		// Just verify it attempts to build (it will fail on Ping)
		_, _ = redpanda.NewProducer(ctx, cfg)
	}
}
