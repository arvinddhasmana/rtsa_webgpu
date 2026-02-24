// CLASSIFICATION: UNCLASSIFIED
// Package handler implements the TrackService.GetTrackHistory gRPC handler.
//
// Feature: FEAT-13 Situational Awareness UI
// UC: UC012, UC013 Historical Query
// Requirements: CR-UI-001, CR-HIS-001, CR-SEC-001
package handler

import (
	"context"

	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/domain"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/mapper"
	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HistoryHandler implements the TrackService.GetTrackHistory unary RPC.
type HistoryHandler struct {
	cache *domain.TrackCache
}

// NewHistoryHandler creates a HistoryHandler backed by the given cache.
func NewHistoryHandler(cache *domain.TrackCache) *HistoryHandler {
	return &HistoryHandler{cache: cache}
}

// GetTrackHistory returns the recent position history for a track.
// Returns NOT_FOUND if the track does not exist or the caller lacks clearance.
func (h *HistoryHandler) GetTrackHistory(ctx context.Context, req *entityv1.GetTrackHistoryRequest) (*entityv1.TrackHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request must not be nil")
	}
	if req.TrackId == "" {
		return nil, status.Error(codes.InvalidArgument, "track_id must not be empty")
	}

	maxPoints := int(req.MaxPoints)
	if maxPoints <= 0 {
		maxPoints = domain.DefaultHistoryMaxPoints
	}

	// Verify the track exists and classify access.
	track := h.cache.Get(req.TrackId)
	if track == nil {
		return nil, status.Errorf(codes.NotFound, "track %q not found", req.TrackId)
	}

	// MANDATORY classification check — return NOT_FOUND to avoid leaking existence.
	if !mapper.TrackPassesClassification(track, req.ClearanceLevel) {
		return nil, status.Errorf(codes.NotFound, "track %q not found", req.TrackId)
	}

	history := h.cache.GetHistory(req.TrackId, maxPoints)

	return &entityv1.TrackHistoryResponse{
		TrackId: req.TrackId,
		Points:  mapper.ToProtoHistoryPoints(history),
	}, nil
}
