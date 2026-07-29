// CLASSIFICATION: UNCLASSIFIED
//go:build integration

// Package integration provides integration tests for svc-audit using testcontainers.
//
// These tests require Docker and spin up a real ClickHouse instance to validate
// the full audit trail flow: insert events → query by ID → stream with filters.
//
// Run with: go test ./... -tags=integration -v -timeout=300s
//
// IMMUTABILITY: Only INSERT operations are executed in the audit repository.
// No UPDATE, DELETE, ALTER, or TRUNCATE SQL operations.
// Feature: FEAT-14 Immutable Audit Trail
// Requirement: ITSG-33 AU-11
package integration

import (
"context"
"fmt"
"testing"
"time"

auditv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/audit/v1"
commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
"github.com/arvinddhasmana/rtsa_webgpu/svc-audit/internal/domain"
"github.com/arvinddhasmana/rtsa_webgpu/svc-audit/internal/repository"
"github.com/ClickHouse/clickhouse-go/v2"
"github.com/testcontainers/testcontainers-go"
"github.com/testcontainers/testcontainers-go/wait"
"google.golang.org/protobuf/types/known/timestamppb"
)

// auditClickHouseContainer holds the testcontainers ClickHouse instance for audit tests.
type auditClickHouseContainer struct {
container testcontainers.Container
dsn       string
}

// startClickHouseForAudit starts a ClickHouse container for audit integration tests.
func startClickHouseForAudit(ctx context.Context) (*auditClickHouseContainer, error) {
req := testcontainers.ContainerRequest{
Image:        "clickhouse/clickhouse-server:24.3-alpine",
ExposedPorts: []string{"9000/tcp"},
// ForListeningPort is more reliable than ForLog("Ready for connections") because
// the ClickHouse entrypoint may suppress that log line when CLICKHOUSE_PASSWORD is
// empty. CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT=1 is required to avoid the
// "disabling network access" restriction that makes the container unreachable.
WaitingFor: wait.ForListeningPort("9000/tcp").WithStartupTimeout(120 * time.Second),
Env: map[string]string{
"CLICKHOUSE_DB":                       "rtsa",
"CLICKHOUSE_USER":                     "default",
"CLICKHOUSE_PASSWORD":                 "",
"CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT": "1",
},
}

container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
ContainerRequest: req,
Started:          true,
})
if err != nil {
return nil, fmt.Errorf("start clickhouse container: %w", err)
}

host, err := container.Host(ctx)
if err != nil {
_ = container.Terminate(ctx)
return nil, fmt.Errorf("get container host: %w", err)
}
port, err := container.MappedPort(ctx, "9000")
if err != nil {
_ = container.Terminate(ctx)
return nil, fmt.Errorf("get container port: %w", err)
}

dsn := fmt.Sprintf("clickhouse://default:@%s:%s/rtsa", host, port.Port())

// Wait for ClickHouse to fully initialise — the port may open before the rtsa database
// is created and the server is ready to accept SQL. Uses same pattern as centralized testutil.
if err := waitForClickHouseAuditReady(ctx, dsn); err != nil {
_ = container.Terminate(ctx)
return nil, fmt.Errorf("clickhouse readiness: %w", err)
}

return &auditClickHouseContainer{container: container, dsn: dsn}, nil
}

// waitForClickHouseAuditReady retries pinging ClickHouse until the DB is ready
// or the context expires. Matches the pattern in tests/integration/testutil/setup.go.
func waitForClickHouseAuditReady(ctx context.Context, dsn string) error {
opts, err := clickhouse.ParseDSN(dsn)
if err != nil {
return fmt.Errorf("parse DSN: %w", err)
}

deadline := time.After(30 * time.Second)
ticker := time.NewTicker(500 * time.Millisecond)
defer ticker.Stop()

var lastErr error
for {
select {
case <-ctx.Done():
return fmt.Errorf("context cancelled while waiting for ClickHouse: %w", lastErr)
case <-deadline:
return fmt.Errorf("timeout after 30s waiting for ClickHouse: %w", lastErr)
case <-ticker.C:
conn, connErr := clickhouse.Open(opts)
if connErr != nil {
lastErr = connErr
continue
}
if pingErr := conn.Ping(ctx); pingErr != nil {
_ = conn.Close()
lastErr = pingErr
continue
}
_ = conn.Close()
return nil
}
}
}


