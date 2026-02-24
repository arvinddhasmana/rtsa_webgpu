// CLASSIFICATION: UNCLASSIFIED
package producer

import (
	"context"
	"fmt"

	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// TrackProducer produces FusedTrack messages to the appropriate entity-type topic.
type TrackProducer struct {
	producer    *redpanda.Producer
	topicPrefix string
	logger      *zap.Logger
}

// NewTrackProducer creates a TrackProducer that publishes to <prefix>.<entity_type> topics.
func NewTrackProducer(brokers []string, topicPrefix string, logger *zap.Logger) (*TrackProducer, error) {
	p, err := redpanda.NewProducer(brokers)
	if err != nil {
		return nil, fmt.Errorf("track_producer: NewTrackProducer: %w", err)
	}
	return &TrackProducer{producer: p, topicPrefix: topicPrefix, logger: logger}, nil
}

// Produce serialises a FusedTrack and publishes it to the matching entity-type topic.
// The Kafka message key is the track_id for partition affinity.
func (tp *TrackProducer) Produce(ctx context.Context, track *entityv1.FusedTrack) error {
	topic := tp.topicFor(track.GetEntityType())
	payload, err := proto.Marshal(track)
	if err != nil {
		return fmt.Errorf("track_producer: marshal: %w", err)
	}
	record := &kgo.Record{
		Topic: topic,
		Key:   []byte(track.GetTrackId()),
		Value: payload,
	}
	if err := tp.producer.Produce(ctx, record); err != nil {
		return fmt.Errorf("track_producer: produce: %w", err)
	}
	tp.logger.Debug("produced fused track", zap.String("track_id", track.GetTrackId()), zap.String("topic", topic))
	return nil
}

// Close shuts down the underlying Redpanda producer.
func (tp *TrackProducer) Close() {
	tp.producer.Close()
}

// topicFor returns the output topic for the given entity type.
func (tp *TrackProducer) topicFor(et commonv1.EntityType) string {
	suffix := "surface"
	switch et {
	case commonv1.EntityType_ENTITY_TYPE_AIR:
		suffix = "air"
	case commonv1.EntityType_ENTITY_TYPE_SUBSURFACE:
		suffix = "subsurface"
	case commonv1.EntityType_ENTITY_TYPE_LAND:
		suffix = "land"
	case commonv1.EntityType_ENTITY_TYPE_CYBER:
		suffix = "cyber"
	}
	return tp.topicPrefix + "." + suffix
}
