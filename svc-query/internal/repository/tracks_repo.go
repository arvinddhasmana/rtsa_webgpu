// CLASSIFICATION: UNCLASSIFIED
package repository

import (
"context"
"fmt"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
queryv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/query/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/security"
"google.golang.org/protobuf/types/known/timestamppb"
)

// TracksRepository handles track queries against ClickHouse.
type TracksRepository struct {
client *ClickHouseClient
filter *security.ClassificationFilter
}

// NewTracksRepository creates a new TracksRepository.
func NewTracksRepository(client *ClickHouseClient) *TracksRepository {
return &TracksRepository{
client: client,
filter: &security.ClassificationFilter{},
}
}

// QueryTracks executes a parameterized track query with classification filtering.
func (r *TracksRepository) QueryTracks(
ctx context.Context,
req *queryv1.QueryTracksRequest,
clearance commonv1.ClassificationLevel,
pageToken *domain.PaginationToken,
pageSize int,
) ([]*entityv1.FusedTrack, *domain.PaginationToken, error) {
baseQuery := `SELECT track_id, entity_type, hostile_classification,
		latitude, longitude, altitude_meters, speed_knots, heading_degrees,
		confidence_score, classification_level, event_time
		FROM tracks_fused
		WHERE event_time >= ? AND event_time <= ?`

params := []interface{}{
req.GetTimeRange().GetStartTime().AsTime(),
req.GetTimeRange().GetEndTime().AsTime(),
}

// Classification filter (always applied — matches LowCardinality(String) column)
query, classParams := r.filter.InjectFilter(baseQuery, clearance)
params = append(params, classParams...)

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
query += " AND (event_time, track_id) > (?, ?)"
params = append(params, pageToken.LastTimestamp, pageToken.LastID)
}
query += fmt.Sprintf(" ORDER BY event_time ASC, track_id ASC LIMIT %d", pageSize)

rows, err := r.client.conn.Query(ctx, query, params...)
if err != nil {
return nil, nil, fmt.Errorf("tracks_repo: query: %w", err)
}
defer rows.Close()

var tracks []*entityv1.FusedTrack
var lastID string
var lastTS time.Time

for rows.Next() {
var (
trackID      string
entityType   string // LowCardinality(String) in ClickHouse
hostileClass string // LowCardinality(String) in ClickHouse
lat, lon     float64
altM         float64
speedKn      float64
headingDeg   float64
confidence   float64
classLevel   string // LowCardinality(String) in ClickHouse
eventTime    time.Time
)
if err := rows.Scan(
&trackID, &entityType, &hostileClass,
&lat, &lon, &altM, &speedKn, &headingDeg,
&confidence, &classLevel, &eventTime,
); err != nil {
return nil, nil, fmt.Errorf("tracks_repo: scan: %w", err)
}

track := &entityv1.FusedTrack{
TrackId:      trackID,
EntityType:   commonv1.EntityType(commonv1.EntityType_value[entityType]),
HostileClass: commonv1.HostileClassification(commonv1.HostileClassification_value[hostileClass]),
EstimatedPosition: &commonv1.Position{
Latitude:       lat,
Longitude:      lon,
AltitudeMeters: &altM,
SpeedKnots:     &speedKn,
HeadingDegrees: &headingDeg,
},
ConfidenceScore: confidence,
Classification:  commonv1.ClassificationLevel(commonv1.ClassificationLevel_value[classLevel]),
UpdatedAt:       timestamppb.New(eventTime),
}
tracks = append(tracks, track)
lastID = trackID
lastTS = eventTime
}
if err := rows.Err(); err != nil {
return nil, nil, fmt.Errorf("tracks_repo: rows error: %w", err)
}

var nextToken *domain.PaginationToken
if len(tracks) == pageSize {
nextToken = &domain.PaginationToken{
LastID:        lastID,
LastTimestamp: lastTS,
PageSize:      pageSize,
}
}

return tracks, nextToken, nil
}
