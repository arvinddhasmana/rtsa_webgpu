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

// AuditRepository handles audit log queries against ClickHouse.
type AuditRepository struct {
client *ClickHouseClient
filter *security.ClassificationFilter
}

// NewAuditRepository creates a new AuditRepository.
func NewAuditRepository(client *ClickHouseClient) *AuditRepository {
return &AuditRepository{
client: client,
filter: &security.ClassificationFilter{},
}
}

// QueryAuditLog executes a parameterized audit log query with classification filtering.
func (r *AuditRepository) QueryAuditLog(
ctx context.Context,
req *queryv1.QueryAuditLogRequest,
clearance commonv1.ClassificationLevel,
pageToken *domain.PaginationToken,
pageSize int,
) ([]*queryv1.AuditLogEntry, *domain.PaginationToken, error) {
baseQuery := `SELECT audit_id, service_id, event_type, actor_id,
actor_type, resource_type, resource_id, action, detail_json,
classification_level, event_time
FROM audit_log
WHERE event_time >= ? AND event_time <= ?`

params := []interface{}{
req.GetTimeRange().GetStartTime().AsTime(),
req.GetTimeRange().GetEndTime().AsTime(),
}

// Classification filter (always applied — matches LowCardinality(String) column)
query, classParams := r.filter.InjectFilter(baseQuery, clearance)
params = append(params, classParams...)

// Optional filters
if req.ServiceId != nil {
query += " AND service_id = ?"
params = append(params, *req.ServiceId)
}
if req.EventType != nil {
query += " AND event_type = ?"
params = append(params, *req.EventType)
}
if req.ActorId != nil {
query += " AND actor_id = ?"
params = append(params, *req.ActorId)
}
if req.ResourceType != nil {
query += " AND resource_type = ?"
params = append(params, *req.ResourceType)
}

// Pagination
if pageToken != nil {
query += " AND (event_time, audit_id) > (?, ?)"
params = append(params, pageToken.LastTimestamp, pageToken.LastID)
}
query += fmt.Sprintf(" ORDER BY event_time ASC, audit_id ASC LIMIT %d", pageSize)

rows, err := r.client.conn.Query(ctx, query, params...)
if err != nil {
return nil, nil, fmt.Errorf("audit_repo: query: %w", err)
}
defer rows.Close()

var entries []*queryv1.AuditLogEntry
var lastID string
var lastTS time.Time

for rows.Next() {
var (
auditID      string
serviceID    string
eventType    string
actorID      string
actorType    string
resourceType string
resourceID   string
action       string
detailJSON   string
classLevel   string // LowCardinality(String) in ClickHouse
eventTime    time.Time
)
if err := rows.Scan(
&auditID, &serviceID, &eventType, &actorID,
&actorType, &resourceType, &resourceID, &action, &detailJSON,
&classLevel, &eventTime,
); err != nil {
return nil, nil, fmt.Errorf("audit_repo: scan: %w", err)
}

entry := &queryv1.AuditLogEntry{
AuditId:             auditID,
ServiceId:           serviceID,
EventType:           eventType,
ActorId:             actorID,
ActorType:           actorType,
ResourceType:        resourceType,
ResourceId:          resourceID,
Action:              action,
DetailJson:          detailJSON,
ClassificationLevel: commonv1.ClassificationLevel(commonv1.ClassificationLevel_value[classLevel]),
EventTime:           eventTime.UTC().Format(time.RFC3339),
}
entries = append(entries, entry)
lastID = auditID
lastTS = eventTime
}
if err := rows.Err(); err != nil {
return nil, nil, fmt.Errorf("audit_repo: rows error: %w", err)
}

var nextToken *domain.PaginationToken
if len(entries) == pageSize {
nextToken = &domain.PaginationToken{
LastID:        lastID,
LastTimestamp: lastTS,
PageSize:      pageSize,
}
}

return entries, nextToken, nil
}
