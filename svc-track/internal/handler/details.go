// CLASSIFICATION: UNCLASSIFIED
// Package handler implements the TrackService.GetTrackDetails gRPC handler.
//
// Feature: FEAT-13 Situational Awareness UI
// UC: UC012 Situational Awareness UI
// Requirements: CR-UI-001, CR-SEC-001
package handler

import (
	"context"

	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/domain"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/mapper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DetailsHandler implements the TrackService.GetTrackDetails unary RPC.
// Classification filtering is mandatory: tracks with classification > caller's
// clearance are returned as NOT_FOUND to avoid leaking existence.
type DetailsHandler struct {
	cache *domain.TrackCache
}

// NewDetailsHandler creates a DetailsHandler backed by the given cache.
func NewDetailsHandler(cache *domain.TrackCache) *DetailsHandler {
	return &DetailsHandler{cache: cache}
}

// GetTrackDetails returns the full FusedTrack for the requested track_id.
// Returns NOT_FOUND if the track does not exist or the caller lacks clearance.
func (h *DetailsHandler) GetTrackDetails(ctx context.Context, req *entityv1.GetTrackDetailsRequest) (*entityv1.FusedTrack, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request must not be nil")
	}
	if req.TrackId == "" {
		return nil, status.Error(codes.InvalidArgument, "track_id must not be empty")
	}

	track := h.cache.Get(req.TrackId)
	if track == nil {
		return nil, status.Errorf(codes.NotFound, "track %q not found", req.TrackId)
	}

	// MANDATORY classification check — return NOT_FOUND to avoid leaking existence.
	if !mapper.TrackPassesClassification(track, req.ClearanceLevel) {
		return nil, status.Errorf(codes.NotFound, "track %q not found", req.TrackId)
	}

	return track, nil
}
