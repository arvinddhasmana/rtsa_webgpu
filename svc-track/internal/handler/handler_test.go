// CLASSIFICATION: UNCLASSIFIED
// Package handler — unit tests for StreamTracks, GetTrackDetails, GetTrackHistory.
//
// Test coverage for T09–T13 per Module 10 specification.
package handler

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/domain"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func testMetrics() *metrics.Metrics {
	reg := prometheus.NewRegistry()
	return metrics.New(reg)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// makeTrack builds a FusedTrack for handler tests.
func makeTrack(id string, status commonv1.TrackStatus, et commonv1.EntityType, cls commonv1.ClassificationLevel, confidence float64) *entityv1.FusedTrack {
	return &entityv1.FusedTrack{
		TrackId:         id,
		Status:          status,
		EntityType:      et,
		Classification:  cls,
		ConfidenceScore: confidence,
		EstimatedPosition: &commonv1.Position{
			Latitude:  45.4215,
			Longitude: -75.6972,
		},
		UpdatedAt: timestamppb.New(time.Now()),
	}
}

// ─── GetTrackDetails tests ──────────────────────────────────────────────────

// T11: GetTrackDetails — existing track returns full track.
func TestDetailsHandler_GetTrackDetails_Exists(t *testing.T) {
	cache := domain.NewTrackCache(10)
	tr := makeTrack("detail-001", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.9)
	cache.Put(tr)

	h := NewDetailsHandler(cache)
	resp, err := h.GetTrackDetails(context.Background(), &entityv1.GetTrackDetailsRequest{
		TrackId:        "detail-001",
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.TrackId != "detail-001" {
		t.Errorf("expected track_id=detail-001, got %q", resp.TrackId)
	}
}

// T12: GetTrackDetails — non-existent track returns NOT_FOUND.
func TestDetailsHandler_GetTrackDetails_NotFound(t *testing.T) {
	cache := domain.NewTrackCache(10)
	h := NewDetailsHandler(cache)

	_, err := h.GetTrackDetails(context.Background(), &entityv1.GetTrackDetailsRequest{
		TrackId:        "nonexistent",
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
	})
	if err == nil {
		t.Fatal("expected NOT_FOUND error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Errorf("expected codes.NotFound, got %v", err)
	}
}

// T14 (details): SECRET track with PROTECTED_B clearance returns NOT_FOUND.
func TestDetailsHandler_GetTrackDetails_ClassificationBlocked(t *testing.T) {
	cache := domain.NewTrackCache(10)
	tr := makeTrack("secret-001", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET, 0.9)
	cache.Put(tr)

	h := NewDetailsHandler(cache)
	_, err := h.GetTrackDetails(context.Background(), &entityv1.GetTrackDetailsRequest{
		TrackId:        "secret-001",
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
	})
	if err == nil {
		t.Fatal("expected NOT_FOUND for SECRET track with PROTECTED_B clearance")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Errorf("expected codes.NotFound, got %v", err)
	}
}

// TestDetailsHandler_NilRequest: nil request returns InvalidArgument.
func TestDetailsHandler_NilRequest(t *testing.T) {
	h := NewDetailsHandler(domain.NewTrackCache(10))
	_, err := h.GetTrackDetails(context.Background(), nil)
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument for nil request, got %v", err)
	}
}

// TestDetailsHandler_EmptyTrackID: empty track_id returns InvalidArgument.
func TestDetailsHandler_EmptyTrackID(t *testing.T) {
	h := NewDetailsHandler(domain.NewTrackCache(10))
	_, err := h.GetTrackDetails(context.Background(), &entityv1.GetTrackDetailsRequest{
		TrackId: "",
	})
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument for empty track_id, got %v", err)
	}
}

// ─── GetTrackHistory tests ──────────────────────────────────────────────────

// T13: GetTrackHistory — returns history points for a known track.
func TestHistoryHandler_GetTrackHistory_ReturnsPoints(t *testing.T) {
	cache := domain.NewTrackCache(50)
	// Insert 5 updates for the same track.
	for i := 0; i < 5; i++ {
		cache.Put(makeTrack("hist-001", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, float64(i)*0.1+0.5))
	}

	h := NewHistoryHandler(cache)
	resp, err := h.GetTrackHistory(context.Background(), &entityv1.GetTrackHistoryRequest{
		TrackId:        "hist-001",
		MaxPoints:      10,
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TrackId != "hist-001" {
		t.Errorf("expected track_id=hist-001, got %q", resp.TrackId)
	}
	if len(resp.Points) != 5 {
		t.Errorf("expected 5 history points, got %d", len(resp.Points))
	}
	for _, pt := range resp.Points {
		if pt.Position == nil {
			t.Error("history point has nil position")
		}
	}
}

// TestHistoryHandler_NotFound: non-existent track returns NOT_FOUND.
func TestHistoryHandler_NotFound(t *testing.T) {
	h := NewHistoryHandler(domain.NewTrackCache(10))
	_, err := h.GetTrackHistory(context.Background(), &entityv1.GetTrackHistoryRequest{
		TrackId:        "ghost",
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
	})
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Errorf("expected codes.NotFound, got %v", err)
	}
}

// TestHistoryHandler_ClassificationBlocked: SECRET track with PROTECTED_B clearance.
func TestHistoryHandler_ClassificationBlocked(t *testing.T) {
	cache := domain.NewTrackCache(10)
	cache.Put(makeTrack("hist-secret", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET, 0.9))

	h := NewHistoryHandler(cache)
	_, err := h.GetTrackHistory(context.Background(), &entityv1.GetTrackHistoryRequest{
		TrackId:        "hist-secret",
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
	})
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Errorf("expected NOT_FOUND for classification violation, got %v", err)
	}
}

// TestHistoryHandler_EmptyTrackID: returns InvalidArgument.
func TestHistoryHandler_EmptyTrackID(t *testing.T) {
	h := NewHistoryHandler(domain.NewTrackCache(10))
	_, err := h.GetTrackHistory(context.Background(), &entityv1.GetTrackHistoryRequest{TrackId: ""})
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

// ─── StreamHandler unit tests ───────────────────────────────────────────────

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

// ─── StreamTracks end-to-end with mock stream ────────────────────────────────

// mockServerStream simulates a gRPC server stream for testing StreamTracks.
type mockServerStream struct {
	sent   []*entityv1.TrackUpdate
	ctx    context.Context
	sendFn func(*entityv1.TrackUpdate) error
}

func (m *mockServerStream) Send(u *entityv1.TrackUpdate) error {
	if m.sendFn != nil {
		return m.sendFn(u)
	}
	m.sent = append(m.sent, u)
	return nil
}

func (m *mockServerStream) Context() context.Context     { return m.ctx }
func (m *mockServerStream) SetHeader(metadata.MD) error  { return nil }
func (m *mockServerStream) SendHeader(metadata.MD) error { return nil }
func (m *mockServerStream) SetTrailer(metadata.MD)       {}
func (m *mockServerStream) RecvMsg(_ interface{}) error  { return nil }
func (m *mockServerStream) SendMsg(v interface{}) error {
	if u, ok := v.(*entityv1.TrackUpdate); ok {
		return m.Send(u)
	}
	return nil
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
