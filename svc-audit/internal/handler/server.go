// CLASSIFICATION: UNCLASSIFIED
package handler

import (
"context"
"fmt"

auditv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/audit/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-audit/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-audit/internal/repository"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-audit/internal/security"
"go.uber.org/zap"
"google.golang.org/grpc"
"google.golang.org/grpc/codes"
"google.golang.org/grpc/status"
)

// AuditServer implements the auditv1.AuditServiceServer interface.
type AuditServer struct {
auditv1.UnimplementedAuditServiceServer
repo      *repository.AuditRepository
guardrail *domain.QueryGuardrail
pageSize  int
logger    *zap.Logger
}

// NewAuditServer creates a new gRPC audit server.
func NewAuditServer(
repo *repository.AuditRepository,
guardrail *domain.QueryGuardrail,
pageSize int,
logger *zap.Logger,
) *AuditServer {
return &AuditServer{
repo:      repo,
guardrail: guardrail,
pageSize:  pageSize,
logger:    logger,
}
}

// GetAuditEntry returns a single audit event by ID with classification filtering.
// Flow:
//  1. Extract caller clearance from gRPC context
//  2. Query ClickHouse with classification filter
//  3. Return event or NOT_FOUND
func (s *AuditServer) GetAuditEntry(
ctx context.Context,
req *auditv1.GetAuditEntryRequest,
) (*auditv1.GetAuditEntryResponse, error) {
if req.GetAuditId() == "" {
return nil, status.Error(codes.InvalidArgument, "audit_id is required")
}

clearance := security.ExtractClearance(ctx)

qctx, cancel := s.guardrail.QueryContext(ctx)
defer cancel()

event, err := s.repo.GetEntry(qctx, req.GetAuditId(), clearance)
if err != nil {
s.logger.Error("GetAuditEntry: repository error",
zap.String("audit_id", req.GetAuditId()), zap.Error(err))
return nil, status.Error(codes.Internal, "query failed")
}
if event == nil {
return nil, status.Error(codes.NotFound,
fmt.Sprintf("audit entry %q not found", req.GetAuditId()))
}

return &auditv1.GetAuditEntryResponse{Event: event}, nil
}

// StreamAuditLog streams filtered audit events from ClickHouse.
// Flow:
//  1. Extract caller clearance from gRPC context
//  2. Validate time_range (required, max 90 days)
//  3. Execute paginated query and stream each page
//  4. Continue until all matching events are sent
//
// NOTE: This queries ClickHouse (historical), not real-time Redpanda.
func (s *AuditServer) StreamAuditLog(
req *auditv1.StreamAuditLogRequest,
stream grpc.ServerStreamingServer[auditv1.StreamAuditLogResponse],
) error {
if err := s.guardrail.ValidateTimeRange(
req.GetTimeRange().GetStartTime(),
req.GetTimeRange().GetEndTime(),
); err != nil {
return err
}

clearance := security.ExtractClearance(stream.Context())

ps := s.pageSize
if req.GetPageSize() > 0 {
ps = int(req.GetPageSize())
}
ps = s.guardrail.EnforceRowLimit(ps)

var pageToken *domain.PaginationToken
for {
qctx, cancel := s.guardrail.QueryContext(stream.Context())
events, nextToken, err := s.repo.QueryAuditLog(qctx, req, clearance, pageToken, ps)
cancel()
if err != nil {
s.logger.Error("StreamAuditLog: repository error", zap.Error(err))
return status.Error(codes.Internal, "query failed")
}

for _, event := range events {
if err := stream.Send(&auditv1.StreamAuditLogResponse{Event: event}); err != nil {
return err
}
}

if nextToken == nil {
break // no more pages
}
pageToken = nextToken
}

return nil
}

// Ensure AuditServer implements the interface at compile time.
var _ auditv1.AuditServiceServer = (*AuditServer)(nil)
