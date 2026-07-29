// CLASSIFICATION: UNCLASSIFIED
package handler

import (
	"context"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/entity/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-track/internal/domain"
)

// T09: StreamTracks initial snapshot.
func TestStreamHandler_Broadcast_FanOut(t *testing.T) {
	cache := domain.NewTrackCache(10)
	filter := &domain.FilterEngine{}
	m := testMetrics()
	logger := testLogger()

	h := NewStreamHandler(cache, filter, m, logger, 64)

	// Register a test client channel manually.
	ch := make(chan *entityv1.TrackUpdate, 16)
	h.clients.Store(uint64(999), ch)

	// Broadcast an update.
	update := &entityv1.TrackUpdate{
		UpdateType: entityv1.TrackUpdate_UPDATE_TYPE_CREATED,
		Track:      makeTrack("broadcast-01", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.8),
	}
	h.Broadcast(update)

	select {
	case got := <-ch:
		if got.Track.TrackId != "broadcast-01" {
			t.Errorf("expected track_id=broadcast-01, got %q", got.Track.TrackId)
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for broadcast")
	}
}

// T10: StreamTracks broadcast — full channel drops without blocking.
func TestStreamHandler_Broadcast_FullChannel_NoPanic(t *testing.T) {
	cache := domain.NewTrackCache(10)
	filter := &domain.FilterEngine{}
	m := testMetrics()
	h := NewStreamHandler(cache, filter, m, testLogger(), 1)

	// Register a client with a channel of size 1 and pre-fill it.
	ch := make(chan *entityv1.TrackUpdate, 1)
	ch <- &entityv1.TrackUpdate{} // fill it
	h.clients.Store(uint64(1), ch)

	// This must not panic or block.
	update := &entityv1.TrackUpdate{
		UpdateType: entityv1.TrackUpdate_UPDATE_TYPE_UPDATED,
		Track:      makeTrack("overflow-01", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.7),
	}
	done := make(chan struct{})
	go func() {
		h.Broadcast(update)
		close(done)
	}()
	select {
	case <-done:
		// passed
	case <-time.After(time.Second):
		t.Error("Broadcast blocked on full channel — must be non-blocking")
	}
}

// T09: StreamTracks sends initial snapshot and then stops when ctx is cancelled.
func TestStreamHandler_StreamTracks_Snapshot(t *testing.T) {
	cache := domain.NewTrackCache(10)
	// Pre-populate cache with active tracks.
	cache.Put(makeTrack("snap-01", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.9))
	cache.Put(makeTrack("snap-02", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_SURFACE, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.8))

	filter := &domain.FilterEngine{}
	m := testMetrics()
	h := NewStreamHandler(cache, filter, m, testLogger(), 64)

	ctx, cancel := context.WithCancel(context.Background())

	stream := &mockServerStream{ctx: ctx}

	// Cancel context immediately after receiving snapshot to end the stream.
	stream.sendFn = func(u *entityv1.TrackUpdate) error {
		stream.sent = append(stream.sent, u)
		// After receiving snapshot items, cancel.
		if len(stream.sent) >= 2 {
			cancel()
		}
		return nil
	}

	req := &entityv1.StreamTracksRequest{
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
	}

	// StreamTracks should return when context is cancelled.
	err := h.StreamTracks(req, stream)
	if err != nil {
		t.Errorf("StreamTracks returned error: %v", err)
	}

	// We should have received snapshot messages.
	if len(stream.sent) < 2 {
		t.Errorf("expected at least 2 snapshot messages, got %d", len(stream.sent))
	}
	for _, u := range stream.sent {
		if u.UpdateType != entityv1.TrackUpdate_UPDATE_TYPE_SNAPSHOT {
			t.Errorf("expected SNAPSHOT update type, got %v", u.UpdateType)
		}
	}
}

// T10: StreamTracks receives incremental updates from cache.
func TestStreamHandler_StreamTracks_IncrementalUpdate(t *testing.T) {
	cache := domain.NewTrackCache(10)
	filter := &domain.FilterEngine{}
	m := testMetrics()
	h := NewStreamHandler(cache, filter, m, testLogger(), 64)

	ctx, cancel := context.WithCancel(context.Background())

	received := make(chan *entityv1.TrackUpdate, 8)
	stream := &mockServerStream{
		ctx: ctx,
		sendFn: func(u *entityv1.TrackUpdate) error {
			received <- u
			return nil
		},
	}

	req := &entityv1.StreamTracksRequest{
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
	}

	// Start streaming in goroutine.
	streamErr := make(chan error, 1)
	go func() {
		streamErr <- h.StreamTracks(req, stream)
	}()

	// Give the stream a moment to start and register.
	time.Sleep(50 * time.Millisecond)

	// Now put a new track — should trigger onChange → client channel → stream.Send.
	newTrack := makeTrack("incr-01", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.9)
	cache.Put(newTrack)

	// Wait for the update or timeout.
	select {
	case u := <-received:
		if u.Track.TrackId != "incr-01" {
			t.Errorf("expected incr-01, got %q", u.Track.TrackId)
		}
		if u.UpdateType != entityv1.TrackUpdate_UPDATE_TYPE_CREATED {
			t.Errorf("expected CREATED, got %v", u.UpdateType)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout: did not receive incremental update")
	}

	cancel()
	<-streamErr
}
