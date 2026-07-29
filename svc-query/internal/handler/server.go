// CLASSIFICATION: UNCLASSIFIED
package handler

import (
"context"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
entityv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/entity/v1"
inferencev1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/rtsa_webgpu/pkg/audit"
queryv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/query/v1"
auditv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/audit/v1"
"github.com/arvinddhasmana/rtsa_webgpu/svc-query/internal/domain"
"github.com/arvinddhasmana/rtsa_webgpu/svc-query/internal/security"
"go.uber.org/zap"
"google.golang.org/grpc/codes"
"google.golang.org/grpc/status"
)

// Repository interfaces for mocking.
type TracksRepository interface {
	QueryTracks(ctx context.Context, req *queryv1.QueryTracksRequest, clearance commonv1.ClassificationLevel, pageToken *domain.PaginationToken, pageSize int) ([]*entityv1.FusedTrack, *domain.PaginationToken, error)
}
type AnomalyRepository interface {
	QueryAnomalies(ctx context.Context, req *queryv1.QueryAnomaliesRequest, clearance commonv1.ClassificationLevel, pageToken *domain.PaginationToken, pageSize int) ([]*inferencev1.AnomalyAlert, *domain.PaginationToken, error)
}
type AuditRepository interface {
	QueryAuditLog(ctx context.Context, req *queryv1.QueryAuditLogRequest, clearance commonv1.ClassificationLevel, pageToken *domain.PaginationToken, pageSize int) ([]*queryv1.AuditLogEntry, *domain.PaginationToken, error)
}
type TimelineRepository interface {
	GetEventTimeline(ctx context.Context, req *queryv1.GetEventTimelineRequest, clearance commonv1.ClassificationLevel) ([]*queryv1.TimelineEvent, error)
}

// QueryServer implements the queryv1.QueryServiceServer interface.
type QueryServer struct {
	queryv1.UnimplementedQueryServiceServer
	tracksRepo   TracksRepository
	anomalyRepo  AnomalyRepository
	auditRepo    AuditRepository
	timelineRepo TimelineRepository
	guardrail    *domain.QueryGuardrail
	auditEmitter AuditEmitter
	pageSize     int
	logger       *zap.Logger
}

// AuditEmitter is a minimal interface for emitting audit events.
type AuditEmitter interface {
	Emit(ctx context.Context, eventType, resourceType, resourceID string)
}

// logAuditEmitter adapts pkg/audit.Emitter to the handler.AuditEmitter interface.
// It emits structured audit events using the real pkg/audit infrastructure.
type logAuditEmitter struct{ emitter *audit.Emitter }

func (l *logAuditEmitter) Emit(ctx context.Context, eventType, resourceType, resourceID string) {
	l.emitter.Emit(ctx, audit.AuditParams{
		EventType:    eventType,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       auditv1.AuditAction_AUDIT_ACTION_READ,
	})
}

// NewQueryServer creates a new gRPC query server.
func NewQueryServer(
	tracksRepo TracksRepository,
	anomalyRepo AnomalyRepository,
	auditRepo AuditRepository,
	timelineRepo TimelineRepository,
	guardrail *domain.QueryGuardrail,
	pageSize int,
	logger *zap.Logger,
) *QueryServer {
	return &QueryServer{
		tracksRepo:   tracksRepo,
		anomalyRepo:  anomalyRepo,
		auditRepo:    auditRepo,
		timelineRepo: timelineRepo,
		guardrail:    guardrail,
		auditEmitter: &logAuditEmitter{emitter: audit.NewLogEmitter(logger)},
		pageSize:     pageSize,
		logger:       logger,
	}
}

// NewQueryServerForTest creates a QueryServer suitable for unit tests.
func NewQueryServerForTest(
	guardrail *domain.QueryGuardrail,
	logger *zap.Logger,
	tracksRepo TracksRepository,
	anomalyRepo AnomalyRepository,
	auditRepo AuditRepository,
	timelineRepo TimelineRepository,
) *QueryServer {
	return &QueryServer{
		tracksRepo:   tracksRepo,
		anomalyRepo:  anomalyRepo,
		auditRepo:    auditRepo,
		timelineRepo: timelineRepo,
		guardrail:    guardrail,
		auditEmitter: &logAuditEmitter{emitter: audit.NewLogEmitter(logger)},
		pageSize:     100,
		logger:       logger,
	}
}

