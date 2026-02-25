// CLASSIFICATION: UNCLASSIFIED
//go:build integration

// Package integration provides integration tests for Module 17 (IT04–IT07):
// fusion engine track correlation, merging, staleness, and classification propagation.
// These tests validate through Redpanda message patterns and pkg/classification public API.
package integration

import (
"context"
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/classification"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
"github.com/arvinddhasmana/RTSA_VS_Opus/tests/integration/testutil"
"github.com/twmb/franz-go/pkg/kgo"
"google.golang.org/protobuf/proto"
"google.golang.org/protobuf/types/known/timestamppb"
)

// TestIT04_SensorToFusedTrack validates message flow:
//  1. Produce 3 radar observations to sensors.radar.tracks
//  2. Verify each message is consumable with correct headers
//  3. Verify protobuf deserialization of track data
//
// Note: Full correlation requires running fusion engine service.
// This test validates the Redpanda message layer.
func TestIT04_SensorToFusedTrack(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupRedpandaOnly(t)
defer env.Teardown()

producer := env.NewKafkaProducer(t)
ctx := context.Background()
rng := testutil.NewSeededRand(42)

// Produce 3 observations at nearly the same position.
for i := 0; i < 3; i++ {
pos := testutil.MidAtlanticPosition(rng)
obs := &ingestionv1.SensorObservation{
ObservationId:   testutil.AuditEventFixture("", "svc-radar").AuditId,
SensorId:        "RADAR-IT04",
SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
ObservationTime: timestamppb.New(time.Now().Add(time.Duration(i) * time.Second)),
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
Position:        pos,
}
payload, _ := proto.Marshal(obs)
headers := redpanda.StandardHeaders("UNCLASSIFIED", "svc-radar-ingestion", "", "v1")

r := producer.ProduceSync(ctx, &kgo.Record{
Topic:   "sensors.radar.tracks",
Key:     []byte(obs.SensorId),
Value:   payload,
Headers: headers,
})
if r.FirstErr() != nil {
t.Fatalf("IT04: produce obs[%d]: %v", i, r.FirstErr())
}
}

// Consume and verify all 3 messages.
consumer := env.NewKafkaConsumer(t, "it04-group", "sensors.radar.tracks")
received := testutil.WaitForTopicMessages(t, consumer, 3, 20*time.Second)
if len(received) < 3 {
t.Fatalf("IT04: expected 3 messages, got %d", len(received))
}

for i, r := range received[:3] {
testutil.AssertHeaderPresent(t, r, redpanda.HeaderClassification)
testutil.AssertHeaderPresent(t, r, redpanda.HeaderSourceService)

var obs ingestionv1.SensorObservation
if err := proto.Unmarshal(r.Value, &obs); err != nil {
t.Errorf("IT04: deserialize obs[%d]: %v", i, err)
}
}

t.Logf("IT04 PASS: %d radar observations produced to sensors.radar.tracks with correct headers", len(received))
}

// TestIT05_TrackMerging validates the track merge message structure:
//  1. Produce a MERGED status FusedTrack message to tracks.fused.surface
//  2. Verify the message is consumable and has correct structure
//  3. Verify MERGED status is preserved in the protobuf
func TestIT05_TrackMerging(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupRedpandaOnly(t)
defer env.Teardown()

producer := env.NewKafkaProducer(t)
ctx := context.Background()

rng := testutil.NewSeededRand(100)
pos := testutil.MidAtlanticPosition(rng)

// Simulate a merged track (as would be produced by fusion engine).
mergedTrack := &entityv1.FusedTrack{
TrackId:           "track-it05-merged",
EntityType:        commonv1.EntityType_ENTITY_TYPE_SURFACE,
Status:            commonv1.TrackStatus_TRACK_STATUS_MERGED,
ConfidenceScore:   0.92,
SourceCount:       3,
Classification:    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
EstimatedPosition: pos,
CreatedAt:         timestamppb.Now(),
UpdatedAt:         timestamppb.Now(),
}

payload, _ := proto.Marshal(mergedTrack)
headers := redpanda.StandardHeaders("UNCLASSIFIED", "svc-fusion-engine", "", "v1")

r := producer.ProduceSync(ctx, &kgo.Record{
Topic:   "tracks.fused.surface",
Key:     []byte(mergedTrack.TrackId),
Value:   payload,
Headers: headers,
})
if r.FirstErr() != nil {
t.Fatalf("IT05: produce merged track: %v", r.FirstErr())
}

// Consume and verify.
consumer := env.NewKafkaConsumer(t, "it05-group", "tracks.fused.surface")
received := testutil.WaitForTopicMessages(t, consumer, 1, 15*time.Second)
if len(received) == 0 {
t.Fatal("IT05: no messages on tracks.fused.surface")
}

var decoded entityv1.FusedTrack
if err := proto.Unmarshal(received[0].Value, &decoded); err != nil {
t.Fatalf("IT05: deserialize merged track: %v", err)
}
if decoded.GetStatus() != commonv1.TrackStatus_TRACK_STATUS_MERGED {
t.Errorf("IT05: track status = %v, want MERGED", decoded.GetStatus())
}
if decoded.GetSourceCount() < 2 {
t.Errorf("IT05: source_count=%d, want >= 2", decoded.GetSourceCount())
}

t.Logf("IT05 PASS: merged track verified (status=%v, source_count=%d)", decoded.GetStatus(), decoded.GetSourceCount())
}