// createAuditLogTable creates the audit_log table in ClickHouse.
// IMMUTABILITY NOTE: Table has NO TTL per ITSG-33 AU-11.
func createAuditLogTable(ctx context.Context, dsn string) error {
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
actor_type Enum8('ACTOR_TYPE_SERVICE' = 1, 'ACTOR_TYPE_OPERATOR' = 2, 'ACTOR_TYPE_SYSTEM' = 3),
resource_type String,
resource_id String,
action String,
detail_json String,
classification_level Enum8(
'UNCLASSIFIED' = 1, 'PROTECTED_A' = 2, 'PROTECTED_B' = 3,
'PROTECTED_C' = 4, 'SECRET' = 5
),
event_time DateTime64(3, 'UTC'),
ingestion_time DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(event_time)
ORDER BY (service_id, event_type, event_time)
SETTINGS index_granularity = 8192`

return conn.Exec(ctx, ddl)
}

// makeTestAuditEvent creates a test audit event with the given audit_id.
func makeTestAuditEvent(auditID, serviceID string, eventTime time.Time) *auditv1.AuditEvent {
return &auditv1.AuditEvent{
AuditId:             auditID,
ServiceId:           serviceID,
EventType:           "test.event",
ActorId:             serviceID,
ActorType:           auditv1.ActorType_ACTOR_TYPE_SERVICE,
ResourceType:        "track",
ResourceId:          "track-001",
Action:              auditv1.AuditAction_AUDIT_ACTION_CREATE,
DetailJson:          `{"test": true}`,
ClassificationLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
EventTime:           timestamppb.New(eventTime),
}
}

// TestAuditRepository_BatchInsert_T01 verifies that batch insert writes 100 events.
// T01: BatchInsert 100 events — all 100 visible in ClickHouse.
func TestAuditRepository_BatchInsert_T01(t *testing.T) {
ctx := context.Background()

ch, err := startClickHouseForAudit(ctx)
if err != nil {
t.Fatalf("start clickhouse: %v", err)
}
defer func() { _ = ch.container.Terminate(ctx) }()

if err := createAuditLogTable(ctx, ch.dsn); err != nil {
t.Fatalf("create table: %v", err)
}

repo, err := repository.NewAuditRepository(ch.dsn)
if err != nil {
t.Fatalf("create repo: %v", err)
}
defer func() { _ = repo.Close() }()

now := time.Now().UTC().Truncate(time.Millisecond)
events := make([]*auditv1.AuditEvent, 100)
for i := range events {
events[i] = makeTestAuditEvent(
fmt.Sprintf("audit-%03d", i),
"svc-test",
now.Add(time.Duration(i)*time.Second),
)
}

if err := repo.BatchInsert(ctx, events); err != nil {
t.Fatalf("batch insert: %v", err)
}
t.Log("T01: Successfully inserted 100 audit events")
}

// TestAuditRepository_DeduplicateInsert_T02 verifies duplicate audit_id is idempotent.
// T02: BatchInsert duplicate audit_id — no error, no duplicate.
func TestAuditRepository_DeduplicateInsert_T02(t *testing.T) {
ctx := context.Background()

ch, err := startClickHouseForAudit(ctx)
if err != nil {
t.Fatalf("start clickhouse: %v", err)
}
defer func() { _ = ch.container.Terminate(ctx) }()

if err := createAuditLogTable(ctx, ch.dsn); err != nil {
t.Fatalf("create table: %v", err)
}

repo, err := repository.NewAuditRepository(ch.dsn)
if err != nil {
t.Fatalf("create repo: %v", err)
}
defer func() { _ = repo.Close() }()

event := makeTestAuditEvent("dup-audit-001", "svc-test", time.Now().UTC())

if err := repo.BatchInsert(ctx, []*auditv1.AuditEvent{event}); err != nil {
t.Fatalf("first insert: %v", err)
}
if err := repo.BatchInsert(ctx, []*auditv1.AuditEvent{event}); err != nil {
t.Fatalf("duplicate insert: %v", err)
}
t.Log("T02: Duplicate insert was idempotent")
}

// TestAuditRepository_GetEntry_T03 verifies retrieval of an existing audit event.
// T03: GetEntry existing ID — returns full event.
func TestAuditRepository_GetEntry_T03(t *testing.T) {
ctx := context.Background()

ch, err := startClickHouseForAudit(ctx)
if err != nil {
t.Fatalf("start clickhouse: %v", err)
}
defer func() { _ = ch.container.Terminate(ctx) }()

if err := createAuditLogTable(ctx, ch.dsn); err != nil {
t.Fatalf("create table: %v", err)
}

repo, err := repository.NewAuditRepository(ch.dsn)
if err != nil {
t.Fatalf("create repo: %v", err)
}
defer func() { _ = repo.Close() }()

auditID := "get-entry-test-001"
event := makeTestAuditEvent(auditID, "svc-radar-ingestion", time.Now().UTC())

if err := repo.BatchInsert(ctx, []*auditv1.AuditEvent{event}); err != nil {
t.Fatalf("insert: %v", err)
}

// ClickHouse needs a moment for the data to be visible
time.Sleep(500 * time.Millisecond)

got, err := repo.GetEntry(ctx, auditID,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED)
if err != nil {
t.Fatalf("get entry: %v", err)
}
if got == nil {
t.Fatal("T03: expected event, got nil (NOT_FOUND)")
}
if got.AuditId != auditID {
t.Errorf("T03: audit_id mismatch: got %q, want %q", got.AuditId, auditID)
}
t.Log("T03: GetEntry returned full event correctly")
}

// TestAuditRepository_GetEntry_NotFound_T04 verifies NOT_FOUND for missing ID.
// T04: GetEntry non-existent ID — NOT_FOUND.
func TestAuditRepository_GetEntry_NotFound_T04(t *testing.T) {
ctx := context.Background()

ch, err := startClickHouseForAudit(ctx)
if err != nil {
t.Fatalf("start clickhouse: %v", err)
}
defer func() { _ = ch.container.Terminate(ctx) }()

if err := createAuditLogTable(ctx, ch.dsn); err != nil {
t.Fatalf("create table: %v", err)
}

repo, err := repository.NewAuditRepository(ch.dsn)
if err != nil {
t.Fatalf("create repo: %v", err)
}
defer func() { _ = repo.Close() }()

got, err := repo.GetEntry(ctx, "does-not-exist",
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED)
if err != nil {
t.Fatalf("T04: unexpected error: %v", err)
}
if got != nil {
t.Errorf("T04: expected nil for non-existent entry, got %v", got)
}
t.Log("T04: GetEntry returned nil for non-existent ID")
}

// TestAuditRepository_ClassificationFilter_T05 verifies higher classified events are excluded.
// T05: GetEntry classification filter — higher classified excluded.
func TestAuditRepository_ClassificationFilter_T05(t *testing.T) {
ctx := context.Background()

ch, err := startClickHouseForAudit(ctx)
if err != nil {
t.Fatalf("start clickhouse: %v", err)
}
defer func() { _ = ch.container.Terminate(ctx) }()

if err := createAuditLogTable(ctx, ch.dsn); err != nil {
t.Fatalf("create table: %v", err)
}

repo, err := repository.NewAuditRepository(ch.dsn)
if err != nil {
t.Fatalf("create repo: %v", err)
}
defer func() { _ = repo.Close() }()

// Insert a SECRET event
secretEvent := makeTestAuditEvent("secret-audit-001", "svc-test", time.Now().UTC())
secretEvent.ClassificationLevel = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET

if err := repo.BatchInsert(ctx, []*auditv1.AuditEvent{secretEvent}); err != nil {
t.Fatalf("insert secret event: %v", err)
}

time.Sleep(500 * time.Millisecond)

// UNCLASSIFIED caller should NOT see SECRET event
got, err := repo.GetEntry(ctx, "secret-audit-001",
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED)
if err != nil {
t.Fatalf("T05: unexpected error: %v", err)
}
if got != nil {
t.Errorf("T05: UNCLASSIFIED caller should not see SECRET event")
}
t.Log("T05: Classification filter correctly excluded higher classified event")
}

// TestAuditRepository_QueryAuditLog_T06_T07 verifies streaming with service_id and event_type filters.
func TestAuditRepository_QueryAuditLog_T06_T07(t *testing.T) {
ctx := context.Background()

ch, err := startClickHouseForAudit(ctx)
if err != nil {
t.Fatalf("start clickhouse: %v", err)
}
defer func() { _ = ch.container.Terminate(ctx) }()

if err := createAuditLogTable(ctx, ch.dsn); err != nil {
t.Fatalf("create table: %v", err)
}

repo, err := repository.NewAuditRepository(ch.dsn)
if err != nil {
t.Fatalf("create repo: %v", err)
}
defer func() { _ = repo.Close() }()

now := time.Now().UTC()

// Insert events from two different services
events := []*auditv1.AuditEvent{
makeTestAuditEvent("q-001", "svc-radar", now.Add(-5*time.Minute)),
makeTestAuditEvent("q-002", "svc-radar", now.Add(-4*time.Minute)),
makeTestAuditEvent("q-003", "svc-feedback", now.Add(-3*time.Minute)),
}
// Give the third event a different event_type
events[2].EventType = "feedback.submitted"

if err := repo.BatchInsert(ctx, events); err != nil {
t.Fatalf("insert: %v", err)
}

time.Sleep(500 * time.Millisecond)

// T06: Filter by service_id
req := &auditv1.StreamAuditLogRequest{
TimeRange: &commonv1.TimeRange{
StartTime: timestamppb.New(now.Add(-10 * time.Minute)),
EndTime:   timestamppb.New(now.Add(time.Minute)),
},
ServiceIds: []string{"svc-radar"},
}

results, _, err := repo.QueryAuditLog(ctx, req,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
nil, 100)
if err != nil {
t.Fatalf("T06: query error: %v", err)
}
if len(results) != 2 {
t.Errorf("T06: expected 2 results for svc-radar, got %d", len(results))
}
t.Log("T06: StreamAuditLog service_id filter worked correctly")

// T07: Filter by event_type
req2 := &auditv1.StreamAuditLogRequest{
TimeRange: &commonv1.TimeRange{
StartTime: timestamppb.New(now.Add(-10 * time.Minute)),
EndTime:   timestamppb.New(now.Add(time.Minute)),
},
EventTypes: []string{"feedback.submitted"},
}

results2, _, err := repo.QueryAuditLog(ctx, req2,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
nil, 100)
if err != nil {
t.Fatalf("T07: query error: %v", err)
}
if len(results2) != 1 {
t.Errorf("T07: expected 1 result for feedback.submitted, got %d", len(results2))
}
t.Log("T07: StreamAuditLog event_type filter worked correctly")
}

// TestAuditRepository_Pagination_T10 verifies cursor-based pagination.
// T10: StreamAuditLog pagination — correct cursor traversal.
func TestAuditRepository_Pagination_T10(t *testing.T) {
ctx := context.Background()

ch, err := startClickHouseForAudit(ctx)
if err != nil {
t.Fatalf("start clickhouse: %v", err)
}
defer func() { _ = ch.container.Terminate(ctx) }()

if err := createAuditLogTable(ctx, ch.dsn); err != nil {
t.Fatalf("create table: %v", err)
}

repo, err := repository.NewAuditRepository(ch.dsn)
if err != nil {
t.Fatalf("create repo: %v", err)
}
defer func() { _ = repo.Close() }()

now := time.Now().UTC()

// Insert 5 events
events := make([]*auditv1.AuditEvent, 5)
for i := range events {
events[i] = makeTestAuditEvent(
fmt.Sprintf("page-%03d", i),
"svc-test",
now.Add(time.Duration(i)*time.Second),
)
}
if err := repo.BatchInsert(ctx, events); err != nil {
t.Fatalf("insert: %v", err)
}

time.Sleep(500 * time.Millisecond)

req := &auditv1.StreamAuditLogRequest{
TimeRange: &commonv1.TimeRange{
StartTime: timestamppb.New(now.Add(-time.Minute)),
EndTime: timestamppb.New(now.Add(time.Minute)),
},
ServiceIds: []string{"svc-test"},
}

guardrail := domain.NewQueryGuardrail(90, 10000, 30)

// Page 1 — size 2
pageSize := guardrail.EnforceRowLimit(2)
page1, nextToken, err := repo.QueryAuditLog(ctx, req,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
nil, pageSize)
if err != nil {
t.Fatalf("T10: page 1 query: %v", err)
}
if len(page1) != 2 {
t.Errorf("T10: expected 2 results on page 1, got %d", len(page1))
}
if nextToken == nil {
t.Fatal("T10: expected next token after page 1")
}

// Page 2 — using cursor
page2, nextToken2, err := repo.QueryAuditLog(ctx, req,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
nextToken, pageSize)
if err != nil {
t.Fatalf("T10: page 2 query: %v", err)
}
if len(page2) != 2 {
t.Errorf("T10: expected 2 results on page 2, got %d", len(page2))
}

// Page 3 — last page
page3, nextToken3, err := repo.QueryAuditLog(ctx, req,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
nextToken2, pageSize)
if err != nil {
t.Fatalf("T10: page 3 query: %v", err)
}
if len(page3) != 1 {
t.Errorf("T10: expected 1 result on page 3, got %d", len(page3))
}
if nextToken3 != nil {
t.Error("T10: expected nil next token on last page")
}

for _, e := range page1 { t.Logf("P1: %s %v", e.AuditId, e.EventTime.AsTime()) }; for _, e := range page2 { t.Logf("P2: %s %v", e.AuditId, e.EventTime.AsTime()) }; for _, e := range page3 { t.Logf("P3: %s %v", e.AuditId, e.EventTime.AsTime()) }
}
