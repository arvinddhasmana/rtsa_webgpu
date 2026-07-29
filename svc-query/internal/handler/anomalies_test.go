// CLASSIFICATION: UNCLASSIFIED
package handler_test

import (
	"context"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	inferencev1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/inference/v1"
	queryv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/query/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-query/internal/domain"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-query/internal/handler"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockAnomalyRepository struct {
	alerts []*inferencev1.AnomalyAlert
	err    error
}

func (m *mockAnomalyRepository) QueryAnomalies(ctx context.Context, req *queryv1.QueryAnomaliesRequest, clearance commonv1.ClassificationLevel, pageToken *domain.PaginationToken, pageSize int) ([]*inferencev1.AnomalyAlert, *domain.PaginationToken, error) {
	return m.alerts, nil, m.err
}

func TestQueryAnomalies_Success(t *testing.T) {
	mockRepo := &mockAnomalyRepository{
		alerts: []*inferencev1.AnomalyAlert{
			{AlertId: "alert-1"},
		},
	}
	guardrail := domain.NewQueryGuardrail(30, 100000, 30)
	srv := handler.NewQueryServerForTest(guardrail, zap.NewNop(), nil, mockRepo, nil, nil)

	req := &queryv1.QueryAnomaliesRequest{
		TimeRange: validTimeRange(),
	}
	resp, err := srv.QueryAnomalies(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Alerts) != 1 {
		t.Errorf("expected 1 alert, got %d", len(resp.Alerts))
	}
}

func TestQueryAnomalies_InvalidTimeRange(t *testing.T) {
	guardrail := domain.NewQueryGuardrail(30, 100000, 30)
	srv := handler.NewQueryServerForTest(guardrail, zap.NewNop(), nil, nil, nil, nil)

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
