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

// TracksQuerier defines the contract for track historical queries.
// This interface enables mock-based handler testing without a ClickHouse dependency.
type TracksQuerier interface {
QueryTracks(ctx context.Context, req *queryv1.QueryTracksRequest, clearance commonv1.ClassificationLevel) (*queryv1.QueryTracksResponse, error)
}

// TracksRepository executes parameterized track queries against ClickHouse.
type TracksRepository struct {
client *ClickHouseClient
filter *security.ClassificationFilter
guard  *domain.QueryGuardrail
}

// NewTracksRepository creates a new TracksRepository.
func NewTracksRepository(client *ClickHouseClient, filter *security.ClassificationFilter, guard *domain.QueryGuardrail) *TracksRepository {
return &TracksRepository{client: client, filter: filter, guard: guard}
}

// trackRow holds a single row scanned from the tracks_fused ClickHouse table.
type trackRow struct {
TrackID               string
EntityType            string
HostileClassification string
Latitude              float64
Longitude             float64
AltitudeMeters        *float64
SpeedKnots            *float64
HeadingDegrees        *float64
ConfidenceScore       float64
SourceCount           uint8
SourceSensors         []string
ClassificationLevel   string
TrackStatus           string
EventTime             time.Time
}

// QueryTracks executes a parameterized ClickHouse query against the tracks_fused table.
//
// Security: classification filter is always injected via server-side clearance.
// Safety:   guardrails are applied before query execution.
// All filter predicates use positional parameter binding — no string interpolation of user input.
func (r *TracksRepository) QueryTracks(
ctx context.Context,
req *queryv1.QueryTracksRequest,
clearance commonv1.ClassificationLevel,
) (*queryv1.QueryTracksResponse, error) {
if req.GetTimeRange() == nil {
return nil, fmt.Errorf("repository.TracksRepository.QueryTracks: time_range is required")
}

startTime := req.GetTimeRange().GetStartTime().AsTime()
endTime := req.GetTimeRange().GetEndTime().AsTime()

query := `SELECT
track_id,
toString(entity_type),
toString(hostile_classification),
latitude,
longitude,
altitude_meters,
speed_knots,
heading_degrees,
confidence_score,
source_count,
source_sensors,
toString(classification_level),
toString(track_status),
event_time
FROM tracks_fused
WHERE event_time >= ? AND event_time <= ?`

params := []interface{}{startTime, endTime}

// Optional entity_type IN filter (parameterized)
if len(req.GetEntityTypes()) > 0 {
query += " AND entity_type IN ("
for i, et := range req.GetEntityTypes() {
if i > 0 {
query += ", "
}
query += "?"
params = append(params, entityTypeToString(et))
}
query += ")"
}

// Optional hostile_classification IN filter (parameterized)
if len(req.GetHostileClasses()) > 0 {
query += " AND hostile_classification IN ("
for i, hc := range req.GetHostileClasses() {
if i > 0 {
query += ", "
}
query += "?"
params = append(params, hostileClassToString(hc))
}
query += ")"
}

// Optional bounding box spatial filter (parameterized)
if bb := req.GetBoundingBox(); bb != nil {
query += " AND latitude >= ? AND latitude <= ? AND longitude >= ? AND longitude <= ?"
params = append(params, bb.GetMinLatitude(), bb.GetMaxLatitude(), bb.GetMinLongitude(), bb.GetMaxLongitude())
}

// Optional minimum confidence filter
if req.GetMinConfidence() > 0 {
query += " AND confidence_score >= ?"
params = append(params, req.GetMinConfidence())
}

// Optional single track_id filter
if req.GetTrackId() != "" {
query += " AND track_id = ?"
params = append(params, req.GetTrackId())
}

// MANDATORY server-side classification filter — always injected, never trusted from client
var classOrdinal int8
query, classOrdinal = r.filter.InjectFilter(query, clearance)
params = append(params, classOrdinal)

// Determine effective page size
pageSize := r.guard.DefaultPageSize
if req.GetPagination() != nil {
pageSize = r.guard.EnforcePageSize(int(req.GetPagination().GetPageSize()))
}

// Decode cursor token (nil on first page)
var pageToken *domain.PaginationToken
if req.GetPagination() != nil && req.GetPagination().GetPageToken() != "" {
var err error
pageToken, err = domain.DecodePaginationToken(req.GetPagination().GetPageToken())
if err != nil {
return nil, fmt.Errorf("repository.TracksRepository.QueryTracks: decode page token: %w", err)
}
}

// Apply cursor-based ORDER BY/LIMIT
query, params = domain.ApplyPagination(query, params, pageToken, "track_id", pageSize)

// Execute with guardrail timeout
qCtx, cancel := r.guard.QueryContext(ctx)
defer cancel()

rows, err := r.client.conn.Query(qCtx, query, params...)
if err != nil {
return nil, fmt.Errorf("repository.TracksRepository.QueryTracks: execute: %w", err)
}
defer rows.Close()

var tracks []*entityv1.FusedTrack
var lastRow *trackRow

for rows.Next() {
var row trackRow
if err := rows.Scan(
&row.TrackID,
&row.EntityType,
&row.HostileClassification,
&row.Latitude,
&row.Longitude,
&row.AltitudeMeters,
&row.SpeedKnots,
&row.HeadingDegrees,
&row.ConfidenceScore,
&row.SourceCount,
&row.SourceSensors,
&row.ClassificationLevel,
&row.TrackStatus,
&row.EventTime,
); err != nil {
return nil, fmt.Errorf("repository.TracksRepository.QueryTracks: scan row: %w", err)
}
tracks = append(tracks, trackRowToProto(&row))
lastRow = &row
}
if err := rows.Err(); err != nil {
return nil, fmt.Errorf("repository.TracksRepository.QueryTracks: rows iteration: %w", err)
}

resp := &queryv1.QueryTracksResponse{
Tracks: tracks,
Pagination: &commonv1.PaginationResponse{
TotalCount: int32(len(tracks)),
},
}

// Emit next page token only when a full page was returned
if len(tracks) == pageSize && lastRow != nil {
nextToken := &domain.PaginationToken{
LastID:        lastRow.TrackID,
LastTimestamp: lastRow.EventTime,
PageSize:      pageSize,
}
resp.Pagination.NextPageToken = domain.EncodePaginationToken(nextToken)
}

return resp, nil
}

