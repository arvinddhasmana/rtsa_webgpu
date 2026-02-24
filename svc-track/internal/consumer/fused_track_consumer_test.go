// CLASSIFICATION: UNCLASSIFIED
// Package consumer — unit tests for FusedTrackConsumer message handling.
package consumer

import (
	"context"
	"log/slog"
	"os"
	"testing"

	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	"google.golang.org/protobuf/proto"
	"github.com/twmb/franz-go/pkg/kgo"
)

// mockTrackPutter records Put calls for test assertions.
type mockTrackPutter struct {
	puts []*entityv1.FusedTrack
}

func (m *mockTrackPutter) Put(track *entityv1.FusedTrack) {
	m.puts = append(m.puts, track)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// TestHandleRecord_ValidTrack: well-formed record puts track into cache.
func TestHandleRecord_ValidTrack(t *testing.T) {
	mock := &mockTrackPutter{}
	c := &FusedTrackConsumer{cache: mock, logger: testLogger()}

	track := &entityv1.FusedTrack{
		TrackId:         "trk-consumer-01",
		EntityType:      commonv1.EntityType_ENTITY_TYPE_AIR,
		Status:          commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
		ConfidenceScore: 0.9,
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
	}
	payload, err := proto.Marshal(track)
	if err != nil {
		t.Fatalf("proto.Marshal: %v", err)
	}

	record := &kgo.Record{
		Topic:  "tracks.fused.air",
		Offset: 1,
		Value:  payload,
	}

	if err := c.handleRecord(context.Background(), record); err != nil {
		t.Fatalf("handleRecord error: %v", err)
	}

	if len(mock.puts) != 1 {
		t.Fatalf("expected 1 Put call, got %d", len(mock.puts))
	}
	if mock.puts[0].TrackId != "trk-consumer-01" {
		t.Errorf("expected track_id=trk-consumer-01, got %q", mock.puts[0].TrackId)
	}
}

// TestHandleRecord_InvalidProto: malformed bytes returns error, no cache put.
func TestHandleRecord_InvalidProto(t *testing.T) {
	mock := &mockTrackPutter{}
	c := &FusedTrackConsumer{cache: mock, logger: testLogger()}

	record := &kgo.Record{
		Topic:  "tracks.fused.air",
		Offset: 2,
		Value:  []byte("not-a-proto"),
	}

	err := c.handleRecord(context.Background(), record)
	if err == nil {
		t.Fatal("expected error for malformed proto, got nil")
	}
	if len(mock.puts) != 0 {
		t.Errorf("expected no Put calls on error, got %d", len(mock.puts))
	}
}

// TestHandleRecord_EmptyTrackID: track with empty ID returns error.
func TestHandleRecord_EmptyTrackID(t *testing.T) {
	mock := &mockTrackPutter{}
	c := &FusedTrackConsumer{cache: mock, logger: testLogger()}

	track := &entityv1.FusedTrack{
		TrackId:    "", // empty — invalid
		EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
	}
	payload, err := proto.Marshal(track)
	if err != nil {
		t.Fatalf("proto.Marshal: %v", err)
	}

	record := &kgo.Record{Topic: "tracks.fused.surface", Offset: 3, Value: payload}
	if err := c.handleRecord(context.Background(), record); err == nil {
		t.Fatal("expected error for empty track_id, got nil")
	}
	if len(mock.puts) != 0 {
		t.Errorf("expected no Put calls, got %d", len(mock.puts))
	}
}

// TestFusedTopics: verifies exactly 5 topics are configured.
func TestFusedTopics(t *testing.T) {
	expected := []string{
		"tracks.fused.surface",
		"tracks.fused.air",
		"tracks.fused.subsurface",
		"tracks.fused.land",
		"tracks.fused.cyber",
	}
	if len(FusedTopics) != len(expected) {
		t.Errorf("expected %d topics, got %d", len(expected), len(FusedTopics))
	}
	for i, topic := range expected {
		if FusedTopics[i] != topic {
			t.Errorf("topic[%d]: expected %q, got %q", i, topic, FusedTopics[i])
		}
	}
}

// TestSlogKgoLogger_Level: verifies Level() returns configured level.
func TestSlogKgoLogger_Level(t *testing.T) {
l := newSlogKgoLogger(testLogger(), kgo.LogLevelWarn)
if l.Level() != kgo.LogLevelWarn {
t.Errorf("expected LogLevelWarn, got %v", l.Level())
}
}

// TestSlogKgoLogger_Log: verifies Log does not panic for all levels.
func TestSlogKgoLogger_Log(t *testing.T) {
l := newSlogKgoLogger(testLogger(), kgo.LogLevelDebug)
levels := []kgo.LogLevel{kgo.LogLevelError, kgo.LogLevelWarn, kgo.LogLevelInfo, kgo.LogLevelDebug}
for _, lvl := range levels {
l.Log(lvl, "test message", "key", "value")
}
}

// TestClassificationFromTopic: returns UNCLASSIFIED for any topic.
func TestClassificationFromTopic(t *testing.T) {
topics := []string{"tracks.fused.air", "tracks.fused.surface", "unknown"}
for _, topic := range topics {
got := classificationFromTopic(topic)
if got != 1 { // CLASSIFICATION_LEVEL_UNCLASSIFIED = 1
t.Errorf("expected UNCLASSIFIED for topic %q, got %v", topic, got)
}
}
}
