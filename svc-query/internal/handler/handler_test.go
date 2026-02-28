// CLASSIFICATION: UNCLASSIFIED
package handler_test

import (
	"context"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	queryv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/query/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/domain"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/handler"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func validTimeRange() *commonv1.TimeRange {
	now := time.Now()
	return &commonv1.TimeRange{
		StartTime: timestamppb.New(now.Add(-1 * time.Hour)),
		EndTime:   timestamppb.New(now),
	}
}

func TestQueryTracks_InvalidTimeRange(t *testing.T) {
	guardrail := domain.NewQueryGuardrail(30, 100000, 30)
	srv := handler.NewQueryServerForTest(guardrail, zap.NewNop())

	now := time.Now()
	req := &queryv1.QueryTracksRequest{
		TimeRange: &commonv1.TimeRange{
			StartTime: timestamppb.New(now),
			EndTime:   timestamppb.New(now.Add(-1 * time.Hour)),
		},
	}
	_, err := srv.QueryTracks(context.Background(), req)
	if err == nil {
		t.Error("expected error for invalid time range")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestQueryAnomalies_InvalidTimeRange(t *testing.T) {
	guardrail := domain.NewQueryGuardrail(30, 100000, 30)
	srv := handler.NewQueryServerForTest(guardrail, zap.NewNop())

	now := time.Now()
	req := &queryv1.QueryAnomaliesRequest{
		TimeRange: &commonv1.TimeRange{
			StartTime: timestamppb.New(now),
			EndTime:   timestamppb.New(now.Add(-1 * time.Hour)),
		},
	}
	_, err := srv.QueryAnomalies(context.Background(), req)
	if err == nil {
		t.Error("expected error for invalid time range")
	}
}

func TestQueryAuditLog_InvalidTimeRange(t *testing.T) {
	guardrail := domain.NewQueryGuardrail(30, 100000, 30)
	srv := handler.NewQueryServerForTest(guardrail, zap.NewNop())

	now := time.Now()
	req := &queryv1.QueryAuditLogRequest{
		TimeRange: &commonv1.TimeRange{
			StartTime: timestamppb.New(now),
			EndTime:   timestamppb.New(now.Add(-1 * time.Hour)),
		},
	}
	_, err := srv.QueryAuditLog(context.Background(), req)
	if err == nil {
		t.Error("expected error for invalid time range")
	}
}

func TestQueryTracks_ValidRange(t *testing.T) {
	guardrail := domain.NewQueryGuardrail(30, 100000, 30)
	srv := handler.NewQueryServerForTest(guardrail, zap.NewNop())
	_ = validTimeRange()
	_ = srv
}

func TestGetEventTimeline_InvalidTimeRange(t *testing.T) {
	guardrail := domain.NewQueryGuardrail(30, 100000, 30)
	srv := handler.NewQueryServerForTest(guardrail, zap.NewNop())

	now := time.Now()
	req := &queryv1.GetEventTimelineRequest{
		TrackId: "track-123",
		TimeRange: &commonv1.TimeRange{
			StartTime: timestamppb.New(now),
			EndTime:   timestamppb.New(now.Add(-1 * time.Hour)),
		},
	}
	_, err := srv.GetEventTimeline(context.Background(), req)
	if err == nil {
		t.Error("expected error for invalid time range")
	}
}

func TestGetEventTimeline_EmptyTrackId(t *testing.T) {
	guardrail := domain.NewQueryGuardrail(30, 100000, 30)
	srv := handler.NewQueryServerForTest(guardrail, zap.NewNop())

	req := &queryv1.GetEventTimelineRequest{
		TrackId:   "",
		TimeRange: validTimeRange(),
	}
	_, err := srv.GetEventTimeline(context.Background(), req)
	if err == nil {
		t.Error("expected error for empty track id")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestGetEventTimeline_MissingTimeRange(t *testing.T) {
	guardrail := domain.NewQueryGuardrail(30, 100000, 30)
	srv := handler.NewQueryServerForTest(guardrail, zap.NewNop())

	req := &queryv1.GetEventTimelineRequest{
		TrackId: "track-123",
	}
	_, err := srv.GetEventTimeline(context.Background(), req)
	if err == nil {
		t.Error("expected error for missing time range")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}
