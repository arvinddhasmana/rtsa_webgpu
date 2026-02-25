// CLASSIFICATION: UNCLASSIFIED
//go:build integration

// Package testutil provides shared test infrastructure helpers for RTSA integration tests.
// All helpers require RTSA_INTEGRATION_TESTS=true and use testcontainers-go for
// isolated, reproducible test infrastructure.
package testutil

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/testcontainers/testcontainers-go"
	tcredpanda "github.com/testcontainers/testcontainers-go/modules/redpanda"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

// AllTopics lists the 21 RTSA Redpanda topics required by the integration tests.
var AllTopics = []string{
"sensors.radar.tracks",
"sensors.ew.intercepts",
"sensors.elint.detections",
"sensors.isr.observations",
"sensors.ais.positions",
"sensors.cyber.iocs",
"tracks.fused.surface",
"tracks.fused.air",
"tracks.fused.subsurface",
"tracks.fused.land",
"tracks.fused.cyber",
"alerts.anomaly.critical",
"alerts.anomaly.elevated",
"alerts.anomaly.watch",
"feedback.operator.submissions",
"feedback.operator.validated",
"models.anomaly.published",
"audit.events",
"dlq.sensors.radar",
"dlq.sensors.ais",
"dlq.sensors.generic",
}

// TestEnv holds references to test infrastructure started via testcontainers-go.
type TestEnv struct {
RedpandaBrokers string
ClickHouseDSN   string

rpContainer testcontainers.Container
chContainer testcontainers.Container
ctx         context.Context
cancel      context.CancelFunc
}

// SkipUnlessEnabled skips the test if RTSA_INTEGRATION_TESTS is not "true".
// Call at the top of every integration test function.
func SkipUnlessEnabled(t *testing.T) {
t.Helper()
if os.Getenv("RTSA_INTEGRATION_TESTS") != "true" {
t.Skip("integration tests disabled: set RTSA_INTEGRATION_TESTS=true to enable")
}
}

// SetupTestEnv creates a complete test environment with Redpanda and ClickHouse.
// It creates all 21 RTSA topics in Redpanda and initialises the ClickHouse schema.
// Call Teardown() in a defer to clean up resources.
func SetupTestEnv(t *testing.T) *TestEnv {
t.Helper()
SkipUnlessEnabled(t)

ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
env := &TestEnv{ctx: ctx, cancel: cancel}

// Start Redpanda.
broker, rpC, err := startRedpandaContainer(ctx, t)
if err != nil {
cancel()
t.Fatalf("testutil: start redpanda: %v", err)
}
env.RedpandaBrokers = broker
env.rpContainer = rpC

// Create all topics.
if err := createTopics(ctx, broker, AllTopics); err != nil {
cancel()
t.Fatalf("testutil: create topics: %v", err)
}

// Start ClickHouse.
dsn, chC, err := startClickHouseContainer(ctx, t)
if err != nil {
cancel()
t.Fatalf("testutil: start clickhouse: %v", err)
}
env.ClickHouseDSN = dsn
env.chContainer = chC

return env
}

// SetupRedpandaOnly starts only Redpanda with all topics. Lighter-weight for
// tests that do not require ClickHouse.
func SetupRedpandaOnly(t *testing.T) *TestEnv {
t.Helper()
SkipUnlessEnabled(t)

ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
env := &TestEnv{ctx: ctx, cancel: cancel}

broker, rpC, err := startRedpandaContainer(ctx, t)
if err != nil {
cancel()
t.Fatalf("testutil: start redpanda: %v", err)
}
env.RedpandaBrokers = broker
env.rpContainer = rpC

if err := createTopics(ctx, broker, AllTopics); err != nil {
cancel()
t.Fatalf("testutil: create topics: %v", err)
}

return env
}

// Teardown terminates all test containers and releases resources.
func (e *TestEnv) Teardown() {
if e.rpContainer != nil {
_ = e.rpContainer.Terminate(context.Background())
}
if e.chContainer != nil {
_ = e.chContainer.Terminate(context.Background())
}
e.cancel()
}

// NewKafkaProducer returns a franz-go client configured for producing to the test broker.
func (e *TestEnv) NewKafkaProducer(t *testing.T) *kgo.Client {
t.Helper()
cl, err := kgo.NewClient(kgo.SeedBrokers(e.RedpandaBrokers))
if err != nil {
t.Fatalf("testutil: new kafka producer: %v", err)
}
t.Cleanup(cl.Close)
return cl
}

