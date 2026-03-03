// CLASSIFICATION: UNCLASSIFIED
//go:build integration

// Package integration_test validates the QueryServer handler pipeline for
// svc-query. Tests interact with two components: the guardrail (time-range/row-limit
// validation) and the query handler routing (classification filtering, pagination).
// Repositories are stubbed via mock implementations — no ClickHouse container required.
package integration_test

import (
	"context"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
	queryv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/query/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/domain"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/handler"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ─── Mock Repositories ───────────────────────────────────────────────────────

// mockTracksRepo returns a fixed list of FusedTrack items.
type mockTracksRepo struct {
	tracks []*entityv1.FusedTrack
}

func (m *mockTracksRepo) QueryTracks(
	_ context.Context,
	_ *queryv1.QueryTracksRequest,
	_ commonv1.ClassificationLevel,
	_ *domain.PaginationToken,
	_ int,
) ([]*entityv1.FusedTrack, *domain.PaginationToken, error) {
	return m.tracks, nil, nil
}

// mockAnomalyRepo returns a fixed list of AnomalyAlert items.
type mockAnomalyRepo struct {
	alerts []*inferencev1.AnomalyAlert
}

func (m *mockAnomalyRepo) QueryAnomalies(
	_ context.Context,
	_ *queryv1.QueryAnomaliesRequest,
	_ commonv1.ClassificationLevel,
	_ *domain.PaginationToken,
	_ int,
) ([]*inferencev1.AnomalyAlert, *domain.PaginationToken, error) {
	return m.alerts, nil, nil
}

// mockAuditRepo returns a fixed list of AuditLogEntry items.
type mockAuditRepo struct {
	entries []*queryv1.AuditLogEntry
}

func (m *mockAuditRepo) QueryAuditLog(
	_ context.Context,
	_ *queryv1.QueryAuditLogRequest,
	_ commonv1.ClassificationLevel,
	_ *domain.PaginationToken,
	_ int,
) ([]*queryv1.AuditLogEntry, *domain.PaginationToken, error) {
	return m.entries, nil, nil
}

// mockTimelineRepo returns an empty timeline.
type mockTimelineRepo struct{}

func (m *mockTimelineRepo) GetEventTimeline(
	_ context.Context,
	_ *queryv1.GetEventTimelineRequest,
	_ commonv1.ClassificationLevel,
) ([]*queryv1.TimelineEvent, error) {
	return nil, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func validTimeRange() *commonv1.TimeRange {
	now := time.Now().UTC()
	return &commonv1.TimeRange{
		StartTime: timestamppb.New(now.Add(-24 * time.Hour)),
		EndTime:   timestamppb.New(now),
	}
}

func newTestQueryServer(
	t *testing.T,
	tracks []*entityv1.FusedTrack,
	alerts []*inferencev1.AnomalyAlert,
) *handler.QueryServer {
	t.Helper()
	guardrail := domain.NewQueryGuardrail(90, 1000, 30)
	return handler.NewQueryServerForTest(
		guardrail,
		zap.NewNop(),
		&mockTracksRepo{tracks: tracks},
		&mockAnomalyRepo{alerts: alerts},
		&mockAuditRepo{},
		&mockTimelineRepo{},
	)
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// TestQueryTracks_ValidTimeRange_ReturnsExpectedTracks validates that QueryTracks
// returns the expected track set from the repository when given a valid time range.
func TestQueryTracks_ValidTimeRange_ReturnsExpectedTracks(t *testing.T) {
	tracks := []*entityv1.FusedTrack{
		{
			TrackId:        "track-q-001",
			EntityType:     commonv1.EntityType_ENTITY_TYPE_SURFACE,
			Status:         commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
			Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
			EstimatedPosition: &commonv1.Position{Latitude: 44.65, Longitude: -63.57},
		},
		{
			TrackId:        "track-q-002",
			EntityType:     commonv1.EntityType_ENTITY_TYPE_AIR,
			Status:         commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
			Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
			EstimatedPosition: &commonv1.Position{Latitude: 45.0, Longitude: -60.0},
		},
	}

	srv := newTestQueryServer(t, tracks, nil)

	resp, err := srv.QueryTracks(context.Background(), &queryv1.QueryTracksRequest{
		TimeRange: validTimeRange(),
	})
	if err != nil {
		t.Fatalf("TestQueryTracks_ValidTimeRange_ReturnsExpectedTracks: unexpected error: %v", err)
	}
	if len(resp.GetTracks()) != 2 {
		t.Errorf("TestQueryTracks_ValidTimeRange_ReturnsExpectedTracks: got %d tracks, want 2", len(resp.GetTracks()))
	}
	if resp.GetPagination().GetTotalCount() != 2 {
		t.Errorf("TestQueryTracks_ValidTimeRange_ReturnsExpectedTracks: total_count=%d, want 2", resp.GetPagination().GetTotalCount())
	}
}

// TestQueryTracks_ExceedsMaxRangeDays_ReturnsInvalidArgument validates that the
// guardrail rejects queries spanning more than MaxRangeDays.
func TestQueryTracks_ExceedsMaxRangeDays_ReturnsInvalidArgument(t *testing.T) {
	srv := newTestQueryServer(t, nil, nil)

	// Request a time-range of 180 days (exceeds guardrail limit of 90).
	now := time.Now().UTC()
	resp, err := srv.QueryTracks(context.Background(), &queryv1.QueryTracksRequest{
		TimeRange: &commonv1.TimeRange{
			StartTime: timestamppb.New(now.Add(-180 * 24 * time.Hour)),
			EndTime:   timestamppb.New(now),
		},
	})
	if err == nil {
		t.Fatalf("TestQueryTracks_ExceedsMaxRangeDays_ReturnsInvalidArgument: expected error, got resp with %d tracks",
			len(resp.GetTracks()))
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("TestQueryTracks_ExceedsMaxRangeDays_ReturnsInvalidArgument: error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("TestQueryTracks_ExceedsMaxRangeDays_ReturnsInvalidArgument: code=%v, want InvalidArgument", st.Code())
	}
}

// TestQueryAnomalies_ValidRequest_ReturnsAlerts validates that QueryAnomalies
// returns the expected alert set.
func TestQueryAnomalies_ValidRequest_ReturnsAlerts(t *testing.T) {
	alerts := []*inferencev1.AnomalyAlert{
		{
			AlertId:         "alert-q-001",
			TrackId:         "track-q-001",
			AnomalyType:     commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
			Severity:        commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL,
			ConfidenceScore: 0.95,
			Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
		},
	}

	srv := newTestQueryServer(t, nil, alerts)

	resp, err := srv.QueryAnomalies(context.Background(), &queryv1.QueryAnomaliesRequest{
		TimeRange: validTimeRange(),
	})
	if err != nil {
		t.Fatalf("TestQueryAnomalies_ValidRequest_ReturnsAlerts: unexpected error: %v", err)
	}
	if len(resp.GetAlerts()) != 1 {
		t.Errorf("TestQueryAnomalies_ValidRequest_ReturnsAlerts: got %d alerts, want 1", len(resp.GetAlerts()))
	}
	got := resp.GetAlerts()[0]
	if got.GetAnomalyType() != commonv1.AnomalyType_ANOMALY_TYPE_SPEED {
		t.Errorf("TestQueryAnomalies_ValidRequest_ReturnsAlerts: anomaly_type=%v, want SPEED", got.GetAnomalyType())
	}
	if got.GetConfidenceScore() < 0.5 {
		t.Errorf("TestQueryAnomalies_ValidRequest_ReturnsAlerts: confidence=%.2f, want >= 0.5", got.GetConfidenceScore())
	}
}

// TestQueryAuditLog_ValidRequest_ReturnsEntries validates that QueryAuditLog
// correctly routes through the audit repository and returns entries.
func TestQueryAuditLog_ValidRequest_ReturnsEntries(t *testing.T) {
	entries := []*queryv1.AuditLogEntry{
		{
			AuditId:   "audit-q-001",
			ServiceId: "svc-fusion-engine",
			EventType: "track.created",
			Action:    "CREATE",
		},
		{
			AuditId:   "audit-q-002",
			ServiceId: "svc-feedback",
			EventType: "feedback.submitted",
			Action:    "CREATE",
		},
	}

	guardrail := domain.NewQueryGuardrail(90, 1000, 30)
	srv := handler.NewQueryServerForTest(
		guardrail,
		zap.NewNop(),
		&mockTracksRepo{},
		&mockAnomalyRepo{},
		&mockAuditRepo{entries: entries},
		&mockTimelineRepo{},
	)

	resp, err := srv.QueryAuditLog(context.Background(), &queryv1.QueryAuditLogRequest{
		TimeRange: validTimeRange(),
	})
	if err != nil {
		t.Fatalf("TestQueryAuditLog_ValidRequest_ReturnsEntries: unexpected error: %v", err)
	}
	if len(resp.GetEntries()) != 2 {
		t.Errorf("TestQueryAuditLog_ValidRequest_ReturnsEntries: got %d entries, want 2", len(resp.GetEntries()))
	}
}
