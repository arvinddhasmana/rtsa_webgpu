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

// AnomalyRepository handles anomaly detection history queries.
type AnomalyRepository struct {
client *ClickHouseClient
filter *security.ClassificationFilter
}

// NewAnomalyRepository creates a new AnomalyRepository.
func NewAnomalyRepository(client *ClickHouseClient) *AnomalyRepository {
return &AnomalyRepository{
client: client,
filter: &security.ClassificationFilter{},
}
}

// QueryAnomalies executes a parameterized anomaly query with classification filtering.
func (r *AnomalyRepository) QueryAnomalies(
ctx context.Context,
req *queryv1.QueryAnomaliesRequest,
clearance commonv1.ClassificationLevel,
pageToken *domain.PaginationToken,
pageSize int,
) ([]*inferencev1.AnomalyAlert, *domain.PaginationToken, error) {
baseQuery := `SELECT alert_id, track_id, anomaly_type, severity,
confidence_score, explanation, model_version, classification_level, event_time
FROM anomaly_detections
WHERE event_time >= ? AND event_time <= ?`

params := []interface{}{
req.GetTimeRange().GetStartTime().AsTime(),
req.GetTimeRange().GetEndTime().AsTime(),
}

// Classification filter (always applied)
query, classParam := r.filter.InjectFilter(baseQuery, clearance)
params = append(params, classParam)

// Optional track ID filter
if req.TrackId != nil {
query += " AND track_id = ?"
params = append(params, *req.TrackId)
}

// Minimum confidence filter
if req.MinConfidence > 0 {
query += " AND confidence_score >= ?"
params = append(params, req.MinConfidence)
}

// Pagination
if pageToken != nil {
query += " AND (event_time, alert_id) > (?, ?)"
params = append(params, pageToken.LastTimestamp, pageToken.LastID)
}
query += fmt.Sprintf(" ORDER BY event_time ASC, alert_id ASC LIMIT %d", pageSize)

rows, err := r.client.conn.Query(ctx, query, params...)
if err != nil {
return nil, nil, fmt.Errorf("anomaly_repo: query: %w", err)
}
defer rows.Close()

var alerts []*inferencev1.AnomalyAlert
var lastID string
var lastTS time.Time

for rows.Next() {
var (
alertID     string
trackID     string
anomalyType int32
severity    int32
confidence  float64
explanation string
modelVer    string
classLevel  int32
eventTime   time.Time
)
if err := rows.Scan(
&alertID, &trackID, &anomalyType, &severity,
&confidence, &explanation, &modelVer, &classLevel, &eventTime,
); err != nil {
return nil, nil, fmt.Errorf("anomaly_repo: scan: %w", err)
}

alert := &inferencev1.AnomalyAlert{
AlertId:         alertID,
TrackId:         trackID,
AnomalyType:     commonv1.AnomalyType(anomalyType),
Severity:        commonv1.AlertSeverity(severity),
ConfidenceScore: confidence,
Explanation:     explanation,
ModelVersion:    modelVer,
Classification:  commonv1.ClassificationLevel(classLevel),
DetectedAt:      timestamppb.New(eventTime),
}
alerts = append(alerts, alert)
lastID = alertID
lastTS = eventTime
}
if err := rows.Err(); err != nil {
return nil, nil, fmt.Errorf("anomaly_repo: rows error: %w", err)
}

var nextToken *domain.PaginationToken
if len(alerts) == pageSize {
nextToken = &domain.PaginationToken{
LastID:        lastID,
LastTimestamp: lastTS,
PageSize:      pageSize,
}
}

return alerts, nextToken, nil
}
