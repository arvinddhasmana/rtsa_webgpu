// CLASSIFICATION: UNCLASSIFIED
package consumer

import (
"context"
"fmt"
"log/slog"

entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
"google.golang.org/protobuf/proto"
)

// TrackHandler is a function that processes a decoded FusedTrack.
type TrackHandler func(ctx context.Context, track *entityv1.FusedTrack) error

// TrackConsumer consumes FusedTrack messages from tracks.fused.* topics.
type TrackConsumer struct {
consumer redpanda.MessageConsumer
logger   *slog.Logger
}

// NewTrackConsumer creates a new TrackConsumer.
func NewTrackConsumer(consumer redpanda.MessageConsumer, logger *slog.Logger) *TrackConsumer {
return &TrackConsumer{
consumer: consumer,
logger:   logger,
}
}

// Start begins consuming from the given topics, calling handler for each decoded track.
func (tc *TrackConsumer) Start(ctx context.Context, topics []string, handler TrackHandler) error {
return tc.consumer.Consume(ctx, topics, func(ctx context.Context, msg *redpanda.Message) error {
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

// decodeTrack deserialises a protobuf-encoded FusedTrack from raw bytes.
func (tc *TrackConsumer) decodeTrack(data []byte) (*entityv1.FusedTrack, error) {
if len(data) == 0 {
return nil, fmt.Errorf("[consumer.TrackConsumer.decodeTrack]: empty message payload")
}
track := &entityv1.FusedTrack{}
if err := proto.Unmarshal(data, track); err != nil {
return nil, fmt.Errorf("[consumer.TrackConsumer.decodeTrack]: %w", err)
}
if track.GetTrackId() == "" {
return nil, fmt.Errorf("[consumer.TrackConsumer.decodeTrack]: track_id is empty")
}
return track, nil
}
