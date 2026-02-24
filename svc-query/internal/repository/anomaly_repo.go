// CLASSIFICATION: UNCLASSIFIED

package repository

import (
"context"
"fmt"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
queryv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/query/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/security"
"google.golang.org/protobuf/types/known/timestamppb"
)

// AnomaliesQuerier defines the contract for anomaly historical queries.
type AnomaliesQuerier interface {
QueryAnomalies(ctx context.Context, req *queryv1.QueryAnomaliesRequest, clearance commonv1.ClassificationLevel) (*queryv1.QueryAnomaliesResponse, error)
}

// AnomalyRepository executes parameterized anomaly queries against ClickHouse.
type AnomalyRepository struct {
client *ClickHouseClient
filter *security.ClassificationFilter
guard  *domain.QueryGuardrail
}

// NewAnomalyRepository creates a new AnomalyRepository.
func NewAnomalyRepository(client *ClickHouseClient, filter *security.ClassificationFilter, guard *domain.QueryGuardrail) *AnomalyRepository {
return &AnomalyRepository{client: client, filter: filter, guard: guard}
}

// anomalyRow holds a single row scanned from the anomaly_detections ClickHouse table.
type anomalyRow struct {
AlertID             string
TrackID             string
AnomalyType         string
Severity            string
ConfidenceScore     float64
Explanation         string
ModelVersion        string
ClassificationLevel string
EventTime           time.Time
}

// QueryAnomalies executes a parameterized ClickHouse query against anomaly_detections.
//
// Security: classification filter is always injected server-side.
// All filter predicates use positional parameter binding.
func (r *AnomalyRepository) QueryAnomalies(
ctx context.Context,
req *queryv1.QueryAnomaliesRequest,
clearance commonv1.ClassificationLevel,
) (*queryv1.QueryAnomaliesResponse, error) {
if req.GetTimeRange() == nil {
return nil, fmt.Errorf("repository.AnomalyRepository.QueryAnomalies: time_range is required")
}

startTime := req.GetTimeRange().GetStartTime().AsTime()
endTime := req.GetTimeRange().GetEndTime().AsTime()

query := `SELECT
alert_id,
track_id,
toString(anomaly_type),
toString(severity),
confidence_score,
explanation,
model_version,
toString(classification_level),
event_time
FROM anomaly_detections
WHERE event_time >= ? AND event_time <= ?`

params := []interface{}{startTime, endTime}

// Optional anomaly_type IN filter
if len(req.GetAnomalyTypes()) > 0 {
query += " AND anomaly_type IN ("
for i, at := range req.GetAnomalyTypes() {
if i > 0 {
query += ", "
}
query += "?"
params = append(params, anomalyTypeToString(at))
}
query += ")"
}

// Optional severity IN filter
if len(req.GetSeverities()) > 0 {
query += " AND severity IN ("
for i, sv := range req.GetSeverities() {
if i > 0 {
query += ", "
}
query += "?"
params = append(params, severityToString(sv))
}
query += ")"
}

// Optional track_id filter
if req.GetTrackId() != "" {
query += " AND track_id = ?"
params = append(params, req.GetTrackId())
}

// Optional minimum confidence filter
if req.GetMinConfidence() > 0 {
query += " AND confidence_score >= ?"
params = append(params, req.GetMinConfidence())
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
return nil, fmt.Errorf("repository.AnomalyRepository.QueryAnomalies: decode page token: %w", err)
}
}

// Apply cursor-based ORDER BY/LIMIT (using alert_id as secondary key)
query, params = domain.ApplyPagination(query, params, pageToken, "alert_id", pageSize)

// Execute with guardrail timeout
qCtx, cancel := r.guard.QueryContext(ctx)
defer cancel()

rows, err := r.client.conn.Query(qCtx, query, params...)
if err != nil {
return nil, fmt.Errorf("repository.AnomalyRepository.QueryAnomalies: execute: %w", err)
}
defer rows.Close()

var alerts []*inferencev1.AnomalyAlert
var lastRow *anomalyRow

