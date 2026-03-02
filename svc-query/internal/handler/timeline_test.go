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

type mockTimelineRepository struct {
	events []*queryv1.TimelineEvent
	err    error
}

func (m *mockTimelineRepository) GetEventTimeline(ctx context.Context, req *queryv1.GetEventTimelineRequest, clearance commonv1.ClassificationLevel) ([]*queryv1.TimelineEvent, error) {
	return m.events, m.err
}

func TestGetEventTimeline_Success(t *testing.T) {
	mockRepo := &mockTimelineRepository{
		events: []*queryv1.TimelineEvent{
			{
				Summary:   "Test event",
				EventType: queryv1.TimelineEventType_TIMELINE_EVENT_TYPE_TRACK_CREATED,
				EventTime: timestamppb.New(time.Now()),
			},
		},
	}
	guardrail := domain.NewQueryGuardrail(30, 100000, 30)
	srv := handler.NewQueryServerForTest(guardrail, zap.NewNop(), nil, nil, nil, mockRepo)

	req := &queryv1.GetEventTimelineRequest{
		TrackId:   "track-123",
		TimeRange: validTimeRange(),
	}
	resp, err := srv.GetEventTimeline(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(resp.Events))
	}
	if resp.TrackId != "track-123" {
		t.Errorf("expected track-123, got %s", resp.TrackId)
	}
}

func TestGetEventTimeline_InvalidTimeRange(t *testing.T) {
	guardrail := domain.NewQueryGuardrail(30, 100000, 30)
	srv := handler.NewQueryServerForTest(guardrail, zap.NewNop(), nil, nil, nil, nil)

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
	srv := handler.NewQueryServerForTest(guardrail, zap.NewNop(), nil, nil, nil, nil)

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
	srv := handler.NewQueryServerForTest(guardrail, zap.NewNop(), nil, nil, nil, nil)

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