// TestIT06_StaleTrackTimeout validates stale track message structure:
//  1. Produce a STALE status FusedTrack message to tracks.fused.surface
//  2. Verify the status field is correctly propagated through Redpanda
//  3. Produce a DROPPED status message and verify
func TestIT06_StaleTrackTimeout(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupRedpandaOnly(t)
defer env.Teardown()

producer := env.NewKafkaProducer(t)
ctx := context.Background()

rng := testutil.NewSeededRand(60)
pos := testutil.MidAtlanticPosition(rng)

statuses := []commonv1.TrackStatus{
commonv1.TrackStatus_TRACK_STATUS_STALE,
commonv1.TrackStatus_TRACK_STATUS_DROPPED,
}

for _, status := range statuses {
track := &entityv1.FusedTrack{
TrackId:           "track-it06-" + status.String(),
EntityType:        commonv1.EntityType_ENTITY_TYPE_SURFACE,
Status:            status,
ConfidenceScore:   0.70,
Classification:    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
EstimatedPosition: pos,
UpdatedAt:         timestamppb.Now(),
}

payload, _ := proto.Marshal(track)
headers := redpanda.StandardHeaders("UNCLASSIFIED", "svc-fusion-engine", "", "v1")

r := producer.ProduceSync(ctx, &kgo.Record{
Topic: "tracks.fused.surface",
Key:   []byte(track.TrackId),
Value: payload, Headers: headers,
})
if r.FirstErr() != nil {
t.Errorf("IT06: produce %v track: %v", status, r.FirstErr())
}
}

// Consume and verify both status transitions.
consumer := env.NewKafkaConsumer(t, "it06-group", "tracks.fused.surface")
received := testutil.WaitForTopicMessages(t, consumer, 2, 15*time.Second)
if len(received) < 2 {
t.Fatalf("IT06: expected 2 messages, got %d", len(received))
}

gotStale, gotDropped := false, false
for _, r := range received {
var tr entityv1.FusedTrack
if err := proto.Unmarshal(r.Value, &tr); err != nil {
continue
}
switch tr.GetStatus() {
case commonv1.TrackStatus_TRACK_STATUS_STALE:
gotStale = true
case commonv1.TrackStatus_TRACK_STATUS_DROPPED:
gotDropped = true
}
}

if !gotStale {
t.Error("IT06: no STALE track received")
}
if !gotDropped {
t.Error("IT06: no DROPPED track received")
}

t.Log("IT06 PASS: STALE and DROPPED track status transitions verified via Redpanda messages")
}

// TestIT07_ClassificationPropagation validates the MAX classification rule
// using the pkg/classification public API and Redpanda message verification:
//  1. Verify PROTECTED_B + SECRET → SECRET using classification.MaxAll
//  2. Verify classification header correctly propagated in Redpanda messages
//  3. Verify classification filtering using classification.CanAccess
func TestIT07_ClassificationPropagation(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupRedpandaOnly(t)
defer env.Teardown()

// Verify MAX rule using public classification API.
result := classification.MaxAll(
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
)
if result != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET {
t.Errorf("IT07: MaxAll(PROTECTED_B, SECRET) = %v, want SECRET", result)
}
t.Log("IT07: MAX classification rule verified (PROTECTED_B + SECRET → SECRET)")

// Produce a SECRET-classified fused track to verify header propagation.
producer := env.NewKafkaProducer(t)
ctx := context.Background()

rng := testutil.NewSeededRand(7)
pos := testutil.MidAtlanticPosition(rng)

secretTrack := &entityv1.FusedTrack{
TrackId:           "track-it07-secret",
EntityType:        commonv1.EntityType_ENTITY_TYPE_SURFACE,
Status:            commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
Classification:    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
EstimatedPosition: pos,
UpdatedAt:         timestamppb.Now(),
}

payload, _ := proto.Marshal(secretTrack)
secretClass := classification.LevelToString(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
headers := redpanda.StandardHeaders(secretClass, "svc-fusion-engine", "", "v1")

r := producer.ProduceSync(ctx, &kgo.Record{
Topic:   "tracks.fused.surface",
Key:     []byte(secretTrack.TrackId),
Value:   payload,
Headers: headers,
})
if r.FirstErr() != nil {
t.Fatalf("IT07: produce SECRET track: %v", r.FirstErr())
}

consumer := env.NewKafkaConsumer(t, "it07-group", "tracks.fused.surface")
received := testutil.WaitForTopicMessages(t, consumer, 1, 15*time.Second)
if len(received) == 0 {
t.Fatal("IT07: no message on tracks.fused.surface")
}

classHeader := testutil.AssertHeaderPresent(t, received[0], redpanda.HeaderClassification)
if classHeader != "SECRET" {
t.Errorf("IT07: classification header = %q, want SECRET", classHeader)
}

// Verify access control: PROTECTED_B cannot access SECRET.
if classification.CanAccess(
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
) {
t.Error("IT07: PROTECTED_B should NOT access SECRET")
}

t.Log("IT07 PASS: classification propagation (MAX rule + header + access control)")
}