for rows.Next() {
var row anomalyRow
if err := rows.Scan(
&row.AlertID,
&row.TrackID,
&row.AnomalyType,
&row.Severity,
&row.ConfidenceScore,
&row.Explanation,
&row.ModelVersion,
&row.ClassificationLevel,
&row.EventTime,
); err != nil {
return nil, fmt.Errorf("repository.AnomalyRepository.QueryAnomalies: scan row: %w", err)
}
alerts = append(alerts, anomalyRowToProto(&row))
lastRow = &row
}
if err := rows.Err(); err != nil {
return nil, fmt.Errorf("repository.AnomalyRepository.QueryAnomalies: rows iteration: %w", err)
}

resp := &queryv1.QueryAnomaliesResponse{
Alerts: alerts,
Pagination: &commonv1.PaginationResponse{
TotalCount: int32(len(alerts)),
},
}

if len(alerts) == pageSize && lastRow != nil {
nextToken := &domain.PaginationToken{
LastID:        lastRow.AlertID,
LastTimestamp: lastRow.EventTime,
PageSize:      pageSize,
}
resp.Pagination.NextPageToken = domain.EncodePaginationToken(nextToken)
}

return resp, nil
}

// anomalyRowToProto converts a scanned anomaly row to an AnomalyAlert proto.
func anomalyRowToProto(row *anomalyRow) *inferencev1.AnomalyAlert {
return &inferencev1.AnomalyAlert{
AlertId:         row.AlertID,
TrackId:         row.TrackID,
AnomalyType:     parseAnomalyType(row.AnomalyType),
Severity:        parseSeverity(row.Severity),
ConfidenceScore: row.ConfidenceScore,
Explanation:     row.Explanation,
ModelVersion:    row.ModelVersion,
Classification:  parseClassificationLevel(row.ClassificationLevel),
DetectedAt:      timestamppb.New(row.EventTime),
}
}

// --- enum converters ---

func anomalyTypeToString(at commonv1.AnomalyType) string {
switch at {
case commonv1.AnomalyType_ANOMALY_TYPE_SPEED:
return "SPEED"
case commonv1.AnomalyType_ANOMALY_TYPE_ROUTE_DEVIATION:
return "ROUTE_DEVIATION"
case commonv1.AnomalyType_ANOMALY_TYPE_AIS_MANIPULATION:
return "AIS_MANIPULATION"
case commonv1.AnomalyType_ANOMALY_TYPE_BEHAVIORAL:
return "BEHAVIORAL"
case commonv1.AnomalyType_ANOMALY_TYPE_TEMPORAL:
return "TEMPORAL"
case commonv1.AnomalyType_ANOMALY_TYPE_PROXIMITY:
return "PROXIMITY"
default:
return "UNSPECIFIED"
}
}

func severityToString(sv commonv1.AlertSeverity) string {
switch sv {
case commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL:
return "NORMAL"
case commonv1.AlertSeverity_ALERT_SEVERITY_WATCH:
return "WATCH"
case commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED:
return "ELEVATED"
case commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL:
return "CRITICAL"
default:
return "NORMAL"
}
}

func parseAnomalyType(s string) commonv1.AnomalyType {
switch s {
case "SPEED":
return commonv1.AnomalyType_ANOMALY_TYPE_SPEED
case "ROUTE_DEVIATION":
return commonv1.AnomalyType_ANOMALY_TYPE_ROUTE_DEVIATION
case "AIS_MANIPULATION":
return commonv1.AnomalyType_ANOMALY_TYPE_AIS_MANIPULATION
case "BEHAVIORAL":
return commonv1.AnomalyType_ANOMALY_TYPE_BEHAVIORAL
case "TEMPORAL":
return commonv1.AnomalyType_ANOMALY_TYPE_TEMPORAL
case "PROXIMITY":
return commonv1.AnomalyType_ANOMALY_TYPE_PROXIMITY
default:
return commonv1.AnomalyType_ANOMALY_TYPE_UNSPECIFIED
}
}

func parseSeverity(s string) commonv1.AlertSeverity {
switch s {
case "NORMAL":
return commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL
case "WATCH":
return commonv1.AlertSeverity_ALERT_SEVERITY_WATCH
case "ELEVATED":
return commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED
case "CRITICAL":
return commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL
default:
return commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL
}
}
