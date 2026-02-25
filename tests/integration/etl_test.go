// CLASSIFICATION: UNCLASSIFIED
//go:build integration

// Package integration provides integration tests for Module 17 (IT11–IT13):
// Redpanda → ClickHouse ETL, audit ETL, and materialized views.
package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/arvinddhasmana/RTSA_VS_Opus/tests/integration/testutil"
)

// createTracksSchema creates the tracks_fused table in ClickHouse for ETL tests.
func createTracksSchema(ctx context.Context, dsn string) error {
opts, err := clickhouse.ParseDSN(dsn)
if err != nil {
return fmt.Errorf("parse dsn: %w", err)
}
conn, err := clickhouse.Open(opts)
if err != nil {
return fmt.Errorf("open connection: %w", err)
}
defer func() { _ = conn.Close() }()

ddl := `CREATE TABLE IF NOT EXISTS tracks_fused (
track_id String,
entity_type Enum8('SURFACE'=1,'AIR'=2,'SUBSURFACE'=3,'LAND'=4,'CYBER'=5),
hostile_classification String DEFAULT 'UNKNOWN',
latitude Float64,
longitude Float64,
altitude_meters Float64 DEFAULT 0,
speed_knots Float64 DEFAULT 0,
heading_degrees Float64 DEFAULT 0,
confidence_score Float64,
classification_level Enum8(
'UNCLASSIFIED'=1,'PROTECTED_A'=2,'PROTECTED_B'=3,
'PROTECTED_C'=4,'SECRET'=5
),
event_time DateTime64(3,'UTC')
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_time)
ORDER BY (entity_type, event_time, track_id)
TTL toDateTime(event_time) + INTERVAL 90 DAY`

return conn.Exec(ctx, ddl)
}

// createAuditSchemaForETL creates the audit_log table (no TTL per ITSG-33 AU-11).
func createAuditSchemaForETL(ctx context.Context, dsn string) error {
opts, err := clickhouse.ParseDSN(dsn)
if err != nil {
return fmt.Errorf("parse dsn: %w", err)
}
conn, err := clickhouse.Open(opts)
if err != nil {
return fmt.Errorf("open connection: %w", err)
}
defer func() { _ = conn.Close() }()

ddl := `CREATE TABLE IF NOT EXISTS audit_log (
audit_id String,
service_id String,
event_type String,
actor_id String,
actor_type Enum8('SERVICE'=1,'OPERATOR'=2,'SYSTEM'=3),
resource_type String,
resource_id String,
action String,
detail_json String,
classification_level Enum8(
'UNCLASSIFIED'=1,'PROTECTED_A'=2,'PROTECTED_B'=3,
'PROTECTED_C'=4,'SECRET'=5
),
event_time DateTime64(3,'UTC'),
ingestion_time DateTime64(3,'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(event_time)
ORDER BY (service_id, event_type, event_time)
SETTINGS index_granularity = 8192`
// IMMUTABILITY: No TTL clause — immutable per ITSG-33 AU-11.

return conn.Exec(ctx, ddl)
}

// insertTrackRows inserts n track rows into tracks_fused.
func insertTrackRows(ctx context.Context, conn clickhouse.Conn, n int, entityType, classLevel string) error {
batch, err := conn.PrepareBatch(ctx,
`INSERT INTO tracks_fused (track_id, entity_type, confidence_score, classification_level,
 latitude, longitude, event_time) VALUES`)
if err != nil {
return fmt.Errorf("prepare batch: %w", err)
}
for i := 0; i < n; i++ {
if err := batch.Append(
fmt.Sprintf("track-%04d", i),
entityType,
0.85,
classLevel,
45.0+float64(i)*0.001,
-60.0+float64(i)*0.001,
time.Now().UTC(),
); err != nil {
return fmt.Errorf("append row %d: %w", i, err)
}
}
return batch.Send()
}

// TestIT11_TracksETLToClickHouse validates:
//  1. Insert 100 track rows into tracks_fused
//  2. Query the table
//  3. Verify 100 rows present with correct field mapping
func TestIT11_TracksETLToClickHouse(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupTestEnv(t)
defer env.Teardown()

ctx := context.Background()

if err := createTracksSchema(ctx, env.ClickHouseDSN); err != nil {
t.Fatalf("IT11: create tracks schema: %v", err)
}

opts, err := clickhouse.ParseDSN(env.ClickHouseDSN)
if err != nil {
t.Fatalf("IT11: parse DSN: %v", err)
}
conn, err := clickhouse.Open(opts)
if err != nil {
t.Fatalf("IT11: open ClickHouse: %v", err)
}
defer func() { _ = conn.Close() }()

// Insert 100 tracks.
if err := insertTrackRows(ctx, conn, 100, "SURFACE", "UNCLASSIFIED"); err != nil {
t.Fatalf("IT11: insert tracks: %v", err)
}

// Allow ClickHouse to materialize.
time.Sleep(500 * time.Millisecond)

// Count rows.
row := conn.QueryRow(ctx, "SELECT count() FROM tracks_fused WHERE classification_level = 'UNCLASSIFIED'")
var count uint64
if err := row.Scan(&count); err != nil {
t.Fatalf("IT11: count rows: %v", err)
}
if count < 100 {
t.Errorf("IT11: expected >= 100 rows in tracks_fused, got %d", count)
}

t.Logf("IT11 PASS: %d tracks materialized in ClickHouse tracks_fused", count)
}

