// CLASSIFICATION: UNCLASSIFIED

//go:build integration

// Package integration contains integration tests for svc-query.
// These tests require a running ClickHouse instance and Redpanda broker.
// Run with: go test -tags integration ./tests/integration/...
package integration

import (
"context"
"fmt"
"testing"
"time"

"github.com/ClickHouse/clickhouse-go/v2"
commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
queryv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/query/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/audit"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/config"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/handler"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/repository"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/security"
"google.golang.org/grpc/metadata"
"google.golang.org/protobuf/types/known/timestamppb"
)

// setupClickHouse creates a connection to the test ClickHouse instance and
// ensures the required tables exist with test data.
func setupClickHouse(t *testing.T) clickhouse.Conn {
t.Helper()

conn, err := clickhouse.Open(&clickhouse.Options{
Addr: []string{"localhost:9000"},
Auth: clickhouse.Auth{
Database: "rtsa",
Username: "default",
Password: "",
},
DialTimeout: 10 * time.Second,
})
if err != nil {
t.Skipf("ClickHouse not available: %v", err)
}

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

if err := conn.Ping(ctx); err != nil {
t.Skipf("ClickHouse ping failed: %v", err)
}

return conn
}

// createTestSchema creates the tables and inserts test rows.
func createTestSchema(t *testing.T, conn clickhouse.Conn) {
t.Helper()
ctx := context.Background()

ddl := []string{
`CREATE DATABASE IF NOT EXISTS rtsa`,
`CREATE TABLE IF NOT EXISTS rtsa.tracks_fused (
track_id String,
entity_type Enum8('UNSPECIFIED'=0, 'SURFACE'=1, 'AIR'=2, 'SUBSURFACE'=3, 'LAND'=4, 'CYBER'=5),
hostile_classification Enum8('UNSPECIFIED'=0, 'HOSTILE'=1, 'FRIENDLY'=2, 'NEUTRAL'=3, 'UNKNOWN'=4),
latitude Float64,
longitude Float64,
altitude_meters Nullable(Float64),
speed_knots Nullable(Float64),
heading_degrees Nullable(Float64),
confidence_score Float64,
source_count UInt8,
source_sensors Array(String),
classification_level Enum8('UNCLASSIFIED'=1, 'PROTECTED_A'=2, 'PROTECTED_B'=3, 'PROTECTED_C'=4, 'SECRET'=5),
track_status Enum8('ACTIVE'=1, 'STALE'=2, 'DROPPED'=3, 'MERGED'=4),
event_time DateTime64(3, 'UTC'),
ingestion_time DateTime64(3, 'UTC') DEFAULT now64(3)
) ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(event_time)
ORDER BY (entity_type, track_id, event_time)
SETTINGS index_granularity = 8192`,
}

for _, stmt := range ddl {
if err := conn.Exec(ctx, stmt); err != nil {
t.Fatalf("DDL failed: %v\nSQL: %s", err, stmt)
}
}

// Insert test tracks
batch, err := conn.PrepareBatch(ctx, "INSERT INTO rtsa.tracks_fused")
if err != nil {
t.Fatalf("prepare batch: %v", err)
}

now := time.Now().UTC().Truncate(time.Millisecond)
testRows := []struct {
TrackID              string
EntityType           string
HostileClass         string
Lat, Lon             float64
Confidence           float64
SourceCount          uint8
SourceSensors        []string
ClassificationLevel  string
TrackStatus          string
EventTime            time.Time
}{
{"track-001", "SURFACE", "HOSTILE", 45.0, -75.0, 0.92, 3, []string{"radar-01"}, "UNCLASSIFIED", "ACTIVE", now.Add(-1 * time.Hour)},
{"track-002", "AIR", "UNKNOWN", 48.0, -73.0, 0.75, 2, []string{"radar-02", "ew-01"}, "PROTECTED_B", "ACTIVE", now.Add(-30 * time.Minute)},
{"track-003", "SURFACE", "FRIENDLY", 43.0, -79.0, 0.88, 1, []string{"ais-01"}, "SECRET", "ACTIVE", now.Add(-15 * time.Minute)},
}

for _, row := range testRows {
if err := batch.Append(
row.TrackID, row.EntityType, row.HostileClass,
row.Lat, row.Lon, (*float64)(nil), (*float64)(nil), (*float64)(nil),
row.Confidence, row.SourceCount, row.SourceSensors,
row.ClassificationLevel, row.TrackStatus, row.EventTime,
); err != nil {
t.Fatalf("batch append: %v", err)
}
}
if err := batch.Send(); err != nil {
t.Fatalf("batch send: %v", err)
}

t.Cleanup(func() {
_ = conn.Exec(context.Background(), "TRUNCATE TABLE rtsa.tracks_fused")
})
}

