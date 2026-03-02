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
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockAuditRepository struct {
	entries []*queryv1.AuditLogEntry
	err     error
}

func (m *mockAuditRepository) QueryAuditLog(ctx context.Context, req *queryv1.QueryAuditLogRequest, clearance commonv1.ClassificationLevel, pageToken *domain.PaginationToken, pageSize int) ([]*queryv1.AuditLogEntry, *domain.PaginationToken, error) {
	return m.entries, nil, m.err
}

func TestQueryAuditLog_Success(t *testing.T) {
	mockRepo := &mockAuditRepository{
		entries: []*queryv1.AuditLogEntry{
			{AuditId: "audit-1"},
		},
	}
	guardrail := domain.NewQueryGuardrail(30, 100000, 30)
	srv := handler.NewQueryServerForTest(guardrail, zap.NewNop(), nil, nil, mockRepo, nil)

	req := &queryv1.QueryAuditLogRequest{
		TimeRange: validTimeRange(),
	}
	resp, err := srv.QueryAuditLog(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(resp.Entries))
	}
}

func TestQueryAuditLog_InvalidTimeRange(t *testing.T) {
	guardrail := domain.NewQueryGuardrail(30, 100000, 30)
	srv := handler.NewQueryServerForTest(guardrail, zap.NewNop(), nil, nil, nil, nil)

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
