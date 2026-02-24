// CLASSIFICATION: UNCLASSIFIED
// Package consumer provides the Redpanda consumer for fused track topics.
//
// The FusedTrackConsumer subscribes to all 5 tracks.fused.* topics and
// calls TrackCache.Put for each deserialized FusedTrack message.
//
// Feature: FEAT-13 Situational Awareness UI
// UC: UC012 Situational Awareness UI
// Requirements: CR-UI-001, CR-ING-010
package consumer

import (
"context"
"fmt"
"log/slog"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
"github.com/twmb/franz-go/pkg/kgo"
"google.golang.org/protobuf/proto"
)

// FusedTopics is the ordered list of Redpanda topics consumed by this service.
var FusedTopics = []string{
"tracks.fused.surface",
"tracks.fused.air",
"tracks.fused.subsurface",
"tracks.fused.land",
"tracks.fused.cyber",
}

// TrackPutter is the interface the consumer uses to push decoded tracks into state.
type TrackPutter interface {
Put(track *entityv1.FusedTrack)
}

// FusedTrackConsumer consumes protobuf-encoded FusedTrack messages from Redpanda
// and forwards them to a TrackPutter (typically the domain.TrackCache).
type FusedTrackConsumer struct {
client *kgo.Client
cache  TrackPutter
logger *slog.Logger
}

// NewFusedTrackConsumer creates a consumer connected to the given brokers, consuming
// from all tracks.fused.* topics with the specified consumer group.
func NewFusedTrackConsumer(brokers []string, groupID string, cache TrackPutter, logger *slog.Logger) (*FusedTrackConsumer, error) {
opts := []kgo.Opt{
kgo.SeedBrokers(brokers...),
kgo.ConsumerGroup(groupID),
kgo.ConsumeTopics(FusedTopics...),
kgo.WithLogger(newSlogKgoLogger(logger, kgo.LogLevelWarn)),
}

client, err := kgo.NewClient(opts...)
if err != nil {
return nil, fmt.Errorf("consumer.NewFusedTrackConsumer: kgo.NewClient: %w", err)
}

return &FusedTrackConsumer{
client: client,
cache:  cache,
logger: logger,
}, nil
}

// Run starts the consume loop. It blocks until ctx is cancelled or an
// unrecoverable error occurs.
func (c *FusedTrackConsumer) Run(ctx context.Context) error {
c.logger.InfoContext(ctx, "fused track consumer starting", slog.Any("topics", FusedTopics))

for {
fetches := c.client.PollFetches(ctx)
if fetches.IsClientClosed() {
return nil
}

// Check for context cancellation.
if err := ctx.Err(); err != nil {
return nil // normal shutdown
}

fetches.EachError(func(t string, p int32, err error) {
c.logger.ErrorContext(ctx, "fetch error",
slog.String("topic", t),
slog.Int("partition", int(p)),
slog.String("error", err.Error()),
)
})

fetches.EachRecord(func(record *kgo.Record) {
if err := c.handleRecord(ctx, record); err != nil {
// Log and continue — do not stop the consumer for one bad message.
c.logger.WarnContext(ctx, "failed to process record",
slog.String("topic", record.Topic),
slog.Int64("offset", record.Offset),
slog.String("error", err.Error()),
)
}
})
}
}

// handleRecord deserializes a single Kafka record and puts it into the cache.
func (c *FusedTrackConsumer) handleRecord(ctx context.Context, record *kgo.Record) error {
var track entityv1.FusedTrack
if err := proto.Unmarshal(record.Value, &track); err != nil {
return fmt.Errorf("consumer.handleRecord(topic=%s, offset=%d): proto.Unmarshal: %w",
record.Topic, record.Offset, err)
}

// Validate minimum required fields.
if track.TrackId == "" {
return fmt.Errorf("consumer.handleRecord(topic=%s, offset=%d): track_id is empty",
record.Topic, record.Offset)
}

c.logger.DebugContext(ctx, "track received",
slog.String("track_id", track.TrackId),
slog.String("entity_type", track.EntityType.String()),
slog.String("status", track.Status.String()),
slog.String("topic", record.Topic),
)

c.cache.Put(&track)
return nil
}

// Close shuts down the Kafka client gracefully.
func (c *FusedTrackConsumer) Close() {
c.client.Close()
}

// HealthCheck verifies connectivity to Redpanda.
func (c *FusedTrackConsumer) HealthCheck(ctx context.Context) error {
if err := c.client.Ping(ctx); err != nil {
return fmt.Errorf("consumer.HealthCheck: broker ping failed: %w", err)
}
return nil
}

// classificationFromTopic infers classification from the topic name.
// The FusedTrack proto field is authoritative; this is defence-in-depth only.
func classificationFromTopic(topic string) commonv1.ClassificationLevel {
_ = topic
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED
}

// slogKgoLogger adapts slog.Logger to the kgo.Logger interface.
type slogKgoLogger struct {
logger *slog.Logger
level  kgo.LogLevel
}

func newSlogKgoLogger(l *slog.Logger, level kgo.LogLevel) kgo.Logger {
return &slogKgoLogger{logger: l, level: level}
}

// Level implements kgo.Logger.
func (a *slogKgoLogger) Level() kgo.LogLevel {
return a.level
}

// Log implements kgo.Logger.
func (a *slogKgoLogger) Log(level kgo.LogLevel, msg string, keyVals ...any) {
switch level {
case kgo.LogLevelError:
a.logger.Error(msg, keyVals...)
case kgo.LogLevelWarn:
a.logger.Warn(msg, keyVals...)
case kgo.LogLevelInfo:
a.logger.Info(msg, keyVals...)
default:
a.logger.Debug(msg, keyVals...)
}
}
