// CLASSIFICATION: UNCLASSIFIED
// Package handler implements the TrackService.StreamTracks gRPC handler.
//
// StreamTracks sends an initial snapshot of all matching tracks then streams
// incremental updates until the client disconnects or context is cancelled.
// Each connected client gets its own buffered channel (fan-out pattern).
// Classification filtering is applied on every message before it is sent.
//
// Feature: FEAT-13 Situational Awareness UI
// UC: UC012 Situational Awareness UI
// Requirements: CR-UI-001, CR-UI-002, CR-SEC-001
package handler

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/domain"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/mapper"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/metrics"
	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StreamHandler implements the StreamTracks gRPC server-streaming handler.
type StreamHandler struct {
	entityv1.UnimplementedTrackServiceServer
	cache      *domain.TrackCache
	filter     *domain.FilterEngine
	metrics    *metrics.Metrics
	logger     *slog.Logger
	chanBuffer int

	// clients maps clientID (string) → chan *entityv1.TrackUpdate
	clients sync.Map

	// clientCounter generates unique client IDs.
	clientCounter atomic.Uint64
}

// NewStreamHandler creates a StreamHandler and wires the cache onChange callback.
func NewStreamHandler(
	cache *domain.TrackCache,
	filter *domain.FilterEngine,
	m *metrics.Metrics,
	logger *slog.Logger,
	chanBuffer int,
) *StreamHandler {
	h := &StreamHandler{
		cache:      cache,
		filter:     filter,
		metrics:    m,
		logger:     logger,
		chanBuffer: chanBuffer,
	}
	cache.SetOnChange(h.Broadcast)
	return h
}

// Broadcast fans out a cache update to all registered client channels.
// Called by the TrackCache with no locks held; must not block.
func (h *StreamHandler) Broadcast(update *entityv1.TrackUpdate) {
	h.clients.Range(func(key, value any) bool {
		ch, ok := value.(chan *entityv1.TrackUpdate)
		if !ok {
			return true
		}
		select {
		case ch <- update:
		default:
			// Channel full — drop this update for this client.
			// The client will receive the next update; snapshot handles initial state.
			h.logger.Warn("client update channel full, dropping update",
				slog.Any("client_id", key),
				slog.String("track_id", update.Track.GetTrackId()),
			)
		}
		return true
	})
}

// StreamTracks implements TrackServiceServer.StreamTracks.
//
// Flow:
//  1. Parse filter from request.
//  2. Register per-client channel BEFORE snapshot to prevent missed updates.
//  3. Send initial snapshot of all matching active tracks.
//  4. Forward incremental updates from the channel until context cancel.
func (h *StreamHandler) StreamTracks(
	req *entityv1.StreamTracksRequest,
	stream grpc.ServerStreamingServer[entityv1.TrackUpdate],
) error {
	ctx := stream.Context()

	// Build the filter for this client.
	trackFilter := mapper.ToTrackFilter(req)

	// Assign a unique client ID.
	clientID := h.clientCounter.Add(1)

	// Create a buffered channel for this client's updates.
	updateCh := make(chan *entityv1.TrackUpdate, h.chanBuffer)

	// Register BEFORE snapshot to prevent race between snapshot and new updates.
	h.clients.Store(clientID, updateCh)
	h.metrics.StreamClients.Inc()

	defer func() {
		h.clients.Delete(clientID)
		close(updateCh)
		h.metrics.StreamClients.Dec()
		h.logger.InfoContext(ctx, "stream client disconnected", slog.Uint64("client_id", clientID))
	}()

	h.logger.InfoContext(ctx, "stream client connected", slog.Uint64("client_id", clientID))

	// Phase 1: Send initial snapshot.
	if err := h.cache.Snapshot(trackFilter, func(update *entityv1.TrackUpdate) error {
		if err := stream.Send(update); err != nil {
			return fmt.Errorf("handler.StreamTracks.Snapshot(client=%d): send: %w", clientID, err)
		}
		h.recordSent(update)
		return nil
	}); err != nil {
		return status.Errorf(codes.Internal, "snapshot failed: %v", err)
	}

	// Phase 2: Stream incremental updates.
	for {
		select {
		case <-ctx.Done():
			return nil
		case update, ok := <-updateCh:
			if !ok {
				return nil
			}
			// Apply filter to each incremental update.
			if !h.filter.Matches(update.Track, trackFilter) {
				continue
			}
			if err := stream.Send(update); err != nil {
				// Client disconnected.
				return nil
			}
			h.recordSent(update)
		}
	}
}

// recordSent increments the updates_sent_total metric.
func (h *StreamHandler) recordSent(update *entityv1.TrackUpdate) {
	if h.metrics == nil || update.Track == nil {
		return
	}
	h.metrics.UpdatesSentTotal.WithLabelValues(
		update.Track.EntityType.String(),
		update.UpdateType.String(),
	).Inc()
}