// newTestHandler creates a fully wired handler connected to the test ClickHouse instance.
func newTestHandler(t *testing.T, cfg *config.Config) *handler.Handler {
t.Helper()

guard, err := domain.NewQueryGuardrail(30, 100000, 30, 100, 1000)
if err != nil {
t.Fatalf("NewQueryGuardrail: %v", err)
}

chClient, err := repository.NewClickHouseClient(cfg)
if err != nil {
t.Skipf("ClickHouse client init failed: %v", err)
}
t.Cleanup(func() { _ = chClient.Close() })

classFilter := security.NewClassificationFilter()
tracksRepo := repository.NewTracksRepository(chClient, classFilter, guard)
anomalyRepo := repository.NewAnomalyRepository(chClient, classFilter, guard)
auditRepoI := repository.NewAuditRepository(chClient, classFilter, guard)
emitter := &audit.NoopEmitter{}

return handler.New(tracksRepo, anomalyRepo, auditRepoI, emitter, guard, "svc-query-test")
}

// IT01: Insert tracks -> query via service (end-to-end ClickHouse)
func TestIT01_QueryTracks_EndToEnd(t *testing.T) {
conn := setupClickHouse(t)
defer conn.Close()

createTestSchema(t, conn)

cfg := &config.Config{
ClickHouseDSN:      "clickhouse://default:@localhost:9000/rtsa",
ClickHouseDatabase: "rtsa",
MaxQueryRangeDays:  30,
MaxResultRows:      100000,
QueryTimeoutSec:    30,
DefaultPageSize:    100,
MaxPageSize:        1000,
}

h := newTestHandler(t, cfg)

now := time.Now().UTC()
ctx := metadata.NewIncomingContext(context.Background(),
metadata.Pairs(security.MetadataKeyClassification, "5")) // SECRET

resp, err := h.QueryTracks(ctx, &queryv1.QueryTracksRequest{
TimeRange: &commonv1.TimeRange{
StartTime: timestamppb.New(now.Add(-2 * time.Hour)),
EndTime:   timestamppb.New(now.Add(time.Minute)),
},
})
if err != nil {
t.Fatalf("QueryTracks: %v", err)
}
// With SECRET clearance, should see all 3 tracks
if len(resp.Tracks) != 3 {
t.Errorf("expected 3 tracks with SECRET clearance, got %d", len(resp.Tracks))
}
}

