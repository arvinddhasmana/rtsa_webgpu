package source
// CLASSIFICATION: UNCLASSIFIED
// Package source adapts the shared Redpanda consumer into a
// webtransport.TrackSource. It consumes Protobuf TrackUpdate messages from the
// fused-track topics, serialises each into a fixed 128-byte GPU-aligned record
// and fans the records out to every connected WebTransport session.
//
// A single consumer goroutine owns the (stateful, non-concurrent) FlatBuffer
// serialiser; fan-out to per-session channels is non-blocking so that one slow
// browser client cannot back-pressure the hot path — records for a saturated
// client are dropped rather than stalling every other operator.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/webtransport_guidelines.md §3
package source

import (
	"context"
	"fmt"

	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/flatbuf"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// Config configures a Source.
type Config struct {
	// Brokers is the list of Redpanda broker addresses.
	Brokers []string
	// Topics is the set of fused-track topics to consume.
	Topics []string
	// ConsumerGroup is the Kafka consumer group.
	ConsumerGroup string
	// StartOffset is "latest" (hot path) or "earliest".
	StartOffset string
	// SubscriberBufferSize is the per-session broadcast channel buffer.
	SubscriberBufferSize int
	// ClientID identifies this consumer to the broker.
	ClientID string

	// TLSEnabled toggles mTLS on the broker connection.
	TLSEnabled bool
	// TLSCAFile is the CA bundle used to verify the broker.
	TLSCAFile string
	// TLSCertFile is the client certificate for broker mTLS.
	TLSCertFile string
	// TLSKeyFile is the client private key for broker mTLS.
	TLSKeyFile string
}

// Source consumes fused tracks from Redpanda and broadcasts serialised records
// to WebTransport sessions. It implements webtransport.TrackSource.
type Source struct {
	hub        *hub
	serializer *flatbuf.Serializer
	consumer   *redpanda.Consumer
	logger     *zap.Logger

	mConsumed   metric.Int64Counter
	mDecodeErrs metric.Int64Counter
	mDropped    metric.Int64Counter
}

// New builds a Source and its underlying Redpanda consumer. The consumer does
// not start until Run is called.
func New(ctx context.Context, cfg Config, logger *zap.Logger, meter metric.Meter) (*Source, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	s := &Source{
		hub:        newHub(cfg.SubscriberBufferSize),
		serializer: flatbuf.NewSerializer(),
		logger:     logger,
	}
	if err := s.initMetrics(meter); err != nil {
		return nil, err
	}

	consumer, err := redpanda.NewConsumer(ctx, redpanda.ConsumerConfig{
		Connection: redpanda.ConnectionOptions{
			Brokers:     cfg.Brokers,
			ClientID:    cfg.ClientID,
			TLSEnabled:  cfg.TLSEnabled,
			TLSCAFile:   cfg.TLSCAFile,
			TLSCertFile: cfg.TLSCertFile,
			TLSKeyFile:  cfg.TLSKeyFile,
		},
		Topics:        cfg.Topics,
		ConsumerGroup: cfg.ConsumerGroup,
		StartOffset:   cfg.StartOffset,
		Handler:       s.handle,
		Logger:        logger,
	})
	if err != nil {
		return nil, fmt.Errorf("source: create consumer: %w", err)
	}
	s.consumer = consumer
	return s, nil
}

// Subscribe implements webtransport.TrackSource. The returned channel receives
// 128-byte records and is closed when ctx is cancelled.
func (s *Source) Subscribe(ctx context.Context) <-chan []byte {
	return s.hub.subscribe(ctx)
}

// Run starts consuming and blocks until ctx is cancelled or a fatal error
// occurs.
func (s *Source) Run(ctx context.Context) error {
	return s.consumer.Start(ctx)
}

// Close releases all subscriber channels and closes the consumer.
func (s *Source) Close() error {
	s.hub.closeAll()
	if s.consumer != nil {
		return s.consumer.Close()
	}
	return nil
}

// Healthy reports whether the underlying consumer can reach the broker.
func (s *Source) Healthy(ctx context.Context) bool {
	return s.consumer != nil && s.consumer.Healthy(ctx)
}

// SubscriberCount returns the number of active WebTransport subscribers.
func (s *Source) SubscriberCount() int {
	return s.hub.subscriberCount()
}

// handle decodes a fused-track record, serialises it and broadcasts it. A
// malformed record is counted and skipped — it never fails the consumer, which
// would stall the entire hot path.
func (s *Source) handle(ctx context.Context, rec *kgo.Record) error {
	record, ok := s.decodeAndSerialize(rec.Value)
	if !ok {
		if s.mDecodeErrs != nil {
			s.mDecodeErrs.Add(ctx, 1)
		}
		return nil
	}
	if s.mConsumed != nil {
		s.mConsumed.Add(ctx, 1)
	}
	if _, dropped := s.hub.broadcast(record); dropped > 0 && s.mDropped != nil {
		s.mDropped.Add(ctx, int64(dropped))
	}
	return nil
}

// decodeAndSerialize unmarshals a Protobuf TrackUpdate and serialises it into a
// freshly allocated 128-byte record. It returns false for nil, malformed or
// position-less messages.
func (s *Source) decodeAndSerialize(value []byte) ([]byte, bool) {
	var update entityv1.TrackUpdate
	if err := proto.Unmarshal(value, &update); err != nil {
		return nil, false
	}
	rec, ok := s.serializer.Serialize(&update)
	if !ok {
		return nil, false
	}
	out := make([]byte, flatbuf.RecordSize)
	copy(out, rec[:])
	return out, true
}

func (s *Source) initMetrics(meter metric.Meter) error {
	if meter == nil {
		return nil
	}
	var err error
	if s.mConsumed, err = meter.Int64Counter(
		"rtsa_wt_source_records_total",
		metric.WithDescription("Fused-track records consumed and serialised"),
	); err != nil {
		return fmt.Errorf("source: consumed counter: %w", err)
	}
	if s.mDecodeErrs, err = meter.Int64Counter(
		"rtsa_wt_source_decode_errors_total",
		metric.WithDescription("Fused-track messages that failed to decode or serialise"),
	); err != nil {
		return fmt.Errorf("source: decode-error counter: %w", err)
	}
	if s.mDropped, err = meter.Int64Counter(
		"rtsa_wt_source_broadcast_dropped_total",
		metric.WithDescription("Records dropped because a subscriber channel was full"),
	); err != nil {
		return fmt.Errorf("source: dropped counter: %w", err)
	}
	return nil
}
