// CLASSIFICATION: UNCLASSIFIED
package source

import (
	"context"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/entity/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/pkg/flatbuf"
	"google.golang.org/protobuf/proto"
)

func TestHub_BroadcastFanOut(t *testing.T) {
	h := newHub(4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a := h.subscribe(ctx)
	b := h.subscribe(ctx)

	delivered, dropped := h.broadcast([]byte{0x01})
	if delivered != 2 || dropped != 0 {
		t.Fatalf("broadcast delivered=%d dropped=%d, want 2/0", delivered, dropped)
	}
	if got := <-a; got[0] != 0x01 {
		t.Errorf("subscriber a got %v", got)
	}
	if got := <-b; got[0] != 0x01 {
		t.Errorf("subscriber b got %v", got)
	}
}

func TestHub_BroadcastDropsWhenFull(t *testing.T) {
	h := newHub(1) // channel buffers exactly one record
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = h.subscribe(ctx)

	if d, drop := h.broadcast([]byte{1}); d != 1 || drop != 0 {
		t.Fatalf("first broadcast d=%d drop=%d, want 1/0", d, drop)
	}
	// Buffer is now full; the next two must be dropped, never blocking.
	if d, drop := h.broadcast([]byte{2}); d != 0 || drop != 1 {
		t.Fatalf("second broadcast d=%d drop=%d, want 0/1", d, drop)
	}
	if d, drop := h.broadcast([]byte{3}); d != 0 || drop != 1 {
		t.Fatalf("third broadcast d=%d drop=%d, want 0/1", d, drop)
	}
}

func TestHub_SubscribeClosesOnCancel(t *testing.T) {
	h := newHub(2)
	ctx, cancel := context.WithCancel(context.Background())

	ch := h.subscribe(ctx)
	cancel()

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected channel to be closed after cancel")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("channel not closed within timeout after context cancel")
	}
	if got := h.subscriberCount(); got != 0 {
		t.Errorf("subscriberCount = %d, want 0 after cancel", got)
	}
}

func TestHub_CloseAll(t *testing.T) {
	h := newHub(2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := h.subscribe(ctx)
	h.closeAll()

	if _, ok := <-ch; ok {
		t.Error("expected channel closed after closeAll")
	}
	if got := h.subscriberCount(); got != 0 {
		t.Errorf("subscriberCount = %d, want 0", got)
	}
}

func newTestSource() *Source {
	return &Source{
		hub:        newHub(8),
		serializer: flatbuf.NewSerializer(),
	}
}

func TestDecodeAndSerialize_ValidTrack(t *testing.T) {
	update := &entityv1.TrackUpdate{
		Track: &entityv1.FusedTrack{
			TrackId:           "trk-1",
			EstimatedPosition: &commonv1.Position{Latitude: 45.42, Longitude: -75.70},
			Velocity:          &commonv1.Velocity{NorthMps: 10, EastMps: 5},
		},
	}
	raw, err := proto.Marshal(update)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	rec, ok := newTestSource().decodeAndSerialize(raw)
	if !ok {
		t.Fatal("decodeAndSerialize returned ok=false for a valid track")
	}
	if len(rec) != flatbuf.RecordSize {
		t.Errorf("record size = %d, want %d", len(rec), flatbuf.RecordSize)
	}
}

func TestDecodeAndSerialize_NoPosition(t *testing.T) {
	// A TrackUpdate with no Track has no position and must be skipped.
	raw, err := proto.Marshal(&entityv1.TrackUpdate{})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if _, ok := newTestSource().decodeAndSerialize(raw); ok {
		t.Error("expected ok=false for a position-less track")
	}
}

func TestDecodeAndSerialize_InvalidProtoBytes(t *testing.T) {
	// 0x08 is a field-1 varint tag with no following value — invalid wire format.
	if _, ok := newTestSource().decodeAndSerialize([]byte{0x08}); ok {
		t.Error("expected ok=false for malformed protobuf bytes")
	}
}
