// CLASSIFICATION: UNCLASSIFIED
package handler

import (
"context"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
queryv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/query/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/repository"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/security"
"go.uber.org/zap"
"google.golang.org/grpc/codes"
"google.golang.org/grpc/status"
)

// QueryServer implements the queryv1.QueryServiceServer interface.
type QueryServer struct {
queryv1.UnimplementedQueryServiceServer
tracksRepo  *repository.TracksRepository
anomalyRepo *repository.AnomalyRepository
auditRepo   *repository.AuditRepository
guardrail   *domain.QueryGuardrail
auditEmitter AuditEmitter
pageSize    int
logger      *zap.Logger
}

// AuditEmitter is a minimal interface for emitting audit events.
type AuditEmitter interface {
Emit(ctx context.Context, eventType, resourceType, resourceID string)
}

// noopAuditEmitter is a no-op audit emitter used when no real emitter is configured.
type noopAuditEmitter struct{}

func (n *noopAuditEmitter) Emit(_ context.Context, _, _, _ string) {}

// NewQueryServer creates a new gRPC query server.
func NewQueryServer(
tracksRepo *repository.TracksRepository,
anomalyRepo *repository.AnomalyRepository,
auditRepo *repository.AuditRepository,
guardrail *domain.QueryGuardrail,
pageSize int,
logger *zap.Logger,
) *QueryServer {
return &QueryServer{
tracksRepo:   tracksRepo,
anomalyRepo:  anomalyRepo,
auditRepo:    auditRepo,
guardrail:    guardrail,
auditEmitter: &noopAuditEmitter{},
pageSize:     pageSize,
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
