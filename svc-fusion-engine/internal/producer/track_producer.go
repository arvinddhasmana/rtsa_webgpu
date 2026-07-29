// CLASSIFICATION: UNCLASSIFIED
package producer

import (
	"context"
	"fmt"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/entity/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/pkg/redpanda"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

// TrackProducer produces FusedTrack messages to the appropriate entity-type topic.
type TrackProducer struct {
	producer    *redpanda.Producer
	topicPrefix string
	logger      *zap.Logger
}

// NewTrackProducer creates a TrackProducer that publishes to <prefix>.<entity_type> topics.
func NewTrackProducer(ctx context.Context, brokers []string, topicPrefix string, logger *zap.Logger) (*TrackProducer, error) {
	p, err := redpanda.NewProducer(ctx, redpanda.ProducerConfig{
		Connection: redpanda.ConnectionOptions{
			Brokers: brokers,
		},
		ServiceName: "svc-fusion-engine",
	})
	if err != nil {
		return nil, fmt.Errorf("track_producer: NewTrackProducer: %w", err)
	}
	return &TrackProducer{producer: p, topicPrefix: topicPrefix, logger: logger}, nil
}

// Produce serialises a FusedTrack and publishes it to the matching entity-type topic.
// The Kafka message key is the track_id for partition affinity.
func (tp *TrackProducer) Produce(ctx context.Context, track *entityv1.FusedTrack) error {
	topic := tp.topicFor(track.GetEntityType())
	payload, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(track)
	if err != nil {
		return fmt.Errorf("track_producer: marshal: %w", err)
	}
	if err := tp.producer.Produce(ctx, topic, []byte(track.GetTrackId()), payload, "UNCLASSIFIED", ""); err != nil {
		return fmt.Errorf("track_producer: produce: %w", err)
	}
	tp.logger.Debug("produced fused track", zap.String("track_id", track.GetTrackId()), zap.String("topic", topic))
	return nil
}

// Close shuts down the underlying Redpanda producer.
func (tp *TrackProducer) Close() {
	if err := tp.producer.Close(); err != nil {
		tp.logger.Warn("track_producer: close error", zap.Error(err))
	}
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
