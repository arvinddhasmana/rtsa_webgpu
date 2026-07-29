// CLASSIFICATION: UNCLASSIFIED
package handler_test

import (
	"context"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/entity/v1"
	queryv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/query/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-query/internal/domain"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-query/internal/handler"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockTracksRepository struct {
	tracks []*entityv1.FusedTrack
	err    error
}

func (m *mockTracksRepository) QueryTracks(ctx context.Context, req *queryv1.QueryTracksRequest, clearance commonv1.ClassificationLevel, pageToken *domain.PaginationToken, pageSize int) ([]*entityv1.FusedTrack, *domain.PaginationToken, error) {
	return m.tracks, nil, m.err
}

func TestQueryTracks_Success(t *testing.T) {
	mockRepo := &mockTracksRepository{
		tracks: []*entityv1.FusedTrack{
			{TrackId: "track-1"},
		},
	}
	guardrail := domain.NewQueryGuardrail(30, 100000, 30)
	srv := handler.NewQueryServerForTest(guardrail, zap.NewNop(), mockRepo, nil, nil, nil)

	req := &queryv1.QueryTracksRequest{
		TimeRange: validTimeRange(),
	}
	resp, err := srv.QueryTracks(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Tracks) != 1 {
		t.Errorf("expected 1 track, got %d", len(resp.Tracks))
	}
}

func TestQueryTracks_InvalidTimeRange(t *testing.T) {
	guardrail := domain.NewQueryGuardrail(30, 100000, 30)
	srv := handler.NewQueryServerForTest(guardrail, zap.NewNop(), nil, nil, nil, nil)

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

func TestQueryTracks_ValidRange(t *testing.T) {
	guardrail := domain.NewQueryGuardrail(30, 100000, 30)
	srv := handler.NewQueryServerForTest(guardrail, zap.NewNop(), nil, nil, nil, nil)
	_ = validTimeRange()
	_ = srv
}