// QueryTracks handles historical track queries from ClickHouse.
func (s *QueryServer) QueryTracks(
	ctx context.Context,
	req *queryv1.QueryTracksRequest,
) (*queryv1.QueryTracksResponse, error) {
if err := s.guardrail.ValidateTimeRange(
req.GetTimeRange().GetStartTime(),
req.GetTimeRange().GetEndTime(),
); err != nil {
return nil, err
}

clearance := security.ExtractClearance(ctx)
pageToken, err := domain.DecodePaginationToken(req.GetPagination().GetPageToken())
if err != nil {
return nil, status.Error(codes.InvalidArgument, err.Error())
}

ps := s.pageSize
if req.GetPagination().GetPageSize() > 0 {
ps = int(req.GetPagination().GetPageSize())
}
ps = s.guardrail.EnforceRowLimit(ps)

qctx, cancel := s.guardrail.QueryContext(ctx)
defer cancel()

tracks, nextToken, err := s.tracksRepo.QueryTracks(qctx, req, clearance, pageToken, ps)
if err != nil {
s.logger.Error("QueryTracks: repository error", zap.Error(err))
return nil, status.Error(codes.Internal, "query failed")
}

s.auditEmitter.Emit(ctx, "query.executed", "tracks_fused", "")

resp := &queryv1.QueryTracksResponse{
Tracks: tracks,
Pagination: &commonv1.PaginationResponse{
NextPageToken: domain.EncodePaginationToken(nextToken),
TotalCount:    int32(len(tracks)),
},
}
return resp, nil
}

// QueryAnomalies handles historical anomaly detection queries from ClickHouse.
func (s *QueryServer) QueryAnomalies(
ctx context.Context,
req *queryv1.QueryAnomaliesRequest,
) (*queryv1.QueryAnomaliesResponse, error) {
if err := s.guardrail.ValidateTimeRange(
req.GetTimeRange().GetStartTime(),
req.GetTimeRange().GetEndTime(),
); err != nil {
return nil, err
}

clearance := security.ExtractClearance(ctx)
pageToken, err := domain.DecodePaginationToken(req.GetPagination().GetPageToken())
if err != nil {
return nil, status.Error(codes.InvalidArgument, err.Error())
}

ps := s.pageSize
if req.GetPagination().GetPageSize() > 0 {
ps = int(req.GetPagination().GetPageSize())
}
ps = s.guardrail.EnforceRowLimit(ps)

qctx, cancel := s.guardrail.QueryContext(ctx)
defer cancel()

alerts, nextToken, err := s.anomalyRepo.QueryAnomalies(qctx, req, clearance, pageToken, ps)
if err != nil {
s.logger.Error("QueryAnomalies: repository error", zap.Error(err))
return nil, status.Error(codes.Internal, "query failed")
}

s.auditEmitter.Emit(ctx, "query.executed", "anomaly_detections", "")

resp := &queryv1.QueryAnomaliesResponse{
Alerts: alerts,
Pagination: &commonv1.PaginationResponse{
NextPageToken: domain.EncodePaginationToken(nextToken),
TotalCount:    int32(len(alerts)),
},
}
return resp, nil
}

// QueryAuditLog handles historical audit log queries from ClickHouse.
func (s *QueryServer) QueryAuditLog(
ctx context.Context,
req *queryv1.QueryAuditLogRequest,
) (*queryv1.QueryAuditLogResponse, error) {
if err := s.guardrail.ValidateTimeRange(
req.GetTimeRange().GetStartTime(),
req.GetTimeRange().GetEndTime(),
); err != nil {
return nil, err
}

clearance := security.ExtractClearance(ctx)
pageToken, err := domain.DecodePaginationToken(req.GetPagination().GetPageToken())
if err != nil {
return nil, status.Error(codes.InvalidArgument, err.Error())
}

ps := s.pageSize
if req.GetPagination().GetPageSize() > 0 {
ps = int(req.GetPagination().GetPageSize())
}
ps = s.guardrail.EnforceRowLimit(ps)

qctx, cancel := s.guardrail.QueryContext(ctx)
defer cancel()

entries, nextToken, err := s.auditRepo.QueryAuditLog(qctx, req, clearance, pageToken, ps)
if err != nil {
s.logger.Error("QueryAuditLog: repository error", zap.Error(err))
return nil, status.Error(codes.Internal, "query failed")
}

s.auditEmitter.Emit(ctx, "query.executed", "audit_log", "")

resp := &queryv1.QueryAuditLogResponse{
Entries: entries,
Pagination: &commonv1.PaginationResponse{
NextPageToken: domain.EncodePaginationToken(nextToken),
TotalCount:    int32(len(entries)),
},
}
return resp, nil
}

// Ensure QueryServer implements the interface at compile time.
var _ queryv1.QueryServiceServer = (*QueryServer)(nil)

// Suppress unused imports.
var _ = entityv1.FusedTrack{}
var _ = inferencev1.AnomalyAlert{}
