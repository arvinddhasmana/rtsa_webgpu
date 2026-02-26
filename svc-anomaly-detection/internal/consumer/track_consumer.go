// CLASSIFICATION: UNCLASSIFIED
package consumer

import (
	"context"
	"fmt"
	"log/slog"

	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// TrackHandler is a function that processes a decoded FusedTrack.
type TrackHandler func(ctx context.Context, track *entityv1.FusedTrack) error

// Message represents a consumed message payload.
type Message struct {
	Topic  string
	Value  []byte
	Offset int64
}

// MessageHandler processes consumed messages.
type MessageHandler func(ctx context.Context, msg *Message) error

// MessageConsumer abstracts track consumption for testability.
type MessageConsumer interface {
	Consume(ctx context.Context, topics []string, handler MessageHandler) error
	Close() error
}

// TrackConsumer consumes FusedTrack messages from tracks.fused.* topics.
type TrackConsumer struct {
	consumer MessageConsumer
	logger   *slog.Logger
}

// NewTrackConsumer creates a new TrackConsumer.
func NewTrackConsumer(consumer MessageConsumer, logger *slog.Logger) *TrackConsumer {
	return &TrackConsumer{
		consumer: consumer,
		logger:   logger,
	}
}

// Start begins consuming from the given topics, calling handler for each decoded track.
func (tc *TrackConsumer) Start(ctx context.Context, topics []string, handler TrackHandler) error {
	return tc.consumer.Consume(ctx, topics, func(ctx context.Context, msg *Message) error {
		track, err := tc.decodeTrack(msg.Value)
		if err != nil {
			tc.logger.Warn("failed to decode fused track",
				"topic", msg.Topic,
				"offset", msg.Offset,
				"error", err,
			)
			return nil // Don't propagate decode errors — log and continue.
		}
		return handler(ctx, track)
	})
}

// Close stops the consumer.
func (tc *TrackConsumer) Close() error {
	return tc.consumer.Close()
}

// decodeTrack deserialises a protojson-encoded FusedTrack from raw bytes.
func (tc *TrackConsumer) decodeTrack(data []byte) (*entityv1.FusedTrack, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("[consumer.TrackConsumer.decodeTrack]: empty message payload")
	}
	track := &entityv1.FusedTrack{}
	if err := protojson.Unmarshal(data, track); err != nil {
		return nil, fmt.Errorf("[consumer.TrackConsumer.decodeTrack]: %w", err)
	}
	if track.GetTrackId() == "" {
		return nil, fmt.Errorf("[consumer.TrackConsumer.decodeTrack]: track_id is empty")
	}
	return track, nil
}
