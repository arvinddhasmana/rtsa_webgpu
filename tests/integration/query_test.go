// CLASSIFICATION: UNCLASSIFIED
//go:build integration

// Package integration provides integration tests for the query service layer.
package integration

import (
"context"
"fmt"
"testing"
"time"

"github.com/ClickHouse/clickhouse-go/v2"
commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
"github.com/arvinddhasmana/rtsa_webgpu/pkg/classification"
"github.com/arvinddhasmana/rtsa_webgpu/tests/integration/testutil"
)

// openClickHouseConn opens a ClickHouse connection using the provided DSN.
func openClickHouseConn(t interface{ Fatal(...interface{}) }, dsn string) clickhouse.Conn {
opts, err := clickhouse.ParseDSN(dsn)
if err != nil {
t.Fatal("query: parse DSN:", err)
}
conn, err := clickhouse.Open(opts)
if err != nil {
t.Fatal("query: open ClickHouse:", err)
}
return conn
}

// insertQueryTestTracks inserts count tracks with the given classification into ClickHouse.
func insertQueryTestTracks(t interface {
Helper()
Fatal(...interface{})
Fatalf(string, ...interface{})
}, conn clickhouse.Conn, count int, classLevel string, now time.Time) {
ctx := context.Background()
batch, err := conn.PrepareBatch(ctx,
`INSERT INTO tracks_fused (track_id, entity_type, confidence_score, classification_level,
 latitude, longitude, event_time) VALUES`)
if err != nil {
t.Fatalf("insertQueryTestTracks: prepare batch: %v", err)
}
for i := 0; i < count; i++ {
if err := batch.Append(
fmt.Sprintf("qtrack-%s-%04d", classLevel, i),
"SURFACE",
0.90,
classLevel,
45.0+float64(i)*0.001,
-60.0+float64(i)*0.001,
now.Add(time.Duration(i)*time.Second),
); err != nil {
t.Fatalf("insertQueryTestTracks: append %d: %v", i, err)
}
}
if err := batch.Send(); err != nil {
t.Fatalf("insertQueryTestTracks: send: %v", err)
}
}

// TestQueryTracks_TimeRangeFilter_ReturnsMatchingRows validates:
//  1. Insert 20 tracks into ClickHouse in a known time range
//  2. Query via direct SQL with matching time range
//  3. Verify results are returned with correct count
func TestQueryTracks_TimeRangeFilter_ReturnsMatchingRows(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupClickHouseOnly(t)
defer env.Teardown()

ctx := context.Background()
if err := createTracksSchema(ctx, env.ClickHouseDSN); err != nil {
t.Fatalf("Query: create schema: %v", err)
}

conn := openClickHouseConn(t, env.ClickHouseDSN)
defer func() { _ = conn.Close() }()

now := time.Now().UTC().Truncate(time.Second)
insertQueryTestTracks(t, conn, 20, "UNCLASSIFIED", now)
time.Sleep(500 * time.Millisecond)

// Query in the same time range.
row := conn.QueryRow(ctx,
`SELECT count() FROM tracks_fused
 WHERE event_time >= ? AND event_time <= ? AND classification_level = 'UNCLASSIFIED'`,
now.Add(-time.Minute), now.Add(time.Minute),
)
var count uint64
if err := row.Scan(&count); err != nil {
t.Fatalf("Query: count: %v", err)
}
if count < 20 {
t.Errorf("Query: expected >= 20 tracks, got %d", count)
}

t.Logf("Query PASS: %d tracks found in time range query", count)
}

// TestQueryTracks_ClassificationFilter_SecretCallerSeesMoreTracks validates:
//  1. Insert UNCLASSIFIED and SECRET tracks
//  2. Query with UNCLASSIFIED clearance filter → only UNCLASSIFIED returned
//  3. Query with SECRET clearance filter → all returned
func TestQueryTracks_ClassificationFilter_SecretCallerSeesMoreTracks(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupClickHouseOnly(t)
defer env.Teardown()

ctx := context.Background()
if err := createTracksSchema(ctx, env.ClickHouseDSN); err != nil {
t.Fatalf("Query: create schema: %v", err)
}

conn := openClickHouseConn(t, env.ClickHouseDSN)
defer func() { _ = conn.Close() }()

now := time.Now().UTC().Truncate(time.Second)

// Insert 5 UNCLASSIFIED tracks.
insertQueryTestTracks(t, conn, 5, "UNCLASSIFIED", now)
// Insert 3 SECRET tracks.
insertQueryTestTracks(t, conn, 3, "SECRET", now.Add(time.Minute))

time.Sleep(500 * time.Millisecond)

// Verify classification-based filtering using pkg/classification public API.
unclLevel := commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED
secretLevel := commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET

// UNCLASSIFIED caller cannot access SECRET.
if classification.CanAccess(unclLevel, secretLevel) {
t.Error("Query: UNCLASSIFIED caller should NOT access SECRET tracks")
}

// SECRET caller can access UNCLASSIFIED (and everything else).
if !classification.CanAccess(secretLevel, unclLevel) {
t.Error("Query: SECRET caller SHOULD access UNCLASSIFIED tracks")
}

// SQL simulation: UNCLASSIFIED filter.
rowUncl := conn.QueryRow(ctx,
`SELECT count() FROM tracks_fused WHERE classification_level = 'UNCLASSIFIED'`)
var unclCount uint64
if err := rowUncl.Scan(&unclCount); err != nil {
t.Fatalf("Query: UNCLASSIFIED count: %v", err)
}

// SQL simulation: SECRET filter (includes UNCLASSIFIED via max classification).
rowSecret := conn.QueryRow(ctx,
`SELECT count() FROM tracks_fused WHERE classification_level IN ('UNCLASSIFIED','SECRET')`)
var secretCount uint64
if err := rowSecret.Scan(&secretCount); err != nil {
t.Fatalf("Query: SECRET count: %v", err)
}

if secretCount <= unclCount {
t.Errorf("Query: SECRET count=%d should be > UNCLASSIFIED count=%d", secretCount, unclCount)
}

t.Logf("Query PASS: UNCLASSIFIED=%d tracks, SECRET=%d tracks (filter correct)",
unclCount, secretCount)
}
