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

// QueryTracks implements QueryServiceServer.QueryTracks.
//
// Flow:
//  1. Extract caller clearance from gRPC context metadata (server-side, never from request)
//  2. Validate time_range is present and within guardrail limits
//  3. Delegate to TracksRepository (applies classification filter + cursor pagination)
//  4. Emit audit event recording the query (query type, filters, result count, clearance)
//  5. Return response
func (h *Handler) QueryTracks(
ctx context.Context,
req *queryv1.QueryTracksRequest,
) (*queryv1.QueryTracksResponse, error) {
// Step 1: Extract server-side clearance — NEVER use req.ClearanceLevel
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

// Step 3: Execute query via repository
resp, err := h.tracksRepo.QueryTracks(ctx, req, clearance)
if err != nil {
slog.ErrorContext(ctx, "handler.QueryTracks: repository error",
"error", err,
"clearance", clearance.String())
return nil, status.Errorf(codes.Internal, "query execution failed: %v", err)
}

resultCount := 0
if resp != nil && resp.Pagination != nil {
resultCount = int(resp.Pagination.TotalCount)
}

// Step 4: Emit audit event — every query is audited
filterSummary := buildTracksFilterSummary(req)
emitErr := h.auditEmitter.Emit(ctx, h.serviceID, audit.AuditParams{
EventType:    "query_executed",
ResourceType: "tracks",
ResourceID:   "",
Action:       auditv1.AuditAction_AUDIT_ACTION_QUERY,
ActorType:    auditv1.ActorType_ACTOR_TYPE_OPERATOR,
// ActorID: in production, extracted from mTLS cert subject; empty in dev
ClassificationLevel: clearance,
Details: map[string]interface{}{
"query_type":          "QueryTracks",
"operator_clearance":  clearance.String(),
"result_count":        resultCount,
"filters":             filterSummary,
},
})
if emitErr != nil {
// Audit failure is logged but MUST NOT abort the query response.
slog.ErrorContext(ctx, "handler.QueryTracks: audit emit failed",
"error", emitErr)
}

return resp, nil
}

// buildTracksFilterSummary creates a non-sensitive summary of the applied filters for audit.
// MUST NOT include raw sensor data or classified content.
func buildTracksFilterSummary(req *queryv1.QueryTracksRequest) map[string]interface{} {
summary := map[string]interface{}{
"entity_type_count":       len(req.GetEntityTypes()),
"hostile_class_count":     len(req.GetHostileClasses()),
"has_bounding_box":        req.GetBoundingBox() != nil,
"has_min_confidence":      req.GetMinConfidence() > 0,
"min_confidence":          fmt.Sprintf("%.2f", req.GetMinConfidence()),
"has_track_id_filter":     req.GetTrackId() != "",
}
if req.GetPagination() != nil {
summary["page_size"] = req.GetPagination().GetPageSize()
summary["has_page_token"] = req.GetPagination().GetPageToken() != ""
}
return summary
}
