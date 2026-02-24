// CLASSIFICATION: UNCLASSIFIED

package handler

import (
"context"
"log/slog"

auditv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/audit/v1"
queryv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/query/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/audit"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/security"
"google.golang.org/grpc/codes"
"google.golang.org/grpc/status"
)

// QueryAuditLog implements QueryServiceServer.QueryAuditLog.
//
// Flow:
//  1. Extract caller clearance from gRPC context metadata (server-side)
//  2. Validate time_range present and within guardrail limits
//  3. Delegate to AuditRepository (classification-filtered)
//  4. Emit meta-audit event (querying the audit log is itself audited)
//  5. Return response
func (h *Handler) QueryAuditLog(
ctx context.Context,
req *queryv1.QueryAuditLogRequest,
) (*queryv1.QueryAuditLogResponse, error) {
// Step 1: Server-side clearance
clearance := security.ExtractClearance(ctx)

// Step 2: Validate time range
if req.GetTimeRange() == nil {
return nil, status.Error(codes.InvalidArgument, "time_range is required")
}
if err := h.guard.ValidateTimeRange(
req.GetTimeRange().GetStartTime(),
req.GetTimeRange().GetEndTime(),
); err != nil {
return nil, err
}

// Step 3: Execute query
resp, err := h.auditRepo.QueryAuditLog(ctx, req, clearance)
if err != nil {
slog.ErrorContext(ctx, "handler.QueryAuditLog: repository error",
"error", err,
"clearance", clearance.String())
return nil, status.Errorf(codes.Internal, "query execution failed: %v", err)
}

resultCount := 0
if resp != nil && resp.Pagination != nil {
resultCount = int(resp.Pagination.TotalCount)
}

// Step 4: Emit meta-audit event (querying the audit log is itself audited)
emitErr := h.auditEmitter.Emit(ctx, h.serviceID, audit.AuditParams{
EventType:    "query_executed",
ResourceType: "audit_log",
ResourceID:   "",
Action:       auditv1.AuditAction_AUDIT_ACTION_QUERY,
ActorType:    auditv1.ActorType_ACTOR_TYPE_OPERATOR,
ClassificationLevel: clearance,
Details: map[string]interface{}{
"query_type":          "QueryAuditLog",
"operator_clearance":  clearance.String(),
"result_count":        resultCount,
"filter_service_id":   req.GetServiceId() != "",
"filter_event_type":   req.GetEventType() != "",
"filter_actor_id":     req.GetActorId() != "",
"filter_resource_type": req.GetResourceType() != "",
},
})
if emitErr != nil {
slog.ErrorContext(ctx, "handler.QueryAuditLog: audit emit failed",
"error", emitErr)
}

return resp, nil
}