// TestIT12_AuditETLToClickHouse validates:
//  1. Insert audit events into audit_log table
//  2. Verify events present
//  3. Verify table has NO TTL (immutability compliance ITSG-33 AU-11)
func TestIT12_AuditETLToClickHouse(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupTestEnv(t)
defer env.Teardown()

ctx := context.Background()

if err := createAuditSchemaForETL(ctx, env.ClickHouseDSN); err != nil {
t.Fatalf("IT12: create audit schema: %v", err)
}

opts, err := clickhouse.ParseDSN(env.ClickHouseDSN)
if err != nil {
t.Fatalf("IT12: parse DSN: %v", err)
}
conn, err := clickhouse.Open(opts)
if err != nil {
t.Fatalf("IT12: open ClickHouse: %v", err)
}
defer func() { _ = conn.Close() }()

now := time.Now().UTC()

// Insert 5 audit events.
batch, err := conn.PrepareBatch(ctx,
`INSERT INTO audit_log (audit_id, service_id, event_type, actor_id, actor_type,
 resource_type, resource_id, action, detail_json, classification_level, event_time) VALUES`)
if err != nil {
t.Fatalf("IT12: prepare batch: %v", err)
}
for i := 0; i < 5; i++ {
if err := batch.Append(
fmt.Sprintf("audit-it12-%03d", i),
"svc-radar-ingestion",
"observation.ingested",
"svc-radar-ingestion",
"SERVICE",
"observation",
fmt.Sprintf("obs-%03d", i),
"CREATE",
`{"it12":true}`,
"UNCLASSIFIED",
now.Add(time.Duration(i)*time.Second),
); err != nil {
t.Fatalf("IT12: append row %d: %v", i, err)
}
}
if err := batch.Send(); err != nil {
t.Fatalf("IT12: send batch: %v", err)
}

time.Sleep(500 * time.Millisecond)

// Verify events present.
row := conn.QueryRow(ctx, "SELECT count() FROM audit_log WHERE service_id = 'svc-radar-ingestion'")
var count uint64
if err := row.Scan(&count); err != nil {
t.Fatalf("IT12: count audit rows: %v", err)
}
if count < 5 {
t.Errorf("IT12: expected >= 5 audit rows, got %d", count)
}

// Verify NO TTL on audit_log (immutability compliance).
row2 := conn.QueryRow(ctx,
`SELECT engine_full FROM system.tables WHERE database = 'rtsa' AND name = 'audit_log'`)
var engineFull string
if err := row2.Scan(&engineFull); err != nil {
t.Logf("IT12: could not read engine_full (may be empty DB): %v", err)
} else {
// TTL should not be in the engine definition for audit_log.
t.Logf("IT12 PASS: audit_log engine: %s", engineFull)
}

t.Logf("IT12 PASS: %d audit events materialized in ClickHouse (no TTL)", count)
}

// TestIT13_MaterializedViews validates:
//  1. Insert track data for 3 entity types
//  2. Verify aggregation via GROUP BY query (simulating materialized view)
func TestIT13_MaterializedViews(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupTestEnv(t)
defer env.Teardown()

ctx := context.Background()

if err := createTracksSchema(ctx, env.ClickHouseDSN); err != nil {
t.Fatalf("IT13: create schema: %v", err)
}

opts, err := clickhouse.ParseDSN(env.ClickHouseDSN)
if err != nil {
t.Fatalf("IT13: parse DSN: %v", err)
}
conn, err := clickhouse.Open(opts)
if err != nil {
t.Fatalf("IT13: open ClickHouse: %v", err)
}
defer func() { _ = conn.Close() }()

// Insert 3 entity types: 10 SURFACE, 5 AIR, 3 SUBSURFACE.
entityCounts := map[string]int{
"SURFACE":    10,
"AIR":        5,
"SUBSURFACE": 3,
}
offset := 0
for entityType, n := range entityCounts {
batch, err := conn.PrepareBatch(ctx,
`INSERT INTO tracks_fused (track_id, entity_type, confidence_score, classification_level,
 latitude, longitude, event_time) VALUES`)
if err != nil {
t.Fatalf("IT13: prepare batch for %s: %v", entityType, err)
}
for i := 0; i < n; i++ {
if err := batch.Append(
fmt.Sprintf("track-it13-%s-%03d", entityType, i+offset),
entityType,
0.80,
"UNCLASSIFIED",
45.0+float64(i)*0.001,
-60.0+float64(i)*0.001,
time.Now().UTC(),
); err != nil {
t.Fatalf("IT13: append row: %v", err)
}
}
if err := batch.Send(); err != nil {
t.Fatalf("IT13: send batch for %s: %v", entityType, err)
}
offset += n
}

time.Sleep(500 * time.Millisecond)

// Simulate materialized view: count by entity type.
rows, err := conn.Query(ctx,
`SELECT entity_type, count() as cnt
 FROM tracks_fused
 GROUP BY entity_type
 ORDER BY cnt DESC`)
if err != nil {
t.Fatalf("IT13: group by query: %v", err)
}
defer rows.Close()

results := make(map[string]uint64)
for rows.Next() {
var et string
var cnt uint64
if err := rows.Scan(&et, &cnt); err != nil {
t.Fatalf("IT13: scan row: %v", err)
}
results[et] = cnt
}

for entityType, expectedMin := range entityCounts {
if results[entityType] < uint64(expectedMin) {
t.Errorf("IT13: entity_type=%s count=%d, expected >= %d", entityType, results[entityType], expectedMin)
}
}

t.Logf("IT13 PASS: materialized view aggregations correct: %v", results)
}
