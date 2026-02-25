// CLASSIFICATION: UNCLASSIFIED
package consumer

import (
"context"
"testing"

entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
"google.golang.org/protobuf/proto"
"log/slog"
)

func TestTrackConsumer_DecodeTrack_Valid(t *testing.T) {
tc := &TrackConsumer{logger: slog.Default()}

track := &entityv1.FusedTrack{
TrackId: "test-track-001",
}
data, err := proto.Marshal(track)
if err != nil {
t.Fatalf("failed to marshal track: %v", err)
}

decoded, err := tc.decodeTrack(data)
if err != nil {
t.Fatalf("decodeTrack returned error: %v", err)
}
if decoded.GetTrackId() != "test-track-001" {
t.Errorf("track_id = %q, want %q", decoded.GetTrackId(), "test-track-001")
}
}

func TestTrackConsumer_DecodeTrack_EmptyPayload(t *testing.T) {
tc := &TrackConsumer{logger: slog.Default()}
_, err := tc.decodeTrack(nil)
if err == nil {
t.Error("Expected error for nil payload")
}
}

func TestTrackConsumer_DecodeTrack_EmptyTrackID(t *testing.T) {
tc := &TrackConsumer{logger: slog.Default()}

// Track with no ID.
track := &entityv1.FusedTrack{}
data, err := proto.Marshal(track)
if err != nil {
t.Fatalf("failed to marshal track: %v", err)
}

_, err = tc.decodeTrack(data)
if err == nil {
t.Error("Expected error for empty track_id")
}
}

func TestTrackConsumer_DecodeTrack_InvalidProtobuf(t *testing.T) {
tc := &TrackConsumer{logger: slog.Default()}
_, err := tc.decodeTrack([]byte{0x01, 0x02, 0x03, 0xFF, 0xFF})
if err == nil {
t.Error("Expected error for invalid protobuf data")
}
}

func TestTrackConsumer_Start_CallsHandler(t *testing.T) {
mockConsumer := &mockMessageConsumer{}
tc := NewTrackConsumer(mockConsumer, slog.Default())

var handledTrack *entityv1.FusedTrack
handler := func(_ context.Context, track *entityv1.FusedTrack) error {
handledTrack = track
return nil
}

track := &entityv1.FusedTrack{TrackId: "test-track-handler"}
data, _ := proto.Marshal(track)
mockConsumer.messageData = data

_ = tc.Start(context.Background(), []string{"tracks.fused.surface"}, handler)

if handledTrack == nil {
t.Error("Handler was not called")
} else if handledTrack.GetTrackId() != "test-track-handler" {
t.Errorf("Handler got track_id=%q, want 'test-track-handler'", handledTrack.GetTrackId())
}
}

func TestTrackConsumer_Close(t *testing.T) {
mock := &mockMessageConsumer{}
tc := NewTrackConsumer(mock, slog.Default())
err := tc.Close()
if err != nil {
t.Fatalf("Close returned error: %v", err)
}
if !mock.closed {
t.Error("Expected underlying consumer to be closed")
}
}

func TestTrackConsumer_Start_InvalidMessage(t *testing.T) {
mock := &mockMessageConsumer{}
mock.messageData = []byte{0xFF, 0xFE} // invalid protobuf
tc := NewTrackConsumer(mock, slog.Default())

called := false
handler := func(_ context.Context, _ *entityv1.FusedTrack) error {
called = true
return nil
}

_ = tc.Start(context.Background(), []string{"tracks.fused.surface"}, handler)
// Handler should NOT be called for invalid messages.
if called {
t.Error("Handler should not be called for invalid protobuf")
}
}
