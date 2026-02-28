// CLASSIFICATION: UNCLASSIFIED
package handler

import (
	"context"

	queryv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/query/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/security"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetEventTimeline handles the federated event timeline query returning an ordered stream of updates.
func (s *QueryServer) GetEventTimeline(
	ctx context.Context,
	req *queryv1.GetEventTimelineRequest,
) (*queryv1.EventTimelineResponse, error) {
	if req.TrackId == "" {
		return nil, status.Error(codes.InvalidArgument, "track_id is required")
	}

	if req.TimeRange == nil {
		return nil, status.Error(codes.InvalidArgument, "time_range is required")
	}

	if err := s.guardrail.ValidateTimeRange(
		req.TimeRange.StartTime,
		req.TimeRange.EndTime,
	); err != nil {
		return nil, err
	}

	clearance := security.ExtractClearance(ctx)

	if req.MaxEvents > int32(s.pageSize) || req.MaxEvents == 0 {
		req.MaxEvents = int32(s.pageSize)
	}

	qctx, cancel := s.guardrail.QueryContext(ctx)
	defer cancel()

	events, err := s.timelineRepo.GetEventTimeline(qctx, req, clearance)
	if err != nil {
		s.logger.Error("GetEventTimeline: repository error", zap.Error(err))
		return nil, status.Error(codes.Internal, "query failed")
	}

	s.auditEmitter.Emit(ctx, "query.executed", "timeline", req.TrackId)

	return &queryv1.EventTimelineResponse{
		TrackId: req.TrackId,
		Events:  events,
	}, nil
}
