// CLASSIFICATION: UNCLASSIFIED

package handler

import (
"context"
"fmt"
"log/slog"

auditv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/audit/v1"
queryv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/query/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/audit"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/security"
"google.golang.org/grpc/codes"
"google.golang.org/grpc/status"
)

// QueryAnomalies implements QueryServiceServer.QueryAnomalies.
//
// Flow:
//  1. Extract caller clearance from gRPC context metadata (server-side)
//  2. Validate time_range present and within guardrail limits
//  3. Delegate to AnomalyRepository
//  4. Emit audit event
//  5. Return response
func (h *Handler) QueryAnomalies(
ctx context.Context,
req *queryv1.QueryAnomaliesRequest,
) (*queryv1.QueryAnomaliesResponse, error) {
// Step 1: Server-side clearance extraction
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
resp, err := h.anomalyRepo.QueryAnomalies(ctx, req, clearance)
if err != nil {
slog.ErrorContext(ctx, "handler.QueryAnomalies: repository error",
"error", err,
"clearance", clearance.String())
return nil, status.Errorf(codes.Internal, "query execution failed: %v", err)
}

resultCount := 0
if resp != nil && resp.Pagination != nil {
resultCount = int(resp.Pagination.TotalCount)
}

// Step 4: Emit audit event
filterSummary := buildAnomaliesFilterSummary(req)
emitErr := h.auditEmitter.Emit(ctx, h.serviceID, audit.AuditParams{
EventType:    "query_executed",
ResourceType: "anomalies",
ResourceID:   "",
Action:       auditv1.AuditAction_AUDIT_ACTION_QUERY,
ActorType:    auditv1.ActorType_ACTOR_TYPE_OPERATOR,
ClassificationLevel: clearance,
Details: map[string]interface{}{
"query_type":         "QueryAnomalies",
"operator_clearance": clearance.String(),
"result_count":       resultCount,
"filters":            filterSummary,
},
})
if emitErr != nil {
slog.ErrorContext(ctx, "handler.QueryAnomalies: audit emit failed",
"error", emitErr)
}

return resp, nil
}

func buildAnomaliesFilterSummary(req *queryv1.QueryAnomaliesRequest) map[string]interface{} {
summary := map[string]interface{}{
"anomaly_type_count":  len(req.GetAnomalyTypes()),
"severity_count":      len(req.GetSeverities()),
"has_track_id_filter": req.GetTrackId() != "",
"min_confidence":      fmt.Sprintf("%.2f", req.GetMinConfidence()),
}
if req.GetPagination() != nil {
summary["page_size"] = req.GetPagination().GetPageSize()
summary["has_page_token"] = req.GetPagination().GetPageToken() != ""
}
return summary
}