// trackRowToProto converts a scanned ClickHouse row to a FusedTrack proto message.
func trackRowToProto(row *trackRow) *entityv1.FusedTrack {
pos := &commonv1.Position{
Latitude:       row.Latitude,
Longitude:      row.Longitude,
AltitudeMeters: row.AltitudeMeters,
SpeedKnots:     row.SpeedKnots,
HeadingDegrees: row.HeadingDegrees,
}

track := &entityv1.FusedTrack{
TrackId:           row.TrackID,
EntityType:        parseEntityType(row.EntityType),
HostileClass:      parseHostileClass(row.HostileClassification),
EstimatedPosition: pos,
ConfidenceScore:   row.ConfidenceScore,
SourceCount:       uint32(row.SourceCount),
Classification:    parseClassificationLevel(row.ClassificationLevel),
Status:            parseTrackStatus(row.TrackStatus),
UpdatedAt:         timestamppb.New(row.EventTime),
}

for _, sensorID := range row.SourceSensors {
track.Sources = append(track.Sources, &entityv1.SourceAttribution{
SensorId: sensorID,
})
}

return track
}

// --- enum string <-> proto converters ---

func parseEntityType(s string) commonv1.EntityType {
switch s {
case "SURFACE":
return commonv1.EntityType_ENTITY_TYPE_SURFACE
case "AIR":
return commonv1.EntityType_ENTITY_TYPE_AIR
case "SUBSURFACE":
return commonv1.EntityType_ENTITY_TYPE_SUBSURFACE
case "LAND":
return commonv1.EntityType_ENTITY_TYPE_LAND
case "CYBER":
return commonv1.EntityType_ENTITY_TYPE_CYBER
default:
return commonv1.EntityType_ENTITY_TYPE_UNSPECIFIED
}
}

func parseHostileClass(s string) commonv1.HostileClassification {
switch s {
case "HOSTILE":
return commonv1.HostileClassification_HOSTILE_CLASSIFICATION_HOSTILE
case "FRIENDLY":
return commonv1.HostileClassification_HOSTILE_CLASSIFICATION_FRIENDLY
case "NEUTRAL":
return commonv1.HostileClassification_HOSTILE_CLASSIFICATION_NEUTRAL
case "UNKNOWN":
return commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN
default:
return commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNSPECIFIED
}
}

func parseClassificationLevel(s string) commonv1.ClassificationLevel {
switch s {
case "UNCLASSIFIED":
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED
case "PROTECTED_A":
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A
case "PROTECTED_B":
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B
case "PROTECTED_C":
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C
case "SECRET":
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET
default:
return commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED
}
}

func parseTrackStatus(s string) commonv1.TrackStatus {
switch s {
case "NEW":
return commonv1.TrackStatus_TRACK_STATUS_NEW
case "ACTIVE":
return commonv1.TrackStatus_TRACK_STATUS_ACTIVE
case "STALE":
return commonv1.TrackStatus_TRACK_STATUS_STALE
case "DROPPED":
return commonv1.TrackStatus_TRACK_STATUS_DROPPED
case "MERGED":
return commonv1.TrackStatus_TRACK_STATUS_MERGED
default:
return commonv1.TrackStatus_TRACK_STATUS_UNSPECIFIED
}
}

func entityTypeToString(et commonv1.EntityType) string {
switch et {
case commonv1.EntityType_ENTITY_TYPE_SURFACE:
return "SURFACE"
case commonv1.EntityType_ENTITY_TYPE_AIR:
return "AIR"
case commonv1.EntityType_ENTITY_TYPE_SUBSURFACE:
return "SUBSURFACE"
case commonv1.EntityType_ENTITY_TYPE_LAND:
return "LAND"
case commonv1.EntityType_ENTITY_TYPE_CYBER:
return "CYBER"
default:
return "UNSPECIFIED"
}
}

func hostileClassToString(hc commonv1.HostileClassification) string {
switch hc {
case commonv1.HostileClassification_HOSTILE_CLASSIFICATION_HOSTILE:
return "HOSTILE"
case commonv1.HostileClassification_HOSTILE_CLASSIFICATION_FRIENDLY:
return "FRIENDLY"
case commonv1.HostileClassification_HOSTILE_CLASSIFICATION_NEUTRAL:
return "NEUTRAL"
case commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN:
return "UNKNOWN"
default:
return "UNSPECIFIED"
}
}