// IT02: Verify classification injection filters results correctly
func TestIT02_ClassificationFilter_Enforced(t *testing.T) {
conn := setupClickHouse(t)
defer conn.Close()

createTestSchema(t, conn)

cfg := &config.Config{
ClickHouseDSN:      "clickhouse://default:@localhost:9000/rtsa",
ClickHouseDatabase: "rtsa",
MaxQueryRangeDays:  30,
MaxResultRows:      100000,
QueryTimeoutSec:    30,
DefaultPageSize:    100,
MaxPageSize:        1000,
}

h := newTestHandler(t, cfg)

now := time.Now().UTC()
tr := &commonv1.TimeRange{
StartTime: timestamppb.New(now.Add(-2 * time.Hour)),
EndTime:   timestamppb.New(now.Add(time.Minute)),
}

tests := []struct {
clearanceOrdinal string
wantMinTracks    int
wantMaxTracks    int
desc             string
}{
{"1", 1, 1, "UNCLASSIFIED clearance sees only UNCLASSIFIED track"},
{"3", 2, 2, "PROTECTED_B clearance sees UNCLASSIFIED + PROTECTED_B tracks"},
{"5", 3, 3, "SECRET clearance sees all tracks"},
}

for _, tc := range tests {
t.Run(tc.desc, func(t *testing.T) {
ctx := metadata.NewIncomingContext(context.Background(),
metadata.Pairs(security.MetadataKeyClassification, tc.clearanceOrdinal))

resp, err := h.QueryTracks(ctx, &queryv1.QueryTracksRequest{TimeRange: tr})
if err != nil {
t.Fatalf("QueryTracks: %v", err)
}
n := len(resp.Tracks)
if n < tc.wantMinTracks || n > tc.wantMaxTracks {
t.Errorf("got %d tracks, want [%d,%d]", n, tc.wantMinTracks, tc.wantMaxTracks)
}
})
}
}

// IT03: Verify pagination across multiple pages
func TestIT03_Pagination(t *testing.T) {
conn := setupClickHouse(t)
defer conn.Close()

createTestSchema(t, conn)

cfg := &config.Config{
ClickHouseDSN:      "clickhouse://default:@localhost:9000/rtsa",
ClickHouseDatabase: "rtsa",
MaxQueryRangeDays:  30,
MaxResultRows:      100000,
QueryTimeoutSec:    30,
DefaultPageSize:    100,
MaxPageSize:        1000,
}

h := newTestHandler(t, cfg)

now := time.Now().UTC()
ctx := metadata.NewIncomingContext(context.Background(),
metadata.Pairs(security.MetadataKeyClassification, "5"))

tr := &commonv1.TimeRange{
StartTime: timestamppb.New(now.Add(-2 * time.Hour)),
EndTime:   timestamppb.New(now.Add(time.Minute)),
}

// Page 1: page size 2
resp1, err := h.QueryTracks(ctx, &queryv1.QueryTracksRequest{
TimeRange: tr,
Pagination: &commonv1.PaginationRequest{PageSize: 2},
})
if err != nil {
t.Fatalf("page 1 QueryTracks: %v", err)
}
if len(resp1.Tracks) != 2 {
t.Errorf("page 1: expected 2 tracks, got %d", len(resp1.Tracks))
}
if resp1.Pagination.NextPageToken == "" {
t.Fatal("expected next page token after page 1")
}

// Page 2: use the token
resp2, err := h.QueryTracks(ctx, &queryv1.QueryTracksRequest{
TimeRange: tr,
Pagination: &commonv1.PaginationRequest{
PageSize:  2,
PageToken: resp1.Pagination.NextPageToken,
},
})
if err != nil {
t.Fatalf("page 2 QueryTracks: %v", err)
}
if len(resp2.Tracks) != 1 {
t.Errorf("page 2: expected 1 track, got %d", len(resp2.Tracks))
}
// No more pages
if resp2.Pagination.NextPageToken != "" {
t.Errorf("expected no next page token after last page, got %q", resp2.Pagination.NextPageToken)
}

// All track IDs should be distinct
seen := make(map[string]bool)
for _, tr2 := range append(resp1.Tracks, resp2.Tracks...) {
if seen[tr2.TrackId] {
t.Errorf("duplicate track_id in pagination: %s", tr2.TrackId)
}
seen[tr2.TrackId] = true
}

fmt.Printf("IT03: paginated across %d+%d=%d tracks\n", len(resp1.Tracks), len(resp2.Tracks), len(resp1.Tracks)+len(resp2.Tracks))
}
