// CLASSIFICATION: UNCLASSIFIED

package repository

import (
"context"
"fmt"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
queryv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/query/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/security"
)

// AuditQuerier defines the contract for audit log historical queries.
type AuditQuerier interface {
QueryAuditLog(ctx context.Context, req *queryv1.QueryAuditLogRequest, clearance commonv1.ClassificationLevel) (*queryv1.QueryAuditLogResponse, error)
}

// AuditRepository executes parameterized audit log queries against ClickHouse.
//
// IMPORTANT: The audit_log table has NO TTL — it is retained indefinitely
// per ITSG-33 accountability requirements. This repository must NEVER issue
// DELETE or TRUNCATE statements against audit_log.
type AuditRepository struct {
client *ClickHouseClient
filter *security.ClassificationFilter
guard  *domain.QueryGuardrail
}

// NewAuditRepository creates a new AuditRepository.
func NewAuditRepository(client *ClickHouseClient, filter *security.ClassificationFilter, guard *domain.QueryGuardrail) *AuditRepository {
return &AuditRepository{client: client, filter: filter, guard: guard}
}

// auditRow holds a single row scanned from the audit_log ClickHouse table.
type auditRow struct {
AuditID             string
ServiceID           string
EventType           string
ActorID             string
ActorType           string
ResourceType        string
ResourceID          string
Action              string
DetailJSON          string
ClassificationLevel string
EventTime           time.Time
}

// QueryAuditLog executes a parameterized ClickHouse query against the audit_log table.
//
// Security: classification filter is always injected server-side.
// All filter predicates use positional parameter binding.
//
// Note: audit_log has no TTL — it may contain many years of records.
// Time range guardrails are especially important here.
func (r *AuditRepository) QueryAuditLog(
ctx context.Context,
req *queryv1.QueryAuditLogRequest,
clearance commonv1.ClassificationLevel,
) (*queryv1.QueryAuditLogResponse, error) {
if req.GetTimeRange() == nil {
return nil, fmt.Errorf("repository.AuditRepository.QueryAuditLog: time_range is required")
}

startTime := req.GetTimeRange().GetStartTime().AsTime()
endTime := req.GetTimeRange().GetEndTime().AsTime()

query := `SELECT
audit_id,
service_id,
event_type,
actor_id,
toString(actor_type),
resource_type,
resource_id,
action,
detail_json,
toString(classification_level),
event_time
FROM audit_log
WHERE event_time >= ? AND event_time <= ?`

params := []interface{}{startTime, endTime}

// Optional service_id filter (scalar, not IN — per proto definition)
if req.GetServiceId() != "" {
query += " AND service_id = ?"
params = append(params, req.GetServiceId())
}

// Optional event_type filter
if req.GetEventType() != "" {
query += " AND event_type = ?"
params = append(params, req.GetEventType())
}

// Optional actor_id filter
if req.GetActorId() != "" {
query += " AND actor_id = ?"
params = append(params, req.GetActorId())
}

// Optional resource_type filter
if req.GetResourceType() != "" {
query += " AND resource_type = ?"
params = append(params, req.GetResourceType())
}

// MANDATORY server-side classification filter
var classOrdinal int8
query, classOrdinal = r.filter.InjectFilter(query, clearance)
params = append(params, classOrdinal)

// Determine effective page size
pageSize := r.guard.DefaultPageSize
if req.GetPagination() != nil {
pageSize = r.guard.EnforcePageSize(int(req.GetPagination().GetPageSize()))
}

// Decode cursor token
var pageToken *domain.PaginationToken
if req.GetPagination() != nil && req.GetPagination().GetPageToken() != "" {
var err error
pageToken, err = domain.DecodePaginationToken(req.GetPagination().GetPageToken())
if err != nil {
return nil, fmt.Errorf("repository.AuditRepository.QueryAuditLog: decode page token: %w", err)
}
}

// Apply cursor-based ORDER BY/LIMIT (using audit_id as secondary key)
query, params = domain.ApplyPagination(query, params, pageToken, "audit_id", pageSize)

// Execute with guardrail timeout
qCtx, cancel := r.guard.QueryContext(ctx)
defer cancel()

rows, err := r.client.conn.Query(qCtx, query, params...)
if err != nil {
return nil, fmt.Errorf("repository.AuditRepository.QueryAuditLog: execute: %w", err)
}
defer rows.Close()

var entries []*queryv1.AuditLogEntry
var lastRow *auditRow

for rows.Next() {
var row auditRow
if err := rows.Scan(
&row.AuditID,
&row.ServiceID,
&row.EventType,
&row.ActorID,
&row.ActorType,
&row.ResourceType,
&row.ResourceID,
&row.Action,
&row.DetailJSON,
&row.ClassificationLevel,
&row.EventTime,
); err != nil {
return nil, fmt.Errorf("repository.AuditRepository.QueryAuditLog: scan row: %w", err)
}
entries = append(entries, auditRowToProto(&row))
lastRow = &row
}
if err := rows.Err(); err != nil {
return nil, fmt.Errorf("repository.AuditRepository.QueryAuditLog: rows iteration: %w", err)
}

resp := &queryv1.QueryAuditLogResponse{
Entries: entries,
Pagination: &commonv1.PaginationResponse{
TotalCount: int32(len(entries)),
},
}

if len(entries) == pageSize && lastRow != nil {
nextToken := &domain.PaginationToken{
LastID:        lastRow.AuditID,
LastTimestamp: lastRow.EventTime,
PageSize:      pageSize,
}
resp.Pagination.NextPageToken = domain.EncodePaginationToken(nextToken)
}

return resp, nil
}

// auditRowToProto converts a scanned audit row to an AuditLogEntry proto.
func auditRowToProto(row *auditRow) *queryv1.AuditLogEntry {
return &queryv1.AuditLogEntry{
AuditId:             row.AuditID,
ServiceId:           row.ServiceID,
EventType:           row.EventType,
ActorId:             row.ActorID,
ActorType:           row.ActorType,
ResourceType:        row.ResourceType,
ResourceId:          row.ResourceID,
Action:              row.Action,
DetailJson:          row.DetailJSON,
ClassificationLevel: parseClassificationLevel(row.ClassificationLevel),
EventTime:           row.EventTime.UTC().Format(time.RFC3339Nano),
}
}
