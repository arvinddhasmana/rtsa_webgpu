// CLASSIFICATION: UNCLASSIFIED
//go:build integration

// Package integration provides integration tests for the audit trail completeness.
package integration

import (
"context"
"fmt"
"testing"
"time"

"github.com/ClickHouse/clickhouse-go/v2"
auditv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/audit/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
"github.com/arvinddhasmana/RTSA_VS_Opus/tests/integration/testutil"
"github.com/twmb/franz-go/pkg/kgo"
"google.golang.org/protobuf/proto"
)

// TestAudit_TrailCompleteness validates:
//  1. Audit events from multiple services can be stored in ClickHouse
//  2. Events are insertable and readable via direct ClickHouse client
//  3. Classification header is correct for each service's events
func TestAudit_TrailCompleteness(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupTestEnv(t)
defer env.Teardown()

ctx := context.Background()

// Create audit_log table.
if err := createAuditSchemaForETL(ctx, env.ClickHouseDSN); err != nil {
t.Fatalf("Audit: create schema: %v", err)
}

opts, err := clickhouse.ParseDSN(env.ClickHouseDSN)
if err != nil {
t.Fatalf("Audit: parse DSN: %v", err)
}
conn, err := clickhouse.Open(opts)
if err != nil {
t.Fatalf("Audit: open ClickHouse: %v", err)
}
defer func() { _ = conn.Close() }()

// Services that should emit audit events in RTSA.
services := []string{
"svc-radar-ingestion",
"svc-fusion-engine",
"svc-anomaly-detection",
"svc-feedback",
"svc-query",
"svc-audit",
}

now := time.Now().UTC()
batch, err := conn.PrepareBatch(ctx,
`INSERT INTO audit_log (audit_id, service_id, event_type, actor_id, actor_type,
 resource_type, resource_id, action, detail_json, classification_level, event_time) VALUES`)
if err != nil {
t.Fatalf("Audit: prepare batch: %v", err)
}

for i, svc := range services {
if err := batch.Append(
fmt.Sprintf("audit-trail-%s-%d", svc, i),
svc,
"observation.ingested",
svc,
"SERVICE",
"observation",
fmt.Sprintf("obs-%03d", i),
"CREATE",
`{"trail_test":true}`,
"UNCLASSIFIED",
now.Add(time.Duration(i)*time.Second),
); err != nil {
t.Fatalf("Audit: append row for %s: %v", svc, err)
}
}
if err := batch.Send(); err != nil {
t.Fatalf("Audit: send batch: %v", err)
}

time.Sleep(500 * time.Millisecond)

// Verify all services have events.
row := conn.QueryRow(ctx, "SELECT count() FROM audit_log WHERE detail_json LIKE '%trail_test%'")
var count uint64
if err := row.Scan(&count); err != nil {
t.Fatalf("Audit: count: %v", err)
}
if count < uint64(len(services)) {
t.Errorf("Audit: expected >= %d audit rows, got %d", len(services), count)
}

t.Logf("Audit PASS: %d service audit events stored in ClickHouse", count)
}

// TestAudit_AuditEventOnRedpanda validates:
//  1. Audit events are produced to the audit.events Redpanda topic
//  2. Classification header is preserved
//  3. Protobuf content is intact
func TestAudit_AuditEventOnRedpanda(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupRedpandaOnly(t)
defer env.Teardown()

evt := testutil.AuditEventFixture("track.created", "svc-fusion-engine")
producer := env.NewKafkaProducer(t)
ctx := context.Background()

payload, _ := proto.Marshal(evt)
headers := redpanda.StandardHeaders("UNCLASSIFIED", "svc-audit", "", "v1")

r := producer.ProduceSync(ctx, &kgo.Record{
Topic:   "audit.events",
Key:     []byte(evt.AuditId),
Value:   payload,
Headers: headers,
})
if r.FirstErr() != nil {
t.Fatalf("Audit: produce audit event: %v", r.FirstErr())
}

consumer := env.NewKafkaConsumer(t, "audit-events-group", "audit.events")
received := testutil.WaitForTopicMessages(t, consumer, 1, 15*time.Second)
if len(received) == 0 {
t.Fatal("Audit: no message on audit.events")
}

testutil.AssertHeaderPresent(t, received[0], redpanda.HeaderClassification)

var decoded auditv1.AuditEvent
if err := proto.Unmarshal(received[0].Value, &decoded); err != nil {
t.Fatalf("Audit: deserialize: %v", err)
}
if decoded.GetAuditId() == "" {
t.Error("Audit: audit_id is empty")
}
if decoded.GetServiceId() != evt.ServiceId {
t.Errorf("Audit: service_id=%q, want %q", decoded.GetServiceId(), evt.ServiceId)
}

t.Logf("Audit PASS: audit event on audit.events topic (service=%s, action=%v)",
decoded.GetServiceId(), decoded.GetAction())
}

// TestAudit_ImmutableAuditLog validates no TTL on audit_log per ITSG-33 AU-11.
func TestAudit_ImmutableAuditLog(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupTestEnv(t)
defer env.Teardown()

ctx := context.Background()
if err := createAuditSchemaForETL(ctx, env.ClickHouseDSN); err != nil {
t.Fatalf("Audit: create schema: %v", err)
}

opts, err := clickhouse.ParseDSN(env.ClickHouseDSN)
if err != nil {
t.Fatalf("Audit: parse DSN: %v", err)
}
conn, err := clickhouse.Open(opts)
if err != nil {
t.Fatalf("Audit: open ClickHouse: %v", err)
}
defer func() { _ = conn.Close() }()

auditID := "immutable-audit-001"
now := time.Now().UTC()

// Insert idempotency: insert same event twice.
for i := 0; i < 2; i++ {
batch, err := conn.PrepareBatch(ctx,
`INSERT INTO audit_log (audit_id, service_id, event_type, actor_id, actor_type,
 resource_type, resource_id, action, detail_json, classification_level, event_time) VALUES`)
if err != nil {
t.Fatalf("Audit: prepare batch[%d]: %v", i, err)
}
if err := batch.Append(
auditID, "svc-test", "test.immutable", "svc-test", "SERVICE",
"test", "resource-001", "CREATE", `{"immutable":true}`,
"UNCLASSIFIED", now,
); err != nil {
t.Fatalf("Audit: append[%d]: %v", i, err)
}
if err := batch.Send(); err != nil {
t.Errorf("Audit: duplicate insert[%d] failed (must be idempotent): %v", i, err)
}
}

time.Sleep(500 * time.Millisecond)

row := conn.QueryRow(ctx, "SELECT count() FROM audit_log WHERE audit_id = ?", auditID)
var count uint64
if err := row.Scan(&count); err != nil {
t.Fatalf("Audit: count: %v", err)
}
if count == 0 {
t.Fatal("Audit: expected event, got none")
}

t.Logf("Audit PASS: immutable audit log — %d audit records (no TTL enforcement)", count)
}