// NewKafkaConsumer returns a franz-go client configured to consume the given topics.
func (e *TestEnv) NewKafkaConsumer(t *testing.T, group string, topics ...string) *kgo.Client {
t.Helper()
cl, err := kgo.NewClient(
kgo.SeedBrokers(e.RedpandaBrokers),
kgo.ConsumerGroup(group),
kgo.ConsumeTopics(topics...),
kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
)
if err != nil {
t.Fatalf("testutil: new kafka consumer: %v", err)
}
t.Cleanup(cl.Close)
return cl
}

// ─── private helpers ──────────────────────────────────────────────────────────

func startRedpandaContainer(ctx context.Context, t *testing.T) (broker string, c testcontainers.Container, err error) {
t.Helper()
rpContainer, err := tcredpanda.Run(ctx, "redpandadata/redpanda:v24.1.1",
tcredpanda.WithAutoCreateTopics(),
)
if err != nil {
return "", nil, fmt.Errorf("start redpanda: %w", err)
}

brokerAddr, err := rpContainer.KafkaSeedBroker(ctx)
if err != nil {
_ = rpContainer.Terminate(ctx)
return "", nil, fmt.Errorf("redpanda broker addr: %w", err)
}

return brokerAddr, rpContainer, nil
}

func startClickHouseContainer(ctx context.Context, t *testing.T) (dsn string, c testcontainers.Container, err error) {
t.Helper()
req := testcontainers.ContainerRequest{
Image:        "clickhouse/clickhouse-server:24.3-alpine",
ExposedPorts: []string{"9000/tcp"},
Env: map[string]string{
"CLICKHOUSE_DB":                       "rtsa",
"CLICKHOUSE_USER":                     "default",
"CLICKHOUSE_PASSWORD":                 "",
"CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT": "1",
},
WaitingFor: wait.ForListeningPort("9000/tcp").WithStartupTimeout(120 * time.Second),
}

c, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
ContainerRequest: req,
Started:          true,
})
if err != nil {
return "", nil, fmt.Errorf("start clickhouse: %w", err)
}

host, err := c.Host(ctx)
if err != nil {
_ = c.Terminate(ctx)
return "", nil, fmt.Errorf("clickhouse host: %w", err)
}
port, err := c.MappedPort(ctx, "9000")
if err != nil {
_ = c.Terminate(ctx)
return "", nil, fmt.Errorf("clickhouse port: %w", err)
}

dsn = fmt.Sprintf("clickhouse://default:@%s:%s/rtsa", host, port.Port())

	// Wait for ClickHouse to fully initialise — the port may open before
	// the rtsa database is created by the CLICKHOUSE_DB env var init.
	if err := waitForClickHouseReady(ctx, dsn); err != nil {
		_ = c.Terminate(ctx)
		return "", nil, fmt.Errorf("clickhouse readiness: %w", err)
	}

	return dsn, c, nil
}

// waitForClickHouseReady retries connecting to ClickHouse until the database is
// ready or the context expires. This handles the race between port readiness
// and database initialisation.
func waitForClickHouseReady(ctx context.Context, dsn string) error {
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
			return fmt.Errorf("context cancelled while waiting: %w", lastErr)
		case <-deadline:
			return fmt.Errorf("timeout after 30s: %w", lastErr)
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

func createTopics(ctx context.Context, broker string, topics []string) error {
cl, err := kgo.NewClient(kgo.SeedBrokers(broker))
if err != nil {
return fmt.Errorf("create kgo client: %w", err)
}
defer cl.Close()

admin := kadm.NewClient(cl)
for _, topic := range topics {
res, err := admin.CreateTopic(ctx, 1, 1, nil, topic)
if err != nil {
return fmt.Errorf("create topic %s: %w", topic, err)
}
if res.Err != nil {
return fmt.Errorf("create topic %s: %w", topic, res.Err)
}
}
return nil
}

// SetupClickHouseOnly starts only ClickHouse with the RTSA schema.
// Use for tests that query ClickHouse without Redpanda.
func SetupClickHouseOnly(t *testing.T) *TestEnv {
t.Helper()
SkipUnlessEnabled(t)

ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
env := &TestEnv{ctx: ctx, cancel: cancel}

dsn, chC, err := startClickHouseContainer(ctx, t)
if err != nil {
cancel()
t.Fatalf("testutil: start clickhouse: %v", err)
}
env.ClickHouseDSN = dsn
env.chContainer = chC

return env
}
